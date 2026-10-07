//go:build unit
// +build unit

package format

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeURI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "leading tab and surrounding spaces", input: "\t python-hyper/h2 ", expected: "python-hyper/h2"},
		{name: "newline and carriage return", input: "foo\r\nbar", expected: "foobar"},
		{name: "no change needed", input: "pkg:pypi/h2@3.1.0", expected: "pkg:pypi/h2@3.1.0"},
		{name: "empty", input: "", expected: ""},
		{name: "only control characters", input: "\t\n\r", expected: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, sanitizeURI(tt.input))
		})
	}
}

func TestSARIFSanitize(t *testing.T) {
	t.Run("cleans all location URIs", func(t *testing.T) {
		sarif := &SARIF{
			Runs: []Runs{
				{
					Results: []Results{
						{
							AnalysisTarget: &ArtifactLocation{URI: "\t pkg:pypi/h2 "},
							Locations: []Location{
								{PhysicalLocation: PhysicalLocation{ArtifactLocation: ArtifactLocation{URI: "\t python-hyper/h2 "}}},
							},
							RelatedLocations: []RelatedLocation{
								{PhysicalLocation: RelatedPhysicalLocation{ArtifactLocation: ArtifactLocation{URI: "related\tname"}}},
							},
						},
					},
					Artifacts: []Artifact{
						{Location: SarifLocation{Uri: " artifact\nname "}},
					},
				},
			},
		}

		sarif.Sanitize()

		assert.Equal(t, "pkg:pypi/h2", sarif.Runs[0].Results[0].AnalysisTarget.URI)
		assert.Equal(t, "python-hyper/h2", sarif.Runs[0].Results[0].Locations[0].PhysicalLocation.ArtifactLocation.URI)
		assert.Equal(t, "relatedname", sarif.Runs[0].Results[0].RelatedLocations[0].PhysicalLocation.ArtifactLocation.URI)
		assert.Equal(t, "artifactname", sarif.Runs[0].Artifacts[0].Location.Uri)
	})

	t.Run("nil receiver is a no-op", func(t *testing.T) {
		var sarif *SARIF
		assert.NotPanics(t, func() { sarif.Sanitize() })
	})
}
