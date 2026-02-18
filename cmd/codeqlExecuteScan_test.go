//go:build unit

package cmd

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SAP/jenkins-library/pkg/codeql"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

type codeqlExecuteScanMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
	*mock.HttpClientMock
}

func newCodeqlExecuteScanTestsUtils() codeqlExecuteScanMockUtils {
	utils := codeqlExecuteScanMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{
			Stub: func(call string, stdoutReturn map[string]string, shouldFailOnCommand map[string]error, stdout io.Writer) error {
				return nil
			},
		},
		FilesMock:      &mock.FilesMock{},
		HttpClientMock: &mock.HttpClientMock{},
	}
	return utils
}

func TestGetMavenSettings(t *testing.T) {
	t.Parallel()
	t.Run("No maven", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "npm"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		assert.Equal(t, "", params)
	})

	t.Run("No build command", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven"}
		params := getMavenSettings("", &config, newCodeqlExecuteScanTestsUtils())
		assert.Equal(t, "", params)
	})

	t.Run("Project Settings file", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", ProjectSettingsFile: "test.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		dir, _ := os.Getwd()
		projectSettingsPath := filepath.Join(dir, "test.xml")
		expectedCommand := fmt.Sprintf(" \"--settings=%s\"", projectSettingsPath)
		assert.Equal(t, expectedCommand, params)
	})

	t.Run("Skip Project Settings file in case already used", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", ProjectSettingsFile: "test.xml"}
		buildCmd := "mvn clean install --settings=project.xml"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		assert.Equal(t, "", params)
	})

	t.Run("Global Settings file", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", GlobalSettingsFile: "global.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		dir, _ := os.Getwd()
		globalSettingsPath := filepath.Join(dir, "global.xml")
		expectedCommand := fmt.Sprintf(" \"--global-settings=%s\"", globalSettingsPath)
		assert.Equal(t, expectedCommand, params)
	})

	t.Run("Project and Global Settings file", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", ProjectSettingsFile: "test.xml", GlobalSettingsFile: "global.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		dir, _ := os.Getwd()
		globalSettingsPath := filepath.Join(dir, "global.xml")
		projectSettingsPath := filepath.Join(dir, "test.xml")
		expectedCommand := fmt.Sprintf(" \"--global-settings=%s\" \"--settings=%s\"", globalSettingsPath, projectSettingsPath)
		assert.Equal(t, expectedCommand, params)
	})

	t.Run("ProjectSettingsFile https url", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", ProjectSettingsFile: "https://jenkins-sap-test.com/test.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		dir, _ := os.Getwd()
		projectSettingsPath := filepath.Join(dir, ".pipeline/mavenProjectSettings.xml")
		expectedCommand := fmt.Sprintf(" \"--settings=%s\"", projectSettingsPath)
		assert.Equal(t, expectedCommand, params)
	})

	t.Run("ProjectSettingsFile http url", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", ProjectSettingsFile: "http://jenkins-sap-test.com/test.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		dir, _ := os.Getwd()
		projectSettingsPath := filepath.Join(dir, ".pipeline/mavenProjectSettings.xml")
		expectedCommand := fmt.Sprintf(" \"--settings=%s\"", projectSettingsPath)
		assert.Equal(t, expectedCommand, params)
	})

	t.Run("GlobalSettingsFile https url", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", GlobalSettingsFile: "https://jenkins-sap-test.com/test.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		dir, _ := os.Getwd()
		globalSettingsPath := filepath.Join(dir, ".pipeline/mavenGlobalSettings.xml")
		expectedCommand := fmt.Sprintf(" \"--global-settings=%s\"", globalSettingsPath)
		assert.Equal(t, expectedCommand, params)
	})

	t.Run("GlobalSettingsFile http url", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", GlobalSettingsFile: "http://jenkins-sap-test.com/test.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		dir, _ := os.Getwd()
		globalSettingsPath := filepath.Join(dir, ".pipeline/mavenGlobalSettings.xml")
		expectedCommand := fmt.Sprintf(" \"--global-settings=%s\"", globalSettingsPath)
		assert.Equal(t, expectedCommand, params)
	})

	t.Run("ProjectSettingsFile and GlobalSettingsFile https url", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", GlobalSettingsFile: "https://jenkins-sap-test.com/test.xml", ProjectSettingsFile: "http://jenkins-sap-test.com/test.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		dir, _ := os.Getwd()
		globalSettingsPath := filepath.Join(dir, ".pipeline/mavenGlobalSettings.xml")
		projectSettingsPath := filepath.Join(dir, ".pipeline/mavenProjectSettings.xml")
		expectedCommand := fmt.Sprintf(" \"--global-settings=%s\" \"--settings=%s\"", globalSettingsPath, projectSettingsPath)
		assert.Equal(t, expectedCommand, params)
	})

	t.Run("ProjectSettingsFile and GlobalSettingsFile http url", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", GlobalSettingsFile: "http://jenkins-sap-test.com/test.xml", ProjectSettingsFile: "http://jenkins-sap-test.com/test.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		dir, _ := os.Getwd()
		globalSettingsPath := filepath.Join(dir, ".pipeline/mavenGlobalSettings.xml")
		projectSettingsPath := filepath.Join(dir, ".pipeline/mavenProjectSettings.xml")
		expectedCommand := fmt.Sprintf(" \"--global-settings=%s\" \"--settings=%s\"", globalSettingsPath, projectSettingsPath)
		assert.Equal(t, expectedCommand, params)
	})

	t.Run("ProjectSettingsFile file and GlobalSettingsFile https url", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", GlobalSettingsFile: "https://jenkins-sap-test.com/test.xml", ProjectSettingsFile: "test.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		dir, _ := os.Getwd()
		globalSettingsPath := filepath.Join(dir, ".pipeline/mavenGlobalSettings.xml")
		projectSettingsPath := filepath.Join(dir, "test.xml")
		expectedCommand := fmt.Sprintf(" \"--global-settings=%s\" \"--settings=%s\"", globalSettingsPath, projectSettingsPath)
		assert.Equal(t, expectedCommand, params)
	})

	t.Run("ProjectSettingsFile file and GlobalSettingsFile https url", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", GlobalSettingsFile: "http://jenkins-sap-test.com/test.xml", ProjectSettingsFile: "test.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		dir, _ := os.Getwd()
		globalSettingsPath := filepath.Join(dir, ".pipeline/mavenGlobalSettings.xml")
		projectSettingsPath := filepath.Join(dir, "test.xml")
		expectedCommand := fmt.Sprintf(" \"--global-settings=%s\" \"--settings=%s\"", globalSettingsPath, projectSettingsPath)
		assert.Equal(t, expectedCommand, params)
	})

	t.Run("ProjectSettingsFile https url and GlobalSettingsFile file", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", GlobalSettingsFile: "global.xml", ProjectSettingsFile: "http://jenkins-sap-test.com/test.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		dir, _ := os.Getwd()
		globalSettingsPath := filepath.Join(dir, "global.xml")
		projectSettingsPath := filepath.Join(dir, ".pipeline/mavenProjectSettings.xml")
		expectedCommand := fmt.Sprintf(" \"--global-settings=%s\" \"--settings=%s\"", globalSettingsPath, projectSettingsPath)
		assert.Equal(t, expectedCommand, params)
	})

	t.Run("ProjectSettingsFile http url and GlobalSettingsFile file", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", GlobalSettingsFile: "global.xml", ProjectSettingsFile: "http://jenkins-sap-test.com/test.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		dir, _ := os.Getwd()
		globalSettingsPath := filepath.Join(dir, "global.xml")
		projectSettingsPath := filepath.Join(dir, ".pipeline/mavenProjectSettings.xml")
		expectedCommand := fmt.Sprintf(" \"--global-settings=%s\" \"--settings=%s\"", globalSettingsPath, projectSettingsPath)
		assert.Equal(t, expectedCommand, params)
	})
}

