//go:build unit

package cmd

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type gitopsAppVaultMock struct {
	paths   []string
	secrets map[string]map[string]string
	err     error
}

func (v *gitopsAppVaultMock) GetKvSecret(path string) (map[string]string, error) {
	v.paths = append(v.paths, path)
	return v.secrets[path], v.err
}

func gitopsAppTestKey(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key, string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}))
}

func TestGitopsAppJWT(t *testing.T) {
	key, pkcs1 := gitopsAppTestKey(t)
	pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	for _, encoded := range []string{pkcs1, string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8}))} {
		now := time.Unix(1700000000, 0)
		token, err := gitopsAppJWT("Iv1.example", encoded, now)
		require.NoError(t, err)
		parts := strings.Split(token, ".")
		require.Len(t, parts, 3)
		claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
		require.NoError(t, err)
		var claims map[string]any
		require.NoError(t, json.Unmarshal(claimsBytes, &claims))
		assert.Equal(t, "Iv1.example", claims["iss"])
		assert.Equal(t, float64(now.Unix()-60), claims["iat"])
		assert.Equal(t, float64(now.Unix()+540), claims["exp"])
		signature, err := base64.RawURLEncoding.DecodeString(parts[2])
		require.NoError(t, err)
		digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
		assert.NoError(t, rsa.VerifyPKCS1v15(&key.PublicKey, crypto.SHA256, digest[:], signature))
	}
	for _, key := range []string{"secret-invalid-key", pkcs1 + pkcs1, "", string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("bad")}))} {
		_, err := gitopsAppJWT("123", key, time.Now())
		require.Error(t, err)
		assert.NotContains(t, err.Error(), "secret-invalid-key")
	}
}

func TestGitopsAppTarget(t *testing.T) {
	for _, tc := range []struct{ url, override, api, owner, repo string }{
		{"https://github.com/org/infra.git", "", "https://api.github.com", "org", "infra"},
		{"https://git.example.com/team/infra/", "", "https://git.example.com/api/v3", "team", "infra"},
		{"https://git.example.com:8443/team/infra", "https://git.example.com:8443/custom/api/", "https://git.example.com:8443/custom/api", "team", "infra"},
	} {
		api, owner, repo, err := gitopsAppTarget(tc.url, tc.override)
		require.NoError(t, err)
		assert.Equal(t, tc.api, api)
		assert.Equal(t, tc.owner, owner)
		assert.Equal(t, tc.repo, repo)
	}
	for _, target := range []string{"http://github.com/a/b", "git@github.com:a/b.git", "https://secret@github.com/a/b", "https://github.com/a/b?secret=1", "https://github.com/a/b#x", "https://github.com/a", "https://github.com/a/b/c", "https://github.com/a/%2e%2e", "https://github.com/a/.."} {
		_, _, _, err := gitopsAppTarget(target, "")
		require.Error(t, err, target)
		assert.NotContains(t, err.Error(), "secret")
	}
	for _, api := range []string{"https://other.example/api/v3", "http://github.com", "https://api.github.com?token=secret", "https://secret@api.github.com"} {
		_, _, _, err := gitopsAppTarget("https://github.com/a/b", api)
		require.Error(t, err, api)
	}
}

