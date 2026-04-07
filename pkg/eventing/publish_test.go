//go:build unit

package eventing

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/stretchr/testify/assert"
)

func TestProcessCDE_NilTokenProvider(t *testing.T) {
	generalConfig := config.GeneralConfigOptions{}
	err := ProcessCDE(nil, &generalConfig, EventContext{
		StepName:  "step1",
		StageName: "stage1",
		ErrorCode: "0",
	})
	assert.NoError(t, err)
}

func TestProcess_NilTokenProvider(t *testing.T) {
	generalConfig := config.GeneralConfigOptions{}
	err := Process(nil, &generalConfig, EventContext{
		StepName:  "step1",
		StageName: "stage1",
		ErrorCode: "0",
	})
	assert.NoError(t, err)
}

func TestProcessCDE_WithTopic(t *testing.T) {
	mockProvider := gcp.OIDCTokenProvider(func(roleID string) (string, error) {
		return "", fmt.Errorf("mock error")
	})

	generalConfig := &config.GeneralConfigOptions{
		HookConfig: config.HookConfiguration{
			GCPPubSubConfig: config.GCPPubSubConfiguration{
				Topic:            "my-topic",
				Source:           "test-source",
				ProjectNumber:    "123456",
				IdentityPool:     "test-pool",
				IdentityProvider: "test-provider",
			},
		},
	}

	err := ProcessCDE(mockProvider, generalConfig, EventContext{
		StepName:  "step1",
		StageName: "stage1",
		ErrorCode: "0",
	})

	// Should fail at publish and not earlier
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event publish failed")
}

func TestProcessCDE_EmptyTopic(t *testing.T) {
	mockProvider := gcp.OIDCTokenProvider(func(roleID string) (string, error) {
		return "test-token", nil
	})

	generalConfig := &config.GeneralConfigOptions{
		HookConfig: config.HookConfiguration{
			GCPPubSubConfig: config.GCPPubSubConfiguration{
				TopicPrefix:      "prefix-",
				Topic:            "", // Empty topic should error
				Source:           "test-source",
				ProjectNumber:    "123456",
				IdentityPool:     "test-pool",
				IdentityProvider: "test-provider",
			},
		},
	}

	err := ProcessCDE(mockProvider, generalConfig, EventContext{
		StepName:  "step1",
		StageName: "stage1",
		ErrorCode: "0",
	})

	// Should return error about missing topic
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "topic is required")
}

func TestProcess_WithTopic(t *testing.T) {
	mockProvider := gcp.OIDCTokenProvider(func(roleID string) (string, error) {
		return "", fmt.Errorf("mock error")
	})

	generalConfig := &config.GeneralConfigOptions{
		HookConfig: config.HookConfiguration{
			GCPPubSubConfig: config.GCPPubSubConfiguration{
				Topic:            "legacy-topic",
				Source:           "test-source",
				TypePrefix:       "com.sap.",
				ProjectNumber:    "123456",
				IdentityPool:     "test-pool",
				IdentityProvider: "test-provider",
			},
		},
	}

	err := Process(mockProvider, generalConfig, EventContext{
		StepName:  "step1",
		StageName: "stage1",
		ErrorCode: "0",
	})

	// Should fail at publish and not earlier
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event publish failed")
}

func TestProcess_EmptyTopic(t *testing.T) {
	mockProvider := gcp.OIDCTokenProvider(func(roleID string) (string, error) {
		return "test-token", nil
	})

	generalConfig := &config.GeneralConfigOptions{
		HookConfig: config.HookConfiguration{
			GCPPubSubConfig: config.GCPPubSubConfiguration{
				TopicPrefix:      "prefix-",
				Topic:            "", // Empty topic should error
				Source:           "test-source",
				TypePrefix:       "com.sap.",
				ProjectNumber:    "123456",
				IdentityPool:     "test-pool",
				IdentityProvider: "test-provider",
			},
		},
	}

	err := Process(mockProvider, generalConfig, EventContext{
		StepName:  "step1",
		StageName: "stage1",
		ErrorCode: "0",
	})

	// Should return error about missing topic
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "topic is required")
}
