package cmd

import (
	"os"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/spf13/cobra"
)

type githubPublishReleaseOptions struct {
	AddClosedIssues       bool     `json:"addClosedIssues,omitempty"`
	AddDeltaToLastRelease bool     `json:"addDeltaToLastRelease,omitempty"`
	AssetPath             string   `json:"assetPath,omitempty"`
	Commitish             string   `json:"commitish,omitempty"`
	ExcludeLabels         []string `json:"excludeLabels,omitempty"`
	APIURL                string   `json:"apiUrl,omitempty"`
	Owner                 string   `json:"owner,omitempty"`
	Repository            string   `json:"repository,omitempty"`
	ServerURL             string   `json:"serverUrl,omitempty"`
	Token                 string   `json:"token,omitempty"`
	UploadURL             string   `json:"uploadUrl,omitempty"`
	Labels                []string `json:"labels,omitempty"`
	ReleaseBodyHeader     string   `json:"releaseBodyHeader,omitempty"`
	UpdateAsset           bool     `json:"updateAsset,omitempty"`
	Version               string   `json:"version,omitempty"`
}

var myGithubPublishReleaseOptions githubPublishReleaseOptions
var githubPublishReleaseStepConfigJSON string

// GithubPublishReleaseCommand Publish a release in GitHub
func GithubPublishReleaseCommand() *cobra.Command {
	metadata := githubPublishReleaseMetadata()
	var createGithubPublishReleaseCmd = &cobra.Command{
		Use:   "githubPublishRelease",
		Short: "Publish a release in GitHub",
		Long: `This step creates a tag in your GitHub repository together with a release.
The release can be filled with text plus additional information like:

* Closed pull request since last release
* Closed issues since last release
* Link to delta information showing all commits since last release

The result looks like

![Example release](../images/githubRelease.png)`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			log.SetStepName("githubPublishRelease")
			log.SetVerbose(GeneralConfig.verbose)
			return PrepareConfig(cmd, &metadata, "githubPublishRelease", &myGithubPublishReleaseOptions, openPiperFile)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return githubPublishRelease(myGithubPublishReleaseOptions)
		},
	}

	addGithubPublishReleaseFlags(createGithubPublishReleaseCmd)
	return createGithubPublishReleaseCmd
}

func addGithubPublishReleaseFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&myGithubPublishReleaseOptions.AddClosedIssues, "addClosedIssues", false, "If set to `true`, closed issues and merged pull-requests since the last release will added below the `releaseBodyHeader`")
	cmd.Flags().BoolVar(&myGithubPublishReleaseOptions.AddDeltaToLastRelease, "addDeltaToLastRelease", false, "If set to `true`, a link will be added to the relese information that brings up all commits since the last release.")
	cmd.Flags().StringVar(&myGithubPublishReleaseOptions.AssetPath, "assetPath", os.Getenv("PIPER_assetPath"), "Path to a release asset which should be uploaded to the list of release assets.")
	cmd.Flags().StringVar(&myGithubPublishReleaseOptions.Commitish, "commitish", "master", "Target git commitish for the release")
	cmd.Flags().StringSliceVar(&myGithubPublishReleaseOptions.ExcludeLabels, "excludeLabels", []string{}, "Allows to exclude issues with dedicated list of labels.")
	cmd.Flags().StringVar(&myGithubPublishReleaseOptions.APIURL, "apiUrl", "https://api.github.com", "Set the GitHub API url.")
	cmd.Flags().StringVar(&myGithubPublishReleaseOptions.Owner, "owner", os.Getenv("PIPER_owner"), "Set the GitHub organization.")
	cmd.Flags().StringVar(&myGithubPublishReleaseOptions.Repository, "repository", os.Getenv("PIPER_repository"), "Set the GitHub repository.")
	cmd.Flags().StringVar(&myGithubPublishReleaseOptions.ServerURL, "serverUrl", "https://github.com", "GitHub server url for end-user access.")
	cmd.Flags().StringVar(&myGithubPublishReleaseOptions.Token, "token", os.Getenv("PIPER_token"), "GitHub personal access token as per https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line")
	cmd.Flags().StringVar(&myGithubPublishReleaseOptions.UploadURL, "uploadUrl", "https://uploads.github.com", "Set the GitHub API url.")
	cmd.Flags().StringSliceVar(&myGithubPublishReleaseOptions.Labels, "labels", []string{}, "Labels to include in issue search.")
	cmd.Flags().StringVar(&myGithubPublishReleaseOptions.ReleaseBodyHeader, "releaseBodyHeader", os.Getenv("PIPER_releaseBodyHeader"), "Content which will appear for the release.")
	cmd.Flags().BoolVar(&myGithubPublishReleaseOptions.UpdateAsset, "updateAsset", false, "Specify if a release asset should be updated only.")
	cmd.Flags().StringVar(&myGithubPublishReleaseOptions.Version, "version", os.Getenv("PIPER_version"), "Define the version number which will be written as tag as well as release name.")

	cmd.MarkFlagRequired("apiUrl")
	cmd.MarkFlagRequired("owner")
	cmd.MarkFlagRequired("repository")
	cmd.MarkFlagRequired("serverUrl")
	cmd.MarkFlagRequired("token")
	cmd.MarkFlagRequired("uploadUrl")
	cmd.MarkFlagRequired("version")
}

// retrieve step metadata
func githubPublishReleaseMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:      "addClosedIssues",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "bool",
						Mandatory: false,
					},
					{
						Name:      "addDeltaToLastRelease",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "bool",
						Mandatory: false,
					},
					{
						Name:      "assetPath",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
					},
					{
						Name:      "commitish",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
					},
					{
						Name:      "excludeLabels",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "[]string",
						Mandatory: false,
					},
					{
						Name:      "apiUrl",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
					},
					{
						Name:      "owner",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
					},
					{
						Name:      "repository",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
					},
					{
						Name:      "serverUrl",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
					},
					{
						Name:      "token",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
					},
					{
						Name:      "uploadUrl",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
					},
					{
						Name:      "labels",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "[]string",
						Mandatory: false,
					},
					{
						Name:      "releaseBodyHeader",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
					},
					{
						Name:      "updateAsset",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "bool",
						Mandatory: false,
					},
					{
						Name:      "version",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
					},
				},
			},
		},
	}
	return theMetaData
}