func TestGitopsAppCredentials(t *testing.T) {
	options := gitopsUpdateDeploymentOptions{VaultPath: "kv/project", VaultBasePath: "kv/base", VaultPipelineName: "pipeline", GithubAppVaultSecretName: "deploymentApp"}
	paths := []string{"kv/project/deploymentApp", "kv/base/pipeline/deploymentApp", "kv/base/GROUP-SECRETS/deploymentApp"}
	for i, location := range paths {
		vault := &gitopsAppVaultMock{secrets: map[string]map[string]string{location: {"appId": "123", "privateKey": "private"}}}
		id, key, err := gitopsAppCredentials(options, vault)
		require.NoError(t, err)
		assert.Equal(t, "123", id)
		assert.Equal(t, "private", key)
		assert.Equal(t, paths[:i+1], vault.paths)
	}
	t.Run("direct pair avoids Vault", func(t *testing.T) {
		options.GithubAppID = "direct"
		options.GithubAppPrivateKey = "direct-key"
		vault := &gitopsAppVaultMock{err: errors.New("must not be called")}
		id, key, err := gitopsAppCredentials(options, vault)
		require.NoError(t, err)
		assert.Equal(t, "direct", id)
		assert.Equal(t, "direct-key", key)
		assert.Empty(t, vault.paths)
		options.GithubAppPrivateKey = ""
		_, _, err = gitopsAppCredentials(options, vault)
		require.Error(t, err)
		assert.Empty(t, vault.paths)
		options.GithubAppID = ""
	})
	for _, secret := range []map[string]string{{"appId": "123"}, {"privateKey": "key"}, {}} {
		vault := &gitopsAppVaultMock{secrets: map[string]map[string]string{paths[0]: secret, paths[1]: {"appId": "other", "privateKey": "other-key"}}}
		_, _, err := gitopsAppCredentials(options, vault)
		require.Error(t, err)
		assert.Equal(t, paths[:1], vault.paths)
	}
	vault := &gitopsAppVaultMock{err: errors.New("response contains secret-value")}
	_, _, err := gitopsAppCredentials(options, vault)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "secret-value")
	assert.Equal(t, paths[:1], vault.paths)
	_, _, err = gitopsAppCredentials(options, nil)
	require.Error(t, err)
	emptyRoots := &gitopsAppVaultMock{}
	_, _, err = gitopsAppCredentials(gitopsUpdateDeploymentOptions{}, emptyRoots)
	require.Error(t, err)
	assert.Empty(t, emptyRoots.paths)
	options.GithubAppVaultSecretName = "../escape"
	_, _, err = gitopsAppCredentials(options, emptyRoots)
	require.Error(t, err)
	assert.Empty(t, emptyRoots.paths)
}

func TestGitopsAppLegacyIsolation(t *testing.T) {
	for _, mode := range []string{"", "legacy"} {
		options := gitopsUpdateDeploymentOptions{AuthenticationMode: mode, ServerURL: "ssh://non-github/repo", Username: "existing", Password: "existing-password", GithubAppID: "incomplete-unused-app"}
		vault := &gitopsAppVaultMock{err: errors.New("must not read")}
		called := false
		err := withGitopsAuthentication(options, vault, nil, func(actual *gitopsUpdateDeploymentOptions) error {
			called = true
			assert.Equal(t, options, *actual)
			return nil
		})
		require.NoError(t, err)
		assert.True(t, called)
		assert.Empty(t, vault.paths)
	}
	err := withGitopsAuthentication(gitopsUpdateDeploymentOptions{AuthenticationMode: "automatic"}, nil, nil, func(*gitopsUpdateDeploymentOptions) error { t.Fatal("must not run"); return nil })
	require.Error(t, err)
}

