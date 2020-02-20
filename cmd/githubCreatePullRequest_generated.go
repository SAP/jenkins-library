package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/spf13/cobra"
)

type githubCreatePullRequestOptions struct {
	Assignees  []string `json:"assignees,omitempty"`
	Base       string   `json:"base,omitempty"`
	Body       string   `json:"body,omitempty"`
	APIURL     string   `json:"apiUrl,omitempty"`
	Head       string   `json:"head,omitempty"`
	Owner      string   `json:"owner,omitempty"`
	Repository string   `json:"repository,omitempty"`
	ServerURL  string   `json:"serverUrl,omitempty"`
	Title      string   `json:"title,omitempty"`
	Token      string   `json:"token,omitempty"`
	Labels     []string `json:"labels,omitempty"`
}

// GithubCreatePullRequestCommand Create a pull request on GitHub
func GithubCreatePullRequestCommand() *cobra.Command {
	metadata := githubCreatePullRequestMetadata()
	var stepConfig githubCreatePullRequestOptions
	var startTime time.Time

	var createGithubCreatePullRequestCmd = &cobra.Command{
		Use:   "githubCreatePullRequest",
		Short: "Create a pull request on GitHub",
		Long: `This step allows you to create a pull request on Github.

It can for example be used for GitOps scenarios or for scenarios where you want to have a manual confirmation step which is delegated to a GitHub pull request.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			startTime = time.Now()
			log.SetStepName("githubCreatePullRequest")
			log.SetVerbose(GeneralConfig.Verbose)
			return PrepareConfig(cmd, &metadata, "githubCreatePullRequest", &stepConfig, config.OpenPiperFile)
		},
		Run: func(cmd *cobra.Command, args []string) {
			telemetryData := telemetry.CustomData{}
			telemetryData.ErrorCode = "1"
			handler := func() {
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetry.Send(&telemetryData)
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, "githubCreatePullRequest")
			githubCreatePullRequest(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
		},
	}

	addGithubCreatePullRequestFlags(createGithubCreatePullRequestCmd, &stepConfig)
	return createGithubCreatePullRequestCmd
}

func addGithubCreatePullRequestFlags(cmd *cobra.Command, stepConfig *githubCreatePullRequestOptions) {
	cmd.Flags().StringSliceVar(&stepConfig.Assignees, "assignees", []string{}, "Login names of users to which the PR should be assigned to.")
	cmd.Flags().StringVar(&stepConfig.Base, "base", os.Getenv("PIPER_base"), "The name of the branch you want the changes pulled into.")
	cmd.Flags().StringVar(&stepConfig.Body, "body", os.Getenv("PIPER_body"), "The description text of the pull request in markdown format.")
	cmd.Flags().StringVar(&stepConfig.APIURL, "apiUrl", "https://api.github.com", "Set the GitHub API url.")
	cmd.Flags().StringVar(&stepConfig.Head, "head", os.Getenv("PIPER_head"), "The name of the branch where your changes are implemented.")
	cmd.Flags().StringVar(&stepConfig.Owner, "owner", os.Getenv("PIPER_owner"), "Set the GitHub organization.")
	cmd.Flags().StringVar(&stepConfig.Repository, "repository", os.Getenv("PIPER_repository"), "Set the GitHub repository.")
	cmd.Flags().StringVar(&stepConfig.ServerURL, "serverUrl", "https://github.com", "GitHub server url for end-user access.")
	cmd.Flags().StringVar(&stepConfig.Title, "title", os.Getenv("PIPER_title"), "Title of the pull request.")
	cmd.Flags().StringVar(&stepConfig.Token, "token", os.Getenv("PIPER_token"), "GitHub personal access token as per https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line")
	cmd.Flags().StringSliceVar(&stepConfig.Labels, "labels", []string{}, "Labels to be added to the pull request.")

	cmd.MarkFlagRequired("base")
	cmd.MarkFlagRequired("body")
	cmd.MarkFlagRequired("apiUrl")
	cmd.MarkFlagRequired("head")
	cmd.MarkFlagRequired("owner")
	cmd.MarkFlagRequired("repository")
	cmd.MarkFlagRequired("serverUrl")
	cmd.MarkFlagRequired("title")
	cmd.MarkFlagRequired("token")
}

// retrieve step metadata
func githubCreatePullRequestMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "assignees",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "base",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "body",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "apiUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "githubApiUrl"}},
					},
					{
						Name:        "head",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "owner",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "githubOrg"}},
					},
					{
						Name:        "repository",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "githubRepo"}},
					},
					{
						Name:        "serverUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "githubServerUrl"}},
					},
					{
						Name:        "title",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "token",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "githubToken"}},
					},
					{
						Name:        "labels",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
				},
			},
		},
	}
	return theMetaData
}
