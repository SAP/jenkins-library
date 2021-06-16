package cpi

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadCpiServiceKeyFile(t *testing.T) {
	properServiceKey := `{
			"url": "https://demo",
			"uaa": {
				"clientid": "demouser",
				"clientsecret": "******",
				"url": "https://demo/oauth/token"
				}
			}`
	faultyServiceKey := `this is not json`

	tests := []struct {
		name              string
		serviceKey        string
		wantCpiServiceKey CpiServiceKey
		wantedErrorMsg    string
	}{
		{
			"happy path",
			properServiceKey,
			CpiServiceKey{
				Host: "https://demo",
				Uaa: OAuth{
					OAuthTokenProviderURL: "https://demo/oauth/token",
					ClientId:              "demouser",
					ClientSecret:          "******",
				},
			},
			"",
		},
		{
			"faulty json",
			faultyServiceKey,
			CpiServiceKey{},
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