func TestGitopsAppTokenLifecycle(t *testing.T) {
	_, key := gitopsAppTestKey(t)
	for _, fails := range []bool{false, true} {
		t.Run(map[bool]string{false: "successful deployment", true: "failed deployment"}[fails], func(t *testing.T) {
			var calls []string
			token := strings.Repeat("variable-length-token", 30)
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls = append(calls, r.Method+" "+r.URL.Path)
				switch r.Method + " " + r.URL.Path {
				case "GET /api/v3/repos/target/infra/installation":
					assert.True(t, strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ey"))
					io.WriteString(w, `{"id":42}`)
				case "POST /api/v3/app/installations/42/access_tokens":
					var body map[string]any
					require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
					assert.Equal(t, map[string]any{"repositories": []any{"infra"}, "permissions": map[string]any{"contents": "write"}}, body)
					w.WriteHeader(http.StatusCreated)
					json.NewEncoder(w).Encode(map[string]any{"token": token, "expires_at": time.Now().Add(time.Hour)})
				case "DELETE /api/v3/installation/token":
					assert.Equal(t, "Bearer "+token, r.Header.Get("Authorization"))
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected request %s", r.URL.Path)
					w.WriteHeader(404)
				}
			}))
			defer server.Close()
			options := gitopsUpdateDeploymentOptions{AuthenticationMode: "githubApp", ServerURL: server.URL + "/target/infra.git", GithubAppID: "123", GithubAppPrivateKey: key, Username: "github-actions", Password: "injected-source-repo-token"}
			deploymentErr := errors.New("deployment failed")
			err := withGitopsAuthentication(options, nil, server.Client(), func(actual *gitopsUpdateDeploymentOptions) error {
				calls = append(calls, "deployment")
				assert.Equal(t, token, actual.Password)
				assert.Equal(t, "github-actions", actual.Username)
				if fails {
					return deploymentErr
				}
				return nil
			})
			if fails {
				assert.ErrorIs(t, err, deploymentErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, []string{"GET /api/v3/repos/target/infra/installation", "POST /api/v3/app/installations/42/access_tokens", "deployment", "DELETE /api/v3/installation/token"}, calls)
			assert.Equal(t, "injected-source-repo-token", options.Password)
		})
	}
}

func TestGitopsAppFailsClosed(t *testing.T) {
	_, key := gitopsAppTestKey(t)
	for _, tc := range []struct {
		name   string
		status int
		body   string
		mint   bool
		revoke bool
	}{
		{"permission denied", 403, `{"message":"secret response"}`, false, false},
		{"redirect", 302, ``, false, false},
		{"invalid installation", 200, `{"id":0}`, false, false},
		{"invalid JSON", 200, `invalid secret response`, false, false},
		{"token failure", 403, `{"message":"secret response"}`, true, false},
		{"empty token", 201, `{"token":"","expires_at":"2099-01-01T00:00:00Z"}`, true, false},
		{"invalid expiration", 201, `{"token":"secret-token","expires_at":"invalid"}`, true, true},
		{"expired token", 201, `{"token":"secret-token","expires_at":"2000-01-01T00:00:00Z"}`, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			revoked := false
			requests := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				if r.Method == http.MethodDelete {
					revoked = true
					w.WriteHeader(204)
					return
				}
				if tc.mint && r.Method == http.MethodGet {
					io.WriteString(w, `{"id":42}`)
					return
				}
				w.Header().Set("Location", "/redirected")
				w.WriteHeader(tc.status)
				io.WriteString(w, tc.body)
			}))
			defer server.Close()
			options := gitopsUpdateDeploymentOptions{AuthenticationMode: "githubApp", ServerURL: server.URL + "/target/infra", GithubAppID: "123", GithubAppPrivateKey: key, Password: "must-not-fallback"}
			err := withGitopsAuthentication(options, nil, server.Client(), func(*gitopsUpdateDeploymentOptions) error { t.Error("deployment must not run"); return nil })
			require.Error(t, err)
			assert.NotContains(t, err.Error(), "secret")
			assert.Equal(t, tc.revoke, revoked)
			if tc.name == "redirect" {
				assert.Equal(t, 1, requests)
			}
		})
	}
}

