package orchestrator

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type GitHubActionsConfigProvider struct {
	client     piperHttp.Client
	actionsURL string
	runData    run
	jobs       []job
	currentJob *job
}

type run struct {
	Status    string    `json:"status"`
	StartedAt time.Time `json:"run_started_at"`
	HtmlURL   string    `json:"html_url"`
}

type job struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	HtmlURL string `json:"html_url"`
}

type stagesID struct {
	Jobs []job `json:"jobs"`
}

type fullLog struct {
	sync.Mutex
	b [][]byte
}

var httpHeaders = http.Header{
	"Accept": {"application/vnd.github+json"},
}

// InitOrchestratorProvider initializes http client for GitHubActionsDevopsConfigProvider
func (g *GitHubActionsConfigProvider) InitOrchestratorProvider(settings *OrchestratorSettings) {
	g.client.SetOptions(piperHttp.ClientOptions{
		Password:         settings.GitHubToken,
		MaxRetries:       3,
		TransportTimeout: time.Second * 10,
	})

	g.actionsURL = actionsURL()
	g.fetchRunData()

	if err := g.fetchJobs(); err != nil {
		// Since InitOrchestratorProvider() does not return an error and changing a public method
		// is currently undesired, log the error here and return.
		log.Entry().Errorf("failed to fetch jobs: %s", err)
		g.jobs = []job{}
		return
	}

	log.Entry().Debug("Successfully initialized GitHubActions config provider")
}

func (g *GitHubActionsConfigProvider) OrchestratorVersion() string {
	log.Entry().Debugf("OrchestratorVersion() for GitHub Actions is not applicable.")
	return "n/a"
}

func (g *GitHubActionsConfigProvider) OrchestratorType() string {
	return "GitHubActions"
}

// GetBuildStatus returns current run status
func (g *GitHubActionsConfigProvider) GetBuildStatus() string {
	switch g.runData.Status {
	case "success":
		return BuildStatusSuccess
	case "cancelled":
		return BuildStatusAborted
	case "in_progress":
		return BuildStatusInProgress
	default:
		return BuildStatusFailure
	}
}

// GetLog returns the whole logfile for the current pipeline run
func (g *GitHubActionsConfigProvider) GetLog() ([]byte, error) {
	if err := g.fetchJobs(); err != nil {
		return nil, err
	}
	// Ignore the last stage (job) as it is not possible in GitHub to fetch logs for a running job.
	jobs := g.jobs[:len(g.jobs)-1]

	fullLogs := fullLog{
		b: make([][]byte, len(jobs)),
	}
	ctx := context.Background()
	sem := semaphore.NewWeighted(10)
	wg := errgroup.Group{}
	for i := range jobs {
		i := i // https://golang.org/doc/faq#closures_and_goroutines
		if err := sem.Acquire(ctx, 1); err != nil {
			return nil, fmt.Errorf("failed to acquire semaphore: %w", err)
		}
		wg.Go(func() error {
			defer sem.Release(1)
			resp, err := g.client.GetRequest(fmt.Sprintf("%s/jobs/%d/logs", g.actionsURL, jobs[i].ID), httpHeaders, nil)
			if err != nil {
				return fmt.Errorf("failed to get API data: %w", err)
			}

			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body: %w", err)
			}
			defer resp.Body.Close()
			fullLogs.Lock()
			defer fullLogs.Unlock()
			fullLogs.b[i] = b

			return nil
		})
	}
	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	return bytes.Join(fullLogs.b, []byte("")), nil
}

// GetBuildID returns current run ID
func (g *GitHubActionsConfigProvider) GetBuildID() string {
	return getEnv("GITHUB_RUN_ID", "n/a")
}

func (g *GitHubActionsConfigProvider) GetChangeSet() []ChangeSet {
	log.Entry().Debug("GetChangeSet for GitHubActions not implemented")
	return []ChangeSet{}
}

// GetPipelineStartTime returns the pipeline start time in UTC
func (g *GitHubActionsConfigProvider) GetPipelineStartTime() time.Time {
	return g.runData.StartedAt.UTC()
}

// GetStageName returns the human-readable name given to a stage.
func (g *GitHubActionsConfigProvider) GetStageName() string {
	if g.currentJob == nil {
		return "n/a"
	}

	return g.currentJob.Name
}

// GetBuildReason returns the reason of workflow trigger.
// BuildReasons are unified with AzureDevOps build reasons, see
// https://docs.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops&tabs=yaml#build-variables-devops-services
func (g *GitHubActionsConfigProvider) GetBuildReason() string {
	switch getEnv("GITHUB_REF", "") {
	case "workflow_dispatch":
		return BuildReasonManual
	case "schedule":
		return BuildReasonSchedule
	case "pull_request":
		return BuildReasonPullRequest
	case "workflow_call":
		return BuildReasonResourceTrigger
	case "push":
		return BuildReasonIndividualCI
	default:
		return BuildReasonUnknown
	}
}

