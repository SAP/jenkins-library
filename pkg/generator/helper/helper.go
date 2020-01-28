package helper

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

type stepInfo struct {
	CobraCmdFuncName string
	CreateCmdVar     string
	ExportPrefix     string
	FlagsFunc        string
	Long             string
	Metadata         []config.StepParameters
	OSImport         bool
	OutputResources  []map[string]string
	Short            string
	StepFunc         string
	StepName         string
}

//StepGoTemplate ...
const stepGoTemplate = `package cmd

import (
	{{ if .OSImport }}"os"{{ end }}
	{{ if .OutputResources }}"fmt"{{ end }}
	{{ if .OutputResources }}"path/filepath"{{ end }}

	{{ if .ExportPrefix}}{{ .ExportPrefix }} "github.com/SAP/jenkins-library/cmd"{{ end -}}
	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	{{ if .OutputResources }}"github.com/SAP/jenkins-library/pkg/piperenv"{{ end }}
	"github.com/spf13/cobra"
)

type {{ .StepName }}Options struct {
	{{- range $key, $value := .Metadata }}
	{{ $value.Name | golangName }} {{ $value.Type }} ` + "`json:\"{{$value.Name}},omitempty\"`" + `{{end}}
}

{{ range $notused, $oRes := .OutputResources }}
{{ index $oRes "def"}}
{{ end }}

var my{{ .StepName | title}}Options {{.StepName}}Options

// {{.CobraCmdFuncName}} {{.Short}}
func {{.CobraCmdFuncName}}() *cobra.Command {
	metadata := {{ .StepName }}Metadata()
	{{- range $notused, $oRes := .OutputResources }}
	var {{ index $oRes "name" }} {{ index $oRes "objectname" }}{{ end }}

	var {{.CreateCmdVar}} = &cobra.Command{
		Use:   "{{.StepName}}",
		Short: "{{.Short}}",
		Long: {{ $tick := "` + "`" + `" }}{{ $tick }}{{.Long | longName }}{{ $tick }},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			log.SetStepName("{{ .StepName }}")
			log.SetVerbose({{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}GeneralConfig.Verbose)
			return {{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}PrepareConfig(cmd, &metadata, "{{ .StepName }}", &my{{ .StepName | title}}Options, config.OpenPiperFile)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			{{ if .OutputResources -}}
			handler := func() {
				{{- range $notused, $oRes := .OutputResources }}
				{{ index $oRes "name" }}.persist(GeneralConfig.EnvRootPath, "{{ index $oRes "name" }}"){{ end }}
			}
			log.DeferExitHandler(handler)
			defer handler()
			{{- end }}
			telemetry.Initialize(GeneralConfig.NoTelemetry, "{{ .StepName }}")
			telemetry.Send(&telemetry.CustomData{})
			return {{.StepName}}(my{{ .StepName | title }}Options{{ range $notused, $oRes := .OutputResources}}, &{{ index $oRes "name" }}{{ end }})
		},
	}

	{{.FlagsFunc}}({{.CreateCmdVar}})
	return {{.CreateCmdVar}}
}

func {{.FlagsFunc}}(cmd *cobra.Command) {
	{{- range $key, $value := .Metadata }}
	cmd.Flags().{{ $value.Type | flagType }}(&my{{ $.StepName | title }}Options.{{ $value.Name | golangName }}, "{{ $value.Name }}", {{ $value.Default }}, "{{ $value.Description }}"){{ end }}
	{{- printf "\n" }}
	{{- range $key, $value := .Metadata }}{{ if $value.Mandatory }}
	cmd.MarkFlagRequired("{{ $value.Name }}"){{ end }}{{ end }}
}

// retrieve step metadata
func {{ .StepName }}Metadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{{- range $key, $value := .Metadata }}
					{
						Name:      "{{ $value.Name }}",
						ResourceRef: []config.ResourceReference{{ "{" }}{{ range $notused, $ref := $value.ResourceRef }}{{ "{" }}Name: "{{ $ref.Name }}", Param: "{{ $ref.Param }}"{{ "}" }},{{ end }}{{ "}" }},
						Scope:     []string{{ "{" }}{{ range $notused, $scope := $value.Scope }}"{{ $scope }}",{{ end }}{{ "}" }},
						Type:      "{{ $value.Type }}",
						Mandatory: {{ $value.Mandatory }},
						Aliases:   []config.Alias{{ "{" }}{{ range $notused, $alias := $value.Aliases }}{{ "{" }}Name: "{{ $alias.Name }}"{{ "}" }},{{ end }}{{ "}" }},
					},{{ end }}
				},
			},
		},
	}
	return theMetaData
}
`