func TestGitopsAppMetadataCompatibility(t *testing.T) {
	metadata := gitopsUpdateDeploymentMetadata()
	for _, param := range metadata.Spec.Inputs.Parameters {
		if strings.HasPrefix(param.Name, "githubApp") {
			assert.Empty(t, param.ResourceRef, "App secrets must not be resolved for legacy runs")
		}
	}
	validator, err := validation.New(validation.WithJSONNamesForStructFields(), validation.WithPredefinedErrorMessages())
	require.NoError(t, err)
	for _, tc := range []struct {
		mode, user, password string
		valid                bool
	}{
		{"legacy", "user", "token", true}, {"legacy", "", "token", false}, {"legacy", "user", "", false}, {"githubApp", "", "", true}, {"unknown", "user", "token", false},
	} {
		err := validator.ValidateStruct(gitopsUpdateDeploymentOptions{AuthenticationMode: tc.mode, Username: tc.user, Password: tc.password, Tool: "kustomize", MaxPushAttempts: 1})
		if tc.valid {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
	// Exercise configuration merging: GENERAL must not accidentally opt in all
	// deployments, while step/stage selection and the existing Vault roots survive.
	for _, tc := range []struct{ yaml, mode string }{
		{"general:\n  authenticationMode: githubApp\n", "legacy"},
		{"steps:\n  gitopsUpdateDeployment:\n    authenticationMode: githubApp\n", "githubApp"},
		{"stages:\n  Release:\n    authenticationMode: githubApp\n", "githubApp"},
	} {
		var c config.Config
		merged, err := c.GetStepConfig(nil, "", io.NopCloser(strings.NewReader(tc.yaml)), nil, true, metadata.GetParameterFilters(), metadata, nil, "Release", "gitopsUpdateDeployment")
		require.NoError(t, err)
		assert.Equal(t, tc.mode, merged.Config["authenticationMode"])
	}
}

func TestGitopsPushRetry(t *testing.T) {
	for _, tc := range []struct {
		name                     string
		attempts, failures, want int
		err                      error
	}{
		{"default stays single attempt", 0, 3, 1, &gitopsPushError{errors.New("non-fast-forward update")}},
		{"conflict recovers", 3, 1, 2, &gitopsPushError{errors.New("non-fast-forward update")}},
		{"fetch first recovers", 3, 2, 3, &gitopsPushError{errors.New("fetch first")}},
		{"lock conflict bounded", 3, 4, 3, &gitopsPushError{errors.New("cannot lock ref refs/heads/main: is at abc123 but expected def456")}},
		{"authentication does not retry", 3, 3, 1, &gitopsPushError{errors.New("authentication required")}},
		{"clone errors do not retry", 3, 3, 1, errors.New("non-fast-forward during clone")},
		{"transport does not retry", 3, 3, 1, &gitopsPushError{errors.New("connection reset")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			var delays []time.Duration
			err := retryGitopsUpdateDeployment(tc.attempts, func(delay time.Duration) { delays = append(delays, delay) }, func() error {
				calls++
				if calls <= tc.failures {
					return tc.err
				}
				return nil
			})
			assert.Equal(t, tc.want, calls)
			assert.Len(t, delays, tc.want-1)
			if tc.want > tc.failures {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tc.err)
			}
			if len(delays) > 0 {
				assert.Equal(t, 2*time.Second, delays[0])
			}
			if len(delays) > 1 {
				assert.Equal(t, 5*time.Second, delays[1])
			}
		})
	}
}

type gitopsAppGitMock struct {
	*gitUtilsMock
	t         *testing.T
	token     string
	pushError error
}

func (g *gitopsAppGitMock) PlainClone(user, password, server, branch, directory string, certs []byte) error {
	assert.Equal(g.t, "deployment-bot", user)
	assert.Equal(g.t, g.token, password)
	return g.gitUtilsMock.PlainClone(user, password, server, branch, directory, certs)
}
func (g *gitopsAppGitMock) CommitFiles(files []string, message, author string) (plumbing.Hash, error) {
	assert.Equal(g.t, "deployment-bot", author)
	return g.gitUtilsMock.CommitFiles(files, message, author)
}
func (g *gitopsAppGitMock) PushChangesToRepository(user, password string, force *bool, certs []byte) error {
	assert.Equal(g.t, "deployment-bot", user)
	assert.Equal(g.t, g.token, password)
	assert.False(g.t, *force)
	return g.pushError
}

func TestGitopsAppDeploymentAndFreshCloneRetry(t *testing.T) {
	_, key := gitopsAppTestKey(t)
	const token = "installation-token-for-fixture"
	revoked := false
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			io.WriteString(w, `{"id":42}`)
		case http.MethodPost:
			w.WriteHeader(201)
			json.NewEncoder(w).Encode(map[string]any{"token": token, "expires_at": time.Now().Add(time.Hour)})
		case http.MethodDelete:
			revoked = true
			w.WriteHeader(204)
		}
	}))
	defer server.Close()
	options := gitopsUpdateDeploymentOptions{
		AuthenticationMode: "githubApp", GithubAppID: "123", GithubAppPrivateKey: key,
		ServerURL: server.URL + "/target/infra", Username: "deployment-bot", Password: "source-token",
		BranchName: "main", FilePath: "dir1/dir2/depl.yaml", Tool: "kubectl", ContainerName: "myContainer",
		ContainerRegistryURL: "https://myregistry.com", ContainerImageNameTag: "myFancyContainer:1337", MaxPushAttempts: 3,
	}
	var directories []string
	err := withGitopsAuthentication(options, nil, server.Client(), func(authenticated *gitopsUpdateDeploymentOptions) error {
		return retryGitopsUpdateDeployment(3, func(time.Duration) {}, func() error {
			mock := &gitopsAppGitMock{gitUtilsMock: &gitUtilsMock{}, t: t, token: token}
			if len(directories) == 0 {
				mock.pushError = errors.New("non-fast-forward update")
			}
			err := runGitopsUpdateDeployment(authenticated, &gitOpsExecRunnerMock{expectedYaml: expectedYaml}, mock, &filesMock{})
			directories = append(directories, mock.temporaryDirectory)
			assert.Equal(t, []string{expectedYaml}, mock.savedFiles)
			return err
		})
	})
	require.NoError(t, err)
	require.Len(t, directories, 2)
	assert.NotEqual(t, directories[0], directories[1])
	for _, directory := range directories {
		_, err := os.Stat(directory)
		assert.True(t, os.IsNotExist(err))
	}
	assert.True(t, revoked)
}

