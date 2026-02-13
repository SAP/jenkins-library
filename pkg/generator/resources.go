package generator

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/SAP/jenkins-library/pkg/piperutils"
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
	Type     string
}

// StructName returns the name of the environment resource struct
func (p *PiperEnvironmentResource) StructName() string {
	return fmt.Sprintf("%v%v", p.StepName, piperutils.Title(p.Name))
}

// StructString returns the golang coding for the struct definition of the environment resource
func (p *PiperEnvironmentResource) StructString() (string, error) {
	funcMap := template.FuncMap{
		"title":             piperutils.Title,
		"golangName":        golangName,
		"resourceFieldType": resourceFieldType,
	}

	templateContent, err := TemplateFS.ReadFile("templates/piper_env.go.tmpl")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("piper_env.go.tmpl").Funcs(funcMap).Parse(string(templateContent))
	if err != nil {
		return "", err
	}

	var generatedCode bytes.Buffer
	err = tmpl.Execute(&generatedCode, &p)
	if err != nil {
		return "", err
	}

	return generatedCode.String(), nil
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

// StructString returns the golang coding for the struct definition of the InfluxResource
func (i *InfluxResource) StructString() (string, error) {
	funcMap := template.FuncMap{
		"title":             piperutils.Title,
		"golangName":        golangName,
		"resourceFieldType": resourceFieldType,
	}

	templateContent, err := TemplateFS.ReadFile("templates/influx.go.tmpl")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("influx.go.tmpl").Funcs(funcMap).Parse(string(templateContent))
	if err != nil {
		return "", err
	}

	var generatedCode bytes.Buffer
	err = tmpl.Execute(&generatedCode, &i)
	if err != nil {
		return "", err
	}

	return generatedCode.String(), nil
}

// StructName returns the name of the influx resource struct
func (i *InfluxResource) StructName() string {
	return fmt.Sprintf("%v%v", i.StepName, piperutils.Title(i.Name))
}

// ReportsResource defines a piper environement resource which stores data across multiple pipeline steps
type ReportsResource struct {
	Name       string
	StepName   string
	Parameters []ReportsParameter
}

// ReportsParameter defines a parameter within the Piper environment
type ReportsParameter struct {
	FilePattern string
	ParamRef    string
	Type        string
}

// StructName returns the name of the environment resource struct
func (p *ReportsResource) StructName() string {
	return fmt.Sprintf("%v%v", p.StepName, piperutils.Title(p.Name))
}

// StructString returns the golang coding for the struct definition of the environment resource
func (p *ReportsResource) StructString() (string, error) {
	funcMap := template.FuncMap{
		"title":             piperutils.Title,
		"golangName":        golangName,
		"resourceFieldType": resourceFieldType,
	}

	templateContent, err := TemplateFS.ReadFile("templates/reports.go.tmpl")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("reports.go.tmpl").Funcs(funcMap).Parse(string(templateContent))
	if err != nil {
		return "", err
	}

	var generatedCode bytes.Buffer
	if err = tmpl.Execute(&generatedCode, &p); err != nil {
		return "", err
	}

	return generatedCode.String(), nil
}
