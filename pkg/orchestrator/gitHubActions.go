package orchestrator

import (
	"bytes"
	"context"
	"fmt"
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"io"
	"sync"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
)

type GitHubActionsConfigProvider struct {
	client piperHttp.Client
	run    run
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

func initGHProvider(settings *OrchestratorSettings) (*GitHubActionsConfigProvider, error) {
	g := GitHubActionsConfigProvider{}
	g.client = piperHttp.Client{}
	g.client.SetOptions(piperHttp.ClientOptions{
		Password:         settings.GitHubToken,
		MaxRetries:       3,
		TransportTimeout: time.Second * 10,
	})
	return &g, nil
}

func (g *GitHubActionsConfigProvider) getData() error {
	resp, err := g.client.GetRequest(
		fmt.Sprintf(
			"%s/runs/%s", getActionsURL(), getEnv("GITHUB_RUN_ID", ""),
		),
		map[string][]string{
			"Accept":        {"application/vnd.github+json"},
			"Authorization": {"Bearer $GITHUB_TOKEN"},
		}, nil)
	if err != nil || resp.StatusCode != 200 {
		return fmt.Errorf("can't get API data: %w", err)

	}
	err = piperHttp.ParseHTTPResponseBodyJSON(resp, &g.run)
	if err != nil {
		return fmt.Errorf("can't parse JSON data: %w", err)
	}

	return nil
}

func (g *GitHubActionsConfigProvider) OrchestratorVersion() string {
	log.Entry().Debugf("OrchestratorVersion() for GitHub Actions is not applicable.")
	return "n/a"
}

func (g *GitHubActionsConfigProvider) OrchestratorType() string {
	return "GitHub Actions"
}

func (g *GitHubActionsConfigProvider) GetBuildStatus() string {
	// By default, we will assume it's a success
	// On error it would be handled by the action itself
	return "SUCCESS"
}

func (g *GitHubActionsConfigProvider) GetLog() ([]byte, error) {
	resp, err := g.client.GetRequest(
		fmt.Sprintf(
			"%s/runs/%s/jobs", getActionsURL(), getEnv("GITHUB_RUN_ID", ""),
		),
		map[string][]string{
			"Accept":        {"application/vnd.github+json"},
			"Authorization": {"Bearer $GITHUB_TOKEN"},
		}, nil)
	if err != nil {
		return nil, fmt.Errorf("can't get API data: %w", err)

	}
	var ids struct {
		Jobs []struct {
			Id string `json:"id"`
		} `json:"jobs"`
	}
	err = piperHttp.ParseHTTPResponseBodyJSON(resp, &ids)
	if err != nil {
		return nil, fmt.Errorf("can't parse JSON data: %w", err)
	}
	ids = struct {
		Jobs []struct {
			Id string `json:"id"`
		} `json:"jobs"`
		// we cant get the log for the last(current) job
	}{Jobs: ids.Jobs[:len(ids.Jobs)-1]}
	if len(ids.Jobs) == 0 {
		return nil, fmt.Errorf("can't get the log form the last(current) job")
	}
	logs := struct {
		b [][]byte
		sync.Mutex
	}{
		b: make([][]byte, len(ids.Jobs)),
	}
	ctx := context.TODO()
	sem := semaphore.NewWeighted(10)
	wg := errgroup.Group{}
	for i := range ids.Jobs {
		if err := sem.Acquire(ctx, 1); err != nil {
			return nil, fmt.Errorf("failed to acquire semaphore: %w", err)
		}
		j := i
		wg.Go(func() error {
			defer sem.Release(1)
			resp, err := g.client.GetRequest(
				fmt.Sprintf(
					"%s/jobs/%s/logs", getActionsURL(), ids.Jobs[j].Id,
				),
				map[string][]string{
					"Accept":        {"application/vnd.github+json"},
					"Authorization": {"Bearer $GITHUB_TOKEN"},
				}, nil)
			if err != nil {
				return fmt.Errorf("can't get API data: %w", err)

			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("can't read response body: %w", err)
			}
			logs.Lock()
			defer logs.Unlock()
			logs.b[j] = append([]byte{}, body...)
			return nil
		})
	}
	if err = wg.Wait(); err != nil {
		return nil, fmt.Errorf("recieving log error: %w", err)
	}
	return bytes.Join(logs.b, []byte("")), nil
}

func (g *GitHubActionsConfigProvider) GetBuildID() string {
	return getEnv("GITHUB_RUN_ID", "n/a")
}

func (g *GitHubActionsConfigProvider) GetChangeSet() []ChangeSet {
	return []ChangeSet{
		{
			g.run.HeadCommit.Id,
			g.run.HeadCommit.Timestamp.String(),
			0,
		},
	}
}

func (g *GitHubActionsConfigProvider) GetPipelineStartTime() time.Time {
	if g.run == (run{}) {
		g.getData()
	}
	return g.run.RunStartedAt.UTC()
}

func (g *GitHubActionsConfigProvider) GetStageName() string {
	return getEnv("GITHUB_JOB", "n/a")
}

func (g *GitHubActionsConfigProvider) GetBuildReason() string {
	return getEnv("GITHUB_EVENT_NAME", "n/a")
}

func (g *GitHubActionsConfigProvider) GetBranch() string {
	return getEnv("GITHUB_REF_NAME", "n/a")
}

func (g *GitHubActionsConfigProvider) GetReference() string {
	return getEnv("GITHUB_REF", "n/a")
}

func (g *GitHubActionsConfigProvider) GetBuildURL() string {
	if g.run == (run{}) {
		g.getData()
	}
	return g.run.HtmlUrl
}

func (g *GitHubActionsConfigProvider) GetJobURL() string {
	log.Entry().Infof("GetJobURL() for GitHub Actions is not applicable.")
	return "n/a"
}

func (g *GitHubActionsConfigProvider) GetJobName() string {
	log.Entry().Debugf("GetJobName() for GitHub Actions is not applicable.")
	return "n/a"
}

func (g *GitHubActionsConfigProvider) GetCommit() string {
	return getEnv("GITHUB_SHA", "n/a")
}

func (g *GitHubActionsConfigProvider) GetRepoURL() string {
	if g.run == (run{}) {
		g.getData()
	}
	return g.run.Repository.HtmlUrl
}

func (g *GitHubActionsConfigProvider) GetPullRequestConfig() PullRequestConfig {
	return PullRequestConfig{
		Branch: getEnv("GITHUB_HEAD_REF", "n/a"),
		Base:   getEnv("GITHUB_BASE_REF", "n/a"),
		Key:    getEnv("GITHUB_EVENT_PULL_REQUEST_NUMBER", "n/a"),
	}
}

func (g *GitHubActionsConfigProvider) IsPullRequest() bool {
	return getEnv("GITHUB_HEAD_REF", "") != ""
}

func isGitHubActions() bool {
	return getEnv("GITHUB_ACTIONS", "") == "true"
}

// https://docs.github.com/en/rest/actions/workflow-runs?apiVersion=2022-11-28#get-a-workflow-run
type run struct {
	//Id               int           `json:"id"`
	//Name             string        `json:"name"`
	//NodeId           string        `json:"node_id"`
	//HeadBranch       string        `json:"head_branch"`
	//HeadSha          string        `json:"head_sha"`
	//RunNumber        int           `json:"run_number"`
	//Event            string        `json:"event"`
	//Status           string        `json:"status"`
	//Conclusion       interface{}   `json:"conclusion"`
	//WorkflowId       int           `json:"workflow_id"`
	//CheckSuiteId     int           `json:"check_suite_id"`
	//CheckSuiteNodeId string        `json:"check_suite_node_id"`
	//Url              string        `json:"url"`
	HtmlUrl string `json:"html_url"`
	//PullRequests     []interface{} `json:"pull_requests"`
	//CreatedAt        time.Time     `json:"created_at"`
	//UpdatedAt        time.Time     `json:"updated_at"`
	//Actor            struct {
	//	Login             string `json:"login"`
	//	Id                int    `json:"id"`
	//	NodeId            string `json:"node_id"`
	//	AvatarUrl         string `json:"avatar_url"`
	//	GravatarId        string `json:"gravatar_id"`
	//	Url               string `json:"url"`
	//	HtmlUrl           string `json:"html_url"`
	//	FollowersUrl      string `json:"followers_url"`
	//	FollowingUrl      string `json:"following_url"`
	//	GistsUrl          string `json:"gists_url"`
	//	StarredUrl        string `json:"starred_url"`
	//	SubscriptionsUrl  string `json:"subscriptions_url"`
	//	OrganizationsUrl  string `json:"organizations_url"`
	//	ReposUrl          string `json:"repos_url"`
	//	EventsUrl         string `json:"events_url"`
	//	ReceivedEventsUrl string `json:"received_events_url"`
	//	Type              string `json:"type"`
	//	SiteAdmin         bool   `json:"site_admin"`
	//} `json:"actor"`
	//RunAttempt      int       `json:"run_attempt"`
	RunStartedAt time.Time `json:"run_started_at"`
	//TriggeringActor struct {
	//	Login             string `json:"login"`
	//	Id                int    `json:"id"`
	//	NodeId            string `json:"node_id"`
	//	AvatarUrl         string `json:"avatar_url"`
	//	GravatarId        string `json:"gravatar_id"`
	//	Url               string `json:"url"`
	//	HtmlUrl           string `json:"html_url"`
	//	FollowersUrl      string `json:"followers_url"`
	//	FollowingUrl      string `json:"following_url"`
	//	GistsUrl          string `json:"gists_url"`
	//	StarredUrl        string `json:"starred_url"`
	//	SubscriptionsUrl  string `json:"subscriptions_url"`
	//	OrganizationsUrl  string `json:"organizations_url"`
	//	ReposUrl          string `json:"repos_url"`
	//	EventsUrl         string `json:"events_url"`
	//	ReceivedEventsUrl string `json:"received_events_url"`
	//	Type              string `json:"type"`
	//	SiteAdmin         bool   `json:"site_admin"`
	//} `json:"triggering_actor"`
	//JobsUrl            string      `json:"jobs_url"`
	//LogsUrl            string      `json:"logs_url"`
	//CheckSuiteUrl      string      `json:"check_suite_url"`
	//ArtifactsUrl       string      `json:"artifacts_url"`
	//CancelUrl          string      `json:"cancel_url"`
	//RerunUrl           string      `json:"rerun_url"`
	//PreviousAttemptUrl interface{} `json:"previous_attempt_url"`
	//WorkflowUrl        string      `json:"workflow_url"`
	HeadCommit struct {
		Id string `json:"id"`
		//	TreeId    string    `json:"tree_id"`
		//	Message   string    `json:"message"`
		Timestamp time.Time `json:"timestamp"`
		//	Author    struct {
		//		Name  string `json:"name"`
		//		Email string `json:"email"`
		//	} `json:"author"`
		//	Committer struct {
		//		Name  string `json:"name"`
		//		Email string `json:"email"`
		//	} `json:"committer"`
	} `json:"head_commit"`
	Repository struct {
		//	Id       int    `json:"id"`
		//	NodeId   string `json:"node_id"`
		//	Name     string `json:"name"`
		//	FullName string `json:"full_name"`
		//	Private  bool   `json:"private"`
		//	Owner    struct {
		//		Login             string `json:"login"`
		//		Id                int    `json:"id"`
		//		NodeId            string `json:"node_id"`
		//		AvatarUrl         string `json:"avatar_url"`
		//		GravatarId        string `json:"gravatar_id"`
		//		Url               string `json:"url"`
		//		HtmlUrl           string `json:"html_url"`
		//		FollowersUrl      string `json:"followers_url"`
		//		FollowingUrl      string `json:"following_url"`
		//		GistsUrl          string `json:"gists_url"`
		//		StarredUrl        string `json:"starred_url"`
		//		SubscriptionsUrl  string `json:"subscriptions_url"`
		//		OrganizationsUrl  string `json:"organizations_url"`
		//		ReposUrl          string `json:"repos_url"`
		//		EventsUrl         string `json:"events_url"`
		//		ReceivedEventsUrl string `json:"received_events_url"`
		//		Type              string `json:"type"`
		//		SiteAdmin         bool   `json:"site_admin"`
		//	} `json:"owner"`
		HtmlUrl string `json:"html_url"`
		//	Description      interface{} `json:"description"`
		//	Fork             bool        `json:"fork"`
		//	Url              string      `json:"url"`
		//	ForksUrl         string      `json:"forks_url"`
		//	KeysUrl          string      `json:"keys_url"`
		//	CollaboratorsUrl string      `json:"collaborators_url"`
		//	TeamsUrl         string      `json:"teams_url"`
		//	HooksUrl         string      `json:"hooks_url"`
		//	IssueEventsUrl   string      `json:"issue_events_url"`
		//	EventsUrl        string      `json:"events_url"`
		//	AssigneesUrl     string      `json:"assignees_url"`
		//	BranchesUrl      string      `json:"branches_url"`
		//	TagsUrl          string      `json:"tags_url"`
		//	BlobsUrl         string      `json:"blobs_url"`
		//	GitTagsUrl       string      `json:"git_tags_url"`
		//	GitRefsUrl       string      `json:"git_refs_url"`
		//	TreesUrl         string      `json:"trees_url"`
		//	StatusesUrl      string      `json:"statuses_url"`
		//	LanguagesUrl     string      `json:"languages_url"`
		//	StargazersUrl    string      `json:"stargazers_url"`
		//	ContributorsUrl  string      `json:"contributors_url"`
		//	SubscribersUrl   string      `json:"subscribers_url"`
		//	SubscriptionUrl  string      `json:"subscription_url"`
		//	CommitsUrl       string      `json:"commits_url"`
		//	GitCommitsUrl    string      `json:"git_commits_url"`
		//	CommentsUrl      string      `json:"comments_url"`
		//	IssueCommentUrl  string      `json:"issue_comment_url"`
		//	ContentsUrl      string      `json:"contents_url"`
		//	CompareUrl       string      `json:"compare_url"`
		//	MergesUrl        string      `json:"merges_url"`
		//	ArchiveUrl       string      `json:"archive_url"`
		//	DownloadsUrl     string      `json:"downloads_url"`
		//	IssuesUrl        string      `json:"issues_url"`
		//	PullsUrl         string      `json:"pulls_url"`
		//	MilestonesUrl    string      `json:"milestones_url"`
		//	NotificationsUrl string      `json:"notifications_url"`
		//	LabelsUrl        string      `json:"labels_url"`
		//	ReleasesUrl      string      `json:"releases_url"`
		//	DeploymentsUrl   string      `json:"deployments_url"`
	} `json:"repository"`
	//HeadRepository struct {
	//	Id       int    `json:"id"`
	//	NodeId   string `json:"node_id"`
	//	Name     string `json:"name"`
	//	FullName string `json:"full_name"`
	//	Private  bool   `json:"private"`
	//	Owner    struct {
	//		Login             string `json:"login"`
	//		Id                int    `json:"id"`
	//		NodeId            string `json:"node_id"`
	//		AvatarUrl         string `json:"avatar_url"`
	//		GravatarId        string `json:"gravatar_id"`
	//		Url               string `json:"url"`
	//		HtmlUrl           string `json:"html_url"`
	//		FollowersUrl      string `json:"followers_url"`
	//		FollowingUrl      string `json:"following_url"`
	//		GistsUrl          string `json:"gists_url"`
	//		StarredUrl        string `json:"starred_url"`
	//		SubscriptionsUrl  string `json:"subscriptions_url"`
	//		OrganizationsUrl  string `json:"organizations_url"`
	//		ReposUrl          string `json:"repos_url"`
	//		EventsUrl         string `json:"events_url"`
	//		ReceivedEventsUrl string `json:"received_events_url"`
	//		Type              string `json:"type"`
	//		SiteAdmin         bool   `json:"site_admin"`
	//	} `json:"owner"`
	//	HtmlUrl          string      `json:"html_url"`
	//	Description      interface{} `json:"description"`
	//	Fork             bool        `json:"fork"`
	//	Url              string      `json:"url"`
	//	ForksUrl         string      `json:"forks_url"`
	//	KeysUrl          string      `json:"keys_url"`
	//	CollaboratorsUrl string      `json:"collaborators_url"`
	//	TeamsUrl         string      `json:"teams_url"`
	//	HooksUrl         string      `json:"hooks_url"`
	//	IssueEventsUrl   string      `json:"issue_events_url"`
	//	EventsUrl        string      `json:"events_url"`
	//	AssigneesUrl     string      `json:"assignees_url"`
	//	BranchesUrl      string      `json:"branches_url"`
	//	TagsUrl          string      `json:"tags_url"`
	//	BlobsUrl         string      `json:"blobs_url"`
	//	GitTagsUrl       string      `json:"git_tags_url"`
	//	GitRefsUrl       string      `json:"git_refs_url"`
	//	TreesUrl         string      `json:"trees_url"`
	//	StatusesUrl      string      `json:"statuses_url"`
	//	LanguagesUrl     string      `json:"languages_url"`
	//	StargazersUrl    string      `json:"stargazers_url"`
	//	ContributorsUrl  string      `json:"contributors_url"`
	//	SubscribersUrl   string      `json:"subscribers_url"`
	//	SubscriptionUrl  string      `json:"subscription_url"`
	//	CommitsUrl       string      `json:"commits_url"`
	//	GitCommitsUrl    string      `json:"git_commits_url"`
	//	CommentsUrl      string      `json:"comments_url"`
	//	IssueCommentUrl  string      `json:"issue_comment_url"`
	//	ContentsUrl      string      `json:"contents_url"`
	//	CompareUrl       string      `json:"compare_url"`
	//	MergesUrl        string      `json:"merges_url"`
	//	ArchiveUrl       string      `json:"archive_url"`
	//	DownloadsUrl     string      `json:"downloads_url"`
	//	IssuesUrl        string      `json:"issues_url"`
	//	PullsUrl         string      `json:"pulls_url"`
	//	MilestonesUrl    string      `json:"milestones_url"`
	//	NotificationsUrl string      `json:"notifications_url"`
	//	LabelsUrl        string      `json:"labels_url"`
	//	ReleasesUrl      string      `json:"releases_url"`
	//	DeploymentsUrl   string      `json:"deployments_url"`
	//} `json:"head_repository"`
}
