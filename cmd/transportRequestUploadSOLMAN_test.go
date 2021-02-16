package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/transportrequest/solman"
	"github.com/stretchr/testify/assert"
	"testing"
)

type transportRequestUploadSOLMANMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newTransportRequestUploadSOLMANTestsUtils() transportRequestUploadSOLMANMockUtils {
	utils := transportRequestUploadSOLMANMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

type ActionMock struct {
	received      solman.UploadAction
	performCalled bool
	failWith      error
}

func (a *ActionMock) WithConnection(c solman.Connection) {
	a.received.Connection = c
}
func (a *ActionMock) WithChangeDocumentID(id string) {
	a.received.ChangeDocumentID = id
}
func (a *ActionMock) WithTransportRequestID(id string) {
	a.received.TransportRequestID = id
}
func (a *ActionMock) WithApplicationID(id string) {
	a.received.ApplicationID = id
}
func (a *ActionMock) WithFile(f string) {
	a.received.File = f
}
func (a *ActionMock) WithCMOpts(opts []string) {
	a.received.CMOpts = opts
}
func (a *ActionMock) Perform(fs solman.FileSystem, command solman.Exec) error {
	a.performCalled = true
	return a.failWith
}

func TestRunTransportRequestUploadSOLMAN(t *testing.T) {
	t.Parallel()

	t.Run("solmand upload", func(t *testing.T) {
		t.Parallel()

		config := transportRequestUploadSOLMANOptions{
			Endpoint:           "https://example.org/solman",
			Username:           "me",
			Password:           "********",
			ApplicationID:      "XYZ",
			ChangeDocumentID:   "12345678",
			TransportRequestID: "87654321",
			FilePath:           "myApp.xxx",
			CmClientOpts:        []string{"-Dtest=abc123"},
		}

		t.Run("straight forward", func(t *testing.T) {
			utils := newTransportRequestUploadSOLMANTestsUtils()
			action := ActionMock{}

			err := runTransportRequestUploadSOLMAN(&config, &action, nil, utils)

			if assert.NoError(t, err) {
				assert.Equal(t, action.received, solman.UploadAction{
					Connection: solman.Connection{
						Endpoint: "https://example.org/solman",
						User:     "me",
						Password: "********",
					},
					ApplicationID:      "XYZ",
					ChangeDocumentID:   "12345678",
					TransportRequestID: "87654321",
					File:               "myApp.xxx",
					CMOpts:             []string{"-Dtest=abc123"},
				})
				assert.True(t, action.performCalled)
			}
		})

		t.Run("Error during deployment", func(t *testing.T) {
			utils := newTransportRequestUploadSOLMANTestsUtils()
			action := ActionMock{failWith: fmt.Errorf("upload failed")}

			err := runTransportRequestUploadSOLMAN(&config, &action, nil, utils)

			assert.Error(t, err, "upload failed")
		})

	})
}
