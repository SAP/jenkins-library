//go:build unit

package orchestrator

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/go-github/v68/github"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestGitHubActionsConfigProvider_GetBuildStatus(t *testing.T) {
	tests := []struct {
		name string
		jobs []job
		want string
	}{
		{"BuildStatusSuccess", []job{{Conclusion: "success"}, {Conclusion: "success"}, {Conclusion: "success"}}, BuildStatusSuccess},
		{"BuildStatusAborted", []job{{Conclusion: "success"}, {Conclusion: "success"}, {Conclusion: "cancelled"}}, BuildStatusAborted},
		{"BuildStatusFailure", []job{{Conclusion: "success"}, {Conclusion: "failure"}, {Conclusion: "cancelled"}}, BuildStatusFailure},
		{"BuildStatusSuccess", []job{{Conclusion: "success"}, {Conclusion: "cancelled"}, {Conclusion: "failure"}}, BuildStatusAborted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &githubActionsConfigProvider{
				jobsFetched: true,
				jobs:        tt.jobs,
			}
			assert.Equalf(t, tt.want, g.BuildStatus(), "BuildStatus()")
		})
	}
}

func TestGitHubActionsConfigProvider_GetBuildReason(t *testing.T) {
	tests := []struct {
		name         string
		envGithubRef string
		want         string
	}{
		{"BuildReasonManual", "workflow_dispatch", BuildReasonManual},
		{"BuildReasonSchedule", "schedule", BuildReasonSchedule},
		{"BuildReasonPullRequest", "pull_request", BuildReasonPullRequest},
		{"BuildReasonResourceTrigger", "workflow_call", BuildReasonResourceTrigger},
		{"BuildReasonIndividualCI", "push", BuildReasonIndividualCI},
		{"BuildReasonUnknown", "qwerty", BuildReasonUnknown},
		{"BuildReasonUnknown", "", BuildReasonUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &githubActionsConfigProvider{}

			_ = os.Setenv("GITHUB_EVENT_NAME", tt.envGithubRef)
			assert.Equalf(t, tt.want, g.BuildReason(), "BuildReason()")
		})
	}
}

func TestGitHubActionsConfigProvider_GetRepoURL(t *testing.T) {
	tests := []struct {
		name         string
		envServerURL string
		envRepo      string
		want         string
	}{
		{"github.com", "https://github.com", "SAP/jenkins-library", "https://github.com/SAP/jenkins-library"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &githubActionsConfigProvider{}

			_ = os.Setenv("GITHUB_SERVER_URL", tt.envServerURL)
			_ = os.Setenv("GITHUB_REPOSITORY", tt.envRepo)
			assert.Equalf(t, tt.want, g.RepoURL(), "RepoURL()")
		})
	}
}

func TestGitHubActionsConfigProvider_GetPullRequestConfig(t *testing.T) {
	tests := []struct {
		name   string
		envRef string
		want   PullRequestConfig
	}{
		{"1", "refs/pull/1234/merge", PullRequestConfig{"n/a", "n/a", "1234"}},
		{"2", "refs/pull/1234", PullRequestConfig{"n/a", "n/a", "1234"}},
		{"2", "1234/merge", PullRequestConfig{"n/a", "n/a", "1234"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &githubActionsConfigProvider{}

			_ = os.Setenv("GITHUB_REF", tt.envRef)
			_ = os.Setenv("GITHUB_HEAD_REF", "n/a")
			_ = os.Setenv("GITHUB_BASE_REF", "n/a")
			assert.Equalf(t, tt.want, g.PullRequestConfig(), "PullRequestConfig()")
		})
	}
}

func TestGitHubActionsConfigProvider_fetchRunData(t *testing.T) {
	// data
	respJson := map[string]interface{}{
		"status":         "completed",
		"run_started_at": "2023-08-11T07:28:24Z",
		"html_url":       "https://github.com/SAP/jenkins-library/actions/runs/11111",
	}
	startedAt, _ := time.Parse(time.RFC3339, "2023-08-11T07:28:24Z")
	wantRunData := run{
		fetched:   true,
		StartedAt: startedAt,
	}

	// setup env vars
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("GITHUB_API_URL", "https://api.github.com")
	_ = os.Setenv("GITHUB_REPOSITORY", "SAP/jenkins-library")
	_ = os.Setenv("GITHUB_RUN_ID", "11111")

	// setup provider
	g := newGithubActionsConfigProvider()
	assert.NoError(t, g.Configure(&Options{}))
	g.client = github.NewClient(http.DefaultClient)

	// setup http mock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(http.MethodGet, "https://api.github.com/repos/SAP/jenkins-library/actions/runs/11111",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, respJson)
		},
	)

	// run
	g.fetchRunData()
	assert.Equal(t, wantRunData, g.runData)
}

