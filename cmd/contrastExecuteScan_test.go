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