//StepTestGoTemplate ...
const stepTestGoTemplate = `package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test{{.CobraCmdFuncName}}(t *testing.T) {

	testCmd := {{.CobraCmdFuncName}}()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "{{.StepName}}", testCmd.Use, "command name incorrect")

}
`

const stepGoImplementationTemplate = `package cmd
import (
	"github.com/SAP/jenkins-library/pkg/log"
)

func {{.StepName}}(config {{ .StepName }}Options{{ range $notused, $oRes := .OutputResources}}, {{ index $oRes "name" }} *{{ index $oRes "objectname" }} {{ end }}) error {
	log.Entry().WithField("customKey", "customValue").Info("This is how you write a log message with a custom field ...")
	return nil
}
`

// ProcessMetaFiles generates step coding based on step configuration provided in yaml files
func ProcessMetaFiles(metadataFiles []string, stepHelperData StepHelperData, docuHelperData DocuHelperData) error {
	for key := range metadataFiles {

		var stepData config.StepData

		configFilePath := metadataFiles[key]

		metadataFile, err := stepHelperData.OpenFile(configFilePath)
		checkError(err)
		defer metadataFile.Close()

		fmt.Printf("Reading file %v\n", configFilePath)

		err = stepData.ReadPipelineStepData(metadataFile)
		checkError(err)

		fmt.Printf("Step name: %v\n", stepData.Metadata.Name)

		//Switch Docu or Step Files
		if !docuHelperData.IsGenerateDocu {
			osImport := false
			osImport, err = setDefaultParameters(&stepData)
			checkError(err)

			myStepInfo, err := getStepInfo(&stepData, osImport, stepHelperData.ExportPrefix)
			checkError(err)

			step := stepTemplate(myStepInfo)
			err = stepHelperData.WriteFile(fmt.Sprintf("cmd/%v_generated.go", stepData.Metadata.Name), step, 0644)
			checkError(err)

			test := stepTestTemplate(myStepInfo)
			err = stepHelperData.WriteFile(fmt.Sprintf("cmd/%v_generated_test.go", stepData.Metadata.Name), test, 0644)
			checkError(err)

			exists, _ := piperutils.FileExists(fmt.Sprintf("cmd/%v.go", stepData.Metadata.Name))
			if !exists {
				impl := stepImplementation(myStepInfo)
				err = stepHelperData.WriteFile(fmt.Sprintf("cmd/%v.go", stepData.Metadata.Name), impl, 0644)
				checkError(err)
			}
		} else {
			err = generateStepDocumentation(stepData, docuHelperData)
			if err != nil {
				fmt.Printf("%v\n", err)
			}
		}
	}
	return nil
}

