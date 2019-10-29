package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunKarma(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		opts := karmaExecuteTestsOptions{ModulePath: "./test", InstallCommand: "npm install test", RunCommand: "npm run test"}

		e := mockRunner{}
		err := runKarma(opts, &e)

		assert.NoError(t, err, "error occured but no error expected")

		assert.Equal(t, e.dir[0], "./test", "install command dir incorrect")
		assert.Equal(t, e.calls[0], execCall{exec: "npm", params: []string{"install", "test"}}, "install command/params incorrect")

		assert.Equal(t, e.dir[1], "./test", "run command dir incorrect")
		assert.Equal(t, e.calls[1], execCall{exec: "npm", params: []string{"run", "test"}}, "run command/params incorrect")

	})

	t.Run("error case install command", func(t *testing.T) {
		opts := karmaExecuteTestsOptions{ModulePath: "./test", InstallCommand: "fail install test", RunCommand: "npm run test"}

		e := mockRunner{}
		err := runKarma(opts, &e)
		assert.Error(t, err, "error expected but none occcured")
	})

	t.Run("error case run command", func(t *testing.T) {
		opts := karmaExecuteTestsOptions{ModulePath: "./test", InstallCommand: "npm install test", RunCommand: "fail run test"}

		e := mockRunner{}
		err := runKarma(opts, &e)
		assert.Error(t, err, "error expected but none occcured")
	})
}
