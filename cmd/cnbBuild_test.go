package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/stretchr/testify/assert"
)

type cnbBuildMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newCnbBuildTestsUtils() cnbBuildMockUtils {
	utils := cnbBuildMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunCnbBuild(t *testing.T) {
	t.Parallel()

	commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}

	t.Run("success case", func(t *testing.T) {
		t.Parallel()
		config := cnbBuildOptions{
			ContainerImage:   "my-image",
			DockerConfigJSON: "/path/to/config.json",
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))

		err := runCnbBuild(&config, &telemetry.CustomData{}, utils, &commonPipelineEnvironment)

		assert.NoError(t, err)
		runner := utils.ExecMockRunner
		assert.Contains(t, runner.Env, "CNB_REGISTRY_AUTH={\"my-registry\":\"Basic dXNlcjpwYXNz\"}")
		assert.Equal(t, "/cnb/lifecycle/detector", runner.Calls[0].Exec)
		assert.Equal(t, "/cnb/lifecycle/builder", runner.Calls[1].Exec)
		assert.Equal(t, "/cnb/lifecycle/exporter", runner.Calls[2].Exec)
		assert.Equal(t, []string{config.ContainerImage}, runner.Calls[2].Params)
	})

	t.Run("error case: Invalid DockerConfigJSON file", func(t *testing.T) {
		t.Parallel()
		config := cnbBuildOptions{
			ContainerImage:   "my-image",
			DockerConfigJSON: "/path/to/config.json",
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":"dXNlcjpwYXNz"}}`))

		err := runCnbBuild(&config, nil, utils, &commonPipelineEnvironment)
		assert.EqualError(t, err, "failed to parse DockerConfigJSON file '/path/to/config.json': json: cannot unmarshal string into Go struct field ConfigFile.auths of type types.AuthConfig")
	})

	t.Run("error case: DockerConfigJSON file not there", func(t *testing.T) {
		t.Parallel()
		config := cnbBuildOptions{
			ContainerImage:   "my-image",
			DockerConfigJSON: "not-there",
		}

		utils := newCnbBuildTestsUtils()
		err := runCnbBuild(&config, nil, utils, &commonPipelineEnvironment)
		assert.EqualError(t, err, "failed to read DockerConfigJSON file 'not-there': could not read 'not-there'")
	})
}
