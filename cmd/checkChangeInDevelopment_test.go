package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type checkChangeInDevelopmentMockUtils struct {
	*mock.ExecMockRunner
}

func newCheckChangeInDevelopmentTestsUtils() checkChangeInDevelopmentMockUtils {
	utils := checkChangeInDevelopmentMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
	}
	return utils
}

func TestRunCheckChangeInDevelopment(t *testing.T) {

	t.Parallel()

	config := checkChangeInDevelopmentOptions{
		Endpoint:                       "https://example.org/cm",
		Username:                       "me",
		Password:                       "****",
		ChangeDocumentID:               "12345678",
		ClientOpts:                     []string{"-Dabc=123", "-Ddef=456"},
		FailIfStatusIsNotInDevelopment: true, // this is the default
	}

	expectedShellCall := mock.ExecCall{
		Exec: "cmclient",
		Params: []string{
			"--endpoint", "https://example.org/cm",
			"--user", "me",
			"--password", "****",
			"--backend-type", "SOLMAN",
			"is-change-in-development",
			"--change-id", "12345678",
			"--return-code",
		},
	}

	t.Run("change found and in status IN_DEVELOPMENT", func(t *testing.T) {

		cmd := newCheckChangeInDevelopmentTestsUtils()
		cmd.ExitCode = 0 // this exit code represents a change in status IN_DEVELOPMENT

		err := runCheckChangeInDevelopment(&config, nil, cmd)

		if assert.NoError(t, err) {
			assert.Equal(t, []string{"CMCLIENT_OPTS=-Dabc=123 -Ddef=456"}, cmd.Env)
			assert.Equal(t, []mock.ExecCall{expectedShellCall}, cmd.Calls)
		}
	})

	t.Run("change found and not in status IN_DEVELOPMENT", func(t *testing.T) {

		cmd := newCheckChangeInDevelopmentTestsUtils()
		cmd.ExitCode = 3 // this exit code represents a change which is not in status IN_DEVELOPMENT

		err := runCheckChangeInDevelopment(&config, nil, cmd)

		if assert.EqualError(t, err, "Change '12345678' is not in status 'in development'") {
			assert.Equal(t, []mock.ExecCall{expectedShellCall}, cmd.Calls)
		}
	})

	t.Run("change found and not in status IN_DEVELOPMENT, but we don't fail", func(t *testing.T) {

		cmd := newCheckChangeInDevelopmentTestsUtils()
		cmd.ExitCode = 3 // this exit code represents a change which is not in status IN_DEVELOPMENT

		myConfig := config
		myConfig.FailIfStatusIsNotInDevelopment = false // needs to be explicitly configured

		err := runCheckChangeInDevelopment(&myConfig, nil, cmd)

		if assert.NoError(t, err) {
			assert.Equal(t, []mock.ExecCall{expectedShellCall}, cmd.Calls)
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {

		cmd := newCheckChangeInDevelopmentTestsUtils()
		cmd.ExitCode = 2 // this exit code represents invalid credentials

		err := runCheckChangeInDevelopment(&config, nil, cmd)

		if assert.EqualError(t, err, "Cannot retrieve change status: Invalid credentials") {
			assert.Equal(t, []mock.ExecCall{expectedShellCall}, cmd.Calls)
		}
	})

	t.Run("generic failure reported via exit code", func(t *testing.T) {

		cmd := newCheckChangeInDevelopmentTestsUtils()
		cmd.ExitCode = 1 // this exit code indicates something went wrong

		err := runCheckChangeInDevelopment(&config, nil, cmd)

		if assert.EqualError(t, err, "Cannot retrieve change status: Check log for details") {
			assert.Equal(t, []mock.ExecCall{expectedShellCall}, cmd.Calls)
		}
	})

	t.Run("generic failure reported via error", func(t *testing.T) {

		cmd := newCheckChangeInDevelopmentTestsUtils()
		cmd.ShouldFailOnCommand = map[string]error{"cm.*": fmt.Errorf("%v", "Something went wrong")}

		err := runCheckChangeInDevelopment(&config, nil, cmd)

		if assert.EqualError(t, err, "Cannot retrieve change status: Something went wrong") {
			assert.Equal(t, []mock.ExecCall{expectedShellCall}, cmd.Calls)
		}
	})
}
