package generator

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/SAP/jenkins-library/pkg/config"
)

// DocuHelperData is used to transport the needed parameters and functions from the step generator to the docu generation.
type DocuHelperData struct {
	DocTemplatePath     string
	OpenDocTemplateFile func(d string) (io.ReadCloser, error)
	DocFileWriter       func(f string, d []byte, p os.FileMode) error
	OpenFile            func(s string) (io.ReadCloser, error)
}

var stepParameterNames []string

func readStepConfiguration(stepMetadata config.StepData, customDefaultFiles []string, docuHelperData DocuHelperData) config.StepConfig {
	filters := stepMetadata.GetParameterFilters()
	filters.All = append(filters.All, "collectTelemetryData")
	filters.General = append(filters.General, "collectTelemetryData")
	filters.Parameters = append(filters.Parameters, "collectTelemetryData")

	defaultFiles := []io.ReadCloser{}
	for _, projectDefaultFile := range customDefaultFiles {
		fc, _ := docuHelperData.OpenFile(projectDefaultFile)
		defer fc.Close()
		defaultFiles = append(defaultFiles, fc)
	}

	configuration := config.Config{}
	stepConfiguration, err := configuration.GetStepConfig(
		map[string]interface{}{},
		"",
		nil,
		defaultFiles,
		false,
		filters,
		stepMetadata.Spec.Inputs.Parameters,
		stepMetadata.Spec.Inputs.Secrets,
		map[string]interface{}{},
		"",
		stepMetadata.Metadata.Name,
		stepMetadata.Metadata.Aliases,
	)
	checkError(err)
	return stepConfiguration
}

// GenerateStepDocumentation generates step coding based on step configuration provided in yaml files
func GenerateStepDocumentation(metadataFiles []string, customDefaultFiles []string, docuHelperData DocuHelperData) error {
	for key := range metadataFiles {
		stepMetadata := readStepMetadata(metadataFiles[key], docuHelperData)

		adjustDefaultValues(&stepMetadata)

		stepConfiguration := readStepConfiguration(stepMetadata, customDefaultFiles, docuHelperData)

		applyCustomDefaultValues(&stepMetadata, stepConfiguration)

		adjustMandatoryFlags(&stepMetadata)

		fmt.Print("  Generate documentation.. ")
		if err := generateStepDocumentation(stepMetadata, docuHelperData); err != nil {
			fmt.Println("")
			fmt.Println(err)
		} else {
			fmt.Println("completed")
		}
	}
	return nil
}

// generates the step documentation and replaces the template with the generated documentation
func generateStepDocumentation(stepData config.StepData, docuHelperData DocuHelperData) error {
	//create the file path for the template and open it.
	docTemplateFilePath := fmt.Sprintf("%v%v.md", docuHelperData.DocTemplatePath, stepData.Metadata.Name)
	docTemplate, err := docuHelperData.OpenDocTemplateFile(docTemplateFilePath)

	if docTemplate != nil {
		defer docTemplate.Close()
	}
	// check if there is an error during opening the template (true : skip docu generation for this meta data file)
	if err != nil {
		return fmt.Errorf("error occurred: %v", err)
	}

	content := readAndAdjustTemplate(docTemplate)
	if len(content) <= 0 {
		return fmt.Errorf("error occurred: no content inside of the template")
	}

	// binding of functions and placeholder
	tmpl, err := template.New("doc").Funcs(template.FuncMap{
		"StepName":    createStepName,
		"Description": createDescriptionSection,
		"Parameters":  createParametersSection,
	}).Parse(content)
	checkError(err)

	// add secrets, context defaults to the step parameters
	handleStepParameters(&stepData)

	// write executed template data to the previously opened file
	var docContent bytes.Buffer
	err = tmpl.Execute(&docContent, &stepData)
	checkError(err)

	// overwrite existing file
	err = docuHelperData.DocFileWriter(docTemplateFilePath, docContent.Bytes(), 644)
	checkError(err)

	return nil
}