func TestUpdateCmdFlag(t *testing.T) {
	t.Parallel()

	t.Run("No maven", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "npm"}
		customFlags := map[string]string{
			"--command": "mvn clean install",
		}
		updateCmdFlag(&config, customFlags, newCodeqlExecuteScanTestsUtils())
		assert.Equal(t, "mvn clean install", customFlags["--command"])
		assert.Equal(t, "", customFlags["-c"])
	})

	t.Run("No custom flags with build command provided", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", ProjectSettingsFile: "test.xml", GlobalSettingsFile: "global.xml"}
		customFlags := map[string]string{}
		updateCmdFlag(&config, customFlags, newCodeqlExecuteScanTestsUtils())
		assert.Equal(t, "", customFlags["--command"])
		assert.Equal(t, "", customFlags["-c"])
	})

	t.Run("Both --command and -c flags are set, no settings file provided", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven"}
		customFlags := map[string]string{
			"--command": "mvn clean install",
			"-c":        "mvn clean install -DskipTests",
		}
		updateCmdFlag(&config, customFlags, newCodeqlExecuteScanTestsUtils())
		assert.Equal(t, "mvn clean install", customFlags["--command"])
		assert.Equal(t, "", customFlags["-c"])
	})

	t.Run("Only -c flag is set, no settings file provided", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven"}
		customFlags := map[string]string{
			"-c": "mvn clean install -DskipTests",
		}
		updateCmdFlag(&config, customFlags, newCodeqlExecuteScanTestsUtils())
		assert.Equal(t, "mvn clean install -DskipTests", customFlags["--command"])
		assert.Equal(t, "", customFlags["-c"])
	})

	t.Run("Update custom command with GlobalSettingsFile and ProjectSettingsFile", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", ProjectSettingsFile: "test.xml", GlobalSettingsFile: "global.xml"}
		customFlags := map[string]string{
			"--command": "mvn clean install",
		}
		updateCmdFlag(&config, customFlags, newCodeqlExecuteScanTestsUtils())
		dir, _ := os.Getwd()
		globalSettingsPath := filepath.Join(dir, "global.xml")
		projectSettingsPath := filepath.Join(dir, "test.xml")
		expectedCommand := fmt.Sprintf("mvn clean install \"--global-settings=%s\" \"--settings=%s\"", globalSettingsPath, projectSettingsPath)
		assert.Equal(t, expectedCommand, customFlags["--command"])
		assert.Equal(t, "", customFlags["-c"])
	})

	t.Run("Custom command has --global-settings and --settings", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", ProjectSettingsFile: "test.xml", GlobalSettingsFile: "global.xml"}
		customFlags := map[string]string{
			"--command": "mvn clean install --settings=test1.xml --global-settings=global1.xml",
		}
		updateCmdFlag(&config, customFlags, newCodeqlExecuteScanTestsUtils())
		assert.Equal(t, "mvn clean install --settings=test1.xml --global-settings=global1.xml", customFlags["--command"])
		assert.Equal(t, "", customFlags["-c"])
	})
}

