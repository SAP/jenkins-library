package cmd

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func TestRunUIVeri5(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		opts := &uiVeri5ExecuteTestsOptions{
			InstallCommand:  "npm install ui5/uiveri5",
			RunCommand:      "uiveri5",
			RunOptions:      []string{"conf.js"},
			TestServerURL:   "http://path/to/deployment",
			NpmConfigPrefix: "/path/to/node",
		}

		e := mock.ExecMockRunner{}
		runUIVeri5(opts, &e)

		fmt.Println(e.Dir)
		fmt.Println(e.Calls)
		fmt.Println(e.Env)

		assert.Equal(t, e.Env[0], "NPM_CONFIG_PREFIX=/path/to/node", "NPM_CONFIG_PREFIX not set as expected")
		assert.Equal(t, e.Env[1], "TARGET_SERVER_URL=http://path/to/deployment", "TARGET_SERVER_URL not set as expected")

		assert.Equal(t, e.Calls[0], mock.ExecCall{Exec: "npm", Params: []string{"install", "ui5/uiveri5"}}, "install command/params incorrect")

		assert.Equal(t, e.Calls[1], mock.ExecCall{Exec: "uiveri5", Params: []string{"conf.js"}}, "run command/params incorrect")

	})

	t.Run("error case install command", func(t *testing.T) {
		var hasFailed bool
		log.Entry().Logger.ExitFunc = func(int) { hasFailed = true }

		opts := &uiVeri5ExecuteTestsOptions{InstallCommand: "fail install test", RunCommand: "uiveri5"}

		e := mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"fail install test": errors.New("error case")}}
		runUIVeri5(opts, &e)
		assert.True(t, hasFailed, "expected command to exit with fatal")
	})

	t.Run("error case run command", func(t *testing.T) {
		var hasFailed bool
		log.Entry().Logger.ExitFunc = func(int) { hasFailed = true }

		opts := &uiVeri5ExecuteTestsOptions{InstallCommand: "npm install ui5/uiveri5", RunCommand: "fail uiveri5"}

		e := mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"fail uiveri5": errors.New("error case")}}
		runUIVeri5(opts, &e)
		assert.True(t, hasFailed, "expected command to exit with fatal")
	})
}
