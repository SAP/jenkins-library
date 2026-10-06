package stagingcredentials

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"
)

const (
	ModeJustInTimeV1 = "justInTimeV1"

	OperationRead  = "read"
	OperationWrite = "write"
)

var safeIdentifier = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// HTTPClient is the interface for making HTTP requests.
type HTTPClient interface {
	Do(request *http.Request) (*http.Response, error)
}

// Request holds all inputs needed to acquire a repository credential.
type Request struct {
	SystemTrustURL          string
	SystemTrustSessionToken string
	StagingServiceURL       string
	GroupID                 string
	RepositoryID            string
	Operation               string
}

// Credential is the repository-level credential returned by the staging service.
type Credential struct {
	RepositoryID  string
	RepositoryURL string
	Username      string
	Password      string
}

// Provider fetches just-in-time staging repository credentials via the
// System Trust → Staging Service exchange.
type Provider struct {
	client HTTPClient
}

// NewProvider returns a Provider with a secure default HTTP client.
func NewProvider() *Provider {
	return &Provider{
		client: &http.Client{
			Timeout: 15 * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// NewProviderWithClient returns a Provider using the supplied HTTP client.
// Intended for testing.
func NewProviderWithClient(client HTTPClient) *Provider {
	return &Provider{client: client}
}

// Acquire performs the three-step exchange to obtain a repository credential:
//  1. POST to System Trust /tokens to exchange the session token for a staging-service token
//  2. GET /trust/impersonate on the staging service to obtain a group-scoped access token
//  3. GET /repository/credentials/{repositoryID} to obtain the actual Nexus user/password
func (p *Provider) Acquire(ctx context.Context, request Request) (Credential, error) {
	if err := validateRequest(request); err != nil {
		return Credential{}, err
	}

	serviceToken, err := p.getStagingServiceToken(ctx, request)
	if err != nil {
		return Credential{}, err
	}

	accessToken, err := p.impersonate(ctx, request, serviceToken)
	if err != nil {
		return Credential{}, err
	}

	credential, err := p.getRepositoryCredential(ctx, request, accessToken)
	if err != nil {
		return Credential{}, err
	}

	repositoryURL, err := CanonicalRepositoryURL(request.StagingServiceURL, request.RepositoryID)
	if err != nil {
		return Credential{}, err
	}
	credential.RepositoryURL = repositoryURL
	return credential, nil
}

func (p *Provider) getStagingServiceToken(ctx context.Context, request Request) (string, error) {
	endpoint, err := joinURL(request.SystemTrustURL, "tokens")
	if err != nil {
		return "", fmt.Errorf("create System Trust token URL: %w", err)
	}
	body, err := json.Marshal([]map[string]string{{
		"system": "staging-service",
		"scope":  "pipeline",
	}})
	if err != nil {
		return "", fmt.Errorf("marshal System Trust token request: %w", err)
	}

	response := map[string]string{}
	if err := p.doJSON(ctx, http.MethodPost, endpoint, request.SystemTrustSessionToken, body, &response); err != nil {
		return "", fmt.Errorf("get Staging Service token from System Trust: %w", err)
	}
	token := response["staging-service"]
	if err := validateSecret("Staging Service token", token); err != nil {
		return "", err
	}
	return token, nil
}

func (p *Provider) impersonate(ctx context.Context, request Request, serviceToken string) (string, error) {
	endpoint, err := joinURL(request.StagingServiceURL, "trust/impersonate")
	if err != nil {
		return "", fmt.Errorf("create Staging Service impersonation URL: %w", err)
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("parse Staging Service impersonation URL: %w", err)
	}
	query := parsed.Query()
	query.Set("group", request.GroupID)
	parsed.RawQuery = query.Encode()

	response := struct {
		AccessToken string `json:"access_token"`
	}{}
	if err := p.doJSON(ctx, http.MethodGet, parsed.String(), serviceToken, nil, &response); err != nil {
		return "", fmt.Errorf("impersonate Staging Service group: %w", err)
	}
	if err := validateSecret("Staging Service access token", response.AccessToken); err != nil {
		return "", err
	}
	return response.AccessToken, nil
}

func (p *Provider) getRepositoryCredential(ctx context.Context, request Request, accessToken string) (Credential, error) {
	endpoint, err := joinURL(
		request.StagingServiceURL,
		"repository/credentials/"+url.PathEscape(request.RepositoryID),
	)
	if err != nil {
		return Credential{}, fmt.Errorf("create repository credentials URL: %w", err)
	}
	response := struct {
		Repository string `json:"repository"`
		User       string `json:"user"`
		Password   string `json:"password"`
	}{}
	if err := p.doJSON(ctx, http.MethodGet, endpoint, accessToken, nil, &response); err != nil {
		return Credential{}, fmt.Errorf("get repository credential: %w", err)
	}
	if response.Repository != request.RepositoryID {
		return Credential{}, errors.New("credential repository ID does not match request")
	}
	if err := validateSecret("repository username", response.User); err != nil {
		return Credential{}, err
	}
	if err := validateSecret("repository password", response.Password); err != nil {
		return Credential{}, err
	}
	return Credential{
		RepositoryID: response.Repository,
		Username:     response.User,
		Password:     response.Password,
	}, nil
}

func (p *Provider) doJSON(ctx context.Context, method, endpoint, token string, body []byte, target any) error {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("request returned HTTP %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// CanonicalRepositoryURL derives the Nexus repository URL from the staging
// service base URL and the repository ID.
func CanonicalRepositoryURL(stagingServiceURL, repositoryID string) (string, error) {
	if err := validateIdentifier("repository ID", repositoryID); err != nil {
		return "", err
	}
	parsed, err := url.Parse(strings.TrimSuffix(stagingServiceURL, "/"))
	if err != nil {
		return "", fmt.Errorf("parse Staging Service URL: %w", err)
	}
	if parsed.Scheme != "https" || parsed.Host == "" {
		return "", errors.New("Staging Service URL must use HTTPS")
	}
	if !strings.HasSuffix(parsed.Path, "/api") {
		return "", errors.New("Staging Service URL must end with /api")
	}
	parsed.Path = path.Join(strings.TrimSuffix(parsed.Path, "/api"), "repository", repositoryID) + "/"
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String(), nil
}

// BasicHeader returns a Base64-encoded Basic Auth header value.
func BasicHeader(username, password string) (string, error) {
	if err := validateSecret("repository username", username); err != nil {
		return "", err
	}
	if err := validateSecret("repository password", password); err != nil {
		return "", err
	}
	value := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	return "Basic " + value, nil
}

func validateRequest(request Request) error {
	if request.Operation != OperationRead && request.Operation != OperationWrite {
		return errors.New("unsupported repository credential operation")
	}
	if err := validateSecret("System Trust session token", request.SystemTrustSessionToken); err != nil {
		return err
	}
	if err := validateIdentifier("staging group ID", request.GroupID); err != nil {
		return err
	}
	if err := validateIdentifier("repository ID", request.RepositoryID); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"System Trust URL":    request.SystemTrustURL,
		"Staging Service URL": request.StagingServiceURL,
	} {
		parsed, err := url.Parse(value)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return fmt.Errorf("%s must use HTTPS", name)
		}
	}
	return nil
}

func validateIdentifier(name, value string) error {
	if !safeIdentifier.MatchString(value) {
		return fmt.Errorf("invalid %s", name)
	}
	return nil
}

func validateSecret(name, value string) error {
	if value == "" || strings.ContainsAny(value, "\r\n\x00") {
		return fmt.Errorf("invalid %s", name)
	}
	return nil
}

func joinURL(baseURL, endpoint string) (string, error) {
	parsed, err := url.Parse(strings.TrimSuffix(baseURL, "/"))
	if err != nil {
		return "", err
	}
	parsed.Path = path.Join(parsed.Path, endpoint)
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String(), nil
}