func TestAddDataToInfluxDB(t *testing.T) {
	repoUrl := "https://github.htllo.test/Testing/codeql"
	repoRef := "https://github.htllo.test/Testing/codeql/tree/branch"
	repoScanUrl := "https://github.htllo.test/Testing/codeql/security/code-scanning"
	querySuite := "security.ql"

	repoInfo := &codeql.RepoInfo{
		FullUrl: repoUrl,
		FullRef: repoRef,
		ScanUrl: repoScanUrl,
	}

	t.Run("No findings", func(t *testing.T) {
		scanResults := []codeql.CodeqlFindings{}
		influx := &codeqlExecuteScanInflux{}
		addDataToInfluxDB(repoInfo, querySuite, scanResults, influx)
		assert.Equal(t, repoUrl, influx.codeql_data.fields.repositoryURL)
		assert.Equal(t, repoRef, influx.codeql_data.fields.repositoryReferenceURL)
		assert.Equal(t, repoScanUrl, influx.codeql_data.fields.codeScanningLink)
		assert.Equal(t, querySuite, influx.codeql_data.fields.querySuite)
		assert.Equal(t, 0, influx.codeql_data.fields.auditAllTotal)
		assert.Equal(t, 0, influx.codeql_data.fields.auditAllAudited)
		assert.Equal(t, 0, influx.codeql_data.fields.optionalTotal)
		assert.Equal(t, 0, influx.codeql_data.fields.optionalAudited)
	})

	t.Run("Audit All findings category only", func(t *testing.T) {
		scanResults := []codeql.CodeqlFindings{
			{
				ClassificationName: codeql.AuditAll,
				Total:              100,
				Audited:            50,
			},
		}
		influx := &codeqlExecuteScanInflux{}
		addDataToInfluxDB(repoInfo, querySuite, scanResults, influx)
		assert.Equal(t, repoUrl, influx.codeql_data.fields.repositoryURL)
		assert.Equal(t, repoRef, influx.codeql_data.fields.repositoryReferenceURL)
		assert.Equal(t, repoScanUrl, influx.codeql_data.fields.codeScanningLink)
		assert.Equal(t, querySuite, influx.codeql_data.fields.querySuite)
		assert.Equal(t, scanResults[0].Total, influx.codeql_data.fields.auditAllTotal)
		assert.Equal(t, scanResults[0].Audited, influx.codeql_data.fields.auditAllAudited)
		assert.Equal(t, 0, influx.codeql_data.fields.optionalTotal)
		assert.Equal(t, 0, influx.codeql_data.fields.optionalAudited)
	})

	t.Run("Optional findings category only", func(t *testing.T) {
		scanResults := []codeql.CodeqlFindings{
			{
				ClassificationName: codeql.Optional,
				Total:              100,
				Audited:            50,
			},
		}
		influx := &codeqlExecuteScanInflux{}
		addDataToInfluxDB(repoInfo, querySuite, scanResults, influx)
		assert.Equal(t, repoUrl, influx.codeql_data.fields.repositoryURL)
		assert.Equal(t, repoRef, influx.codeql_data.fields.repositoryReferenceURL)
		assert.Equal(t, repoScanUrl, influx.codeql_data.fields.codeScanningLink)
		assert.Equal(t, querySuite, influx.codeql_data.fields.querySuite)
		assert.Equal(t, 0, influx.codeql_data.fields.auditAllTotal)
		assert.Equal(t, 0, influx.codeql_data.fields.auditAllAudited)
		assert.Equal(t, scanResults[0].Total, influx.codeql_data.fields.optionalTotal)
		assert.Equal(t, scanResults[0].Audited, influx.codeql_data.fields.optionalAudited)
	})

	t.Run("Both findings category", func(t *testing.T) {
		scanResults := []codeql.CodeqlFindings{
			{
				ClassificationName: codeql.AuditAll,
				Total:              100,
				Audited:            50,
			},
			{
				ClassificationName: codeql.Optional,
				Total:              100,
				Audited:            50,
			},
		}
		influx := &codeqlExecuteScanInflux{}
		addDataToInfluxDB(repoInfo, querySuite, scanResults, influx)
		assert.Equal(t, repoUrl, influx.codeql_data.fields.repositoryURL)
		assert.Equal(t, repoRef, influx.codeql_data.fields.repositoryReferenceURL)
		assert.Equal(t, repoScanUrl, influx.codeql_data.fields.codeScanningLink)
		assert.Equal(t, querySuite, influx.codeql_data.fields.querySuite)
		assert.Equal(t, scanResults[0].Total, influx.codeql_data.fields.auditAllTotal)
		assert.Equal(t, scanResults[0].Audited, influx.codeql_data.fields.auditAllAudited)
		assert.Equal(t, scanResults[1].Total, influx.codeql_data.fields.optionalTotal)
		assert.Equal(t, scanResults[1].Audited, influx.codeql_data.fields.optionalAudited)
	})
}

