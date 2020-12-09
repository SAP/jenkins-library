// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/spf13/cobra"
)

type sonarExecuteScanOptions struct {
	Instance                  string   `json:"instance,omitempty"`
	ServerURL                 string   `json:"serverUrl,omitempty"`
	Token                     string   `json:"token,omitempty"`
	Organization              string   `json:"organization,omitempty"`
	CustomTLSCertificateLinks []string `json:"customTlsCertificateLinks,omitempty"`
	SonarScannerDownloadURL   string   `json:"sonarScannerDownloadUrl,omitempty"`
	ProjectVersion            string   `json:"projectVersion,omitempty"`
	ProjectKey                string   `json:"projectKey,omitempty"`
	CoverageExclusions        []string `json:"coverageExclusions,omitempty"`
	InferJavaBinaries         bool     `json:"inferJavaBinaries,omitempty"`
	InferJavaLibraries        bool     `json:"inferJavaLibraries,omitempty"`
	Options                   []string `json:"options,omitempty"`
	BranchName                string   `json:"branchName,omitempty"`
	InferBranchName           bool     `json:"inferBranchName,omitempty"`
	ChangeID                  string   `json:"changeId,omitempty"`
	ChangeBranch              string   `json:"changeBranch,omitempty"`
	ChangeTarget              string   `json:"changeTarget,omitempty"`
	PullRequestProvider       string   `json:"pullRequestProvider,omitempty"`
	Owner                     string   `json:"owner,omitempty"`
	Repository                string   `json:"repository,omitempty"`
	GithubToken               string   `json:"githubToken,omitempty"`
	DisableInlineComments     bool     `json:"disableInlineComments,omitempty"`
	LegacyPRHandling          bool     `json:"legacyPRHandling,omitempty"`
	GithubAPIURL              string   `json:"githubApiUrl,omitempty"`
	M2Path                    string   `json:"m2Path,omitempty"`
}

type sonarExecuteScanInflux struct {
	step_data struct {
		fields struct {
			sonar bool
		}
		tags struct {
		}
	}
}

func (i *sonarExecuteScanInflux) persist(path, resourceName string) {
	measurementContent := []struct {
		measurement string
		valType     string
		name        string
		value       interface{}
	}{
		{valType: config.InfluxField, measurement: "step_data", name: "sonar", value: i.step_data.fields.sonar},
	}

	errCount := 0
	for _, metric := range measurementContent {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(metric.measurement, fmt.Sprintf("%vs", metric.valType), metric.name), metric.value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting influx environment.")
			errCount++
		}
	}
	if errCount > 0 {
		log.Entry().Fatal("failed to persist Influx environment")
	}
}

// SonarExecuteScanCommand Executes the Sonar scanner
func SonarExecuteScanCommand() *cobra.Command {
	const STEP_NAME = "sonarExecuteScan"

	metadata := sonarExecuteScanMetadata()
	var stepConfig sonarExecuteScanOptions
	var startTime time.Time
	var influx sonarExecuteScanInflux

	var createSonarExecuteScanCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Executes the Sonar scanner",
		Long:  `The step executes the [sonar-scanner](https://docs.sonarqube.org/display/SCAN/Analyzing+with+SonarQube+Scanner) cli command to scan the defined sources and publish the results to a SonarQube instance.`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err := PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}
			log.RegisterSecret(stepConfig.Token)
			log.RegisterSecret(stepConfig.GithubToken)

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			telemetryData := telemetry.CustomData{}
			telemetryData.ErrorCode = "1"
			handler := func() {
				config.RemoveVaultSecretFiles()
				influx.persist(GeneralConfig.EnvRootPath, "influx")
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetryData.ErrorCategory = log.GetErrorCategory().String()
				telemetry.Send(&telemetryData)
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			sonarExecuteScan(stepConfig, &telemetryData, &influx)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addSonarExecuteScanFlags(createSonarExecuteScanCmd, &stepConfig)
	return createSonarExecuteScanCmd
}

