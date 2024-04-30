//go:build unit
// +build unit

package cmd

import (
	"bytes"
	"context"
	"fmt"

	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	bd "github.com/SAP/jenkins-library/pkg/blackduck"
	piperDocker "github.com/SAP/jenkins-library/pkg/docker"
	piperGithub "github.com/SAP/jenkins-library/pkg/github"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/orchestrator"

	"github.com/google/go-github/v45/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type detectTestUtilsBundle struct {
	expectedError   error
	downloadedFiles map[string]string // src, dest
	*mock.ShellMockRunner
	*mock.FilesMock
	customEnv    []string
	orchestrator *orchestratorConfigProviderMock
	dClient      *mock.DownloadMock
}

func (d *detectTestUtilsBundle) GetProvider() orchestrator.ConfigProvider {
	return d.orchestrator
}

func (d *detectTestUtilsBundle) GetIssueService() *github.IssuesService {
	return nil
}

func (d *detectTestUtilsBundle) GetSearchService() *github.SearchService {
	return nil
}

func (d *detectTestUtilsBundle) GetDockerClient(options piperDocker.ClientOptions) piperDocker.Download {
	return d.dClient
}

type orchestratorConfigProviderMock struct {
	orchestrator.UnknownOrchestratorConfigProvider
	isPullRequest bool
}

func (o *orchestratorConfigProviderMock) IsPullRequest() bool {
	return o.isPullRequest
}

type httpMockClient struct {
	responseBodyForURL map[string]string
	errorMessageForURL map[string]string
	header             map[string]http.Header
}

func (c *httpMockClient) SetOptions(opts piperhttp.ClientOptions) {}
func (c *httpMockClient) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	c.header[url] = header
	response := http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte(""))),
	}

	if c.errorMessageForURL[url] != "" {
		response.StatusCode = 400
		return &response, fmt.Errorf(c.errorMessageForURL[url])
	}

	if c.responseBodyForURL[url] != "" {
		response.Body = io.NopCloser(bytes.NewReader([]byte(c.responseBodyForURL[url])))
		return &response, nil
	}

	return &response, nil
}

func newBlackduckMockSystem(config detectExecuteScanOptions) blackduckSystem {
	myTestClient := httpMockClient{
		responseBodyForURL: map[string]string{
			"https://my.blackduck.system/api/tokens/authenticate":                                                                               authContent,
			"https://my.blackduck.system/api/projects?q=name%3ASHC-PiperTest":                                                                   projectContent,
			"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions?limit=100&offset=0":                         projectVersionContent,
			"https://my.blackduck.system/api/projects/5ca86e11/versions/a6c94786/components?limit=999&offset=0":                                 componentsContent,
			"https://my.blackduck.system/api/projects/5ca86e11/versions/a6c94786/vunlerable-bom-components?limit=999&offset=0":                  vulnerabilitiesContent,
			"https://my.blackduck.system/api/projects/5ca86e11/versions/a6c94786/components?filter=policyCategory%3Alicense&limit=999&offset=0": componentsContent,
			"https://my.blackduck.system/api/projects/5ca86e11/versions/a6c94786/policy-status":                                                 policyStatusContent,
			"https://my.blackduck.system/api/projects?q=name%3ARapid_scan_on_PRs":                                                               projectContentRapidScan,
			"https://my.blackduck.system/api/projects/654ggfdgf-1983-4e7b-97d4-eb1a0aeffbbf/versions?limit=100&offset=0":                        projectVersionContentRapid,
		},
		header: map[string]http.Header{},
	}
	sys := blackduckSystem{
		Client: bd.NewClient(config.Token, config.ServerURL, &myTestClient),
	}
	return sys
}

