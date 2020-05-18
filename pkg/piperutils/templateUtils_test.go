package piperutils

import (
	"github.com/SAP/jenkins-library/pkg/versioning"
	"github.com/stretchr/testify/assert"
	"testing"
	"text/template"
)

func TestExecuteTemplate(t *testing.T) {
	t.Run("test success", func(t *testing.T) {
		context := versioning.MavenDescriptor{GroupID: "com.sap.cp.jenkins", ArtifactID: "piper", Version: "1.2.3"}
		result, err := ExecuteTemplate("{{ .GroupID }}-{{ .ArtifactID }}:{{ .Version}}", context)
		assert.NoError(t, err, "Didn't expect error but got one")
		assert.Equal(t, "com.sap.cp.jenkins-piper:1.2.3", result, "Expected different result")
	})

	t.Run("test template error", func(t *testing.T) {
		context := versioning.MavenDescriptor{GroupID: "com.sap.cp.jenkins", ArtifactID: "piper", Version: "1.2.3"}
		_, err := ExecuteTemplate("{{ $+++.+++GroupID }}-{{ .ArtifactID }}:{{ .Version}}", context)
		assert.Error(t, err, "Expected error but got none")
	})

	t.Run("test functions", func(t *testing.T) {
		functions := template.FuncMap{
			"testFunc": reverse,
		}
		context := versioning.MavenDescriptor{GroupID: "com.sap.cp.jenkins", ArtifactID: "piper", Version: "1.2.3"}
		result, err := ExecuteTemplateFunctions("{{ testFunc .GroupID }}-{{ .ArtifactID }}:{{ .Version}}", functions, context)
		assert.NoError(t, err, "Didn't expect error but got one")
		assert.Equal(t, "sniknej.pc.pas.moc-piper:1.2.3", result, "Expected different result")
	})
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