func TestGitHubActionsConfigProvider_fetchJobs(t *testing.T) {
	// data
	respJson := map[string]interface{}{"jobs": []map[string]interface{}{{
		"id":        111,
		"name":      "Piper / Init",
		"html_url":  "https://github.com/SAP/jenkins-library/actions/runs/11111/jobs/111",
		"runner_id": 12345,
	}, {
		"id":        222,
		"name":      "Piper / Build",
		"html_url":  "https://github.com/SAP/jenkins-library/actions/runs/11111/jobs/222",
		"runner_id": 12345,
	}, {
		"id":        333,
		"name":      "Piper / Acceptance",
		"html_url":  "https://github.com/SAP/jenkins-library/actions/runs/11111/jobs/333",
		"runner_id": 12345,
	},
	}}
	wantJobs := []job{{
		ID:      111,
		Name:    "Piper / Init",
		HtmlURL: "https://github.com/SAP/jenkins-library/actions/runs/11111/jobs/111",
	}, {
		ID:      222,
		Name:    "Piper / Build",
		HtmlURL: "https://github.com/SAP/jenkins-library/actions/runs/11111/jobs/222",
	}, {
		ID:      333,
		Name:    "Piper / Acceptance",
		HtmlURL: "https://github.com/SAP/jenkins-library/actions/runs/11111/jobs/333",
	}}

	// setup env vars
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("GITHUB_API_URL", "https://api.github.com")
	_ = os.Setenv("GITHUB_REPOSITORY", "SAP/jenkins-library")
	_ = os.Setenv("GITHUB_RUN_ID", "11111")

	// setup provider
	g := newGithubActionsConfigProvider()
	assert.NoError(t, g.Configure(&Options{}))
	g.client = github.NewClient(http.DefaultClient)

	// setup http mock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		http.MethodGet,
		"https://api.github.com/repos/SAP/jenkins-library/actions/runs/11111/jobs",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, respJson)
		},
	)

	// run
	err := g.fetchJobs()
	assert.NoError(t, err)
	assert.Equal(t, wantJobs, g.jobs)
}

func TestGitHubActionsConfigProvider_GetLog(t *testing.T) {
	// data
	respLogs := []string{
		"log_record11\nlog_record12\nlog_record13\n",
		"log_record21\nlog_record22\n",
		"log_record31\nlog_record32\n",
		"log_record41\n",
	}
	wantLogs := "log_record11\nlog_record12\nlog_record13\nlog_record21\n" +
		"log_record22\nlog_record31\nlog_record32\nlog_record41\n"
	jobs := []job{
		{ID: 111}, {ID: 222}, {ID: 333}, {ID: 444}, {ID: 555},
	}

	// setup env vars
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("GITHUB_API_URL", "https://api.github.com")
	_ = os.Setenv("GITHUB_REPOSITORY", "SAP/jenkins-library")

	// setup provider
	g := newGithubActionsConfigProvider()
	g.jobs = jobs
	g.jobsFetched = true

	assert.NoError(t, g.Configure(&Options{}))
	g.client = github.NewClient(http.DefaultClient)

	// setup http mock
	latencyMin, latencyMax := 15, 500 // milliseconds
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	for i, j := range jobs {
		idx := i
		httpmock.RegisterResponder(
			http.MethodGet,
			fmt.Sprintf("https://api.github.com/repos/SAP/jenkins-library/actions/jobs/%d/logs", j.ID),
			func(jobId int64) func(req *http.Request) (*http.Response, error) {
				return func(req *http.Request) (*http.Response, error) {
					resp := httpmock.NewStringResponse(http.StatusFound, respLogs[idx])
					logsDownloadUrl := fmt.Sprintf("https://api.github.com/repos/SAP/jenkins-library/actions/jobs/%d/logs/download", jobId)
					resp.Header.Set("Location", logsDownloadUrl)
					return resp, nil
				}
			}(j.ID),
		)
		httpmock.RegisterResponder(
			http.MethodGet,
			fmt.Sprintf("https://api.github.com/repos/SAP/jenkins-library/actions/jobs/%d/logs/download", j.ID),
			func(req *http.Request) (*http.Response, error) {
				// simulate response delay
				latency := rand.Intn(latencyMax-latencyMin) + latencyMin
				time.Sleep(time.Duration(latency) * time.Millisecond)
				return httpmock.NewStringResponse(200, respLogs[idx]), nil
			},
		)
	}
	// run
	logs, err := g.FullLogs()
	assert.NoError(t, err)
	assert.Equal(t, wantLogs, string(logs))
}

