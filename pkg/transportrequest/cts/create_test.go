package cts

import (
	"bytes"
	"errors"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateTransportRequest(t *testing.T) {

	defaultCreateAction := CreateAction{
		Connection: Connection{
			Endpoint: "https://example.org/cts",
			User:     "me",
			Password: "******",
		},
		TransportType:  "X",
		TargetSystemID: "XYZ",
		Description:    "Lorem ipsum",
		CMOpts:         []string{"-Dx=y", "-Dabc=123"},
	}

	t.Run("straight forward", func(t *testing.T) {
		examinee := defaultCreateAction
		cmd := getExecMock()
		cmd.StdoutReturn = map[string]string{"cmclient:*": "XXXK12345678"}
		transportRequestId, err := examinee.Perform(cmd)

		if assert.NoError(t, err) {
			t.Run("assert the call", func(t *testing.T) {
				assert.Equal(t, []mock.ExecCall{
					mock.ExecCall{
						Exec: "cmclient",
						Params: []string{
							"--endpoint", "https://example.org/cts",
							"--user", "me",
							"--password", "******",
							"-t", "CTS",
							"create-transport",
							"-tt", "X",
							"-ts", "XYZ",
							"-d", "Lorem ipsum",
						},
					},
				}, cmd.Calls)
			})
			t.Run("assert the transport request id", func(t *testing.T) {
				assert.Equal(t, "XXXK12345678", transportRequestId)
			})
			t.Run("assert the additional environment", func(t *testing.T) {
				assert.Equal(t, []string{"CMCLIENT_OPTS=-Dx=y -Dabc=123"}, cmd.Env)
			})
		}
	})

	t.Run("no transport request id provided via stdout by cm client", func(t *testing.T) {
		examinee := defaultCreateAction
		cmd := getExecMock()
		cmd.StdoutReturn = map[string]string{"cmclient:*": ""}
		_, err := examinee.Perform(cmd)

		assert.EqualError(t, err, "cannot create transport request at 'https://example.org/cts', transport type: 'X', target system: 'XYZ': no transport request id received")
	})

	t.Run("create transport request fails with rc not equal zero", func(t *testing.T) {
		examinee := defaultCreateAction
		cmd := getExecMock()
		cmd.ExitCode = 1
		_, err := examinee.Perform(cmd)

		assert.EqualError(t, err, "cannot create transport request at 'https://example.org/cts', transport type: 'X', target system: 'XYZ': Create transport request command returned with exit code '1'")
	})

	t.Run("create transport request fails with error", func(t *testing.T) {
		examinee := defaultCreateAction
		cmd := getExecMock()
		cmd.ShouldFailOnCommand = map[string]error{"cmclient:*": errors.New("We have a problem")}
		_, err := examinee.Perform(cmd)

		assert.EqualError(t, err, "cannot create transport request at 'https://example.org/cts', transport type: 'X', target system: 'XYZ': We have a problem")
	})

	t.Run("check missing parameters", func(t *testing.T) {
		examinee := defaultCreateAction
		examinee.Connection.Endpoint = ""
		examinee.TransportType = ""
		cmd := getExecMock()

		_, err := examinee.Perform(cmd)

		if assert.EqualError(t, err, "cannot create transport request at '', transport type: '', target system: 'XYZ': the following parameters are not available [Connection.Endpoint TransportType]") {
			t.Run("no calls", func(t *testing.T) {
				assert.Empty(t, cmd.Calls)
			})
		}
	})
}

func getExecMock() *mock.ExecMockRunner {
	var out bytes.Buffer
	e := &mock.ExecMockRunner{}
	e.Stdout(&out)
	return e
}
