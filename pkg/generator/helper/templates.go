package helper

import (
	"embed"
	"fmt"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

func getStepGoTemplate() string {
	return mustReadTemplate("templates/step.go.tmpl")
}

func mustReadTemplate(path string) string {
	content, err := templateFS.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("failed to read embedded template %s: %v", path, err))
	}
	return string(content)
}
