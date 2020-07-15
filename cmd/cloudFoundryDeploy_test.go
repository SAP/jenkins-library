package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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

func TestCfDeployment(t *testing.T) {

	defer func() {
		fileUtils = &piperutils.Files{}
		_substitute = yaml.Substitute
	}()

	filesMock := mock.FilesMock{}
	filesMock.AddDir("/home/me")
	filesMock.Chdir("/home/me")
	fileUtils = &filesMock

	// everything below in the config map annotated with '//default' is a default in the metadata
	// since we don't get injected these values during the tests we set it here.
	defaultConfig := cloudFoundryDeployOptions{
		Org:                 "myOrg",
		Space:               "mySpace",
		Username:            "me",
		Password:            "******",
		APIEndpoint:         "https://examples.sap.com/cf",
		SmokeTestStatusCode: 200,            // default
		Manifest:            "manifest.yml", //default
		MtaDeployParameters: "-f",           // default
		DeployType:          "standard",     // default
	}

	config := defaultConfig

	successfulLogin := cloudfoundry.LoginOptions{
		CfAPIEndpoint: "https://examples.sap.com/cf",
		CfOrg:         "myOrg",
		CfSpace:       "mySpace",
		Username:      "me",
		Password:      "******",
		CfAPIOpts:     []string{},
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
			filesMock.FileRemove(manifestName) // slightly mis-use since that is intended to be used by code under test, not test code
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

	_substitute = func(manifest string, replacements map[string]interface{}, replacementsFiles []string) (bool, error) {
		return false, nil
	}

	t.Run("Manifest substitution", func(t *testing.T) {

		defer func() {
			cleanup()
			_substitute = func(manifest string, replacements map[string]interface{}, replacementsFiles []string) (bool, error) {
				return false, nil
			}
		}()

		s := mock.ExecMockRunner{}

		var manifestForSubstitution string
		var replacements map[string]interface{}
		var replacementFiles []string

		defer prepareDefaultManifestMocking("substitute-manifest.yml", []string{"testAppName"})()
		config.DeployTool = "cf_native"
		config.DeployType = "blue-green"
		config.AppName = "myApp"
		config.Manifest = "substitute-manifest.yml"

		_substitute = func(manifest string, _replacements map[string]interface{}, _replacementsFiles []string) (bool, error) {
			manifestForSubstitution = manifest
			replacements = _replacements
			replacementFiles = _replacementsFiles
			return false, nil
		}

		t.Run("straight forward", func(t *testing.T) {

			defer func() {
				config.ManifestVariables = []string{}
				config.ManifestVariablesFiles = []string{}
			}()

			config.ManifestVariables = []string{"k1=v1"}
			config.ManifestVariablesFiles = []string{"myVars.yml"}

			err := runCloudFoundryDeploy(&config, nil, nil, &s)

			if assert.NoError(t, err) {
				assert.Equal(t, "substitute-manifest.yml", manifestForSubstitution)
				assert.Equal(t, map[string]interface{}{"k1": "v1"}, replacements)
				assert.Equal(t, []string{"myVars.yml"}, replacementFiles)
			}
		})

		t.Run("empty", func(t *testing.T) {

			defer func() {
				config.ManifestVariables = []string{}
				config.ManifestVariablesFiles = []string{}
			}()

			config.ManifestVariables = []string{}
			config.ManifestVariablesFiles = []string{}

			err := runCloudFoundryDeploy(&config, nil, nil, &s)

			if assert.NoError(t, err) {
				assert.Equal(t, "substitute-manifest.yml", manifestForSubstitution)
				assert.Equal(t, map[string]interface{}{}, replacements)
				assert.Equal(t, []string{}, replacementFiles)
			}
		})
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
						mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
						mock.ExecCall{Exec: "cf", Params: []string{"push", "-f", "manifest.yml"}},
					}, s.Calls)
				})
			})

			t.Run("check environment variables", func(t *testing.T) {
				assert.Contains(t, s.Env, "CF_HOME=/home/me1")
				assert.Contains(t, s.Env, "CF_PLUGIN_HOME=/home/me2")
				assert.Contains(t, s.Env, "STATUS_CODE=200")
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
			// There was the big eclise in Karlsruhe
			return time.Date(1999, time.August, 11, 12, 32, 0, 0, time.UTC)
		}

		defer prepareDefaultManifestMocking("manifest.yml", []string{"testAppName"})()

		config.DeployTool = "cf_native"

		influxData := cloudFoundryDeployInflux{}

		err := runCloudFoundryDeploy(&config, nil, &influxData, &s)

		if assert.NoError(t, err) {

			expected := cloudFoundryDeployInflux{}

			expected.deployment_data.fields.artifactURL = "n/a"
			expected.deployment_data.fields.deployTime = "AUG 11 1999 12:32:00"
			expected.deployment_data.fields.jobTrigger = "<n/a>"

			expected.deployment_data.tags.artifactVersion = "<n/a>" // TODO revisit
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
					mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
					mock.ExecCall{Exec: "cf", Params: []string{"push",
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
						mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
						mock.ExecCall{Exec: "cf", Params: []string{"push",
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
				//REVISIT: in the corresponding groovy test we checked for "${'********'}"
				// I don't understand why, but we should discuss ...
				assert.Contains(t, s.Env, "CF_DOCKER_PASSWORD=********")
			})
		}
	})

	t.Run("deploy cf native blue green with manifest and docker credentials", func(t *testing.T) {

		defer cleanup()

		// Blue Green Deploy cf cli plugin does not support --docker-username and --docker-image parameters
		// docker username and docker image have to be set in the manifest file
		// if a private docker repository is used the CF_DOCKER_PASSWORD env variable must be set

		config.DeployTool = "cf_native"
		config.DeployType = "blue-green"
		config.DockerUsername = "test_cf_docker"
		config.DockerPassword = "********"
		config.AppName = "testAppName"

		defer prepareDefaultManifestMocking("manifest.yml", []string{"testAppName"})()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, func(t *testing.T) {

					assert.Equal(t, []mock.ExecCall{
						mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
						mock.ExecCall{Exec: "cf", Params: []string{
							"blue-green-deploy",
							"testAppName",
							"--delete-old-apps",
							"-f",
							"manifest.yml",
						}},
					}, s.Calls)
				})
			})

			t.Run("check environment variables", func(t *testing.T) {
				//REVISIT: in the corresponding groovy test we checked for "${'********'}"
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
						mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
						mock.ExecCall{Exec: "cf", Params: []string{
							"push",
							"-f",
							"test-manifest.yml",
						}},
					}, s.Calls)

				})
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

		if assert.EqualError(t, err, "No appName available in manifest 'test-manifest.yml'") {

			t.Run("check shell calls", func(t *testing.T) {
				noopCfAPICalls(t, s)
			})
		}
	})

	// tests from groovy checking for keep old instances are already contained above. Search for '--delete-old-apps'

	t.Run("deploy cf native blue green keep old instance", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.DeployType = "blue-green"
		config.Manifest = "test-manifest.yml"
		config.AppName = "myTestApp"
		config.KeepOldInstance = true

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, func(t *testing.T) {

					assert.Equal(t, []mock.ExecCall{
						mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
						mock.ExecCall{Exec: "cf", Params: []string{
							"blue-green-deploy",
							"myTestApp",
							"-f",
							"test-manifest.yml",
						}},
						mock.ExecCall{Exec: "cf", Params: []string{
							"stop",
							"myTestApp-old",
							// MIGRATE FFROM GROOVY: in contrast to groovy there is not redirect of everything &> to a file since we
							// read the stream directly now.
						}},
					}, s.Calls)
				})
			})
		}
	})

	t.Run("cf deploy blue green multiple applications", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.DeployType = "blue-green"
		config.Manifest = "test-manifest.yml"
		config.AppName = "myTestApp"

		defer prepareDefaultManifestMocking("test-manifest.yml", []string{"app1", "app2"})()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s)

		if assert.EqualError(t, err, "Your manifest contains more than one application. For blue green deployments your manifest file may contain only one application") {
			t.Run("check shell calls", func(t *testing.T) {
				noopCfAPICalls(t, s)
			})
		}
	})

	t.Run("cf native deploy blue green with no route", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.DeployType = "blue-green"
		config.Manifest = "test-manifest.yml"
		config.AppName = "myTestApp"

		defer func() {
			filesMock.FileRemove("test-manifest.yml")
			_getManifest = getManifest
		}()

		filesMock.AddFile("test-manifest.yml", []byte("Content does not matter"))

		_getManifest = func(name string) (cloudfoundry.Manifest, error) {
			return manifestMock{
					manifestFileName: "test-manifest.yml",
					apps: []map[string]interface{}{
						map[string]interface{}{
							"name":     "app1",
							"no-route": true,
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

					assert.Equal(t, []mock.ExecCall{
						mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
						mock.ExecCall{Exec: "cf", Params: []string{
							"push",
							"myTestApp",
							"-f",
							"test-manifest.yml",
						}},
					}, s.Calls)
				})
			})
		}
	})

	t.Run("cf native deployment failure", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.DeployType = "blue-green"
		config.Manifest = "test-manifest.yml"
		config.AppName = "myTestApp"

		defer prepareDefaultManifestMocking("test-manifest.yml", []string{"app"})()

		s := mock.ExecMockRunner{}

		s.ShouldFailOnCommand = map[string]error{"cf.*": fmt.Errorf("cf deploy failed")}
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
		config.DeployType = "blue-green"
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
				assert.Empty(t, s.Calls)
				// no logout
				assert.False(t, logoutCalled)
			})
		}
	})

	// TODO testCfNativeBlueGreenKeepOldInstanceShouldThrowErrorOnStopError

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
						mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
						mock.ExecCall{Exec: "cf", Params: []string{
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

	t.Run("testCfNativeWithoutAppNameBlueGreen", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.DeployType = "blue-green"
		config.Manifest = "test-manifest.yml"

		defer func() {
			filesMock.FileRemove("test-manifest.yml")
			_getManifest = getManifest
		}()

		filesMock.AddFile("test-manifest.yml", []byte("The content does not matter"))

		_getManifest = func(name string) (cloudfoundry.Manifest, error) {
			return manifestMock{
					manifestFileName: "test-manifest.yml",
					apps: []map[string]interface{}{
						map[string]interface{}{
							"there-is": "no-app-name",
						},
					},
				},
				nil
		}

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s)

		if assert.EqualError(t, err, "Blue-green plugin requires app name to be passed (see https://github.com/bluemixgaragelondon/cf-blue-green-deploy/issues/27)") {

			t.Run("check shell calls", func(t *testing.T) {
				noopCfAPICalls(t, s)
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
			filesMock.FileRemove("target/test.mtar")
		}()

		filesMock.AddFile("target/test.mtar", []byte("content does not matter"))

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, func(t *testing.T) {

					assert.Equal(t, []mock.ExecCall{
						mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
						mock.ExecCall{Exec: "cf", Params: []string{
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
		}()

		filesMock.AddFile("vars.yaml", []byte("content does not matter"))
		filesMock.AddFile("test-manifest.yml", []byte("content does not matter"))

		_getManifest = func(name string) (cloudfoundry.Manifest, error) {
			return manifestMock{
					manifestFileName: "test-manifest.yml",
					apps: []map[string]interface{}{
						map[string]interface{}{
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
						mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
						mock.ExecCall{Exec: "cf", Params: []string{
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
			filesMock.FileRemove("test-manifest.yml")
			filesMock.FileRemove("vars.yaml")
			_getManifest = getManifest
		}()

		filesMock.AddFile("test-manifest.yml", []byte("content does not matter"))
		filesMock.AddFile("vars.yaml", []byte("content does not matter"))

		_getManifest = func(name string) (cloudfoundry.Manifest, error) {
			return manifestMock{
					manifestFileName: "test-manifest.yml",
					apps: []map[string]interface{}{
						map[string]interface{}{
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
						mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
						mock.ExecCall{Exec: "cf", Params: []string{
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

	// TODO: testCfBlueGreenDeploymentWithVariableSubstitution variable substitution is not handled at the moment (pr pending).
	// but anyway we should not test the full cycle here, but only that the variables substitution tool is called in the appropriate way.
	// variable substitution should be tested at the variables substitution tool itself (yaml util)

	t.Run("deploytool mtaDeployPlugin", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "mtaDeployPlugin"
		config.MtaDeployParameters = "-f"

		t.Run("mta config file from project sources", func(t *testing.T) {

			defer filesMock.FileRemove("xyz.mtar")

			// The mock is inaccurat here.
			// AddFile() adds the file absolute, prefix with the current working directory
			// Glob() returns the absolute path - but without leading slash - , whereas
			// the real Glob returns the path relative to the current workdir.
			// In order to mimic the behavour in the free wild we add the mtar at the root dir.
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
						mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
						mock.ExecCall{Exec: "cf", Params: []string{"deploy", "xyz.mtar", "-f"}}})

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

func TestManifestVariableFiles(t *testing.T) {

	defer func() {
		fileUtils = &piperutils.Files{}
	}()

	filesMock := mock.FilesMock{}
	fileUtils = &filesMock

	filesMock.AddFile("a/varsA.txt", []byte("content does not matter"))
	filesMock.AddFile("varsB.txt", []byte("content does not matter"))

	t.Run("straight forward", func(t *testing.T) {
		varOpts, err := getVarFileOptions([]string{"a/varsA.txt", "varsB.txt"})
		if assert.NoError(t, err) {
			assert.Equal(t, []string{"--vars-file", "a/varsA.txt", "--vars-file", "varsB.txt"}, varOpts)
		}
	})

	t.Run("no var filesprovided", func(t *testing.T) {
		varOpts, err := getVarFileOptions([]string{})
		if assert.NoError(t, err) {
			assert.Equal(t, []string{}, varOpts)
		}
	})

	t.Run("one var file does not exist", func(t *testing.T) {
		varOpts, err := getVarFileOptions([]string{"a/varsA.txt", "doesNotExist.txt"})
		if assert.NoError(t, err) {
			assert.Equal(t, []string{"--vars-file", "a/varsA.txt"}, varOpts)
		}
	})
}

func TestManifestVariables(t *testing.T) {
	t.Run("straight forward", func(t *testing.T) {
		varOpts, err := getVarOptions([]string{"a=b", "c=d"})
		if assert.NoError(t, err) {
			assert.Equal(t, []string{"--var", "a=b", "--var", "c=d"}, varOpts)
		}
	})

	t.Run("empty variabls list", func(t *testing.T) {
		varOpts, err := getVarOptions([]string{})
		if assert.NoError(t, err) {
			assert.Equal(t, []string{}, varOpts)
		}
	})

	t.Run("no equal sign in variable", func(t *testing.T) {
		_, err := getVarOptions([]string{"ab"})
		assert.EqualError(t, err, "Invalid parameter provided (expected format <key>=<val>: 'ab'")
	})
}

func TestMtarLookup(t *testing.T) {

	defer func() {
		fileUtils = piperutils.Files{}
	}()

	filesMock := mock.FilesMock{}
	fileUtils = &filesMock

	t.Run("One MTAR", func(t *testing.T) {

		defer filesMock.FileRemove("x.mtar")
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
			filesMock.FileRemove("x.mtar")
			filesMock.FileRemove("y.mtar")
		}()

		filesMock.AddFile("x.mtar", []byte("content does not matter"))
		filesMock.AddFile("y.mtar", []byte("content does not matter"))

		_, err := findMtar()
		assert.EqualError(t, err, "Found multiple mtar files matching pattern '**/*.mtar' (x.mtar,y.mtar), please specify file via mtaPath parameter 'mtarPath'")
	})
}
