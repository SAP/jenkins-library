package helper

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// PiperEnvironmentResource defines a piper environement resource which stores data across multiple pipeline steps
type PiperEnvironmentResource struct {
	Name       string
	StepName   string
	Parameters []PiperEnvironmentParameter
	Categories []string
}

// PiperEnvironmentParameter defines a parameter within the Piper environment
type PiperEnvironmentParameter struct {
	Category string
	Name     string
}

const piperEnvStructTemplate = `type {{ .StepName }}{{ .Name | title}} struct {
	{{- range $notused, $param := .Parameters }}
	{{- if not $param.Category}}
	{{ $param.Name | golangName }} string
	{{- end }}
	{{- end }}
	{{- range $notused, $category := .Categories }}
	{{ $category }} struct {
		{{- range $notused, $param := $.Parameters }}
		{{- if eq $category $param.Category }}
		{{ $param.Name | golangName }} string
		{{- end }}
		{{- end }}
	}
	{{- end }}
}

func (p *{{ .StepName }}{{ .Name | title}}) persist(path, resourceName string) {
	content := []struct{
		category string
		name string
		value interface{}
	}{
		{{- range $notused, $param := .Parameters }}
		{{- if not $param.Category}}
		{category: "", name: "{{ $param.Name }}", value: p.{{ $param.Name | golangName}}},
		{{- else }}
		{category: "{{ $param.Category }}", name: "{{ $param.Name }}", value: p.{{ $param.Category }}.{{ $param.Name | golangName}}},
		{{- end }}
		{{- end }}
	}

	errCount := 0
	for _, param := range content {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(param.category, param.name), param.value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting piper environment.")
			errCount++
		}
	}
	if errCount > 0 {
		log.Entry().Fatal("failed to persist Piper environment")
	}
}`

// StructName returns the name of the influx resource struct
func (p *PiperEnvironmentResource) StructName() string {
	return fmt.Sprintf("%v%v", p.StepName, strings.Title(p.Name))
}

// StructString returns the golang coding for the struct definition of the InfluxResource
func (p *PiperEnvironmentResource) StructString() (string, error) {
	funcMap := template.FuncMap{
		"title":      strings.Title,
		"golangName": golangName,
	}

	tmpl, err := template.New("resources").Funcs(funcMap).Parse(piperEnvStructTemplate)
	if err != nil {
		return "", err
	}

	var generatedCode bytes.Buffer
	err = tmpl.Execute(&generatedCode, &p)
	if err != nil {
		return "", err
	}

	return string(generatedCode.Bytes()), nil
}

// InfluxResource defines an Influx resouece that holds measurement information for a pipeline run
type InfluxResource struct {
	Name         string
	StepName     string
	Measurements []InfluxMeasurement
}

// InfluxMeasurement defines a measurement for Influx reporting which is defined via a step resource
type InfluxMeasurement struct {
	Name   string
	Fields []InfluxMetric
	Tags   []InfluxMetric
}

// InfluxMetric defines a metric (column) in an influx measurement
type InfluxMetric struct {
	Name string
	Type string
}

// InfluxMetricContent defines the content of an Inflx metric
type InfluxMetricContent struct {
	Measurement string
	ValType     string
	Name        string
	Value       *string
}

const influxStructTemplate = `type {{ .StepName }}{{ .Name | title}} struct {
	{{- range $notused, $measurement := .Measurements }}
	{{ $measurement.Name }} struct {
		fields struct {
			{{- range $notused, $field := $measurement.Fields }}
			{{ $field.Name | golangName }} {{ $field.Type | influxType }}
			{{- end }}
		}
		tags struct {
			{{- range $notused, $tag := $measurement.Tags }}
			{{ $tag.Name | golangName }} {{ $tag.Type | influxType }}
			{{- end }}
		}
	}
	{{- end }}
}

func (i *{{ .StepName }}{{ .Name | title}}) persist(path, resourceName string) {
	measurementContent := []struct{
		measurement string
		valType     string
		name        string
		value       interface{}
	}{
		{{- range $notused, $measurement := .Measurements }}
		{{- range $notused, $field := $measurement.Fields }}
		{valType: config.InfluxField, measurement: "{{ $measurement.Name }}" , name: "{{ $field.Name }}", value: i.{{ $measurement.Name }}.fields.{{ $field.Name | golangName }}},
		{{- end }}
		{{- range $notused, $tag := $measurement.Tags }}
		{valType: config.InfluxTag, measurement: "{{ $measurement.Name }}" , name: "{{  $tag.Name }}", value: i.{{ $measurement.Name }}.tags.{{  $tag.Name | golangName }}},
		{{- end }}
		{{- end }}
	}

	errCount := 0
	for _, metric := range measurementContent {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(metric.measurement, fmt.Sprintf("%vs", metric.valType), metric.name), metric.value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting influx environment.")
			errCount++
		}
	}
	if errCount > 0 {
		log.Entry().Fatal("failed to persist Influx environment")
	}
}`

// StructString returns the golang coding for the struct definition of the InfluxResource
func (i *InfluxResource) StructString() (string, error) {
	funcMap := template.FuncMap{
		"title":      strings.Title,
		"golangName": golangName,
		"influxType": influxType,
	}

	tmpl, err := template.New("resources").Funcs(funcMap).Parse(influxStructTemplate)
	if err != nil {
		return "", err
	}

	var generatedCode bytes.Buffer
	err = tmpl.Execute(&generatedCode, &i)
	if err != nil {
		return "", err
	}

	return string(generatedCode.Bytes()), nil
}

// StructName returns the name of the influx resource struct
func (i *InfluxResource) StructName() string {
	return fmt.Sprintf("%v%v", i.StepName, strings.Title(i.Name))
}
