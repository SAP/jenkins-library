//go:build unit
// +build unit

package cmd

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/mock"

	"github.com/stretchr/testify/assert"
)

func TestCfDeployment(t *testing.T) {

	t.Chdir(t.TempDir())

	// everything below in the config map annotated with '//default' is a default in the metadata
	// since we don't get injected these values during the tests we set it here.
	defaultConfig := cloudFoundryDeployOptions{
		Org:                 "myOrg",
		Space:               "mySpace",
		Username:            "me",
		Password:            "******",
		APIEndpoint:         "https://examples.sap.com/cf",
		Manifest:            "manifest.yml", // default
		MtaDeployParameters: "-f",           // default
		DeployType:          "standard",     // default
	}

	config := defaultConfig

	noopCfAPICalls := func(t *testing.T, s mock.ExecMockRunner) {
		assert.Empty(t, s.Calls) // invalid deploy tools must not execute CF commands
	}

	prepareManifest := func(manifestName string, appNames []string) func() {
		applications := ""
		for _, appName := range appNames {
			if appName == "" {
				appName = `""`
			}
			applications += fmt.Sprintf("  - name: %s\n", appName)
		}
		manifest := fmt.Sprintf("applications:\n%s", applications)
		assert.NoError(t, os.WriteFile(manifestName, []byte(manifest), 0644))
		return func() {
			assert.NoError(t, os.Remove(manifestName))
		}
	}

	withLoginAndLogout := func(t *testing.T, runner *mock.ExecMockRunner, asserts func(t *testing.T)) {
		t.Helper()
		assert.GreaterOrEqual(t, len(runner.Calls), 3)
		assert.Equal(t, []string{"login", "-a", defaultConfig.APIEndpoint, "-o", defaultConfig.Org, "-s", defaultConfig.Space, "-u", defaultConfig.Username, "-p", defaultConfig.Password}, runner.Calls[1].Params)
		assert.Equal(t, []string{"logout"}, runner.Calls[len(runner.Calls)-1].Params)

		calls := runner.Calls
		runner.Calls = append(append([]mock.ExecCall{}, calls[:1]...), calls[2:len(calls)-1]...)
		defer func() { runner.Calls = calls }()
		asserts(t)
	}

	cleanup := func() {
		config = defaultConfig
	}

	t.Run("deploy with token authentication", func(t *testing.T) {
		defer cleanup()

		config.Token = "testToken"
		config.TokenOrigin = "testOrigin"
		s := mock.ExecMockRunner{}

		err := cfDeploy(&config, []string{"push", "testApp"}, nil, &s)

		if assert.NoError(t, err) {
			assert.Equal(t, []mock.ExecCall{
				{Exec: "cf", Params: []string{"version"}},
				{Exec: "cf", Params: []string{"api", config.APIEndpoint}},
				{Exec: "cf", Params: []string{"auth", "--assertion", "testToken", "--origin", "testOrigin"}},
				{Exec: "cf", Params: []string{"target", "-o", config.Org, "-s", config.Space}},
				{Exec: "cf", Params: []string{"plugins"}},
				{Exec: "cf", Params: []string{"push", "testApp"}},
				{Exec: "cf", Params: []string{"logout"}},
			}, s.Calls)
		}
	})

	t.Run("Test invalid appname", func(t *testing.T) {

		defer cleanup()
		config.AppName = "a_z"
		s := mock.ExecMockRunner{}
		err := runCloudFoundryDeploy(&config, nil, nil, &s, time.Now)

		assert.EqualError(t, err, "Your application name 'a_z' contains a '_' (underscore) which is not allowed, only letters, dashes and numbers can be used. Please change the name to fit this requirement(s). For more details please visit https://docs.cloudfoundry.org/devguide/deploy-apps/deploy-app.html#basic-settings.")
	})

	t.Run("Invalid deploytool", func(t *testing.T) {

		defer cleanup()

		s := mock.ExecMockRunner{}

		config.DeployTool = "invalid"

		err := runCloudFoundryDeploy(&config, nil, nil, &s, time.Now)

		if assert.NoError(t, err) {
			noopCfAPICalls(t, s)
		}
	})

	t.Run("deploytool cf native", func(t *testing.T) {

		defer cleanup()

		defer prepareManifest("manifest.yml", []string{"testAppName"})()

		config.DeployTool = "cf_native"
		config.CfHome = "/home/me1"
		config.CfPluginHome = "/home/me2"

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s, time.Now)

		if assert.NoError(t, err) {

			t.Run("check cf api calls", func(t *testing.T) {

				withLoginAndLogout(t, &s, func(t *testing.T) {
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

		defer prepareManifest("manifest.yml", []string{"testAppName"})()

		config.DeployTool = "cf_native"
		config.ArtifactVersion = "0.1.2"
		config.CommitHash = "123456"

		influxData := cloudFoundryDeployInflux{}

		now := func() time.Time {
			// There was the big eclipse in Karlsruhe.
			return time.Date(1999, time.August, 11, 12, 32, 0, 0, time.UTC)
		}

		err := runCloudFoundryDeploy(&config, nil, &influxData, &s, now)

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

		err := runCloudFoundryDeploy(&config, nil, nil, &s, time.Now)

		if assert.NoError(t, err) {

			withLoginAndLogout(t, &s, func(t *testing.T) {
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

		err := runCloudFoundryDeploy(&config, nil, nil, &s, time.Now)

		if assert.NoError(t, err) {
			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, &s, func(t *testing.T) {

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
		defer prepareManifest("test-manifest.yml", []string{"dummyApp"})()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s, time.Now)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, &s, func(t *testing.T) {

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
		defer prepareManifest("manifest.yml", []string{"newAppName"})()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s, time.Now)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, &s, func(t *testing.T) {

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
		defer prepareManifest("manifest.yml", []string{"newAppName"})()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s, time.Now)

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
		defer prepareManifest("manifest.yml", []string{"newAppName"})()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s, time.Now)

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
		defer prepareManifest("test-manifest.yml", []string{""})()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s, time.Now)

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

		defer prepareManifest("test-manifest.yml", []string{"app"})()

		s := mock.ExecMockRunner{}

		s.ShouldFailOnCommand = map[string]error{"cf.*push.*": fmt.Errorf("cf deploy failed")}
		err := runCloudFoundryDeploy(&config, nil, nil, &s, time.Now)

		if assert.EqualError(t, err, "cf deploy failed") {
			assert.Equal(t, []string{"logout"}, s.Calls[len(s.Calls)-1].Params)
		}
	})

	t.Run("cf native deployment failure when logging in", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.DeployType = "standard"
		config.Manifest = "test-manifest.yml"
		config.AppName = "myTestApp"

		s := mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"cf login .*": fmt.Errorf("Unable to login")}}

		err := runCloudFoundryDeploy(&config, nil, nil, &s, time.Now)

		if assert.EqualError(t, err, "Failed to login to Cloud Foundry: Unable to login") {
			assert.Equal(t, []mock.ExecCall{
				{Exec: "cf", Params: []string{"version"}},
				{Exec: "cf", Params: []string{"login", "-a", config.APIEndpoint, "-o", config.Org, "-s", config.Space, "-u", config.Username, "-p", config.Password}},
			}, s.Calls)
		}
	})

	t.Run("cf native deploy standard should not stop instance", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.DeployType = "standard"
		config.Manifest = "test-manifest.yml"
		config.AppName = "myTestApp"
		config.KeepOldInstance = true

		defer prepareManifest("test-manifest.yml", []string{"app"})()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s, time.Now)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, &s, func(t *testing.T) {

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
			_ = os.Remove("target/test.mtar")
		}()

		assert.NoError(t, os.MkdirAll("target", 0755))
		assert.NoError(t, os.WriteFile("target/test.mtar", []byte("content does not matter"), 0644))

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s, time.Now)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, &s, func(t *testing.T) {

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

		assert.NoError(t, os.WriteFile("test-manifest.yml", []byte("applications:\n  - name: myApp\n"), 0644))
		assert.NoError(t, os.WriteFile("vars.yaml", []byte("appName: testApplicationFromVarsFile\n"), 0644))

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s, time.Now)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, &s, func(t *testing.T) {

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

		assert.NoError(t, os.WriteFile("test-manifest.yml", []byte("applications:\n  - name: myApp\n"), 0644))
		assert.NoError(t, os.WriteFile("vars.yaml", []byte("appName: testApplicationFromVarsFile\n"), 0644))

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, nil, &s, time.Now)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, &s, func(t *testing.T) {
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

			assert.NoError(t, os.WriteFile("xyz.mtar", []byte("content does not matter"), 0644))
			s := mock.ExecMockRunner{}
			err := runCloudFoundryDeploy(&config, nil, nil, &s, time.Now)

			if assert.NoError(t, err) {

				withLoginAndLogout(t, &s, func(t *testing.T) {

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
			err := runCloudFoundryDeploy(&config, nil, nil, &s, time.Now)
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

	t.Chdir(t.TempDir())

	t.Run("One MTAR", func(t *testing.T) {

		assert.NoError(t, os.WriteFile("x.mtar", []byte("content does not matter"), 0644))
		defer os.Remove("x.mtar")

		path, err := findMtar()

		if assert.NoError(t, err) {
			assert.Equal(t, "x.mtar", path)
		}
	})

	t.Run("No MTAR", func(t *testing.T) {

		// nothing needs to be configured. There is simply no
		// mtar in the temporary file system, so no mtar will be found.

		_, err := findMtar()

		assert.EqualError(t, err, "No mtar file matching pattern '**/*.mtar' found")
	})

	t.Run("Several MTARs", func(t *testing.T) {

		assert.NoError(t, os.WriteFile("x.mtar", []byte("content does not matter"), 0644))
		assert.NoError(t, os.WriteFile("y.mtar", []byte("content does not matter"), 0644))
		defer os.Remove("x.mtar")
		defer os.Remove("y.mtar")

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

	t.Chdir(t.TempDir())
	t.Setenv("MY_CRED_ENV_VAR1", "**$0****")
	t.Setenv("MY_CRED_ENV_VAR2", "++$1++++")

	writeFixture := func(name string, content []byte) {
		t.Helper()
		assert.NoError(t, os.WriteFile(name, content, 0644))
	}

	t.Run("extension file does not exist", func(t *testing.T) {
		_, _, err := handleMtaExtensionCredentials("mtaextDoesNotExist.mtaext", map[string]interface{}{})
		assert.EqualError(t, err, "Cannot handle credentials for mta extension file 'mtaextDoesNotExist.mtaext': open mtaextDoesNotExist.mtaext: no such file or directory")
	})

	t.Run("credential cannot be retrieved", func(t *testing.T) {

		writeFixture("mtaext.mtaext", []byte(
			`'_schema-version: '3.1'
				ID: test.ext
				extends: test
				parameters
					test-credentials1: "<%= testCred1 %>"
					test-credentials2: "<%=testCred2%>"`))
		_, _, err := handleMtaExtensionCredentials(
			"mtaext.mtaext",
			map[string]interface{}{
				"testCred1": "myCredEnvVar1NotDefined",
				"testCred2": "myCredEnvVar2NotDefined",
			},
		)
		assert.EqualError(t, err, "cannot handle mta extension credentials: No credentials found for '[myCredEnvVar1NotDefined myCredEnvVar2NotDefined]'/'[MY_CRED_ENV_VAR1_NOT_DEFINED MY_CRED_ENV_VAR2_NOT_DEFINED]'. Are these credentials maintained?")
	})

	t.Run("irrelevant credentials do not cause failures", func(t *testing.T) {

		writeFixture("mtaext.mtaext", []byte(
			`'_schema-version: '3.1'
				ID: test.ext
				extends: test
				parameters
					test-credentials1: "<%= testCred1 %>"
					test-credentials2: "<%=testCred2%>`))
		_, _, err := handleMtaExtensionCredentials(
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
		writeFixture("mtaext.mtaext", []byte(
			`'_schema-version: '3.1'
				ID: test.ext
				extends: test
				parameters
					test-credentials1: "<%= testCred1 %>"
					test-credentials2: "<%=testCred2%>`))
		_, _, err := handleMtaExtensionCredentials("mtaext.mtaext",
			map[string]interface{}{
				"test.*Cred1": "myCredEnvVar1",
			},
		)
		assert.EqualError(t, err, "credential key name 'test.*Cred1' contains unsupported character. Must contain only ^[-_A-Za-z0-9]+$")
	})

	t.Run("unresolved placeholders does not cause an error", func(t *testing.T) {
		// we emit a log message, but it does not fail
		writeFixture("mtaext-unresolved.mtaext", []byte("<%= unresolved %>"))
		updated, containsUnresolved, err := handleMtaExtensionCredentials("mtaext-unresolved.mtaext", map[string]interface{}{})
		assert.True(t, containsUnresolved)
		assert.False(t, updated)
		assert.NoError(t, err)
	})

	t.Run("replace straight forward", func(t *testing.T) {
		mtaFileName := "mtaext.mtaext"
		writeFixture(mtaFileName, []byte(
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
			mtaFileName,
			map[string]interface{}{
				"testCred1": "myCredEnvVar1",
				"testCred2": "myCredEnvVar2",
			},
		)
		if assert.NoError(t, err) {
			b, e := os.ReadFile(mtaFileName)
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
