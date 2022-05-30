// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/gcs"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/bmatcuk/doublestar"
	"github.com/spf13/cobra"
)

type npmExecuteScriptsOptions struct {
	Tool                       string   `json:"tool,omitempty"`
	Install                    bool     `json:"install,omitempty"`
	RunScripts                 []string `json:"runScripts,omitempty"`
	DefaultNpmRegistry         string   `json:"defaultNpmRegistry,omitempty"`
	VirtualFrameBuffer         bool     `json:"virtualFrameBuffer,omitempty"`
	ScriptOptions              []string `json:"scriptOptions,omitempty"`
	BuildDescriptorExcludeList []string `json:"buildDescriptorExcludeList,omitempty"`
	BuildDescriptorList        []string `json:"buildDescriptorList,omitempty"`
	CreateBOM                  bool     `json:"createBOM,omitempty"`
	Publish                    bool     `json:"publish,omitempty"`
	RepositoryURL              string   `json:"repositoryUrl,omitempty"`
	RepositoryPassword         string   `json:"repositoryPassword,omitempty"`
	RepositoryUsername         string   `json:"repositoryUsername,omitempty"`
	BuildSettingsInfo          string   `json:"buildSettingsInfo,omitempty"`
	PackBeforePublish          bool     `json:"packBeforePublish,omitempty"`
}

type npmExecuteScriptsCommonPipelineEnvironment struct {
	custom struct {
		buildSettingsInfo string
	}
}

func (p *npmExecuteScriptsCommonPipelineEnvironment) persist(path, resourceName string) {
	content := []struct {
		category string
		name     string
		value    interface{}
	}{
		{category: "custom", name: "buildSettingsInfo", value: p.custom.buildSettingsInfo},
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
		log.Entry().Error("failed to persist Piper environment")
	}
}

type npmExecuteScriptsReports struct {
}

func (p *npmExecuteScriptsReports) persist(stepConfig npmExecuteScriptsOptions, gcpJsonKeyFilePath string, gcsBucketId string, gcsFolderPath string, gcsSubFolder string) {
	if gcsBucketId == "" {
		log.Entry().Info("persisting reports to GCS is disabled, because gcsBucketId is empty")
		return
	}
	log.Entry().Info("Uploading reports to Google Cloud Storage...")
	content := []gcs.ReportOutputParam{
		{FilePattern: "**/bom.xml", ParamRef: "", StepResultType: "sbom"},
		{FilePattern: "**/TEST-*.xml", ParamRef: "", StepResultType: "junit"},
		{FilePattern: "**/cobertura-coverage.xml", ParamRef: "", StepResultType: "cobertura-coverage"},
		{FilePattern: "**/e2e/*.json", ParamRef: "", StepResultType: "cucumber"},
	}
	envVars := []gcs.EnvVar{
		{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: gcpJsonKeyFilePath, Modified: false},
	}
	gcsClient, err := gcs.NewClient(gcs.WithEnvVars(envVars))
	if err != nil {
		log.Entry().Errorf("creation of GCS client failed: %v", err)
		return
	}
	defer gcsClient.Close()
	structVal := reflect.ValueOf(&stepConfig).Elem()
	inputParameters := map[string]string{}
	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Type().Field(i)
		if field.Type.String() == "string" {
			paramName := strings.Split(field.Tag.Get("json"), ",")
			paramValue, _ := structVal.Field(i).Interface().(string)
			inputParameters[paramName[0]] = paramValue
		}
	}
	if err := gcs.PersistReportsToGCS(gcsClient, content, inputParameters, gcsFolderPath, gcsBucketId, gcsSubFolder, doublestar.Glob, os.Stat); err != nil {
		log.Entry().Errorf("failed to persist reports: %v", err)
	}
}

