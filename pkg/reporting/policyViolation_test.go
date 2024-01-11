//go:build unit
// +build unit

package reporting

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPolicyViolationToMarkdown(t *testing.T) {
	t.Parallel()
	t.Run("success - empty", func(t *testing.T) {
		t.Parallel()
		policyReport := PolicyViolationReport{}
		_, err := policyReport.ToMarkdown()
		assert.NoError(t, err)
	})

	t.Run("success - filled", func(t *testing.T) {
		t.Parallel()
		policyReport := PolicyViolationReport{
			ArtifactID:       "theArtifact",
			Branch:           "main",
			CommitID:         "acb123",
			Description:      "This is the test description.",
			DirectDependency: "true",
			Footer:           "This is the test footer",
			Group:            "the.group",
			PipelineName:     "thePipelineName",
			PipelineLink:     "https://the.link.to.the.pipeline",
			Version:          "1.2.3",
			PackageURL:       "pkg:generic/the.group/theArtifact@1.2.3",
		}
		goldenFilePath := filepath.Join("testdata", "markdownPolicyViolation.golden")
		expected, err := os.ReadFile(goldenFilePath)
		assert.NoError(t, err)

		res, err := policyReport.ToMarkdown()
		assert.NoError(t, err)
		assert.Equal(t, string(expected), string(res))
	})
}
