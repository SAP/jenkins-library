package events

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestPublishTaskRunFinishedEvent_NilTokenProvider(t *testing.T) {
	generalConfig := config.GeneralConfigOptions{}
	err := PublishTaskRunFinishedEvent(nil, generalConfig, "stage1", "step1", "0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no OIDC token provider")
}
