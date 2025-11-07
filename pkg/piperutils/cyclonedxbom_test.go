package piperutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTempFile(t *testing.T, content string) (string, func()) {
	dir := t.TempDir()
	fileName := filepath.Join(dir, "test.xml")
	err := os.WriteFile(fileName, []byte(content), 0666)
	if err != nil {
		t.Fatalf("Failed to create temp file: %s", err)
	}
	return fileName, func() {
		os.Remove(fileName)
	}
}

const validBom = `<bom>
					<metadata>
						<component>
							<purl>pkg:maven/com.example/mycomponent@1.0.0</purl>
						</component>
					</metadata>
				</bom>`

func TestGetBom(t *testing.T) {
	tests := []struct {
		name          string
		xmlContent    string
		expectedBom   Bom
		errorContains string
	}{
		{
			name:       "valid file",
			xmlContent: validBom,
			expectedBom: Bom{
				Metadata: Metadata{
					Component: BomComponent{
						Purl: "pkg:maven/com.example/mycomponent@1.0.0",
					},
				},
			},
		},
		{
			name:          "file not found",
			xmlContent:    "",
			expectedBom:   Bom{},
			errorContains: "no such file or directory",
		},
		{
			name:          "invalid XML file",
			xmlContent:    "<bom><metadata><component><purl>invalid xml</metadata></bom>",
			expectedBom:   Bom{},
			errorContains: "XML syntax error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fileName string
			var cleanup func()
			if tt.xmlContent != "" {
				fileName, cleanup = createTempFile(t, tt.xmlContent)
				defer cleanup()
			} else {
				// Use a non-existent file path
				fileName = "nonexistent.xml"
			}

			bom, err := GetBom(fileName)

			if tt.errorContains != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBom.Metadata.Component.Purl, bom.Metadata.Component.Purl)
			}
		})
	}
}

func TestGetPurl(t *testing.T) {
	tests := []struct {
		name         string
		xmlContent   string
		expectedPurl string
	}{
		{
			name:         "valid BOM file",
			xmlContent:   validBom,
			expectedPurl: "pkg:maven/com.example/mycomponent@1.0.0",
		},
		{
			name:         "BOM file not found",
			xmlContent:   "",
			expectedPurl: "",
		},
		{
			name:         "invalid BOM file",
			xmlContent:   "<bom><metadata><component><purl>invalid xml</metadata></bom>",
			expectedPurl: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filePath string
			var cleanup func()
			if tt.xmlContent != "" {
				filePath, cleanup = createTempFile(t, tt.xmlContent)
				defer cleanup()
			} else {
				// Use a non-existent file path
				filePath = "nonexistent.xml"
			}

			purl := GetComponent(filePath).Purl
			assert.Equal(t, tt.expectedPurl, purl)
		})
	}
}

