package config

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// PiperEnvironmentParameter defines a parameter within the Piper environment
type PiperEnvironmentParameter struct {
	Name string `json:"name"`
}

// InfluxResource defines an Influx resouece that holds measurement information for a pipeline run
type InfluxResource struct {
	Name         string              `json:"name"`
	StepName     string              `json:"stepName"`
	Measurements []InfluxMeasurement `json:"parameters"`
}

// InfluxMeasurement defines a measurement for Influx reporting which is defined via a step resource
type InfluxMeasurement struct {
	Name   string         `json:"name"`
	Fields []InfluxMetric `json:"fields"`
	Tags   []InfluxMetric `json:"tags"`
}

// InfluxMetric defines a metric (column) in an influx measurement
type InfluxMetric struct {
	Name string `json:"name"`
}

// InfluxMetricContent defines the content of an Inflx metric
type InfluxMetricContent struct {
	Measurement string
	ValType     string
	Name        string
	Value       *string
}

// InfluxField is the constant for an Inflx field
const InfluxField = "field"

// InfluxTag is the constant for an Inflx field
const InfluxTag = "tag"

const influxStructTemplate = `type {{ .StepName }}{{ .Name | title}} struct {
	{{- range $notused, $measurement := .Measurements }}
	{{ $measurement.Name }} struct {
		fields struct {
			{{- range $notused, $field := $measurement.Fields }}
			{{ $field.Name }} string
			{{- end }}
		}
		tags struct {
			{{- range $notused, $tag := $measurement.Tags }}
			{{ $tag.Name }} string
			{{- end }}
		}
	}
	{{- end }}
}

func (i *{{ .StepName }}{{ .Name | title}}) persist(path, resourceName string) {
	measurementContent := []config.InfluxMetricContent{
		{{- range $notused, $measurement := .Measurements }}
		{{- range $notused, $field := $measurement.Fields }}
		{ValType: config.InfluxField, Measurement: "{{ $measurement.Name }}" , Name: "{{ $field.Name }}", Value: &i.{{ $measurement.Name }}.fields.{{ $field.Name }}},
		{{- end }}
		{{- range $notused, $tag := $measurement.Tags }}
		{ValType: config.InfluxTag, Measurement: "{{ $measurement.Name }}" , Name: "{{  $tag.Name }}", Value: &i.{{ $measurement.Name }}.tags.{{  $tag.Name }}},
		{{- end }}
		{{- end }}
	}

	errCount := 0
	for _, metric := range measurementContent {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(metric.Measurement, fmt.Sprintf("%vs", metric.ValType), metric.Name), *metric.Value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting influx environment.")
			errCount++
		}
	}
	if errCount > 0 {
		os.Exit(1)
	}
}`

// StructString returns the golang coding for the struct definition of the InfluxResource
func (i *InfluxResource) StructString() (string, error) {

	funcMap := template.FuncMap{
		"title": strings.Title,
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
