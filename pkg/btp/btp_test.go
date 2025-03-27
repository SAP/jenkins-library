package btp

import (
	"bytes"
	"errors"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	m := &mock.BtpExecuterMock{
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
	m := &mock.BtpExecuterMock{
		StdoutReturn: map[string]string{
			"btp check": `dummy
							dummy

							OK`,
		},
	}

	m.Stdout(new(bytes.Buffer))

	// Test successful polling execution
	err := m.RunSync("btp deploy", "btp check", 1, 30, false)
	assert.NoError(t, err)
}

func TestRunSync_Timeout(t *testing.T) {
	m := &mock.BtpExecuterMock{
		ShouldFailOnCommand: map[string]error{
			"btp check": errors.New("Bad Request"),
		},
	}

	m.Stdout(new(bytes.Buffer))

	timeoutMin := 1
	err := m.RunSync("btp deploy", "btp check", timeoutMin, 20, false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Command did not complete within the timeout period")
}
