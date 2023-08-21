package orchestrator

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"

	"golang.org/x/sync/errgroup"
)

type GitHubActionsConfigProvider struct {
	client      piperHttp.Client
	runData     run
	jobs        []job
	jobsFetched bool
	currentJob  job
}

type run struct {
	fetched   bool
	Status    string    `json:"status"`
	StartedAt time.Time `json:"run_started_at"`
}

type job struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	HtmlURL string `json:"html_url"`
}

type fullLog struct {
	sync.Mutex
	b [][]byte
}

var httpHeaders = http.Header{
	"Accept":               {"application/vnd.github+json"},
	"X-GitHub-Api-Version": {"2022-11-28"},
}

// InitOrchestratorProvider initializes http client for GitHubActionsDevopsConfigProvider
func (g *GitHubActionsConfigProvider) InitOrchestratorProvider(settings *OrchestratorSettings) {
	g.client.SetOptions(piperHttp.ClientOptions{
		Token:            "Bearer " + settings.GitHubToken,
		MaxRetries:       3,
		TransportTimeout: time.Second * 10,
	})

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
	g.fetchRunData()
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

	fullLogs := fullLog{b: make([][]byte, len(jobs))}
	wg := errgroup.Group{}
	wg.SetLimit(10)
	for i := range jobs {
		i := i // https://golang.org/doc/faq#closures_and_goroutines
		wg.Go(func() error {
			resp, err := g.client.GetRequest(fmt.Sprintf("%s/jobs/%d/logs", actionsURL(), jobs[i].ID), httpHeaders, nil)
			if err != nil {
				return fmt.Errorf("failed to get API data: %w", err)
			}
			defer resp.Body.Close()

			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body: %w", err)
			}

			fullLogs.Lock()
			fullLogs.b[i] = b
			fullLogs.Unlock()

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
	g.fetchRunData()
	return g.runData.StartedAt.UTC()
}

// GetStageName returns the human-readable name given to a stage.
func (g *GitHubActionsConfigProvider) GetStageName() string {
	return getEnv("GITHUB_JOB", "unknown")
}

// GetBuildReason returns the reason of workflow trigger.
// BuildReasons are unified with AzureDevOps build reasons, see
// https://docs.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops&tabs=yaml#build-variables-devops-services
func (g *GitHubActionsConfigProvider) GetBuildReason() string {
	switch getEnv("GITHUB_EVENT_NAME", "") {
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
	return g.GetRepoURL() + "/actions/runs/" + getEnv("GITHUB_RUN_ID", "n/a")
}

// GetJobURL returns the current job HTML URL (not API URL).
// For example, https://github.com/SAP/jenkins-library/actions/runs/123456/jobs/7654321
func (g *GitHubActionsConfigProvider) GetJobURL() string {
	// We need to query the GitHub API here because the environment variable GITHUB_JOB returns
	// the name of the job, not a numeric ID (which we need to form the URL)
	g.guessCurrentJob()
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
	return getEnv("GITHUB_API_URL", "") + "/repos/" + getEnv("GITHUB_REPOSITORY", "") + "/actions"
}

func (g *GitHubActionsConfigProvider) fetchRunData() {
	if g.runData.fetched {
		return
	}

	url := fmt.Sprintf("%s/runs/%s", actionsURL(), getEnv("GITHUB_RUN_ID", ""))
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
	g.runData.fetched = true
}

func (g *GitHubActionsConfigProvider) fetchJobs() error {
	if g.jobsFetched {
		return nil
	}

	url := fmt.Sprintf("%s/runs/%s/jobs", actionsURL(), getEnv("GITHUB_RUN_ID", ""))
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
	g.jobsFetched = true

	return nil
}

func (g *GitHubActionsConfigProvider) guessCurrentJob() {
	// check if the current job has already been guessed
	if g.currentJob.ID != 0 {
		return
	}

	// fetch jobs if they haven't been fetched yet
	if err := g.fetchJobs(); err != nil {
		log.Entry().Errorf("failed to fetch jobs: %s", err)
		g.jobs = []job{}
		return
	}

	targetJobName := getEnv("GITHUB_JOB", "unknown")
	log.Entry().Debugf("looking for job '%s' in jobs list: %v", targetJobName, g.jobs)
	for _, j := range g.jobs {
		// j.Name may be something like "piper / Init / Init"
		// but GITHUB_JOB env may contain only "Init"
		if strings.HasSuffix(j.Name, targetJobName) {
			log.Entry().Debugf("current job id: %d", j.ID)
			g.currentJob = j
			return
		}
	}
}
