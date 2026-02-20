package helper

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

type stepInfo struct {
	CobraCmdFuncName  string
	CreateCmdVar      string
	ExportPrefix      string
	FlagsFunc         string
	Long              string
	StepParameters    []config.StepParameters
	StepAliases       []config.Alias
	OutputResources   []OutputResource
	Short             string
	StepFunc          string
	StepName          string
	StepSecrets       []string
	Containers        []config.Container
	Sidecars          []config.Container
	Outputs           config.StepOutputs
	Resources         []config.StepResources
	Secrets           []config.StepSecrets
	StepErrors        []config.StepError
	HasReportsOutput  bool
	HasInfluxOutput   bool
	HasPiperEnvOutput bool
}

// OutputResource represents a generated output resource (piperEnvironment, influx, reports).
type OutputResource struct {
	Name       string // Variable name in generated code
	Type       string // "piperEnvironment", "influx", or "reports"
	Def        string // Struct definition code
	ObjectName string // Type name of the struct
}

// StepTestGoTemplate ...
const stepTestGoTemplate = `//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test{{.CobraCmdFuncName}}(t *testing.T) {
	t.Parallel()

	testCmd := {{.CobraCmdFuncName}}()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, {{ .StepName | quote }}, testCmd.Use, "command name incorrect")

}
`

const stepGoImplementationTemplate = `package cmd
import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

type {{.StepName}}Utils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The {{.StepName}}Utils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type {{.StepName}}UtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to {{.StepName}}Utils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// {{.StepName}}UtilsBundle and forward to the implementation of the dependency.
}

func new{{.StepName | title}}Utils() {{.StepName}}Utils {
	utils := {{.StepName}}UtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func {{.StepName}}(config {{ .StepName }}Options, telemetryData *telemetry.CustomData{{ range $notused, $oRes := .OutputResources}}, {{ index $oRes "name" }} *{{ index $oRes "objectname" }}{{ end }}) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := new{{.StepName | title}}Utils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := run{{.StepName | title}}(&config, telemetryData, utils{{ range $notused, $oRes := .OutputResources}}, {{ index $oRes "name" }}{{ end }})
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func run{{.StepName | title}}(config *{{ .StepName }}Options, telemetryData *telemetry.CustomData, utils {{.StepName}}Utils{{ range $notused, $oRes := .OutputResources}}, {{ index $oRes "name" }} *{{ index $oRes "objectname" }} {{ end }}) error {
	log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")

	// Example of calling methods from external dependencies directly on utils:
	exists, err := utils.FileExists("file.txt")
	if err != nil {
		// It is good practice to set an error category.
		// Most likely you want to do this at the place where enough context is known.
		log.SetErrorCategory(log.ErrorConfiguration)
		// Always wrap non-descriptive errors to enrich them with context for when they appear in the log:
		return fmt.Errorf("failed to check for important file: %w", err)
	}
	if !exists {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("cannot run without important file")
	}

	return nil
}
`

const stepGoImplementationTestTemplate = `package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type {{.StepName}}MockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func new{{.StepName | title}}TestsUtils() {{.StepName}}MockUtils {
	utils := {{.StepName}}MockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRun{{.StepName | title}}(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := {{.StepName}}Options{}

		utils := new{{.StepName | title}}TestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := run{{.StepName | title}}(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := {{.StepName}}Options{}

		utils := new{{.StepName | title}}TestsUtils()

		// test
		err := run{{.StepName | title}}(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
`

const metadataGeneratedFileName = "metadata_generated.go"
const metadataGeneratedTemplate = `// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import "github.com/SAP/jenkins-library/pkg/config"

// GetStepMetadata return a map with all the step metadata mapped to their stepName
func GetAllStepMetadata() map[string]config.StepData {
	return map[string]config.StepData{
		{{range $stepName := .Steps }} {{ $stepName | quote }}: {{$stepName}}Metadata(),
		{{end}}
	}
}
`

