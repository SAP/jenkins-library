package cmd

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/config/interpolation"
	"github.com/SAP/jenkins-library/pkg/log"
)

type gitopsAppVault interface {
	GetKvSecret(string) (map[string]string, error)
}

// Keep credentials on a copy of the configuration; never persist installation tokens
// to the common pipeline environment or change the global credential resolver.
func withGitopsAuthentication(options gitopsUpdateDeploymentOptions, vault gitopsAppVault, client *http.Client, run func(*gitopsUpdateDeploymentOptions) error) error {
	switch options.AuthenticationMode {
	case "", "legacy":
		return run(&options)
	case "githubApp":
	default:
		return errors.New("unsupported GitOps authenticationMode")
	}
	api, owner, repo, err := gitopsAppTarget(options.ServerURL, options.GithubAppAPIURL)
	if err != nil {
		return err
	}
	appID, privateKey, err := gitopsAppCredentials(options, vault)
	if err != nil {
		return err
	}
	jwt, err := gitopsAppJWT(appID, privateKey, time.Now())
	if err != nil {
		return err
	}
	log.RegisterSecret(jwt)
	// Redirects can move signed credentials to another endpoint. Fail closed, also
	// for same-host redirects: the repository installation must match the target.
	restrictedClient := *client
	restrictedClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	var installation struct {
		ID int64 `json:"id"`
	}
	err = gitopsAppRequest(&restrictedClient, http.MethodGet, api+"/repos/"+owner+"/"+repo+"/installation", jwt, nil, http.StatusOK, &installation)
	if err != nil {
		return fmt.Errorf("discover GitHub App installation: %w", err)
	}
	if installation.ID <= 0 {
		return errors.New("GitHub App installation response contains no installation ID")
	}
	body, _ := json.Marshal(map[string]any{
		"repositories": []string{repo},
		"permissions":  map[string]string{"contents": "write"},
	})
	var access struct {
		Token     string `json:"token"`
		ExpiresAt string `json:"expires_at"`
	}
	err = gitopsAppRequest(&restrictedClient, http.MethodPost, fmt.Sprintf("%s/app/installations/%d/access_tokens", api, installation.ID), jwt, body, http.StatusCreated, &access)
	if err != nil {
		return fmt.Errorf("create GitHub App installation token: %w", err)
	}
	if strings.TrimSpace(access.Token) == "" {
		return errors.New("GitHub App returned an empty installation token")
	}
	log.RegisterSecret(access.Token)
	defer func() {
		if err := gitopsAppRequest(&restrictedClient, http.MethodDelete, api+"/installation/token", access.Token, nil, http.StatusNoContent, nil); err != nil {
			// A successful push must not be retried because token cleanup failed.
			log.Entry().Warn("Could not revoke GitHub App installation token; it remains valid until its expiration")
		}
	}()
	expiresAt, parseErr := time.Parse(time.RFC3339, access.ExpiresAt)
	if parseErr != nil || !expiresAt.After(time.Now()) {
		return errors.New("GitHub App returned an expired installation token or no expiration")
	}
	options.Password = access.Token
	if options.Username == "" {
		options.Username = "x-access-token"
	}
	return run(&options)
}

func gitopsAppCredentials(options gitopsUpdateDeploymentOptions, vault gitopsAppVault) (string, string, error) {
	if options.GithubAppID != "" || options.GithubAppPrivateKey != "" {
		if strings.TrimSpace(options.GithubAppID) == "" || strings.TrimSpace(options.GithubAppPrivateKey) == "" {
			return "", "", errors.New("githubAppId and githubAppPrivateKey must be supplied together")
		}
		log.RegisterSecret(options.GithubAppPrivateKey)
		return options.GithubAppID, options.GithubAppPrivateKey, nil
	}
	if vault == nil {
		return "", "", errors.New("GitHub App credentials require configured Vault or both githubAppId and githubAppPrivateKey")
	}
	name := options.GithubAppVaultSecretName
	if name == "" {
		name = "githubApp"
	}
	if path.IsAbs(name) || path.Clean(name) != name || name == "." || name == ".." || strings.HasPrefix(name, "../") {
		return "", "", errors.New("githubAppVaultSecretName must be a relative secret name within the configured Vault roots")
	}
	roots := map[string]any{}
	for key, value := range map[string]string{"vaultPath": options.VaultPath, "vaultBasePath": options.VaultBasePath, "vaultPipelineName": options.VaultPipelineName} {
		if value != "" {
			roots[key] = value
		}
	}
	visited := map[string]bool{}
	for _, template := range config.VaultRootPaths {
		root, ok := interpolation.ResolveString(template, roots)
		if !ok {
			continue
		}
		secretPath := path.Join(root, name)
		if visited[secretPath] {
			continue
		}
		visited[secretPath] = true
		secret, err := vault.GetKvSecret(secretPath)
		// Vault errors can contain response bodies. Do not include them in logs.
		if err != nil {
			return "", "", errors.New("GitHub App Vault lookup failed; check Vault access and connectivity")
		}
		if secret == nil {
			continue
		}
		appID, key := secret["appId"], secret["privateKey"]
		if strings.TrimSpace(appID) == "" || strings.TrimSpace(key) == "" {
			return "", "", errors.New("GitHub App Vault secret must contain both appId and privateKey")
		}
		log.RegisterSecret(key)
		return appID, key, nil
	}
	return "", "", errors.New("GitHub App secret was not found under the configured Vault roots")
}

