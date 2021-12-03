package cmd

import (
	"io"
	"io/ioutil"
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
stages:
  testStage:
    stepConditions:
      testStep:
        config: testConfig`
	case ".pipeline/config.yml":
		fileContent = `
steps: 
  testStep: 
    testConfig: 'testValue'`
	default:
		fileContent = ""
	}
	return ioutil.NopCloser(strings.NewReader(fileContent)), nil
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
		exp := []string{"stage", "stageConfig"}
		assert.Equal(t, exp, gotOpt, "optional flags incorrect")
	})

	t.Run("Run", func(t *testing.T) {
		t.Run("Success case", func(t *testing.T) {
			checkStepActiveOptions.openFile = checkStepActiveOpenFileMock
			checkStepActiveOptions.stageName = "testStage1"
			checkStepActiveOptions.stepName = "testStep"
			checkStepActiveOptions.stageConfigFile = "stage-config.yml"
			GeneralConfig.CustomConfig = ".pipeline/config.yml"
			GeneralConfig.DefaultConfig = []string{".pipeline/defaults.yaml"}
			GeneralConfig.StageName = "testStage"
			cmd.Run(cmd, []string{})
		})
	})
}
