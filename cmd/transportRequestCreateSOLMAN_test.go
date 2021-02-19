package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/transportrequest/solman"
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

func TestRunTransportRequestCreateSOLMAN(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := transportRequestCreateSOLMANOptions{

			Endpoint:            "https://example.org/solman",
			Username:            "me",
			Password:            "secret",
			ChangeDocumentID:    "123",
			DevelopmentSystemID: "XXX",
			CmClientOpts: []string{"-Dabc=123", "-Dxyz=456"},
		}

		utils := newTransportRequestCreateSOLMANTestsUtils()
		create := solman.CreateAction{}
		cpe := &transportRequestCreateSOLMANCommonPipelineEnvironment{}

		// test
		err := runTransportRequestCreateSOLMAN(&config, &create, nil, utils, cpe)

		// assert
		assert.NoError(t, err)
	})
}
