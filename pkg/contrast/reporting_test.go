package contrast

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type contrastExecuteScanMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newContrastExecuteScanTestsUtils() contrastExecuteScanMockUtils {
	return contrastExecuteScanMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
}

func TestCreateToolRecordContrast(t *testing.T) {
	modulePath := "./"

	t.Run("Valid toolrun file", func(t *testing.T) {
		appInfo := &ApplicationInfo{
			Url:    "https://server.com/application",
			Id:     "application-id",
			Name:   "app name",
			Server: "https://server.com",
		}
		toolRecord, err := createToolRecordContrast(newContrastExecuteScanTestsUtils(), appInfo, modulePath)
		assert.NoError(t, err)
		assert.Equal(t, "contrast", toolRecord.ToolName)
		assert.Equal(t, appInfo.Server, toolRecord.ToolInstance)
		assert.Equal(t, appInfo.Name, toolRecord.DisplayName)
		assert.Equal(t, appInfo.Url, toolRecord.DisplayURL)
		assert.Equal(t, 1, len(toolRecord.Keys))
		assert.Equal(t, "application", toolRecord.Keys[0].Name)
		assert.Equal(t, appInfo.Url, toolRecord.Keys[0].URL)
		assert.Equal(t, appInfo.Id, toolRecord.Keys[0].Value)
		assert.Equal(t, appInfo.Name, toolRecord.Keys[0].DisplayName)
	})

	t.Run("Empty server", func(t *testing.T) {
		appInfo := &ApplicationInfo{
			Url:  "https://server.com/application",
			Id:   "application-id",
			Name: "app name",
		}
		toolRecord, err := createToolRecordContrast(newContrastExecuteScanTestsUtils(), appInfo, modulePath)
		assert.NoError(t, err)
		assert.Equal(t, "contrast", toolRecord.ToolName)
		assert.Equal(t, "", toolRecord.ToolInstance)
		assert.Equal(t, appInfo.Name, toolRecord.DisplayName)
		assert.Equal(t, appInfo.Url, toolRecord.DisplayURL)
		assert.Equal(t, 1, len(toolRecord.Keys))
		assert.Equal(t, "application", toolRecord.Keys[0].Name)
		assert.Equal(t, appInfo.Url, toolRecord.Keys[0].URL)
		assert.Equal(t, appInfo.Id, toolRecord.Keys[0].Value)
		assert.Equal(t, appInfo.Name, toolRecord.Keys[0].DisplayName)
	})

	t.Run("Empty application id", func(t *testing.T) {
		appInfo := &ApplicationInfo{
			Url:    "https://server.com/application",
			Name:   "app name",
			Server: "https://server.com",
		}
		_, err := createToolRecordContrast(newContrastExecuteScanTestsUtils(), appInfo, modulePath)
		assert.Error(t, err)
	})

	t.Run("Empty application name", func(t *testing.T) {
		appInfo := &ApplicationInfo{
			Url:    "https://contrastsecurity.com",
			Id:     "application-id",
			Server: "https://server.com",
		}
		toolRecord, err := createToolRecordContrast(newContrastExecuteScanTestsUtils(), appInfo, modulePath)
		assert.NoError(t, err)
		assert.Equal(t, "contrast", toolRecord.ToolName)
		assert.Equal(t, appInfo.Server, toolRecord.ToolInstance)
		assert.Equal(t, "", toolRecord.DisplayName)
		assert.Equal(t, appInfo.Url, toolRecord.DisplayURL)
		assert.Equal(t, 1, len(toolRecord.Keys))
		assert.Equal(t, "application", toolRecord.Keys[0].Name)
		assert.Equal(t, appInfo.Url, toolRecord.Keys[0].URL)
		assert.Equal(t, appInfo.Id, toolRecord.Keys[0].Value)
		assert.Equal(t, "", toolRecord.Keys[0].DisplayName)
	})

	t.Run("Empty application url", func(t *testing.T) {
		appInfo := &ApplicationInfo{
			Name:   "app name",
			Id:     "application-id",
			Server: "https://server.com",
		}
		toolRecord, err := createToolRecordContrast(newContrastExecuteScanTestsUtils(), appInfo, modulePath)
		assert.NoError(t, err)
		assert.Equal(t, "contrast", toolRecord.ToolName)
		assert.Equal(t, appInfo.Server, toolRecord.ToolInstance)
		assert.Equal(t, appInfo.Name, toolRecord.DisplayName)
		assert.Equal(t, "", toolRecord.DisplayURL)
		assert.Equal(t, 1, len(toolRecord.Keys))
		assert.Equal(t, "application", toolRecord.Keys[0].Name)
		assert.Equal(t, "", toolRecord.Keys[0].URL)
		assert.Equal(t, appInfo.Id, toolRecord.Keys[0].Value)
		assert.Equal(t, appInfo.Name, toolRecord.Keys[0].DisplayName)
	})
}