// ProcessMetaFiles generates step coding based on step configuration provided in yaml files
func ProcessMetaFiles(metadataFiles []string, targetDir string, stepHelperData StepHelperData) error {
	allSteps := struct{ Steps []string }{}
	for key := range metadataFiles {
		var stepData config.StepData

		configFilePath := metadataFiles[key]

		metadataFile, err := stepHelperData.OpenFile(configFilePath)
		if err != nil {
			log.Fatalf("Error occurred: %v\n", err)
		}
		defer metadataFile.Close()

		fmt.Printf("Reading file %v\n", configFilePath)

		if err = stepData.ReadPipelineStepData(metadataFile); err != nil {
			log.Fatalf("Error occurred: %v\n", err)
		}

		stepName := stepData.Metadata.Name
		fmt.Println("Step name: ", stepName)
		if stepName+".yaml" != filepath.Base(configFilePath) {
			log.Fatalf("Expected file %s to have name %s.yaml (<stepName>.yaml)\n", configFilePath, filepath.Join(filepath.Dir(configFilePath), stepName))
		}
		allSteps.Steps = append(allSteps.Steps, stepName)

		for _, parameter := range stepData.Spec.Inputs.Parameters {
			for _, mandatoryIfCase := range parameter.MandatoryIf {
				if mandatoryIfCase.Name == "" || mandatoryIfCase.Value == "" {
					return errors.New("invalid mandatoryIf option")
				}
			}
		}

		if err = setDefaultParameters(&stepData); err != nil {
			log.Fatalf("Error occurred: %v\n", err)
		}

		myStepInfo, err := getStepInfo(&stepData, stepHelperData.ExportPrefix)
		if err != nil {
			log.Fatalf("Error occurred: %v\n", err)
		}

		step := stepTemplate(myStepInfo, "step", getStepGoTemplate())
		if err = stepHelperData.WriteFile(filepath.Join(targetDir, fmt.Sprintf("%v_generated.go", stepName)), step, 0644); err != nil {
			log.Fatalf("Error occurred: %v\n", err)
		}

		test := stepTemplate(myStepInfo, "stepTest", stepTestGoTemplate)
		if err = stepHelperData.WriteFile(filepath.Join(targetDir, fmt.Sprintf("%v_generated_test.go", stepName)), test, 0644); err != nil {
			log.Fatalf("Error occurred: %v\n", err)
		}

		exists, _ := piperutils.FileExists(filepath.Join(targetDir, fmt.Sprintf("%v.go", stepName)))
		if !exists {
			impl := stepImplementation(myStepInfo, "impl", stepGoImplementationTemplate)
			if err = stepHelperData.WriteFile(filepath.Join(targetDir, fmt.Sprintf("%v.go", stepName)), impl, 0644); err != nil {
				log.Fatalf("Error occurred: %v\n", err)
			}
		}

		exists, _ = piperutils.FileExists(filepath.Join(targetDir, fmt.Sprintf("%v_test.go", stepName)))
		if !exists {
			impl := stepImplementation(myStepInfo, "implTest", stepGoImplementationTestTemplate)
			if err = stepHelperData.WriteFile(filepath.Join(targetDir, fmt.Sprintf("%v_test.go", stepName)), impl, 0644); err != nil {
				log.Fatalf("Error occurred: %v\n", err)
			}
		}
	}

	// expose metadata functions
	code := generateCode(allSteps, "metadata", metadataGeneratedTemplate, sprig.HermeticTxtFuncMap())
	if err := stepHelperData.WriteFile(filepath.Join(targetDir, metadataGeneratedFileName), code, 0644); err != nil {
		log.Fatalf("Error occurred: %v\n", err)
	}

	return nil
}

