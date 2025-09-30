package btp

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	m := &BtpExecutorMock{
		StdoutReturn: map[string]string{
			"btp login": "Login successful",
		},
		ShouldFailOnCommand: map[string]error{
			"btp logout": errors.New("Logout failed"),
		},
	}

	m.Stdout(new(bytes.Buffer))

	// Test successful command execution
	err := m.Run("btp login")
	assert.NoError(t, err)
	assert.Contains(t, m.GetStdoutValue(), "Login successful")

	// Test failing command execution
	err = m.Run("btp logout")
	assert.Error(t, err)
	assert.Equal(t, "Logout failed", err.Error())
}

func TestRunSync_Success(t *testing.T) {
	m := &BtpExecutorMock{
		StdoutReturn: map[string]string{
			"btp check": `dummy
							dummy

							OK`,
		},
	}

	m.Stdout(new(bytes.Buffer))

	// Test successful polling execution
	err := m.RunSync(RunSyncOptions{
		CmdScript:      "btp deploy",
		TimeoutSeconds: 1,
		PollInterval:   30,
		CheckFunc: func() bool {
			return true // Simulate a successful check
		},
	})
	assert.NoError(t, err)
}

func TestRunSync_Erro_On_Check(t *testing.T) {
	m := &BtpExecutorMock{
		ShouldFailOnCommand: map[string]error{
			"btp check": errors.New("Bad Request"),
		},
	}

	m.Stdout(new(bytes.Buffer))

	timeoutMin := 1
	err := m.RunSync(RunSyncOptions{
		CmdScript:      "btp deploy",
		TimeoutSeconds: timeoutMin,
		PollInterval:   20,
		CheckFunc: func() bool {
			return false
		},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Command did not complete within the timeout period")
}