func addSonarExecuteScanFlags(cmd *cobra.Command, stepConfig *sonarExecuteScanOptions) {
	cmd.Flags().StringVar(&stepConfig.Instance, "instance", `SonarCloud`, "Jenkins only: The name of the SonarQube instance defined in the Jenkins settings. DEPRECATED: use serverUrl parameter instead")
	cmd.Flags().StringVar(&stepConfig.ServerURL, "serverUrl", os.Getenv("PIPER_serverUrl"), "The URL to the Sonar backend.")
	cmd.Flags().StringVar(&stepConfig.Token, "token", os.Getenv("PIPER_token"), "Token used to authenticate with the Sonar Server.")
	cmd.Flags().StringVar(&stepConfig.Organization, "organization", os.Getenv("PIPER_organization"), "SonarCloud.io only: Organization that the project will be assigned to in SonarCloud.io.")
	cmd.Flags().StringSliceVar(&stepConfig.CustomTLSCertificateLinks, "customTlsCertificateLinks", []string{}, "List of download links to custom TLS certificates. This is required to ensure trusted connections to instances with custom certificates.")
	cmd.Flags().StringVar(&stepConfig.SonarScannerDownloadURL, "sonarScannerDownloadUrl", `https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-4.5.0.2216-linux.zip`, "URL to the sonar-scanner-cli archive.")
	cmd.Flags().StringVar(&stepConfig.ProjectVersion, "projectVersion", os.Getenv("PIPER_projectVersion"), "The project version that is reported to SonarQube.")
	cmd.Flags().StringVar(&stepConfig.ProjectKey, "projectKey", os.Getenv("PIPER_projectKey"), "The project key identifies the project in SonarQube.")
	cmd.Flags().StringSliceVar(&stepConfig.CoverageExclusions, "coverageExclusions", []string{}, "A list of patterns that should be excluded from the coverage scan.")
	cmd.Flags().BoolVar(&stepConfig.InferJavaBinaries, "inferJavaBinaries", false, "Find the location of generated Java class files in all modules and pass the option `sonar.java.binaries to the sonar tool.")
	cmd.Flags().BoolVar(&stepConfig.InferJavaLibraries, "inferJavaLibraries", false, "If the parameter `m2Path` is configured for the step `mavenExecute` in the general section of the configuration, pass it as option `sonar.java.libraries` to the sonar tool.")
	cmd.Flags().StringSliceVar(&stepConfig.Options, "options", []string{}, "A list of options which are passed to the sonar-scanner.")
	cmd.Flags().StringVar(&stepConfig.BranchName, "branchName", os.Getenv("PIPER_branchName"), "Non-Pull-Request only: Name of the SonarQube branch that should be used to report findings to.")
	cmd.Flags().BoolVar(&stepConfig.InferBranchName, "inferBranchName", false, "Jenkins only: Whether to infer the `branchName` parameter automatically based on the `BRANCH_NAME` environment variable in non-productive runs of the pipeline.")
	cmd.Flags().StringVar(&stepConfig.ChangeID, "changeId", os.Getenv("PIPER_changeId"), "Pull-Request only: The id of the pull-request.")
	cmd.Flags().StringVar(&stepConfig.ChangeBranch, "changeBranch", os.Getenv("PIPER_changeBranch"), "Pull-Request only: The name of the pull-request branch.")
	cmd.Flags().StringVar(&stepConfig.ChangeTarget, "changeTarget", os.Getenv("PIPER_changeTarget"), "Pull-Request only: The name of the base branch.")
	cmd.Flags().StringVar(&stepConfig.PullRequestProvider, "pullRequestProvider", `GitHub`, "Pull-Request only: The scm provider.")
	cmd.Flags().StringVar(&stepConfig.Owner, "owner", os.Getenv("PIPER_owner"), "Pull-Request only: The owner of the scm repository.")
	cmd.Flags().StringVar(&stepConfig.Repository, "repository", os.Getenv("PIPER_repository"), "Pull-Request only: The scm repository.")
	cmd.Flags().StringVar(&stepConfig.GithubToken, "githubToken", os.Getenv("PIPER_githubToken"), "Pull-Request only: Token for Github to set status on the Pull-Request.")
	cmd.Flags().BoolVar(&stepConfig.DisableInlineComments, "disableInlineComments", false, "Pull-Request only: Disables the pull-request decoration with inline comments. DEPRECATED: only supported in SonarQube < 7.2")
	cmd.Flags().BoolVar(&stepConfig.LegacyPRHandling, "legacyPRHandling", false, "Pull-Request only: Activates the pull-request handling using the [GitHub Plugin](https://docs.sonarqube.org/display/PLUG/GitHub+Plugin). DEPRECATED: only supported in SonarQube < 7.2")
	cmd.Flags().StringVar(&stepConfig.GithubAPIURL, "githubApiUrl", `https://api.github.com`, "Pull-Request only: The URL to the Github API. See [GitHub plugin docs](https://docs.sonarqube.org/display/PLUG/GitHub+Plugin#GitHubPlugin-Usage) DEPRECATED: only supported in SonarQube < 7.2")
	cmd.Flags().StringVar(&stepConfig.M2Path, "m2Path", os.Getenv("PIPER_m2Path"), "Path to the location of the local repository that should be used.")

}

