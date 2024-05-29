//go:build unit
// +build unit

package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/codeql"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type codeqlExecuteScanMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
	*mock.HttpClientMock
}

func newCodeqlExecuteScanTestsUtils() codeqlExecuteScanMockUtils {
	utils := codeqlExecuteScanMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
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
		assert.Equal(t, " --settings=test.xml", params)
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
		assert.Equal(t, " --global-settings=global.xml", params)
	})

	t.Run("Project and Global Settings file", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", ProjectSettingsFile: "test.xml", GlobalSettingsFile: "global.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		assert.Equal(t, " --global-settings=global.xml --settings=test.xml", params)
	})

	t.Run("ProjectSettingsFile https url", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", ProjectSettingsFile: "https://jenkins-sap-test.com/test.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		assert.Equal(t, " --settings=.pipeline/mavenProjectSettings.xml", params)
	})

	t.Run("ProjectSettingsFile http url", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", ProjectSettingsFile: "http://jenkins-sap-test.com/test.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		assert.Equal(t, " --settings=.pipeline/mavenProjectSettings.xml", params)
	})

	t.Run("GlobalSettingsFile https url", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", GlobalSettingsFile: "https://jenkins-sap-test.com/test.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		assert.Equal(t, " --global-settings=.pipeline/mavenGlobalSettings.xml", params)
	})

	t.Run("GlobalSettingsFile http url", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", GlobalSettingsFile: "http://jenkins-sap-test.com/test.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		assert.Equal(t, " --global-settings=.pipeline/mavenGlobalSettings.xml", params)
	})

	t.Run("ProjectSettingsFile and GlobalSettingsFile https url", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", GlobalSettingsFile: "https://jenkins-sap-test.com/test.xml", ProjectSettingsFile: "http://jenkins-sap-test.com/test.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		assert.Equal(t, " --global-settings=.pipeline/mavenGlobalSettings.xml --settings=.pipeline/mavenProjectSettings.xml", params)
	})

	t.Run("ProjectSettingsFile and GlobalSettingsFile http url", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", GlobalSettingsFile: "http://jenkins-sap-test.com/test.xml", ProjectSettingsFile: "http://jenkins-sap-test.com/test.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		assert.Equal(t, " --global-settings=.pipeline/mavenGlobalSettings.xml --settings=.pipeline/mavenProjectSettings.xml", params)
	})

	t.Run("ProjectSettingsFile file and GlobalSettingsFile https url", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", GlobalSettingsFile: "https://jenkins-sap-test.com/test.xml", ProjectSettingsFile: "test.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		assert.Equal(t, " --global-settings=.pipeline/mavenGlobalSettings.xml --settings=test.xml", params)
	})

	t.Run("ProjectSettingsFile file and GlobalSettingsFile https url", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", GlobalSettingsFile: "http://jenkins-sap-test.com/test.xml", ProjectSettingsFile: "test.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		assert.Equal(t, " --global-settings=.pipeline/mavenGlobalSettings.xml --settings=test.xml", params)
	})

	t.Run("ProjectSettingsFile https url and GlobalSettingsFile file", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", GlobalSettingsFile: "global.xml", ProjectSettingsFile: "http://jenkins-sap-test.com/test.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		assert.Equal(t, " --global-settings=global.xml --settings=.pipeline/mavenProjectSettings.xml", params)
	})

	t.Run("ProjectSettingsFile http url and GlobalSettingsFile file", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", GlobalSettingsFile: "global.xml", ProjectSettingsFile: "http://jenkins-sap-test.com/test.xml"}
		buildCmd := "mvn clean install"
		params := getMavenSettings(buildCmd, &config, newCodeqlExecuteScanTestsUtils())
		assert.Equal(t, " --global-settings=global.xml --settings=.pipeline/mavenProjectSettings.xml", params)
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
		assert.Equal(t, "mvn clean install --global-settings=global.xml --settings=test.xml", customFlags["--command"])
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
		cmd, err := prepareCmdForDatabaseCreate(map[string]string{}, config, newCodeqlExecuteScanTestsUtils())
		assert.NoError(t, err)
		assert.NotEmpty(t, cmd)
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
		cmd, err := prepareCmdForDatabaseCreate(map[string]string{}, config, newCodeqlExecuteScanTestsUtils())
		assert.NoError(t, err)
		assert.NotEmpty(t, cmd)
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
		_, err := prepareCmdForDatabaseCreate(map[string]string{}, config, newCodeqlExecuteScanTestsUtils())
		assert.Error(t, err)
	})

	t.Run("No custom flags, invalid build tool, no language specified", func(t *testing.T) {
		config := &codeqlExecuteScanOptions{
			Database:   "codeqlDB",
			ModulePath: "./",
			BuildTool:  "test",
		}
		_, err := prepareCmdForDatabaseCreate(map[string]string{}, config, newCodeqlExecuteScanTestsUtils())
		assert.Error(t, err)
	})

	t.Run("No custom flags, invalid build tool, language specified", func(t *testing.T) {
		config := &codeqlExecuteScanOptions{
			Database:   "codeqlDB",
			ModulePath: "./",
			BuildTool:  "test",
			Language:   "javascript",
		}
		cmd, err := prepareCmdForDatabaseCreate(map[string]string{}, config, newCodeqlExecuteScanTestsUtils())
		assert.NoError(t, err)
		assert.NotEmpty(t, cmd)
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
		cmd, err := prepareCmdForDatabaseCreate(customFlags, config, newCodeqlExecuteScanTestsUtils())
		assert.NoError(t, err)
		assert.NotEmpty(t, cmd)
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
		cmd, err := prepareCmdForDatabaseCreate(customFlags, config, newCodeqlExecuteScanTestsUtils())
		assert.NoError(t, err)
		assert.NotEmpty(t, cmd)
		assert.Equal(t, 10, len(cmd))
		assert.True(t, "database create codeqlDB --overwrite --working-dir ./ --language=javascript --ram=2000 -j=1 --source-root=customSrcRoot/" == strings.Join(cmd, " ") ||
			"database create codeqlDB --overwrite --working-dir ./ --language=javascript --ram=2000 --source-root=customSrcRoot/ -j=1" == strings.Join(cmd, " "))
	})

}

