//go:build unit

package solman

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestSolmanCreateTransportRequest(t *testing.T) {

	a := CreateAction{
		Connection: Connection{
			Endpoint: "https://example.org/solman",
			User:     "me",
			Password: "******",
		},
		ChangeDocumentID:    "123",
		DevelopmentSystemID: "XXX~EXT_SRV",
		CMOpts:              []string{"-Dprop1=abc", "-Dprop2=123"},
	}

	t.Run("straight forward", func(t *testing.T) {

		e := getExecMock()
		e.StdoutReturn = map[string]string{"^cmclient.*": "ABCK123456"}

		examinee := a
		transportRequestId, err := examinee.Perform(e)

		if assert.NoError(t, err) {
			assert.Equal(t, []mock.ExecCall{mock.ExecCall{
				Exec: "cmclient",
				Params: []string{
					"--endpoint", "https://example.org/solman",
					"--user", "me",
					"--password", "******",
					"create-transport",
					"-cID", "123",
					"-dID", "XXX~EXT_SRV",
				},
			}}, e.Calls)
			assert.Equal(t, "ABCK123456", transportRequestId)
			assert.Equal(t, []string{"CMCLIENT_OPTS=-Dprop1=abc -Dprop2=123"}, e.Env)
		}
	})

	t.Run("fail with error", func(t *testing.T) {

		e := getExecMock()
		e.ShouldFailOnCommand = map[string]error{"^cmclient.*": fmt.Errorf("creating transport request failed")}

		examinee := a
		_, err := examinee.Perform(e)

		assert.EqualError(t, err, "cannot create transport request: creating transport request failed")
	})

	t.Run("fail via return code", func(t *testing.T) {

		e := getExecMock()
		e.ExitCode = 42

		examinee := a
		_, err := examinee.Perform(e)

		assert.EqualError(t, err, "cannot create transport request: create transport request command returned with exit code '42'")
	})

	t.Run("input missing", func(t *testing.T) {

		e := getExecMock()

		examinee := a
		examinee.Connection = Connection{}
		examinee.ChangeDocumentID = ""

		_, err := examinee.Perform(e)

		if assert.Error(t, err) {
			// I don't want to rely on the order of the parameters
			assert.Contains(t, err.Error(), "cannot create transport request: the following parameters are not available")
			assert.Contains(t, err.Error(), "Connection.Endpoint")
			assert.Contains(t, err.Error(), "Connection.User")
			assert.Contains(t, err.Error(), "Connection.Password")
			assert.Contains(t, err.Error(), "ChangeDocumentID")
			assert.Empty(t, e.Calls)
		}
	})
}

func getExecMock() *mock.ExecMockRunner {
	var out bytes.Buffer
	e := &mock.ExecMockRunner{}
	e.Stdout(&out)
	return e
}