const (
	authContent = `{
        "bearerToken":"bearerTestToken",
        "expiresInMilliseconds":7199997
    }`
	projectContent = `{
        "totalCount": 1,
        "items": [
            {
                "name": "SHC-PiperTest",
                "_meta": {
                    "href": "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf",
                    "links": [
                        {
                            "rel": "versions",
                            "href": "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions"
                        }
                    ]
                }
            }
        ]
    }`
	projectVersionContent = `{
        "totalCount": 1,
        "items": [
            {
                "versionName": "1.0",
                "_meta": {
                    "href": "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36",
                    "links": [
                        {
                            "rel": "components",
                            "href": "https://my.blackduck.system/api/projects/5ca86e11/versions/a6c94786/components"
                        },
                        {
                            "rel": "vulnerable-components",
                            "href": "https://my.blackduck.system/api/projects/5ca86e11/versions/a6c94786/vunlerable-bom-components"
                        },
                        {
                            "rel": "policy-status",
                            "href": "https://my.blackduck.system/api/projects/5ca86e11/versions/a6c94786/policy-status"
                        }
                    ]
                }
            }
        ]
    }`
	componentsContent = `{
        "totalCount": 3,
        "items" : [
            {
                "componentName": "Spring Framework",
                "componentVersionName": "5.3.9",
                "policyStatus": "IN_VIOLATION"
            }, {
                "componentName": "Apache Tomcat",
                "componentVersionName": "9.0.52",
                "policyStatus": "IN_VIOLATION"
            }, {
                "componentName": "Apache Log4j",
                "componentVersionName": "4.5.16",
                "policyStatus": "UNKNOWN"
            }
        ]
    }`
	vulnerabilitiesContent = `{
        "totalCount": 3,
        "items": [
            {
                "componentName": "Spring Framework",
                "componentVersionName": "5.3.9",
                "vulnerabilityWithRemediation" : {
                    "vulnerabilityName" : "BDSA-2019-2021",
                    "baseScore" : 7.5,
                    "overallScore" : 7.5,
                    "severity" : "HIGH",
                    "remediationStatus" : "IGNORED",
                    "description" : "description"
                }
            }, {
                "componentName": "Apache Log4j",
                "componentVersionName": "4.5.16",
                "vulnerabilityWithRemediation" : {
                    "vulnerabilityName" : "BDSA-2020-4711",
                    "baseScore" : 7.5,
                    "overallScore" : 7.5,
                    "severity" : "HIGH",
                    "remediationStatus" : "IGNORED",
                    "description" : "description"
                }
            }, {
                "componentName": "Apache Log4j",
                "componentVersionName": "4.5.16",
                "vulnerabilityWithRemediation" : {
                    "vulnerabilityName" : "BDSA-2020-4712",
                    "baseScore" : 4.5,
                    "overallScore" : 4.5,
                    "severity" : "MEDIUM",
                    "remediationStatus" : "IGNORED",
                    "description" : "description"
                }
            }
        ]
    }`
	policyStatusContent = `{
        "overallStatus": "IN_VIOLATION",
        "componentVersionPolicyViolationDetails": {
            "name": "IN_VIOLATION",
            "severityLevels": [{"name":"BLOCKER", "value": 1}, {"name": "CRITICAL", "value": 1}]
        }
    }`
	projectContentRapidScan = `{
        "totalCount": 1,
        "items": [
            {
                "name": "Rapid_scan_on_PRs",
                "_meta": {
                    "href": "https://my.blackduck.system/api/projects/654ggfdgf-1983-4e7b-97d4-eb1a0aeffbbf",
                    "links": [
                        {
                            "rel": "versions",
                            "href": "https://my.blackduck.system/api/projects/654ggfdgf-1983-4e7b-97d4-eb1a0aeffbbf/versions"
                        }
                    ]
                }
            }
        ]
    }`
	projectVersionContentRapid = `{
        "totalCount": 1,
        "items": [
            {
                "versionName": "1.0",
                "_meta": {
                    "href": "https://my.blackduck.system/api/projects/654ggfdgf-1983-4e7b-97d4-eb1a0aeffbbf/versions/54357fds-0ee6-414f-9054-90d549c69c36",
                    "links": [
                        {
                            "rel": "components",
                            "href": "https://my.blackduck.system/api/projects/5ca86e11/versions/654784382/components"
                        },
                        {
                            "rel": "vulnerable-components",
                            "href": "https://my.blackduck.system/api/projects/5ca86e11/versions/654784382/vunlerable-bom-components"
                        },
                        {
                            "rel": "policy-status",
                            "href": "https://my.blackduck.system/api/projects/5ca86e11/versions/654784382/policy-status"
                        }
                    ]
                }
            }
        ]
    }`
)

func (c *detectTestUtilsBundle) RunExecutable(string, ...string) error {
	panic("not expected to be called in test")
}

func (c *detectTestUtilsBundle) SetOptions(piperhttp.ClientOptions) {
}

func (c *detectTestUtilsBundle) GetOsEnv() []string {
	return c.customEnv
}

func (c *detectTestUtilsBundle) DownloadFile(url, filename string, _ http.Header, _ []*http.Cookie) error {
	if c.expectedError != nil {
		return c.expectedError
	}

	if c.downloadedFiles == nil {
		c.downloadedFiles = make(map[string]string)
	}
	c.downloadedFiles[url] = filename
	return nil
}

func (w *detectTestUtilsBundle) CreateIssue(ghCreateIssueOptions *piperGithub.CreateIssueOptions) error {
	return nil
}

func newDetectTestUtilsBundle(isPullRequest bool) *detectTestUtilsBundle {
	utilsBundle := detectTestUtilsBundle{
		ShellMockRunner: &mock.ShellMockRunner{},
		FilesMock:       &mock.FilesMock{},
		orchestrator:    &orchestratorConfigProviderMock{isPullRequest: isPullRequest},
		dClient:         &mock.DownloadMock{},
	}
	return &utilsBundle
}

