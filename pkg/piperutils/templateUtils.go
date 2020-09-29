package piperutils

import (
	"bytes"
	"fmt"
	"text/template"
)

// ExecuteTemplate parses the provided template, substitutes values and returns the output
func ExecuteTemplate(txtTemplate string, context interface{}) (string, error) {
	return ExecuteTemplateFunctions(txtTemplate, nil, context)
}

// ExecuteTemplateFunctions parses the provided template, applies the transformation functions, substitutes values and returns the output
func ExecuteTemplateFunctions(txtTemplate string, functionMap template.FuncMap, context interface{}) (string, error) {
	template := template.New("tmp")
	if functionMap != nil {
		template = template.Funcs(functionMap)
	}
	template, err := template.Parse(txtTemplate)
	if err != nil {
		return "<nil>", fmt.Errorf("Failed to parse template definition %v: %w", txtTemplate, err)
	}
	var output bytes.Buffer
	err = template.Execute(&output, context)
	if err != nil {
		return "<nil>", fmt.Errorf("Failed to transform template definition %v: %w", txtTemplate, err)
	}
	return output.String(), nil
}
