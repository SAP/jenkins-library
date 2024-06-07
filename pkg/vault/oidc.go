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
	"github.com/pkg/errors"
)

type JwtPayload struct {
	Expire int64 `json:"exp"`
}

// getOIDCToken returns the generated OIDC token and sets it in the env
func (v Client) getOIDCToken(roleID string) (string, error) {
	oidcPath := sanitizePath(path.Join("identity/oidc/token/", roleID))
	c := v.lClient
	jwt, err := c.Read(oidcPath)
	if err != nil {
		return "", err
	}

	token := jwt.Data["token"].(string)
	log.RegisterSecret(token)
	os.Setenv("PIPER_OIDCIdentityToken", token)

	return token, nil
}

// getJWTTokenPayload returns the payload of the JWT token using base64 decoding
func getJWTTokenPayload(token string) ([]byte, error) {
	parts := strings.Split(token, ".")
	if len(parts) >= 2 {
		substr := parts[1]
		decodedBytes, err := base64.RawStdEncoding.DecodeString(substr)
		if err != nil {
			return nil, errors.Wrap(err, "JWT payload couldn't be decoded: %s")
		}
		return decodedBytes, nil
	}

	return nil, fmt.Errorf("Not a valid JWT token")
}

func oidcTokenIsValid(token string) bool {
	payload, err := getJWTTokenPayload(token)
	if err != nil {
		log.Entry().Debugf("OIDC token couldn't be validated: %s", err)
		return false
	}

	var jwtPayload JwtPayload
	err = json.Unmarshal(payload, &jwtPayload)
	if err != nil {
		log.Entry().Debugf("OIDC token couldn't be validated: %s", err)
		return false
	}

	expiryTime := time.Unix(jwtPayload.Expire, 0)
	currentTime := time.Now()

	return expiryTime.After(currentTime)
}

// GetOIDCTokenByValidation returns the token if token is expired then get a new token else return old token
func (v Client) GetOIDCTokenByValidation(roleID string) (string, error) {
	token := os.Getenv("PIPER_OIDCIdentityToken")
	if token != "" && oidcTokenIsValid(token) {
		return token, nil
	}

	token, err := v.getOIDCToken(roleID)
	if token == "" || err != nil {
		return "", errors.Wrap(err, "failed to get OIDC token")
	}

	return token, nil
}
