//go:build unit
// +build unit

package cmd

import (
	"testing"

	"errors"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestRunUIVeri5(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		opts := &uiVeri5ExecuteTestsOptions{
			InstallCommand: "npm install ui5/uiveri5",
			RunCommand:     "uiveri5",
			RunOptions:     []string{"conf.js"},
			TestServerURL:  "http://path/to/deployment",
		}

		e := mock.ExecMockRunner{}
		runUIVeri5(opts, &e)

		assert.Equal(t, e.Env[0], "NPM_CONFIG_PREFIX=~/.npm-global", "NPM_CONFIG_PREFIX not set as expected")
		assert.Contains(t, e.Env[1], "PATH", "PATH not in env list")
		assert.Equal(t, e.Env[2], "TARGET_SERVER_URL=http://path/to/deployment", "TARGET_SERVER_URL not set as expected")

		assert.Equal(t, e.Calls[0], mock.ExecCall{Exec: "npm", Params: []string{"install", "ui5/uiveri5"}}, "install command/params incorrect")

		assert.Equal(t, e.Calls[1], mock.ExecCall{Exec: "uiveri5", Params: []string{"conf.js"}}, "run command/params incorrect")

	})

	t.Run("error case install command", func(t *testing.T) {
		wantError := "failed to execute install command: fail install test: error case"

		opts := &uiVeri5ExecuteTestsOptions{InstallCommand: "fail install test", RunCommand: "uiveri5"}

		e := mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"fail install test": errors.New("error case")}}
		err := runUIVeri5(opts, &e)
		assert.EqualErrorf(t, err, wantError, "expected comman to exit with error")
	})

	t.Run("error case run command", func(t *testing.T) {
		wantError := "failed to execute run command: fail uiveri5 testParam: error case"

		opts := &uiVeri5ExecuteTestsOptions{InstallCommand: "npm install ui5/uiveri5", RunCommand: "fail uiveri5", RunOptions: []string{"testParam"}}

		e := mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"fail uiveri5": errors.New("error case")}}
		err := runUIVeri5(opts, &e)
		assert.EqualErrorf(t, err, wantError, "expected comman to exit with error")
	})
}
