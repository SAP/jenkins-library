package orchestrator

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"

	"github.com/SAP/jenkins-library/pkg/log"
)

type GitHubActionsConfigProvider struct {
	client piperHttp.Client
	Token  string
}

func (g *GitHubActionsConfigProvider) InitOrchestratorProvider(settings *OrchestratorSettings) {
	g.Token = settings.GitHubActionToken
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

// func initGHProvider(settings *OrchestratorSettings) (*GitHubActionsConfigProvider, error) {
// 	g := GitHubActionsConfigProvider{}
// 	g.client = piperHttp.Client{}
// 	g.client.SetOptions(piperHttp.ClientOptions{
// 		Password:         settings.GitHubToken,
// 		MaxRetries:       3,
// 		TransportTimeout: time.Second * 10,
// 	})
// 	return &g, nil
// }

func (g *GitHubActionsConfigProvider) OrchestratorVersion() string {
	log.Entry().Debugf("OrchestratorVersion() for GitHub Actions is not applicable.")
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
	// token needs to be available in g
	// get IDs of all jobs
	// get logs of all job IDs
	// merge them

	ids, err := g.GetStageIds()
	if err != nil {
		return nil, err
	}

	logs := struct {
		b [][]byte
		// sync.Mutex
	}{
		b: make([][]byte, len(ids)),
	}
	for i, id := range ids {
		resp, err := g.client.GetRequest(
			fmt.Sprintf(
				"%s/jobs/%d/logs", getActionsURL(), id,
			),
			map[string][]string{
				"Accept":        {"application/vnd.github+json"},
				"Authorization": {fmt.Sprintf("Bearer %s", g.Token)},
			}, nil)
		if err != nil {
			return nil, fmt.Errorf("can't get API data: %w", err)

		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("can't read response body: %w", err)
		}
		defer resp.Body.Close()
		fmt.Println(string(body))
		logs.b[i] = append([]byte{}, body...)
		fmt.Println(string(logs.b[i]))
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

func (g *GitHubActionsConfigProvider) GetStageIds() ([]int64, error) {
	resp, err := g.client.GetRequest(fmt.Sprintf("%s/runs/%s/jobs", getActionsURL(), getEnv("GITHUB_RUN_ID", "")),
		g.getHeader(),
		nil)

	if err != nil {
		return nil, fmt.Errorf("can't get API data: %w", err)
	}

	var ids struct {
		Jobs []struct {
			Id int64 `json:"id"`
		} `json:"jobs"`
	}

	err = piperHttp.ParseHTTPResponseBodyJSON(resp, &ids)
	if err != nil {
		return nil, fmt.Errorf("can't parse JSON data: %w", err)
	}

	result := make([]int64, 0)

	for _, job := range ids.Jobs {
		result = append(result, job.Id)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("can't get the log form the last(current) job")
	}

	// last stage hasn't finished yet
	return result[:len(ids.Jobs)-1], nil
}

func (g *GitHubActionsConfigProvider) getHeader() http.Header {
	header := http.Header{
		"Accept":        {"application/vnd.github+json"},
		"Authorization": {fmt.Sprintf("Bearer %s", g.Token)},
	}
	return header
}
