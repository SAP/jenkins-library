//go:build unit
// +build unit

package cmd

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"

	"github.com/SAP/jenkins-library/pkg/mock"
)

func checkStepActiveOpenFileMock(name string, tokens map[string]string) (io.ReadCloser, error) {
	var fileContent string
	switch name {
	case ".pipeline/defaults.yaml":
		fileContent = `
general:
stages:
steps:`
	case "stage-config.yml":
		fileContent = `
spec:
  stages:
    - name: testStage
      displayName: testStage
      steps:
        - name: testStep
          conditions:
            - configKey: testConfig`
	case ".pipeline/config.yml":
		fileContent = `
steps: 
  testStep: 
    testConfig: 'testValue'`
	default:
		fileContent = ""
	}
	return io.NopCloser(strings.NewReader(fileContent)), nil
}

func checkStepActiveFileExistsMock(filename string) (bool, error) {
	switch filename {
	case ".pipeline/config.yml":
		return true, nil
	default:
		return false, nil
	}
}

func TestCheckStepActiveCommand(t *testing.T) {
	cmd := CheckStepActiveCommand()

	gotReq := []string{}
	gotOpt := []string{}

	cmd.Flags().VisitAll(func(pflag *flag.Flag) {
		annotations, found := pflag.Annotations[cobra.BashCompOneRequiredFlag]
		if found && annotations[0] == "true" {
			gotReq = append(gotReq, pflag.Name)
		} else {
			gotOpt = append(gotOpt, pflag.Name)
		}
	})

	t.Run("Required flags", func(t *testing.T) {
		exp := []string{"step"}
		assert.Equal(t, exp, gotReq, "required flags incorrect")
	})

	t.Run("Optional flags", func(t *testing.T) {
		exp := []string{"stage", "stageConfig", "stageOutputFile", "stagesWithExtensions", "stepOutputFile", "useV1"}
		assert.Equal(t, exp, gotOpt, "optional flags incorrect")
	})

	t.Run("Run", func(t *testing.T) {
		t.Run("Success case - set stage and stageName parameters", func(t *testing.T) {
			checkStepActiveOptions.openFile = checkStepActiveOpenFileMock
			checkStepActiveOptions.fileExists = checkStepActiveFileExistsMock
			checkStepActiveOptions.stageName = "testStage"
			checkStepActiveOptions.stepName = "testStep"
			checkStepActiveOptions.stageConfigFile = "stage-config.yml"
			GeneralConfig.CustomConfig = ".pipeline/config.yml"
			GeneralConfig.DefaultConfig = []string{".pipeline/defaults.yaml"}
			GeneralConfig.StageName = "testStage1"
			cmd.Run(cmd, []string{})
		})
		t.Run("Success case - set only stage parameter", func(t *testing.T) {
			checkStepActiveOptions.openFile = checkStepActiveOpenFileMock
			checkStepActiveOptions.fileExists = checkStepActiveFileExistsMock
			checkStepActiveOptions.stageName = "testStage"
			checkStepActiveOptions.stepName = "testStep"
			checkStepActiveOptions.stageConfigFile = "stage-config.yml"
			GeneralConfig.CustomConfig = ".pipeline/config.yml"
			GeneralConfig.DefaultConfig = []string{".pipeline/defaults.yaml"}
			cmd.Run(cmd, []string{})
		})
		t.Run("Success case - set only stageName parameter", func(t *testing.T) {
			checkStepActiveOptions.openFile = checkStepActiveOpenFileMock
			checkStepActiveOptions.fileExists = checkStepActiveFileExistsMock
			checkStepActiveOptions.stepName = "testStep"
			checkStepActiveOptions.stageConfigFile = "stage-config.yml"
			GeneralConfig.CustomConfig = ".pipeline/config.yml"
			GeneralConfig.DefaultConfig = []string{".pipeline/defaults.yaml"}
			GeneralConfig.StageName = "testStage"
			cmd.Run(cmd, []string{})
		})
	})
}
func TestCheckIfStepActiveWithStagesWithExtensions(t *testing.T) {
	housekeepingOpenFileMock := func(name string, tokens map[string]string) (io.ReadCloser, error) {
		var fileContent string
		switch name {
		case ".pipeline/defaults.yaml":
			fileContent = `
general:
stages:
steps:`
		case "stage-config.yml":
			fileContent = `
spec:
  stages:
    - name: testStage
      displayName: testStage
      steps:
        - name: housekeepingStep
          housekeeping: true`
		default:
			fileContent = ""
		}
		return io.NopCloser(strings.NewReader(fileContent)), nil
	}

	setup := func(t *testing.T, stagesWithExtensions string) *mock.FilesMock {
		optionsBackup := checkStepActiveOptions
		t.Cleanup(func() { checkStepActiveOptions = optionsBackup })
		checkStepActiveOptions.openFile = housekeepingOpenFileMock
		checkStepActiveOptions.fileExists = func(string) (bool, error) { return false, nil }
		checkStepActiveOptions.stageName = "testStage"
		checkStepActiveOptions.stepName = "housekeepingStep"
		checkStepActiveOptions.stageConfigFile = "stage-config.yml"
		checkStepActiveOptions.stageOutputFile = "stage_out.json"
		checkStepActiveOptions.stagesWithExtensions = stagesWithExtensions
		GeneralConfig.DefaultConfig = []string{".pipeline/defaults.yaml"}
		return &mock.FilesMock{}
	}

	t.Run("housekeeping-only stage is inactive without extensions", func(t *testing.T) {
		utils := setup(t, "")
		assert.NoError(t, checkIfStepActive(utils))
		content, err := utils.FileRead("stage_out.json")
		assert.NoError(t, err)
		assert.JSONEq(t, `{"testStage": false}`, string(content))
	})

	t.Run("housekeeping-only stage is active when announced via stagesWithExtensions", func(t *testing.T) {
		utils := setup(t, "otherStage, testStage")
		assert.NoError(t, checkIfStepActive(utils))
		content, err := utils.FileRead("stage_out.json")
		assert.NoError(t, err)
		assert.JSONEq(t, `{"testStage": true}`, string(content))
	})
}

