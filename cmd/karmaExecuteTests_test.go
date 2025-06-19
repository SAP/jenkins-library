//go:build unit

package cmd

import (
	"errors"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestRunKarma(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		opts := karmaExecuteTestsOptions{Modules: []string{"./test"}, InstallCommand: "npm install test", RunCommand: "npm run test"}

		e := mock.ExecMockRunner{}
		runKarma(opts, &e)

		assert.Equal(t, e.Dir[0], "./test", "install command dir incorrect")
		assert.Equal(t, e.Calls[0], mock.ExecCall{Exec: "npm", Params: []string{"install", "test"}}, "install command/params incorrect")

		assert.Equal(t, e.Dir[1], "./test", "run command dir incorrect")
		assert.Equal(t, e.Calls[1], mock.ExecCall{Exec: "npm", Params: []string{"run", "test"}}, "run command/params incorrect")

	})

	t.Run("success case - verbose logging", func(t *testing.T) {
		GeneralConfig.Verbose = true
		defer func() { GeneralConfig.Verbose = false }()

		opts := karmaExecuteTestsOptions{Modules: []string{"./test"}, InstallCommand: "npm install test", RunCommand: "npm run test"}

		e := mock.ExecMockRunner{}
		runKarma(opts, &e)

		assert.Equal(t, "./test", e.Dir[1], "run command dir incorrect")
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "test", "--", "--log-level", "DEBUG"}}, e.Calls[1], "run command/params incorrect")
	})

	t.Run("error case install command", func(t *testing.T) {
		var hasFailed bool
		log.Entry().Logger.ExitFunc = func(int) { hasFailed = true }

		opts := karmaExecuteTestsOptions{Modules: []string{"./test"}, InstallCommand: "fail install test", RunCommand: "npm run test"}

		e := mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"fail install test": errors.New("error case")}}
		runKarma(opts, &e)
		assert.True(t, hasFailed, "expected command to exit with fatal")
	})

	t.Run("error case run command", func(t *testing.T) {
		var hasFailed bool
		log.Entry().Logger.ExitFunc = func(int) { hasFailed = true }

		opts := karmaExecuteTestsOptions{Modules: []string{"./test"}, InstallCommand: "npm install test", RunCommand: "npm run test"}

		e := mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"npm install test": errors.New("error case")}}
		runKarma(opts, &e)
		assert.True(t, hasFailed, "expected command to exit with fatal")
	})
}
