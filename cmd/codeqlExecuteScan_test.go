//go:build unit
// +build unit

package cmd

import (
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/codeql"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type codeqlExecuteScanMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newCodeqlExecuteScanTestsUtils() codeqlExecuteScanMockUtils {
	utils := codeqlExecuteScanMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunCodeqlExecuteScan(t *testing.T) {

	t.Run("Valid CodeqlExecuteScan", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", ModulePath: "./"}
		_, err := runCodeqlExecuteScan(&config, nil, newCodeqlExecuteScanTestsUtils())
		assert.NoError(t, err)
	})

	t.Run("No auth token passed on upload results", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", UploadResults: true, ModulePath: "./"}
		_, err := runCodeqlExecuteScan(&config, nil, newCodeqlExecuteScanTestsUtils())
		assert.Error(t, err)
	})

	t.Run("GitCommitID is NA on upload results", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", UploadResults: true, ModulePath: "./", CommitID: "NA"}
		_, err := runCodeqlExecuteScan(&config, nil, newCodeqlExecuteScanTestsUtils())
		assert.Error(t, err)
	})

	t.Run("Custom buildtool", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "custom", Language: "javascript", ModulePath: "./"}
		_, err := runCodeqlExecuteScan(&config, nil, newCodeqlExecuteScanTestsUtils())
		assert.NoError(t, err)
	})

	t.Run("Custom buildtool but no language specified", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "custom", ModulePath: "./", GithubToken: "test"}
		_, err := runCodeqlExecuteScan(&config, nil, newCodeqlExecuteScanTestsUtils())
		assert.Error(t, err)
	})

	t.Run("Invalid buildtool and no language specified", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "test", ModulePath: "./", GithubToken: "test"}
		_, err := runCodeqlExecuteScan(&config, nil, newCodeqlExecuteScanTestsUtils())
		assert.Error(t, err)
	})

	t.Run("Invalid buildtool but language specified", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "test", Language: "javascript", ModulePath: "./", GithubToken: "test"}
		_, err := runCodeqlExecuteScan(&config, nil, newCodeqlExecuteScanTestsUtils())
		assert.NoError(t, err)
	})
}