func TestValidateBOM(t *testing.T) {
	tests := []struct {
		name          string
		bomContent    string
		errorContains string
	}{
		{
			name: "valid CycloneDX 1.4 BOM with PURL",
			bomContent: `<?xml version="1.0"?>
<bom xmlns="http://cyclonedx.org/schema/bom/1.4" version="1">
	<metadata>
		<component type="application">
			<name>test-app</name>
			<version>1.0.0</version>
			<purl>pkg:maven/com.example/myapp@1.0.0</purl>
		</component>
	</metadata>
</bom>`,
		},
		{
			name: "BOM missing PURL",
			bomContent: `<?xml version="1.0"?>
<bom xmlns="http://cyclonedx.org/schema/bom/1.4" version="1">
	<metadata>
		<component type="application">
			<name>test-app</name>
			<version>1.0.0</version>
			<purl></purl>
		</component>
	</metadata>
</bom>`,
			errorContains: "purl is mandatory",
		},
		{
			name: "BOM with invalid PURL format",
			bomContent: `<?xml version="1.0"?>
<bom xmlns="http://cyclonedx.org/schema/bom/1.4" version="1">
	<metadata>
		<component type="application">
			<name>test-app</name>
			<version>1.0.0</version>
			<purl>invalid-purl</purl>
		</component>
	</metadata>
</bom>`,
			errorContains: "must start with 'pkg:'",
		},
		{
			name: "BOM missing component name",
			bomContent: `<?xml version="1.0"?>
<bom xmlns="http://cyclonedx.org/schema/bom/1.4" version="1">
	<metadata>
		<component type="application">
			<purl>pkg:maven/com.example/myapp@1.0.0</purl>
		</component>
	</metadata>
</bom>`,
			errorContains: "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBOM([]byte(tt.bomContent))

			if tt.errorContains != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePurl(t *testing.T) {
	tests := []struct {
		name          string
		purl          string
		errorContains string
	}{
		{
			name: "valid Maven PURL",
			purl: "pkg:maven/com.example/myapp@1.0.0",
		},
		{
			name: "valid NPM PURL",
			purl: "pkg:npm/express@4.18.2",
		},
		{
			name: "valid PyPI PURL",
			purl: "pkg:pypi/django@3.2.0",
		},
		{
			name:          "empty PURL",
			purl:          "",
			errorContains: "mandatory but was empty",
		},
		{
			name:          "invalid PURL - no pkg prefix",
			purl:          "maven/com.example/myapp@1.0.0",
			errorContains: "must start with 'pkg:'",
		},
		{
			name:          "invalid PURL - incomplete",
			purl:          "pkg:",
			errorContains: "invalid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePurl(tt.purl)

			if tt.errorContains != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetBomVersion(t *testing.T) {
	tests := []struct {
		name            string
		bomContent      string
		expectedVersion string
	}{
		{
			name: "CycloneDX 1.4",
			bomContent: `<?xml version="1.0"?>
<bom xmlns="http://cyclonedx.org/schema/bom/1.4" version="1">
	<metadata>
		<component>
			<name>test</name>
			<purl>pkg:maven/com.example/test@1.0.0</purl>
		</component>
	</metadata>
</bom>`,
			expectedVersion: "1.4",
		},
		{
			name: "CycloneDX 1.5",
			bomContent: `<?xml version="1.0"?>
<bom xmlns="http://cyclonedx.org/schema/bom/1.5" version="1">
	<metadata>
		<component>
			<name>test</name>
			<purl>pkg:maven/com.example/test@1.0.0</purl>
		</component>
	</metadata>
</bom>`,
			expectedVersion: "1.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileName, cleanup := createTempFile(t, tt.bomContent)
			defer cleanup()

			version, err := GetBomVersion(fileName)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedVersion, version)
		})
	}
}

func TestParseMTASampleBOM(t *testing.T) {
	// Test parsing the actual MTA sample BOM
	bomPath := "testdata/mta-sbom-sample.xml"

	// Check if file exists first
	if _, err := os.Stat(bomPath); os.IsNotExist(err) {
		t.Skipf("Sample BOM file not found: %s", bomPath)
	}

	bom, err := GetBom(bomPath)
	assert.NoError(t, err, "Failed to parse MTA sample BOM")

	// Verify structure
	assert.Equal(t, "http://cyclonedx.org/schema/bom/1.4", bom.Xmlns, "Expected CycloneDX 1.4 xmlns")
	assert.Equal(t, "test-mta", bom.Metadata.Component.Name, "Expected component name to be 'test-mta'")
	assert.NotEmpty(t, bom.Metadata.Component.Purl, "Expected PURL to be present in metadata component")
	assert.NotEmpty(t, bom.Components, "Expected at least one component in components list")

	// Validate the BOM
	bomContent, err := os.ReadFile(bomPath)
	assert.NoError(t, err, "Failed to read BOM file")
	err = ValidateBOM(bomContent)
	assert.NoError(t, err, "MTA sample BOM validation failed")

	// Verify version detection
	version, err := GetBomVersion(bomPath)
	assert.NoError(t, err, "Failed to get BOM version")
	assert.Equal(t, "1.4", version, "Expected version 1.4")
}
