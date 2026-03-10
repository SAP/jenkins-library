//go:build unit

package eventing

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestProcessCDE_NilTokenProvider(t *testing.T) {
	generalConfig := config.GeneralConfigOptions{}
	err := ProcessCDE(nil, &generalConfig, EventContext{
		StepName:  "step1",
		StageName: "stage1",
		ErrorCode: "0",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no OIDC token provider")
}

func TestProcess_NilTokenProvider(t *testing.T) {
	generalConfig := config.GeneralConfigOptions{}
	err := Process(nil, &generalConfig, EventContext{
		StepName:  "step1",
		StageName: "stage1",
		ErrorCode: "0",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no OIDC token provider")
}
