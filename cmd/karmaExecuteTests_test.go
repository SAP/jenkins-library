package cmd

import (
	"errors"
	"testing"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestRunKarma(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		opts := karmaExecuteTestsOptions{ModulePath: "./test", InstallCommand: "npm install test", RunCommand: "npm run test"}

		e := execMockRunner{}
		runKarma(opts, &e)

		assert.Equal(t, e.dir[0], "./test", "install command dir incorrect")
		assert.Equal(t, e.calls[0], execCall{exec: "npm", params: []string{"install", "test"}}, "install command/params incorrect")

		assert.Equal(t, e.dir[1], "./test", "run command dir incorrect")
		assert.Equal(t, e.calls[1], execCall{exec: "npm", params: []string{"run", "test"}}, "run command/params incorrect")

	})

	t.Run("error case install command", func(t *testing.T) {
		var hasFailed bool
		log.Entry().Logger.ExitFunc = func(int) { hasFailed = true }

		opts := karmaExecuteTestsOptions{ModulePath: "./test", InstallCommand: "fail install test", RunCommand: "npm run test"}

		e := execMockRunner{shouldFailWith: errors.New("error case")}
		runKarma(opts, &e)
		assert.True(t, hasFailed, "expected command to exit with fatal")
	})

	t.Run("error case run command", func(t *testing.T) {
		var hasFailed bool
		log.Entry().Logger.ExitFunc = func(int) { hasFailed = true }

		opts := karmaExecuteTestsOptions{ModulePath: "./test", InstallCommand: "npm install test", RunCommand: "npm run test"}

		e := execMockRunner{shouldFailWith: errors.New("error case")}
		runKarma(opts, &e)
		assert.True(t, hasFailed, "expected command to exit with fatal")
	})
}
