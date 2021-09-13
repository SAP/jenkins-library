// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
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
}

// MavenBuildCommand This step will install the maven project into the local maven repository.
func MavenBuildCommand() *cobra.Command {
	const STEP_NAME = "mavenBuild"

	metadata := mavenBuildMetadata()
	var stepConfig mavenBuildOptions
	var startTime time.Time
	var logCollector *log.CollectorHook

	var createMavenBuildCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "This step will install the maven project into the local maven repository.",
		Long: `This step will install the maven project into the local maven repository.
It will also prepare jacoco to record the code coverage and
supports ci friendly versioning by flattening the pom before installing.`,
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
				logCollector = &log.CollectorHook{CorrelationID: GeneralConfig.CorrelationID}
				log.RegisterHook(logCollector)
			}

			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			telemetryData := telemetry.CustomData{}
			telemetryData.ErrorCode = "1"
			handler := func() {
				config.RemoveVaultSecretFiles()
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetryData.ErrorCategory = log.GetErrorCategory().String()
				telemetry.Send(&telemetryData)
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunk.Send(&telemetryData, logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunk.Initialize(GeneralConfig.CorrelationID,
					GeneralConfig.HookConfig.SplunkConfig.Dsn,
					GeneralConfig.HookConfig.SplunkConfig.Token,
					GeneralConfig.HookConfig.SplunkConfig.Index,
					GeneralConfig.HookConfig.SplunkConfig.SendLogs)
			}
			mavenBuild(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
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
				Parameters: []config.StepParameters{
					{
						Name:        "pomPath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STEPS"},
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
						Scope:       []string{"PARAMETERS"},
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
								Param: "custom/repositoryPassword",
							},

							{
								Name: "altDeploymentRepositoryPasswordId",
								Type: "secret",
							},

							{
								Name:  "",
								Paths: []string{"$(vaultPath)/alt-deployment-repository-passowrd", "$(vaultBasePath)/$(vaultPipelineName)/alt-deployment-repository-passowrd", "$(vaultBasePath)/GROUP-SECRETS/alt-deployment-repository-passowrd"},
								Type:  "vaultSecretFile",
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
				},
			},
			Containers: []config.Container{
				{Name: "mvn", Image: "maven:3.6-jdk-8"},
			},
		},
	}
	return theMetaData
}