func openMetaFile(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

func fileWriter(filename string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}

func setDefaultParameters(stepData *config.StepData) (bool, error) {
	//ToDo: custom function for default handling, support all relevant parameter types
	osImportRequired := false
	for k, param := range stepData.Spec.Inputs.Parameters {

		if param.Default == nil {
			switch param.Type {
			case "bool":
				// ToDo: Check if default should be read from env
				param.Default = "false"
			case "int":
				param.Default = "0"
			case "string":
				param.Default = fmt.Sprintf("os.Getenv(\"PIPER_%v\")", param.Name)
				osImportRequired = true
			case "[]string":
				// ToDo: Check if default should be read from env
				param.Default = "[]string{}"
			default:
				return false, fmt.Errorf("Meta data type not set or not known: '%v'", param.Type)
			}
		} else {
			switch param.Type {
			case "bool":
				boolVal := "false"
				if param.Default.(bool) == true {
					boolVal = "true"
				}
				param.Default = boolVal
			case "int":
				param.Default = fmt.Sprintf("%v", param.Default)
			case "string":
				param.Default = fmt.Sprintf("\"%v\"", param.Default)
			case "[]string":
				param.Default = fmt.Sprintf("[]string{\"%v\"}", strings.Join(getStringSliceFromInterface(param.Default), "\", \""))
			default:
				return false, fmt.Errorf("Meta data type not set or not known: '%v'", param.Type)
			}
		}

		stepData.Spec.Inputs.Parameters[k] = param
	}
	return osImportRequired, nil
}

func getStepInfo(stepData *config.StepData, osImport bool, exportPrefix string) (stepInfo, error) {
	oRes, err := getOutputResourceDetails(stepData)

	return stepInfo{
			StepName:         stepData.Metadata.Name,
			CobraCmdFuncName: fmt.Sprintf("%vCommand", strings.Title(stepData.Metadata.Name)),
			CreateCmdVar:     fmt.Sprintf("create%vCmd", strings.Title(stepData.Metadata.Name)),
			Short:            stepData.Metadata.Description,
			Long:             stepData.Metadata.LongDescription,
			Metadata:         stepData.Spec.Inputs.Parameters,
			FlagsFunc:        fmt.Sprintf("add%vFlags", strings.Title(stepData.Metadata.Name)),
			OSImport:         osImport,
			OutputResources:  oRes,
			ExportPrefix:     exportPrefix,
		},
		err
}

func getOutputResourceDetails(stepData *config.StepData) ([]map[string]string, error) {
	outputResources := []map[string]string{}

	for _, res := range stepData.Spec.Outputs.Resources {
		currentResource := map[string]string{}
		currentResource["name"] = res.Name

		switch res.Type {
		case "piperEnvironment":
			var envResource PiperEnvironmentResource
			envResource.Name = res.Name
			envResource.StepName = stepData.Metadata.Name
			for _, param := range res.Parameters {
				paramSections := strings.Split(fmt.Sprintf("%v", param["name"]), "/")
				category := ""
				name := paramSections[0]
				if len(paramSections) > 1 {
					name = strings.Join(paramSections[1:], "_")
					category = paramSections[0]
					if !contains(envResource.Categories, category) {
						envResource.Categories = append(envResource.Categories, category)
					}
				}
				envParam := PiperEnvironmentParameter{Category: category, Name: name}
				envResource.Parameters = append(envResource.Parameters, envParam)
			}
			def, err := envResource.StructString()
			if err != nil {
				return outputResources, err
			}
			currentResource["def"] = def
			currentResource["objectname"] = envResource.StructName()
			outputResources = append(outputResources, currentResource)
		case "influx":
			var influxResource InfluxResource
			influxResource.Name = res.Name
			influxResource.StepName = stepData.Metadata.Name
			for _, measurement := range res.Parameters {
				influxMeasurement := InfluxMeasurement{Name: fmt.Sprintf("%v", measurement["name"])}
				if fields, ok := measurement["fields"].([]interface{}); ok {
					for _, field := range fields {
						if fieldParams, ok := field.(map[string]interface{}); ok {
							influxMeasurement.Fields = append(influxMeasurement.Fields, InfluxMetric{Name: fmt.Sprintf("%v", fieldParams["name"])})
						}
					}
				}

				if tags, ok := measurement["tags"].([]interface{}); ok {
					for _, tag := range tags {
						if tagParams, ok := tag.(map[string]interface{}); ok {
							influxMeasurement.Tags = append(influxMeasurement.Tags, InfluxMetric{Name: fmt.Sprintf("%v", tagParams["name"])})
						}
					}
				}
				influxResource.Measurements = append(influxResource.Measurements, influxMeasurement)
			}
			def, err := influxResource.StructString()
			if err != nil {
				return outputResources, err
			}
			currentResource["def"] = def
			currentResource["objectname"] = influxResource.StructName()
			outputResources = append(outputResources, currentResource)
		}
	}

	return outputResources, nil
}

// MetadataFiles provides a list of all step metadata files
func MetadataFiles(sourceDirectory string) ([]string, error) {

	var metadataFiles []string

	err := filepath.Walk(sourceDirectory, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".yaml" {
			metadataFiles = append(metadataFiles, path)
		}
		return nil
	})
	if err != nil {
		return metadataFiles, nil
	}
	return metadataFiles, nil
}

