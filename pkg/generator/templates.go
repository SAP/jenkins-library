package generator

import "embed"

//go:embed templates/*.tmpl
var TemplateFS embed.FS
