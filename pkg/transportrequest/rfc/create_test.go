package rfc

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
			Instance: "001",
			Client:   "DEV",
			User:     "me",
			Password: "******",
		},
		TransportType:  "X",
		TargetSystemID: "XYZ",
		Description:    "Lorem ipsum",
	}

	t.Run("straight forward", func(t *testing.T) {
		examinee := defaultCreateAction
		cmd := getExecMock()
		cmd.StdoutReturn = map[string]string{"cts.*": "{\"REQUESTID\":\"XXXK12345678\"}"}
		transportRequestId, err := examinee.Perform(cmd)

		if assert.NoError(t, err) {
			t.Run("assert the call", func(t *testing.T) {
				assert.Equal(t, []mock.ExecCall{
					mock.ExecCall{
						Exec: "cts",
						Params: []string{
							"createTransportRequest",
						},
					},
				}, cmd.Calls)
			})
			t.Run("assert the transport request id", func(t *testing.T) {
				assert.Equal(t, "XXXK12345678", transportRequestId)
			})
			t.Run("assert the additional environment", func(t *testing.T) {

				expected := []string{
					"ABAP_DEVELOPMENT_SERVER=https://example.org/cts",
					"ABAP_DEVELOPMENT_USER=me",
					"ABAP_DEVELOPMENT_PASSWORD=******",
					"TRANSPORT_DESCRIPTION=Lorem ipsum",
					"ABAP_DEVELOPMENT_INSTANCE=001",
					"ABAP_DEVELOPMENT_CLIENT=DEV",
				}

				if assert.Len(t, expected, 6) {
					assert.Subset(t, expected, cmd.Env)
				}
			})
		}
	})

	t.Run("no transport request id provided via stdout by cm client", func(t *testing.T) {
		examinee := defaultCreateAction
		cmd := getExecMock()
		cmd.StdoutReturn = map[string]string{"cts.*": ""}
		_, err := examinee.Perform(cmd)

		assert.EqualError(t, err, "Cannot create transport request at 'https://example.org/cts': No transport request id received.")
	})

	t.Run("create transport request fails with rc not equal zero", func(t *testing.T) {
		examinee := defaultCreateAction
		cmd := getExecMock()
		cmd.ExitCode = 1
		_, err := examinee.Perform(cmd)

		assert.EqualError(t, err, "Cannot create transport request at 'https://example.org/cts': Create transport request command returned with exit code '1'")
	})

	t.Run("create transport request fails with error", func(t *testing.T) {
		examinee := defaultCreateAction
		cmd := getExecMock()
		cmd.ShouldFailOnCommand = map[string]error{"cts.*": errors.New("We have a problem")}
		_, err := examinee.Perform(cmd)

		assert.EqualError(t, err, "Cannot create transport request at 'https://example.org/cts': We have a problem")
	})

	t.Run("check missing parameters", func(t *testing.T) {
		examinee := defaultCreateAction
		examinee.Connection.Endpoint = ""
		examinee.TransportType = ""
		cmd := getExecMock()

		_, err := examinee.Perform(cmd)

		if assert.EqualError(t, err, "Cannot create transport request at '': the following parameters are not available [Connection.Endpoint TransportType]") {
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
