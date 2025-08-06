package cpi

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadCpiServiceKeyFile(t *testing.T) {
	properServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
				}
			}`
	faultyServiceKey := `this is not json`

	tests := []struct {
		name              string
		serviceKey        string
		wantCpiServiceKey ServiceKey
		wantedErrorMsg    string
	}{
		{
			"happy path",
			properServiceKey,
			ServiceKey{
				OAuth: OAuth{
					Host:                  "https://demo",
					OAuthTokenProviderURL: "https://demo/oauth/token",
					ClientID:              "demouser",
					ClientSecret:          "******",
				},
			},
			"",
		},
		{
			"faulty json",
			faultyServiceKey,
			ServiceKey{},
			"error unmarshalling serviceKey: invalid character 'h' in literal true (expecting 'r')",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCpiServiceKey, err := ReadCpiServiceKey(tt.serviceKey)
			if tt.wantedErrorMsg != "" {
				assert.EqualError(t, err, tt.wantedErrorMsg)
			}
			assert.Equal(t, tt.wantCpiServiceKey, gotCpiServiceKey)
		})
	}
}
