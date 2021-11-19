package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMixinReportingConfig(t *testing.T) {
	gcpJsonKeyFilePath := "path/key.json"
	gcsFolderPath := "test/folder/path"
	gcsBucketID := "testBucketId"
	config := StepConfig{
		Config:     map[string]interface{}{},
		HookConfig: nil,
	}
	general := map[string]interface{}{
		"gcpJsonKeyFilePath": gcpJsonKeyFilePath,
		"gcsFolderPath":      gcsFolderPath,
		"gcsBucketId":        "generalBucketId",
	}
	steps := map[string]interface{}{
		"gcsBucketId":   gcsBucketID,
		"unknownConfig": "test",
	}

	config.mixinReportingConfig(nil, general, steps)

	assert.Contains(t, config.Config, "gcpJsonKeyFilePath")
	assert.Equal(t, gcpJsonKeyFilePath, config.Config["gcpJsonKeyFilePath"])
	assert.Contains(t, config.Config, "gcpJsonKeyFilePath")
	assert.Equal(t, gcsFolderPath, config.Config["gcsFolderPath"])
	assert.Contains(t, config.Config, "gcsBucketId")
	assert.Equal(t, gcsBucketID, config.Config["gcsBucketId"])
	assert.NotContains(t, config.Config, "unknownConfig")

}
