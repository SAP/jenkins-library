package cmd

import (
	"encoding/base64"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type contrastExecuteScanMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newContrastExecuteScanTestsUtils() contrastExecuteScanMockUtils {
	utils := contrastExecuteScanMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestGetAuth(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		config := &contrastExecuteScanOptions{
			UserAPIKey: "user-api-key",
			Username:   "username",
			ServiceKey: "service-key",
		}
		authString := getAuth(config)
		assert.NotEmpty(t, authString)
		data, err := base64.StdEncoding.DecodeString(authString)
		assert.NoError(t, err)
		assert.Equal(t, "username:service-key", string(data))
	})
}

func TestGetApplicationUrls(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		config := &contrastExecuteScanOptions{
			Server:         "https://server.com",
			OrganizationID: "orgId",
			ApplicationID:  "appId",
		}
		appUrl, guiUrl := getApplicationUrls(config)
		assert.Equal(t, "https://server.com/api/v4/organizations/orgId/applications/appId", appUrl)
		assert.Equal(t, "https://server.com/Contrast/static/ng/index.html#/orgId/applications/appId", guiUrl)
	})
}

func TestValidateConfigs(t *testing.T) {
	t.Parallel()
	validConfig := contrastExecuteScanOptions{
		UserAPIKey:     "user-api-key",
		ServiceKey:     "service-key",
		Username:       "username",
		Server:         "https://server.com",
		OrganizationID: "orgId",
		ApplicationID:  "appId",
	}

	t.Run("Valid config", func(t *testing.T) {
		config := validConfig
		err := validateConfigs(&config)
		assert.NoError(t, err)
	})

	t.Run("Valid config, server url without https://", func(t *testing.T) {
		config := validConfig
		config.Server = "server.com"
		err := validateConfigs(&config)
		assert.NoError(t, err)
		assert.Equal(t, config.Server, "https://server.com")
	})

	t.Run("Empty config", func(t *testing.T) {
		config := contrastExecuteScanOptions{}

		err := validateConfigs(&config)
		assert.Error(t, err)
	})

	t.Run("Empty userAPIKey", func(t *testing.T) {
		config := validConfig
		config.UserAPIKey = ""

		err := validateConfigs(&config)
		assert.Error(t, err)
	})

	t.Run("Empty username", func(t *testing.T) {
		config := validConfig
		config.Username = ""

		err := validateConfigs(&config)
		assert.Error(t, err)
	})

	t.Run("Empty serviceKey", func(t *testing.T) {
		config := validConfig
		config.ServiceKey = ""

		err := validateConfigs(&config)
		assert.Error(t, err)
	})

	t.Run("Empty server", func(t *testing.T) {
		config := validConfig
		config.Server = ""

		err := validateConfigs(&config)
		assert.Error(t, err)
	})

	t.Run("Empty organizationId", func(t *testing.T) {
		config := validConfig
		config.OrganizationID = ""

		err := validateConfigs(&config)
		assert.Error(t, err)
	})

	t.Run("Empty applicationID", func(t *testing.T) {
		config := validConfig
		config.ApplicationID = ""

		err := validateConfigs(&config)
		assert.Error(t, err)
	})
}
