package whitesource

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		// assert
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
			ProjectName:      "mock-project",
			AgentDownloadURL: "https://download.ua.org/agent.jar",
			AgentFileName:    "unified-agent.jar",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.AddFile("wss-generated-file.config", []byte("dummy"))
		scan := newTestScan(&config)
		// test
		err := scan.ExecuteUAScan(&config, utilsMock)
		// assert
		require.NoError(t, err)
		require.Len(t, utilsMock.DownloadedFiles, 1)
		assert.Equal(t, "https://download.ua.org/agent.jar", utilsMock.DownloadedFiles[0].SourceURL)
		assert.Equal(t, "unified-agent.jar", utilsMock.DownloadedFiles[0].FilePath)
	})
	t.Run("UA is NOT downloaded", func(t *testing.T) {
		// init
		config := ScanOptions{
			ScanType:         "unified-agent",
			ProjectName:      "mock-project",
			AgentDownloadURL: "https://download.ua.org/agent.jar",
			AgentFileName:    "unified-agent.jar",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.AddFile("wss-generated-file.config", []byte("dummy"))
		utilsMock.AddFile("unified-agent.jar", []byte("dummy"))
		scan := newTestScan(&config)
		// test
		err := scan.ExecuteUAScan(&config, utilsMock)
		// assert
		require.NoError(t, err)
		assert.Len(t, utilsMock.DownloadedFiles, 0)
	})
}

func TestDownloadAgent(t *testing.T) {
	t.Parallel()

	t.Run("success - download", func(t *testing.T) {
		config := ScanOptions{
			AgentDownloadURL: "https://download.ua.org/agent.jar",
			AgentFileName:    "unified-agent.jar",
		}
		utilsMock := NewScanUtilsMock()

		err := downloadAgent(&config, utilsMock)
		assert.NoError(t, err, "error occured although none expected")
		assert.Len(t, utilsMock.DownloadedFiles, 1)
		assert.Equal(t, "https://download.ua.org/agent.jar", utilsMock.DownloadedFiles[0].SourceURL)
		assert.Equal(t, "unified-agent.jar", utilsMock.DownloadedFiles[0].FilePath)
	})

	t.Run("success - no download", func(t *testing.T) {
		config := ScanOptions{
			AgentDownloadURL: "https://download.ua.org/agent.jar",
			AgentFileName:    "unified-agent.jar",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.AddFile("unified-agent.jar", []byte("dummy"))

		err := downloadAgent(&config, utilsMock)
		assert.NoError(t, err, "error occured although none expected")
		assert.Len(t, utilsMock.DownloadedFiles, 0)
	})

	t.Run("error - file existence", func(t *testing.T) {
		config := ScanOptions{
			AgentDownloadURL: "https://download.ua.org/agent.jar",
			AgentFileName:    "unified-agent.jar",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.FileExistsErrorMessage = "failed to check existence"

		err := downloadAgent(&config, utilsMock)
		assert.Contains(t, fmt.Sprint(err), "failed to check if file 'unified-agent.jar' exists")
	})

	t.Run("error - download", func(t *testing.T) {
		config := ScanOptions{
			AgentDownloadURL: "https://download.ua.org/agent.jar",
			AgentFileName:    "unified-agent.jar",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.DownloadErrorMessage = "failed to download file"

		err := downloadAgent(&config, utilsMock)
		assert.Contains(t, fmt.Sprint(err), "failed to download unified agent from URL")
	})
}

func TestDownloadJre(t *testing.T) {
	t.Parallel()

	t.Run("success - no download required", func(t *testing.T) {
		config := ScanOptions{
			JreDownloadURL: "https://download.jre.org/jvm.jar",
		}
		utilsMock := NewScanUtilsMock()

		jre, err := downloadJre(&config, utilsMock)
		assert.NoError(t, err)
		assert.Equal(t, "java", jre)
		assert.Equal(t, "java", utilsMock.Calls[0].Exec)
		assert.Equal(t, []string{"--version"}, utilsMock.Calls[0].Params)
	})

	t.Run("success - jre downloaded", func(t *testing.T) {
		config := ScanOptions{
			JreDownloadURL: "https://download.jre.org/jvm.jar",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.ShouldFailOnCommand = map[string]error{"java": fmt.Errorf("failed to run java")}

		jre, err := downloadJre(&config, utilsMock)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(jvmDir, "bin", "java"), jre)
		assert.Equal(t, "https://download.jre.org/jvm.jar", utilsMock.DownloadedFiles[0].SourceURL)
		exists, _ := utilsMock.DirExists(jvmDir)
		assert.True(t, exists)
		assert.Equal(t, "tar", utilsMock.Calls[1].Exec)
		assert.Equal(t, fmt.Sprintf("--directory=%v", jvmDir), utilsMock.Calls[1].Params[0])
	})

	t.Run("error - download", func(t *testing.T) {
		config := ScanOptions{
			JreDownloadURL: "https://download.jre.org/jvm.jar",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.ShouldFailOnCommand = map[string]error{"java": fmt.Errorf("failed to run java")}
		utilsMock.DownloadErrorMessage = "failed to download file"

		_, err := downloadJre(&config, utilsMock)
		assert.Contains(t, fmt.Sprint(err), "failed to download jre from URL")
	})

	t.Run("error - tar execution", func(t *testing.T) {
		config := ScanOptions{
			JreDownloadURL: "https://download.jre.org/jvm.jar",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.ShouldFailOnCommand = map[string]error{
			"java": fmt.Errorf("failed to run java"),
			"tar":  fmt.Errorf("failed to run tar"),
		}
		_, err := downloadJre(&config, utilsMock)
		assert.Contains(t, fmt.Sprint(err), "failed to extract")
	})
}

func TestRemoveJre(t *testing.T) {
	t.Parallel()

	t.Run("success - no removal required", func(t *testing.T) {
		utilsMock := NewScanUtilsMock()
		err := removeJre("java", utilsMock)
		assert.NoError(t, err, "error occured although none expected")
	})

	t.Run("success - with removal", func(t *testing.T) {
		utilsMock := NewScanUtilsMock()
		utilsMock.AddFile(jvmTarGz, []byte("dummy"))
		err := removeJre("./jvm/bin/java", utilsMock)
		assert.NoError(t, err, "error occured although none expected")
		assert.Contains(t, utilsMock.RemoveAllDirs, jvmDir)
		assert.True(t, utilsMock.HasRemovedFile(jvmTarGz))
	})

	t.Run("error - remove jvm directory", func(t *testing.T) {
		utilsMock := NewScanUtilsMock()
		utilsMock.RemoveAllErrorMessage = "failed to remove directory"

		err := removeJre("./jvm/bin/java", utilsMock)
		assert.Contains(t, fmt.Sprint(err), "failed to remove downloaded and extracted jvm")
	})

	t.Run("error - remove jvm tar.gz", func(t *testing.T) {
		utilsMock := NewScanUtilsMock()

		err := removeJre("./jvm/bin/java", utilsMock)
		assert.Contains(t, fmt.Sprint(err), "failed to remove downloaded")
	})
}

func TestAutoGenerateWhitesourceConfig(t *testing.T) {
	t.Parallel()

}