func TestPrepareCmdForDatabaseCreate(t *testing.T) {
	t.Parallel()

	t.Run("No custom flags", func(t *testing.T) {
		config := &codeqlExecuteScanOptions{
			Database:     "codeqlDB",
			ModulePath:   "./",
			BuildTool:    "maven",
			BuildCommand: "mvn clean install",
		}
		isMultiLang, cmd, err := prepareCmdForDatabaseCreate(map[string]string{}, config, newCodeqlExecuteScanTestsUtils())
		assert.NoError(t, err)
		assert.NotEmpty(t, cmd)
		assert.False(t, isMultiLang)
		assert.Equal(t, 10, len(cmd))
		assert.Equal(t, "database create codeqlDB --overwrite --source-root . --working-dir ./ --language=java --command=mvn clean install",
			strings.Join(cmd, " "))
	})

	t.Run("No custom flags, custom build tool", func(t *testing.T) {
		config := &codeqlExecuteScanOptions{
			Database:   "codeqlDB",
			ModulePath: "./",
			BuildTool:  "custom",
			Language:   "javascript",
		}
		isMultiLang, cmd, err := prepareCmdForDatabaseCreate(map[string]string{}, config, newCodeqlExecuteScanTestsUtils())
		assert.NoError(t, err)
		assert.NotEmpty(t, cmd)
		assert.False(t, isMultiLang)
		assert.Equal(t, 9, len(cmd))
		assert.Equal(t, "database create codeqlDB --overwrite --source-root . --working-dir ./ --language=javascript",
			strings.Join(cmd, " "))
	})

	t.Run("No custom flags, custom build tool, no language specified", func(t *testing.T) {
		config := &codeqlExecuteScanOptions{
			Database:   "codeqlDB",
			ModulePath: "./",
			BuildTool:  "custom",
		}
		_, _, err := prepareCmdForDatabaseCreate(map[string]string{}, config, newCodeqlExecuteScanTestsUtils())
		assert.Error(t, err)
	})

	t.Run("No custom flags, invalid build tool, no language specified", func(t *testing.T) {
		config := &codeqlExecuteScanOptions{
			Database:   "codeqlDB",
			ModulePath: "./",
			BuildTool:  "test",
		}
		_, _, err := prepareCmdForDatabaseCreate(map[string]string{}, config, newCodeqlExecuteScanTestsUtils())
		assert.Error(t, err)
	})

	t.Run("No custom flags, invalid build tool, language specified", func(t *testing.T) {
		config := &codeqlExecuteScanOptions{
			Database:   "codeqlDB",
			ModulePath: "./",
			BuildTool:  "test",
			Language:   "javascript",
		}
		isMultiLang, cmd, err := prepareCmdForDatabaseCreate(map[string]string{}, config, newCodeqlExecuteScanTestsUtils())
		assert.NoError(t, err)
		assert.NotEmpty(t, cmd)
		assert.False(t, isMultiLang)
		assert.Equal(t, 9, len(cmd))
		assert.Equal(t, "database create codeqlDB --overwrite --source-root . --working-dir ./ --language=javascript",
			strings.Join(cmd, " "))
	})

	t.Run("Custom flags, overwriting source-root", func(t *testing.T) {
		config := &codeqlExecuteScanOptions{
			Database:   "codeqlDB",
			ModulePath: "./",
			Language:   "javascript",
		}
		customFlags := map[string]string{
			"--source-root": "--source-root=customSrcRoot/",
		}
		isMultiLang, cmd, err := prepareCmdForDatabaseCreate(customFlags, config, newCodeqlExecuteScanTestsUtils())
		assert.NoError(t, err)
		assert.NotEmpty(t, cmd)
		assert.False(t, isMultiLang)
		assert.Equal(t, 8, len(cmd))
		assert.Equal(t, "database create codeqlDB --overwrite --working-dir ./ --language=javascript --source-root=customSrcRoot/",
			strings.Join(cmd, " "))
	})

	t.Run("Custom flags, overwriting threads from config", func(t *testing.T) {
		config := &codeqlExecuteScanOptions{
			Database:   "codeqlDB",
			ModulePath: "./",
			Language:   "javascript",
			Threads:    "0",
			Ram:        "2000",
		}
		customFlags := map[string]string{
			"--source-root": "--source-root=customSrcRoot/",
			"-j":            "-j=1",
		}
		isMultiLang, cmd, err := prepareCmdForDatabaseCreate(customFlags, config, newCodeqlExecuteScanTestsUtils())
		assert.NoError(t, err)
		assert.NotEmpty(t, cmd)
		assert.False(t, isMultiLang)
		assert.Equal(t, 10, len(cmd))
		assert.True(t, "database create codeqlDB --overwrite --working-dir ./ --language=javascript --ram=2000 -j=1 --source-root=customSrcRoot/" == strings.Join(cmd, " ") ||
			"database create codeqlDB --overwrite --working-dir ./ --language=javascript --ram=2000 --source-root=customSrcRoot/ -j=1" == strings.Join(cmd, " "))
	})

	t.Run("Multi-language adds --db-cluster", func(t *testing.T) {
		config := &codeqlExecuteScanOptions{
			Database:   "codeqlDB",
			ModulePath: "./",
			BuildTool:  "custom",
			Language:   "javascript,python",
		}
		isMultiLang, cmd, err := prepareCmdForDatabaseCreate(map[string]string{}, config, newCodeqlExecuteScanTestsUtils())
		assert.NoError(t, err)
		assert.NotEmpty(t, cmd)
		assert.True(t, isMultiLang)
		assert.Equal(t, 10, len(cmd))
		assert.Equal(t, "database create codeqlDB --overwrite --source-root . --working-dir ./ --db-cluster --language=javascript,python",
			strings.Join(cmd, " "))
	})

	t.Run("Multi-language with build command", func(t *testing.T) {
		config := &codeqlExecuteScanOptions{
			Database:     "codeqlDB",
			ModulePath:   "./",
			BuildTool:    "custom",
			Language:     "go,python",
			BuildCommand: "make build",
		}
		isMultiLang, cmd, err := prepareCmdForDatabaseCreate(map[string]string{}, config, newCodeqlExecuteScanTestsUtils())
		assert.NoError(t, err)
		assert.NotEmpty(t, cmd)
		assert.True(t, isMultiLang)
		assert.Equal(t, 11, len(cmd))
		assert.Equal(t, "database create codeqlDB --overwrite --source-root . --working-dir ./ --db-cluster --language=go,python --command=make build",
			strings.Join(cmd, " "))
	})
}

