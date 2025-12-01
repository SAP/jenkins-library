package helper

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

type stepInfo struct {
	CobraCmdFuncName string
	CreateCmdVar     string
	ExportPrefix     string
	FlagsFunc        string
	Long             string
	StepParameters   []config.StepParameters
	StepAliases      []config.Alias
	OSImport         bool
	OutputResources  []map[string]string
	Short            string
	StepFunc         string
	StepName         string
	StepSecrets      []string
	Containers       []config.Container
	Sidecars         []config.Container
	Outputs          config.StepOutputs
	Resources        []config.StepResources
	Secrets          []config.StepSecrets
	StepErrors       []config.StepError
}

//go:embed templates/step.go.tmpl
var stepGoTemplate string

//go:embed templates/step_test.go.tmpl
var stepTestGoTemplate string

//go:embed templates/step_impl.go.tmpl
var stepGoImplementationTemplate string

//go:embed templates/step_impl_test.go.tmpl
var stepGoImplementationTestTemplate string

//go:embed templates/metadata.go.tmpl
var metadataGeneratedTemplate string

const metadataGeneratedFileName = "metadata_generated.go"

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
		fmt.Printf("Step name: %v\n", stepName)
		if stepName+".yaml" != filepath.Base(configFilePath) {
			fmt.Printf("Expected file %s to have name %s.yaml (<stepName>.yaml)\n", configFilePath, filepath.Join(filepath.Dir(configFilePath), stepName))
			os.Exit(1)
		}
		allSteps.Steps = append(allSteps.Steps, stepName)

		for _, parameter := range stepData.Spec.Inputs.Parameters {
			for _, mandatoryIfCase := range parameter.MandatoryIf {
				if mandatoryIfCase.Name == "" || mandatoryIfCase.Value == "" {
					return errors.New("invalid mandatoryIf option")
				}
			}
		}

		osImport := false
		osImport, err = setDefaultParameters(&stepData)
		if err != nil {
			log.Fatalf("Error occurred: %v\n", err)
		}

		myStepInfo, err := getStepInfo(&stepData, osImport, stepHelperData.ExportPrefix)
		if err != nil {
			log.Fatalf("Error occurred: %v\n", err)
		}

		step := stepTemplate(myStepInfo, "step", stepGoTemplate)
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
			err = stepHelperData.WriteFile(filepath.Join(targetDir, fmt.Sprintf("%v.go", stepName)), impl, 0644)
			if err != nil {
				log.Fatalf("Error occurred: %v\n", err)
			}
		}

		exists, _ = piperutils.FileExists(filepath.Join(targetDir, fmt.Sprintf("%v_test.go", stepName)))
		if !exists {
			impl := stepImplementation(myStepInfo, "implTest", stepGoImplementationTestTemplate)
			err = stepHelperData.WriteFile(filepath.Join(targetDir, fmt.Sprintf("%v_test.go", stepName)), impl, 0644)
			if err != nil {
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

func setDefaultParameters(stepData *config.StepData) (bool, error) {
	// ToDo: custom function for default handling, support all relevant parameter types
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
			case "map[string]interface{}", "[]map[string]interface{}":
				// Currently we don't need to set a default here since in this case the default
				// is never used. Needs to be changed in case we enable cli parameter handling
				// for that type.
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
				param.Default = fmt.Sprintf("`%v`", param.Default)
			case "[]string":
				param.Default = fmt.Sprintf("[]string{`%v`}", strings.Join(getStringSliceFromInterface(param.Default), "`, `"))
			case "map[string]interface{}", "[]map[string]interface{}":
				// Currently we don't need to set a default here since in this case the default
				// is never used. Needs to be changed in case we enable cli parameter handling
				// for that type.
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
			CobraCmdFuncName: fmt.Sprintf("%vCommand", piperutils.Title(stepData.Metadata.Name)),
			CreateCmdVar:     fmt.Sprintf("create%vCmd", piperutils.Title(stepData.Metadata.Name)),
			Short:            stepData.Metadata.Description,
			Long:             stepData.Metadata.LongDescription,
			StepParameters:   stepData.Spec.Inputs.Parameters,
			StepAliases:      stepData.Metadata.Aliases,
			FlagsFunc:        fmt.Sprintf("add%vFlags", piperutils.Title(stepData.Metadata.Name)),
			OSImport:         osImport,
			OutputResources:  oRes,
			ExportPrefix:     exportPrefix,
			StepSecrets:      getSecretFields(stepData),
			Containers:       stepData.Spec.Containers,
			Sidecars:         stepData.Spec.Sidecars,
			Outputs:          stepData.Spec.Outputs,
			Resources:        stepData.Spec.Inputs.Resources,
			Secrets:          stepData.Spec.Inputs.Secrets,
			StepErrors:       stepData.Metadata.Errors,
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

func getOutputResourceDetails(stepData *config.StepData) ([]map[string]string, error) {
	outputResources := []map[string]string{}

	for _, res := range stepData.Spec.Outputs.Resources {
		currentResource := map[string]string{}
		currentResource["name"] = res.Name
		currentResource["type"] = res.Type

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
							influxMeasurement.Fields = append(influxMeasurement.Fields, InfluxMetric{Name: fmt.Sprintf("%v", fieldParams["name"]), Type: fmt.Sprintf("%v", fieldParams["type"])})
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
			currentResource["def"] = def
			currentResource["objectname"] = reportsResource.StructName()
			outputResources = append(outputResources, currentResource)
		}
	}

	return outputResources, nil
}

// MetadataFiles provides a list of all step metadata files
func MetadataFiles(sourceDirectory string) ([]string, error) {
	var metadataFiles []string

	_ = filepath.Walk(sourceDirectory, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".yaml" {
			metadataFiles = append(metadataFiles, path)
		}
		return nil
	})

	return metadataFiles, nil
}

func isCLIParam(myType string) bool {
	return myType != "map[string]interface{}" && myType != "[]map[string]interface{}"
}

func stepTemplate(myStepInfo stepInfo, templateName, goTemplate string) []byte {
	funcMap := sprig.HermeticTxtFuncMap()
	funcMap["flagType"] = flagType
	funcMap["golangName"] = golangNameTitle
	funcMap["title"] = piperutils.Title
	funcMap["longName"] = longName
	funcMap["uniqueName"] = mustUniqName
	funcMap["isCLIParam"] = isCLIParam

	return generateCode(myStepInfo, templateName, goTemplate, funcMap)
}

func stepImplementation(myStepInfo stepInfo, templateName, goTemplate string) []byte {
	funcMap := sprig.HermeticTxtFuncMap()
	funcMap["title"] = piperutils.Title
	funcMap["uniqueName"] = mustUniqName

	return generateCode(myStepInfo, templateName, goTemplate, funcMap)
}

func generateCode(dataObject interface{}, templateName, goTemplate string, funcMap template.FuncMap) []byte {
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
	properName := strings.Replace(name, "Api", "API", -1)
	properName = strings.Replace(properName, "api", "API", -1)
	properName = strings.Replace(properName, "Url", "URL", -1)
	properName = strings.Replace(properName, "Id", "ID", -1)
	properName = strings.Replace(properName, "Json", "JSON", -1)
	properName = strings.Replace(properName, "json", "JSON", -1)
	properName = strings.Replace(properName, "Tls", "TLS", -1)
	return properName
}

// golangNameTitle returns name in title case with abbriviations in capital (API, URL, ID, JSON, TLS)
func golangNameTitle(name string) string {
	return piperutils.Title(golangName(name))
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
		log.Fatalf("Meta data type not set or not known: '%v'\n", paramType)
	}
	return theFlagType
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
		return nil, fmt.Errorf("cannot find uniq on type %s", tp)
	}
}
