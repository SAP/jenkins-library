package apim

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsJSONFile(t *testing.T) {
	properServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
				}
			}`
	apimData := Bundle{Payload: properServiceKey}
	assert.Equal(t, apimData.IsJSON(), true)
}

func TestIsJSONInvalidFile(t *testing.T) {
	properServiceKey := `{
			"oauth" {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
				}
			}`
	apimData := Bundle{Payload: properServiceKey}
	assert.Equal(t, apimData.IsJSON(), false)
}
