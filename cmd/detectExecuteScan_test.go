package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	bd "github.com/SAP/jenkins-library/pkg/blackduck"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"

	"github.com/stretchr/testify/assert"
)

type detectTestUtilsBundle struct {
	expectedError   error
	downloadedFiles map[string]string // src, dest
	*mock.ShellMockRunner
	*mock.FilesMock
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
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
	}

	if c.errorMessageForURL[url] != "" {
		response.StatusCode = 400
		return &response, fmt.Errorf(c.errorMessageForURL[url])
	}

	if c.responseBodyForURL[url] != "" {
		response.Body = ioutil.NopCloser(bytes.NewReader([]byte(c.responseBodyForURL[url])))
		return &response, nil
	}

	return &response, nil
}

func newBlackduckMockSystem(config detectExecuteScanOptions) blackduckSystem {
	myTestClient := httpMockClient{
		responseBodyForURL: map[string]string{
			"https://my.blackduck.system/api/tokens/authenticate":                                                                               authContent,
			"https://my.blackduck.system/api/projects?q=name%3ASHC-PiperTest":                                                                   projectContent,
			"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions":                                            projectVersionContent,
			"https://my.blackduck.system/api/projects/5ca86e11/versions/a6c94786/components?limit=999&offset=0":                                 componentsContent,
			"https://my.blackduck.system/api/projects/5ca86e11/versions/a6c94786/vunlerable-bom-components?limit=999&offset=0":                  vulnerabilitiesContent,
			"https://my.blackduck.system/api/projects/5ca86e11/versions/a6c94786/components?filter=policyCategory%3Alicense&limit=999&offset=0": componentsContent,
			"https://my.blackduck.system/api/projects/5ca86e11/versions/a6c94786/policy-status":                                                 policyStatusContent,
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
		"totalCount": 2,
		"items" : [
			{
				"componentName": "Spring Framework",
				"componentVersionName": "5.3.9",
				"policyStatus": "IN_VIOLATION"
			}, {
				"componentName": "Apache Tomcat",
				"componentVersionName": "9.0.52",
				"policyStatus": "IN_VIOLATION"
			}
		]
	}`
	vulnerabilitiesContent = `{
		"totalCount": 1,
		"items": [
			{
				"componentName": "Spring Framework",
				"componentVersionName": "5.3.2",
				"vulnerabilityWithRemediation" : {
					"vulnerabilityName" : "BDSA-2019-2021",
					"baseScore" : 7.5,
					"overallScore" : 7.5,
					"severity" : "HIGH",
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
)

func (c *detectTestUtilsBundle) RunExecutable(string, ...string) error {
	panic("not expected to be called in test")
}

func (c *detectTestUtilsBundle) SetOptions(piperhttp.ClientOptions) {

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

func newDetectTestUtilsBundle() *detectTestUtilsBundle {
	utilsBundle := detectTestUtilsBundle{
		ShellMockRunner: &mock.ShellMockRunner{},
		FilesMock:       &mock.FilesMock{},
	}
	return &utilsBundle
}

func TestRunDetect(t *testing.T) {
	t.Parallel()
	t.Run("success case", func(t *testing.T) {
		t.Parallel()
		utilsMock := newDetectTestUtilsBundle()
		utilsMock.AddFile("detect.sh", []byte(""))
		err := runDetect(detectExecuteScanOptions{}, utilsMock, &detectExecuteScanInflux{})

		assert.Equal(t, utilsMock.downloadedFiles["https://detect.synopsys.com/detect7.sh"], "detect.sh")
		assert.True(t, utilsMock.HasRemovedFile("detect.sh"))
		assert.NoError(t, err)
		assert.Equal(t, ".", utilsMock.Dir, "Wrong execution directory used")
		assert.Equal(t, "/bin/bash", utilsMock.Shell[0], "Bash shell expected")
		expectedScript := "./detect.sh --blackduck.url= --blackduck.api.token= \"--detect.project.name=''\" \"--detect.project.version.name=''\" \"--detect.code.location.name=''\" --detect.source.path='.'"
		assert.Equal(t, expectedScript, utilsMock.Calls[0])
	})

	t.Run("success case detect 6", func(t *testing.T) {
		t.Parallel()
		utilsMock := newDetectTestUtilsBundle()
		utilsMock.AddFile("detect.sh", []byte(""))
		options := detectExecuteScanOptions{
			CustomEnvironmentVariables: []string{"DETECT_LATEST_RELEASE_VERSION=6.8.0"},
		}
		err := runDetect(options, utilsMock, &detectExecuteScanInflux{})

		assert.Equal(t, utilsMock.downloadedFiles["https://detect.synopsys.com/detect.sh"], "detect.sh")
		assert.True(t, utilsMock.HasRemovedFile("detect.sh"))
		assert.NoError(t, err)
		assert.Equal(t, ".", utilsMock.Dir, "Wrong execution directory used")
		assert.Equal(t, "/bin/bash", utilsMock.Shell[0], "Bash shell expected")
		expectedScript := "./detect.sh --blackduck.url= --blackduck.api.token= \"--detect.project.name=''\" \"--detect.project.version.name=''\" \"--detect.code.location.name=''\" --detect.source.path='.'"
		assert.Equal(t, expectedScript, utilsMock.Calls[0])
	})
	//Creation of report is covered in the test case for postScanChecks
	/*
		t.Run("success case - with report", func(t *testing.T) {
			t.Parallel()
			utilsMock := newDetectTestUtilsBundle()
			utilsMock.AddFile("detect.sh", []byte(""))
			utilsMock.AddFile("my_BlackDuck_RiskReport.pdf", []byte(""))
			err := runDetect(detectExecuteScanOptions{FailOn: []string{"BLOCKER"}}, utilsMock, &detectExecuteScanInflux{})

			assert.Equal(t, utilsMock.downloadedFiles["https://detect.synopsys.com/detect.sh"], "detect.sh")
			assert.True(t, utilsMock.HasRemovedFile("detect.sh"))
			assert.NoError(t, err)
			assert.Equal(t, ".", utilsMock.Dir, "Wrong execution directory used")
			assert.Equal(t, "/bin/bash", utilsMock.Shell[0], "Bash shell expected")
			expectedScript := "./detect.sh --blackduck.url= --blackduck.api.token= \"--detect.project.name=''\" \"--detect.project.version.name=''\" --detect.policy.check.fail.on.severities=BLOCKER \"--detect.code.location.name=''\" --detect.source.path='.'"
			assert.Equal(t, expectedScript, utilsMock.Calls[0])

			content, err := utilsMock.FileRead("blackduck-ip.json")
			assert.NoError(t, err)
			assert.Contains(t, string(content), `"policyViolations":0`)
		})
	*/
	t.Run("failure case", func(t *testing.T) {
		t.Parallel()
		utilsMock := newDetectTestUtilsBundle()
		utilsMock.ShouldFailOnCommand = map[string]error{"./detect.sh --blackduck.url= --blackduck.api.token= \"--detect.project.name=''\" \"--detect.project.version.name=''\" \"--detect.code.location.name=''\" --detect.source.path='.'": fmt.Errorf("")}
		utilsMock.ExitCode = 3
		utilsMock.AddFile("detect.sh", []byte(""))
		err := runDetect(detectExecuteScanOptions{}, utilsMock, &detectExecuteScanInflux{})
		assert.Equal(t, utilsMock.ExitCode, 3)
		assert.Contains(t, err.Error(), "FAILURE_POLICY_VIOLATION => Detect found policy violations.")
		assert.True(t, utilsMock.HasRemovedFile("detect.sh"))
	})

	t.Run("maven parameters", func(t *testing.T) {
		t.Parallel()
		utilsMock := newDetectTestUtilsBundle()
		utilsMock.CurrentDir = "root_folder"
		utilsMock.AddFile("detect.sh", []byte(""))
		err := runDetect(detectExecuteScanOptions{
			M2Path:              ".pipeline/local_repo",
			ProjectSettingsFile: "project-settings.xml",
			GlobalSettingsFile:  "global-settings.xml",
		}, utilsMock, &detectExecuteScanInflux{})

		assert.NoError(t, err)
		assert.Equal(t, ".", utilsMock.Dir, "Wrong execution directory used")
		assert.Equal(t, "/bin/bash", utilsMock.Shell[0], "Bash shell expected")
		absoluteLocalPath := string(os.PathSeparator) + filepath.Join("root_folder", ".pipeline", "local_repo")

		expectedParam := "\"--detect.maven.build.command='--global-settings global-settings.xml --settings project-settings.xml -Dmaven.repo.local=" + absoluteLocalPath + "'\""
		assert.Contains(t, utilsMock.Calls[0], expectedParam)
	})
}

func TestAddDetectArgs(t *testing.T) {
	t.Parallel()
	testData := []struct {
		args     []string
		options  detectExecuteScanOptions
		expected []string
	}{
		{
			args: []string{"--testProp1=1"},
			options: detectExecuteScanOptions{
				ScanProperties:  []string{"--scan1=1", "--scan2=2"},
				ServerURL:       "https://server.url",
				Token:           "apiToken",
				ProjectName:     "testName",
				Version:         "1.0",
				VersioningModel: "major-minor",
				CodeLocation:    "",
				Scanners:        []string{"signature"},
				ScanPaths:       []string{"path1", "path2"},
			},
			expected: []string{
				"--testProp1=1",
				"--scan1=1",
				"--scan2=2",
				"--blackduck.url=https://server.url",
				"--blackduck.api.token=apiToken",
				"\"--detect.project.name='testName'\"",
				"\"--detect.project.version.name='1.0'\"",
				"\"--detect.code.location.name='testName/1.0'\"",
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
				"\"--detect.project.name='testName'\"",
				"\"--detect.project.version.name='1.0'\"",
				"\"--detect.project.user.groups='testGroup'\"",
				"--detect.policy.check.fail.on.severities=BLOCKER,MAJOR",
				"\"--detect.code.location.name='testLocation'\"",
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
				"\"--detect.project.name='testName'\"",
				"\"--detect.project.version.name='1.0'\"",
				"\"--detect.project.user.groups='testGroup,testGroup2'\"",
				"--detect.policy.check.fail.on.severities=BLOCKER,MAJOR",
				"\"--detect.code.location.name='testLocation'\"",
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
				"\"--detect.project.name='testName'\"",
				"\"--detect.project.version.name='1.0'\"",
				"\"--detect.project.user.groups='testGroup,testGroup2'\"",
				"--detect.policy.check.fail.on.severities=BLOCKER,MAJOR",
				"\"--detect.code.location.name='testLocation'\"",
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
				"\"--detect.project.name='testName'\"",
				"\"--detect.project.version.name='1.0'\"",
				"\"--detect.project.user.groups='testGroup,testGroup2'\"",
				"--detect.policy.check.fail.on.severities=BLOCKER,MAJOR",
				"\"--detect.code.location.name='testLocation'\"",
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
				"\"--detect.project.name='testName'\"",
				"\"--detect.project.version.name='1.0'\"",
				"\"--detect.project.user.groups='testGroup,testGroup2'\"",
				"--detect.policy.check.fail.on.severities=BLOCKER,MAJOR",
				"\"--detect.code.location.name='testLocation'\"",
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
				ScanOnChanges:           true,
			},
			expected: []string{
				"--testProp1=1",
				"--report",
				"--blackduck.url=https://server.url",
				"--blackduck.api.token=apiToken",
				"\"--detect.project.name='testName'\"",
				"\"--detect.project.version.name='1.0'\"",
				"\"--detect.project.user.groups='testGroup,testGroup2'\"",
				"--detect.policy.check.fail.on.severities=BLOCKER,MAJOR",
				"\"--detect.code.location.name='testLocation'\"",
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
				ScanOnChanges:           true,
			},
			expected: []string{
				"--testProp1=1",
				"--report",
				"--blackduck.url=https://server.url",
				"--blackduck.api.token=apiToken",
				"\"--detect.project.name='testName'\"",
				"\"--detect.project.version.name='1.0'\"",
				"\"--detect.project.user.groups='testGroup,testGroup2'\"",
				"--detect.policy.check.fail.on.severities=BLOCKER,MAJOR",
				"\"--detect.code.location.name='testLocation'\"",
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
				ScanOnChanges:           true,
			},
			expected: []string{
				"--testProp1=1",
				"--report",
				"--scan=1",
				"--blackduck.url=https://server.url",
				"--blackduck.api.token=apiToken",
				"\"--detect.project.name='testName'\"",
				"\"--detect.project.version.name='1.0'\"",
				"\"--detect.project.user.groups='testGroup,testGroup2'\"",
				"--detect.policy.check.fail.on.severities=BLOCKER,MAJOR",
				"\"--detect.code.location.name='testLocation'\"",
				"--detect.blackduck.signature.scanner.paths=path1,path2",
				"--detect.source.path=pathx",
				"--detect.included.detector.types=MAVEN,GRADLE",
				"--detect.excluded.detector.types=NPM,NUGET",
				"--detect.maven.excluded.scopes=test,compile",
				"--detect.tools=DETECTOR",
			},
		},
	}

	for k, v := range testData {
		v := v
		t.Run(fmt.Sprintf("run %v", k), func(t *testing.T) {
			t.Parallel()
			got, err := addDetectArgs(v.args, v.options, newDetectTestUtilsBundle())
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
		config := detectExecuteScanOptions{Token: "token", ServerURL: "https://my.blackduck.system", ProjectName: "SHC-PiperTest", Version: "", CustomScanVersion: "1.0"}
		utils := newDetectTestUtilsBundle()
		sys := newBlackduckMockSystem(config)
		err := postScanChecksAndReporting(config, &detectExecuteScanInflux{}, utils, &sys)

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
		}
		assert.True(t, isMajorVulnerability(v))
	})
	t.Run("Case False", func(t *testing.T) {
		vr := bd.VulnerabilityWithRemediation{
			OverallScore: 7.5,
			Severity:     "MEDIUM",
		}
		v := bd.Vulnerability{
			Name:                         "",
			VulnerabilityWithRemediation: vr,
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
		assert.Equal(t, getActivePolicyViolations(components), 2)
	})
}