// retrieve step metadata
func sonarExecuteScanMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "sonarExecuteScan",
			Aliases:     []config.Alias{},
			Description: "Executes the Sonar scanner",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "sonarTokenCredentialsId", Description: "Jenkins 'Secret text' credentials ID containing the token used to authenticate with the Sonar Server.", Type: "jenkins"},
					{Name: "githubTokenCredentialsId", Description: "Jenkins 'Secret text' credentials ID containing the token used to authenticate with the Github Server.", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "instance",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Default:     `SonarCloud`,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "serverUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_serverUrl"),
						Aliases:     []config.Alias{{Name: "host"}, {Name: "sonarServerUrl"}},
					},
					{
						Name: "token",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "",
								Paths: []string{"$(vaultPath)/sonar", "$(vaultBasePath)/$(vaultPipelineName)/sonar", "$(vaultBasePath)/GROUP-SECRETS/sonar"},
								Type:  "vaultSecret",
							},

							{
								Name: "sonarTokenCredentialsId",
								Type: "secret",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: false,
						Default:   os.Getenv("PIPER_token"),
						Aliases:   []config.Alias{{Name: "sonarToken"}},
					},
					{
						Name:        "organization",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_organization"),
						Aliases:     []config.Alias{},
					},
					{
						Name:        "customTlsCertificateLinks",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Default:     []string{},
						Aliases:     []config.Alias{},
					},
					{
						Name:        "sonarScannerDownloadUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Default:     `https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-4.5.0.2216-linux.zip`,
						Aliases:     []config.Alias{},
					},
					{
						Name: "projectVersion",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "artifactVersion",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Default:   os.Getenv("PIPER_projectVersion"),
						Aliases:   []config.Alias{},
					},
					{
						Name:        "projectKey",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_projectKey"),
						Aliases:     []config.Alias{},
					},
					{
						Name:        "coverageExclusions",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Default:     []string{},
						Aliases:     []config.Alias{},
					},
					{
						Name:        "inferJavaBinaries",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Default:     false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "inferJavaLibraries",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Default:     false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "options",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Default:     []string{},
						Aliases:     []config.Alias{{Name: "sonarProperties"}},
					},
					{
						Name:        "branchName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_branchName"),
						Aliases:     []config.Alias{},
					},
					{
						Name:        "inferBranchName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Default:     false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "changeId",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_changeId"),
						Aliases:     []config.Alias{},
					},
					{
						Name:        "changeBranch",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_changeBranch"),
						Aliases:     []config.Alias{},
					},
					{
						Name:        "changeTarget",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_changeTarget"),
						Aliases:     []config.Alias{},
					},
					{
						Name:        "pullRequestProvider",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Default:     `GitHub`,
						Aliases:     []config.Alias{},
					},
					{
						Name: "owner",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "github/owner",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Default:   os.Getenv("PIPER_owner"),
						Aliases:   []config.Alias{{Name: "githubOrg"}},
					},
					{
						Name: "repository",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "github/repository",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Default:   os.Getenv("PIPER_repository"),
						Aliases:   []config.Alias{{Name: "githubRepo"}},
					},
					{
						Name: "githubToken",
						ResourceRef: []config.ResourceReference{
							{
								Name: "githubTokenCredentialsId",
								Type: "secret",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: false,
						Default:   os.Getenv("PIPER_githubToken"),
						Aliases:   []config.Alias{},
					},
					{
						Name:        "disableInlineComments",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Default:     false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "legacyPRHandling",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Default:     false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "githubApiUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Default:     `https://api.github.com`,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "m2Path",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Default:     os.Getenv("PIPER_m2Path"),
						Aliases:     []config.Alias{{Name: "maven/m2Path"}},
					},
				},
			},
			Containers: []config.Container{
				{Name: "sonar", Image: "sonarsource/sonar-scanner-cli:4.5"},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "influx",
						Type: "influx",
						Parameters: []map[string]interface{}{
							{"Name": "step_data"}, {"fields": []map[string]string{{"name": "sonar"}}},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