func readContextInformation(contextDetailsPath string, contextDetails *config.StepData) {
	contextDetailsFile, err := os.Open(contextDetailsPath)
	checkError(err)
	defer contextDetailsFile.Close()

	err = contextDetails.ReadPipelineStepData(contextDetailsFile)
	checkError(err)
}

func getContainerParameters(container config.Container, sidecar bool) map[string]interface{} {
	containerParams := map[string]interface{}{}

	if len(container.Command) > 0 {
		containerParams[ifThenElse(sidecar, "sidecarCommand", "containerCommand")] = container.Command[0]
	}
	if len(container.EnvVars) > 0 {
		containerParams[ifThenElse(sidecar, "sidecarEnvVars", "dockerEnvVars")] = config.EnvVarsAsMap(container.EnvVars)
	}
	containerParams[ifThenElse(sidecar, "sidecarImage", "dockerImage")] = container.Image
	containerParams[ifThenElse(sidecar, "sidecarPullImage", "dockerPullImage")] = container.ImagePullPolicy != "Never"
	if len(container.Name) > 0 {
		containerParams[ifThenElse(sidecar, "sidecarName", "containerName")] = container.Name
		containerParams["dockerName"] = container.Name
	}
	if len(container.Options) > 0 {
		containerParams[ifThenElse(sidecar, "sidecarOptions", "dockerOptions")] = container.Options
	}
	if len(container.WorkingDir) > 0 {
		containerParams[ifThenElse(sidecar, "sidecarWorkspace", "dockerWorkspace")] = container.WorkingDir
	}

	if sidecar {
		if len(container.ReadyCommand) > 0 {
			containerParams["sidecarReadyCommand"] = container.ReadyCommand
		}
	} else {
		if len(container.Shell) > 0 {
			containerParams["containerShell"] = container.Shell
		}
	}

	//ToDo? add dockerVolumeBind, sidecarVolumeBind -> so far not part of config.Container

	return containerParams
}

func handleStepParameters(stepData *config.StepData) {

	stepParameterNames = stepData.GetParameterFilters().All

	//add general options like script, verbose, etc.
	//ToDo: add to context.yaml
	appendGeneralOptionsToParameters(stepData)

	//consolidate conditional parameters:
	//- remove duplicate parameter entries
	//- combine defaults (consider conditions)
	consolidateConditionalParameters(stepData)

	//get the context defaults
	appendContextParameters(stepData)

	//consolidate context defaults:
	//- combine defaults (consider conditions)
	consolidateContextDefaults(stepData)

	setDefaultAndPossisbleValues(stepData)
}

func setDefaultAndPossisbleValues(stepData *config.StepData) {
	for k, param := range stepData.Spec.Inputs.Parameters {

		//fill default if not set
		if param.Default == nil {
			switch param.Type {
			case "bool":
				param.Default = false
			case "int":
				param.Default = 0
			}
		}

		//add possible values where known for certain types
		switch param.Type {
		case "bool":
			if param.PossibleValues == nil {
				param.PossibleValues = []interface{}{true, false}
			}
		}

		stepData.Spec.Inputs.Parameters[k] = param
	}
}

func appendGeneralOptionsToParameters(stepData *config.StepData) {
	script := config.StepParameters{
		Name: "script", Type: "Jenkins Script", Mandatory: true,
		Description: "The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the `this` parameter, as in `script: this`. This allows the function to access the `commonPipelineEnvironment` for retrieving, e.g. configuration parameters.",
	}
	verbose := config.StepParameters{
		Name: "verbose", Type: "bool", Mandatory: false, Default: false, Scope: []string{"PARAMETERS", "GENERAL", "STEPS", "STAGES"},
		Description: "verbose output",
	}
	stepData.Spec.Inputs.Parameters = append(stepData.Spec.Inputs.Parameters, script, verbose)
}
