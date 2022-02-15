package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const imageRegistry = "some-registry"

func newCnbBuildTestsUtils() cnbutils.MockUtils {
	utils := cnbutils.MockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func addBuilderFiles(utils *cnbutils.MockUtils) {
	utils.FilesMock.AddFile(creatorPath, []byte(`xyz`))
}

func assertLifecycleCalls(t *testing.T, runner *mock.ExecMockRunner, callNo int) {
	require.GreaterOrEqual(t, len(runner.Calls), callNo)
	assert.Equal(t, creatorPath, runner.Calls[callNo-1].Exec)
	for _, arg := range []string{"-no-color", "-buildpacks", "/cnb/buildpacks", "-order", "/cnb/order.toml", "-platform", "/tmp/platform"} {
		assert.Contains(t, runner.Calls[callNo-1].Params, arg)
	}
}

func TestRunCnbBuild(t *testing.T) {
	t.Parallel()

	t.Run("prefers direct configuration", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "0.0.1",
			ContainerRegistryURL: fmt.Sprintf("https://%s", imageRegistry),
			DockerConfigJSON:     "/path/to/config.json",
		}

		projectToml := `[project]
		id = "io.buildpacks.my-app"
		`

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		utils.FilesMock.AddFile("project.toml", []byte(projectToml))
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &cnbBuildTelemetry{}, &utils, &commonPipelineEnvironment, &piperhttp.Client{})

		require.NoError(t, err)
		runner := utils.ExecMockRunner
		assert.Contains(t, runner.Env, "CNB_REGISTRY_AUTH={\"my-registry\":\"Basic dXNlcjpwYXNz\"}")
		assertLifecycleCalls(t, runner, 1)
		assert.Contains(t, runner.Calls[0].Params, fmt.Sprintf("%s/%s:%s", imageRegistry, config.ContainerImageName, config.ContainerImageTag))
		assert.Equal(t, config.ContainerRegistryURL, commonPipelineEnvironment.container.registryURL)
		assert.Equal(t, "my-image:0.0.1", commonPipelineEnvironment.container.imageNameTag)
	})

	t.Run("prefers project descriptor", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{
			ContainerImageTag:    "0.0.1",
			ContainerRegistryURL: fmt.Sprintf("https://%s", imageRegistry),
			DockerConfigJSON:     "/path/to/config.json",
			ProjectDescriptor:    "project.toml",
		}

		projectToml := `[project]
		id = "io.buildpacks.my-app"
		`

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		utils.FilesMock.AddFile("project.toml", []byte(projectToml))
		addBuilderFiles(&utils)

		telemetryData := telemetry.CustomData{}
		err := callCnbBuild(&config, &telemetryData, &cnbBuildTelemetry{}, &utils, &commonPipelineEnvironment, &piperhttp.Client{})

		require.NoError(t, err)
		runner := utils.ExecMockRunner
		assert.Contains(t, runner.Env, "CNB_REGISTRY_AUTH={\"my-registry\":\"Basic dXNlcjpwYXNz\"}")
		assertLifecycleCalls(t, runner, 1)
		assert.Contains(t, runner.Calls[0].Params, fmt.Sprintf("%s/%s:%s", imageRegistry, "io-buildpacks-my-app", config.ContainerImageTag))
		assert.Equal(t, config.ContainerRegistryURL, commonPipelineEnvironment.container.registryURL)
		assert.Equal(t, "io-buildpacks-my-app:0.0.1", commonPipelineEnvironment.container.imageNameTag)

		customDataAsString := telemetryData.Custom1
		customData := cnbBuildTelemetry{}
		err = json.Unmarshal([]byte(customDataAsString), &customData)
		require.NoError(t, err)
		assert.Equal(t, 1, len(customData.Data))
		assert.Equal(t, "root", string(customData.Data[0].Path))
	})

	t.Run("success case (registry with https)", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "0.0.1",
			ContainerRegistryURL: fmt.Sprintf("https://%s", imageRegistry),
			DockerConfigJSON:     "/path/to/config.json",
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &cnbBuildTelemetry{}, &utils, &commonPipelineEnvironment, &piperhttp.Client{})

		require.NoError(t, err)
		runner := utils.ExecMockRunner
		assert.Contains(t, runner.Env, "CNB_REGISTRY_AUTH={\"my-registry\":\"Basic dXNlcjpwYXNz\"}")
		assertLifecycleCalls(t, runner, 1)
		assert.Contains(t, runner.Calls[0].Params, fmt.Sprintf("%s/%s:%s", imageRegistry, config.ContainerImageName, config.ContainerImageTag))
		assert.Equal(t, config.ContainerRegistryURL, commonPipelineEnvironment.container.registryURL)
		assert.Equal(t, "my-image:0.0.1", commonPipelineEnvironment.container.imageNameTag)
	})

	t.Run("success case (registry without https)", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "0.0.1",
			ContainerRegistryURL: imageRegistry,
			DockerConfigJSON:     "/path/to/config.json",
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &cnbBuildTelemetry{}, &utils, &commonPipelineEnvironment, &piperhttp.Client{})

		require.NoError(t, err)
		runner := utils.ExecMockRunner
		assert.Contains(t, runner.Env, "CNB_REGISTRY_AUTH={\"my-registry\":\"Basic dXNlcjpwYXNz\"}")
		assertLifecycleCalls(t, runner, 1)
		assert.Contains(t, runner.Calls[0].Params, fmt.Sprintf("%s/%s:%s", config.ContainerRegistryURL, config.ContainerImageName, config.ContainerImageTag))
		assert.Equal(t, fmt.Sprintf("https://%s", config.ContainerRegistryURL), commonPipelineEnvironment.container.registryURL)
		assert.Equal(t, "my-image:0.0.1", commonPipelineEnvironment.container.imageNameTag)
	})

	t.Run("success case (custom buildpacks and custom env variables, renaming docker conf file, additional tag)", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "0.0.1",
			ContainerRegistryURL: imageRegistry,
			DockerConfigJSON:     "/path/to/test.json",
			Buildpacks:           []string{"test"},
			BuildEnvVars: map[string]interface{}{
				"FOO": "BAR",
			},
			AdditionalTags: []string{"latest"},
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &cnbBuildTelemetry{}, &utils, &commonPipelineEnvironment, &piperhttp.Client{})

		require.NoError(t, err)
		runner := utils.ExecMockRunner
		assert.Contains(t, runner.Env, "CNB_REGISTRY_AUTH={\"my-registry\":\"Basic dXNlcjpwYXNz\"}")
		assert.Equal(t, creatorPath, runner.Calls[0].Exec)
		assert.Contains(t, runner.Calls[0].Params, "/tmp/buildpacks")
		assert.Contains(t, runner.Calls[0].Params, "/tmp/buildpacks/order.toml")
		assert.Contains(t, runner.Calls[0].Params, fmt.Sprintf("%s/%s:%s", config.ContainerRegistryURL, config.ContainerImageName, config.ContainerImageTag))
		assert.Contains(t, runner.Calls[0].Params, fmt.Sprintf("%s/%s:latest", config.ContainerRegistryURL, config.ContainerImageName))

		initialFileExists, _ := utils.FileExists("/path/to/test.json")
		renamedFileExists, _ := utils.FileExists("/path/to/config.json")

		assert.False(t, initialFileExists)
		assert.True(t, renamedFileExists)
	})

	t.Run("success case (customTlsCertificates)", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, "https://test-cert.com/cert.crt", httpmock.NewStringResponder(200, "testCert"))
		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		caCertsFile := "/etc/ssl/certs/ca-certificates.crt"
		caCertsTmpFile := "/tmp/ca-certificates.crt"
		registry := "some-registry"
		config := cnbBuildOptions{
			ContainerImageName:        "my-image",
			ContainerImageTag:         "0.0.1",
			ContainerRegistryURL:      registry,
			DockerConfigJSON:          "/path/to/config.json",
			CustomTLSCertificateLinks: []string{"https://test-cert.com/cert.crt", "https://test-cert.com/cert.crt"},
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(caCertsFile, []byte("test\n"))
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &cnbBuildTelemetry{}, &utils, &commonPipelineEnvironment, client)
		require.NoError(t, err)

		result, err := utils.FilesMock.FileRead(caCertsTmpFile)
		require.NoError(t, err)
		assert.Equal(t, "test\ntestCert\ntestCert\n", string(result))

		require.NoError(t, err)
		runner := utils.ExecMockRunner
		assert.Contains(t, runner.Env, "CNB_REGISTRY_AUTH={\"my-registry\":\"Basic dXNlcjpwYXNz\"}")
		assert.Contains(t, runner.Env, fmt.Sprintf("SSL_CERT_FILE=%s", caCertsTmpFile))
		assertLifecycleCalls(t, runner, 1)
		assert.Contains(t, runner.Calls[0].Params, fmt.Sprintf("%s/%s:%s", config.ContainerRegistryURL, config.ContainerImageName, config.ContainerImageTag))
	})

	t.Run("success case (additionalTags)", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "3.1.5",
			ContainerRegistryURL: imageRegistry,
			DockerConfigJSON:     "/path/to/config.json",
			AdditionalTags:       []string{"3", "3.1", "3.1", "3.1.5"},
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &cnbBuildTelemetry{}, &utils, &commonPipelineEnvironment, &piperhttp.Client{})
		require.NoError(t, err)

		runner := utils.ExecMockRunner
		assertLifecycleCalls(t, runner, 1)
		assert.Contains(t, runner.Calls[0].Params, fmt.Sprintf("%s/%s:%s", config.ContainerRegistryURL, config.ContainerImageName, config.ContainerImageTag))
		assert.Contains(t, runner.Calls[0].Params, fmt.Sprintf("%s/%s:3", config.ContainerRegistryURL, config.ContainerImageName))
		assert.Contains(t, runner.Calls[0].Params, fmt.Sprintf("%s/%s:3.1", config.ContainerRegistryURL, config.ContainerImageName))
		assert.Contains(t, runner.Calls[0].Params, fmt.Sprintf("%s/%s:3.1.5", config.ContainerRegistryURL, config.ContainerImageName))
	})

	t.Run("pom.xml exists (symlink for the target folder)", func(t *testing.T) {
		t.Parallel()
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "3.1.5",
			ContainerRegistryURL: imageRegistry,
			DockerConfigJSON:     "/path/to/config.json",
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.CurrentDir = "/jenkins"
		utils.FilesMock.AddDir("/jenkins")
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		utils.FilesMock.AddFile("/workspace/pom.xml", []byte("test"))
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &cnbBuildTelemetry{}, &utils, &cnbBuildCommonPipelineEnvironment{}, &piperhttp.Client{})
		require.NoError(t, err)

		runner := utils.ExecMockRunner
		assertLifecycleCalls(t, runner, 1)

		assert.True(t, utils.FilesMock.HasCreatedSymlink("/jenkins/target", "/workspace/target"))
	})

	t.Run("no pom.xml exists (no symlink for the target folder)", func(t *testing.T) {
		t.Parallel()
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "3.1.5",
			ContainerRegistryURL: imageRegistry,
			DockerConfigJSON:     "/path/to/config.json",
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.CurrentDir = "/jenkins"
		utils.FilesMock.AddDir("/jenkins")
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &cnbBuildTelemetry{}, &utils, &cnbBuildCommonPipelineEnvironment{}, &piperhttp.Client{})
		require.NoError(t, err)

		runner := utils.ExecMockRunner
		assertLifecycleCalls(t, runner, 1)

		assert.False(t, utils.FilesMock.HasCreatedSymlink("/jenkins/target", "/workspace/target"))
	})

	t.Run("error case: Invalid DockerConfigJSON file", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{
			ContainerImageTag:    "0.0.1",
			ContainerRegistryURL: imageRegistry,
			ContainerImageName:   "my-image",
			DockerConfigJSON:     "/path/to/config.json",
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":"dXNlcjpwYXNz"}}`))
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &cnbBuildTelemetry{}, &utils, &commonPipelineEnvironment, &piperhttp.Client{})
		assert.EqualError(t, err, "failed to generate CNB_REGISTRY_AUTH: json: cannot unmarshal string into Go struct field ConfigFile.auths of type types.AuthConfig")
	})

	t.Run("error case: DockerConfigJSON file not there (config.json)", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{
			ContainerImageTag:    "0.0.1",
			ContainerRegistryURL: imageRegistry,
			ContainerImageName:   "my-image",
			DockerConfigJSON:     "not-there/config.json",
		}

		utils := newCnbBuildTestsUtils()
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &cnbBuildTelemetry{}, &utils, &commonPipelineEnvironment, &piperhttp.Client{})
		assert.EqualError(t, err, "failed to generate CNB_REGISTRY_AUTH: could not read 'not-there/config.json'")
	})

	t.Run("error case: DockerConfigJSON file not there (not config.json)", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{
			ContainerImageTag:    "0.0.1",
			ContainerRegistryURL: imageRegistry,
			ContainerImageName:   "my-image",
			DockerConfigJSON:     "not-there",
		}

		utils := newCnbBuildTestsUtils()
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &cnbBuildTelemetry{}, &utils, &commonPipelineEnvironment, &piperhttp.Client{})
		assert.EqualError(t, err, "failed to rename DockerConfigJSON file 'not-there': renaming file 'not-there' is not supported, since it does not exist, or is not a leaf-entry")
	})

	t.Run("error case: dockerImage is not a valid builder", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{}

		utils := newCnbBuildTestsUtils()

		err := callCnbBuild(&config, &telemetry.CustomData{}, &cnbBuildTelemetry{}, &utils, &commonPipelineEnvironment, &piperhttp.Client{})
		assert.EqualError(t, err, "the provided dockerImage is not a valid builder: binary '/cnb/lifecycle/creator' not found")
	})

	t.Run("error case: builder image does not contain tls certificates", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{
			ContainerImageName:        "my-image",
			ContainerImageTag:         "0.0.1",
			ContainerRegistryURL:      imageRegistry,
			DockerConfigJSON:          "/path/to/config.json",
			Buildpacks:                []string{"test"},
			CustomTLSCertificateLinks: []string{"http://example.com/certs.pem"},
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &cnbBuildTelemetry{}, &utils, &commonPipelineEnvironment, &piperhttp.Client{})
		assert.EqualError(t, err, "failed to copy certificates: cannot copy '/etc/ssl/certs/ca-certificates.crt': file does not exist")
	})

	t.Run("success case (telemetry was added)", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		registry := "some-registry"
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "3.1.5",
			ContainerRegistryURL: registry,
			DockerConfigJSON:     "/path/to/config.json",
			ProjectDescriptor:    "project.toml",
			AdditionalTags:       []string{"latest"},
			Buildpacks:           []string{"paketobuildpacks/java", "gcr.io/paketo-buildpacks/node"},
			Bindings:             map[string]interface{}{"SECRET": map[string]string{"key": "KEY", "file": "a_file"}},
			Path:                 "target",
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		utils.FilesMock.AddDir("target")
		utils.FilesMock.AddFile("target/project.toml", []byte(`[project]
id = "test"
name = "test"
version = "1.0.0"

[build]
include = []
exclude = ["*.tar"]

[[build.buildpacks]]
uri = "some-buildpack"`))
		utils.FilesMock.AddFile("a_file", []byte(`{}`))
		utils.FilesMock.AddFile("target/somelib.jar", []byte(`FFFFFF`))

		addBuilderFiles(&utils)

		telemetryData := telemetry.CustomData{}
		cnbTelemetry := cnbBuildTelemetry{Version: 2}
		err := callCnbBuild(&config, &telemetryData, &cnbTelemetry, &utils, &commonPipelineEnvironment, &piperhttp.Client{})
		require.NoError(t, err)

		customDataAsString := telemetryData.Custom1
		customData := cnbBuildTelemetry{}
		err = json.Unmarshal([]byte(customDataAsString), &customData)

		require.NoError(t, err)
		assert.Equal(t, 2, customData.Version)
		require.Equal(t, 1, len(customData.Data))
		assert.Equal(t, "3.1.5", customData.Data[0].ImageTag)
		assert.Equal(t, "folder", string(customData.Data[0].Path))
		assert.Contains(t, customData.Data[0].AdditionalTags, "latest")
		assert.Contains(t, customData.Data[0].BindingKeys, "SECRET")

		assert.Contains(t, customData.Data[0].Buildpacks.FromConfig, "paketobuildpacks/java")
		assert.NotContains(t, customData.Data[0].Buildpacks.FromProjectDescriptor, "paketobuildpacks/java")
		assert.Contains(t, customData.Data[0].Buildpacks.FromProjectDescriptor, "<redacted>")
		assert.NotContains(t, customData.Data[0].Buildpacks.Overall, "<redacted>")
		assert.Contains(t, customData.Data[0].Buildpacks.Overall, "paketobuildpacks/java")

		assert.True(t, customData.Data[0].ProjectDescriptor.Used)
		assert.False(t, customData.Data[0].ProjectDescriptor.IncludeUsed)
		assert.True(t, customData.Data[0].ProjectDescriptor.ExcludeUsed)
	})

	t.Run("success case (build env telemetry was added)", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		registry := "some-registry"
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "3.1.5",
			ContainerRegistryURL: registry,
			ProjectDescriptor:    "project.toml",
			BuildEnvVars:         map[string]interface{}{"CONFIG_KEY": "var", "BP_JVM_VERSION": "8"},
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile("project.toml", []byte(`[project]
id = "test"

[build]
include = []

[[build.env]]
name='PROJECT_KEY'
value='var'

[[build.env]]
name='BP_NODE_VERSION'
value='11'

[[build.buildpacks]]
uri = "some-buildpack"
`))

		addBuilderFiles(&utils)

		telemetryData := telemetry.CustomData{}
		cnbTelemetry := cnbBuildTelemetry{Version: 2}
		err := callCnbBuild(&config, &telemetryData, &cnbTelemetry, &utils, &commonPipelineEnvironment, &piperhttp.Client{})
		require.NoError(t, err)

		customDataAsString := telemetryData.Custom1
		customData := cnbBuildTelemetry{}
		err = json.Unmarshal([]byte(customDataAsString), &customData)

		require.NoError(t, err)
		require.Equal(t, 1, len(customData.Data))
		assert.Contains(t, customData.Data[0].BuildEnv.KeysFromConfig, "CONFIG_KEY")
		assert.NotContains(t, customData.Data[0].BuildEnv.KeysFromProjectDescriptor, "CONFIG_KEY")
		assert.Contains(t, customData.Data[0].BuildEnv.KeysOverall, "CONFIG_KEY")

		assert.NotContains(t, customData.Data[0].BuildEnv.KeysFromConfig, "PROJECT_KEY")
		assert.Contains(t, customData.Data[0].BuildEnv.KeysFromProjectDescriptor, "PROJECT_KEY")
		assert.Contains(t, customData.Data[0].BuildEnv.KeysOverall, "PROJECT_KEY")

		assert.Equal(t, "8", customData.Data[0].BuildEnv.KeyValues["BP_JVM_VERSION"])
		assert.Equal(t, "11", customData.Data[0].BuildEnv.KeyValues["BP_NODE_VERSION"])
		assert.NotContains(t, customData.Data[0].BuildEnv.KeyValues, "PROJECT_KEY")

		assert.Contains(t, customData.Data[0].Buildpacks.Overall, "some-buildpack")
	})

	t.Run("success case (multiple images configured)", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{
			ContainerImageTag:    "3.1.5",
			ContainerRegistryURL: imageRegistry,
			DockerConfigJSON:     "/path/to/my-config.json",
			AdditionalTags:       []string{"3", "3.1", "3.1", "3.1.5"},
			MultipleImages:       []map[string]interface{}{{"ContainerImageName": "my-image-0"}, {"ContainerImageName": "my-image-1"}},
		}

		expectedImageCount := len(config.MultipleImages)

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		telemetryData := telemetry.CustomData{}
		err := callCnbBuild(&config, &telemetryData, &cnbBuildTelemetry{}, &utils, &commonPipelineEnvironment, &piperhttp.Client{})
		require.NoError(t, err)

		customDataAsString := telemetryData.Custom1
		customData := cnbBuildTelemetry{}
		err = json.Unmarshal([]byte(customDataAsString), &customData)
		require.Equal(t, expectedImageCount, len(customData.Data))

		runner := utils.ExecMockRunner
		require.Equal(t, expectedImageCount, len(runner.Calls))
		for i, call := range runner.Calls {
			assert.Equal(t, 4, len(customData.Data[i].AdditionalTags))
			assertLifecycleCalls(t, runner, i+1)
			containerImageName := fmt.Sprintf("my-image-%d", i)
			assert.Contains(t, call.Params, fmt.Sprintf("%s/%s:%s", config.ContainerRegistryURL, containerImageName, config.ContainerImageTag))
			assert.Contains(t, call.Params, fmt.Sprintf("%s/%s:3", config.ContainerRegistryURL, containerImageName))
			assert.Contains(t, call.Params, fmt.Sprintf("%s/%s:3.1", config.ContainerRegistryURL, containerImageName))
			assert.Contains(t, call.Params, fmt.Sprintf("%s/%s:3.1.5", config.ContainerRegistryURL, containerImageName))
		}

		assert.Equal(t, "my-image-0:3.1.5", commonPipelineEnvironment.container.imageNameTag)
	})

}