func TestGitHubActionsConfigProvider_Others(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	_ = os.Setenv("GITHUB_ACTION", "1")
	_ = os.Setenv("GITHUB_JOB", "Build")
	_ = os.Setenv("GITHUB_RUN_ID", "11111")
	_ = os.Setenv("GITHUB_REF_NAME", "main")
	_ = os.Setenv("GITHUB_HEAD_REF", "feature-branch-1")
	_ = os.Setenv("GITHUB_REF", "refs/pull/42/merge")
	_ = os.Setenv("GITHUB_WORKFLOW", "Piper workflow")
	_ = os.Setenv("GITHUB_SHA", "ffac537e6cbbf934b08745a378932722df287a53")
	_ = os.Setenv("GITHUB_API_URL", "https://api.github.com")
	_ = os.Setenv("GITHUB_SERVER_URL", "https://github.com")
	_ = os.Setenv("GITHUB_REPOSITORY", "SAP/jenkins-library")
	_ = os.Setenv("GITHUB_WORKFLOW_REF", "SAP/jenkins-library/.github/workflows/piper.yml@refs/heads/main")

	p := githubActionsConfigProvider{}
	startedAt, _ := time.Parse(time.RFC3339, "2023-08-11T07:28:24Z")
	p.runData = run{
		fetched:   true,
		StartedAt: startedAt,
	}

	assert.Equal(t, "n/a", p.OrchestratorVersion())
	assert.Equal(t, "GitHubActions", p.OrchestratorType())
	assert.Equal(t, "11111", p.BuildID())
	assert.Equal(t, []ChangeSet{}, p.ChangeSets())
	assert.Equal(t, startedAt, p.PipelineStartTime())
	assert.Equal(t, "Build", p.StageName())
	assert.Equal(t, "main", p.Branch())
	assert.Equal(t, "refs/pull/42/merge", p.GitReference())
	assert.Equal(t, "https://github.com/SAP/jenkins-library/actions/runs/11111", p.BuildURL())
	assert.Equal(t, "https://github.com/SAP/jenkins-library/actions/workflows/piper.yml", p.JobURL())
	assert.Equal(t, "Piper workflow", p.JobName())
	assert.Equal(t, "ffac537e6cbbf934b08745a378932722df287a53", p.CommitSHA())
	assert.Equal(t, "https://api.github.com/repos/SAP/jenkins-library/actions", actionsURL())
	assert.True(t, p.IsPullRequest())
	assert.True(t, isGitHubActions())
}

func TestWorkflowFileName(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()

	tests := []struct {
		name, workflowRef, want string
	}{
		{
			name:        "valid file name (yaml)",
			workflowRef: "owner/repo/.github/workflows/test-workflow.yaml@refs/heads/main",
			want:        "test-workflow.yaml",
		},
		{
			name:        "valid file name (yml)",
			workflowRef: "owner/repo/.github/workflows/test-workflow.yml@refs/heads/main",
			want:        "test-workflow.yml",
		},
		{
			name:        "invalid file name",
			workflowRef: "owner/repo/.github/workflows/test-workflow@refs/heads/main",
			want:        "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv("GITHUB_WORKFLOW_REF", tt.workflowRef)
			result := workflowFileName()
			assert.Equal(t, tt.want, result)
		})
	}
}
func Test_filterJobs(t *testing.T) {
	tests := []struct {
		name string
		jobs []*github.WorkflowJob
		want []*github.WorkflowJob
	}{
		{
			name: "all jobs have runner id",
			jobs: []*github.WorkflowJob{
				{RunnerID: github.Ptr(int64(1))},
				{RunnerID: github.Ptr(int64(2))},
			},
			want: []*github.WorkflowJob{
				{RunnerID: github.Ptr(int64(1))},
				{RunnerID: github.Ptr(int64(2))},
			},
		},
		{
			name: "no jobs have runner id",
			jobs: []*github.WorkflowJob{
				{RunnerID: nil},
				{RunnerID: github.Ptr(int64(0))},
			},
			want: []*github.WorkflowJob{},
		},
		{
			name: "some jobs have runner id",
			jobs: []*github.WorkflowJob{
				{RunnerID: github.Ptr(int64(1))},
				{RunnerID: github.Ptr(int64(0))},
				{RunnerID: github.Ptr(int64(3))},
			},
			want: []*github.WorkflowJob{
				{RunnerID: github.Ptr(int64(1))},
				{RunnerID: github.Ptr(int64(3))},
			},
		},
		{
			name: "empty input",
			jobs: []*github.WorkflowJob{},
			want: []*github.WorkflowJob{},
		},
		{
			name: "nil input",
			jobs: nil,
			want: []*github.WorkflowJob{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterJobs(tt.jobs)
			assert.Equal(t, len(tt.want), len(got))
			for i := range tt.want {
				assert.Equal(t, tt.want[i].GetRunnerID(), got[i].GetRunnerID())
			}
		})
	}
}
