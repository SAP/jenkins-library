//go:build unit
// +build unit

package cmd

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/yaml"
	"github.com/stretchr/testify/assert"
)

type manifestMock struct {
	manifestFileName string
	apps             []map[string]interface{}
}

func (m manifestMock) GetAppName(index int) (string, error) {
	val, err := m.GetApplicationProperty(index, "name")
	if err != nil {
		return "", err
	}
	if v, ok := val.(string); ok {
		return v, nil
	}
	return "", fmt.Errorf("Cannot resolve application name")
}
func (m manifestMock) ApplicationHasProperty(index int, name string) (bool, error) {
	_, exists := m.apps[index][name]
	return exists, nil
}
func (m manifestMock) GetApplicationProperty(index int, name string) (interface{}, error) {
	return m.apps[index][name], nil
}
func (m manifestMock) GetFileName() string {
	return m.manifestFileName
}
func (m manifestMock) Transform() error {
	return nil
}
func (m manifestMock) IsModified() bool {
	return false
}
func (m manifestMock) GetApplications() ([]map[string]interface{}, error) {
	return m.apps, nil
}
func (m manifestMock) WriteManifest() error {
	return nil
}

// everything below in the config map annotated with '//default' is a default in the metadata
// since we don't get injected these values during the tests we set it here.
var defaultConfig = cloudFoundryDeployOptions{
	Org:                     "myOrg",
	Space:                   "mySpace",
	Username:                "me",
	Password:                "******",
	APIEndpoint:             "https://examples.sap.com/cf",
	Manifest:                "manifest.yml", // default
	MtaDeployParameters:     "-f",           // default
	DeployType:              "standard",     // default
	CheckMissingCredentials: true,           // default
}

