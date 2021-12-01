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
	assert.Equal(t, analyzerPath, runner.Calls[0].Exec)
	assert.Equal(t, detectorPath, runner.Calls[1].Exec)
	assert.Equal(t, builderPath, runner.Calls[2].Exec)
	assert.Equal(t, restorerPath, runner.Calls[3].Exec)
	assert.Equal(t, exporterPath, runner.Calls[4].Exec)
}

func TestRunCnbBuild(t *testing.T) {
	t.Parallel()

	commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}

	t.Run("success case (registry with https)", func(t *testing.T) {
		t.Parallel()
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
		assert.Equal(t, []string{"-buildpacks", "/cnb/buildpacks", "-order", "/cnb/order.toml", "-platform", "/tmp/platform", "-no-color"}, runner.Calls[1].Params)
		assert.Equal(t, []string{"-buildpacks", "/cnb/buildpacks", "-platform", "/tmp/platform", "-no-color"}, runner.Calls[2].Params)
		assert.Equal(t, []string{"-no-color", fmt.Sprintf("%s/%s:%s", registry, config.ContainerImageName, config.ContainerImageTag)}, runner.Calls[4].Params)
		assert.Equal(t, commonPipelineEnvironment.container.registryURL, fmt.Sprintf("https://%s", registry))
		assert.Equal(t, commonPipelineEnvironment.container.imageNameTag, "my-image:0.0.1")
	})

	t.Run("success case (registry without https)", func(t *testing.T) {
		t.Parallel()
		registry := "some-registry"
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
		assert.Equal(t, []string{"-buildpacks", "/cnb/buildpacks", "-order", "/cnb/order.toml", "-platform", "/tmp/platform", "-no-color"}, runner.Calls[1].Params)
		assert.Equal(t, []string{"-buildpacks", "/cnb/buildpacks", "-platform", "/tmp/platform", "-no-color"}, runner.Calls[2].Params)
		assert.Equal(t, []string{"-no-color", fmt.Sprintf("%s/%s:%s", registry, config.ContainerImageName, config.ContainerImageTag)}, runner.Calls[4].Params)
		assert.Equal(t, commonPipelineEnvironment.container.registryURL, fmt.Sprintf("https://%s", registry))
		assert.Equal(t, commonPipelineEnvironment.container.imageNameTag, "my-image:0.0.1")
	})

	t.Run("success case (custom buildpacks and custom env variables, renaming docker conf file, additional tag)", func(t *testing.T) {
		t.Parallel()
		registry := "some-registry"
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
		assertLifecycleCalls(t, runner)
		assert.Equal(t, []string{"-buildpacks", "/tmp/buildpacks", "-order", "/tmp/buildpacks/order.toml", "-platform", "/tmp/platform", "-no-color"}, runner.Calls[1].Params)
		assert.Equal(t, []string{"-buildpacks", "/tmp/buildpacks", "-platform", "/tmp/platform", "-no-color"}, runner.Calls[2].Params)
		assert.Equal(t, []string{"-no-color", fmt.Sprintf("%s/%s:%s", registry, config.ContainerImageName, config.ContainerImageTag), fmt.Sprintf("%s/%s:latest", registry, config.ContainerImageName)}, runner.Calls[4].Params)
	})

	t.Run("success case (customTlsCertificates)", func(t *testing.T) {
		t.Parallel()

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
			Buildpacks:                []string{"test"},
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
		assert.Equal(t, []string{"-buildpacks", "/tmp/buildpacks", "-order", "/tmp/buildpacks/order.toml", "-platform", "/tmp/platform", "-no-color"}, runner.Calls[1].Params)
		assert.Equal(t, []string{"-buildpacks", "/tmp/buildpacks", "-platform", "/tmp/platform", "-no-color"}, runner.Calls[2].Params)
		assert.Equal(t, []string{"-no-color", fmt.Sprintf("%s/%s:%s", registry, config.ContainerImageName, config.ContainerImageTag)}, runner.Calls[4].Params)
	})

	t.Run("success case (additionalTags)", func(t *testing.T) {
		t.Parallel()

		registry := "some-registry"
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "3.1.5",
			ContainerRegistryURL: registry,
			DockerConfigJSON:     "/path/to/config.json",
			Buildpacks:           []string{"test"},
			AdditionalTags:       []string{"3", "3.1", "3.1", "3.1.5"},
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		err := runCnbBuild(&config, &telemetry.CustomData{}, &utils, &commonPipelineEnvironment, &piperhttp.Client{})
		assert.NoError(t, err)

		runner := utils.ExecMockRunner
		assert.Equal(t, exporterPath, runner.Calls[4].Exec)
		assert.ElementsMatch(t, []string{"-no-color", fmt.Sprintf("%s/%s:%s", registry, config.ContainerImageName, config.ContainerImageTag), fmt.Sprintf("%s/%s:3", registry, config.ContainerImageName), fmt.Sprintf("%s/%s:3.1", registry, config.ContainerImageName)}, runner.Calls[4].Params)
	})

	t.Run("error case: Invalid DockerConfigJSON file", func(t *testing.T) {
		t.Parallel()
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
		config := cnbBuildOptions{}

		utils := newCnbBuildTestsUtils()

		err := runCnbBuild(&config, nil, &utils, &commonPipelineEnvironment, &piperhttp.Client{})
		assert.EqualError(t, err, "the provided dockerImage is not a valid builder: binary '/cnb/lifecycle/analyzer' not found")
	})

	t.Run("error case: builder image does not contain tls certificates", func(t *testing.T) {
		t.Parallel()

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