func TestGetGitRepoInfo(t *testing.T) {
	t.Run("Valid https URL1", func(t *testing.T) {
		var repoInfo codeql.RepoInfo
		err := getGitRepoInfo("https://github.hello.test/Testing/fortify.git", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "fortify", repoInfo.Repo)
		assert.Equal(t, "Testing", repoInfo.Owner)
	})

	t.Run("Valid https URL2", func(t *testing.T) {
		var repoInfo codeql.RepoInfo
		err := getGitRepoInfo("https://github.hello.test/Testing/fortify", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "fortify", repoInfo.Repo)
		assert.Equal(t, "Testing", repoInfo.Owner)
	})
	t.Run("Valid https URL1 with dots", func(t *testing.T) {
		var repoInfo codeql.RepoInfo
		err := getGitRepoInfo("https://github.hello.test/Testing/com.sap.fortify.git", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "com.sap.fortify", repoInfo.Repo)
		assert.Equal(t, "Testing", repoInfo.Owner)
	})

	t.Run("Valid https URL2 with dots", func(t *testing.T) {
		var repoInfo codeql.RepoInfo
		err := getGitRepoInfo("https://github.hello.test/Testing/com.sap.fortify", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "com.sap.fortify", repoInfo.Repo)
		assert.Equal(t, "Testing", repoInfo.Owner)
	})
	t.Run("Valid https URL1 with username and token", func(t *testing.T) {
		var repoInfo codeql.RepoInfo
		err := getGitRepoInfo("https://username:token@github.hello.test/Testing/fortify.git", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "fortify", repoInfo.Repo)
		assert.Equal(t, "Testing", repoInfo.Owner)
	})

	t.Run("Valid https URL2 with username and token", func(t *testing.T) {
		var repoInfo codeql.RepoInfo
		err := getGitRepoInfo("https://username:token@github.hello.test/Testing/fortify", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "fortify", repoInfo.Repo)
		assert.Equal(t, "Testing", repoInfo.Owner)
	})

	t.Run("Invalid https URL as no org/Owner passed", func(t *testing.T) {
		var repoInfo codeql.RepoInfo
		assert.Error(t, getGitRepoInfo("https://github.com/fortify", &repoInfo))
	})

	t.Run("Invalid URL as no protocol passed", func(t *testing.T) {
		var repoInfo codeql.RepoInfo
		assert.Error(t, getGitRepoInfo("github.hello.test/Testing/fortify", &repoInfo))
	})

	t.Run("Valid ssh URL1", func(t *testing.T) {
		var repoInfo codeql.RepoInfo
		err := getGitRepoInfo("git@github.hello.test/Testing/fortify.git", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "fortify", repoInfo.Repo)
		assert.Equal(t, "Testing", repoInfo.Owner)
	})

	t.Run("Valid ssh URL2", func(t *testing.T) {
		var repoInfo codeql.RepoInfo
		err := getGitRepoInfo("git@github.hello.test/Testing/fortify", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "fortify", repoInfo.Repo)
		assert.Equal(t, "Testing", repoInfo.Owner)
	})
	t.Run("Valid ssh URL1 with dots", func(t *testing.T) {
		var repoInfo codeql.RepoInfo
		err := getGitRepoInfo("git@github.hello.test/Testing/com.sap.fortify.git", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "com.sap.fortify", repoInfo.Repo)
		assert.Equal(t, "Testing", repoInfo.Owner)
	})

	t.Run("Valid ssh URL2 with dots", func(t *testing.T) {
		var repoInfo codeql.RepoInfo
		err := getGitRepoInfo("git@github.hello.test/Testing/com.sap.fortify", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "com.sap.fortify", repoInfo.Repo)
		assert.Equal(t, "Testing", repoInfo.Owner)
	})

	t.Run("Invalid ssh URL as no org/Owner passed", func(t *testing.T) {
		var repoInfo codeql.RepoInfo
		assert.Error(t, getGitRepoInfo("git@github.com/fortify", &repoInfo))
	})
}

