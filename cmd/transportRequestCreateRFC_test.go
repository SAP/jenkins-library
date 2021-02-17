package cmd

import (
	"bytes"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/transportrequest/rfc"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type transportRequestCreateRFCMockUtils struct {
	*mock.ExecMockRunner
}

func newTransportRequestCreateRFCTestsUtils() transportRequestCreateRFCMockUtils {
	utils := transportRequestCreateRFCMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
	}
	return utils
}

type createActionMock struct {
	received           rfc.CreateAction
	transportRequestID string
	shouldFail         error
}

// WithConnection Set the connection details
func (c *createActionMock) WithConnection(con rfc.Connection) {
	c.received.Connection = con
}

// WithTransportType Sets the transport type
func (c *createActionMock) WithTransportType(t string) {
	c.received.TransportType = t
}

// WithTargetSystemID Sets the target system
func (c *createActionMock) WithTargetSystemID(t string) {
	c.received.TargetSystemID = t
}

// WithDescription Sets the description
func (c *createActionMock) WithDescription(d string) {
	c.received.Description = d
}

// Perform Creates the transport request
func (c *createActionMock) Perform(command rfc.Exec) (string, error) {
	return c.transportRequestID, c.shouldFail
}

func TestRunTransportRequestCreateRFC(t *testing.T) {
	t.Parallel()

	defaultConfig := transportRequestCreateRFCOptions{
		Endpoint:      "https://example.org/rfc",
		Username:      "me",
		Password:      "********",
		Client:        "001",
		Instance:      "DEV",
		TransportType: "X",
		TargetSystem:  "YYY",
		Description:   "Lorem ipsum",
	}

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init

		config := defaultConfig
		utils := newTransportRequestCreateRFCTestsUtils()
		cpe := transportRequestCreateRFCCommonPipelineEnvironment{}

		createMock := createActionMock{transportRequestID: "YYYK123456"}
		var stdout bytes.Buffer

		// test
		err := runTransportRequestCreateRFC(&config, &createMock, nil, utils, &cpe, &stdout)

		// assert
		if assert.NoError(t, err) {
			t.Run("Assert the parameters", func(t *testing.T) {
				assert.Equal(t, rfc.CreateAction{
					Connection: rfc.Connection{
						Endpoint: "https://example.org/rfc",
						Client:   "001",
						Instance: "DEV",
						User:     "me",
						Password: "********",
					},
					TransportType:  "X",
					TargetSystemID: "YYY",
					Description:    "Lorem ipsum",
				}, createMock.received)
			})

			t.Run("Assert the Common Pipeline Environment", func(t *testing.T) {
				assert.Equal(t, "YYYK123456", cpe.custom.transportRequestID)
			})

			t.Run("Assert transport request id returned via stdout", func(t *testing.T) {
				// will only work if the framework does not write to stdout, that we can't test here ...
				assert.Equal(t, "YYYK123456", stdout.String())
			})
		}
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := defaultConfig
		utils := newTransportRequestCreateRFCTestsUtils()
		cpe := transportRequestCreateRFCCommonPipelineEnvironment{}
		createMock := createActionMock{transportRequestID: "YYYK123456", shouldFail: errors.New("something went wrong")}
		var stdout bytes.Buffer
		// test
		err := runTransportRequestCreateRFC(&config, &createMock, nil, utils, &cpe, &stdout)
		// assert
		if assert.EqualError(t, err, "something went wrong") {

			t.Run("assert CPE unset", func(t *testing.T) {
				assert.Empty(t, cpe.custom.transportRequestID)
			})
			t.Run("assert stdout empty", func(t *testing.T) {
				assert.Empty(t, stdout.String())
			})
		}
	})
}
