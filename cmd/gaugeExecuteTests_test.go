package cmd

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type gaugeExecuteTestsMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func (utils gaugeExecuteTestsMockUtils) Getenv(key string) string {
	if key == "HOME" {
		return filepath.FromSlash("/home/node")
	}
	return ""
}

func TestRunGaugeExecuteTests(t *testing.T) {
	t.Parallel()

	allFineConfig := gaugeExecuteTestsOptions{
		InstallCommand: "npm install -g @getgauge/cli",
		LanguageRunner: "java",
		RunCommand:     "run",
		TestOptions:    "specs",
	}
	gaugeBinRegEx := "home.node..npm-global.bin.gauge"

	t.Run("success case", func(t *testing.T) {
		t.Parallel()

		mockUtils := gaugeExecuteTestsMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{},
			FilesMock:      &mock.FilesMock{},
		}

		err := runGaugeExecuteTests(&allFineConfig, nil, &mockUtils)

		assert.NoError(t, err)
		assert.Equal(t, mockUtils.ExecMockRunner.Calls[0].Exec, "npm")
		assert.Equal(t, mockUtils.ExecMockRunner.Calls[0].Params, []string{"install", "-g", "@getgauge/cli", "--prefix=~/.npm-global"})
		assert.Equal(t, mockUtils.ExecMockRunner.Calls[1].Exec, filepath.FromSlash("/home/node/.npm-global/bin/gauge"))
		assert.Equal(t, mockUtils.ExecMockRunner.Calls[1].Params, []string{"install", "java"})
		assert.Equal(t, mockUtils.ExecMockRunner.Calls[2].Exec, filepath.FromSlash("/home/node/.npm-global/bin/gauge"))
		assert.Equal(t, mockUtils.ExecMockRunner.Calls[2].Params, []string{"run", "specs"})
	})

	t.Run("fail on installation", func(t *testing.T) {
		t.Parallel()

		badInstallConfig := allFineConfig
		badInstallConfig.InstallCommand = "npm install -g @wronggauge/cli"

		mockUtils := gaugeExecuteTestsMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"npm install -g @wronggauge/cli": errors.New("cannot find module")}},
			FilesMock:      &mock.FilesMock{},
		}

		err := runGaugeExecuteTests(&badInstallConfig, nil, &mockUtils)
		assert.True(t, errors.Is(err, ErrorGaugeInstall))

		assert.Equal(t, len(mockUtils.ExecMockRunner.Calls), 1)
		assert.Equal(t, mockUtils.ExecMockRunner.Calls[0].Exec, "npm")
		assert.Equal(t, mockUtils.ExecMockRunner.Calls[0].Params, []string{"install", "-g", "@wronggauge/cli", "--prefix=~/.npm-global"})
	})

	t.Run("fail on installing language runner", func(t *testing.T) {
		t.Parallel()
		badInstallConfig := allFineConfig
		badInstallConfig.LanguageRunner = "wrong"

		mockUtils := gaugeExecuteTestsMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{gaugeBinRegEx + " install wrong": errors.New("error installing runner")}},
			FilesMock:      &mock.FilesMock{},
		}

		err := runGaugeExecuteTests(&badInstallConfig, nil, &mockUtils)
		assert.True(t, errors.Is(err, ErrorGaugeRunnerInstall))

		assert.Equal(t, len(mockUtils.ExecMockRunner.Calls), 2)

		assert.Equal(t, mockUtils.ExecMockRunner.Calls[0].Exec, "npm")
		assert.Equal(t, mockUtils.ExecMockRunner.Calls[0].Params, []string{"install", "-g", "@getgauge/cli", "--prefix=~/.npm-global"})
		assert.Equal(t, mockUtils.ExecMockRunner.Calls[1].Exec, filepath.FromSlash("/home/node/.npm-global/bin/gauge"))
		assert.Equal(t, mockUtils.ExecMockRunner.Calls[1].Params, []string{"install", "wrong"})
	})
	t.Run("fail on gauge run", func(t *testing.T) {
		t.Parallel()

		mockUtils := gaugeExecuteTestsMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{gaugeBinRegEx + " run specs": errors.New("error running gauge")}},
			FilesMock:      &mock.FilesMock{},
		}

		err := runGaugeExecuteTests(&allFineConfig, nil, &mockUtils)
		assert.True(t, errors.Is(err, ErrorGaugeRun))

		assert.Equal(t, len(mockUtils.ExecMockRunner.Calls), 3)

		assert.Equal(t, mockUtils.ExecMockRunner.Calls[0].Exec, "npm")
		assert.Equal(t, mockUtils.ExecMockRunner.Calls[0].Params, []string{"install", "-g", "@getgauge/cli", "--prefix=~/.npm-global"})
		assert.Equal(t, mockUtils.ExecMockRunner.Calls[1].Exec, filepath.FromSlash("/home/node/.npm-global/bin/gauge"))
		assert.Equal(t, mockUtils.ExecMockRunner.Calls[1].Params, []string{"install", "java"})
		assert.Equal(t, mockUtils.ExecMockRunner.Calls[2].Exec, filepath.FromSlash("/home/node/.npm-global/bin/gauge"))
		assert.Equal(t, mockUtils.ExecMockRunner.Calls[2].Params, []string{"run", "specs"})
	})
}
