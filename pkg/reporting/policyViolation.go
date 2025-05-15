package reporting

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/SAP/jenkins-library/pkg/orchestrator"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type PolicyViolationReport struct {
	ArtifactID       string
	Branch           string
	CommitID         string
	Description      string
	DirectDependency string
	Footer           string
	Group            string
	PackageURL       string
	PipelineName     string
	PipelineLink     string
	Version          string
}

const policyViolationMdTemplate string = `# Policy Violation - {{ .PackageURL }}

## Description

{{ .Description }}

## Context

{{if .PipelineLink -}}
### Pipeline

Pipeline run: [{{ .PipelineName }}]({{ .PipelineLink }})
{{- end}}

### Detected in

{{if .Branch}}**Branch:** {{ .Branch }}{{- end}}
{{if .CommitID}}**CommitId:** {{ .CommitID }}{{- end}}
{{if .DirectDependency}}**Dependency:** {{if (eq .DirectDependency "true")}}direct{{ else }}indirect{{ end }}{{- end}}
{{if .ArtifactID}}**ArtifactId:** {{ .ArtifactID }}{{- end}}
{{if .Group}}**Group:** {{ .Group }}{{- end}}
{{if .Version}}**Version:** {{ .Version }}{{- end}}
{{if .PackageURL}}**Package URL:** {{ .PackageURL }}{{- end}}

---

{{.Footer}}
`

func (p *PolicyViolationReport) ToMarkdown() ([]byte, error) {
	funcMap := template.FuncMap{
		"date": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
		"title": func(s string) string {
			caser := cases.Title(language.AmericanEnglish)
			return caser.String(s)
		},
	}

	// only fill with orchestrator information if orchestrator can be identified properly
	if provider, err := orchestrator.GetOrchestratorConfigProvider(nil); err == nil {
		// only add information if not yet provided
		if len(p.CommitID) == 0 {
			p.CommitID = provider.CommitSHA()
		}
		if len(p.PipelineLink) == 0 {
			p.PipelineLink = provider.JobURL()
			p.PipelineName = provider.JobName()
		}
	}

	md := []byte{}
	tmpl, err := template.New("report").Funcs(funcMap).Parse(policyViolationMdTemplate)
	if err != nil {
		return md, fmt.Errorf("failed to create  markdown issue template: %w", err)
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, p)
	if err != nil {
		return md, fmt.Errorf("failed to execute markdown issue template: %w", err)
	}
	md = buf.Bytes()
	return md, nil
}
