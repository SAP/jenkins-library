//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type validateBOMMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newValidateBOMTestsUtils() validateBOMMockUtils {
	utils := validateBOMMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

const validBOM = `<?xml version="1.0" encoding="UTF-8"?>
<bom xmlns="http://cyclonedx.org/schema/bom/1.4" version="1">
  <metadata>
    <component type="application">
      <name>test-app</name>
      <version>1.0.0</version>
      <purl>pkg:npm/test-app@1.0.0</purl>
    </component>
  </metadata>
</bom>`

const invalidBOM = `<?xml version="1.0" encoding="UTF-8"?>
<invalid>Not a valid BOM</invalid>`

func TestRunValidateBOM(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		config        validateBOMOptions
		setupFiles    map[string]string
		expectError   bool
		errorContains string
	}{
		{
			name: "success - no BOM files found",
			config: validateBOMOptions{
				BomPattern:            "**/bom-*.xml",
				FailOnValidationError: false,
				ValidatePurl:          true,
				Skip:                  false,
			},
			setupFiles: map[string]string{
				"other-file.txt": "content",
			},
			expectError: false,
		},
		{
			name: "success - skip validation",
			config: validateBOMOptions{
				BomPattern:            "**/bom-*.xml",
				FailOnValidationError: false,
				ValidatePurl:          true,
				Skip:                  true,
			},
			setupFiles:  map[string]string{},
			expectError: false,
		},
		{
			name: "success - valid BOM file",
			config: validateBOMOptions{
				BomPattern:            "bom-test.xml",
				FailOnValidationError: false,
				ValidatePurl:          true,
				Skip:                  false,
			},
			setupFiles: map[string]string{
				"bom-test.xml": validBOM,
			},
			expectError: false,
		},
		{
			name: "warn - invalid BOM with failOnValidationError false",
			config: validateBOMOptions{
				BomPattern:            "bom-invalid.xml",
				FailOnValidationError: false,
				ValidatePurl:          true,
				Skip:                  false,
			},
			setupFiles: map[string]string{
				"bom-invalid.xml": invalidBOM,
			},
			expectError: false,
		},
		{
			name: "error - invalid BOM with failOnValidationError true",
			config: validateBOMOptions{
				BomPattern:            "bom-invalid.xml",
				FailOnValidationError: true,
				ValidatePurl:          true,
				Skip:                  false,
			},
			setupFiles: map[string]string{
				"bom-invalid.xml": invalidBOM,
			},
			expectError:   true,
			errorContains: "BOM validation failed",
		},
		{
			name: "success - multiple BOM files",
			config: validateBOMOptions{
				BomPattern:            "bom-*.xml",
				FailOnValidationError: false,
				ValidatePurl:          true,
				Skip:                  false,
			},
			setupFiles: map[string]string{
				"bom-docker-0.xml": validBOM,
				"bom-mta.xml":      validBOM,
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			utils := newValidateBOMTestsUtils()
			for filename, content := range tc.setupFiles {
				utils.AddFile(filename, []byte(content))
			}

			// Test
			err := runValidateBOM(&tc.config, nil, utils)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