func TestRunDetect(t *testing.T) {
	t.Parallel()
	t.Run("success case", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		utilsMock := newDetectTestUtilsBundle(false)
		utilsMock.AddFile("detect.sh", []byte(""))
		err := runDetect(ctx, detectExecuteScanOptions{}, utilsMock, &detectExecuteScanInflux{})

		assert.Equal(t, utilsMock.downloadedFiles["https://detect.synopsys.com/detect8.sh"], "detect.sh")
		assert.True(t, utilsMock.HasRemovedFile("detect.sh"))
		assert.NoError(t, err)
		assert.Equal(t, ".", utilsMock.Dir, "Wrong execution directory used")
		assert.Equal(t, "/bin/bash", utilsMock.Shell[0], "Bash shell expected")
		expectedScript := "./detect.sh --blackduck.url= --blackduck.api.token= \"--detect.project.name=\" \"--detect.project.version.name=\" \"--detect.code.location.name=\" \"--detect.force.success.on.skip=true\" --detect.source.path='.'"
		assert.Equal(t, expectedScript, utilsMock.Calls[0])
	})

	t.Run("failure case", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		utilsMock := newDetectTestUtilsBundle(false)
		utilsMock.ShouldFailOnCommand = map[string]error{"./detect.sh --blackduck.url= --blackduck.api.token= \"--detect.project.name=\" \"--detect.project.version.name=\" \"--detect.code.location.name=\" \"--detect.force.success.on.skip=true\" --detect.source.path='.'": fmt.Errorf("")}
		utilsMock.ExitCode = 3
		utilsMock.AddFile("detect.sh", []byte(""))
		err := runDetect(ctx, detectExecuteScanOptions{FailOnSevereVulnerabilities: true}, utilsMock, &detectExecuteScanInflux{})
		assert.Equal(t, utilsMock.ExitCode, 3)
		assert.Contains(t, err.Error(), "FAILURE_POLICY_VIOLATION => Detect found policy violations.")
		assert.True(t, utilsMock.HasRemovedFile("detect.sh"))
	})

	t.Run("maven parameters", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		utilsMock := newDetectTestUtilsBundle(false)
		utilsMock.CurrentDir = "root_folder"
		utilsMock.AddFile("detect.sh", []byte(""))
		err := runDetect(ctx, detectExecuteScanOptions{
			M2Path:              ".pipeline/local_repo",
			ProjectSettingsFile: "project-settings.xml",
			GlobalSettingsFile:  "global-settings.xml",
		}, utilsMock, &detectExecuteScanInflux{})

		assert.NoError(t, err)
		assert.Equal(t, ".", utilsMock.Dir, "Wrong execution directory used")
		assert.Equal(t, "/bin/bash", utilsMock.Shell[0], "Bash shell expected")
		absoluteLocalPath := string(os.PathSeparator) + filepath.Join("root_folder", ".pipeline", "local_repo")

		expectedParam := "\"--detect.maven.build.command=--global-settings global-settings.xml --settings project-settings.xml -Dmaven.repo.local=" + absoluteLocalPath + "\""
		assert.Contains(t, utilsMock.Calls[0], expectedParam)
	})

	t.Run("images scan", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		utilsMock := newDetectTestUtilsBundle(false)
		utilsMock.CurrentDir = "root_folder"
		utilsMock.AddFile("detect.sh", []byte(""))
		err := runDetect(ctx, detectExecuteScanOptions{
			ScanContainerDistro: "ubuntu",
			ImageNameTags:       []string{"foo/bar:latest", "bar/bazz:latest"},
		}, utilsMock, &detectExecuteScanInflux{})

		assert.NoError(t, err)
		assert.Equal(t, ".", utilsMock.Dir, "Wrong execution directory used")
		require.Equal(t, 3, len(utilsMock.Calls))

		expectedParam1 := "--detect.docker.tar=./foo_bar_latest.tar --detect.target.type=IMAGE --detect.tools.excluded=DETECTOR --detect.docker.passthrough.shared.dir.path.local=/opt/blackduck/blackduck-imageinspector/shared/ --detect.docker.passthrough.shared.dir.path.imageinspector=/opt/blackduck/blackduck-imageinspector/shared --detect.docker.passthrough.imageinspector.service.distro.default=ubuntu --detect.docker.passthrough.imageinspector.service.start=false --detect.docker.passthrough.output.include.squashedimage=false --detect.docker.passthrough.imageinspector.service.url=http://localhost:8082"
		assert.Contains(t, utilsMock.Calls[1], expectedParam1)

		expectedParam2 := "--detect.docker.tar=./bar_bazz_latest.tar --detect.target.type=IMAGE --detect.tools.excluded=DETECTOR --detect.docker.passthrough.shared.dir.path.local=/opt/blackduck/blackduck-imageinspector/shared/ --detect.docker.passthrough.shared.dir.path.imageinspector=/opt/blackduck/blackduck-imageinspector/shared --detect.docker.passthrough.imageinspector.service.distro.default=ubuntu --detect.docker.passthrough.imageinspector.service.start=false --detect.docker.passthrough.output.include.squashedimage=false --detect.docker.passthrough.imageinspector.service.url=http://localhost:8082"
		assert.Contains(t, utilsMock.Calls[2], expectedParam2)
	})
}

