package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/transportrequest/rfc"
	"github.com/stretchr/testify/assert"
	"testing"
)

type transportRequestUploadRFCMockUtils struct {
	*mock.ExecMockRunner
}

func newTransportRequestUploadRFCTestsUtils() transportRequestUploadRFCMockUtils {
	utils := transportRequestUploadRFCMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
	}
	return utils
}

type uploadMock struct {
	received     rfc.UploadAction
	uploadCalled bool
}

// WithApplicationURL The location of the deployable
func (m *uploadMock) WithApplicationURL(z string) {
	m.received.ApplicationURL = z
}

// WithTransportRequestID The transport request ID for the upload
func (m *uploadMock) WithTransportRequestID(t string) {
	m.received.TransportRequestID = t
}

// WithApplication Everything we need to know about the application
func (m *uploadMock) WithApplication(a rfc.Application) {
	m.received.Application = a
}

// WithConfiguration Everything we need to know in order to perform the upload
func (m *uploadMock) WithConfiguration(c rfc.UploadConfig) {
	m.received.Configuration = c
}

// WithConnection Everything we need to know about the connection
func (m *uploadMock) WithConnection(c rfc.Connection) {
	m.received.Connection = c
}

func (m *uploadMock) Perform(exec rfc.Exec) error {
	m.uploadCalled = true
	return nil
}

func TestRunTransportRequestUploadRFC(t *testing.T) {
	t.Parallel()

	config := transportRequestUploadRFCOptions{
		Endpoint:                   "https://my.abap.server",
		Client:                     "001",
		Instance:                   "00",
		Username:                   "me",
		Password:                   "******",
		ApplicationName:            "MyApp",
		ApplicationDescription:     "Lorem impsum",
		AbapPackage:                "XX",
		ApplicationURL:             "http://example.org/myDeployable.zip",
		CodePage:                   "UTF-8",
		AcceptUnixStyleLineEndings: true,
		FailUploadOnWarning:        true,
		TransportRequestID:         "XXXK12345678",
	}

	utils := newTransportRequestUploadRFCTestsUtils()

	mock := uploadMock{}

	err := runTransportRequestUploadRFC(&config, &mock, nil, utils)

	if assert.NoError(t, err) {
		t.Run("upload triggered", func(t *testing.T) {
			assert.True(t, mock.uploadCalled)
		})
		t.Run("parameters has been marshalled", func(t *testing.T) {
			assert.Equal(t, rfc.UploadAction{
				Connection: rfc.Connection{
					Endpoint: "https://my.abap.server",
					Client:   "001",
					Instance: "00",
					User:     "me",
					Password: "******",
				},
				Application: rfc.Application{
					Name:        "MyApp",
					Description: "Lorem impsum",
					AbapPackage: "XX",
				},
				Configuration: rfc.UploadConfig{
					AcceptUnixStyleEndOfLine: true,
					CodePage:                 "UTF-8",
					FailUploadOnWarning:      true,
					Verbose:                  false, // comes from general config
				},
				TransportRequestID: "XXXK12345678",
				ApplicationURL:     "http://example.org/myDeployable.zip",
			}, mock.received)
		})
	}
}
