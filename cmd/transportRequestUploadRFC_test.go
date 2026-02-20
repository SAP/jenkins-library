//go:build unit
// +build unit

package cmd

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/transportrequest/rfc"
	"github.com/stretchr/testify/assert"
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
	failWith     error
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
	return m.failWith
}

type configMock struct {
	config *transportRequestUploadRFCOptions
}

func TestTrRfcRunTransportRequestUpload(t *testing.T) {
	t.Parallel()

	t.Run("good", func(t *testing.T) {
		t.Parallel()

		utils := newTransportRequestUploadRFCTestsUtils()
		configMock := newRfcConfigMock()
		actionMock := uploadMock{}
		cpe := &transportRequestUploadRFCCommonPipelineEnvironment{}

		err := runTransportRequestUploadRFC(configMock.config, &actionMock, nil, utils, cpe)

		if assert.NoError(t, err) {
			t.Run("upload triggered", func(t *testing.T) {
				assert.True(t, actionMock.uploadCalled)
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
						AbapPackage: "ABC",
					},
					Configuration: rfc.UploadConfig{
						AcceptUnixStyleEndOfLine: true,
						CodePage:                 "UTF-8",
						FailUploadOnWarning:      true,
						Verbose:                  false, // comes from general config
					},
					TransportRequestID: "K12345678",
					ApplicationURL:     "http://example.org/myDeployable.zip",
				}, actionMock.received)

				assert.Equal(t, cpe.custom.transportRequestID, "K12345678")
			})
		}
	})
	t.Run("bad", func(t *testing.T) {
		t.Parallel()

		t.Run("Error during deployment", func(t *testing.T) {
			utilsMock := newTransportRequestUploadSOLMANTestsUtils(0)
			configMock := newRfcConfigMock()
			actionMock := uploadMock{failWith: fmt.Errorf("upload failed")}
			cpe := &transportRequestUploadRFCCommonPipelineEnvironment{}

			err := runTransportRequestUploadRFC(configMock.config, &actionMock, nil, utilsMock, cpe)

			assert.Error(t, err, "upload failed")
		})
	})
}

func newRfcConfigMock() *configMock {
	return &configMock{
		config: &transportRequestUploadRFCOptions{
			Endpoint:                   "https://my.abap.server",
			Client:                     "001",
			Instance:                   "00",
			Username:                   "me",
			Password:                   "******",
			ApplicationName:            "MyApp",
			ApplicationDescription:     "Lorem impsum",
			AbapPackage:                "ABC",
			ApplicationURL:             "http://example.org/myDeployable.zip",
			CodePage:                   "UTF-8",
			AcceptUnixStyleLineEndings: true,
			FailUploadOnWarning:        true,
			TransportRequestID:         "K12345678",
		},
	}
}
