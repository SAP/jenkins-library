package util

import (
	"bytes"
	"fmt"
	"text/template"
)

func ExecuteTemplate(templ string, data interface{}) (string, error) {
	tmpl, err := template.New(``).Parse(templ)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}
	description := &bytes.Buffer{}
	err = tmpl.Execute(description, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return description.String(), nil
}
