package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/transportrequest/solman"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

type transportRequestUploadSOLMANMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newTransportRequestUploadSOLMANTestsUtils(exitcode int) transportRequestUploadSOLMANMockUtils {
	utils := transportRequestUploadSOLMANMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{ExitCode: exitcode},
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

type ConfigMock struct {
	config *transportRequestUploadSOLMANOptions
}

func (m *ConfigMock) with(field string, value string) *ConfigMock {
	r := reflect.ValueOf(m.config)
	f := reflect.Indirect(r).FieldByName(field)
	f.SetString(value)
	return m
}

func (m *ConfigMock) without(field string) *ConfigMock {
	return m.with(field, "")
}

type transportRequestUtilsMock struct {
	trID string
	cdID string
}

func (m *transportRequestUtilsMock) FindIDInRange(label, from, to string) (string, error) {
	if strings.HasPrefix(label, "TransportRequest") {
		return m.trID, nil
	}
	if strings.HasPrefix(label, "ChangeDocument") {
		return m.cdID, nil
	}

	return "invalid", fmt.Errorf("invalid label passed: %s", label)
}

func TestTrSolmanRunTransportRequestUpload(t *testing.T) {
	t.Parallel()

	t.Run("solmand upload", func(t *testing.T) {
		t.Parallel()

		t.Run("straight forward", func(t *testing.T) {
			utilsMock := newTransportRequestUploadSOLMANTestsUtils(0)
			configMock := newConfigMock()
			actionMock := ActionMock{}
			cpe := &transportRequestUploadSOLMANCommonPipelineEnvironment{}

			err := runTransportRequestUploadSOLMAN(configMock.config, &actionMock, nil, utilsMock, cpe)

			if assert.NoError(t, err) {
				assert.Equal(t, actionMock.received, solman.UploadAction{
					Connection: solman.Connection{
						Endpoint: "https://example.org/solman",
						User:     "me",
						Password: "********",
					},
					ApplicationID:      "XYZ",
					ChangeDocumentID:   "12345678",
					TransportRequestID: "87654321",
					File:               "myApp.abc",
					CMOpts:             []string{"-Dtest=abc123"},
				})
				assert.True(t, actionMock.performCalled)
			}
		})

		t.Run("Error during deployment", func(t *testing.T) {
			utilsMock := newTransportRequestUploadSOLMANTestsUtils(0)
			configMock := newConfigMock()
			actionMock := ActionMock{failWith: fmt.Errorf("upload failed")}
			cpe := &transportRequestUploadSOLMANCommonPipelineEnvironment{}

			err := runTransportRequestUploadSOLMAN(configMock.config, &actionMock, nil, utilsMock, cpe)

			assert.Error(t, err, "upload failed")
		})

	})
}

func TestTrSolmanGetTransportRequestID(t *testing.T) {
	t.Parallel()

	t.Run("get transport request id", func(t *testing.T) {
		t.Parallel()

		t.Run("TransportRequestID from config", func(t *testing.T) {
			configMock := newConfigMock()

			id, err := getTransportRequestID(configMock.config, &transportRequestUtilsMock{trID: "43218765", cdID: "56781234"})

			if assert.NoError(t, err) {
				assert.Equal(t, id, "87654321")
			}
		})
		t.Run("TransportRequestID from git commit", func(t *testing.T) {
			configMock := newConfigMock().without("TransportRequestID")

			id, err := getTransportRequestID(configMock.config, &transportRequestUtilsMock{trID: "43218765", cdID: "56781234"})

			if assert.NoError(t, err) {
				assert.Equal(t, id, "43218765")
			}
		})
	})
}

func TestTrSolmanGetChangeDocumentID(t *testing.T) {
	t.Parallel()

	t.Run("get change document id", func(t *testing.T) {
		t.Parallel()

		t.Run("ChangeDocumentID from config", func(t *testing.T) {
			configMock := newConfigMock()

			id, err := getChangeDocumentID(configMock.config, &transportRequestUtilsMock{trID: "43218765", cdID: "56781234"})

			if assert.NoError(t, err) {
				assert.Equal(t, id, "12345678")
			}
		})
		t.Run("ChangeDocumentID from git commit", func(t *testing.T) {
			configMock := newConfigMock().without("ChangeDocumentID")

			id, err := getChangeDocumentID(configMock.config, &transportRequestUtilsMock{trID: "43218765", cdID: "56781234"})

			if assert.NoError(t, err) {
				assert.Equal(t, id, "56781234")
			}
		})
	})
}

func newConfigMock() *ConfigMock {
	return &ConfigMock{
		config: &transportRequestUploadSOLMANOptions{
			Endpoint:              "https://example.org/solman",
			Username:              "me",
			Password:              "********",
			ApplicationID:         "XYZ",
			ChangeDocumentID:      "12345678",
			TransportRequestID:    "87654321",
			FilePath:              "myApp.abc",
			CmClientOpts:          []string{"-Dtest=abc123"},
			GitFrom:               "origin/master",
			GitTo:                 "HEAD",
			ChangeDocumentLabel:   "ChangeDocument",
			TransportRequestLabel: "TransportRequest",
		},
	}
}
