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

func TestSaveReportFileSuccess(t *testing.T) {
	t.Run("SaveSarifReport", func(t *testing.T) {
		utils := newContrastExecuteScanTestsUtils()
		testData := []byte(`{"version": "2.1.0"}`)

		paths, err := SaveReportFile(utils, "piper_contrast.sarif", "Contrast SARIF Report", testData)

		assert.NoError(t, err)
		assert.NotEmpty(t, paths)
		assert.Equal(t, 1, len(paths))
		assert.Equal(t, "Contrast SARIF Report", paths[0].Name)
		assert.Equal(t, "contrast/piper_contrast.sarif", paths[0].Target)
	})

	t.Run("SavePdfReport", func(t *testing.T) {
		utils := newContrastExecuteScanTestsUtils()
		testData := []byte("PDF content here")

		paths, err := SaveReportFile(utils, "piper_contrast_attestation.pdf", "Contrast PDF Attestation Report", testData)

		assert.NoError(t, err)
		assert.NotEmpty(t, paths)
		assert.Equal(t, 1, len(paths))
		assert.Equal(t, "Contrast PDF Attestation Report", paths[0].Name)
		assert.Equal(t, "contrast/piper_contrast_attestation.pdf", paths[0].Target)
	})

	t.Run("SaveJsonReport", func(t *testing.T) {
		utils := newContrastExecuteScanTestsUtils()
		testData := []byte(`{"toolName":"contrast"}`)

		paths, err := SaveReportFile(utils, "piper_contrast_report.json", "Contrast JSON Compliance Report", testData)

		assert.NoError(t, err)
		assert.NotEmpty(t, paths)
		assert.Equal(t, 1, len(paths))
		assert.Equal(t, "Contrast JSON Compliance Report", paths[0].Name)
		assert.Equal(t, "contrast/piper_contrast_report.json", paths[0].Target)
	})
}

func TestSaveReportFileFileWriteFails(t *testing.T) {
	utils := newContrastExecuteScanTestsUtils()
	utils.FileWriteError = assert.AnError

	paths, err := SaveReportFile(utils, "test.txt", "Test File", []byte("test"))

	assert.Error(t, err)
	assert.Nil(t, paths)
	assert.Contains(t, err.Error(), "failed to write test.txt file")
}
