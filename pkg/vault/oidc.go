package vault

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
)

type jwtPayload struct {
	Expire int64 `json:"exp"`
}

// getOIDCToken returns the generated OIDC token and sets it in the env
func (c *Client) getOIDCToken(roleID string) (string, error) {
	oidcPath := sanitizePath(path.Join("identity/oidc/token/", roleID))
	jwt, err := c.logical.Read(oidcPath)
	if err != nil {
		return "", err
	}

	token := jwt.Data["token"].(string)
	if token == "" {
		return "", fmt.Errorf("received an empty token")
	}

	log.RegisterSecret(token)
	os.Setenv("PIPER_OIDCIdentityToken", token)

	return token, nil
}

// getJWTPayload returns the payload of the JWT token using base64 decoding
func getJWTPayload(token string) (*jwtPayload, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("not a valid JWT token")
	}

	decodedBytes, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("JWT payload couldn't be decoded: %w", err)
	}

	var payload jwtPayload
	if err = json.Unmarshal(decodedBytes, &payload); err != nil {
		return nil, fmt.Errorf("JWT unmarshal failed: %w", err)
	}

	return &payload, nil
}

func oidcTokenIsValid(token string) bool {
	if token == "" {
		return false
	}

	jwtTokenPayload, err := getJWTPayload(token)
	if err != nil {
		log.Entry().Debugf("OIDC token couldn't be validated: %s", err)
		return false
	}

	expiryTime := time.Unix(jwtTokenPayload.Expire, 0)
	currentTime := time.Now()

	return expiryTime.After(currentTime)
}

// GetOIDCTokenByValidation returns the token if token is expired then get a new token else return old token
func (c *Client) GetOIDCTokenByValidation(roleID string) (string, error) {
	token := os.Getenv("PIPER_OIDCIdentityToken")
	if oidcTokenIsValid(token) {
		return token, nil
	}

	log.Entry().Debug("obtaining new OIDC token")
	token, err := c.getOIDCToken(roleID)
	if err != nil {
		return "", err
	}

	return token, nil
}