func TestPrepareCmdForDatabaseAnalyze(t *testing.T) {
	t.Parallel()

	t.Run("No additional flags, no querySuite, sarif format", func(t *testing.T) {
		config := &codeqlExecuteScanOptions{
			Database: "codeqlDB",
		}
		cmd, err := prepareCmdForDatabaseAnalyze(map[string]string{}, config, "sarif-latest", "target/codeqlReport.sarif")
		assert.NoError(t, err)
		assert.NotEmpty(t, cmd)
		assert.Equal(t, 5, len(cmd))
		assert.Equal(t, "database analyze --format=sarif-latest --output=target/codeqlReport.sarif codeqlDB", strings.Join(cmd, " "))
	})

	t.Run("No additional flags, no querySuite, csv format", func(t *testing.T) {
		config := &codeqlExecuteScanOptions{
			Database: "codeqlDB",
		}
		cmd, err := prepareCmdForDatabaseAnalyze(map[string]string{}, config, "csv", "target/codeqlReport.csv")
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
		cmd, err := prepareCmdForDatabaseAnalyze(map[string]string{}, config, "sarif-latest", "target/codeqlReport.sarif")
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
		cmd, err := prepareCmdForDatabaseAnalyze(map[string]string{}, config, "sarif-latest", "target/codeqlReport.sarif")
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
		cmd, err := prepareCmdForDatabaseAnalyze(customFlags, config, "sarif-latest", "target/codeqlReport.sarif")
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
		cmd, err := prepareCmdForDatabaseAnalyze(customFlags, config, "sarif-latest", "target/codeqlReport.sarif")
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
		cmd, err := prepareCmdForDatabaseAnalyze(customFlags, config, "sarif-latest", "target/codeqlReport.sarif")
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
		cmd := prepareCmdForUploadResults(config, repoInfo, "token")
		assert.NotEmpty(t, cmd)
		assert.Equal(t, 8, len(cmd))
	})

	t.Run("Configs are set partially", func(t *testing.T) {
		repoInfo := &codeql.RepoInfo{
			CommitId:  "commitId",
			ServerUrl: "http://github.com",
			Repo:      "repo",
		}
		cmd := prepareCmdForUploadResults(config, repoInfo, "token")
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
		cmd := prepareCmdForUploadResults(config, repoInfo, "")
		assert.NotEmpty(t, cmd)
		assert.Equal(t, 7, len(cmd))
	})

	t.Run("Empty configs and token", func(t *testing.T) {
		repoInfo := &codeql.RepoInfo{}
		cmd := prepareCmdForUploadResults(config, repoInfo, "")
		assert.NotEmpty(t, cmd)
		assert.Equal(t, 3, len(cmd))
	})
}

func TestAppendCodeqlQuery(t *testing.T) {
	t.Parallel()

	t.Run("Empty query", func(t *testing.T) {
		cmd := []string{"database", "analyze"}
		query := ""
		cmd = appendCodeqlQuery(cmd, query)
		assert.Equal(t, 2, len(cmd))
	})

	t.Run("Not empty query", func(t *testing.T) {
		cmd := []string{"database", "analyze"}
		query := "java-extended.ql"
		cmd = appendCodeqlQuery(cmd, query)
		assert.Equal(t, 3, len(cmd))
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
