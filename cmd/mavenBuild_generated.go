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

type mavenBuildOptions struct {
	PomPath                         string   `json:"pomPath,omitempty"`
	Profiles                        []string `json:"profiles,omitempty"`
	Flatten                         bool     `json:"flatten,omitempty"`
	Verify                          bool     `json:"verify,omitempty"`
	ProjectSettingsFile             string   `json:"projectSettingsFile,omitempty"`
	GlobalSettingsFile              string   `json:"globalSettingsFile,omitempty"`
	M2Path                          string   `json:"m2Path,omitempty"`
	LogSuccessfulMavenTransfers     bool     `json:"logSuccessfulMavenTransfers,omitempty"`
	CreateBOM                       bool     `json:"createBOM,omitempty"`
	AltDeploymentRepositoryPassword string   `json:"altDeploymentRepositoryPassword,omitempty"`
	AltDeploymentRepositoryUser     string   `json:"altDeploymentRepositoryUser,omitempty"`
	AltDeploymentRepositoryURL      string   `json:"altDeploymentRepositoryUrl,omitempty"`
	AltDeploymentRepositoryID       string   `json:"altDeploymentRepositoryID,omitempty"`
	CustomTLSCertificateLinks       []string `json:"customTlsCertificateLinks,omitempty"`
	Publish                         bool     `json:"publish,omitempty"`
	JavaCaCertFilePath              string   `json:"javaCaCertFilePath,omitempty"`
	BuildSettingsInfo               string   `json:"buildSettingsInfo,omitempty"`
}

type mavenBuildCommonPipelineEnvironment struct {
	custom struct {
		buildSettingsInfo string
	}
}

