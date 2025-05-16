//go:build unit

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
		exp := []string{"stage", "stageConfig", "stageOutputFile", "stepOutputFile", "useV1"}
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
