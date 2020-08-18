package helper

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
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
}

func readTemplateResource(baseName string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get cwd: %w", err)
	}
	fmt.Printf("current dir: %s\n", cwd)

	packagePath := filepath.Join("pkg", "generator", "helper")
	resourcePath := baseName + "Template.txt"
	if !strings.HasSuffix(cwd, packagePath) {
		resourcePath = filepath.Join(packagePath, resourcePath)
	}

	contents, err := ioutil.ReadFile(resourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file '%s': %w", resourcePath, err)
	}
	return string(contents), nil
}

// ProcessMetaFiles generates step coding based on step configuration provided in yaml files
func ProcessMetaFiles(metadataFiles []string, targetDir string, stepHelperData StepHelperData, docuHelperData DocuHelperData) error {

	templateBundle := make(map[string]string)
	templateBaseNames := []string{
		"stepGoImplementation",
		"stepGoImplementationTest",
		"stepGo",
		"stepGoTest",
	}

	for _, baseName := range templateBaseNames {
		contents, err := readTemplateResource(baseName)
		if err != nil {
			return err
		}
		templateBundle[baseName] = contents
	}

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

			step := stepTemplate(myStepInfo, "step", templateBundle["stepGo"])
			err = stepHelperData.WriteFile(filepath.Join(targetDir, fmt.Sprintf("%v_generated.go", stepData.Metadata.Name)), step, 0644)
			checkError(err)

			test := stepTemplate(myStepInfo, "stepTest", templateBundle["stepGoTest"])
			err = stepHelperData.WriteFile(filepath.Join(targetDir, fmt.Sprintf("%v_generated_test.go", stepData.Metadata.Name)), test, 0644)
			checkError(err)

			exists, _ := piperutils.FileExists(filepath.Join(targetDir, fmt.Sprintf("%v.go", stepData.Metadata.Name)))
			if !exists {
				impl := stepImplementation(myStepInfo, "impl", templateBundle["stepGoImplementation"])
				err = stepHelperData.WriteFile(filepath.Join(targetDir, fmt.Sprintf("%v.go", stepData.Metadata.Name)), impl, 0644)
				checkError(err)
			}

			exists, _ = piperutils.FileExists(filepath.Join(targetDir, fmt.Sprintf("%v_test.go", stepData.Metadata.Name)))
			if !exists {
				impl := stepImplementation(myStepInfo, "implTest", templateBundle["stepGoImplementationTest"])
				err = stepHelperData.WriteFile(filepath.Join(targetDir, fmt.Sprintf("%v_test.go", stepData.Metadata.Name)), impl, 0644)
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
				param.Default = fmt.Sprintf("`%v`", param.Default)
			case "[]string":
				param.Default = fmt.Sprintf("[]string{`%v`}", strings.Join(getStringSliceFromInterface(param.Default), "`, `"))
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
			StepParameters:   stepData.Spec.Inputs.Parameters,
			StepAliases:      stepData.Metadata.Aliases,
			FlagsFunc:        fmt.Sprintf("add%vFlags", strings.Title(stepData.Metadata.Name)),
			OSImport:         osImport,
			OutputResources:  oRes,
			ExportPrefix:     exportPrefix,
			StepSecrets:      getSecretFields(stepData),
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

func stepTemplate(myStepInfo stepInfo, templateName, goTemplate string) []byte {
	funcMap := sprig.HermeticTxtFuncMap()
	funcMap["flagType"] = flagType
	funcMap["golangName"] = golangNameTitle
	funcMap["title"] = strings.Title
	funcMap["longName"] = longName
	funcMap["uniqueName"] = mustUniqName

	return generateCode(myStepInfo, templateName, goTemplate, funcMap)
}

func stepImplementation(myStepInfo stepInfo, templateName, goTemplate string) []byte {
	funcMap := sprig.HermeticTxtFuncMap()
	funcMap["title"] = strings.Title
	funcMap["uniqueName"] = mustUniqName

	return generateCode(myStepInfo, templateName, goTemplate, funcMap)
}

func generateCode(myStepInfo stepInfo, templateName, goTemplate string, funcMap template.FuncMap) []byte {
	tmpl, err := template.New(templateName).Funcs(funcMap).Parse(goTemplate)
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
	properName = strings.Replace(properName, "Tls", "TLS", -1)
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
			if !piperutils.ContainsString(names, item.Name) {
				names = append(names, item.Name)
				dest = append(dest, item)
			}
		}

		return dest, nil
	default:
		return nil, fmt.Errorf("Cannot find uniq on type %s", tp)
	}
}
