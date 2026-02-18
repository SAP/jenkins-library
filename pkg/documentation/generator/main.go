package generator

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/piperutils"

	"github.com/ghodss/yaml"
)

// DocuHelperData is used to transport the needed parameters and functions from the step generator to the docu generation.
type DocuHelperData struct {
	DocTemplatePath     string
	OpenDocTemplateFile func(d string) (io.ReadCloser, error)
	DocFileWriter       func(f string, d []byte, p os.FileMode) error
	OpenFile            func(s string) (io.ReadCloser, error)
}

var stepParameterNames []string
var includeAzure, includeGHA bool

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
		map[string]any{},
		"",
		nil,
		defaultFiles,
		false,
		filters,
		stepMetadata,
		map[string]any{},
		"",
		stepMetadata.Metadata.Name,
	)
	checkError(err)
	return stepConfiguration
}

// GenerateStepDocumentation generates step coding based on step configuration provided in yaml files
func GenerateStepDocumentation(metadataFiles []string, customDefaultFiles []string, docuHelperData DocuHelperData, azure bool, githubAction bool) error {
	includeAzure = azure
	includeGHA = githubAction
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

func getContainerParameters(container config.Container, sidecar bool) map[string]any {
	containerParams := map[string]any{}

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
				param.PossibleValues = []any{true, false}
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

// GenerateStepDocumentation generates pipeline stage documentation based on pipeline configuration provided in a yaml file
func GenerateStageDocumentation(stageMetadataPath, stageTargetPath, relativeStepsPath string, utils piperutils.FileUtils) error {
	if len(stageTargetPath) == 0 {
		return fmt.Errorf("stageTargetPath cannot be empty")
	}
	if len(stageMetadataPath) == 0 {
		return fmt.Errorf("stageMetadataPath cannot be empty")
	}

	if err := utils.MkdirAll(stageTargetPath, 0777); err != nil {
		return fmt.Errorf("failed to create directory '%v': %w", stageTargetPath, err)
	}

	stageMetadataContent, err := utils.FileRead(stageMetadataPath)
	if err != nil {
		return fmt.Errorf("failed to read stage metadata file '%v': %w", stageMetadataPath, err)
	}

	stageRunConfig := config.RunConfigV1{}

	err = yaml.Unmarshal(stageMetadataContent, &stageRunConfig.PipelineConfig)
	if err != nil {
		return fmt.Errorf("format of configuration is invalid %q: %w", stageMetadataContent, err)
	}

	err = createPipelineDocumentation(&stageRunConfig, stageTargetPath, relativeStepsPath, utils)
	if err != nil {
		return fmt.Errorf("failed to create pipeline documentation: %w", err)
	}

	return nil
}

func createPipelineDocumentation(stageRunConfig *config.RunConfigV1, stageTargetPath, relativeStepsPath string, utils piperutils.FileUtils) error {
	if err := createPipelineOverviewDocumentation(stageRunConfig, stageTargetPath, utils); err != nil {
		return fmt.Errorf("failed to create pipeline overview: %w", err)
	}

	if err := createPipelineStageDocumentation(stageRunConfig, stageTargetPath, relativeStepsPath, utils); err != nil {
		return fmt.Errorf("failed to create pipeline stage details: %w", err)
	}

	return nil
}

func createPipelineOverviewDocumentation(stageRunConfig *config.RunConfigV1, stageTargetPath string, utils piperutils.FileUtils) error {
	overviewFileName := "overview.md"
	var overviewDoc strings.Builder
	overviewDoc.WriteString(fmt.Sprintf("# %v\n\n", stageRunConfig.PipelineConfig.Metadata.DisplayName))
	overviewDoc.WriteString(fmt.Sprintf("%v\n\n", stageRunConfig.PipelineConfig.Metadata.Description))
	overviewDoc.WriteString(fmt.Sprintf("The %v comprises following stages\n\n", stageRunConfig.PipelineConfig.Metadata.DisplayName))
	for _, stage := range stageRunConfig.PipelineConfig.Spec.Stages {
		stageFilePath := filepath.Join(stageTargetPath, fmt.Sprintf("%v.md", stage.Name))
		overviewDoc.WriteString(fmt.Sprintf("* [%v Stage](%v)\n", stage.DisplayName, stageFilePath))
	}
	overviewFilePath := filepath.Join(stageTargetPath, overviewFileName)
	fmt.Println("writing file", overviewFilePath)
	return utils.FileWrite(overviewFilePath, []byte(overviewDoc.String()), 0666)
}

const stepConditionDetails = `!!! note "Step condition details"
    There are currently several conditions which can be checked.<br />**Important: It will be sufficient that any one condition per step is met.**

    * ` + "`" + `config` + "`" + `: Checks if a configuration parameter has a defined value.
	* ` + "`" + `config key` + "`" + `: Checks if a defined configuration parameter is set.
    * ` + "`" + `file pattern` + "`" + `: Checks if files according a defined pattern exist in the project.
	* ` + "`" + `file pattern from config` + "`" + `: Checks if files according a pattern defined in the custom configuration exist in the project.
    * ` + "`" + `npm script` + "`" + `: Checks if a npm script exists in one of the package.json files in the repositories.

`
const overrulingStepActivation = `!!! note "Overruling step activation conditions"
    It is possible to overrule the automatically detected step activation status.

    * In case a step will be **active** you can add to your stage configuration ` + "`" + `<stepName>: false` + "`" + ` to explicitly **deactivate** the step.
    * In case a step will be **inactive** you can add to your stage configuration ` + "`" + `<stepName>: true` + "`" + ` to explicitly **activate** the step.

`

func createPipelineStageDocumentation(stageRunConfig *config.RunConfigV1, stageTargetPath, relativeStepsPath string, utils piperutils.FileUtils) error {
	for _, stage := range stageRunConfig.PipelineConfig.Spec.Stages {
		var stageDoc strings.Builder
		stageDoc.WriteString(fmt.Sprintf("# %v\n\n", stage.DisplayName))
		stageDoc.WriteString(fmt.Sprintf("%v\n\n", stage.Description))

		if len(stage.Steps) > 0 {
			stageDoc.WriteString("## Stage Content\n\nThis stage comprises following steps which are activated depending on your use-case/configuration:\n\n")

			for i, step := range stage.Steps {
				if i == 0 {
					stageDoc.WriteString("| step | step description |\n")
					stageDoc.WriteString("| ---- | ---------------- |\n")
				}

				var orchestratorBadges strings.Builder
				for _, orchestrator := range step.Orchestrators {
					orchestratorBadges.WriteString(getBadge(orchestrator) + " ")
				}

				stageDoc.WriteString(fmt.Sprintf("| [%v](%v/%v.md) | %v%v |\n", step.Name, relativeStepsPath, step.Name, orchestratorBadges.String(), step.Description))
			}

			stageDoc.WriteString("\n")

			stageDoc.WriteString("## Stage & Step Activation\n\nThis stage will be active in case one of following conditions are met:\n\n")
			stageDoc.WriteString("* One of the steps is explicitly activated by using `<stepName>: true` in the stage configuration\n")
			stageDoc.WriteString("* At least one of the step conditions is met and steps are not explicitly deactivated by using `<stepName>: false` in the stage configuration\n\n")

			stageDoc.WriteString(stepConditionDetails)
			stageDoc.WriteString(overrulingStepActivation)

			stageDoc.WriteString("Following conditions apply for activation of steps contained in the stage:\n\n")

			stageDoc.WriteString("| step | active if one of following conditions is met |\n")
			stageDoc.WriteString("| ---- | -------------------------------------------- |\n")

			// add step condition details
			for _, step := range stage.Steps {
				stageDoc.WriteString(fmt.Sprintf("| [%v](%v/%v.md) | %v |\n", step.Name, relativeStepsPath, step.Name, getStepConditionDetails(step)))
			}
		}

		stageFilePath := filepath.Join(stageTargetPath, fmt.Sprintf("%v.md", stage.Name))
		fmt.Println("writing file", stageFilePath)
		if err := utils.FileWrite(stageFilePath, []byte(stageDoc.String()), 0666); err != nil {
			return fmt.Errorf("failed to write stage file '%v': %w", stageFilePath, err)
		}
	}
	return nil
}

func getBadge(orchestrator string) string {
	orchestratorOnly := piperutils.Title(strings.ToLower(orchestrator)) + " only"
	urlPath := &url.URL{Path: orchestratorOnly}
	orchestratorOnlyString := urlPath.String()

	return fmt.Sprintf("[![%v](https://img.shields.io/badge/-%v-yellowgreen)](#)", orchestratorOnly, orchestratorOnlyString)
}

func getStepConditionDetails(step config.Step) string {
	stepConditions := ""
	if step.Conditions == nil || len(step.Conditions) == 0 {
		return "**active** by default - deactivate explicitly"
	}

	if len(step.Orchestrators) > 0 {
		var orchestratorBadges strings.Builder
		for _, orchestrator := range step.Orchestrators {
			orchestratorBadges.WriteString(getBadge(orchestrator) + " ")
		}
		stepConditions = orchestratorBadges.String() + "<br />"
	}

	for _, condition := range step.Conditions {
		if condition.Config != nil && len(condition.Config) > 0 {
			stepConditions += "<i>config:</i><ul>"
			for param, activationValues := range condition.Config {
				for _, activationValue := range activationValues {
					stepConditions += fmt.Sprintf("<li>`%v`: `%v`</li>", param, activationValue)
				}
				// config condition only covers first entry
				break
			}
			stepConditions += "</ul>"
			continue
		}

		if len(condition.ConfigKey) > 0 {
			stepConditions += fmt.Sprintf("<i>config key:</i>&nbsp;`%v`<br />", condition.ConfigKey)
			continue
		}

		if len(condition.FilePattern) > 0 {
			stepConditions += fmt.Sprintf("<i>file pattern:</i>&nbsp;`%v`<br />", condition.FilePattern)
			continue
		}

		if len(condition.FilePatternFromConfig) > 0 {
			stepConditions += fmt.Sprintf("<i>file pattern from config:</i>&nbsp;`%v`<br />", condition.FilePatternFromConfig)
			continue
		}

		if len(condition.NpmScript) > 0 {
			stepConditions += fmt.Sprintf("<i>npm script:</i>&nbsp;`%v`<br />", condition.NpmScript)
			continue
		}

		if condition.Inactive {
			stepConditions += "**inactive** by default - activate explicitly"
			continue
		}
	}

	return stepConditions
}
