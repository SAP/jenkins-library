//go:build unit
// +build unit

package cmd

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/python"
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
		assert.Equal(t, []string{"install", "--upgrade", "setuptools"}, utils.ExecMockRunner.Calls[2].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[3].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "wheel"}, utils.ExecMockRunner.Calls[3].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "python"), utils.ExecMockRunner.Calls[4].Exec)
		assert.Equal(t, []string{"setup.py", "sdist", "bdist_wheel"}, utils.ExecMockRunner.Calls[4].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[5].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "twine"}, utils.ExecMockRunner.Calls[5].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "twine"), utils.ExecMockRunner.Calls[6].Exec)
		assert.Equal(t, []string{"upload", "--username", config.TargetRepositoryUser,
			"--password", config.TargetRepositoryPassword, "--repository-url", config.TargetRepositoryURL,
			"--disable-progress-bar", "dist/*"}, utils.ExecMockRunner.Calls[6].Params)
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
		assert.Equal(t, []string{"install", "--upgrade", "setuptools"}, utils.ExecMockRunner.Calls[2].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[3].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "wheel"}, utils.ExecMockRunner.Calls[3].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "python"), utils.ExecMockRunner.Calls[4].Exec)
		assert.Equal(t, []string{"setup.py", "sdist", "bdist_wheel"}, utils.ExecMockRunner.Calls[4].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[5].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "."}, utils.ExecMockRunner.Calls[5].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[6].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "cyclonedx-bom==7.3.0"}, utils.ExecMockRunner.Calls[6].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "cyclonedx-py"), utils.ExecMockRunner.Calls[7].Exec)
		assert.Equal(t, []string{"env", "--output-file", "bom-pip.xml", "--output-format", "XML", "--spec-version", "1.4"}, utils.ExecMockRunner.Calls[7].Params)
	})
}

func TestRunPythonBuildWithToml(t *testing.T) {
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
		utils.AddFile("pyproject.toml", []byte(minimalSetupPyFileContent))
		utils.AddDir("dummy")
		telemetryData := telemetry.CustomData{}

		err := runPythonBuild(&config, &telemetryData, utils, &cpe)
		assert.NoError(t, err)
		// assert.Equal(t, 3, len(utils.ExecMockRunner.Calls))
		assert.Equal(t, "python3", utils.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{"-m", "venv", "dummy"}, utils.ExecMockRunner.Calls[0].Params)
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
		utils.AddFile("pyproject.toml", []byte(minimalSetupPyFileContent))
		utils.AddDir("dummy")
		telemetryData := telemetry.CustomData{}

		err := runPythonBuild(&config, &telemetryData, utils, &cpe)
		assert.NoError(t, err)
		assert.Equal(t, "python3", utils.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{"-m", "venv", config.VirtualEnvironmentName}, utils.ExecMockRunner.Calls[0].Params)
		assert.Equal(t, "bash", utils.ExecMockRunner.Calls[1].Exec)
		assert.Equal(t, []string{"-c", "source " + filepath.Join("dummy", "bin", "activate")}, utils.ExecMockRunner.Calls[1].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[2].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "setuptools"}, utils.ExecMockRunner.Calls[2].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[3].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "pip"}, utils.ExecMockRunner.Calls[3].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[4].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "."}, utils.ExecMockRunner.Calls[4].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[5].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "build"}, utils.ExecMockRunner.Calls[5].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[6].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "wheel"}, utils.ExecMockRunner.Calls[6].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "python"), utils.ExecMockRunner.Calls[7].Exec)
		assert.Equal(t, []string{"-m", "build", "--no-isolation"}, utils.ExecMockRunner.Calls[7].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[8].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "twine"}, utils.ExecMockRunner.Calls[8].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "twine"), utils.ExecMockRunner.Calls[9].Exec)
		assert.Equal(t, []string{"upload", "--username", config.TargetRepositoryUser,
			"--password", config.TargetRepositoryPassword, "--repository-url", config.TargetRepositoryURL,
			"--disable-progress-bar", "dist/*"}, utils.ExecMockRunner.Calls[9].Params)
	})

	t.Run("success - create BOM", func(t *testing.T) {
		config := pythonBuildOptions{
			CreateBOM:              true,
			Publish:                false,
			VirtualEnvironmentName: "dummy",
		}
		utils := newPythonBuildTestsUtils()
		utils.AddFile("pyproject.toml", []byte(minimalSetupPyFileContent))
		utils.AddDir("dummy")
		telemetryData := telemetry.CustomData{}

		err := runPythonBuild(&config, &telemetryData, utils, &cpe)
		assert.NoError(t, err)
		assert.Equal(t, "python3", utils.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{"-m", "venv", config.VirtualEnvironmentName}, utils.ExecMockRunner.Calls[0].Params)
		assert.Equal(t, "bash", utils.ExecMockRunner.Calls[1].Exec)
		assert.Equal(t, []string{"-c", "source " + filepath.Join("dummy", "bin", "activate")}, utils.ExecMockRunner.Calls[1].Params)
		assert.Equal(t, []string{"install", "--upgrade", "setuptools"}, utils.ExecMockRunner.Calls[2].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[2].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "pip"}, utils.ExecMockRunner.Calls[3].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[3].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "."}, utils.ExecMockRunner.Calls[4].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[4].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "build"}, utils.ExecMockRunner.Calls[5].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[5].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "wheel"}, utils.ExecMockRunner.Calls[6].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[6].Exec)
		assert.Equal(t, []string{"-m", "build", "--no-isolation"}, utils.ExecMockRunner.Calls[7].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "python"), utils.ExecMockRunner.Calls[7].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "."}, utils.ExecMockRunner.Calls[8].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[8].Exec)
		assert.Equal(t, []string{"install", "--upgrade", "--root-user-action=ignore", "cyclonedx-bom==7.3.0"}, utils.ExecMockRunner.Calls[9].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "pip"), utils.ExecMockRunner.Calls[9].Exec)
		assert.Equal(t, []string{"env", "--output-file", "bom-pip.xml", "--output-format", "XML", "--spec-version", "1.4"}, utils.ExecMockRunner.Calls[10].Params)
		assert.Equal(t, filepath.Join("dummy", "bin", "cyclonedx-py"), utils.ExecMockRunner.Calls[10].Exec)
	})
}