func TestCfDeployment(t *testing.T) {

	defer func() {
		fileUtils = &piperutils.Files{}
		_replaceVariables = yaml.Substitute
	}()

	filesMock := mock.FilesMock{}
	filesMock.AddDir("/home/me")
	err := filesMock.Chdir("/home/me")
	assert.NoError(t, err)
	fileUtils = &filesMock

	config := defaultConfig

	successfulLogin := cloudfoundry.LoginOptions{
		CfAPIEndpoint: "https://examples.sap.com/cf",
		CfOrg:         "myOrg",
		CfSpace:       "mySpace",
		Username:      "me",
		Password:      "******",
		CfLoginOpts:   []string{},
	}

	var loginOpts cloudfoundry.LoginOptions
	var logoutCalled bool

	noopCfAPICalls := func(t *testing.T, s mock.ExecMockRunner) {
		assert.Empty(t, s.Calls)   // --> in case of an invalid deploy tool there must be no cf api calls
		assert.Empty(t, loginOpts) // no login options: login has not been called
		assert.False(t, logoutCalled)
	}

	prepareDefaultManifestMocking := func(manifestName string, appNames []string) func() {

		filesMock.AddFile(manifestName, []byte("file content does not matter"))

		apps := []map[string]interface{}{}

		for _, appName := range appNames {
			apps = append(apps, map[string]interface{}{"name": appName})
		}

		_getManifest = func(name string) (cloudfoundry.Manifest, error) {
			return manifestMock{
				manifestFileName: manifestName,
				apps:             apps,
			}, nil
		}

		return func() {
			_ = filesMock.FileRemove(manifestName) // slightly mis-use since that is intended to be used by code under test, not test code
			_getManifest = getManifest
		}
	}

	withLoginAndLogout := func(t *testing.T, asserts func(t *testing.T)) {
		assert.Equal(t, successfulLogin, loginOpts)
		asserts(t)
		assert.True(t, logoutCalled)
	}

	cleanup := func() {
		loginOpts = cloudfoundry.LoginOptions{}
		logoutCalled = false
		config = defaultConfig
	}

	defer func() {
		_cfLogin = cfLogin
		_cfLogout = cfLogout
	}()

	_cfLogin = func(c command.ExecRunner, opts cloudfoundry.LoginOptions) error {
		loginOpts = opts
		return nil
	}

	_cfLogout = func(c command.ExecRunner) error {
		logoutCalled = true
		return nil
	}

	_replaceVariables = func(manifest string, replacements map[string]interface{}, replacementsFiles []string) (bool, error) {
		return false, nil
	}

	t.Run("Test invalid appname", func(t *testing.T) {

		defer cleanup()
		config.AppName = "a_z"
		s := mock.ExecMockRunner{}
		err := runCloudFoundryDeploy(&config, nil, nil, &s)

		assert.EqualError(t, err, "Your application name 'a_z' contains a '_' (underscore) which is not allowed, only letters, dashes and numbers can be used. Please change the name to fit this requirement(s). For more details please visit https://docs.cloudfoundry.org/devguide/deploy-apps/deploy-app.html#basic-settings.")
	})

	t.Run("Invalid deploytool", func(t *testing.T) {

		defer cleanup()

		s := mock.ExecMockRunner{}

		config.DeployTool = "invalid"

		err := runCloudFoundryDeploy(&config, nil, nil, &s)

		if assert.NoError(t, err) {
			noopCfAPICalls(t, s)
		}
	})

	t.Run("deploytool cf native", func(t *testing.T) {

		defer cleanup()

		defer prepareDefaultManifestMocking("manifest.yml", []string{"testAppName"})()

		config.DeployTool = "cf_native"
		config.CfHome = "/home/me1"
		config.CfPluginHome = "/home/me2"

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check cf api calls", func(t *testing.T) {

				withLoginAndLogout(t, func(t *testing.T) {
					assert.Equal(t, []mock.ExecCall{
						{Exec: "cf", Params: []string{"version"}},
						{Exec: "cf", Params: []string{"plugins"}},
						{Exec: "cf", Params: []string{"push", "-f", "manifest.yml"}},
					}, s.Calls)
				})
			})

			t.Run("check environment variables", func(t *testing.T) {
				assert.Contains(t, s.Env, "CF_HOME=/home/me1")
				assert.Contains(t, s.Env, "CF_PLUGIN_HOME=/home/me2")
			})
		}
	})

	t.Run("influx reporting", func(t *testing.T) {

		defer cleanup()

		s := mock.ExecMockRunner{}

		defer func() {
			_now = time.Now
		}()

		_now = func() time.Time {
			// There was the big eclipse in Karlsruhe
			return time.Date(1999, time.August, 11, 12, 32, 0, 0, time.UTC)
		}

		defer prepareDefaultManifestMocking("manifest.yml", []string{"testAppName"})()

		config.DeployTool = "cf_native"
		config.ArtifactVersion = "0.1.2"
		config.CommitHash = "123456"

		influxData := cloudFoundryDeployInflux{}

		err := runCloudFoundryDeploy(&config, nil, &influxData, &s)

		if assert.NoError(t, err) {

			expected := cloudFoundryDeployInflux{}

			expected.deployment_data.fields.artifactURL = "n/a"
			expected.deployment_data.fields.deployTime = "AUG 11 1999 12:32:00"
			expected.deployment_data.fields.jobTrigger = "n/a"
			expected.deployment_data.fields.commitHash = "123456"

			expected.deployment_data.tags.artifactVersion = "0.1.2"
			expected.deployment_data.tags.deployUser = "me"
			expected.deployment_data.tags.deployResult = "SUCCESS"
			expected.deployment_data.tags.cfAPIEndpoint = "https://examples.sap.com/cf"
			expected.deployment_data.tags.cfOrg = "myOrg"
			expected.deployment_data.tags.cfSpace = "mySpace"

			assert.Equal(t, expected, influxData)

		}

	})

	t.Run("deploy cf native with docker image and docker username", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.DeployDockerImage = "repo/image:tag"
		config.DockerUsername = "me"
		config.AppName = "testAppName"

		config.Manifest = ""

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s)

		if assert.NoError(t, err) {

			withLoginAndLogout(t, func(t *testing.T) {
				assert.Equal(t, []mock.ExecCall{
					{Exec: "cf", Params: []string{"version"}},
					{Exec: "cf", Params: []string{"plugins"}},
					{Exec: "cf", Params: []string{"push",
						"testAppName",
						"--docker-image",
						"repo/image:tag",
						"--docker-username",
						"me",
					}},
				}, s.Calls)
			})
		}
	})

	t.Run("deploy_cf_native with manifest and docker credentials", func(t *testing.T) {

		defer cleanup()

		// Docker image can be done via manifest.yml.
		// if a private Docker registry is used, --docker-username and DOCKER_PASSWORD
		// must be set; this is checked by this test

		config.DeployTool = "cf_native"
		config.DeployDockerImage = "repo/image:tag"
		config.DockerUsername = "test_cf_docker"
		config.DockerPassword = "********"
		config.AppName = "testAppName"

		config.Manifest = ""

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s)

		if assert.NoError(t, err) {
			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, func(t *testing.T) {

					assert.Equal(t, []mock.ExecCall{
						{Exec: "cf", Params: []string{"version"}},
						{Exec: "cf", Params: []string{"plugins"}},
						{Exec: "cf", Params: []string{"push",
							"testAppName",
							"--docker-image",
							"repo/image:tag",
							"--docker-username",
							"test_cf_docker",
						}},
					}, s.Calls)
				})
			})

			t.Run("check environment variables", func(t *testing.T) {
				// REVISIT: in the corresponding groovy test we checked for "${'********'}"
				// I don't understand why, but we should discuss ...
				assert.Contains(t, s.Env, "CF_DOCKER_PASSWORD=********")
			})
		}
	})

	t.Run("deploy cf native app name from manifest", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.Manifest = "test-manifest.yml"

		// app name is not asserted since it does not appear in the cf calls
		// but it is checked that an app name is present, hence we need it here.
		defer prepareDefaultManifestMocking("test-manifest.yml", []string{"dummyApp"})()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, func(t *testing.T) {

					assert.Equal(t, []mock.ExecCall{
						{Exec: "cf", Params: []string{"version"}},
						{Exec: "cf", Params: []string{"plugins"}},
						{Exec: "cf", Params: []string{
							"push",
							"-f",
							"test-manifest.yml",
						}},
					}, s.Calls)

				})
			})
		}
	})

	t.Run("get app name from default manifest with cf native deployment", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.Manifest = ""
		config.AppName = ""

		// app name does not need to be set if it can be found in the manifest.yml
		// manifest name does not need to be set- the default manifest.yml will be used if not set
		defer prepareDefaultManifestMocking("manifest.yml", []string{"newAppName"})()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, func(t *testing.T) {

					assert.Equal(t, []mock.ExecCall{
						{Exec: "cf", Params: []string{"version"}},
						{Exec: "cf", Params: []string{"plugins"}},
						{Exec: "cf", Params: []string{
							"push",
						}},
					}, s.Calls)

				})
			})
		}
	})

	t.Run("cf native deploy fail when deployType is blue-green", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.DeployType = "blue-green"
		config.Manifest = ""
		config.AppName = ""

		// app name does not need to be set if it can be found in the manifest.yml
		// manifest name does not need to be set- the default manifest.yml will be used if not set
		defer prepareDefaultManifestMocking("manifest.yml", []string{"newAppName"})()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s)

		if assert.EqualError(t, err, "Blue-green deployment type is deprecated for cf native builds."+
			"Instead set parameter `cfNativeDeployParameters: '--strategy rolling'`. "+
			"Please refer to the Cloud Foundry documentation for further information: "+
			"https://docs.cloudfoundry.org/devguide/deploy-apps/rolling-deploy.html."+
			"Or alternatively, switch to mta build tool. Please refer to mta build tool"+
			"documentation for further information: https://sap.github.io/cloud-mta-build-tool/configuration/.") {

			t.Run("check shell calls", func(t *testing.T) {
				noopCfAPICalls(t, s)
			})
		}
	})

	t.Run("cf native deploy fail when unknown deployType is set", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.DeployType = "blue"
		config.Manifest = ""
		config.AppName = ""

		// app name does not need to be set if it can be found in the manifest.yml
		// manifest name does not need to be set- the default manifest.yml will be used if not set
		defer prepareDefaultManifestMocking("manifest.yml", []string{"newAppName"})()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s)

		if assert.EqualError(t, err, "Invalid deploy type received: 'blue'. Supported value: standard") {

			t.Run("check shell calls", func(t *testing.T) {
				noopCfAPICalls(t, s)
			})
		}
	})

	t.Run("deploy cf native without app name", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.Manifest = "test-manifest.yml"

		// Here we don't provide an application name from the mock. To make that
		// more explicit we provide the empty string default explicitly.
		defer prepareDefaultManifestMocking("test-manifest.yml", []string{""})()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s)

		if assert.EqualError(t, err, "appName from manifest 'test-manifest.yml' is empty") {

			t.Run("check shell calls", func(t *testing.T) {
				noopCfAPICalls(t, s)
			})
		}
	})

	t.Run("cf native deployment failure", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.DeployType = "standard"
		config.Manifest = "test-manifest.yml"
		config.AppName = "myTestApp"

		defer prepareDefaultManifestMocking("test-manifest.yml", []string{"app"})()

		s := mock.ExecMockRunner{}

		s.ShouldFailOnCommand = map[string]error{"cf.*push.*": fmt.Errorf("cf deploy failed")}
		err := runCloudFoundryDeploy(&config, nil, nil, &s)

		if assert.EqualError(t, err, "cf deploy failed") {
			t.Run("check shell calls", func(t *testing.T) {

				// we should try to logout in this case
				assert.True(t, logoutCalled)
			})
		}
	})

	t.Run("cf native deployment failure when logging in", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.DeployType = "standard"
		config.Manifest = "test-manifest.yml"
		config.AppName = "myTestApp"

		defer func() {

			_cfLogin = func(c command.ExecRunner, opts cloudfoundry.LoginOptions) error {
				loginOpts = opts
				return nil
			}
		}()

		_cfLogin = func(c command.ExecRunner, opts cloudfoundry.LoginOptions) error {
			loginOpts = opts
			return fmt.Errorf("Unable to login")
		}

		defer prepareDefaultManifestMocking("test-manifest.yml", []string{"app1"})()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s)

		if assert.EqualError(t, err, "Unable to login") {
			t.Run("check shell calls", func(t *testing.T) {

				// no calls to the cf client in this case
				assert.Equal(t,
					[]mock.ExecCall{
						{Exec: "cf", Params: []string{"version"}},
					}, s.Calls)
				// no logout
				assert.False(t, logoutCalled)
			})
		}
	})

	t.Run("cf native deploy standard should not stop instance", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.DeployType = "standard"
		config.Manifest = "test-manifest.yml"
		config.AppName = "myTestApp"
		config.KeepOldInstance = true

		defer prepareDefaultManifestMocking("test-manifest.yml", []string{"app"})()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, func(t *testing.T) {

					assert.Equal(t, []mock.ExecCall{
						{Exec: "cf", Params: []string{"version"}},
						{Exec: "cf", Params: []string{"plugins"}},
						{Exec: "cf", Params: []string{
							"push",
							"myTestApp",
							"-f",
							"test-manifest.yml",
						}},

						//
						// There is no cf stop
						//

					}, s.Calls)
				})
			})
		}
	})

	// TODO add test for testCfNativeFailureInShellCall

	t.Run("deploytool mtaDeployPlugin blue green", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "mtaDeployPlugin"
		config.DeployType = "blue-green"
		config.MtaPath = "target/test.mtar"

		defer func() {
			_ = filesMock.FileRemove("target/test.mtar")
		}()

		filesMock.AddFile("target/test.mtar", []byte("content does not matter"))

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, func(t *testing.T) {

					assert.Equal(t, []mock.ExecCall{
						{Exec: "cf", Params: []string{"version"}},
						{Exec: "cf", Params: []string{"plugins"}},
						{Exec: "cf", Params: []string{
							"bg-deploy",
							"target/test.mtar",
							"-f",
							"--no-confirm",
						}},

						//
						// There is no cf stop
						//

					}, s.Calls)
				})
			})
		}
	})

	// TODO: add test for influx reporting (influx reporting is missing at the moment)

	t.Run("cf push with variables from file and as list", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.Manifest = "test-manifest.yml"
		config.ManifestVariablesFiles = []string{"vars.yaml"}
		config.ManifestVariables = []string{"appName=testApplicationFromVarsList"}
		config.AppName = "testAppName"

		defer func() {
			_getManifest = getManifest
			_getVarsOptions = cloudfoundry.GetVarsOptions
			_getVarsFileOptions = cloudfoundry.GetVarsFileOptions
		}()

		_getVarsOptions = func(vars []string) ([]string, error) {
			return []string{"--var", "appName=testApplicationFromVarsList"}, nil
		}
		_getVarsFileOptions = func(varFiles []string) ([]string, error) {
			return []string{"--vars-file", "vars.yaml"}, nil
		}

		filesMock.AddFile("test-manifest.yml", []byte("content does not matter"))

		_getManifest = func(name string) (cloudfoundry.Manifest, error) {
			return manifestMock{
					manifestFileName: "test-manifest.yml",
					apps: []map[string]interface{}{
						{
							"name": "myApp",
						},
					},
				},
				nil
		}

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, func(t *testing.T) {

					// Revisit: we don't verify a log message in case of a non existing vars file

					assert.Equal(t, []mock.ExecCall{
						{Exec: "cf", Params: []string{"version"}},
						{Exec: "cf", Params: []string{"plugins"}},
						{Exec: "cf", Params: []string{
							"push",
							"testAppName",
							"--var",
							"appName=testApplicationFromVarsList",
							"--vars-file",
							"vars.yaml",
							"-f",
							"test-manifest.yml",
						}},
					}, s.Calls)
				})
			})
		}
	})

	t.Run("cf push with variables from file which does not exist", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.Manifest = "test-manifest.yml"
		config.ManifestVariablesFiles = []string{"vars.yaml", "vars-does-not-exist.yaml"}
		config.AppName = "testAppName"

		defer func() {
			_ = filesMock.FileRemove("test-manifest.yml")
			_ = filesMock.FileRemove("vars.yaml")
			_getManifest = getManifest
			_getVarsOptions = cloudfoundry.GetVarsOptions
			_getVarsFileOptions = cloudfoundry.GetVarsFileOptions
		}()

		filesMock.AddFile("test-manifest.yml", []byte("content does not matter"))

		_getManifest = func(name string) (cloudfoundry.Manifest, error) {
			return manifestMock{
					manifestFileName: "test-manifest.yml",
					apps: []map[string]interface{}{
						{
							"name": "myApp",
						},
					},
				},
				nil
		}

		s := mock.ExecMockRunner{}

		var receivedVarOptions []string
		var receivedVarsFileOptions []string

		_getVarsOptions = func(vars []string) ([]string, error) {
			receivedVarOptions = vars
			return []string{}, nil
		}
		_getVarsFileOptions = func(varFiles []string) ([]string, error) {
			receivedVarsFileOptions = varFiles
			return []string{"--vars-file", "vars.yaml"}, nil
		}

		err := runCloudFoundryDeploy(&config, nil, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check received vars options", func(t *testing.T) {
				assert.Empty(t, receivedVarOptions)
			})

			t.Run("check received vars file options", func(t *testing.T) {
				assert.Equal(t, []string{"vars.yaml", "vars-does-not-exist.yaml"}, receivedVarsFileOptions)
			})

			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, func(t *testing.T) {
					// Revisit: we don't verify a log message in case of a non existing vars file

					assert.Equal(t, []mock.ExecCall{
						{Exec: "cf", Params: []string{"version"}},
						{Exec: "cf", Params: []string{"plugins"}},
						{Exec: "cf", Params: []string{
							"push",
							"testAppName",
							"--vars-file",
							"vars.yaml",
							"-f",
							"test-manifest.yml",
						}},
					}, s.Calls)
				})
			})
		}
	})

	// TODO: testCfPushDeploymentWithoutVariableSubstitution is already handled above (?)

	t.Run("deploytool mtaDeployPlugin", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "mtaDeployPlugin"
		config.MtaDeployParameters = "-f"

		t.Run("mta config file from project sources", func(t *testing.T) {

			defer func() { _ = filesMock.FileRemove("xyz.mtar") }()

			// The mock is inaccurat here.
			// AddFile() adds the file absolute, prefix with the current working directory
			// Glob() returns the absolute path - but without leading slash - , whereas
			// the real Glob returns the path relative to the current workdir.
			// In order to mimic the behavior in the free wild we add the mtar at the root dir.
			filesMock.AddDir("/")
			assert.NoError(t, filesMock.Chdir("/"))
			filesMock.AddFile("xyz.mtar", []byte("content does not matter"))
			// restor the expected working dir.
			assert.NoError(t, filesMock.Chdir("/home/me"))
			s := mock.ExecMockRunner{}
			err := runCloudFoundryDeploy(&config, nil, nil, &s)

			if assert.NoError(t, err) {

				withLoginAndLogout(t, func(t *testing.T) {

					assert.Equal(t, s.Calls, []mock.ExecCall{
						{Exec: "cf", Params: []string{"version"}},
						{Exec: "cf", Params: []string{"plugins"}},
						{Exec: "cf", Params: []string{"deploy", "xyz.mtar", "-f"}}})

				})
			}
		})

		t.Run("mta config file from project config does not exist", func(t *testing.T) {
			defer func() { config.MtaPath = "" }()
			config.MtaPath = "my.mtar"
			s := mock.ExecMockRunner{}
			err := runCloudFoundryDeploy(&config, nil, nil, &s)
			assert.EqualError(t, err, "mtar file 'my.mtar' retrieved from configuration does not exist")
		})

		// TODO: add test for mtar file from project config which does exist in project sources
	})
}

