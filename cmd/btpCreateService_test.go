package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/btp"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/stretchr/testify/assert"
)

func btpMockCleanup(m *btp.BtpExecutorMock) {
	m.ShouldFailOnCommand = map[string]error{}
	m.StdoutReturn = map[string]string{}
	m.Calls = []btp.BtpExecCall{}
}

func TestRunBtpCreateService(t *testing.T) {
	m := &btp.BtpExecutorMock{}
	utils := btp.NewBTPUtils(m)

	var telemetryData telemetry.CustomData

	t.Run("happy path", func(t *testing.T) {
		defer btpMockCleanup(m)
		// init
		config := btpCreateServiceOptions{}

		// test
		err := runBtpCreateService(&config, &telemetryData, *utils)

		// assert
		assert.NoError(t, err)
	})
}
