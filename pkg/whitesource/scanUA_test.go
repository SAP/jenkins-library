package whitesource

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExecuteScanUA(t *testing.T) {
	t.Parallel()
	t.Run("happy path UA", func(t *testing.T) {
		// init
		config := ScanOptions{
			ScanType:         "unified-agent",
			OrgToken:         "org-token",
			UserToken:        "user-token",
			ProductName:      "mock-product",
			ProjectName:      "mock-project",
			AgentDownloadURL: "https://download.ua.org/agent.jar",
			AgentFileName:    "unified-agent.jar",
			ConfigFilePath:   "ua.cfg",
			M2Path:           ".pipeline/m2",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.AddFile("wss-generated-file.config", []byte("key=value"))
		scan := newTestScan(&config)
		// test
		err := scan.ExecuteUAScan(&config, utilsMock)
		// many assert
		require.NoError(t, err)

		content, err := utilsMock.FileRead("ua.cfg")
		require.NoError(t, err)
		contentAsString := string(content)
		assert.Contains(t, contentAsString, "key=value\n")
		assert.Contains(t, contentAsString, "gradle.aggregateModules=true\n")
		assert.Contains(t, contentAsString, "maven.aggregateModules=true\n")
		assert.Contains(t, contentAsString, "maven.m2RepositoryPath=.pipeline/m2\n")
		assert.Contains(t, contentAsString, "excludes=")

		require.Len(t, utilsMock.Calls, 4)
		fmt.Printf("calls: %v\n", utilsMock.Calls)
		expectedCall := mock.ExecCall{
			Exec: "java",
			Params: []string{
				"-jar",
				config.AgentFileName,
				"-d", ".",
				"-c", config.ConfigFilePath,
				"-apiKey", config.OrgToken,
				"-userKey", config.UserToken,
				"-project", config.ProjectName,
				"-product", config.ProductName,
				"-productVersion", scan.ProductVersion,
			},
		}
		assert.Equal(t, expectedCall, utilsMock.Calls[3])
	})
	t.Run("UA is downloaded", func(t *testing.T) {
		// init
		config := ScanOptions{
			ScanType:         "unified-agent",
			AgentDownloadURL: "https://download.ua.org/agent.jar",
			AgentFileName:    "unified-agent.jar",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.AddFile("wss-generated-file.config", []byte("dummy"))
		scan := newTestScan(&config)
		// test
		err := scan.ExecuteUAScan(&config, utilsMock)
		// many assert
		require.NoError(t, err)
		require.Len(t, utilsMock.DownloadedFiles, 1)
		assert.Equal(t, "https://download.ua.org/agent.jar", utilsMock.DownloadedFiles[0].sourceURL)
		assert.Equal(t, "unified-agent.jar", utilsMock.DownloadedFiles[0].filePath)
	})
	t.Run("UA is NOT downloaded", func(t *testing.T) {
		// init
		config := ScanOptions{
			ScanType:         "unified-agent",
			AgentDownloadURL: "https://download.ua.org/agent.jar",
			AgentFileName:    "unified-agent.jar",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.AddFile("wss-generated-file.config", []byte("dummy"))
		utilsMock.AddFile("unified-agent.jar", []byte("dummy"))
		scan := newTestScan(&config)
		// test
		err := scan.ExecuteUAScan(&config, utilsMock)
		// many assert
		require.NoError(t, err)
		assert.Len(t, utilsMock.DownloadedFiles, 0)
	})
}