func TestPrepareCmdForDatabaseAnalyze(t *testing.T) {
	t.Parallel()
	utils := codeqlExecuteScanMockUtils{}

	t.Run("No additional flags, no querySuite, sarif format", func(t *testing.T) {
		config := &codeqlExecuteScanOptions{
			Database: "codeqlDB",
		}
		cmd, err := prepareCmdForDatabaseAnalyze(utils, map[string]string{}, config, "sarif-latest", "target/codeqlReport.sarif", config.Database)
		assert.NoError(t, err)
		assert.NotEmpty(t, cmd)
		assert.Equal(t, 5, len(cmd))
		assert.Equal(t, "database analyze --format=sarif-latest --output=target/codeqlReport.sarif codeqlDB", strings.Join(cmd, " "))
	})

	t.Run("No additional flags, no querySuite, csv format", func(t *testing.T) {
		config := &codeqlExecuteScanOptions{
			Database: "codeqlDB",
		}
		cmd, err := prepareCmdForDatabaseAnalyze(utils, map[string]string{}, config, "csv", "target/codeqlReport.csv", config.Database)
		assert.NoError(t, err)
		assert.NotEmpty(t, cmd)
		assert.Equal(t, 5, len(cmd))
		assert.Equal(t, "database analyze --format=csv --output=target/codeqlReport.csv codeqlDB", strings.Join(cmd, " "))
	})

	t.Run("No additional flags, set querySuite", func(t *testing.T) {
		config := &codeqlExecuteScanOptions{
			Database:   "codeqlDB",
			QuerySuite: "security.ql",
		}
		cmd, err := prepareCmdForDatabaseAnalyze(utils, map[string]string{}, config, "sarif-latest", "target/codeqlReport.sarif", config.Database)
		assert.NoError(t, err)
		assert.NotEmpty(t, cmd)
		assert.Equal(t, 6, len(cmd))
		assert.Equal(t, "database analyze --format=sarif-latest --output=target/codeqlReport.sarif codeqlDB security.ql", strings.Join(cmd, " "))
	})

	t.Run("No custom flags, flags from config", func(t *testing.T) {
		config := &codeqlExecuteScanOptions{
			Database:   "codeqlDB",
			QuerySuite: "security.ql",
			Threads:    "1",
			Ram:        "2000",
		}
		cmd, err := prepareCmdForDatabaseAnalyze(utils, map[string]string{}, config, "sarif-latest", "target/codeqlReport.sarif", config.Database)
		assert.NoError(t, err)
		assert.NotEmpty(t, cmd)
		assert.Equal(t, 8, len(cmd))
		assert.Equal(t, "database analyze --format=sarif-latest --output=target/codeqlReport.sarif codeqlDB --threads=1 --ram=2000 security.ql", strings.Join(cmd, " "))
	})

	t.Run("Custom flags, overwriting threads", func(t *testing.T) {
		config := &codeqlExecuteScanOptions{
			Database:   "codeqlDB",
			QuerySuite: "security.ql",
			Threads:    "1",
			Ram:        "2000",
		}
		customFlags := map[string]string{
			"--threads": "--threads=2",
		}
		cmd, err := prepareCmdForDatabaseAnalyze(utils, customFlags, config, "sarif-latest", "target/codeqlReport.sarif", config.Database)
		assert.NoError(t, err)
		assert.NotEmpty(t, cmd)
		assert.Equal(t, 8, len(cmd))
		assert.Equal(t, "database analyze --format=sarif-latest --output=target/codeqlReport.sarif codeqlDB --ram=2000 --threads=2 security.ql", strings.Join(cmd, " "))
	})

	t.Run("Custom flags, overwriting threads (-j)", func(t *testing.T) {
		config := &codeqlExecuteScanOptions{
			Database:   "codeqlDB",
			QuerySuite: "security.ql",
			Threads:    "1",
			Ram:        "2000",
		}
		customFlags := map[string]string{
			"-j": "-j=2",
		}
		cmd, err := prepareCmdForDatabaseAnalyze(utils, customFlags, config, "sarif-latest", "target/codeqlReport.sarif", config.Database)
		assert.NoError(t, err)
		assert.NotEmpty(t, cmd)
		assert.Equal(t, 8, len(cmd))
		assert.Equal(t, "database analyze --format=sarif-latest --output=target/codeqlReport.sarif codeqlDB --ram=2000 -j=2 security.ql", strings.Join(cmd, " "))
	})

	t.Run("Custom flags, no overwriting", func(t *testing.T) {
		config := &codeqlExecuteScanOptions{
			Database:   "codeqlDB",
			QuerySuite: "security.ql",
			Threads:    "1",
			Ram:        "2000",
		}
		customFlags := map[string]string{
			"--no-download": "--no-download",
		}
		cmd, err := prepareCmdForDatabaseAnalyze(utils, customFlags, config, "sarif-latest", "target/codeqlReport.sarif", config.Database)
		assert.NoError(t, err)
		assert.NotEmpty(t, cmd)
		assert.Equal(t, 9, len(cmd))
		assert.Equal(t, "database analyze --format=sarif-latest --output=target/codeqlReport.sarif codeqlDB --threads=1 --ram=2000 --no-download security.ql", strings.Join(cmd, " "))
	})
}

func TestPrepareCmdForUploadResults(t *testing.T) {
	t.Parallel()

	config := &codeqlExecuteScanOptions{
		ModulePath: "./",
	}

	t.Run("All configs are set", func(t *testing.T) {
		repoInfo := &codeql.RepoInfo{
			CommitId:    "commitId",
			ServerUrl:   "http://github.com",
			Repo:        "repo",
			Owner:       "owner",
			AnalyzedRef: "refs/heads/branch",
		}
		cmd := prepareCmdForUploadResults(repoInfo, "token", filepath.Join(config.ModulePath, "target", "codeqlReport.sarif"))
		assert.NotEmpty(t, cmd)
		assert.Equal(t, 8, len(cmd))
	})

	t.Run("Configs are set partially", func(t *testing.T) {
		repoInfo := &codeql.RepoInfo{
			CommitId:  "commitId",
			ServerUrl: "http://github.com",
			Repo:      "repo",
		}
		cmd := prepareCmdForUploadResults(repoInfo, "token", filepath.Join(config.ModulePath, "target", "codeqlReport.sarif"))
		assert.NotEmpty(t, cmd)
		assert.Equal(t, 6, len(cmd))
	})

	t.Run("Empty token", func(t *testing.T) {
		repoInfo := &codeql.RepoInfo{
			CommitId:    "commitId",
			ServerUrl:   "http://github.com",
			Repo:        "repo",
			Owner:       "owner",
			AnalyzedRef: "refs/heads/branch",
		}
		cmd := prepareCmdForUploadResults(repoInfo, "", filepath.Join(config.ModulePath, "target", "codeqlReport.sarif"))
		assert.NotEmpty(t, cmd)
		assert.Equal(t, 7, len(cmd))
	})

	t.Run("Empty configs and token", func(t *testing.T) {
		repoInfo := &codeql.RepoInfo{}
		cmd := prepareCmdForUploadResults(repoInfo, "", filepath.Join(config.ModulePath, "target", "codeqlReport.sarif"))
		assert.NotEmpty(t, cmd)
		assert.Equal(t, 3, len(cmd))
	})
}