func TestAddDetectArgs(t *testing.T) {
	t.Parallel()
	testData := []struct {
		args          []string
		options       detectExecuteScanOptions
		isPullRequest bool
		expected      []string
	}{
		{
			args: []string{"--testProp1=1"},
			options: detectExecuteScanOptions{
				BuildTool:           "mta",
				ExcludedDirectories: []string{"dir1", "dir2"},
				ScanProperties:      []string{"--scan1=1", "--scan2=2"},
				ServerURL:           "https://server.url",
				Token:               "apiToken",
				ProjectName:         "testName",
				Version:             "1.0",
				VersioningModel:     "major-minor",
				CodeLocation:        "",
				Scanners:            []string{"signature"},
				ScanPaths:           []string{"path1", "path2"},
			},
			expected: []string{
				"--testProp1=1",
				"--detect.detector.search.depth=100",
				"--detect.detector.search.continue=true",
				"--detect.excluded.directories=dir1,dir2",
				"--scan1=1",
				"--scan2=2",
				"--blackduck.url=https://server.url",
				"--blackduck.api.token=apiToken",
				"\"--detect.project.name=testName\"",
				"\"--detect.project.version.name=1.0\"",
				"\"--detect.code.location.name=testName/1.0\"",
				"\"--detect.force.success.on.skip=true\"",
				"--detect.blackduck.signature.scanner.paths=path1,path2",
				"--detect.source.path='.'",
			},
		},
		{
			args: []string{"--testProp1=1"},
			options: detectExecuteScanOptions{
				ServerURL:       "https://server.url",
				Token:           "apiToken",
				ProjectName:     "testName",
				Version:         "1.0",
				VersioningModel: "major-minor",
				CodeLocation:    "testLocation",
				FailOn:          []string{"BLOCKER", "MAJOR"},
				Scanners:        []string{"source"},
				ScanPaths:       []string{"path1", "path2"},
				Groups:          []string{"testGroup"},
			},
			expected: []string{
				"--testProp1=1",
				"--blackduck.url=https://server.url",
				"--blackduck.api.token=apiToken",
				"\"--detect.project.name=testName\"",
				"\"--detect.project.version.name=1.0\"",
				"\"--detect.project.user.groups=testGroup\"",
				"--detect.policy.check.fail.on.severities=BLOCKER,MAJOR",
				"\"--detect.code.location.name=testLocation\"",
				"\"--detect.force.success.on.skip=true\"",
				"--detect.blackduck.signature.scanner.paths=path1,path2",
				"--detect.source.path='.'",
			},
		},
		{
			args: []string{"--testProp1=1"},
			options: detectExecuteScanOptions{
				ServerURL:       "https://server.url",
				Token:           "apiToken",
				ProjectName:     "testName",
				CodeLocation:    "testLocation",
				FailOn:          []string{"BLOCKER", "MAJOR"},
				Scanners:        []string{"source"},
				ScanPaths:       []string{"path1", "path2"},
				Groups:          []string{"testGroup", "testGroup2"},
				Version:         "1.0",
				VersioningModel: "major-minor",
			},
			expected: []string{
				"--testProp1=1",
				"--blackduck.url=https://server.url",
				"--blackduck.api.token=apiToken",
				"\"--detect.project.name=testName\"",
				"\"--detect.project.version.name=1.0\"",
				"\"--detect.project.user.groups=testGroup,testGroup2\"",
				"--detect.policy.check.fail.on.severities=BLOCKER,MAJOR",
				"\"--detect.code.location.name=testLocation\"",
				"\"--detect.force.success.on.skip=true\"",
				"--detect.blackduck.signature.scanner.paths=path1,path2",
				"--detect.source.path='.'",
			},
		},
		{
			args: []string{"--testProp1=1"},
			options: detectExecuteScanOptions{
				ServerURL:       "https://server.url",
				Token:           "apiToken",
				ProjectName:     "testName",
				CodeLocation:    "testLocation",
				FailOn:          []string{"BLOCKER", "MAJOR"},
				Scanners:        []string{"source"},
				ScanPaths:       []string{"path1", "path2"},
				Groups:          []string{"testGroup", "testGroup2"},
				Version:         "1.0",
				VersioningModel: "major-minor",
				DependencyPath:  "pathx",
			},
			expected: []string{
				"--testProp1=1",
				"--blackduck.url=https://server.url",
				"--blackduck.api.token=apiToken",
				"\"--detect.project.name=testName\"",
				"\"--detect.project.version.name=1.0\"",
				"\"--detect.project.user.groups=testGroup,testGroup2\"",
				"--detect.policy.check.fail.on.severities=BLOCKER,MAJOR",
				"\"--detect.code.location.name=testLocation\"",
				"\"--detect.force.success.on.skip=true\"",
				"--detect.blackduck.signature.scanner.paths=path1,path2",
				"--detect.source.path=pathx",
			},
		},
		{
			args: []string{"--testProp1=1"},
			options: detectExecuteScanOptions{
				ServerURL:       "https://server.url",
				Token:           "apiToken",
				ProjectName:     "testName",
				CodeLocation:    "testLocation",
				FailOn:          []string{"BLOCKER", "MAJOR"},
				Scanners:        []string{"source"},
				ScanPaths:       []string{"path1", "path2"},
				Groups:          []string{"testGroup", "testGroup2"},
				Version:         "1.0",
				VersioningModel: "major-minor",
				DependencyPath:  "pathx",
				Unmap:           true,
			},
			expected: []string{
				"--testProp1=1",
				"--detect.project.codelocation.unmap=true",
				"--blackduck.url=https://server.url",
				"--blackduck.api.token=apiToken",
				"\"--detect.project.name=testName\"",
				"\"--detect.project.version.name=1.0\"",
				"\"--detect.project.user.groups=testGroup,testGroup2\"",
				"--detect.policy.check.fail.on.severities=BLOCKER,MAJOR",
				"\"--detect.code.location.name=testLocation\"",
				"\"--detect.force.success.on.skip=true\"",
				"--detect.blackduck.signature.scanner.paths=path1,path2",
				"--detect.source.path=pathx",
			},
		},
		{
			args: []string{"--testProp1=1"},
			options: detectExecuteScanOptions{
				ServerURL:               "https://server.url",
				Token:                   "apiToken",
				ProjectName:             "testName",
				CodeLocation:            "testLocation",
				FailOn:                  []string{"BLOCKER", "MAJOR"},
				Scanners:                []string{"source"},
				ScanPaths:               []string{"path1", "path2"},
				Groups:                  []string{"testGroup", "testGroup2"},
				Version:                 "1.0",
				VersioningModel:         "major-minor",
				DependencyPath:          "pathx",
				Unmap:                   true,
				IncludedPackageManagers: []string{"maven", "GRADLE"},
				ExcludedPackageManagers: []string{"npm", "NUGET"},
				MavenExcludedScopes:     []string{"TEST", "compile"},
				DetectTools:             []string{"DETECTOR"},
			},
			expected: []string{
				"--testProp1=1",
				"--detect.project.codelocation.unmap=true",
				"--blackduck.url=https://server.url",
				"--blackduck.api.token=apiToken",
				"\"--detect.project.name=testName\"",
				"\"--detect.project.version.name=1.0\"",
				"\"--detect.project.user.groups=testGroup,testGroup2\"",
				"--detect.policy.check.fail.on.severities=BLOCKER,MAJOR",
				"\"--detect.code.location.name=testLocation\"",
				"\"--detect.force.success.on.skip=true\"",
				"--detect.blackduck.signature.scanner.paths=path1,path2",
				"--detect.source.path=pathx",
				"--detect.included.detector.types=MAVEN,GRADLE",
				"--detect.excluded.detector.types=NPM,NUGET",
				"--detect.maven.excluded.scopes=test,compile",
				"--detect.tools=DETECTOR",
			},
		},
		{
			args: []string{"--testProp1=1"},
			options: detectExecuteScanOptions{
				ServerURL:               "https://server.url",
				Token:                   "apiToken",
				ProjectName:             "testName",
				CodeLocation:            "testLocation",
				FailOn:                  []string{"BLOCKER", "MAJOR"},
				Scanners:                []string{"source"},
				ScanPaths:               []string{"path1", "path2"},
				Groups:                  []string{"testGroup", "testGroup2"},
				Version:                 "1.0",
				VersioningModel:         "major-minor",
				DependencyPath:          "pathx",
				Unmap:                   true,
				IncludedPackageManagers: []string{"maven", "GRADLE"},
				ExcludedPackageManagers: []string{"npm", "NUGET"},
				MavenExcludedScopes:     []string{"TEST", "compile"},
				DetectTools:             []string{"DETECTOR"},
			},
			expected: []string{
				"--testProp1=1",
				"--detect.project.codelocation.unmap=true",
				"--blackduck.url=https://server.url",
				"--blackduck.api.token=apiToken",
				"\"--detect.project.name=testName\"",
				"\"--detect.project.version.name=1.0\"",
				"\"--detect.project.user.groups=testGroup,testGroup2\"",
				"--detect.policy.check.fail.on.severities=BLOCKER,MAJOR",
				"\"--detect.code.location.name=testLocation\"",
				"\"--detect.force.success.on.skip=true\"",
				"--detect.blackduck.signature.scanner.paths=path1,path2",
				"--detect.source.path=pathx",
				"--detect.included.detector.types=MAVEN,GRADLE",
				"--detect.excluded.detector.types=NPM,NUGET",
				"--detect.maven.excluded.scopes=test,compile",
				"--detect.tools=DETECTOR",
			},
		},
		{
			args: []string{"--testProp1=1"},
			options: detectExecuteScanOptions{
				ServerURL:               "https://server.url",
				Token:                   "apiToken",
				ProjectName:             "testName",
				CodeLocation:            "testLocation",
				FailOn:                  []string{"BLOCKER", "MAJOR"},
				Scanners:                []string{"source"},
				ScanPaths:               []string{"path1", "path2"},
				Groups:                  []string{"testGroup", "testGroup2"},
				Version:                 "1.0",
				VersioningModel:         "major-minor",
				DependencyPath:          "pathx",
				Unmap:                   true,
				IncludedPackageManagers: []string{"maven", "GRADLE"},
				ExcludedPackageManagers: []string{"npm", "NUGET"},
				MavenExcludedScopes:     []string{"TEST", "compile"},
				DetectTools:             []string{"DETECTOR"},
			},
			expected: []string{
				"--testProp1=1",
				"--detect.project.codelocation.unmap=true",
				"--blackduck.url=https://server.url",
				"--blackduck.api.token=apiToken",
				"\"--detect.project.name=testName\"",
				"\"--detect.project.version.name=1.0\"",
				"\"--detect.project.user.groups=testGroup,testGroup2\"",
				"--detect.policy.check.fail.on.severities=BLOCKER,MAJOR",
				"\"--detect.code.location.name=testLocation\"",
				"\"--detect.force.success.on.skip=true\"",
				"--detect.blackduck.signature.scanner.paths=path1,path2",
				"--detect.source.path=pathx",
				"--detect.included.detector.types=MAVEN,GRADLE",
				"--detect.excluded.detector.types=NPM,NUGET",
				"--detect.maven.excluded.scopes=test,compile",
				"--detect.tools=DETECTOR",
			},
		},
		{
			args: []string{"--testProp1=1"},
			options: detectExecuteScanOptions{
				ScanProperties:          []string{"--scan=1", "--detect.project.codelocation.unmap=true"},
				ServerURL:               "https://server.url",
				Token:                   "apiToken",
				ProjectName:             "testName",
				CodeLocation:            "testLocation",
				FailOn:                  []string{"BLOCKER", "MAJOR"},
				Scanners:                []string{"source"},
				ScanPaths:               []string{"path1", "path2"},
				Groups:                  []string{"testGroup", "testGroup2"},
				Version:                 "1.0",
				VersioningModel:         "major-minor",
				DependencyPath:          "pathx",
				Unmap:                   true,
				IncludedPackageManagers: []string{"maven", "GRADLE"},
				ExcludedPackageManagers: []string{"npm", "NUGET"},
				MavenExcludedScopes:     []string{"TEST", "compile"},
				DetectTools:             []string{"DETECTOR"},
			},
			expected: []string{
				"--testProp1=1",
				"--scan=1",
				"--detect.project.codelocation.unmap=true",
				"--blackduck.url=https://server.url",
				"--blackduck.api.token=apiToken",
				"\"--detect.project.name=testName\"",
				"\"--detect.project.version.name=1.0\"",
				"\"--detect.project.user.groups=testGroup,testGroup2\"",
				"--detect.policy.check.fail.on.severities=BLOCKER,MAJOR",
				"\"--detect.code.location.name=testLocation\"",
				"\"--detect.force.success.on.skip=true\"",
				"--detect.blackduck.signature.scanner.paths=path1,path2",
				"--detect.source.path=pathx",
				"--detect.included.detector.types=MAVEN,GRADLE",
				"--detect.excluded.detector.types=NPM,NUGET",
				"--detect.maven.excluded.scopes=test,compile",
				"--detect.tools=DETECTOR",
			},
		},
		{
			args: []string{"--testProp1=1"},
			options: detectExecuteScanOptions{
				ServerURL:       "https://server.url",
				Token:           "apiToken",
				ProjectName:     "testName",
				Version:         "1.0",
				VersioningModel: "major-minor",
				CodeLocation:    "",
				ScanPaths:       []string{"path1", "path2"},
			},
			expected: []string{
				"--testProp1=1",
				"--blackduck.url=https://server.url",
				"--blackduck.api.token=apiToken",
				"\"--detect.project.name=testName\"",
				"\"--detect.project.version.name=1.0\"",
				"\"--detect.code.location.name=testName/1.0\"",
				"\"--detect.force.success.on.skip=true\"",
				"--detect.blackduck.signature.scanner.paths=path1,path2",
				"--detect.source.path='.'",
			},
		},
		{
			args: []string{"--testProp1=1"},
			options: detectExecuteScanOptions{
				ServerURL:         "https://server.url",
				Token:             "apiToken",
				ProjectName:       "Rapid_scan_on_PRs",
				Version:           "1.0",
				VersioningModel:   "major-minor",
				CodeLocation:      "",
				ScanPaths:         []string{"path1", "path2"},
				CustomScanVersion: "1.0",
			},
			isPullRequest: true,
			expected: []string{
				"--testProp1=1",
				"--blackduck.url=https://server.url",
				"--blackduck.api.token=apiToken",
				"\"--detect.project.name=Rapid_scan_on_PRs\"",
				"\"--detect.project.version.name=1.0\"",
				"\"--detect.code.location.name=Rapid_scan_on_PRs/1.0\"",
				"\"--detect.force.success.on.skip=true\"",
				"--detect.blackduck.signature.scanner.paths=path1,path2",
				"--detect.source.path='.'",
				"--detect.blackduck.scan.mode='RAPID'",
				"--detect.blackduck.rapid.compare.mode='BOM_COMPARE_STRICT'",
				"--detect.cleanup=false",
				"--detect.output.path='report'",
			},
		},
		{
			args: []string{"--testProp1=1"},
			options: detectExecuteScanOptions{
				ServerURL:       "https://server.url",
				BuildTool:       "mta",
				Token:           "apiToken",
				ProjectName:     "Rapid_scan_on_PRs",
				Version:         "2.0",
				VersioningModel: "major-minor",
				CodeLocation:    "",
				ScanPaths:       []string{"path1", "path2"},
				ScanProperties: []string{
					"--detect.detector.search.depth=5",
					"--detect.detector.search.continue=false",
					"--detect.excluded.directories=dir1,dir2",
				},
				ExcludedDirectories: []string{"dir3,dir4"},
				CustomScanVersion:   "2.0",
			},
			isPullRequest: true,
			expected: []string{
				"--testProp1=1",
				"--detect.detector.search.depth=5",
				"--detect.detector.search.continue=false",
				"--detect.excluded.directories=dir1,dir2",
				"--blackduck.url=https://server.url",
				"--blackduck.api.token=apiToken",
				"\"--detect.project.name=Rapid_scan_on_PRs\"",
				"\"--detect.project.version.name=2.0\"",
				"\"--detect.code.location.name=Rapid_scan_on_PRs/2.0\"",
				"\"--detect.force.success.on.skip=true\"",
				"--detect.blackduck.signature.scanner.paths=path1,path2",
				"--detect.source.path='.'",
				"--detect.blackduck.scan.mode='RAPID'",
				"--detect.cleanup=false",
				"--detect.output.path='report'",
			},
		},
		{
			args: []string{"--testProp1=1"},
			options: detectExecuteScanOptions{
				ServerURL:          "https://server.url",
				BuildTool:          "maven",
				Token:              "apiToken",
				ProjectName:        "Rapid_scan_on_PRs",
				Version:            "2.0",
				VersioningModel:    "major-minor",
				CodeLocation:       "",
				ScanPaths:          []string{"path1", "path2"},
				M2Path:             "./m2",
				GlobalSettingsFile: "pipeline/settings.xml",
				ScanProperties: []string{
					"--detect.maven.build.command= --settings .pipeline/settings.xml -DskipTests install",
				},
				CustomScanVersion: "2.0",
			},
			isPullRequest: true,
			expected: []string{
				"--testProp1=1",
				"--detect.maven.build.command=",
				"--settings",
				".pipeline/settings.xml",
				"-DskipTests",
				"install",
				"--blackduck.url=https://server.url",
				"--blackduck.api.token=apiToken",
				"\"--detect.project.name=Rapid_scan_on_PRs\"",
				"\"--detect.project.version.name=2.0\"",
				"\"--detect.code.location.name=Rapid_scan_on_PRs/2.0\"",
				"\"--detect.force.success.on.skip=true\"",
				"--detect.blackduck.signature.scanner.paths=path1,path2",
				"--detect.source.path='.'",
				"--detect.blackduck.scan.mode='RAPID'",
				"--detect.cleanup=false",
				"--detect.output.path='report'",
			},
		},
	}

	for k, v := range testData {
		v := v
		t.Run(fmt.Sprintf("run %v", k), func(t *testing.T) {
			t.Parallel()

			config := detectExecuteScanOptions{Token: "token", ServerURL: "https://my.blackduck.system", ProjectName: v.options.ProjectName, Version: v.options.Version, CustomScanVersion: v.options.CustomScanVersion}
			sys := newBlackduckMockSystem(config)

			got, err := addDetectArgs(v.args, v.options, newDetectTestUtilsBundle(v.isPullRequest), &sys, NO_VERSION_SUFFIX, "")
			assert.NoError(t, err)
			assert.Equal(t, v.expected, got)
		})
	}
}

