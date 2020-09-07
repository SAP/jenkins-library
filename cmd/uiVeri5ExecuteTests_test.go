package cmd

import (
	"errors"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestRunUIVeri5(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		opts := &uiVeri5ExecuteTestsOptions{ModulePath: "./test", InstallCommand: "npm install ui5/uiveri5", RunCommand: "uiveri5", ConfPath: "conf.js"}

		e := mock.ExecMockRunner{}
		runUIVeri5(opts, &e)

		assert.Equal(t, e.Dir[0], "./test", "install command dir incorrect")
		assert.Equal(t, e.Calls[0], mock.ExecCall{Exec: "npm", Params: []string{"install", "ui5/uiveri5"}}, "install command/params incorrect")

		assert.Equal(t, e.Dir[1], "./test", "run command dir incorrect")
		assert.Equal(t, e.Calls[1], mock.ExecCall{Exec: "uiveri5", Params: []string{"conf.js"}}, "run command/params incorrect")

	})

	t.Run("error case install command", func(t *testing.T) {
		var hasFailed bool
		log.Entry().Logger.ExitFunc = func(int) { hasFailed = true }

		opts := &uiVeri5ExecuteTestsOptions{ModulePath: "./test", InstallCommand: "fail install test", RunCommand: "uiveri5", ConfPath: "conf.js"}

		e := mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"fail install test": errors.New("error case")}}
		runUIVeri5(opts, &e)
		assert.True(t, hasFailed, "expected command to exit with fatal")
	})

	t.Run("error case run command", func(t *testing.T) {
		var hasFailed bool
		log.Entry().Logger.ExitFunc = func(int) { hasFailed = true }

		opts := &uiVeri5ExecuteTestsOptions{ModulePath: "./test", InstallCommand: "npm install ui5/uiveri5", RunCommand: "fail uiveri5", ConfPath: "conf.js"}

		e := mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"fail uiveri5": errors.New("error case")}}
		runUIVeri5(opts, &e)
		assert.True(t, hasFailed, "expected command to exit with fatal")
	})
}