func TestValidateDeployTool(t *testing.T) {
	testCases := []struct {
		runName            string
		deployToolGiven    string
		buildTool          string
		deployToolExpected string
	}{
		{"no params", "", "", ""},
		{"build tool MTA", "", "mta", "mtaDeployPlugin"},
		{"build tool other", "", "other", "cf_native"},
		{"deploy and build tool given", "given", "unknown", "given"},
		{"only deploy tool given", "given", "", "given"},
	}

	t.Parallel()

	for _, test := range testCases {
		t.Run(test.runName, func(t *testing.T) {
			config := cloudFoundryDeployOptions{BuildTool: test.buildTool, DeployTool: test.deployToolGiven}
			validateDeployTool(&config)
			assert.Equal(t, test.deployToolExpected, config.DeployTool,
				"expected different deployTool result")
		})
	}
}

func TestMtarLookup(t *testing.T) {

	defer func() {
		fileUtils = piperutils.Files{}
	}()

	filesMock := mock.FilesMock{}
	fileUtils = &filesMock

	t.Run("One MTAR", func(t *testing.T) {

		defer func() { _ = filesMock.FileRemove("x.mtar") }()
		filesMock.AddFile("x.mtar", []byte("content does not matter"))

		path, err := findMtar()

		if assert.NoError(t, err) {
			assert.Equal(t, "x.mtar", path)
		}
	})

	t.Run("No MTAR", func(t *testing.T) {

		// nothing needs to be configures. There is simply no
		// mtar in the file system mock, so no mtar will be found.

		_, err := findMtar()

		assert.EqualError(t, err, "No mtar file matching pattern '**/*.mtar' found")
	})

	t.Run("Several MTARs", func(t *testing.T) {

		defer func() {
			_ = filesMock.FileRemove("x.mtar")
			_ = filesMock.FileRemove("y.mtar")
		}()

		filesMock.AddFile("x.mtar", []byte("content does not matter"))
		filesMock.AddFile("y.mtar", []byte("content does not matter"))

		_, err := findMtar()
		assert.EqualError(t, err, "Found multiple mtar files matching pattern '**/*.mtar' (x.mtar,y.mtar), please specify file via parameter 'mtarPath'")
	})
}

