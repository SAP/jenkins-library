package events

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestPublishTaskRunFinishedEvent_Disabled(t *testing.T) {
	generalConfig := config.GeneralConfigOptions{}
	err := PublishTaskRunFinishedEvent(nil, generalConfig, "stage1", "step1", "0")
	assert.NoError(t, err)
}

func TestPublishTaskRunFinishedEvent_NilTokenProvider(t *testing.T) {
	generalConfig := config.GeneralConfigOptions{
		HookConfig: config.HookConfiguration{
			GCPPubSubConfig: config.GCPPubSubConfiguration{
				Enabled:       true,
				ProjectNumber: "123",
			},
		},
	}
	err := PublishTaskRunFinishedEvent(nil, generalConfig, "stage1", "step1", "0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no OIDC token provider")
}