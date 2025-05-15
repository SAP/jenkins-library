//go:build unit
// +build unit

package config

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func initRunConfigGlobMock(pattern string) ([]string, error) {
	matches := []string{}
	switch pattern {
	case "**/file1":
		matches = append(matches, "file1")
	case "directory/file2":
		matches = append(matches, "file2")
	}
	return matches, nil
}

func TestInitRunConfigV1(t *testing.T) {
	tt := []struct {
		name              string
		config            Config
		stageConfig       string
		runStagesExpected map[string]bool
		runStepsExpected  map[string]map[string]bool
		expectedError     error
		errorContains     string
	}{
		{
			name:             "success",
			config:           Config{Stages: map[string]map[string]interface{}{"testStage": {"testKey": "testVal"}}},
			stageConfig:      "spec:\n  stages:\n  - name: testStage\n    displayName: testStage\n    steps:\n    - name: testStep\n      conditions:\n      - configKey: testKey",
			runStepsExpected: map[string]map[string]bool{},
		},
		{
			name:             "error - load conditions",
			stageConfig:      "wrong stage config format",
			runStepsExpected: map[string]map[string]bool{},
			errorContains:    "failed to load pipeline run conditions",
		},
		{
			name:             "error - evaluate conditions",
			config:           Config{Stages: map[string]map[string]interface{}{"testStage": {"testKey": "testVal"}}},
			runStepsExpected: map[string]map[string]bool{},
			stageConfig:      "spec:\n  stages:\n  - name: testStage\n    displayName: testStage\n    steps:\n    - name: testStep\n      conditions:\n      - config:\n          configKey1:\n          - configVal1\n          configKey2:\n          - configVal2",
			errorContains:    "failed to evaluate step conditions",
		},
	}

	filesMock := mock.FilesMock{}

	for _, test := range tt {
		stageConfig := io.NopCloser(strings.NewReader(test.stageConfig))
		runConfig := RunConfig{StageConfigFile: stageConfig}
		runConfigV1 := RunConfigV1{RunConfig: runConfig}
		err := runConfigV1.InitRunConfigV1(&test.config, &filesMock, ".pipeline")
		if len(test.errorContains) > 0 {
			assert.Contains(t, fmt.Sprint(err), test.errorContains)
		} else {
			assert.NoError(t, err)
		}

	}
}