// Testing exit code mapping method
func TestExitCodeMapping(t *testing.T) {
	cases := []struct {
		exitCode int
		expected string
	}{
		{1, "FAILURE_BLACKDUCK_CONNECTIVITY"},
		{-1, "Not known exit code key"},
		{8, "Not known exit code key"},
		{100, "FAILURE_UNKNOWN_ERROR"},
	}

	for _, c := range cases {
		response := exitCodeMapping(c.exitCode)
		assert.Contains(t, response, c.expected)
	}
}

func TestPostScanChecksAndReporting(t *testing.T) {
	t.Parallel()
	t.Run("Reporting after scan", func(t *testing.T) {
		ctx := context.Background()
		config := detectExecuteScanOptions{Token: "token", ServerURL: "https://my.blackduck.system", ProjectName: "SHC-PiperTest", Version: "", CustomScanVersion: "1.0"}
		utils := newDetectTestUtilsBundle(false)
		sys := newBlackduckMockSystem(config)
		err := postScanChecksAndReporting(ctx, config, &detectExecuteScanInflux{}, utils, &sys)

		assert.EqualError(t, err, "License Policy Violations found")
		content, err := utils.FileRead("blackduck-ip.json")
		assert.NoError(t, err)
		assert.Contains(t, string(content), `"policyViolations":2`)
	})
}

