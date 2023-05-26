//go:build unit
// +build unit

package cmd

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type isChangeInDevelopmentMockUtils struct {
	*mock.ExecMockRunner
}

func newIsChangeInDevelopmentTestsUtils() isChangeInDevelopmentMockUtils {
	utils := isChangeInDevelopmentMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
	}
	return utils
}

func TestRunIsChangeInDevelopment(t *testing.T) {

	t.Parallel()

	config := isChangeInDevelopmentOptions{
		Endpoint:                       "https://example.org/cm",
		Username:                       "me",
		Password:                       "****",
		ChangeDocumentID:               "12345678",
		CmClientOpts:                   []string{"-Dabc=123", "-Ddef=456"},
		FailIfStatusIsNotInDevelopment: true, // this is the default
	}

	expectedShellCall := mock.ExecCall{
		Exec: "cmclient",
		Params: []string{
			"--endpoint", "https://example.org/cm",
			"--user", "me",
			"--password", "****",
			"is-change-in-development",
			"--change-id", "12345678",
			"--return-code",
		},
	}

	t.Run("change found and in status IN_DEVELOPMENT", func(t *testing.T) {

		cmd := newIsChangeInDevelopmentTestsUtils()
		cmd.ExitCode = 0 // this exit code represents a change in status IN_DEVELOPMENT
		cpe := &isChangeInDevelopmentCommonPipelineEnvironment{}

		err := runIsChangeInDevelopment(&config, nil, cmd, cpe)

		assert.Equal(t, cpe.custom.isChangeInDevelopment, true)
		if assert.NoError(t, err) {
			assert.Equal(t, []string{"CMCLIENT_OPTS=-Dabc=123 -Ddef=456"}, cmd.Env)
			assert.Equal(t, []mock.ExecCall{expectedShellCall}, cmd.Calls)
		}
	})

	t.Run("change found and not in status IN_DEVELOPMENT", func(t *testing.T) {

		cmd := newIsChangeInDevelopmentTestsUtils()
		cmd.ExitCode = 3 // this exit code represents a change which is not in status IN_DEVELOPMENT
		cpe := &isChangeInDevelopmentCommonPipelineEnvironment{}

		err := runIsChangeInDevelopment(&config, nil, cmd, cpe)

		assert.Equal(t, cpe.custom.isChangeInDevelopment, false)
		if assert.EqualError(t, err, "change '12345678' is not in status 'in development'") {
			assert.Equal(t, []mock.ExecCall{expectedShellCall}, cmd.Calls)
		}
	})

	t.Run("change found and not in status IN_DEVELOPMENT, but we don't fail", func(t *testing.T) {

		cmd := newIsChangeInDevelopmentTestsUtils()
		cmd.ExitCode = 3 // this exit code represents a change which is not in status IN_DEVELOPMENT

		myConfig := config
		myConfig.FailIfStatusIsNotInDevelopment = false // needs to be explicitly configured
		cpe := &isChangeInDevelopmentCommonPipelineEnvironment{}

		err := runIsChangeInDevelopment(&myConfig, nil, cmd, cpe)

		assert.Equal(t, cpe.custom.isChangeInDevelopment, false)
		if assert.NoError(t, err) {
			assert.Equal(t, []mock.ExecCall{expectedShellCall}, cmd.Calls)
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {

		cmd := newIsChangeInDevelopmentTestsUtils()
		cmd.ExitCode = 2 // this exit code represents invalid credentials
		cpe := &isChangeInDevelopmentCommonPipelineEnvironment{}

		err := runIsChangeInDevelopment(&config, nil, cmd, cpe)

		if assert.EqualError(t, err, "cannot retrieve change status: invalid credentials") {
			assert.Equal(t, []mock.ExecCall{expectedShellCall}, cmd.Calls)
		}
	})

	t.Run("generic failure reported via exit code", func(t *testing.T) {

		cmd := newIsChangeInDevelopmentTestsUtils()
		cmd.ExitCode = 1 // this exit code indicates something went wrong
		cpe := &isChangeInDevelopmentCommonPipelineEnvironment{}

		err := runIsChangeInDevelopment(&config, nil, cmd, cpe)

		if assert.EqualError(t, err, "cannot retrieve change status: check log for details") {
			assert.Equal(t, []mock.ExecCall{expectedShellCall}, cmd.Calls)
		}
	})

	t.Run("generic failure reported via error", func(t *testing.T) {

		cmd := newIsChangeInDevelopmentTestsUtils()
		cmd.ExitCode = 1 // this exit code indicates something went wrong
		cmd.ShouldFailOnCommand = map[string]error{"cm.*": fmt.Errorf("%v", "Something went wrong")}
		cpe := &isChangeInDevelopmentCommonPipelineEnvironment{}

		err := runIsChangeInDevelopment(&config, nil, cmd, cpe)

		if assert.EqualError(t, err, "cannot retrieve change status: Something went wrong") {
			assert.Equal(t, []mock.ExecCall{expectedShellCall}, cmd.Calls)
		}
	})
}