func TestAppendCodeqlQuerySuite(t *testing.T) {
	t.Parallel()

	t.Run("Empty query", func(t *testing.T) {
		utils := newCodeqlExecuteScanTestsUtils()
		cmd := []string{"database", "analyze"}
		querySuite := ""
		cmd = appendCodeqlQuerySuite(utils, cmd, querySuite, "")
		assert.Equal(t, 2, len(cmd))
	})

	t.Run("Not empty query", func(t *testing.T) {
		utils := newCodeqlExecuteScanTestsUtils()
		cmd := []string{"database", "analyze"}
		querySuite := "java-extended.ql"
		cmd = appendCodeqlQuerySuite(utils, cmd, querySuite, "")
		assert.Equal(t, 3, len(cmd))
	})

	t.Run("Add prefix to querySuite", func(t *testing.T) {
		utils := codeqlExecuteScanMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{
				Stub: func(call string, stdoutReturn map[string]string, shouldFailOnCommand map[string]error, stdout io.Writer) error {
					stdout.Write([]byte("test-java-security-extended.qls"))
					return nil
				},
			},
		}
		cmd := []string{"database", "analyze"}
		querySuite := "java-security-extended.qls"
		cmd = appendCodeqlQuerySuite(utils, cmd, querySuite, `s/^(java|python)-(security-extended\.qls|security-and-quality\.qls)/test-\1-\2/`)
		assert.Equal(t, 3, len(cmd))
		assert.Equal(t, "test-java-security-extended.qls", cmd[2])
	})

	t.Run("Don't add prefix to querySuite", func(t *testing.T) {
		utils := codeqlExecuteScanMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{
				Stub: func(call string, stdoutReturn map[string]string, shouldFailOnCommand map[string]error, stdout io.Writer) error {
					stdout.Write([]byte("php-security-extended.qls"))
					return nil
				},
			},
		}
		cmd := []string{"database", "analyze"}
		querySuite := "php-security-extended.qls"
		cmd = appendCodeqlQuerySuite(utils, cmd, querySuite, `s/^(java|python)-(security-extended\.qls|security-and-quality\.qls)/test-\1-\2/`)
		assert.Equal(t, 3, len(cmd))
		assert.Equal(t, "php-security-extended.qls", cmd[2])
	})

	t.Run("Error while transforming querySuite", func(t *testing.T) {
		utils := codeqlExecuteScanMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{
				Stub: func(call string, stdoutReturn map[string]string, shouldFailOnCommand map[string]error, stdout io.Writer) error {
					return fmt.Errorf("error")
				},
			},
		}
		cmd := []string{"database", "analyze"}
		querySuite := "php-security-extended.qls"
		cmd = appendCodeqlQuerySuite(utils, cmd, querySuite, `s/^(java|python)-(security-extended\.qls|security-and-quality\.qls)`)
		assert.Equal(t, 3, len(cmd))
		assert.Equal(t, "php-security-extended.qls", cmd[2])
	})

	t.Run("Empty transformed querySuite", func(t *testing.T) {
		utils := codeqlExecuteScanMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{
				Stub: func(call string, stdoutReturn map[string]string, shouldFailOnCommand map[string]error, stdout io.Writer) error {
					stdout.Write([]byte(""))
					return nil
				},
			},
		}
		cmd := []string{"database", "analyze"}
		querySuite := "python-security-extended.qls"
		cmd = appendCodeqlQuerySuite(utils, cmd, querySuite, `s/^(java|python)-(security-extended\.qls|security-and-quality\.qls)//`)
		assert.Equal(t, 2, len(cmd))
	})
}

func TestTransformQuerySuite(t *testing.T) {
	t.Run("Add prefix to querySuite", func(t *testing.T) {
		utils := codeqlExecuteScanMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{
				Stub: func(call string, stdoutReturn map[string]string, shouldFailOnCommand map[string]error, stdout io.Writer) error {
					stdout.Write([]byte("test-java-security-extended.qls"))
					return nil
				},
			},
		}
		input := "java-security-extended.qls"
		transformString := `s/^(java|python)-(security-extended.qls|security-and-quality.qls)/test-\1-\2/`
		expect := "test-java-security-extended.qls"
		result := transformQuerySuite(utils, input, transformString)
		assert.Equal(t, expect, result)
	})

	t.Run("Don't add prefix to querySuite", func(t *testing.T) {
		utils := codeqlExecuteScanMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{
				Stub: func(call string, stdoutReturn map[string]string, shouldFailOnCommand map[string]error, stdout io.Writer) error {
					stdout.Write([]byte("php-security-extended.qls"))
					return nil
				},
			},
		}
		input := "php-security-extended.qls"
		transformString := `s/^(java|python)-(security-extended.qls|security-and-quality.qls)/test-\1-\2/`
		expected := "php-security-extended.qls"
		result := transformQuerySuite(utils, input, transformString)
		assert.Equal(t, expected, result)

	})

	t.Run("Failed running transform cmd", func(t *testing.T) {
		utils := codeqlExecuteScanMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{
				Stub: func(call string, stdoutReturn map[string]string, shouldFailOnCommand map[string]error, stdout io.Writer) error {
					return fmt.Errorf("error")
				},
			},
		}
		input := "php-security-extended.qls"
		transformString := `s//test-\1-\2/`
		result := transformQuerySuite(utils, input, transformString)
		assert.Equal(t, input, result)
	})

	t.Run("Transform querySuite to empty string", func(t *testing.T) {
		utils := codeqlExecuteScanMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{
				Stub: func(call string, stdoutReturn map[string]string, shouldFailOnCommand map[string]error, stdout io.Writer) error {
					stdout.Write([]byte(""))
					return nil
				},
			},
		}
		input := "java-security-extended.qls"
		transformString := `s/^(java|python)-(security-extended.qls|security-and-quality.qls)//`
		expect := ""
		result := transformQuerySuite(utils, input, transformString)
		assert.Equal(t, expect, result)
	})
}

func TestGetLangFromBuildTool(t *testing.T) {
	t.Parallel()

	t.Run("Build tool Maven", func(t *testing.T) {
		assert.Equal(t, "java", getLangFromBuildTool("maven"))
	})
	t.Run("Build tool Pip", func(t *testing.T) {
		assert.Equal(t, "python", getLangFromBuildTool("pip"))
	})
	t.Run("Build tool Npm", func(t *testing.T) {
		assert.Equal(t, "javascript", getLangFromBuildTool("npm"))
	})
	t.Run("Build tool Yarn", func(t *testing.T) {
		assert.Equal(t, "javascript", getLangFromBuildTool("yarn"))
	})
	t.Run("Build tool Golang", func(t *testing.T) {
		assert.Equal(t, "go", getLangFromBuildTool("golang"))
	})
	t.Run("Build tool Unknown", func(t *testing.T) {
		assert.Equal(t, "", getLangFromBuildTool("unknown"))
	})
}

func TestGetToken(t *testing.T) {
	t.Run("Token is set in config", func(t *testing.T) {
		config := &codeqlExecuteScanOptions{GithubToken: "token"}
		os.Setenv("GITHUB_TOKEN", "token_from_env")
		hasToken, token := getToken(config)
		os.Clearenv()
		assert.True(t, hasToken)
		assert.NotEmpty(t, token)
		assert.Equal(t, "token", token)
	})

	t.Run("Token is set in env", func(t *testing.T) {
		config := &codeqlExecuteScanOptions{}
		os.Setenv("GITHUB_TOKEN", "token_from_env")
		hasToken, token := getToken(config)
		os.Clearenv()
		assert.True(t, hasToken)
		assert.NotEmpty(t, token)
		assert.Equal(t, "token_from_env", token)
	})

	t.Run("Token is not set", func(t *testing.T) {
		config := &codeqlExecuteScanOptions{}
		hasToken, token := getToken(config)
		assert.False(t, hasToken)
		assert.Empty(t, token)
	})
}