func TestIsMajorVulnerability(t *testing.T) {
	t.Parallel()
	t.Run("Case True", func(t *testing.T) {
		vr := bd.VulnerabilityWithRemediation{
			OverallScore: 7.5,
			Severity:     "HIGH",
		}
		v := bd.Vulnerability{
			Name:                         "",
			VulnerabilityWithRemediation: vr,
			Ignored:                      false,
		}
		assert.True(t, isMajorVulnerability(v))
	})
	t.Run("Case Ignored Components", func(t *testing.T) {
		vr := bd.VulnerabilityWithRemediation{
			OverallScore: 7.5,
			Severity:     "HIGH",
		}
		v := bd.Vulnerability{
			Name:                         "",
			VulnerabilityWithRemediation: vr,
			Ignored:                      true,
		}
		assert.False(t, isMajorVulnerability(v))
	})
	t.Run("Case False", func(t *testing.T) {
		vr := bd.VulnerabilityWithRemediation{
			OverallScore: 7.5,
			Severity:     "MEDIUM",
		}
		v := bd.Vulnerability{
			Name:                         "",
			VulnerabilityWithRemediation: vr,
			Ignored:                      false,
		}
		assert.False(t, isMajorVulnerability(v))
	})
}

func TestIsActiveVulnerability(t *testing.T) {
	t.Parallel()
	t.Run("Case true", func(t *testing.T) {
		vr := bd.VulnerabilityWithRemediation{
			OverallScore:      7.5,
			Severity:          "HIGH",
			RemediationStatus: "NEW",
		}
		v := bd.Vulnerability{
			Name:                         "",
			VulnerabilityWithRemediation: vr,
		}
		assert.True(t, isActiveVulnerability(v))
	})
	t.Run("Case False", func(t *testing.T) {
		vr := bd.VulnerabilityWithRemediation{
			OverallScore:      7.5,
			Severity:          "HIGH",
			RemediationStatus: "IGNORED",
		}
		v := bd.Vulnerability{
			Name:                         "",
			VulnerabilityWithRemediation: vr,
		}
		assert.False(t, isActiveVulnerability(v))
	})
}

