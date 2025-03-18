package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/google/go-github/v45/github"

	"github.com/SAP/jenkins-library/pkg/command"
	piperGithub "github.com/SAP/jenkins-library/pkg/github"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type githubClientWrapper struct {
	client *github.Client
}

func (gcw *githubClientWrapper) GetRepo(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error) {
	return gcw.client.Repositories.Get(ctx, owner, repo)
}

func (gcw *githubClientWrapper) ListAlertsForRepo(ctx context.Context, owner, repo string, opts *github.SecretScanningAlertListOptions) ([]*github.SecretScanningAlert, *github.Response, error) {
	return gcw.client.SecretScanning.ListAlertsForRepo(ctx, owner, repo, opts)
}

type githubSecretScanningReportUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)
	Create(name string) (io.ReadWriteCloser, error)

	GetRepo(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error)
	ListAlertsForRepo(ctx context.Context, owner, repo string, opts *github.SecretScanningAlertListOptions) ([]*github.SecretScanningAlert, *github.Response, error)
}

type githubSecretScanningReportUtilsBundle struct {
	*command.Command
	*piperutils.Files
	*githubClientWrapper
}

func newGithubSecretScanningReportUtils(ghClient *github.Client) githubSecretScanningReportUtils {
	utils := githubSecretScanningReportUtilsBundle{
		Command:             &command.Command{},
		Files:               &piperutils.Files{},
		githubClientWrapper: &githubClientWrapper{ghClient},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func githubSecretScanningReport(config githubSecretScanningReportOptions, telemetryData *telemetry.CustomData) {
	ctx, ghClient, err := piperGithub.
		NewClientBuilder(config.Token, config.APIURL).
		Build()

	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to get GitHub client.")
	}

	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newGithubSecretScanningReportUtils(ghClient)

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	if err = runGithubSecretScanningReport(ctx, &config, telemetryData, utils); err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runGithubSecretScanningReport(ctx context.Context, config *githubSecretScanningReportOptions, telemetryData *telemetry.CustomData, utils githubSecretScanningReportUtils) error {
	log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")

	report, err := generateGithubSecretScanningReport(ctx, config, utils)

	if err != nil {
		return fmt.Errorf("couldn't generate the github secret scanning report: %w", err)
	}

	reportFile, err := utils.Create("github-secretscanning.report.json")
	if err != nil {
		return fmt.Errorf("couldn't create 'secretscan.json': %w", err)
	}

	defer reportFile.Close()

	if err = json.NewEncoder(reportFile).Encode(report); err != nil {
		return fmt.Errorf("couldn't save the github secret scanning report: %w", err)
	}

	return nil
}

// generateGithubSecretScanningReport generates a secret scanning report for a specified GitHub repository.
// It retrieves all open secret scanning alerts, compiles their details including secret type, state,
// and locations, and returns a structured report. If any error occurs during the retrieval process,
// the function returns an error.
func generateGithubSecretScanningReport(ctx context.Context, config *githubSecretScanningReportOptions, utils githubSecretScanningReportUtils) (*githubSecretScanningReportType, error) {
	repo, _, err := utils.GetRepo(ctx, config.Owner, config.Repository)

	if err != nil {
		return nil, err
	}

	// query github for alerts
	secretAlerts, _, err := utils.ListAlertsForRepo(ctx, config.Owner, config.Repository, &github.SecretScanningAlertListOptions{})
	if err != nil {
		return nil, err
	}

	alertsTotal := len(secretAlerts)
	alertsAudited := 0

	// query actual finding locations
	for _, alert := range secretAlerts {
		if alert.State != nil && *alert.State == "resolved" {
			alertsAudited = alertsAudited + 1
		}
	}

	report := &githubSecretScanningReportType{
		ToolName: "GitHubSecretScanning",
		Findings: []githubSecretScanningFinding{
			githubSecretScanningFinding{
				ClassificationName: "Audit All",
				Total:              alertsTotal,
				Audited:            alertsAudited,
			},
		},
	}

	if repo.HTMLURL != nil {
		report.RepositoryURL = *repo.HTMLURL
		report.SecretScanningURL = fmt.Sprintf("%s/security/secret-scanning", *repo.HTMLURL)
	}

	return report, nil
}

type githubSecretScanningReportType struct {
	ToolName          string                        `json:"toolName"`
	RepositoryURL     string                        `json:"repositoryUrl,omitempty"`
	SecretScanningURL string                        `json:"secretScanningUrl,omitempty"`
	Findings          []githubSecretScanningFinding `json:"findings"`
}

type githubSecretScanningFinding struct {
	ClassificationName string `json:"classificationName"`
	Total              int    `json:"total"`
	Audited            int    `json:"audited"`
}