func TestCheckForCompliance(t *testing.T) {
	t.Parallel()

	config := &codeqlExecuteScanOptions{VulnerabilityThresholdTotal: 0}
	repoInfo := &codeql.RepoInfo{
		FullUrl:     "http://github.com/Test/repo",
		AnalyzedRef: "refs/heads/branch",
	}

	t.Run("Project is compliant", func(t *testing.T) {
		scanResults := []codeql.CodeqlFindings{
			{
				ClassificationName: codeql.AuditAll,
				Total:              10,
				Audited:            10,
			},
		}
		assert.NoError(t, checkForCompliance(scanResults, config, repoInfo))
	})

	t.Run("Project is not compliant", func(t *testing.T) {
		scanResults := []codeql.CodeqlFindings{
			{
				ClassificationName: codeql.AuditAll,
				Total:              20,
				Audited:            10,
			},
		}
		assert.Error(t, checkForCompliance(scanResults, config, repoInfo))
	})

	t.Run("Don't check Optional findings", func(t *testing.T) {
		scanResults := []codeql.CodeqlFindings{
			{
				ClassificationName: codeql.Optional,
				Total:              10,
				Audited:            0,
			},
		}
		assert.NoError(t, checkForCompliance(scanResults, config, repoInfo))
	})
}

func TestGetLanguageList(t *testing.T) {
	t.Parallel()

	t.Run("Comma separated with spaces and empties", func(t *testing.T) {
		cfg := &codeqlExecuteScanOptions{
			Language:  "javascript, python, ,go ,",
			BuildTool: "npm",
		}
		got := getLanguageList(cfg)
		assert.Equal(t, []string{"javascript", "python", "go"}, got)
	})

	t.Run("Single explicit language", func(t *testing.T) {
		cfg := &codeqlExecuteScanOptions{
			Language: "python",
		}
		got := getLanguageList(cfg)
		assert.Equal(t, []string{"python"}, got)
	})

	t.Run("Inferred from build tool", func(t *testing.T) {
		cfg := &codeqlExecuteScanOptions{
			BuildTool: "maven",
		}
		got := getLanguageList(cfg)
		assert.Equal(t, []string{"java"}, got)
	})

	t.Run("None available returns nil", func(t *testing.T) {
		cfg := &codeqlExecuteScanOptions{}
		got := getLanguageList(cfg)
		assert.Nil(t, got)
	})
}

func TestCloneFlags(t *testing.T) {
	src := map[string]string{"--threads": "--threads=2", "--foo": "bar"}
	dst := cloneFlags(src)

	assert.Equal(t, src, dst)

	// check that changes in the dst map does not affect scr
	dst["--threads"] = "--threads=4"
	delete(dst, "--foo")

	assert.Equal(t, "--threads=2", src["--threads"])
	assert.Equal(t, "bar", src["--foo"])
}

func TestRunDatabaseAnalyze_SingleLanguage(t *testing.T) {
	t.Parallel()
	var calls []string
	utils := codeqlExecuteScanMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{
			Stub: func(call string, stdoutReturn map[string]string, shouldFailOnCommand map[string]error, stdout io.Writer) error {
				calls = append(calls, call)
				return nil
			},
		},
		FilesMock:      &mock.FilesMock{},
		HttpClientMock: &mock.HttpClientMock{},
	}
	cfg := &codeqlExecuteScanOptions{
		Database:   "codeqlDB",
		ModulePath: ".",
		Language:   "javascript",
	}
	custom := map[string]string{}

	reports, sarifs, err := runDatabaseAnalyze(cfg, custom, utils, false)
	assert.NoError(t, err)

	expectSarif := filepath.Join(".", "target", "codeqlReport.sarif")
	expectCSV := filepath.Join(".", "target", "codeqlReport.csv")

	assert.ElementsMatch(t,
		[]piperutils.Path{{Target: expectSarif}, {Target: expectCSV}},
		reports,
	)
	assert.Equal(t, []string{expectSarif}, sarifs)

	joined := strings.Join(calls, "\n")
	assert.Contains(t, joined, "database analyze --format=sarif-latest --output="+expectSarif+" codeqlDB")
	assert.Contains(t, joined, "database analyze --format=csv --output="+expectCSV+" codeqlDB")
}

func TestRunDatabaseAnalyze_MultiLanguage(t *testing.T) {
	t.Parallel()
	var calls []string
	utils := codeqlExecuteScanMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{
			Stub: func(call string, stdoutReturn map[string]string, shouldFailOnCommand map[string]error, stdout io.Writer) error {
				calls = append(calls, call)
				return nil
			},
		},
		FilesMock:      &mock.FilesMock{},
		HttpClientMock: &mock.HttpClientMock{},
	}

	cfg := &codeqlExecuteScanOptions{
		Database:   "codeqlDB",
		ModulePath: ".",
		Language:   "javascript,python",
	}
	custom := map[string]string{}

	reports, sarifs, err := runDatabaseAnalyze(cfg, custom, utils, true)
	assert.NoError(t, err)

	jsSarif := filepath.Join(".", "target", "javascript.sarif")
	pySarif := filepath.Join(".", "target", "python.sarif")
	jsCSV := filepath.Join(".", "target", "javascript.csv")
	pyCSV := filepath.Join(".", "target", "python.csv")

	// reports should include all 4 files
	assert.ElementsMatch(t,
		[]piperutils.Path{{Target: jsSarif}, {Target: jsCSV}, {Target: pySarif}, {Target: pyCSV}},
		reports,
	)
	// sarifFiles should be both per-language sarif outputs
	assert.ElementsMatch(t, []string{jsSarif, pySarif}, sarifs)

	joined := strings.Join(calls, "\n")

	assert.Contains(t, joined, "database analyze --format=sarif-latest --output="+jsSarif+" codeqlDB/javascript")
	assert.Contains(t, joined, "database analyze --format=sarif-latest --output="+pySarif+" codeqlDB/python")

	assert.Contains(t, joined, "--sarif-category=javascript")
	assert.Contains(t, joined, "--sarif-category=python")

	assert.Contains(t, joined, "database analyze --format=csv --output="+jsCSV+" codeqlDB/javascript")
	assert.Contains(t, joined, "database analyze --format=csv --output="+pyCSV+" codeqlDB/python")
}

