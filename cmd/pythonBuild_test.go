package cmd

import (
	"fmt"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/telemetry"

	"github.com/stretchr/testify/assert"

	"github.com/SAP/jenkins-library/pkg/mock"
)

type pythonBuildMockUtils struct {
	t      *testing.T
	config *pythonBuildOptions
	*mock.ExecMockRunner
	*mock.FilesMock
}

type puthonBuildMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock

	clientOptions []piperhttp.ClientOptions // set by mock
	fileUploads   map[string]string         // set by mock
}

func newPythonBuildTestsUtils() pythonBuildMockUtils {
	utils := pythonBuildMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func (f *pythonBuildMockUtils) GetConfig() *pythonBuildOptions {
	return f.config
}

func TestRunPythonBuild(t *testing.T) {

	t.Run("success - build", func(t *testing.T) {
		config := pythonBuildOptions{}
		utils := newPythonBuildTestsUtils()
		telemetryData := telemetry.CustomData{}

		err := runPythonBuild(&config, &telemetryData, utils)
		assert.NoError(t, err)
		assert.Equal(t, "python3", utils.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{"setup.py", "sdist", "bdist_wheel"}, utils.ExecMockRunner.Calls[0].Params)
	})

	t.Run("failure - build failure", func(t *testing.T) {
		config := pythonBuildOptions{}
		utils := newPythonBuildTestsUtils()
		utils.ShouldFailOnCommand = map[string]error{"python3 setup.py sdist bdist_wheel": fmt.Errorf("build failure")}
		telemetryData := telemetry.CustomData{}

		err := runPythonBuild(&config, &telemetryData, utils)
		assert.EqualError(t, err, "Python build failed with error: build failure")
	})

	t.Run("success - publishes binaries", func(t *testing.T) {
		config := pythonBuildOptions{
			Publish:                  true,
			TargetRepositoryURL:      "https://my.target.repository.local",
			TargetRepositoryUser:     "user",
			TargetRepositoryPassword: "password",
		}
		utils := newPythonBuildTestsUtils()
		telemetryData := telemetry.CustomData{}

		err := runPythonBuild(&config, &telemetryData, utils)
		assert.NoError(t, err)
		assert.Equal(t, "python3", utils.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{"setup.py", "sdist", "bdist_wheel"}, utils.ExecMockRunner.Calls[0].Params)
		assert.Equal(t, "python3", utils.ExecMockRunner.Calls[1].Exec)
		assert.Equal(t, []string{"-m", "pip", "install", "--upgrade", "twine"}, utils.ExecMockRunner.Calls[1].Params)
		assert.Equal(t, "twine", utils.ExecMockRunner.Calls[2].Exec)
		assert.Equal(t, []string{"upload", "--username", config.TargetRepositoryUser,
			"--password", config.TargetRepositoryPassword, "--repository-url", config.TargetRepositoryURL,
			"dist/*"}, utils.ExecMockRunner.Calls[2].Params)
	})

	t.Run("success - create BOM", func(t *testing.T) {
		config := pythonBuildOptions{
			CreateBOM: true,
		}
		utils := newPythonBuildTestsUtils()
		telemetryData := telemetry.CustomData{}

		err := runPythonBuild(&config, &telemetryData, utils)
		assert.NoError(t, err)
		assert.Equal(t, "python3", utils.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{"setup.py", "sdist", "bdist_wheel"}, utils.ExecMockRunner.Calls[0].Params)
		assert.Equal(t, "python3", utils.ExecMockRunner.Calls[1].Exec)
		assert.Equal(t, []string{"-m", "pip", "install", "--upgrade", "cyclonedx-bom"}, utils.ExecMockRunner.Calls[1].Params)
		assert.Equal(t, "cyclonedx-bom", utils.ExecMockRunner.Calls[2].Exec)
		assert.Equal(t, []string{"--e", "--output", "bom.xml"}, utils.ExecMockRunner.Calls[2].Params)
	})

	t.Run("failure - install pre-requisites for BOM creation", func(t *testing.T) {
		config := pythonBuildOptions{
			CreateBOM: true,
		}
		utils := newPythonBuildTestsUtils()
		utils.ShouldFailOnCommand = map[string]error{"python3 -m pip install --upgrade cyclonedx-bom": fmt.Errorf("install failure")}
		telemetryData := telemetry.CustomData{}

		err := runPythonBuild(&config, &telemetryData, utils)
		assert.EqualError(t, err, "BOM creation failed: install failure")
	})

	t.Run("failure - install pre-requisites for Twine upload", func(t *testing.T) {
		config := pythonBuildOptions{
			Publish: true,
		}
		utils := newPythonBuildTestsUtils()
		utils.ShouldFailOnCommand = map[string]error{"python3 -m pip install --upgrade twine": fmt.Errorf("install failure")}
		telemetryData := telemetry.CustomData{}

		err := runPythonBuild(&config, &telemetryData, utils)
		assert.EqualError(t, err, "failed to publish: install failure")
	})
}