func TestSplitAndTrim(t *testing.T) {
	assert.Equal(t, []string{"A", "B", "C"}, splitAndTrim("A, B,,C"))
	assert.Nil(t, splitAndTrim(""))
	assert.Nil(t, splitAndTrim(" , "))
}

func TestFailIfNoConfigFound(t *testing.T) {
	if os.Getenv("TEST_FAIL_IF_NO_CONFIG_FOUND") == "1" {
		cmd := CheckStepActiveCommand()

		gotReq := []string{}
		gotOpt := []string{}

		cmd.Flags().VisitAll(func(pflag *flag.Flag) {
			annotations, found := pflag.Annotations[cobra.BashCompOneRequiredFlag]
			if found && annotations[0] == "true" {
				gotReq = append(gotReq, pflag.Name)
			} else {
				gotOpt = append(gotOpt, pflag.Name)
			}
		})
		checkStepActiveOptions.openFile = checkStepActiveOpenFileMock
		checkStepActiveOptions.fileExists = checkStepActiveFileExistsMock
		checkStepActiveOptions.stageName = "testStage"
		checkStepActiveOptions.stepName = "testStep"
		checkStepActiveOptions.stageConfigFile = "stage-config.yml"
		GeneralConfig.CustomConfig = ".pipeline/unknown.yml"
		GeneralConfig.DefaultConfig = []string{".pipeline/defaults.yaml"}
		GeneralConfig.StageName = "testStage1"
		cmd.Run(cmd, []string{})
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestFailIfNoConfigFound")
	cmd.Env = append(os.Environ(), "TEST_FAIL_IF_NO_CONFIG_FOUND=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		t.Log(e.Error())
		t.Log("Stderr: ", string(e.Stderr))
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}
