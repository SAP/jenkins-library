package cmd

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/telemetry"
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
	for _, path := range []string{detectorPath, builderPath, exporterPath} {
		utils.FilesMock.AddFile(path, []byte(`xyz`))
	}
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

		httpClient := &kanikoMockClient{
			responseBody: "testCert",
		}

		err := runCnbBuild(&config, &telemetry.CustomData{}, &utils, &commonPipelineEnvironment, httpClient)

		assert.NoError(t, err)
		runner := utils.ExecMockRunner
		assert.Contains(t, runner.Env, "CNB_REGISTRY_AUTH={\"my-registry\":\"Basic dXNlcjpwYXNz\"}")
		assert.Equal(t, "/cnb/lifecycle/detector", runner.Calls[0].Exec)
		assert.Equal(t, "/cnb/lifecycle/builder", runner.Calls[1].Exec)
		assert.Equal(t, "/cnb/lifecycle/exporter", runner.Calls[2].Exec)
		assert.Equal(t, []string{"-buildpacks", "/cnb/buildpacks", "-order", "/cnb/order.toml", "-platform", "/platform"}, runner.Calls[0].Params)
		assert.Equal(t, []string{"-buildpacks", "/cnb/buildpacks", "-platform", "/platform"}, runner.Calls[1].Params)
		assert.Equal(t, []string{fmt.Sprintf("%s/%s:%s", registry, config.ContainerImageName, config.ContainerImageTag), fmt.Sprintf("%s/%s:latest", registry, config.ContainerImageName)}, runner.Calls[2].Params)
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

		httpClient := &kanikoMockClient{
			responseBody: "testCert",
		}

		err := runCnbBuild(&config, &telemetry.CustomData{}, &utils, &commonPipelineEnvironment, httpClient)

		assert.NoError(t, err)
		runner := utils.ExecMockRunner
		assert.Contains(t, runner.Env, "CNB_REGISTRY_AUTH={\"my-registry\":\"Basic dXNlcjpwYXNz\"}")
		assert.Equal(t, "/cnb/lifecycle/detector", runner.Calls[0].Exec)
		assert.Equal(t, "/cnb/lifecycle/builder", runner.Calls[1].Exec)
		assert.Equal(t, "/cnb/lifecycle/exporter", runner.Calls[2].Exec)
		assert.Equal(t, []string{"-buildpacks", "/cnb/buildpacks", "-order", "/cnb/order.toml", "-platform", "/platform"}, runner.Calls[0].Params)
		assert.Equal(t, []string{"-buildpacks", "/cnb/buildpacks", "-platform", "/platform"}, runner.Calls[1].Params)
		assert.Equal(t, []string{fmt.Sprintf("%s/%s:%s", registry, config.ContainerImageName, config.ContainerImageTag), fmt.Sprintf("%s/%s:latest", registry, config.ContainerImageName)}, runner.Calls[2].Params)
	})

	t.Run("success case (custom buildpacks and custom env variables, renaming docker conf file)", func(t *testing.T) {
		t.Parallel()
		registry := "some-registry"
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "0.0.1",
			ContainerRegistryURL: registry,
			DockerConfigJSON:     "/path/to/test.json",
			Buildpacks:           []string{"test"},
			BuildEnvVars:         []string{"FOO=BAR"},
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		httpClient := &kanikoMockClient{
			responseBody: "testCert",
		}

		err := runCnbBuild(&config, &telemetry.CustomData{}, &utils, &commonPipelineEnvironment, httpClient)

		assert.NoError(t, err)
		runner := utils.ExecMockRunner
		assert.Contains(t, runner.Env, "CNB_REGISTRY_AUTH={\"my-registry\":\"Basic dXNlcjpwYXNz\"}")
		assert.Equal(t, "/cnb/lifecycle/detector", runner.Calls[0].Exec)
		assert.Equal(t, "/cnb/lifecycle/builder", runner.Calls[1].Exec)
		assert.Equal(t, "/cnb/lifecycle/exporter", runner.Calls[2].Exec)
		assert.Equal(t, []string{"-buildpacks", "/tmp/buildpacks", "-order", "/tmp/buildpacks/order.toml", "-platform", "/tmp/platform"}, runner.Calls[0].Params)
		assert.Equal(t, []string{"-buildpacks", "/tmp/buildpacks", "-platform", "/tmp/platform"}, runner.Calls[1].Params)
		assert.Equal(t, []string{fmt.Sprintf("%s/%s:%s", registry, config.ContainerImageName, config.ContainerImageTag), fmt.Sprintf("%s/%s:latest", registry, config.ContainerImageName)}, runner.Calls[2].Params)
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

		httpClient := &kanikoMockClient{
			responseBody: "testCert",
		}

		err := runCnbBuild(&config, nil, &utils, &commonPipelineEnvironment, httpClient)
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

		httpClient := &kanikoMockClient{
			responseBody: "testCert",
		}

		err := runCnbBuild(&config, nil, &utils, &commonPipelineEnvironment, httpClient)
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

		httpClient := &kanikoMockClient{
			responseBody: "testCert",
		}

		err := runCnbBuild(&config, nil, &utils, &commonPipelineEnvironment, httpClient)
		assert.EqualError(t, err, "failed to rename DockerConfigJSON file 'not-there': renaming file 'not-there' is not supported, since it does not exist, or is not a leaf-entry")
	})

	t.Run("error case: dockerImage is not a valid builder", func(t *testing.T) {
		t.Parallel()
		config := cnbBuildOptions{}

		utils := newCnbBuildTestsUtils()

		httpClient := &kanikoMockClient{
			responseBody: "testCert",
		}

		err := runCnbBuild(&config, nil, &utils, &commonPipelineEnvironment, httpClient)
		assert.EqualError(t, err, "the provided dockerImage is not a valid builder")
	})
}