func TestInitGitInfo(t *testing.T) {
	t.Run("Valid URL1", func(t *testing.T) {
		config := codeqlExecuteScanOptions{Repository: "https://github.hello.test/Testing/codeql.git", AnalyzedRef: "refs/head/branch", CommitID: "abcd1234"}
		repoInfo, err := initGitInfo(&config)
		assert.NoError(t, err)
		assert.Equal(t, "abcd1234", repoInfo.CommitId)
		assert.Equal(t, "Testing", repoInfo.Owner)
		assert.Equal(t, "codeql", repoInfo.Repo)
		assert.Equal(t, "refs/head/branch", repoInfo.Ref)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
	})

	t.Run("Valid URL2", func(t *testing.T) {
		config := codeqlExecuteScanOptions{Repository: "https://github.hello.test/Testing/codeql", AnalyzedRef: "refs/head/branch", CommitID: "abcd1234"}
		repoInfo, err := initGitInfo(&config)
		assert.NoError(t, err)
		assert.Equal(t, "abcd1234", repoInfo.CommitId)
		assert.Equal(t, "Testing", repoInfo.Owner)
		assert.Equal(t, "codeql", repoInfo.Repo)
		assert.Equal(t, "refs/head/branch", repoInfo.Ref)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
	})

	t.Run("Valid url with dots URL1", func(t *testing.T) {
		config := codeqlExecuteScanOptions{Repository: "https://github.hello.test/Testing/com.sap.codeql.git", AnalyzedRef: "refs/head/branch", CommitID: "abcd1234"}
		repoInfo, err := initGitInfo(&config)
		assert.NoError(t, err)
		assert.Equal(t, "abcd1234", repoInfo.CommitId)
		assert.Equal(t, "Testing", repoInfo.Owner)
		assert.Equal(t, "com.sap.codeql", repoInfo.Repo)
		assert.Equal(t, "refs/head/branch", repoInfo.Ref)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
	})

	t.Run("Valid url with dots URL2", func(t *testing.T) {
		config := codeqlExecuteScanOptions{Repository: "https://github.hello.test/Testing/com.sap.codeql", AnalyzedRef: "refs/head/branch", CommitID: "abcd1234"}
		repoInfo, err := initGitInfo(&config)
		assert.NoError(t, err)
		assert.Equal(t, "abcd1234", repoInfo.CommitId)
		assert.Equal(t, "Testing", repoInfo.Owner)
		assert.Equal(t, "com.sap.codeql", repoInfo.Repo)
		assert.Equal(t, "refs/head/branch", repoInfo.Ref)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
	})

	t.Run("Valid url with username and token URL1", func(t *testing.T) {
		config := codeqlExecuteScanOptions{Repository: "https://username:token@github.hello.test/Testing/codeql.git", AnalyzedRef: "refs/head/branch", CommitID: "abcd1234"}
		repoInfo, err := initGitInfo(&config)
		assert.NoError(t, err)
		assert.Equal(t, "abcd1234", repoInfo.CommitId)
		assert.Equal(t, "Testing", repoInfo.Owner)
		assert.Equal(t, "codeql", repoInfo.Repo)
		assert.Equal(t, "refs/head/branch", repoInfo.Ref)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
	})

	t.Run("Valid url with username and token URL2", func(t *testing.T) {
		config := codeqlExecuteScanOptions{Repository: "https://username:token@github.hello.test/Testing/codeql", AnalyzedRef: "refs/head/branch", CommitID: "abcd1234"}
		repoInfo, err := initGitInfo(&config)
		assert.NoError(t, err)
		assert.Equal(t, "abcd1234", repoInfo.CommitId)
		assert.Equal(t, "Testing", repoInfo.Owner)
		assert.Equal(t, "codeql", repoInfo.Repo)
		assert.Equal(t, "refs/head/branch", repoInfo.Ref)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
	})

	t.Run("Invalid URL with no org/reponame", func(t *testing.T) {
		config := codeqlExecuteScanOptions{Repository: "https://github.hello.test", AnalyzedRef: "refs/head/branch", CommitID: "abcd1234"}
		repoInfo, err := initGitInfo(&config)
		assert.NoError(t, err)
		_, err = orchestrator.NewOrchestratorSpecificConfigProvider()
		assert.Equal(t, "abcd1234", repoInfo.CommitId)
		assert.Equal(t, "refs/head/branch", repoInfo.Ref)
		if err != nil {
			assert.Equal(t, "", repoInfo.Owner)
			assert.Equal(t, "", repoInfo.Repo)
			assert.Equal(t, "", repoInfo.ServerUrl)
		}
	})
}

func TestWaitSarifUploaded(t *testing.T) {
	t.Parallel()
	config := codeqlExecuteScanOptions{SarifCheckRetryInterval: 1, SarifCheckMaxRetries: 5}
	t.Run("Fast complete upload", func(t *testing.T) {
		codeqlScanAuditMock := CodeqlSarifUploaderMock{counter: 0}
		timerStart := time.Now()
		err := waitSarifUploaded(&config, &codeqlScanAuditMock)
		assert.Less(t, time.Now().Sub(timerStart), time.Second)
		assert.NoError(t, err)
	})
	t.Run("Long completed upload", func(t *testing.T) {
		codeqlScanAuditMock := CodeqlSarifUploaderMock{counter: 2}
		timerStart := time.Now()
		err := waitSarifUploaded(&config, &codeqlScanAuditMock)
		assert.GreaterOrEqual(t, time.Now().Sub(timerStart), time.Second*2)
		assert.NoError(t, err)
	})
	t.Run("Failed upload", func(t *testing.T) {
		codeqlScanAuditMock := CodeqlSarifUploaderMock{counter: -1}
		err := waitSarifUploaded(&config, &codeqlScanAuditMock)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to upload sarif file")
	})
	t.Run("Error while checking sarif uploading", func(t *testing.T) {
		codeqlScanAuditErrorMock := CodeqlSarifUploaderErrorMock{counter: -1}
		err := waitSarifUploaded(&config, &codeqlScanAuditErrorMock)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "test error")
	})
	t.Run("Completed upload after getting errors from server", func(t *testing.T) {
		codeqlScanAuditErrorMock := CodeqlSarifUploaderErrorMock{counter: 3}
		err := waitSarifUploaded(&config, &codeqlScanAuditErrorMock)
		assert.NoError(t, err)
	})
	t.Run("Max retries reached", func(t *testing.T) {
		codeqlScanAuditErrorMock := CodeqlSarifUploaderErrorMock{counter: 6}
		err := waitSarifUploaded(&config, &codeqlScanAuditErrorMock)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "max retries reached")
	})
}

