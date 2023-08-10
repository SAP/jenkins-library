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
}

type run struct {
	Status    string    `json:"status"`
	StartedAt time.Time `json:"run_started_at"`
}

type job struct {
	ID int `json:"id"`
}

type stagesID struct {
	Jobs []job `json:"jobs"`
}

type logs struct {
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
	ids, err := g.getStageIds()
	if err != nil {
		return nil, err
	}

	logs := logs{
		b: make([][]byte, len(ids)),
	}

	ctx := context.Background()
	sem := semaphore.NewWeighted(10)
	wg := errgroup.Group{}
	for i := range ids {
		i := i // https://golang.org/doc/faq#closures_and_goroutines
		if err := sem.Acquire(ctx, 1); err != nil {
			return nil, fmt.Errorf("failed to acquire semaphore: %w", err)
		}
		wg.Go(func() error {
			defer sem.Release(1)
			resp, err := g.client.GetRequest(fmt.Sprintf("%s/jobs/%d/logs", g.actionsURL, ids[i]), httpHeaders, nil)
			if err != nil {
				return fmt.Errorf("failed to get API data: %w", err)
			}

			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body: %w", err)
			}
			defer resp.Body.Close()
			logs.Lock()
			defer logs.Unlock()
			logs.b[i] = b

			return nil
		})
	}
	if err = wg.Wait(); err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	return bytes.Join(logs.b, []byte("")), nil
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

// GetStageName returns the human-readable name given to a stage. e.g. "Promote" or "Init"
// TODO
func (g *GitHubActionsConfigProvider) GetStageName() string {
	return "GITHUB_WORKFLOW" // TODO: is there something like is "stage" in GH Actions?
}

// GetBuildReason returns the build reason
// TODO
func (g *GitHubActionsConfigProvider) GetBuildReason() string {
	log.Entry().Infof("GetBuildReason() for GitHub Actions not yet implemented.")
	return "n/a"
}

// GetBranch returns the source branch name, e.g. main
func (g *GitHubActionsConfigProvider) GetBranch() string {
	// TODO trim different prefixes. See GITHUB_REF description in
	// https://docs.github.com/en/actions/learn-github-actions/variables#default-environment-variables
	// or just use GITHUB_REF_NAME ???
	return strings.TrimPrefix(getEnv("GITHUB_REF", "n/a"), "refs/heads/")
}

// GetReference return the git reference. For example, refs/heads/your_branch_name
func (g *GitHubActionsConfigProvider) GetReference() string {
	return getEnv("GITHUB_REF", "n/a")
}

// GetBuildURL returns the builds URL. For example, https://github.com/SAP/jenkins-library/actions/runs/5815297487
func (g *GitHubActionsConfigProvider) GetBuildURL() string {
	return g.GetRepoURL() + "/actions/runs/" + getEnv("GITHUB_RUN_ID", "n/a")
}

// GetJobURL returns tje current job URL. For example, TODO
func (g *GitHubActionsConfigProvider) GetJobURL() string {
	log.Entry().Debugf("Not yet implemented.")
	return g.GetRepoURL() + "/actions/runs/" + getEnv("GITHUB_RUN_ID", "n/a")
}

// GetJobName TODO
func (g *GitHubActionsConfigProvider) GetJobName() string {
	log.Entry().Debugf("GetJobName() for GitHubActions not yet implemented.")
	return "n/a"
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

func (g *GitHubActionsConfigProvider) getStageIds() ([]int, error) {
	resp, err := g.client.GetRequest(fmt.Sprintf("%s/runs/%s/jobs", g.actionsURL, getEnv("GITHUB_RUN_ID", "")), httpHeaders, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get API data: %w", err)
	}

	var stagesID stagesID
	err = piperHttp.ParseHTTPResponseBodyJSON(resp, &stagesID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON data: %w", err)
	}

	ids := make([]int, len(stagesID.Jobs))
	for i, job := range stagesID.Jobs {
		ids[i] = job.ID
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("failed to get IDs")
	}

	// execution of the last stage hasn't finished yet - we can't get logs of the last stage
	return ids[:len(stagesID.Jobs)-1], nil
}
