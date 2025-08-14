package piperutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
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
					<name>com.example/mycomponent</name>
					<version>1.0.0</version>
					<purl>pkg:maven/com.example/mycomponent@1.0.0</purl>
				</component>
				<properties>
					<property name="name1" value="value1" />
					<property name="name2" value="value2" />
				</properties>
			</metadata>
		</bom>`

func TestGetBom(t *testing.T) {
	tests := []struct {
		name          string
		xmlContent    string
		expectedBom   Bom
		expectError   bool
		expectedError string
	}{
		{
			name:       "valid file",
			xmlContent: validBom,
			expectedBom: Bom{
				Metadata: Metadata{
					Component: BomComponent{
						Purl: "pkg:maven/com.example/mycomponent@1.0.0",
					},
					Properties: []BomProperty{
						{Name: "name1", Value: "value1"},
						{Name: "name2", Value: "value2"},
					},
				},
			},
			expectError: false,
		},
		{
			name:          "file not found",
			xmlContent:    "",
			expectedBom:   Bom{},
			expectError:   true,
			expectedError: "no such file or directory",
		},
		{
			name:          "invalid XML file",
			xmlContent:    "<bom><metadata><component><purl>invalid xml</metadata></bom>",
			expectedBom:   Bom{},
			expectError:   true,
			expectedError: "XML syntax error",
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
			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
			}

			if err != nil && !tt.expectError {
				if !tt.expectError && !containsSubstring(err.Error(), tt.expectedError) {
					t.Errorf("Expected error message: %v, got: %v", tt.expectedError, err.Error())
				}
			}

			if !tt.expectError && !bomEquals(bom, tt.expectedBom) {
				t.Errorf("Expected BOM: %+v, got: %+v", tt.expectedBom, bom)
			}
		})
	}
}

func TestGetPurl(t *testing.T) {
	tests := []struct {
		name          string
		filePath      string
		bomFilename   string
		xmlContent    string
		expectedPurl  string
		expectError   bool
		expectedError string
	}{
		{
			name:         "valid BOM file",
			xmlContent:   validBom,
			expectedPurl: "pkg:maven/com.example/mycomponent@1.0.0",
		},
		{
			name:          "BOM file not found",
			xmlContent:    "",
			expectedPurl:  "",
			expectError:   true,
			expectedError: "no such file or directory",
		},
		{
			name:          "invalid BOM file",
			xmlContent:    "<bom><metadata><component><purl>invalid xml</metadata></bom>",
			expectedPurl:  "",
			expectError:   true,
			expectedError: "XML syntax error",
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

			purl := GetPurl(filePath)
			if purl != tt.expectedPurl {
				t.Errorf("Expected PURL: %v, got: %v", tt.expectedPurl, purl)
			}
		})
	}
}

func TestGetBOMName(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		bomFilename  string
		xmlContent   string
		expectedName string
	}{
		{
			name:         "valid BOM file",
			xmlContent:   validBom,
			expectedName: "com.example/mycomponent",
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

			name := GetName(filePath)
			if name != tt.expectedName {
				t.Errorf("Expected PURL: %v, got: %v", tt.expectedName, name)
			}
		})
	}
}

func TestGetBOMVersion(t *testing.T) {
	tests := []struct {
		name            string
		filePath        string
		bomFilename     string
		xmlContent      string
		expectedVersion string
	}{
		{
			name:            "valid BOM file",
			xmlContent:      validBom,
			expectedVersion: "1.0.0",
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

			version := GetVersion(filePath)
			if version != tt.expectedVersion {
				t.Errorf("Expected PURL: %v, got: %v", tt.expectedVersion, version)
			}
		})
	}
}

func TestUpdateBOMPurl(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		bomFilename  string
		xmlContent   string
		expectedPurl string
	}{
		{
			name:         "valid BOM file",
			xmlContent:   validBom,
			expectedPurl: "pkg:maven/com.example/mycomponent@1.0.0",
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

			err := UpdatePurl(filePath, "pkg:maven/com.example/mycomponent@1.0.0")
			require.NoError(t, err)
			purl := GetPurl(filePath)
			if purl != tt.expectedPurl {
				t.Errorf("Expected PURL: %v, got: %v", tt.expectedPurl, purl)
			}
		})
	}
}

func bomEquals(a, b Bom) bool {
	// compare a and b manually since reflect.DeepEqual can be problematic with slices and nil values
	return a.Metadata.Component.Purl == b.Metadata.Component.Purl &&
		len(a.Metadata.Properties) == len(b.Metadata.Properties) &&
		propertiesMatch(a.Metadata.Properties, b.Metadata.Properties)
}

func propertiesMatch(a, b []BomProperty) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func containsSubstring(str, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	return len(str) >= len(substr) && str[:len(substr)] == substr
}