// NpmExecuteScriptsCommand Execute npm run scripts on all npm packages in a project
func NpmExecuteScriptsCommand() *cobra.Command {
	const STEP_NAME = "npmExecuteScripts"

	metadata := npmExecuteScriptsMetadata()
	var stepConfig npmExecuteScriptsOptions
	var startTime time.Time
	var commonPipelineEnvironment npmExecuteScriptsCommonPipelineEnvironment
	var reports npmExecuteScriptsReports
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createNpmExecuteScriptsCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Execute npm run scripts on all npm packages in a project",
		Long: `Execute npm run scripts in all package json files, if they implement the scripts.

### build with depedencies from a private repository
if your build has scoped/unscoped dependencies from a private repository you can include a .npmrc into the source code
repository as below (replace the ` + "`" + `@privateScope:registry` + "`" + ` value(s) with a valid private repo url) :

` + "`" + `` + "`" + `` + "`" + `
@privateScope:registry=https://private.repository.com/
//private.repository.com/:username=${PIPER_VAULTCREDENTIAL_USER}
//private.repository.com/:_password=${PIPER_VAULTCREDENTIAL_PASSWORD_BASE64}
//private.repository.com/:always-auth=true
registry=https://registry.npmjs.org
` + "`" + `` + "`" + `` + "`" + `
` + "`" + `PIPER_VAULTCREDENTIAL_USER` + "`" + ` and ` + "`" + `PIPER_VAULTCREDENTIAL_PASSWORD_BASE64` + "`" + ` (Base64 encoded password) are the username and password for the private repository
and are exposed are environment variables that must be present in the environment where the Piper step runs or alternatively can be created using :
[vault general purpose credentials](../infrastructure/vault.md#using-vault-for-general-purpose-and-test-credentials)`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err := PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}
			log.RegisterSecret(stepConfig.RepositoryPassword)
			log.RegisterSecret(stepConfig.RepositoryUsername)

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunkClient = &splunk.Splunk{}
				logCollector = &log.CollectorHook{CorrelationID: GeneralConfig.CorrelationID}
				log.RegisterHook(logCollector)
			}

			validation, err := validation.New(validation.WithJSONNamesForStructFields(), validation.WithPredefinedErrorMessages())
			if err != nil {
				return err
			}
			if err = validation.ValidateStruct(stepConfig); err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			stepTelemetryData := telemetry.CustomData{}
			stepTelemetryData.ErrorCode = "1"
			handler := func() {
				commonPipelineEnvironment.persist(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
				reports.persist(stepConfig, GeneralConfig.GCPJsonKeyFilePath, GeneralConfig.GCSBucketId, GeneralConfig.GCSFolderPath, GeneralConfig.GCSSubFolder)
				config.RemoveVaultSecretFiles()
				stepTelemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				stepTelemetryData.ErrorCategory = log.GetErrorCategory().String()
				stepTelemetryData.PiperCommitHash = GitCommit
				telemetryClient.SetData(&stepTelemetryData)
				telemetryClient.Send()
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetryClient.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunkClient.Initialize(GeneralConfig.CorrelationID,
					GeneralConfig.HookConfig.SplunkConfig.Dsn,
					GeneralConfig.HookConfig.SplunkConfig.Token,
					GeneralConfig.HookConfig.SplunkConfig.Index,
					GeneralConfig.HookConfig.SplunkConfig.SendLogs)
			}
			npmExecuteScripts(stepConfig, &stepTelemetryData, &commonPipelineEnvironment)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addNpmExecuteScriptsFlags(createNpmExecuteScriptsCmd, &stepConfig)
	return createNpmExecuteScriptsCmd
}

