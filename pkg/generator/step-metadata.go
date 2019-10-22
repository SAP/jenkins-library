package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/SAP/jenkins-library/pkg/config"
)

/*
//StepParameterDefinition defines one configuration parameter of a step
type StepParameterDefinition struct {
	Short     string      `json:"description"`
	Long      string      `json:"longDescription"`
	Markdown  string      `json:"markdownDescription"`
	Mandatory bool        `json:"mandatory"`
	Scope     []string    `json:"scope"`
	Type      string      `json:"type"`
	Default   interface{} `json:"default"`
}

//StepMetaData defines configuration options of the step
type StepMetaData struct {
	Long         string                             `json:"longDescription"`
	Short        string                             `json:"description"`
	Markdown     string                             `json:"markdownDescription"`
	OSImport bool                               `json:"osDependency,omitempty"`
	Metadata     map[string]StepParameterDefinition `json:"parameters"`
}
*/

type stepInfo struct {
	CobraCmdFuncName string
	CreateCmdVar     string
	FlagsFunc        string
	Long             string
	Metadata         []config.StepParameters
	OSImport         bool
	Short            string
	StepFunc         string
	StepName         string
}

//StepGoTemplate ...
const stepGoTemplate = `package cmd

import (
	//"os"

	"github.com/spf13/cobra"
	"github.com/SAP/jenkins-library/pkg/config"
)

type {{ .StepName }}Options struct {
	{{range $key, $value := .Metadata }}
	{{ $value.Name | golangName }} {{ $value.Type }} ` + "`json:\"{{$value.Name}},omitempty\"`" + `{{end}}
}

var my{{ .StepName | title}}Options {{.StepName}}Options
var {{ .StepName }}StepConfigJSON string

// {{.CobraCmdFuncName}} {{.Short}}
func {{.CobraCmdFuncName}}() *cobra.Command {
	var {{.CreateCmdVar}} = &cobra.Command{
		Use:   "{{.StepName}}",
		Short: "{{.Short}}",
		Long:   {{ $tick := "` + "`" + `" }}{{ $tick }}{{.Long | longName }}{{ $tick }},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			metadata := {{ .StepName }}Metadata()
			return prepareConfig(cmd, &metadata, "{{ .StepName }}", &my{{ .StepName | title}}Options)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return {{.StepName}}(my{{ .StepName | title }}Options)
		},
	}

	{{.FlagsFunc}}({{.CreateCmdVar}})
	return {{.CreateCmdVar}}
}

// {{.FlagsFunc}} defines the flags for the {{.StepName}} command
func {{.FlagsFunc}}(cmd *cobra.Command) {
	{{ range $key, $value := .Metadata }}
	cmd.Flags().{{ $value.Type | flagType }}(&my{{ $.StepName | title }}Options.{{ $value.Name | golangName }}, "{{ $value.Name }}", {{ $value.Default }}, "{{ $value.Description }}"){{ end }}

	{{ range $key, $value := .Metadata }}{{ if $value.Mandatory }}cmd.MarkFlagRequired("{{ $value.Name }}"){{ printf "\n\t" }}{{ end }}{{ end }}
}

// retrieve step metadata
func {{ .StepName }}Metadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{{range $key, $value := .Metadata }}
					{
						Name: "{{ $value.Name }}",
						Scope: []string{{ "{" }}{{ range $notused, $scope := $value.Scope }}"{{ $scope }}",{{ end }}{{ "}" }},
						Type: "{{ $value.Type }}",
						Mandatory: {{ $value.Mandatory }},
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
)

func Test{{.CobraCmdFuncName}}(t *testing.T) {

	testCmd := {{.CobraCmdFuncName}}()

	// only high level testing performed - details are tested in step generation procudure
	if testCmd.Use != "{{.StepName}}" {
		t.Errorf("Expected command name to be '{{.StepName}}' but was '%v'", testCmd.Use)
	}

}
`

