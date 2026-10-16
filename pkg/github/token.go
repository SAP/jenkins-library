package github

import (
	"os"

	"github.com/SAP/jenkins-library/pkg/log"
)

const (
	gitHubTokenEnv       = "GITHUB_TOKEN"
	defaultTokenUsername = "x-access-token"
)

// WithActionsTokenFallback completes the credentials resolved from configuration, System Trust and Vault:
// if no token is set, the GITHUB_TOKEN environment variable is used as last resort,
// and a token without username gets a default username.
func WithActionsTokenFallback(username, token string) (string, string) {
	if token == "" {
		if envToken := os.Getenv(gitHubTokenEnv); envToken != "" {
			log.RegisterSecret(envToken)
			log.Entry().Infof("Using the %s environment variable: no token from configuration, System Trust or Vault", gitHubTokenEnv)
			token = envToken
		}
	}
	if token != "" && username == "" {
		username = defaultTokenUsername
	}
	return username, token
}