func gitopsAppTarget(repositoryURL, apiURL string) (string, string, string, error) {
	target, err := url.Parse(repositoryURL)
	if err != nil || target.Scheme != "https" || target.Host == "" || target.User != nil || target.RawQuery != "" || target.Fragment != "" || target.RawPath != "" {
		return "", "", "", errors.New("GitHub App authentication requires an HTTPS repository URL without credentials, query or fragment")
	}
	parts := strings.Split(strings.TrimPrefix(strings.TrimSuffix(target.Path, "/"), "/"), "/")
	if len(parts) != 2 {
		return "", "", "", errors.New("GitHub App repository URL must identify exactly one owner and repository")
	}
	owner, repo := parts[0], strings.TrimSuffix(parts[1], ".git")
	valid := regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
	if !valid.MatchString(owner) || !valid.MatchString(repo) || owner == "." || owner == ".." || repo == "." || repo == ".." {
		return "", "", "", errors.New("invalid GitHub App repository owner or name")
	}
	apiHost := target.Host
	if strings.EqualFold(target.Host, "github.com") {
		apiHost = "api.github.com"
	}
	if apiURL == "" {
		apiURL = "https://" + apiHost
		if apiHost != "api.github.com" {
			apiURL += "/api/v3"
		}
	}
	api, err := url.Parse(apiURL)
	if err != nil || api.Scheme != "https" || !strings.EqualFold(api.Host, apiHost) || api.User != nil || api.RawQuery != "" || api.Fragment != "" || api.RawPath != "" {
		return "", "", "", errors.New("githubAppApiUrl must use HTTPS on the target repository's API host without credentials, query or fragment")
	}
	return strings.TrimRight(api.String(), "/"), owner, repo, nil
}

func gitopsAppJWT(appID, privateKey string, now time.Time) (string, error) {
	block, rest := pem.Decode([]byte(privateKey))
	if block == nil || len(bytes.TrimSpace(rest)) != 0 {
		return "", errors.New("githubAppPrivateKey must contain one PEM encoded RSA private key")
	}
	var key *rsa.PrivateKey
	switch block.Type {
	case "RSA PRIVATE KEY":
		key, _ = x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err == nil {
			key, _ = parsed.(*rsa.PrivateKey)
		}
	}
	if key == nil || key.Validate() != nil {
		return "", errors.New("githubAppPrivateKey must be an unencrypted RSA private key in PKCS#1 or PKCS#8 format")
	}
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	claims, _ := json.Marshal(map[string]any{"iss": appID, "iat": now.Add(-time.Minute).Unix(), "exp": now.Add(9 * time.Minute).Unix()})
	unsigned := header + "." + base64.RawURLEncoding.EncodeToString(claims)
	digest := sha256.Sum256([]byte(unsigned))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		return "", errors.New("could not sign GitHub App JWT")
	}
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func gitopsAppRequest(client *http.Client, method, endpoint, token string, body []byte, status int, result any) error {
	req, err := http.NewRequest(method, endpoint, bytes.NewReader(body))
	if err != nil {
		return errors.New("invalid GitHub App request")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("GitHub App request failed; check API connectivity and TLS trust")
	}
	defer resp.Body.Close()
	if resp.StatusCode != status {
		return fmt.Errorf("GitHub App API returned HTTP %d", resp.StatusCode)
	}
	if result != nil {
		if err := json.NewDecoder(io.LimitReader(resp.Body, 1024*1024)).Decode(result); err != nil {
			return errors.New("GitHub App API returned an invalid response")
		}
	}
	return nil
}