func TestExtensionDescriptorsWithMinusE(t *testing.T) {

	t.Run("ExtensionDescriptorsWithMinusE", func(t *testing.T) {
		extDesc, _ := handleMtaExtensionDescriptors("-e 1.yaml -e 2.yaml")
		assert.Equal(t, []string{
			"-e",
			"1.yaml,2.yaml",
		}, extDesc)
	})

	t.Run("ExtensionDescriptorsFirstOneWithoutMinusE", func(t *testing.T) {
		extDesc, _ := handleMtaExtensionDescriptors("1.yaml -e 2.yaml")
		assert.Equal(t, []string{
			"-e",
			"1.yaml,2.yaml",
		}, extDesc)
	})

	t.Run("NoExtensionDescriptors", func(t *testing.T) {
		extDesc, _ := handleMtaExtensionDescriptors("")
		assert.Equal(t, []string{}, extDesc)
	})
}

func TestAppNameChecks(t *testing.T) {

	t.Run("appName with alpha-numeric chars should work", func(t *testing.T) {
		err := validateAppName("myValidAppName123")
		assert.NoError(t, err)
	})

	t.Run("appName with alpha-numeric chars and dash should work", func(t *testing.T) {
		err := validateAppName("my-Valid-AppName123")
		assert.NoError(t, err)
	})

	t.Run("empty appName should work", func(t *testing.T) {
		// we consider the empty string as valid appname since we only check app names handed over from outside
		// in case there is no (real) app name provided from outside we might still find an appname in the metadata
		// That app name in turn is not checked.
		err := validateAppName("")
		assert.NoError(t, err)
	})

	t.Run("single char appName should work", func(t *testing.T) {
		err := validateAppName("a")
		assert.NoError(t, err)
	})

	t.Run("appName with alpha-numeric chars and trailing dash should throw an error", func(t *testing.T) {
		err := validateAppName("my-Invalid-AppName123-")
		assert.EqualError(t, err, "Your application name 'my-Invalid-AppName123-' starts or ends with a '-' (dash) which is not allowed, only letters and numbers can be used. Please change the name to fit this requirement(s). For more details please visit https://docs.cloudfoundry.org/devguide/deploy-apps/deploy-app.html#basic-settings.")
	})

	t.Run("appName with underscores should throw an error", func(t *testing.T) {
		err := validateAppName("my_invalid_app_name")
		assert.EqualError(t, err, "Your application name 'my_invalid_app_name' contains a '_' (underscore) which is not allowed, only letters, dashes and numbers can be used. Please change the name to fit this requirement(s). For more details please visit https://docs.cloudfoundry.org/devguide/deploy-apps/deploy-app.html#basic-settings.")
	})

}