func stepTemplate(myStepInfo stepInfo) []byte {

	funcMap := template.FuncMap{
		"flagType":   flagType,
		"golangName": golangNameTitle,
		"title":      strings.Title,
		"longName":   longName,
	}

	tmpl, err := template.New("step").Funcs(funcMap).Parse(stepGoTemplate)
	checkError(err)

	var generatedCode bytes.Buffer
	err = tmpl.Execute(&generatedCode, myStepInfo)
	checkError(err)

	return generatedCode.Bytes()
}

func stepTestTemplate(myStepInfo stepInfo) []byte {

	funcMap := template.FuncMap{
		"flagType":   flagType,
		"golangName": golangNameTitle,
		"title":      strings.Title,
	}

	tmpl, err := template.New("stepTest").Funcs(funcMap).Parse(stepTestGoTemplate)
	checkError(err)

	var generatedCode bytes.Buffer
	err = tmpl.Execute(&generatedCode, myStepInfo)
	checkError(err)

	return generatedCode.Bytes()
}

func stepImplementation(myStepInfo stepInfo) []byte {

	funcMap := template.FuncMap{
		"title": strings.Title,
	}

	tmpl, err := template.New("impl").Funcs(funcMap).Parse(stepGoImplementationTemplate)
	checkError(err)

	var generatedCode bytes.Buffer
	err = tmpl.Execute(&generatedCode, myStepInfo)
	checkError(err)

	return generatedCode.Bytes()
}

func longName(long string) string {
	l := strings.ReplaceAll(long, "`", "` + \"`\" + `")
	l = strings.TrimSpace(l)
	return l
}

func golangName(name string) string {
	properName := strings.Replace(name, "Api", "API", -1)
	properName = strings.Replace(properName, "api", "API", -1)
	properName = strings.Replace(properName, "Url", "URL", -1)
	properName = strings.Replace(properName, "Id", "ID", -1)
	properName = strings.Replace(properName, "Json", "JSON", -1)
	properName = strings.Replace(properName, "json", "JSON", -1)
	return properName
}

func golangNameTitle(name string) string {
	return strings.Title(golangName(name))
}

func flagType(paramType string) string {
	var theFlagType string
	switch paramType {
	case "bool":
		theFlagType = "BoolVar"
	case "int":
		theFlagType = "IntVar"
	case "string":
		theFlagType = "StringVar"
	case "[]string":
		theFlagType = "StringSliceVar"
	default:
		fmt.Printf("Meta data type not set or not known: '%v'\n", paramType)
		os.Exit(1)
	}
	return theFlagType
}

func getStringSliceFromInterface(iSlice interface{}) []string {
	s := []string{}

	t, ok := iSlice.([]interface{})
	if ok {
		for _, v := range t {
			s = append(s, fmt.Sprintf("%v", v))
		}
	} else {
		s = append(s, fmt.Sprintf("%v", iSlice))
	}

	return s
}