func main() {

	metadataPath := "./resources/metadata"

	metadataFiles, err := metadataFiles(metadataPath)
	checkError(err)

	for key := range metadataFiles {

		var stepData config.StepData

		configFilePath := metadataFiles[key]

		metadataFile, err := os.Open(configFilePath)
		checkError(err)
		defer metadataFile.Close()

		fmt.Printf("Reading file %v\n", configFilePath)
		//configFile := filepath.Base(configFilePath)

		err = stepData.ReadPipelineStepData(metadataFile)
		checkError(err)

		fmt.Printf("Step name: %v\n", stepData.Metadata.Name)

		//ToDo: custom function for default handling, support all parameter types
		for k, param := range stepData.Spec.Inputs.Parameters {

			if param.Default == nil {
				switch param.Type {
				case "string":
					param.Default = fmt.Sprintf("os.Getenv(\"PIPER_%v\")", param.Name)
				case "bool":
					param.Default = false
					//ToDo: custom function
					//currentParameter.Default = fmt.Sprintf("strconv.ParseBool(os.Getenv(\"PIPER_%v\"))", k)
				case "[]string":
					param.Default = "[]string{}"
				default:
					fmt.Printf("Meta data type not set or not known: '%v'\n", param.Type)
					os.Exit(1)
				}
			} else {
				switch param.Type {
				case "string":
					param.Default = fmt.Sprintf("\"%v\"", param.Default)
				case "bool":
					param.Default = param.Default
				case "[]string":
					//ToDo: generate proper default by looping over value
				default:
					fmt.Printf("Meta data type not set or not known: '%v'\n", param.Type)
					os.Exit(1)
				}
			}

			stepData.Spec.Inputs.Parameters[k] = param
		}

		myStepInfo := stepInfo{
			StepName:         stepData.Metadata.Name,
			CobraCmdFuncName: fmt.Sprintf("%vCommand", strings.Title(stepData.Metadata.Name)),
			CreateCmdVar:     fmt.Sprintf("create%vCmd", strings.Title(stepData.Metadata.Name)),
			Short:            stepData.Metadata.Description,
			Long:             stepData.Metadata.LongDescription,
			Metadata:         stepData.Spec.Inputs.Parameters,
			FlagsFunc:        fmt.Sprintf("Add%vFlags", strings.Title(stepData.Metadata.Name)),
		}

		step := stepTemplate(myStepInfo)
		err = ioutil.WriteFile(fmt.Sprintf("cmd/%v_generated.go", stepData.Metadata.Name), step, 0644)
		checkError(err)

		test := stepTestTemplate(myStepInfo)
		err = ioutil.WriteFile(fmt.Sprintf("cmd/%v_generated_test.go", stepData.Metadata.Name), test, 0644)
		checkError(err)
	}

}

func checkError(err error) {
	if err != nil {
		fmt.Printf("Error occured: %v\n", err)
		os.Exit(1)
	}
}

func metadataFiles(sourceDirectory string) ([]string, error) {

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
		"golangName": golangName,
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
		"golangName": golangName,
		"title":      strings.Title,
	}

	tmpl, err := template.New("stepTest").Funcs(funcMap).Parse(stepTestGoTemplate)
	checkError(err)

	var generatedCode bytes.Buffer
	err = tmpl.Execute(&generatedCode, myStepInfo)
	checkError(err)

	return generatedCode.Bytes()
}

func longName(long string) string {
	l := strings.ReplaceAll(long, "`", "` + \"`\" + `")
	return l
}

func golangName(name string) string {
	properName := strings.Replace(name, "Api", "API", -1)
	properName = strings.Replace(properName, "Url", "URL", -1)
	properName = strings.Replace(properName, "Id", "ID", -1)
	properName = strings.Replace(properName, "Json", "JSON", -1)
	properName = strings.Replace(properName, "json", "JSON", -1)
	return strings.Title(properName)
}

func flagType(paramType string) string {
	var theFlagType string
	switch paramType {
	case "bool":
		theFlagType = "BoolVar"
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