func TestMtaExtensionCredentials(t *testing.T) {

	filesMock := mock.FilesMock{}
	filesMock.AddDir("/home/me")
	err := filesMock.Chdir("/home/me")
	assert.NoError(t, err)
	fileUtils = &filesMock
	config := defaultConfig

	_environ = func() []string {
		return []string{
			"MY_CRED_ENV_VAR1=**$0****",
			"MY_CRED_ENV_VAR2=++$1++++",
		}
	}

	defer func() {
		fileUtils = piperutils.Files{}
		_environ = os.Environ
	}()

	t.Run("extension file does not exist", func(t *testing.T) {
		_, _, err := handleMtaExtensionCredentials(&config, "mtaextDoesNotExist.mtaext", map[string]interface{}{})
		assert.EqualError(t, err, "Cannot handle credentials for mta extension file 'mtaextDoesNotExist.mtaext': could not read 'mtaextDoesNotExist.mtaext'")
	})

	t.Run("credential cannot be retrieved", func(t *testing.T) {

		filesMock.AddFile("mtaext.mtaext", []byte(
			`'_schema-version: '3.1'
				ID: test.ext
				extends: test
				parameters
					test-credentials1: "<%= testCred1 %>"
					test-credentials2: "<%=testCred2%>"`))
		_, _, err := handleMtaExtensionCredentials(
			&config,
			"mtaext.mtaext",
			map[string]interface{}{
				"testCred1": "myCredEnvVar1NotDefined",
				"testCred2": "myCredEnvVar2NotDefined",
			},
		)
		assert.EqualError(t, err, "cannot handle mta extension credentials: No credentials found for '[myCredEnvVar1NotDefined myCredEnvVar2NotDefined]'/'[MY_CRED_ENV_VAR1_NOT_DEFINED MY_CRED_ENV_VAR2_NOT_DEFINED]'. Are these credentials maintained?")
	})

	t.Run("irrelevant credentials do not cause failures", func(t *testing.T) {

		filesMock.AddFile("mtaext.mtaext", []byte(
			`'_schema-version: '3.1'
				ID: test.ext
				extends: test
				parameters
					test-credentials1: "<%= testCred1 %>"
					test-credentials2: "<%=testCred2%>`))
		_, _, err := handleMtaExtensionCredentials(
			&config,
			"mtaext.mtaext",
			map[string]interface{}{
				"testCred1":       "myCredEnvVar1",
				"testCred2":       "myCredEnvVar2",
				"testCredNotUsed": "myCredEnvVarWhichDoesNotExist", //<-- This here is not used.
			},
		)
		assert.NoError(t, err)
	})

	t.Run("invalid chars in credential key name", func(t *testing.T) {
		filesMock.AddFile("mtaext.mtaext", []byte(
			`'_schema-version: '3.1'
				ID: test.ext
				extends: test
				parameters
					test-credentials1: "<%= testCred1 %>"
					test-credentials2: "<%=testCred2%>`))
		_, _, err := handleMtaExtensionCredentials(
			&config,
			"mtaext.mtaext",
			map[string]interface{}{
				"test.*Cred1": "myCredEnvVar1",
			},
		)
		assert.EqualError(t, err, "credential key name 'test.*Cred1' contains unsupported character. Must contain only ^[-_A-Za-z0-9]+$")
	})

	t.Run("unresolved placeholders does not cause an error", func(t *testing.T) {
		// we emit a log message, but it does not fail
		filesMock.AddFile("mtaext-unresolved.mtaext", []byte("<%= unresolved %>"))
		updated, containsUnresolved, err := handleMtaExtensionCredentials(&config, "mtaext-unresolved.mtaext", map[string]interface{}{})
		assert.True(t, containsUnresolved)
		assert.False(t, updated)
		assert.NoError(t, err)
	})

	t.Run("replace straight forward", func(t *testing.T) {
		mtaFileName := "mtaext.mtaext"
		filesMock.AddFile(mtaFileName, []byte(
			`'_schema-version: '3.1'
			ID: test.ext
			extends: test
			parameters
				test-credentials1: "<%= testCred1 %>"
				test-credentials2: "<%=testCred2%>"
				test-credentials3: "<%= testCred2%>"
				test-credentials4: "<%=testCred2 %>"
				test-credentials5: "<%=  testCred2    %>"`))
		updated, containsUnresolved, err := handleMtaExtensionCredentials(
			&config,
			mtaFileName,
			map[string]interface{}{
				"testCred1": "myCredEnvVar1",
				"testCred2": "myCredEnvVar2",
			},
		)
		if assert.NoError(t, err) {
			b, e := fileUtils.FileRead(mtaFileName)
			if e != nil {
				assert.Fail(t, "Cannot read mta extension file: %v", e)
			}
			content := string(b)
			assert.Contains(t, content, "test-credentials1: \"**$0****\"")
			assert.Contains(t, content, "test-credentials2: \"++$1++++\"")
			assert.Contains(t, content, "test-credentials3: \"++$1++++\"")
			assert.Contains(t, content, "test-credentials4: \"++$1++++\"")
			assert.Contains(t, content, "test-credentials5: \"++$1++++\"")

			assert.True(t, updated)
			assert.False(t, containsUnresolved)
		}
	})
}

func TestEnvVarKeyModification(t *testing.T) {
	envVarCompatibleKey := toEnvVarKey("Mta.EXtensionCredential~Credential_Id1Abc")
	assert.Equal(t, "MTA_EXTENSION_CREDENTIAL_CREDENTIAL_ID1_ABC", envVarCompatibleKey)
}