func addNpmExecuteScriptsFlags(cmd *cobra.Command, stepConfig *npmExecuteScriptsOptions) {
	cmd.Flags().StringVar(&stepConfig.Tool, "tool", `auto`, "The to run the script. Can be one of: auto, npm, yarn, pnpm.\nDefault is 'auto' to auto-detect the tool based on the lock file.\n")
	cmd.Flags().BoolVar(&stepConfig.Install, "install", true, "Run npm install or similar commands depending on the project structure.")
	cmd.Flags().StringSliceVar(&stepConfig.RunScripts, "runScripts", []string{}, "List of additional run scripts to execute from package.json.")
	cmd.Flags().StringVar(&stepConfig.DefaultNpmRegistry, "defaultNpmRegistry", os.Getenv("PIPER_defaultNpmRegistry"), "URL of the npm registry to use. Defaults to https://registry.npmjs.org/")
	cmd.Flags().BoolVar(&stepConfig.VirtualFrameBuffer, "virtualFrameBuffer", false, "(Linux only) Start a virtual frame buffer in the background. This allows you to run a web browser without the need for an X server. Note that xvfb needs to be installed in the execution environment.")
	cmd.Flags().StringSliceVar(&stepConfig.ScriptOptions, "scriptOptions", []string{}, "Options are passed to all runScripts calls separated by a '--'. './piper npmExecuteScripts --runScripts ci-e2e --scriptOptions '--tag1' will correspond to 'npm run ci-e2e -- --tag1'")
	cmd.Flags().StringSliceVar(&stepConfig.BuildDescriptorExcludeList, "buildDescriptorExcludeList", []string{`deployment/**`}, "List of build descriptors and therefore modules to exclude from execution of the npm scripts. The elements can either be a path to the build descriptor or a pattern.")
	cmd.Flags().StringSliceVar(&stepConfig.BuildDescriptorList, "buildDescriptorList", []string{}, "List of build descriptors and therefore modules for execution of the npm scripts. The elements have to be paths to the build descriptors. **If set, buildDescriptorExcludeList will be ignored.**")
	cmd.Flags().BoolVar(&stepConfig.CreateBOM, "createBOM", false, "Create a BOM xml using CycloneDX.")
	cmd.Flags().BoolVar(&stepConfig.Publish, "publish", false, "Configures npm to publish the artifact to a repository.")
	cmd.Flags().StringVar(&stepConfig.RepositoryURL, "repositoryUrl", os.Getenv("PIPER_repositoryUrl"), "Url to the repository to which the project artifacts should be published.")
	cmd.Flags().StringVar(&stepConfig.RepositoryPassword, "repositoryPassword", os.Getenv("PIPER_repositoryPassword"), "Password for the repository to which the project artifacts should be published.")
	cmd.Flags().StringVar(&stepConfig.RepositoryUsername, "repositoryUsername", os.Getenv("PIPER_repositoryUsername"), "Username for the repository to which the project artifacts should be published.")
	cmd.Flags().StringVar(&stepConfig.BuildSettingsInfo, "buildSettingsInfo", os.Getenv("PIPER_buildSettingsInfo"), "build settings info is typically filled by the step automatically to create information about the build settings that were used during the npm build . This information is typically used for compliance related processes.")
	cmd.Flags().BoolVar(&stepConfig.PackBeforePublish, "packBeforePublish", false, "used for executing npm pack first, followed by npm publish. This two step maybe required when you are building a scoped packages and have npm dependencies from the same scope")

}

// retrieve step metadata
func npmExecuteScriptsMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "npmExecuteScripts",
			Aliases:     []config.Alias{{Name: "executeNpm", Deprecated: false}},
			Description: "Execute npm run scripts on all npm packages in a project",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Resources: []config.StepResources{
					{Name: "source", Type: "stash"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "tool",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `auto`,
					},
					{
						Name:        "install",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     true,
					},
					{
						Name:        "runScripts",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "defaultNpmRegistry",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "GENERAL", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "npm/defaultNpmRegistry"}},
						Default:     os.Getenv("PIPER_defaultNpmRegistry"),
					},
					{
						Name:        "virtualFrameBuffer",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "scriptOptions",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "buildDescriptorExcludeList",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{`deployment/**`},
					},
					{
						Name:        "buildDescriptorList",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "createBOM",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "publish",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name: "repositoryUrl",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/npmRepositoryURL",
							},

							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryUrl",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_repositoryUrl"),
					},
					{
						Name: "repositoryPassword",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/npmRepositoryPassword",
							},

							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryPassword",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_repositoryPassword"),
					},
					{
						Name: "repositoryUsername",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/npmRepositoryUsername",
							},

							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryUsername",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_repositoryUsername"),
					},
					{
						Name: "buildSettingsInfo",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/buildSettingsInfo",
							},
						},
						Scope:     []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_buildSettingsInfo"),
					},
					{
						Name:        "packBeforePublish",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
				},
			},
			Containers: []config.Container{
				{Name: "node", Image: "node:lts-stretch"},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "commonPipelineEnvironment",
						Type: "piperEnvironment",
						Parameters: []map[string]interface{}{
							{"name": "custom/buildSettingsInfo"},
						},
					},
					{
						Name: "reports",
						Type: "reports",
						Parameters: []map[string]interface{}{
							{"filePattern": "**/bom.xml", "type": "sbom"},
							{"filePattern": "**/TEST-*.xml", "type": "junit"},
							{"filePattern": "**/cobertura-coverage.xml", "type": "cobertura-coverage"},
							{"filePattern": "**/e2e/*.json", "type": "cucumber"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
