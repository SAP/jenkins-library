package orchestrator

import (
	"fmt"
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
)

type GitHubActionsConfigProvider struct {
	run run
}

func (g *GitHubActionsConfigProvider) initOrchestratorProvider(settings *OrchestratorSettings) error {
	client := piperHttp.Client{}
	client.SetOptions(piperHttp.ClientOptions{
		Password:         settings.GitHubToken,
		MaxRetries:       3,
		TransportTimeout: time.Second * 10,
	})
	ghURL := getEnv("GITHUB_URL", "")
	switch ghURL {
	case "https://github.com/":
		ghURL = "https://api.github.com/"
	default:
		ghURL += "api/v3/"
	}
	resp, err := client.GetRequest(
		fmt.Sprintf(
			"%s/repos/%s/actions/runs/%s", ghURL, getEnv("GITHUB_REPOSITORY", ""), getEnv("GITHUB_RUN_ID", ""),
		),
		map[string][]string{
			"Accept":        {"application/vnd.github+json"},
			"Authorization": {"Bearer $GITHUB_TOKEN"},
		}, nil)
	if err != nil {
		return fmt.Errorf("can't get API data: %w", err)

	}
	err = piperHttp.ParseHTTPResponseBodyJSON(resp, g.run)
	if err != nil {
		return fmt.Errorf("can't parse JSON data: %w", err)
	}
	return err
}

func (g *GitHubActionsConfigProvider) OrchestratorVersion() string {
	log.Entry().Debugf("OrchestratorVersion() for GitHub Actions is not applicable.")
	return "n/a"
}

func (g *GitHubActionsConfigProvider) OrchestratorType() string {
	return "GitHub Actions"
}

func (g *GitHubActionsConfigProvider) GetBuildStatus() string {
	// By default, we will assume it ass a success
	// On error it would be handled by the action itself
	return "SUCCESS"
}

func (g *GitHubActionsConfigProvider) GetLog() ([]byte, error) {
	// It's not possible to get log during workflow execution
	log.Entry().Debugf("GetLog() for GitHub Actions is not applicable.")
	return []byte{}, nil
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
