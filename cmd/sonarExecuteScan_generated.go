package cmd

import (
	"os"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/spf13/cobra"
)

type sonarExecuteScanOptions struct {
	Host                      string `json:"host,omitempty"`
	Token                     string `json:"token,omitempty"`
	GithubToken               string `json:"githubToken,omitempty"`
	ChangeID                  string `json:"changeId,omitempty"`
	DisableInlineComments     bool   `json:"disableInlineComments,omitempty"`
	LegacyPRHandling          bool   `json:"legacyPRHandling,omitempty"`
	Owner                     string `json:"owner,omitempty"`
	Repository                string `json:"repository,omitempty"`
	GithubAPIURL              string `json:"githubApiUrl,omitempty"`
	Organization              string `json:"organization,omitempty"`
	Options                   string `json:"options,omitempty"`
	CustomTLSCertificateLinks string `json:"customTlsCertificateLinks,omitempty"`
	PullRequestProvider       string `json:"pullRequestProvider,omitempty"`
	SonarScannerDownloadURL   string `json:"sonarScannerDownloadUrl,omitempty"`
	ProjectVersion            string `json:"projectVersion,omitempty"`
}

var mySonarExecuteScanOptions sonarExecuteScanOptions
var sonarExecuteScanStepConfigJSON string

// SonarExecuteScanCommand Executes the Sonar scanner
func SonarExecuteScanCommand() *cobra.Command {
	metadata := sonarExecuteScanMetadata()
	var createSonarExecuteScanCmd = &cobra.Command{
		Use:   "sonarExecuteScan",
		Short: "Executes the Sonar scanner",
		Long:  `The step executes the [sonar-scanner](https://docs.sonarqube.org/display/SCAN/Analyzing+with+SonarQube+Scanner) cli command to scan the defined sources and publish the results to a SonarQube instance.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			log.SetStepName("sonarExecuteScan")
			log.SetVerbose(GeneralConfig.Verbose)
			return PrepareConfig(cmd, &metadata, "sonarExecuteScan", &mySonarExecuteScanOptions, config.OpenPiperFile)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return sonarExecuteScan(mySonarExecuteScanOptions)
		},
	}

	addSonarExecuteScanFlags(createSonarExecuteScanCmd)
	return createSonarExecuteScanCmd
}

func addSonarExecuteScanFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&mySonarExecuteScanOptions.Host, "host", os.Getenv("PIPER_host"), "")
	cmd.Flags().StringVar(&mySonarExecuteScanOptions.Token, "token", os.Getenv("PIPER_token"), "")
	cmd.Flags().StringVar(&mySonarExecuteScanOptions.GithubToken, "githubToken", os.Getenv("PIPER_githubToken"), "")
	cmd.Flags().StringVar(&mySonarExecuteScanOptions.ChangeID, "changeId", os.Getenv("PIPER_changeId"), "")
	cmd.Flags().BoolVar(&mySonarExecuteScanOptions.DisableInlineComments, "disableInlineComments", false, "Pull-Request voting only: Disables the pull-request decoration with inline comments. deprecated: only supported in < 7.2")
	cmd.Flags().BoolVar(&mySonarExecuteScanOptions.LegacyPRHandling, "legacyPRHandling", false, "Pull-Request voting only: Activates the pull-request handling using the [GitHub Plugin](https://docs.sonarqube.org/display/PLUG/GitHub+Plugin) (deprecated). deprecated: only supported in < 7.2")
	cmd.Flags().StringVar(&mySonarExecuteScanOptions.Owner, "owner", os.Getenv("PIPER_owner"), "Pull-Request voting only: The Github organization. @default: `commonPipelineEnvironment.getGithubOrg()`")
	cmd.Flags().StringVar(&mySonarExecuteScanOptions.Repository, "repository", os.Getenv("PIPER_repository"), "Pull-Request voting only: The Github repository. @default: `commonPipelineEnvironment.getGithubRepo()`")
	cmd.Flags().StringVar(&mySonarExecuteScanOptions.GithubAPIURL, "githubApiUrl", "https://api.github.com", "Pull-Request voting only: The URL to the Github API. see [GitHub plugin docs](https://docs.sonarqube.org/display/PLUG/GitHub+Plugin#GitHubPlugin-Usage) deprecated: only supported in < 7.2")
	cmd.Flags().StringVar(&mySonarExecuteScanOptions.Organization, "organization", os.Getenv("PIPER_organization"), "Organization that the project will be assigned to in SonarCloud.io.")
	cmd.Flags().StringVar(&mySonarExecuteScanOptions.Options, "options", "[]", "A list of options which are passed to the `sonar-scanner`.")
	cmd.Flags().StringVar(&mySonarExecuteScanOptions.CustomTLSCertificateLinks, "customTlsCertificateLinks", "[]", "List containing download links of custom TLS certificates. This is required to ensure trusted connections to instances with custom certificates.")
	cmd.Flags().StringVar(&mySonarExecuteScanOptions.PullRequestProvider, "pullRequestProvider", "GitHub", "")
	cmd.Flags().StringVar(&mySonarExecuteScanOptions.SonarScannerDownloadURL, "sonarScannerDownloadUrl", "https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-4.2.0.1873-linux.zip", "")
	cmd.Flags().StringVar(&mySonarExecuteScanOptions.ProjectVersion, "projectVersion", os.Getenv("PIPER_projectVersion"), "The project version that is reported to SonarQube. @default: major number of `commonPipelineEnvironment.getArtifactVersion()`")

}

// retrieve step metadata
func sonarExecuteScanMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:      "host",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "token",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "githubToken",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "changeId",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "disableInlineComments",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "bool",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "legacyPRHandling",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "bool",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "owner",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "githubOrg"}},
					},
					{
						Name:      "repository",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "githubRepo"}},
					},
					{
						Name:      "githubApiUrl",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "organization",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "options",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "customTlsCertificateLinks",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "pullRequestProvider",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "sonarScannerDownloadUrl",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "projectVersion",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
				},
			},
		},
	}
	return theMetaData
}