func setDefaultParameters(stepData *config.StepData) error {
	// ToDo: custom function for default handling, support all relevant parameter types
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
			case "[]string":
				// ToDo: Check if default should be read from env
				param.Default = "[]string{}"
			case "map[string]interface{}", "[]map[string]interface{}", "map[string]any", "[]map[string]any":
				// Currently we don't need to set a default here since in this case the default
				// is never used. Needs to be changed in case we enable cli parameter handling
				// for that type.
			default:
				return fmt.Errorf("meta data type not set or not known: '%v'", param.Type)
			}
		} else {
			switch param.Type {
			case "bool":
				boolVal := "false"
				if d, ok := param.Default.(bool); ok && d {
					boolVal = "true"
				}
				param.Default = boolVal
			case "int":
				switch v := param.Default.(type) {
				case int:
					param.Default = fmt.Sprintf("%d", v)
				case int64:
					param.Default = fmt.Sprintf("%d", v)
				case float64:
					param.Default = fmt.Sprintf("%d", int(v))
				case string:
					if _, err := strconv.Atoi(v); err != nil {
						return fmt.Errorf("parameter %q: invalid int default %q", param.Name, v)
					}
					param.Default = v
				default:
					return fmt.Errorf("parameter %q: expected int, got %T", param.Name, param.Default)
				}
			case "string":
				param.Default = fmt.Sprintf("`%v`", param.Default)
			case "[]string":
				param.Default = fmt.Sprintf("[]string{`%v`}", strings.Join(getStringSliceFromInterface(param.Default), "`, `"))
			case "map[string]interface{}", "[]map[string]interface{}", "map[string]any", "[]map[string]any":
				// Currently we don't need to set a default here since in this case the default
				// is never used. Needs to be changed in case we enable cli parameter handling
				// for that type.
			default:
				return fmt.Errorf("meta data type not set or not known: '%v'", param.Type)
			}
		}

		stepData.Spec.Inputs.Parameters[k] = param
	}
	return nil
}

func getStepInfo(stepData *config.StepData, exportPrefix string) (stepInfo, error) {
	oRes, err := getOutputResourceDetails(stepData)

	// Pre-compute output resource type flags for template
	var hasReports, hasInflux, hasPiperEnv bool
	for _, res := range oRes {
		switch res.Type {
		case "reports":
			hasReports = true
		case "influx":
			hasInflux = true
		case "piperEnvironment":
			hasPiperEnv = true
		}
	}

	return stepInfo{
			StepName:          stepData.Metadata.Name,
			CobraCmdFuncName:  fmt.Sprintf("%vCommand", piperutils.Title(stepData.Metadata.Name)),
			CreateCmdVar:      fmt.Sprintf("create%vCmd", piperutils.Title(stepData.Metadata.Name)),
			Short:             stepData.Metadata.Description,
			Long:              stepData.Metadata.LongDescription,
			StepParameters:    stepData.Spec.Inputs.Parameters,
			StepAliases:       stepData.Metadata.Aliases,
			FlagsFunc:         fmt.Sprintf("add%vFlags", piperutils.Title(stepData.Metadata.Name)),
			OutputResources:   oRes,
			HasReportsOutput:  hasReports,
			HasInfluxOutput:   hasInflux,
			HasPiperEnvOutput: hasPiperEnv,
			ExportPrefix:      exportPrefix,
			StepSecrets:       getSecretFields(stepData),
			Containers:        stepData.Spec.Containers,
			Sidecars:          stepData.Spec.Sidecars,
			Outputs:           stepData.Spec.Outputs,
			Resources:         stepData.Spec.Inputs.Resources,
			Secrets:           stepData.Spec.Inputs.Secrets,
			StepErrors:        stepData.Metadata.Errors,
		},
		err
}

func getSecretFields(stepData *config.StepData) []string {
	var secretFields []string

	for _, parameter := range stepData.Spec.Inputs.Parameters {
		if parameter.Secret {
			secretFields = append(secretFields, parameter.Name)
		}
	}
	return secretFields
}