func TestRunCustomCommand(t *testing.T) {
	t.Parallel()

	t.Run("Success: simple command with args", func(t *testing.T) {
		var calls []string
		utils := codeqlExecuteScanMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{
				Stub: func(call string, _ map[string]string, _ map[string]error, _ io.Writer) error {
					calls = append(calls, call)
					return nil
				},
			},
			FilesMock:      &mock.FilesMock{},
			HttpClientMock: &mock.HttpClientMock{},
		}

		err := runCustomCommand(utils, `echo "hello world"`)
		assert.NoError(t, err)
		if assert.Len(t, calls, 1) {
			assert.Equal(t, "echo hello world", calls[0])
		}
	})

	t.Run("Parse error: invalid quoting", func(t *testing.T) {
		utils := newCodeqlExecuteScanTestsUtils() // stub isn't invoked because split fails first
		err := runCustomCommand(utils, `echo "unterminated`)
		assert.Error(t, err)
	})

	t.Run("Exec error: command fails to run", func(t *testing.T) {
		utils := codeqlExecuteScanMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{
				Stub: func(call string, _ map[string]string, _ map[string]error, _ io.Writer) error {
					return fmt.Errorf("boom")
				},
			},
			FilesMock:      &mock.FilesMock{},
			HttpClientMock: &mock.HttpClientMock{},
		}

		err := runCustomCommand(utils, `false --flag`)
		assert.Error(t, err)
	})
}

func Test_prepareCodeQLConfigFile(t *testing.T) {
	t.Run("creates_or_updates_default_config_next_to_codeql_binary", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("requires POSIX exec bits for the fake binary")
		}
		dir := t.TempDir()

		codeqlBin := filepath.Join(dir, "codeql")
		writeExec(t, codeqlBin, "#!/bin/sh\necho CodeQL\n")

		origPath := os.Getenv("PATH")
		t.Cleanup(func() { _ = os.Setenv("PATH", origPath) })
		require.NoError(t, os.Setenv("PATH", dir+string(os.PathListSeparator)+origPath))

		opts := &codeqlExecuteScanOptions{
			Paths:       " src \nlib/utils\n",
			PathsIgnore: "vendor\n**/*.gen.go\n",
		}

		err := prepareCodeQLConfigFile(opts)
		require.NoError(t, err)

		qlPath, err := codeql.Which("codeql")
		require.NoError(t, err)

		loc, found := strings.CutSuffix(qlPath, "codeql")
		require.True(t, found)

		cfgPath := path.Join(loc, "default-codeql-config.yml")
		b, err := os.ReadFile(cfgPath)
		require.NoError(t, err)
		out := normalizeNL(string(b))

		// Assert it contains the expected lists (don’t depend on ordering)
		assert.Contains(t, out, "paths:")
		assert.Contains(t, out, "- src")
		assert.Contains(t, out, "- lib/utils")

		assert.Contains(t, out, "paths-ignore:")
		assert.Contains(t, out, "- vendor")
		// yaml.v3 single-quotes strings with '*' etc.
		assert.Contains(t, out, "- '**/*.gen.go'")
	})

	t.Run("when_codeql_binary_missing_returns_error", func(t *testing.T) {
		dir := t.TempDir()
		origPath := os.Getenv("PATH")
		t.Cleanup(func() { _ = os.Setenv("PATH", origPath) })

		// PATH without codeql
		require.NoError(t, os.Setenv("PATH", dir))

		opts := &codeqlExecuteScanOptions{
			Paths:       "a",
			PathsIgnore: "b",
		}
		err := prepareCodeQLConfigFile(opts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not locate codeql executable")
	})

	t.Run("propagates_append_error_when_default_config_path_is_a_directory", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("requires POSIX exec bits and reliable directory perms")
		}

		dir := t.TempDir()

		// Create fake codeql binary
		codeqlBin := filepath.Join(dir, "codeql")
		writeExec(t, codeqlBin, "#!/bin/sh\necho CodeQL\n")

		defaultCfgDir := filepath.Join(dir, "default-codeql-config.yml")
		require.NoError(t, os.Mkdir(defaultCfgDir, 0o755))

		origPath := os.Getenv("PATH")
		t.Cleanup(func() { _ = os.Setenv("PATH", origPath) })
		require.NoError(t, os.Setenv("PATH", dir+string(os.PathListSeparator)+origPath))

		opts := &codeqlExecuteScanOptions{
			Paths:       "x",
			PathsIgnore: "y",
		}
		err := prepareCodeQLConfigFile(opts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "append paths and paths ignore to the default config")
	})

	t.Run("no_changes_when_both_paths_empty_but_still_needs_binary", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("requires POSIX exec bits for the fake binary")
		}
		dir := t.TempDir()

		codeqlBin := filepath.Join(dir, "codeql")
		writeExec(t, codeqlBin, "#!/bin/sh\necho CodeQL\n")
		origPath := os.Getenv("PATH")
		t.Cleanup(func() { _ = os.Setenv("PATH", origPath) })
		require.NoError(t, os.Setenv("PATH", dir+string(os.PathListSeparator)+origPath))

		cfgFile := filepath.Join(dir, "default-codeql-config.yml")
		writeCodeQLFile(t, cfgFile, "")

		opts := &codeqlExecuteScanOptions{
			Paths:       "",
			PathsIgnore: "",
		}

		before := readCodeQLFile(t, cfgFile)
		time.Sleep(10 * time.Millisecond) // avoid flakiness on very fast FS

		err := prepareCodeQLConfigFile(opts)
		require.NoError(t, err)

		after := readCodeQLFile(t, cfgFile)
		assert.Equal(t, before, after, "file should remain unchanged when nothing to write")
	})
}

func writeExec(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), 0o755))
	require.NoError(t, os.Chmod(path, 0o755))
	info, err := os.Stat(path)
	require.NoError(t, err)
	require.False(t, info.IsDir())
	require.NotZero(t, info.Mode()&0o111, "should be executable")
}

func writeCodeQLFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func readCodeQLFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(b)
}

func normalizeNL(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}
