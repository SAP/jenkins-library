package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/transportrequest/solman"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type transportRequestCreateSOLMANMockUtils struct {
	*mock.ExecMockRunner
}

func newTransportRequestCreateSOLMANTestsUtils() transportRequestCreateSOLMANMockUtils {
	utils := transportRequestCreateSOLMANMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
	}
	return utils
}

type createMock struct {
	received           solman.CreateAction
	transportRequestID string
	shouldFail         error
}

func (a *createMock) WithConnection(c solman.Connection) {
	a.received.Connection = c
}

func (a *createMock) WithChangeDocumentID(id string) {
	a.received.ChangeDocumentID = id
}

func (a *createMock) WithDevelopmentSystemID(id string) {
	a.received.DevelopmentSystemID = id
}

func (a *createMock) WithCMOpts(opts []string) {
	a.received.CMOpts = opts
}

func (a *createMock) Perform(command solman.Exec) (string, error) {
	return a.transportRequestID, a.shouldFail
}

func TestRunTransportRequestCreateSOLMAN(t *testing.T) {
	t.Parallel()

	defaultConfig := transportRequestCreateSOLMANOptions{

		Endpoint:            "https://example.org/solman",
		Username:            "me",
		Password:            "secret",
		ChangeDocumentID:    "123",
		DevelopmentSystemID: "XXX",
		CmClientOpts:        []string{"-Dabc=123", "-Dxyz=456"},
	}

	t.Run("straight forward", func(t *testing.T) {
		t.Parallel()
		// init
		config := defaultConfig

		utils := newTransportRequestCreateSOLMANTestsUtils()
		create := createMock{transportRequestID: "XYZK12345678"}
		cpe := &transportRequestCreateSOLMANCommonPipelineEnvironment{}

		// test
		err := runTransportRequestCreateSOLMAN(&config, &create, nil, utils, cpe)

		// assert
		if assert.NoError(t, err) {

			t.Run("assert received parameters", func(t *testing.T) {
				assert.Equal(t, solman.CreateAction{
					Connection: solman.Connection{
						Endpoint: "https://example.org/solman",
						User:     "me",
						Password: "secret",
					},
					ChangeDocumentID:    "123",
					DevelopmentSystemID: "XXX",
					CMOpts: []string{
						"-Dabc=123",
						"-Dxyz=456",
					},
				}, create.received)
			})

			t.Run("assert transport request id on CPE", func(t *testing.T) {
				assert.Equal(t, "XYZK12345678", cpe.custom.transportRequestID)
			})
		}
	})

	t.Run("error case", func(t *testing.T) {
		t.Parallel()
		// init
		config := defaultConfig

		utils := newTransportRequestCreateSOLMANTestsUtils()
		create := createMock{transportRequestID: "XYZK12345678", shouldFail: errors.New("something went wrong")}
		cpe := &transportRequestCreateSOLMANCommonPipelineEnvironment{}

		// test
		err := runTransportRequestCreateSOLMAN(&config, &create, nil, utils, cpe)

		// assert
		if assert.EqualError(t, err, "something went wrong") {

			t.Run("assert transport request id on CPE unset", func(t *testing.T) {
				assert.Empty(t, cpe.custom.transportRequestID)
			})
		}
	})
}
