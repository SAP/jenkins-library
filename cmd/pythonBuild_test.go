//go:build unit
// +build unit

package cmd

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/telemetry"

	"github.com/stretchr/testify/assert"
)

type pythonBuildMockUtils struct {
	config *pythonBuildOptions
	*mock.ExecMockRunner
	*mock.FilesMock
}

const minimalSetupPyFileContent = "from setuptools import setup\n\nsetup(name='MyPackageName',version='1.0.0')"

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
	cpe := pythonBuildCommonPipelineEnvironment{}
	// utils := newPythonBuildTestsUtils()

	SetConfigOptions(ConfigCommandOptions{
		// OpenFile: utils.FilesMock.OpenFile,
		OpenFile: config.OpenPiperFile,
	})

	t.Run("success - build", func(t *testing.T) {
		config := pythonBuildOptions{
			VirtualEnvironmentName: "dummy",
		}
		utils := newPythonBuildTestsUtils()
		utils.AddFile("setup.py", []byte(minimalSetupPyFileContent))
		utils.AddDir("dummy")
		telemetryData := telemetry.CustomData{}

		err := runPythonBuild(&config, &telemetryData, utils, &cpe)
		assert.NoError(t, err)
		// assert.Equal(t, 3, len(utils.ExecMockRunner.Calls))
		assert.Equal(t, "python3", utils.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{"-m", "venv", "dummy"}, utils.ExecMockRunner.Calls[0].Params)
	})

	t.Run("failure - build failure", func(t *testing.T) {
		config := pythonBuildOptions{}
		utils := newPythonBuildTestsUtils()
		utils.AddFile("setup.py", []byte(minimalSetupPyFileContent))
		utils.ShouldFailOnCommand = map[string]error{"python setup.py sdist bdist_wheel": fmt.Errorf("build failure")}
		telemetryData := telemetry.CustomData{}

		err := runPythonBuild(&config, &telemetryData, utils, &cpe)
		assert.EqualError(t, err, "failed to build python project: build failure")
	})

	t.Run("success - publishes binaries", func(t *testing.T) {
		config := pythonBuildOptions{
			Publish:                  true,
			TargetRepositoryURL:      "https://my.target.repository.local",
			TargetRepositoryUser:     "user",
			TargetRepositoryPassword: "password",
			VirtualEnvironmentName:   "dummy",
		}
		utils := newPythonBuildTestsUtils()
		utils.AddFile("setup.py", []byte(minimalSetupPyFileContent))
		utils.AddDir("dummy")
		telemetryData := telemetry.CustomData{}

		err := runPythonBuild(&config, &telemetryData, utils, &cpe)
		assert.NoError(t, err)
		assert.Equal(t, "python3", utils.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{"-m", "venv", config.VirtualEnvironmentName}, utils.ExecMockRunner.Calls[0].Params)
		assert.Equal(t, "bash", utils.ExecMockRunner.Calls[1].Exec)
		assert.Equal(t, []string{"-c", "source " + filepath.Join("dummy", "bin", "activate")}, utils.ExecMockRunner.Calls[1].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[2].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "wheel"}, utils.ExecMockRunner.Calls[2].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "python"), utils.ExecMockRunner.Calls[3].Exec)
		assert.Equal(t, []string{"setup.py", "sdist", "bdist_wheel"}, utils.ExecMockRunner.Calls[3].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[4].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "twine"}, utils.ExecMockRunner.Calls[4].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "twine"), utils.ExecMockRunner.Calls[5].Exec)
		assert.Equal(t, []string{"upload", "--username", config.TargetRepositoryUser,
			"--password", config.TargetRepositoryPassword, "--repository-url", config.TargetRepositoryURL,
			"--disable-progress-bar", "dist/*"}, utils.ExecMockRunner.Calls[5].Params)
	})

	t.Run("success - create BOM", func(t *testing.T) {
		config := pythonBuildOptions{
			CreateBOM:              true,
			Publish:                false,
			VirtualEnvironmentName: "dummy",
		}
		utils := newPythonBuildTestsUtils()
		utils.AddFile("setup.py", []byte(minimalSetupPyFileContent))
		utils.AddDir("dummy")
		telemetryData := telemetry.CustomData{}

		err := runPythonBuild(&config, &telemetryData, utils, &cpe)
		assert.NoError(t, err)
		assert.Equal(t, "python3", utils.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{"-m", "venv", config.VirtualEnvironmentName}, utils.ExecMockRunner.Calls[0].Params)
		assert.Equal(t, "bash", utils.ExecMockRunner.Calls[1].Exec)
		assert.Equal(t, []string{"-c", "source " + filepath.Join("dummy", "bin", "activate")}, utils.ExecMockRunner.Calls[1].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[2].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "wheel"}, utils.ExecMockRunner.Calls[2].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "python"), utils.ExecMockRunner.Calls[3].Exec)
		assert.Equal(t, []string{"setup.py", "sdist", "bdist_wheel"}, utils.ExecMockRunner.Calls[3].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[4].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "cyclonedx-bom==6.1.1"}, utils.ExecMockRunner.Calls[4].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "cyclonedx-py"), utils.ExecMockRunner.Calls[5].Exec)
		assert.Equal(t, []string{"env", "--output-file", "bom-pip.xml", "--output-format", "XML", "--spec-version", "1.4"}, utils.ExecMockRunner.Calls[5].Params)
	})
}
