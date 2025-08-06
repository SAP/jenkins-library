package config

import "github.com/SAP/jenkins-library/pkg/config"

// GeneralConfigOptions contains all global configuration options for piper binary
type GeneralConfigOptions struct {
	GitHubAccessTokens   map[string]string // map of tokens with url as key in order to maintain url-specific tokens
	CorrelationID        string
	CustomConfig         string
	GitHubTokens         []string // list of entries in form of <server>:<token> to allow token authentication for downloading config / defaults
	DefaultConfig        []string // ordered list of Piper default configurations. Can be filePath or ENV containing JSON in format 'ENV:MY_ENV_VAR'
	IgnoreCustomDefaults bool
	ParametersJSON       string
	EnvRootPath          string
	NoTelemetry          bool
	StageName            string
	StepConfigJSON       string
	StepMetadata         string // metadata to be considered, can be filePath or ENV containing JSON in format 'ENV:MY_ENV_VAR'
	StepName             string
	Verbose              bool
	LogFormat            string
	VaultRoleID          string
	VaultRoleSecretID    string
	VaultToken           string
	VaultServerURL       string
	VaultNamespace       string
	VaultPath            string
	SystemTrustToken     string
	HookConfig           HookConfiguration
	MetaDataResolver     func() map[string]config.StepData
	GCPJsonKeyFilePath   string
	GCSFolderPath        string
	GCSBucketId          string
	GCSSubFolder         string
}

// HookConfiguration contains the configuration for supported hooks, so far Sentry and Splunk are supported.
type HookConfiguration struct {
	GCPPubSubConfig   GCPPubSubConfiguration   `json:"gcpPubSub,omitempty"`
	SentryConfig      SentryConfiguration      `json:"sentry,omitempty"`
	SplunkConfig      SplunkConfiguration      `json:"splunk,omitempty"`
	OIDCConfig        OIDCConfiguration        `json:"oidc,omitempty"`
	SystemTrustConfig SystemTrustConfiguration `json:"systemtrust,omitempty"`
}

type GCPPubSubConfiguration struct {
	Enabled          bool   `json:"enabled"`
	ProjectNumber    string `json:"projectNumber,omitempty"`
	IdentityPool     string `json:"identityPool,omitempty"`
	IdentityProvider string `json:"identityProvider,omitempty"`
	Topic            string `json:"topic,omitempty"`
}

// SentryConfiguration defines the configuration options for the Sentry logging system
type SentryConfiguration struct {
	Dsn string `json:"dsn,omitempty"`
}

// SplunkConfiguration defines the configuration options for the Splunk logging system
type SplunkConfiguration struct {
	Dsn               string `json:"dsn,omitempty"`
	Token             string `json:"token,omitempty"`
	Index             string `json:"index,omitempty"`
	SendLogs          bool   `json:"sendLogs"`
	ProdCriblEndpoint string `json:"prodCriblEndpoint,omitempty"`
	ProdCriblToken    string `json:"prodCriblToken,omitempty"`
	ProdCriblIndex    string `json:"prodCriblIndex,omitempty"`
}

// OIDCConfiguration defines the configuration options for the OpenID Connect authentication system
type OIDCConfiguration struct {
	RoleID string `json:",roleID,omitempty"`
}

type SystemTrustConfiguration struct {
	ServerURL           string `json:"baseURL,omitempty"`
	TokenEndPoint       string `json:"tokenEndPoint,omitempty"`
	TokenQueryParamName string `json:"tokenQueryParamName,omitempty"`
}