func TestGitopsVaultFlagCompatibility(t *testing.T) {
	for _, args := range [][]string{{"--vaultPath", "kv/cli", "gitopsUpdateDeployment"}, {"gitopsUpdateDeployment", "--vaultPath", "kv/cli"}} {
		root := &cobra.Command{Use: "piper"}
		root.PersistentFlags().String("vaultPath", "", "existing global flag")
		options := gitopsUpdateDeploymentOptions{}
		step := &cobra.Command{Use: "gitopsUpdateDeployment", RunE: func(cmd *cobra.Command, _ []string) error {
			metadata := gitopsUpdateDeploymentMetadata()
			filters := metadata.GetParameterFilters()
			flags := config.AvailableFlagValues(cmd, &filters)
			assert.Equal(t, "kv/cli", flags["vaultPath"])
			return nil
		}}
		addGitopsUpdateDeploymentFlags(step, &options)
		for _, name := range []string{"branchName", "serverUrl", "filePath", "containerRegistryUrl", "containerImageNameTag", "tool"} {
			delete(step.Flags().Lookup(name).Annotations, cobra.BashCompOneRequiredFlag)
		}
		root.AddCommand(step)
		root.SetArgs(args)
		require.NoError(t, root.Execute())
	}
	// Brace globs allow selecting only dev and staging, without matching prod.
	root := t.TempDir()
	for _, name := range []string{"dev", "staging", "prod"} {
		require.NoError(t, os.MkdirAll(filepath.Join(root, name), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(root, name, "kustomization.yaml"), []byte("images: []"), 0600))
	}
	files := filesMock{}
	matches, err := files.Glob(filepath.Join(root, "{dev,staging}", "kustomization.yaml"))
	require.NoError(t, err)
	assert.Len(t, matches, 2)
	for _, match := range matches {
		assert.NotContains(t, match, "/prod/")
	}
}