func (p *mavenBuildCommonPipelineEnvironment) persist(path, resourceName string) {
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

type mavenBuildReports struct {
}

func (p *mavenBuildReports) persist(stepConfig mavenBuildOptions, gcpJsonKeyFilePath string, gcsBucketId string, gcsFolderPath string, gcsSubFolder string) {
	if gcsBucketId == "" {
		log.Entry().Info("persisting reports to GCS is disabled, because gcsBucketId is empty")
		return
	}
	log.Entry().Info("Uploading reports to Google Cloud Storage...")
	content := []gcs.ReportOutputParam{
		{FilePattern: "**/bom-maven.xml", ParamRef: "", StepResultType: "sbom"},
		{FilePattern: "**/TEST-*.xml", ParamRef: "", StepResultType: "junit"},
		{FilePattern: "**/jacoco.xml", ParamRef: "", StepResultType: "jacoco-coverage"},
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

// MavenBuildCommand This step will install the maven project into the local maven repository.
func MavenBuildCommand() *cobra.Command {
	const STEP_NAME = "mavenBuild"

	metadata := mavenBuildMetadata()
	var stepConfig mavenBuildOptions
	var startTime time.Time
	var commonPipelineEnvironment mavenBuildCommonPipelineEnvironment
	var reports mavenBuildReports
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createMavenBuildCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "This step will install the maven project into the local maven repository.",
		Long: `This step will install the maven project into the local maven repository.
It will also prepare jacoco to record the code coverage and
supports ci friendly versioning by flattening the pom before installing.

### build with depedencies from a private repository
if your build has dependencies from a private repository you can include a project settings xml into the source code repository as below (replace the ` + "`" + `<url>` + "`" + `
tag with a valid private repo url).
` + "`" + `` + "`" + `` + "`" + `xml
<?xml version="1.0" encoding="UTF-8"?>
<settings xmlns="http://maven.apache.org/SETTINGS/1.0.0"
xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://maven.apache.org/SETTINGS/1.0.0 https://maven.apache.org/xsd/settings-1.0.0.xsd">
  <servers>
    <server>
        <id>private.repo.id</id>
        <username>${env.PIPER_VAULTCREDENTIAL_USERNAME}</username>
        <password>${env.PIPER_VAULTCREDENTIAL_PASSWORD}</password>
    </server>
  </servers>
</settings>
` + "`" + `` + "`" + `` + "`" + `
` + "`" + `PIPER_VAULTCREDENTIAL_USERNAME` + "`" + ` and ` + "`" + `PIPER_VAULTCREDENTIAL_PASSWORD` + "`" + ` are the username and password for the private repository and are exposed as environment variables that must be present
in the environment where the Piper step runs or alternatively can be created using :
[vault general purpose credentials](../infrastructure/vault.md#using-vault-for-general-purpose-and-test-credentials)

include the below ` + "`" + `<repositories>` + "`" + ` tag in your ` + "`" + `pom.xml` + "`" + ` to reference the ` + "`" + `<server>` + "`" + ` and make sure the values in the ` + "`" + `<id>` + "`" + ` tags match
` + "`" + `` + "`" + `` + "`" + `xml
<repositories>
    <repository>
      <id>private.repo.id</id>
      <url>https://private.repo.com/</url>
    </repository>
</repositories>
` + "`" + `` + "`" + `` + "`" + `

Ensure the following configuration in the Piper ` + "`" + `config.yaml` + "`" + ` to ensure the above settings xml is included and all steps can consume this parameter:

` + "`" + `` + "`" + `` + "`" + `yaml
general:
  projectSettingsFile: <path to the above settings.xml>
` + "`" + `` + "`" + `` + "`" + ``,
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
			log.RegisterSecret(stepConfig.AltDeploymentRepositoryPassword)

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunkClient = &splunk.Splunk{}
				logCollector = &log.CollectorHook{CorrelationID: GeneralConfig.CorrelationID}
				log.RegisterHook(logCollector)
			}

			if err = log.RegisterANSHookIfConfigured(GeneralConfig.CorrelationID); err != nil {
				log.Entry().WithError(err).Warn("failed to set up SAP Alert Notification Service log hook")
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
					splunkClient.Initialize(GeneralConfig.CorrelationID,
						GeneralConfig.HookConfig.SplunkConfig.Dsn,
						GeneralConfig.HookConfig.SplunkConfig.Token,
						GeneralConfig.HookConfig.SplunkConfig.Index,
						GeneralConfig.HookConfig.SplunkConfig.SendLogs)
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
				if len(GeneralConfig.HookConfig.SplunkConfig.ProdCriblEndpoint) > 0 {
					splunkClient.Initialize(GeneralConfig.CorrelationID,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblEndpoint,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblToken,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblIndex,
						GeneralConfig.HookConfig.SplunkConfig.SendLogs)
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetryClient.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			mavenBuild(stepConfig, &stepTelemetryData, &commonPipelineEnvironment)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addMavenBuildFlags(createMavenBuildCmd, &stepConfig)
	return createMavenBuildCmd
}

func addMavenBuildFlags(cmd *cobra.Command, stepConfig *mavenBuildOptions) {
	cmd.Flags().StringVar(&stepConfig.PomPath, "pomPath", `pom.xml`, "Path to the pom file which should be installed including all children.")
	cmd.Flags().StringSliceVar(&stepConfig.Profiles, "profiles", []string{}, "Defines list of maven build profiles to be used.")
	cmd.Flags().BoolVar(&stepConfig.Flatten, "flatten", true, "Defines if the pom files should be flattened to support ci friendly maven versioning.")
	cmd.Flags().BoolVar(&stepConfig.Verify, "verify", false, "Instead of installing the artifact only the verify lifecycle phase is executed.")
	cmd.Flags().StringVar(&stepConfig.ProjectSettingsFile, "projectSettingsFile", os.Getenv("PIPER_projectSettingsFile"), "Path to the mvn settings file that should be used as project settings file.")
	cmd.Flags().StringVar(&stepConfig.GlobalSettingsFile, "globalSettingsFile", os.Getenv("PIPER_globalSettingsFile"), "Path to the mvn settings file that should be used as global settings file.")
	cmd.Flags().StringVar(&stepConfig.M2Path, "m2Path", os.Getenv("PIPER_m2Path"), "Path to the location of the local repository that should be used.")
	cmd.Flags().BoolVar(&stepConfig.LogSuccessfulMavenTransfers, "logSuccessfulMavenTransfers", false, "Configures maven to log successful downloads. This is set to `false` by default to reduce the noise in build logs.")
	cmd.Flags().BoolVar(&stepConfig.CreateBOM, "createBOM", false, "Creates the bill of materials (BOM) using CycloneDX Maven plugin.")
	cmd.Flags().StringVar(&stepConfig.AltDeploymentRepositoryPassword, "altDeploymentRepositoryPassword", os.Getenv("PIPER_altDeploymentRepositoryPassword"), "Password for the alternative deployment repository to which the project artifacts should be deployed ( other than those specified in <distributionManagement> ). This password will be updated in settings.xml . When no settings.xml is provided a new one is created corresponding with <servers> tag")
	cmd.Flags().StringVar(&stepConfig.AltDeploymentRepositoryUser, "altDeploymentRepositoryUser", os.Getenv("PIPER_altDeploymentRepositoryUser"), "User for the alternative deployment repository to which the project artifacts should be deployed ( other than those specified in <distributionManagement> ). This user will be updated in settings.xml . When no settings.xml is provided a new one is created corresponding with <servers> tag")
	cmd.Flags().StringVar(&stepConfig.AltDeploymentRepositoryURL, "altDeploymentRepositoryUrl", os.Getenv("PIPER_altDeploymentRepositoryUrl"), "Url for the alternative deployment repository to which the project artifacts should be deployed ( other than those specified in <distributionManagement> ). This Url will be updated in settings.xml . When no settings.xml is provided a new one is created corresponding with <servers> tag")
	cmd.Flags().StringVar(&stepConfig.AltDeploymentRepositoryID, "altDeploymentRepositoryID", os.Getenv("PIPER_altDeploymentRepositoryID"), "Id for the alternative deployment repository to which the project artifacts should be deployed ( other than those specified in <distributionManagement> ). This id will be updated in settings.xml and will be used as a flag with DaltDeploymentRepository along with mavenAltDeploymentRepositoryUrl during maven deploy . When no settings.xml is provided a new one is created corresponding with <servers> tag")
	cmd.Flags().StringSliceVar(&stepConfig.CustomTLSCertificateLinks, "customTlsCertificateLinks", []string{}, "List of download links to custom TLS certificates. This is required to ensure trusted connections to instances with repositories (like nexus) when publish flag is set to true.")
	cmd.Flags().BoolVar(&stepConfig.Publish, "publish", false, "Configures maven to run the deploy plugin to publish artifacts to a repository.")
	cmd.Flags().StringVar(&stepConfig.JavaCaCertFilePath, "javaCaCertFilePath", os.Getenv("PIPER_javaCaCertFilePath"), "path to the cacerts file used by Java. When maven publish is set to True and customTlsCertificateLinks (to deploy the artifact to a repository with a self signed cert) are provided to trust the self signed certs, Piper will extend the existing Java cacerts to include the new self signed certs. if not provided Piper will search for the cacerts in $JAVA_HOME/jre/lib/security/cacerts")
	cmd.Flags().StringVar(&stepConfig.BuildSettingsInfo, "buildSettingsInfo", os.Getenv("PIPER_buildSettingsInfo"), "build settings info is typically filled by the step automatically to create information about the build settings that were used during the maven build . This information is typically used for compliance related processes.")

}

// retrieve step metadata
func mavenBuildMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "mavenBuild",
			Aliases:     []config.Alias{{Name: "mavenExecute", Deprecated: false}},
			Description: "This step will install the maven project into the local maven repository.",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "altDeploymentRepositoryPasswordId", Description: "Jenkins credentials ID containing the artifact deployment repository password.", Type: "jenkins"},
				},
				Resources: []config.StepResources{
					{Name: "source", Type: "stash"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "pomPath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `pom.xml`,
					},
					{
						Name:        "profiles",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "GENERAL", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "flatten",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     true,
					},
					{
						Name:        "verify",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "projectSettingsFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/projectSettingsFile"}},
						Default:     os.Getenv("PIPER_projectSettingsFile"),
					},
					{
						Name: "globalSettingsFile",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/mavenGlobalSettingsFile",
							},
						},
						Scope:     []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "maven/globalSettingsFile"}},
						Default:   os.Getenv("PIPER_globalSettingsFile"),
					},
					{
						Name:        "m2Path",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/m2Path"}},
						Default:     os.Getenv("PIPER_m2Path"),
					},
					{
						Name:        "logSuccessfulMavenTransfers",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/logSuccessfulMavenTransfers"}},
						Default:     false,
					},
					{
						Name:        "createBOM",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/createBOM"}},
						Default:     false,
					},
					{
						Name: "altDeploymentRepositoryPassword",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/mavenRepositoryPassword",
							},

							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryPassword",
							},

							{
								Name: "altDeploymentRepositoryPasswordId",
								Type: "secret",
							},

							{
								Name:    "altDeploymentRepositoryPasswordFileVaultSecretName",
								Type:    "vaultSecretFile",
								Default: "alt-deployment-repository-passowrd",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_altDeploymentRepositoryPassword"),
					},
					{
						Name: "altDeploymentRepositoryUser",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/mavenRepositoryUsername",
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
						Default:   os.Getenv("PIPER_altDeploymentRepositoryUser"),
					},
					{
						Name: "altDeploymentRepositoryUrl",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/mavenRepositoryURL",
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
						Default:   os.Getenv("PIPER_altDeploymentRepositoryUrl"),
					},
					{
						Name: "altDeploymentRepositoryID",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryId",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_altDeploymentRepositoryID"),
					},
					{
						Name:        "customTlsCertificateLinks",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "publish",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/publish"}},
						Default:     false,
					},
					{
						Name:        "javaCaCertFilePath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/javaCaCertFilePath"}},
						Default:     os.Getenv("PIPER_javaCaCertFilePath"),
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
				},
			},
			Containers: []config.Container{
				{Name: "mvn", Image: "maven:3.6-jdk-8"},
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
							{"filePattern": "**/bom-maven.xml", "type": "sbom"},
							{"filePattern": "**/TEST-*.xml", "type": "junit"},
							{"filePattern": "**/jacoco.xml", "type": "jacoco-coverage"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