// GetBranch returns the source branch name, e.g. main
func (g *GitHubActionsConfigProvider) GetBranch() string {
	return getEnv("GITHUB_REF_NAME", "n/a")
}

// GetReference return the git reference. For example, refs/heads/your_branch_name
func (g *GitHubActionsConfigProvider) GetReference() string {
	return getEnv("GITHUB_REF", "n/a")
}

// GetBuildURL returns the builds URL. For example, https://github.com/SAP/jenkins-library/actions/runs/5815297487
func (g *GitHubActionsConfigProvider) GetBuildURL() string {
	return g.runData.HtmlURL
}

// GetJobURL returns the current job HTML URL (not API URL).
// For example, https://github.com/SAP/jenkins-library/actions/runs/123456/jobs/7654321
func (g *GitHubActionsConfigProvider) GetJobURL() string {
	if g.currentJob == nil {
		return "n/a"
	}

	return g.currentJob.HtmlURL
}

// GetJobName returns the current workflow name. For example, "Piper workflow"
func (g *GitHubActionsConfigProvider) GetJobName() string {
	return getEnv("GITHUB_WORKFLOW", "unknown")
}

// GetCommit returns the commit SHA that triggered the workflow. For example, ffac537e6cbbf934b08745a378932722df287a53
func (g *GitHubActionsConfigProvider) GetCommit() string {
	return getEnv("GITHUB_SHA", "n/a")
}

// GetRepoURL returns full url to repository. For example, https://github.com/SAP/jenkins-library
func (g *GitHubActionsConfigProvider) GetRepoURL() string {
	return getEnv("GITHUB_SERVER_URL", "n/a") + "/" + getEnv("GITHUB_REPOSITORY", "n/a")
}

// GetPullRequestConfig returns pull request configuration
func (g *GitHubActionsConfigProvider) GetPullRequestConfig() PullRequestConfig {
	// See https://docs.github.com/en/enterprise-server@3.6/actions/learn-github-actions/variables#default-environment-variables
	githubRef := getEnv("GITHUB_REF", "n/a")
	prNumber := strings.TrimSuffix(strings.TrimPrefix(githubRef, "refs/pull/"), "/merge")
	return PullRequestConfig{
		Branch: getEnv("GITHUB_HEAD_REF", "n/a"),
		Base:   getEnv("GITHUB_BASE_REF", "n/a"),
		Key:    prNumber,
	}
}

// IsPullRequest indicates whether the current build is triggered by a PR
func (g *GitHubActionsConfigProvider) IsPullRequest() bool {
	return truthy("GITHUB_HEAD_REF")
}

func isGitHubActions() bool {
	envVars := []string{"GITHUB_ACTION", "GITHUB_ACTIONS"}
	return areIndicatingEnvVarsSet(envVars)
}

// actionsURL returns URL to actions resource. For example,
// https://api.github.com/repos/SAP/jenkins-library/actions              - if it's github.com
// https://github.tools.sap/api/v3/repos/project-piper/sap-piper/actions - if it's GitHub Enterprise
func actionsURL() string {
	return fmt.Sprintf("%s/repos/%s/actions", getEnv("GITHUB_API_URL", ""), getEnv("GITHUB_REPOSITORY", ""))
}

func (g *GitHubActionsConfigProvider) fetchRunData() {
	url := fmt.Sprintf("%s/runs/%s", g.actionsURL, getEnv("GITHUB_RUN_ID", ""))
	resp, err := g.client.GetRequest(url, httpHeaders, nil)
	if err != nil || resp.StatusCode != 200 {
		log.Entry().Errorf("failed to get API data: %s", err)
		return
	}

	err = piperHttp.ParseHTTPResponseBodyJSON(resp, &g.runData)
	if err != nil {
		log.Entry().Errorf("failed to parse JSON data: %s", err)
		return
	}
}

func (g *GitHubActionsConfigProvider) fetchJobs() error {
	if len(g.jobs) != 0 {
		// already fetched once
		return nil
	}

	url := fmt.Sprintf("%s/runs/%s/jobs", g.actionsURL, getEnv("GITHUB_RUN_ID", ""))
	resp, err := g.client.GetRequest(url, httpHeaders, nil)
	if err != nil {
		return fmt.Errorf("failed to get API data: %w", err)
	}

	var result struct {
		Jobs []job `json:"jobs"`
	}
	err = piperHttp.ParseHTTPResponseBodyJSON(resp, &result)
	if err != nil {
		return fmt.Errorf("failed to parse JSON data: %w", err)
	}

	if len(result.Jobs) == 0 {
		return fmt.Errorf("no jobs found in response")
	}
	g.jobs = result.Jobs

	return nil
}

func (g *GitHubActionsConfigProvider) guessCurrentJob() {
	for _, j := range g.jobs {
		if j.Name == getEnv("GITHUB_JOB", "unknown") {
			g.currentJob = &j
		}
	}
}