func TestGetMavenSettings(t *testing.T) {
	t.Parallel()
	t.Run("No maven", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "npm"}
		params := getMavenSettings(&config)
		assert.Equal(t, 0, len(params))
	})

	t.Run("No build command", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven"}
		params := getMavenSettings(&config)
		assert.Equal(t, 0, len(params))
	})

	t.Run("Project Settings file", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", BuildCommand: "mvn clean install", ProjectSettingsFile: "test.xml"}
		params := getMavenSettings(&config)
		assert.Equal(t, 1, len(params))
		assert.Equal(t, "--settings=test.xml", params[0])
	})

	t.Run("Skip Project Settings file incase already used", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", BuildCommand: "mvn clean install --settings=project.xml", ProjectSettingsFile: "test.xml"}
		params := getMavenSettings(&config)
		assert.Equal(t, 0, len(params))
	})

	t.Run("Global Settings file", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", BuildCommand: "mvn clean install", GlobalSettingsFile: "gloabl.xml"}
		params := getMavenSettings(&config)
		assert.Equal(t, 1, len(params))
		assert.Equal(t, "--global-settings=gloabl.xml", params[0])
	})

	t.Run("Project and Global Settings file", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", BuildCommand: "mvn clean install", ProjectSettingsFile: "test.xml", GlobalSettingsFile: "global.xml"}
		params := getMavenSettings(&config)
		assert.Equal(t, 2, len(params))
		assert.Equal(t, "--settings=test.xml", params[0])
		assert.Equal(t, "--global-settings=global.xml", params[1])
	})

	t.Run("Skip incase of https url", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", BuildCommand: "mvn clean install", ProjectSettingsFile: "https://jenkins-sap-test.com/test.xml"}
		params := getMavenSettings(&config)
		assert.Equal(t, 0, len(params))
	})

	t.Run("Skip incase of http url", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", BuildCommand: "mvn clean install", ProjectSettingsFile: "http://jenkins-sap-test.com/test.xml"}
		params := getMavenSettings(&config)
		assert.Equal(t, 0, len(params))
	})
}

type CodeqlSarifUploaderMock struct {
	counter int
}

func (c *CodeqlSarifUploaderMock) GetSarifStatus() (codeql.SarifFileInfo, error) {
	if c.counter == 0 {
		return codeql.SarifFileInfo{
			ProcessingStatus: "complete",
			Errors:           nil,
		}, nil
	}
	if c.counter == -1 {
		return codeql.SarifFileInfo{
			ProcessingStatus: "failed",
			Errors:           []string{"upload error"},
		}, nil
	}
	c.counter--
	return codeql.SarifFileInfo{
		ProcessingStatus: "pending",
		Errors:           nil,
	}, nil
}

type CodeqlSarifUploaderErrorMock struct {
	counter int
}

func (c *CodeqlSarifUploaderErrorMock) GetSarifStatus() (codeql.SarifFileInfo, error) {
	if c.counter == -1 {
		return codeql.SarifFileInfo{}, errors.New("test error")
	}
	if c.counter == 0 {
		return codeql.SarifFileInfo{
			ProcessingStatus: "complete",
			Errors:           nil,
		}, nil
	}
	c.counter--
	return codeql.SarifFileInfo{ProcessingStatus: "Service unavailable"}, nil
}