func TestRunPythonBuildWithTests(t *testing.T) {
	cpe := pythonBuildCommonPipelineEnvironment{}

	SetConfigOptions(ConfigCommandOptions{
		OpenFile: config.OpenPiperFile,
	})

	// Each subtest runs for both build descriptors. Descriptor-specific call
	// indices (e.g. how many pip installs precede the build) are not asserted
	// here — only runTests-relevant calls (pytest, pip install pytest/pytest-cov)
	// and their relative ordering.
	descriptors := []struct {
		name string
		file string
	}{
		{"setup.py", "setup.py"},
		{"pyproject.toml", "pyproject.toml"},
	}

	for _, d := range descriptors {
		descriptor := d // capture loop variable

		t.Run("runTests=false - no pytest calls/"+descriptor.name, func(t *testing.T) {
			cfg := pythonBuildOptions{
				VirtualEnvironmentName: "dummy",
				RunTests:               false,
			}
			utils := newPythonBuildTestsUtils()
			utils.AddFile(descriptor.file, []byte(minimalSetupPyFileContent))
			utils.AddDir("dummy")
			telemetryData := telemetry.CustomData{}

			err := runPythonBuild(&cfg, &telemetryData, utils, &cpe)
			assert.NoError(t, err)
			for _, call := range utils.ExecMockRunner.Calls {
				assert.NotEqual(t, filepath.Join("dummy", "bin", "pytest"), call.Exec)
				assert.NotContains(t, call.Params, "pytest")
				assert.NotContains(t, call.Params, "pytest-cov")
			}
		})

		t.Run("runTests=true - happy path: install test deps then run pytest/"+descriptor.name, func(t *testing.T) {
			cfg := pythonBuildOptions{
				VirtualEnvironmentName: "dummy",
				RunTests:               true,
			}
			utils := newPythonBuildTestsUtils()
			utils.AddFile(descriptor.file, []byte(minimalSetupPyFileContent))
			utils.AddDir("dummy")
			telemetryData := telemetry.CustomData{}

			err := runPythonBuild(&cfg, &telemetryData, utils, &cpe)
			assert.NoError(t, err)

			pipExec := filepath.Join("dummy", "bin", "pip")
			pytestExec := filepath.Join("dummy", "bin", "pytest")
			installPytestIdx, installPytestCovIdx, pytestIdx := -1, -1, -1
			for i, call := range utils.ExecMockRunner.Calls {
				switch {
				case call.Exec == pipExec && slices.Contains(call.Params, "pytest") && !slices.Contains(call.Params, "pytest-cov") && installPytestIdx == -1:
					installPytestIdx = i
				case call.Exec == pipExec && slices.Contains(call.Params, "pytest-cov") && installPytestCovIdx == -1:
					installPytestCovIdx = i
				case call.Exec == pytestExec && pytestIdx == -1:
					pytestIdx = i
				}
			}
			assert.GreaterOrEqual(t, installPytestIdx, 0, "pip install pytest not found")
			assert.GreaterOrEqual(t, installPytestCovIdx, 0, "pip install pytest-cov not found")
			assert.GreaterOrEqual(t, pytestIdx, 0, "pytest execution not found")
			assert.Less(t, installPytestIdx, pytestIdx, "pip install pytest must occur before pytest")
			assert.Less(t, installPytestCovIdx, pytestIdx, "pip install pytest-cov must occur before pytest")

			pytestCall := utils.ExecMockRunner.Calls[pytestIdx]
			assert.Equal(t, pytestExec, pytestCall.Exec)
			assert.Equal(t, []string{
				"--junitxml=" + python.JUnitReportFile,
				"--cov",
				"--cov-report=xml:" + python.CoverageReportFile,
			}, pytestCall.Params)
		})

		t.Run("runTests=true - testOptions appended after report flags/"+descriptor.name, func(t *testing.T) {
			cfg := pythonBuildOptions{
				VirtualEnvironmentName: "dummy",
				RunTests:               true,
				TestOptions:            []string{"-v", "--tb=short"},
			}
			utils := newPythonBuildTestsUtils()
			utils.AddFile(descriptor.file, []byte(minimalSetupPyFileContent))
			utils.AddDir("dummy")
			telemetryData := telemetry.CustomData{}

			err := runPythonBuild(&cfg, &telemetryData, utils, &cpe)
			assert.NoError(t, err)

			var pytestCall *mock.ExecCall
			for i := range utils.ExecMockRunner.Calls {
				if utils.ExecMockRunner.Calls[i].Exec == filepath.Join("dummy", "bin", "pytest") {
					pytestCall = &utils.ExecMockRunner.Calls[i]
					break
				}
			}
			assert.NotNil(t, pytestCall, "pytest call not found")
			assert.Equal(t, []string{
				"--junitxml=" + python.JUnitReportFile,
				"--cov",
				"--cov-report=xml:" + python.CoverageReportFile,
				"-v",
				"--tb=short",
			}, pytestCall.Params)
		})

		t.Run("runTests=true - pytest failure sets ErrorTest category/"+descriptor.name, func(t *testing.T) {
			log.SetErrorCategory(log.ErrorUndefined)
			defer log.SetErrorCategory(log.ErrorUndefined)
			cfg := pythonBuildOptions{
				VirtualEnvironmentName: "dummy",
				RunTests:               true,
			}
			utils := newPythonBuildTestsUtils()
			utils.AddFile(descriptor.file, []byte(minimalSetupPyFileContent))
			utils.AddDir("dummy")
			utils.ExecMockRunner.ShouldFailOnCommand = map[string]error{
				filepath.Join("dummy", "bin", "pytest"): fmt.Errorf("exit status 1"),
			}
			telemetryData := telemetry.CustomData{}

			err := runPythonBuild(&cfg, &telemetryData, utils, &cpe)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "python tests")
			assert.Equal(t, log.ErrorTest, log.GetErrorCategory())
		})

		t.Run("runTests=true - test dep install failure sets ErrorBuild category/"+descriptor.name, func(t *testing.T) {
			log.SetErrorCategory(log.ErrorUndefined)
			defer log.SetErrorCategory(log.ErrorUndefined)
			cfg := pythonBuildOptions{
				VirtualEnvironmentName: "dummy",
				RunTests:               true,
			}
			utils := newPythonBuildTestsUtils()
			utils.AddFile(descriptor.file, []byte(minimalSetupPyFileContent))
			utils.AddDir("dummy")
			utils.ExecMockRunner.ShouldFailOnCommand = map[string]error{
				filepath.Join("dummy", "bin", "pip") + " " + strings.Join(append(python.PipInstallFlags, "pytest"), " "): fmt.Errorf("pip install failed"),
			}
			telemetryData := telemetry.CustomData{}

			err := runPythonBuild(&cfg, &telemetryData, utils, &cpe)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "install test dependencies")
			assert.Equal(t, log.ErrorBuild, log.GetErrorCategory())
		})

		t.Run("runTests=true, createBOM=true - pytest runs before BOM/"+descriptor.name, func(t *testing.T) {
			cfg := pythonBuildOptions{
				VirtualEnvironmentName: "dummy",
				RunTests:               true,
				CreateBOM:              true,
			}
			utils := newPythonBuildTestsUtils()
			utils.AddFile(descriptor.file, []byte(minimalSetupPyFileContent))
			utils.AddDir("dummy")
			telemetryData := telemetry.CustomData{}

			err := runPythonBuild(&cfg, &telemetryData, utils, &cpe)
			assert.NoError(t, err)

			pytestIdx := -1
			cyclonedxIdx := -1
			for i, call := range utils.ExecMockRunner.Calls {
				if call.Exec == filepath.Join("dummy", "bin", "pytest") {
					pytestIdx = i
				}
				if call.Exec == filepath.Join("dummy", "bin", "cyclonedx-py") {
					cyclonedxIdx = i
				}
			}
			assert.GreaterOrEqual(t, pytestIdx, 0, "pytest not found in calls")
			assert.GreaterOrEqual(t, cyclonedxIdx, 0, "cyclonedx not found in calls")
			assert.Less(t, pytestIdx, cyclonedxIdx, "pytest must run before BOM creation")
		})
	}

	// Full-pipeline subtests (descriptor-agnostic: only call ordering matters)
	t.Run("runTests=true, createBOM=true, publish=true - full pipeline ordering", func(t *testing.T) {
		cfg := pythonBuildOptions{
			VirtualEnvironmentName:   "dummy",
			RunTests:                 true,
			CreateBOM:                true,
			Publish:                  true,
			TargetRepositoryURL:      "https://my.target.repository.local",
			TargetRepositoryUser:     "user",
			TargetRepositoryPassword: "password",
		}
		utils := newPythonBuildTestsUtils()
		utils.AddFile("setup.py", []byte(minimalSetupPyFileContent))
		utils.AddDir("dummy")
		telemetryData := telemetry.CustomData{}

		err := runPythonBuild(&cfg, &telemetryData, utils, &cpe)
		assert.NoError(t, err)

		pytestIdx, cyclonedxIdx, twineIdx := -1, -1, -1
		for i, call := range utils.ExecMockRunner.Calls {
			switch call.Exec {
			case filepath.Join("dummy", "bin", "pytest"):
				pytestIdx = i
			case filepath.Join("dummy", "bin", "cyclonedx-py"):
				cyclonedxIdx = i
			case filepath.Join("dummy", "bin", "twine"):
				twineIdx = i
			}
		}
		assert.GreaterOrEqual(t, pytestIdx, 0, "pytest not found in calls")
		assert.GreaterOrEqual(t, cyclonedxIdx, 0, "cyclonedx-py not found in calls")
		assert.GreaterOrEqual(t, twineIdx, 0, "twine not found in calls")
		assert.Less(t, pytestIdx, cyclonedxIdx, "pytest must run before BOM creation")
		assert.Less(t, pytestIdx, twineIdx, "pytest must run before publish")
	})

	t.Run("runTests=true, publish=true - test failure short-circuits publish", func(t *testing.T) {
		cfg := pythonBuildOptions{
			VirtualEnvironmentName:   "dummy",
			RunTests:                 true,
			Publish:                  true,
			TargetRepositoryURL:      "https://my.target.repository.local",
			TargetRepositoryUser:     "user",
			TargetRepositoryPassword: "password",
		}
		utils := newPythonBuildTestsUtils()
		utils.AddFile("setup.py", []byte(minimalSetupPyFileContent))
		utils.AddDir("dummy")
		utils.ExecMockRunner.ShouldFailOnCommand = map[string]error{
			filepath.Join("dummy", "bin", "pytest"): fmt.Errorf("exit status 1"),
		}
		telemetryData := telemetry.CustomData{}

		err := runPythonBuild(&cfg, &telemetryData, utils, &cpe)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "python tests")
		assert.NotContains(t, err.Error(), "password")
		for _, call := range utils.ExecMockRunner.Calls {
			assert.NotEqual(t, filepath.Join("dummy", "bin", "twine"), call.Exec,
				"twine must not be called when tests fail")
		}
	})
}
