//go:build unit

package config

import (
	"fmt"
	"os"
	"path/filepath"
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
	assert.Contains(t, config.Config, "gcsFolderPath")
	assert.Equal(t, gcsFolderPath, config.Config["gcsFolderPath"])
	assert.Contains(t, config.Config, "gcsBucketId")
	assert.Equal(t, gcsBucketID, config.Config["gcsBucketId"])
	assert.NotContains(t, config.Config, "unknownConfig")
}

func TestReportingParams_GetResourceParameters(t *testing.T) {
	tt := []struct {
		in       ReportingParams
		expected map[string]interface{}
	}{
		{
			in:       ReportingParams{Parameters: []StepParameters{}},
			expected: map[string]interface{}{},
		},
		{
			in: ReportingParams{Parameters: []StepParameters{
				{Name: "param1"},
				{Name: "param2"},
			}},
			expected: map[string]interface{}{},
		},
		{
			in: ReportingParams{Parameters: []StepParameters{
				{Name: "param1", ResourceRef: []ResourceReference{}},
				{Name: "param2", ResourceRef: []ResourceReference{}},
			}},
			expected: map[string]interface{}{},
		},
		{
			in: ReportingParams{Parameters: []StepParameters{
				{Name: "param1", ResourceRef: []ResourceReference{{Name: "notAvailable", Param: "envparam1"}}},
				{Name: "param2", ResourceRef: []ResourceReference{{Name: "commonPipelineEnvironment", Param: "envparam2"}}, Type: "string"},
			}},
			expected: map[string]interface{}{"param2": "val2"},
		},
		{
			in: ReportingParams{Parameters: []StepParameters{
				{Name: "param2", ResourceRef: []ResourceReference{{Name: "commonPipelineEnvironment", Param: "envparam2"}}, Type: "string"},
				{Name: "param3", ResourceRef: []ResourceReference{{Name: "commonPipelineEnvironment", Param: "jsonList"}}, Type: "[]string"},
			}},
			expected: map[string]interface{}{"param2": "val2", "param3": []interface{}{"value1", "value2"}},
		},
		{
			in: ReportingParams{Parameters: []StepParameters{
				{Name: "param4", ResourceRef: []ResourceReference{{Name: "commonPipelineEnvironment", Param: "jsonKeyValue"}}, Type: "map[string]interface{}"},
			}},
			expected: map[string]interface{}{"param4": map[string]interface{}{"key": "value"}},
		},
		{
			in: ReportingParams{Parameters: []StepParameters{
				{Name: "param1", ResourceRef: []ResourceReference{{Name: "commonPipelineEnvironment", Param: "envparam1"}}, Type: "noString"},
				{Name: "param4", ResourceRef: []ResourceReference{{Name: "commonPipelineEnvironment", Param: "jsonKeyValueString"}}, Type: "string"},
			}},
			expected: map[string]interface{}{"param4": "{\"key\":\"valueString\"}"},
		},
	}

	dir := t.TempDir()

	cpeDir := filepath.Join(dir, "commonPipelineEnvironment")
	err := os.MkdirAll(cpeDir, 0700)
	if err != nil {
		t.Fatal("Failed to create sub directory")
	}

	os.WriteFile(filepath.Join(cpeDir, "envparam1"), []byte("val1"), 0700)
	os.WriteFile(filepath.Join(cpeDir, "envparam2"), []byte("val2"), 0700)
	os.WriteFile(filepath.Join(cpeDir, "jsonList.json"), []byte("[\"value1\",\"value2\"]"), 0700)
	os.WriteFile(filepath.Join(cpeDir, "jsonKeyValue.json"), []byte("{\"key\":\"value\"}"), 0700)
	os.WriteFile(filepath.Join(cpeDir, "jsonKeyValueString"), []byte("{\"key\":\"valueString\"}"), 0700)

	for run, test := range tt {
		t.Run(fmt.Sprintf("Run %v", run), func(t *testing.T) {
			actual := test.in.GetResourceParameters(dir, "commonPipelineEnvironment")
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestReportingParams_GetGetStepFilters(t *testing.T) {
	tt := []struct {
		in       ReportingParams
		expected StepFilters
	}{
		{
			in:       ReportingParams{Parameters: []StepParameters{}},
			expected: StepFilters{},
		},
		{
			in: ReportingParams{Parameters: []StepParameters{
				{Name: "param1"},
				{Name: "param2"},
			}},
			expected: StepFilters{
				All:     []string{"param1", "param2"},
				General: []string{"param1", "param2"},
				Steps:   []string{"param1", "param2"},
				Stages:  []string{"param1", "param2"},
			},
		},
		{
			in: ReportingParams{Parameters: []StepParameters{
				{Name: "param1"},
				{Name: "param2"},
				{Name: "param3"},
				{Name: "param4"},
				{Name: "param5"},
				{Name: "param6"},
			}},
			expected: StepFilters{
				All:     []string{"param1", "param2", "param3", "param4", "param5", "param6"},
				General: []string{"param1", "param2", "param3", "param4", "param5", "param6"},
				Steps:   []string{"param1", "param2", "param3", "param4", "param5", "param6"},
				Stages:  []string{"param1", "param2", "param3", "param4", "param5", "param6"},
			},
		},
	}

	for run, test := range tt {
		t.Run(fmt.Sprintf("Run %v", run), func(t *testing.T) {
			actual := test.in.getStepFilters()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestReportingParams_GetReportingFilter(t *testing.T) {
	tt := []struct {
		in       ReportingParams
		expected []string
	}{
		{
			in:       ReportingParams{Parameters: []StepParameters{}},
			expected: nil,
		},
		{
			in: ReportingParams{Parameters: []StepParameters{
				{Name: "param1"},
				{Name: "param2"},
			}},
			expected: []string{"param1", "param2"},
		},
		{
			in: ReportingParams{Parameters: []StepParameters{
				{Name: "param1"},
				{Name: "param2"},
				{Name: "param3"},
				{Name: "param4"},
				{Name: "param5"},
				{Name: "param6"},
			}},
			expected: []string{"param1", "param2", "param3", "param4", "param5", "param6"},
		},
	}

	for run, test := range tt {
		t.Run(fmt.Sprintf("Run %v", run), func(t *testing.T) {
			actual := test.in.getReportingFilter()
			assert.Equal(t, test.expected, actual)
		})
	}
}
