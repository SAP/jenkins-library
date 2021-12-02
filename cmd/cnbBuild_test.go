package cmd

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func newCnbBuildTestsUtils() cnbutils.MockUtils {
	utils := cnbutils.MockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
		DockerMock:     &cnbutils.DockerMock{},
	}
	return utils
}

func addBuilderFiles(utils *cnbutils.MockUtils) {
	for _, path := range []string{analyzerPath, detectorPath, builderPath, restorerPath, exporterPath} {
		utils.FilesMock.AddFile(path, []byte(`xyz`))
	}
}

func assertLifecycleCalls(t *testing.T, runner *mock.ExecMockRunner) {
	assert.Equal(t, creatorPath, runner.Calls[0].Exec)
	for _, arg := range []string{"-no-color", "-buildpacks", "/cnb/buildpacks", "-order", "/cnb/order.toml", "-platform", "/tmp/platform"} {
		assert.Contains(t, runner.Calls[0].Params, arg)
	}
}

func TestRunCnbBuild(t *testing.T) {
	t.Parallel()

	t.Run("success case (registry with https)", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		registry := "some-registry"
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "0.0.1",
			ContainerRegistryURL: fmt.Sprintf("https://%s", registry),
			DockerConfigJSON:     "/path/to/config.json",
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		err := runCnbBuild(&config, &telemetry.CustomData{}, &utils, &commonPipelineEnvironment, &piperhttp.Client{})

		assert.NoError(t, err)
		runner := utils.ExecMockRunner
		assert.Contains(t, runner.Env, "CNB_REGISTRY_AUTH={\"my-registry\":\"Basic dXNlcjpwYXNz\"}")
		assertLifecycleCalls(t, runner)
		assert.Contains(t, runner.Calls[0].Params, fmt.Sprintf("%s/%s:%s", registry, config.ContainerImageName, config.ContainerImageTag))
		assert.Equal(t, fmt.Sprintf("https://%s", registry), commonPipelineEnvironment.container.registryURL)
		assert.Equal(t, "my-image:0.0.1", commonPipelineEnvironment.container.imageNameTag)
	})

	t.Run("success case (registry without https)", func(t *testing.T) {
		t.Parallel()
		registry := "some-registry"
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "0.0.1",
			ContainerRegistryURL: registry,
			DockerConfigJSON:     "/path/to/config.json",
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		err := runCnbBuild(&config, &telemetry.CustomData{}, &utils, &commonPipelineEnvironment, &piperhttp.Client{})

		assert.NoError(t, err)
		runner := utils.ExecMockRunner
		assert.Contains(t, runner.Env, "CNB_REGISTRY_AUTH={\"my-registry\":\"Basic dXNlcjpwYXNz\"}")
		assertLifecycleCalls(t, runner)
		assert.Contains(t, runner.Calls[0].Params, fmt.Sprintf("%s/%s:%s", registry, config.ContainerImageName, config.ContainerImageTag))
		assert.Equal(t, fmt.Sprintf("https://%s", registry), commonPipelineEnvironment.container.registryURL)
		assert.Equal(t, "my-image:0.0.1", commonPipelineEnvironment.container.imageNameTag)
	})

	t.Run("success case (custom buildpacks and custom env variables, renaming docker conf file, additional tag)", func(t *testing.T) {
		t.Parallel()
		registry := "some-registry"
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "0.0.1",
			ContainerRegistryURL: registry,
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

		err := runCnbBuild(&config, &telemetry.CustomData{}, &utils, &commonPipelineEnvironment, &piperhttp.Client{})

		assert.NoError(t, err)
		runner := utils.ExecMockRunner
		assert.Contains(t, runner.Env, "CNB_REGISTRY_AUTH={\"my-registry\":\"Basic dXNlcjpwYXNz\"}")
		assert.Equal(t, creatorPath, runner.Calls[0].Exec)
		assert.Contains(t, runner.Calls[0].Params, "/tmp/buildpacks")
		assert.Contains(t, runner.Calls[0].Params, "/tmp/buildpacks/order.toml")
		assert.Contains(t, runner.Calls[0].Params, fmt.Sprintf("%s/%s:%s", registry, config.ContainerImageName, config.ContainerImageTag))
		assert.Contains(t, runner.Calls[0].Params, fmt.Sprintf("%s/%s:latest", registry, config.ContainerImageName))
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

		err := runCnbBuild(&config, &telemetry.CustomData{}, &utils, &commonPipelineEnvironment, client)
		assert.NoError(t, err)

		result, err := utils.FilesMock.FileRead(caCertsTmpFile)
		assert.NoError(t, err)
		assert.Equal(t, "test\ntestCert\ntestCert\n", string(result))

		assert.NoError(t, err)
		runner := utils.ExecMockRunner
		assert.Contains(t, runner.Env, "CNB_REGISTRY_AUTH={\"my-registry\":\"Basic dXNlcjpwYXNz\"}")
		assert.Contains(t, runner.Env, fmt.Sprintf("SSL_CERT_FILE=%s", caCertsTmpFile))
		assertLifecycleCalls(t, runner)
		assert.Contains(t, runner.Calls[0].Params, fmt.Sprintf("%s/%s:%s", registry, config.ContainerImageName, config.ContainerImageTag))
	})

	t.Run("success case (additionalTags)", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		registry := "some-registry"
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "3.1.5",
			ContainerRegistryURL: registry,
			DockerConfigJSON:     "/path/to/config.json",
			AdditionalTags:       []string{"3", "3.1", "3.1", "3.1.5"},
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		err := runCnbBuild(&config, &telemetry.CustomData{}, &utils, &commonPipelineEnvironment, &piperhttp.Client{})
		assert.NoError(t, err)

		runner := utils.ExecMockRunner
		assertLifecycleCalls(t, runner)
		assert.Contains(t, runner.Calls[0].Params, fmt.Sprintf("%s/%s:%s", registry, config.ContainerImageName, config.ContainerImageTag))
		assert.Contains(t, runner.Calls[0].Params, fmt.Sprintf("%s/%s:3", registry, config.ContainerImageName))
		assert.Contains(t, runner.Calls[0].Params, fmt.Sprintf("%s/%s:3.1", registry, config.ContainerImageName))
		assert.Contains(t, runner.Calls[0].Params, fmt.Sprintf("%s/%s:3.1.5", registry, config.ContainerImageName))
	})

	t.Run("error case: Invalid DockerConfigJSON file", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{
			ContainerImageName: "my-image",
			DockerConfigJSON:   "/path/to/config.json",
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":"dXNlcjpwYXNz"}}`))
		addBuilderFiles(&utils)

		err := runCnbBuild(&config, nil, &utils, &commonPipelineEnvironment, &piperhttp.Client{})
		assert.EqualError(t, err, "failed to parse DockerConfigJSON file '/path/to/config.json': json: cannot unmarshal string into Go struct field ConfigFile.auths of type types.AuthConfig")
	})

	t.Run("error case: DockerConfigJSON file not there (config.json)", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{
			ContainerImageName: "my-image",
			DockerConfigJSON:   "not-there/config.json",
		}

		utils := newCnbBuildTestsUtils()
		addBuilderFiles(&utils)

		err := runCnbBuild(&config, nil, &utils, &commonPipelineEnvironment, &piperhttp.Client{})
		assert.EqualError(t, err, "failed to read DockerConfigJSON file 'not-there/config.json': could not read 'not-there/config.json'")
	})

	t.Run("error case: DockerConfigJSON file not there (not config.json)", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{
			ContainerImageName: "my-image",
			DockerConfigJSON:   "not-there",
		}

		utils := newCnbBuildTestsUtils()
		addBuilderFiles(&utils)

		err := runCnbBuild(&config, nil, &utils, &commonPipelineEnvironment, &piperhttp.Client{})
		assert.EqualError(t, err, "failed to rename DockerConfigJSON file 'not-there': renaming file 'not-there' is not supported, since it does not exist, or is not a leaf-entry")
	})

	t.Run("error case: dockerImage is not a valid builder", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{}

		utils := newCnbBuildTestsUtils()

		err := runCnbBuild(&config, nil, &utils, &commonPipelineEnvironment, &piperhttp.Client{})
		assert.EqualError(t, err, "the provided dockerImage is not a valid builder: binary '/cnb/lifecycle/analyzer' not found")
	})

	t.Run("error case: builder image does not contain tls certificates", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		registry := "some-registry"
		config := cnbBuildOptions{
			ContainerImageName:        "my-image",
			ContainerImageTag:         "0.0.1",
			ContainerRegistryURL:      registry,
			DockerConfigJSON:          "/path/to/config.json",
			Buildpacks:                []string{"test"},
			CustomTLSCertificateLinks: []string{"http://example.com/certs.pem"},
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		err := runCnbBuild(&config, nil, &utils, &commonPipelineEnvironment, &piperhttp.Client{})
		assert.EqualError(t, err, "failed to copy certificates: cannot copy '/etc/ssl/certs/ca-certificates.crt': file does not exist")
	})
}