func getOutputResourceDetails(stepData *config.StepData) ([]OutputResource, error) {
	var outputResources []OutputResource

	for _, res := range stepData.Spec.Outputs.Resources {
		currentResource := OutputResource{
			Name: res.Name,
			Type: res.Type,
		}

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
					if !slices.Contains(envResource.Categories, category) {
						envResource.Categories = append(envResource.Categories, category)
					}
				}
				envParam := PiperEnvironmentParameter{Category: category, Name: name, Type: fmt.Sprint(param["type"])}
				envResource.Parameters = append(envResource.Parameters, envParam)
			}
			def, err := envResource.StructString()
			if err != nil {
				return outputResources, err
			}
			currentResource.Def = def
			currentResource.ObjectName = envResource.StructName()
			outputResources = append(outputResources, currentResource)
		case "influx":
			var influxResource InfluxResource
			influxResource.Name = res.Name
			influxResource.StepName = stepData.Metadata.Name
			for _, measurement := range res.Parameters {
				influxMeasurement := InfluxMeasurement{Name: fmt.Sprintf("%v", measurement["name"])}
				if fields, ok := measurement["fields"].([]any); ok {
					for _, field := range fields {
						if fieldParams, ok := field.(map[string]any); ok {
							influxMeasurement.Fields = append(influxMeasurement.Fields, InfluxMetric{Name: fmt.Sprintf("%v", fieldParams["name"]), Type: fmt.Sprintf("%v", fieldParams["type"])})
						}
					}
				}

				if tags, ok := measurement["tags"].([]any); ok {
					for _, tag := range tags {
						if tagParams, ok := tag.(map[string]any); ok {
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
			currentResource.Def = def
			currentResource.ObjectName = influxResource.StructName()
			outputResources = append(outputResources, currentResource)
		case "reports":
			var reportsResource ReportsResource
			reportsResource.Name = res.Name
			reportsResource.StepName = stepData.Metadata.Name
			for _, param := range res.Parameters {
				filePattern, _ := param["filePattern"].(string)
				paramRef, _ := param["paramRef"].(string)
				if filePattern == "" && paramRef == "" {
					return outputResources, errors.New("both filePattern and paramRef cannot be empty at the same time")
				}
				stepResultType, _ := param["type"].(string)
				reportsParam := ReportsParameter{FilePattern: filePattern, ParamRef: paramRef, Type: stepResultType}
				reportsResource.Parameters = append(reportsResource.Parameters, reportsParam)
			}
			def, err := reportsResource.StructString()
			if err != nil {
				return outputResources, err
			}
			currentResource.Def = def
			currentResource.ObjectName = reportsResource.StructName()
			outputResources = append(outputResources, currentResource)
		}
	}

	return outputResources, nil
}

// MetadataFiles provides a list of all step metadata files
func MetadataFiles(sourceDirectory string) ([]string, error) {
	var metadataFiles []string

	if err := filepath.Walk(sourceDirectory, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".yaml" {
			metadataFiles = append(metadataFiles, path)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return metadataFiles, nil
}

func isCLIParam(myType string) bool {
	return myType != "map[string]interface{}" && myType != "[]map[string]interface{}"
}

func stepTemplate(myStepInfo stepInfo, templateName, goTemplate string) []byte {
	funcMap := sprig.HermeticTxtFuncMap()
	funcMap["flagType"] = flagType
	funcMap["golangName"] = GolangNameTitle
	funcMap["title"] = piperutils.Title
	funcMap["longName"] = longName
	funcMap["uniqueName"] = mustUniqName
	funcMap["isCLIParam"] = isCLIParam
	funcMap["configPrefix"] = configPrefix
	funcMap["structTag"] = structTag

	return generateCode(myStepInfo, templateName, goTemplate, funcMap)
}

// structTag generates the struct field tag for a step parameter.
// Example output: `json:"paramName,omitempty" validate:"possible-values=a b c"`
func structTag(param config.StepParameters) string {
	tag := fmt.Sprintf(`json:"%s,omitempty"`, param.Name)

	var validators []string
	if len(param.PossibleValues) > 0 {
		var values []string
		for _, v := range param.PossibleValues {
			values = append(values, fmt.Sprintf("%v", v))
		}
		validators = append(validators, "possible-values="+strings.Join(values, " "))
	}
	if len(param.MandatoryIf) > 0 {
		var conditions []string
		for _, m := range param.MandatoryIf {
			conditions = append(conditions, piperutils.Title(m.Name)+" "+m.Value)
		}
		validators = append(validators, "required_if="+strings.Join(conditions, " "))
	}
	if len(validators) > 0 {
		tag += fmt.Sprintf(` validate:"%s"`, strings.Join(validators, ","))
	}

	return "`" + tag + "`"
}

// configPrefix returns "prefix." if prefix is non-empty, otherwise empty string.
// Used in templates to simplify repeated {{if .ExportPrefix}}{{.ExportPrefix}}.{{end}} patterns.
func configPrefix(prefix string) string {
	if prefix != "" {
		return prefix + "."
	}
	return ""
}

func stepImplementation(myStepInfo stepInfo, templateName, goTemplate string) []byte {
	funcMap := sprig.HermeticTxtFuncMap()
	funcMap["title"] = piperutils.Title
	funcMap["uniqueName"] = mustUniqName

	return generateCode(myStepInfo, templateName, goTemplate, funcMap)
}

func generateCode(dataObject any, templateName, goTemplate string, funcMap template.FuncMap) []byte {
	tmpl, err := template.New(templateName).Funcs(funcMap).Parse(goTemplate)
	if err != nil {
		log.Fatalf("Error occurred: %v\n", err)
	}

	var generatedCode bytes.Buffer
	if err = tmpl.Execute(&generatedCode, dataObject); err != nil {
		log.Fatalf("Error occurred: %v\n", err)
	}

	return generatedCode.Bytes()
}

func longName(long string) string {
	l := strings.ReplaceAll(long, "`", "` + \"`\" + `")
	l = strings.TrimSpace(l)
	return l
}

func resourceFieldType(fieldType string) string {
	// TODO: clarify why fields are initialized with <nil> and tags are initialized with ''
	if len(fieldType) == 0 || fieldType == "<nil>" {
		return "string"
	}
	return fieldType
}

func golangName(name string) string {
	name = strings.ReplaceAll(name, "Api", "API")
	name = strings.ReplaceAll(name, "api", "API")
	name = strings.ReplaceAll(name, "Url", "URL")
	name = strings.ReplaceAll(name, "Id", "ID")
	name = strings.ReplaceAll(name, "Json", "JSON")
	name = strings.ReplaceAll(name, "json", "JSON")
	name = strings.ReplaceAll(name, "Tls", "TLS")
	return name
}

// GolangNameTitle returns name in title case with abbriviations in capital (API, URL, ID, JSON, TLS)
func GolangNameTitle(name string) string {
	return piperutils.Title(golangName(name))
}

func flagType(paramType string) string {
	switch paramType {
	case "bool":
		return "BoolVar"
	case "int":
		return "IntVar"
	case "string":
		return "StringVar"
	case "[]string":
		return "StringSliceVar"
	default: // TODO: Should it be fatal or just log and ignore the parameter?
		log.Fatalf("Meta data type not set or not known: '%v'\n", paramType)
	}
	return ""
}

func getStringSliceFromInterface(iSlice any) []string {
	t, ok := iSlice.([]any)
	if !ok {
		return []string{fmt.Sprintf("%v", iSlice)}
	}
	s := []string{}
	for _, v := range t {
		s = append(s, fmt.Sprintf("%v", v))
	}

	return s
}

func mustUniqName(list []config.StepParameters) ([]config.StepParameters, error) {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		names := []string{}
		dest := []config.StepParameters{}
		var item config.StepParameters
		for i := 0; i < l; i++ {
			item = l2.Index(i).Interface().(config.StepParameters)
			if !slices.Contains(names, item.Name) {
				names = append(names, item.Name)
				dest = append(dest, item)
			}
		}

		return dest, nil
	default:
		return nil, fmt.Errorf("Cannot find uniq on type %s", tp)
	}
}
