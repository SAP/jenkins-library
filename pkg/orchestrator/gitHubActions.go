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
	client piperHttp.Client
}

type Job struct {
	ID int `json:"id"`
}

type StagesID struct {
	Jobs []Job `json:"jobs"`
}

type Logs struct {
	sync.Mutex
	b [][]byte
}

func (g *GitHubActionsConfigProvider) InitOrchestratorProvider(settings *OrchestratorSettings) {
	log.Entry().Debug("Successfully initialized GitHubActions config provider")
}

func getActionsURL() string {
	ghURL := getEnv("GITHUB_URL", "")
	switch ghURL {
	case "https://github.com/":
		ghURL = "https://api.github.com"
	default:
		ghURL += "api/v3"
	}
	return fmt.Sprintf("%s/repos/%s/actions", ghURL, getEnv("GITHUB_REPOSITORY", ""))
}

func gitHubActionsConfigProvider(settings *OrchestratorSettings) (*GitHubActionsConfigProvider, error) {
	g := GitHubActionsConfigProvider{}
	g.client = piperHttp.Client{}
	g.client.SetOptions(piperHttp.ClientOptions{
		Password:         settings.GitHubToken,
		MaxRetries:       3,
		TransportTimeout: time.Second * 10,
	})
	return &g, nil
}

func (g *GitHubActionsConfigProvider) OrchestratorVersion() string {
	return "n/a"
}

func (g *GitHubActionsConfigProvider) OrchestratorType() string {
	return "GitHubActions"
}

func (g *GitHubActionsConfigProvider) GetBuildStatus() string {
	log.Entry().Infof("GetBuildStatus() for GitHub Actions not yet implemented.")
	return "FAILURE"
}

func (g *GitHubActionsConfigProvider) GetLog() ([]byte, error) {
	ids, err := g.GetStageIds()
	if err != nil {
		return nil, err
	}

	logs := Logs{
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
			resp, err := g.client.GetRequest(fmt.Sprintf("%s/jobs/%d/logs", getActionsURL(), ids[i]), g.getHeader(), nil)
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
			logs.b[i] = append([]byte{}, b...)

			return nil
		})
	}
	if err = wg.Wait(); err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	return bytes.Join(logs.b, []byte("")), nil
}

func (g *GitHubActionsConfigProvider) GetBuildID() string {
	log.Entry().Infof("GetBuildID() for GitHub Actions not yet implemented.")
	return "n/a"
}

func (g *GitHubActionsConfigProvider) GetChangeSet() []ChangeSet {
	log.Entry().Warn("GetChangeSet for GitHubActions not yet implemented")
	return []ChangeSet{}
}

func (g *GitHubActionsConfigProvider) GetPipelineStartTime() time.Time {
	log.Entry().Infof("GetPipelineStartTime() for GitHub Actions not yet implemented.")
	return time.Time{}.UTC()
}
func (g *GitHubActionsConfigProvider) GetStageName() string {
	return "GITHUB_WORKFLOW" // TODO: is there something like is "stage" in GH Actions?
}

func (g *GitHubActionsConfigProvider) GetBuildReason() string {
	log.Entry().Infof("GetBuildReason() for GitHub Actions not yet implemented.")
	return "n/a"
}

func (g *GitHubActionsConfigProvider) GetBranch() string {
	return strings.TrimPrefix(getEnv("GITHUB_REF", "n/a"), "refs/heads/")
}

func (g *GitHubActionsConfigProvider) GetReference() string {
	return getEnv("GITHUB_REF", "n/a")
}

func (g *GitHubActionsConfigProvider) GetBuildURL() string {
	return g.GetRepoURL() + "/actions/runs/" + getEnv("GITHUB_RUN_ID", "n/a")
}

func (g *GitHubActionsConfigProvider) GetJobURL() string {
	log.Entry().Debugf("Not yet implemented.")
	return g.GetRepoURL() + "/actions/runs/" + getEnv("GITHUB_RUN_ID", "n/a")
}

func (g *GitHubActionsConfigProvider) GetJobName() string {
	log.Entry().Debugf("GetJobName() for GitHubActions not yet implemented.")
	return "n/a"
}

func (g *GitHubActionsConfigProvider) GetCommit() string {
	return getEnv("GITHUB_SHA", "n/a")
}

func (g *GitHubActionsConfigProvider) GetRepoURL() string {
	return getEnv("GITHUB_SERVER_URL", "n/a") + "/" + getEnv("GITHUB_REPOSITORY", "n/a")
}

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

func (g *GitHubActionsConfigProvider) IsPullRequest() bool {
	return truthy("GITHUB_HEAD_REF")
}

func isGitHubActions() bool {
	envVars := []string{"GITHUB_ACTION", "GITHUB_ACTIONS"}
	return areIndicatingEnvVarsSet(envVars)
}

func (g *GitHubActionsConfigProvider) GetStageIds() ([]int, error) {
	resp, err := g.client.GetRequest(fmt.Sprintf("%s/runs/%s/jobs", getActionsURL(), getEnv("GITHUB_RUN_ID", "")), g.getHeader(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get API data: %w", err)
	}

	var stagesID StagesID
	err = piperHttp.ParseHTTPResponseBodyJSON(resp, &stagesID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON data: %w", err)
	}

	ids := make([]int, len(stagesID.Jobs))
	for i, job := range stagesID.Jobs {
		ids[i] = job.ID
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("failed to get logs")
	}

	// execution of the last stage hasn't finished yet - we can't get logs of the last stage
	return ids[:len(stagesID.Jobs)-1], nil
}

func (g *GitHubActionsConfigProvider) getHeader() http.Header {
	header := http.Header{
		"Accept":        {"application/vnd.github+json"},
		"Authorization": {fmt.Sprintf("Bearer %s", getEnv("GITHUB_TOKEN", ""))},
	}
	return header
}