func TestIsActivePolicyViolation(t *testing.T) {
	t.Parallel()
	t.Run("Case true", func(t *testing.T) {
		assert.True(t, isActivePolicyViolation("IN_VIOLATION"))
	})
	t.Run("Case False", func(t *testing.T) {
		assert.False(t, isActivePolicyViolation("NOT_IN_VIOLATION"))
	})
}

func TestGetActivePolicyViolations(t *testing.T) {
	t.Parallel()
	t.Run("Case true", func(t *testing.T) {
		config := detectExecuteScanOptions{Token: "token", ServerURL: "https://my.blackduck.system", ProjectName: "SHC-PiperTest", Version: "", CustomScanVersion: "1.0"}
		sys := newBlackduckMockSystem(config)

		components, err := sys.Client.GetComponents("SHC-PiperTest", "1.0")
		assert.NoError(t, err)
		assert.Equal(t, 2, getActivePolicyViolations(components))
	})
}

func TestGetVulnerabilitiesWithComponents(t *testing.T) {
	t.Parallel()
	t.Run("Case true", func(t *testing.T) {
		config := detectExecuteScanOptions{Token: "token", ServerURL: "https://my.blackduck.system", ProjectName: "SHC-PiperTest", Version: "", CustomScanVersion: "1.0"}
		sys := newBlackduckMockSystem(config)

		vulns, err := getVulnerabilitiesWithComponents(config, &detectExecuteScanInflux{}, &sys)
		assert.NoError(t, err)
		vulnerabilitySpring := bd.Vulnerability{}
		vulnerabilityLog4j1 := bd.Vulnerability{}
		vulnerabilityLog4j2 := bd.Vulnerability{}
		for _, v := range vulns.Items {
			if v.VulnerabilityWithRemediation.VulnerabilityName == "BDSA-2019-2021" {
				vulnerabilitySpring = v
			}
			if v.VulnerabilityWithRemediation.VulnerabilityName == "BDSA-2020-4711" {
				vulnerabilityLog4j1 = v
			}
			if v.VulnerabilityWithRemediation.VulnerabilityName == "BDSA-2020-4712" {
				vulnerabilityLog4j2 = v
			}
		}
		vulnerableComponentSpring := &bd.Component{}
		vulnerableComponentLog4j := &bd.Component{}
		for i := 0; i < len(vulns.Items); i++ {
			if vulns.Items[i].Component != nil && vulns.Items[i].Component.Name == "Spring Framework" {
				vulnerableComponentSpring = vulns.Items[i].Component
			}
			if vulns.Items[i].Component != nil && vulns.Items[i].Component.Name == "Apache Log4j" {
				vulnerableComponentLog4j = vulns.Items[i].Component
			}
		}
		assert.Equal(t, vulnerableComponentSpring, vulnerabilitySpring.Component)
		assert.Equal(t, vulnerableComponentLog4j, vulnerabilityLog4j1.Component)
		assert.Equal(t, vulnerableComponentLog4j, vulnerabilityLog4j2.Component)
	})
}
