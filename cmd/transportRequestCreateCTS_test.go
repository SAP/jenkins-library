package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/transportrequest/cts"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type transportRequestCreateCTSMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newTransportRequestCreateCTSTestsUtils() transportRequestCreateCTSMockUtils {
	utils := transportRequestCreateCTSMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

type createMock struct {
	received           cts.CreateAction
	transportRequestId string
	shouldFail         error
}

func (m *createMock) WithConnection(c cts.Connection) {
	m.received.Connection = c
}

func (m *createMock) WithTransportType(t string) {
	m.received.TransportType = t
}

func (m *createMock) WithTargetSystemID(t string) {
	m.received.TargetSystemID = t
}

func (m *createMock) WithDescription(d string) {
	m.received.Description = d
}

func (m *createMock) WithCMOpts(opts []string) {
	m.received.CMOpts = opts
}

func (m *createMock) Perform(command cts.Exec) (string, error) {
	return m.transportRequestId, m.shouldFail
}

func TestRunTransportRequestCreateCTS(t *testing.T) {
	t.Parallel()

	defaultConfig := transportRequestCreateCTSOptions{
		Endpoint:      "https://example.org/cts",
		Username:      "me",
		Password:      "******",
		Description:   "Lorem ipsum",
		TransportType: "X",
		TargetSystem:  "XYZ",
		CmClientOpts:  []string{"-Dx=y", "-Dabc=123"},
	}

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		config := defaultConfig
		cpe := transportRequestCreateCTSCommonPipelineEnvironment{}
		utils := newTransportRequestCreateCTSTestsUtils()
		create := createMock{
			transportRequestId: "XYZK12345678",
		}

		err := runTransportRequestCreateCTS(
			&config,
			nil,
			&utils,
			&cpe,
			&create,
		)

		if assert.NoError(t, err) {
			t.Run("assert parameters", func(t *testing.T) {
				assert.Equal(t, cts.CreateAction{
					Connection: cts.Connection{
						Endpoint: "https://example.org/cts",
						User:     "me",
						Password: "******",
					},
					Description:    "Lorem ipsum",
					TransportType:  "X",
					TargetSystemID: "XYZ",
					CMOpts:         []string{"-Dx=y", "-Dabc=123"},
				}, create.received)
			})

			t.Run("assert transport request id in cpe", func(t *testing.T) {
				assert.Equal(t, "XYZK12345678", cpe.custom.transportRequestID)
			})
		}
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()

		config := defaultConfig
		cpe := transportRequestCreateCTSCommonPipelineEnvironment{}
		utils := newTransportRequestCreateCTSTestsUtils()
		create := createMock{
			transportRequestId: "XYZK12345678",
			shouldFail:         errors.New("Cannot create transport request"),
		}

		err := runTransportRequestCreateCTS(
			&config,
			nil,
			&utils,
			&cpe,
			&create,
		)

		if assert.EqualError(t, err, "Cannot create transport request") {

			t.Run("assert transport request id in cpe is empty ", func(t *testing.T) {
				assert.Empty(t, cpe.custom.transportRequestID)
			})
		}
	})
}
