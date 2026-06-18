//go:build unit

package eventing

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestPublishTaskRunFinishedCDEvent_NilTokenProvider(t *testing.T) {
	generalConfig := config.GeneralConfigOptions{}
	err := PublishTaskRunFinishedCDEvent(nil, &generalConfig, EventContext{
		StepName:  "step1",
		StageName: "stage1",
		ErrorCode: "0",
	})
	assert.NoError(t, err)
}

func TestPublishTaskRunFinishedEvent_NilTokenProvider(t *testing.T) {
	generalConfig := config.GeneralConfigOptions{}
	err := PublishTaskRunFinishedEvent(nil, &generalConfig, EventContext{
		StepName:  "step1",
		StageName: "stage1",
		ErrorCode: "0",
	})
	assert.NoError(t, err)
}
