package orchestrator

import (
	"github.com/SAP/jenkins-library/pkg/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAzure(t *testing.T) {
	t.Run("Azure - BranchBuild", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Setenv("AZURE_HTTP_USER_AGENT", "FOO BAR BAZ")
		os.Setenv("BUILD_SOURCEBRANCH", "refs/heads/feat/test-azure")
		os.Setenv("SYSTEM_TEAMFOUNDATIONCOLLECTIONURI", "https://pogo.foo")
		os.Setenv("SYSTEM_TEAMPROJECT", "bar")
		os.Setenv("BUILD_BUILDID", "42")
		os.Setenv("BUILD_SOURCEVERSION", "abcdef42713")
		os.Setenv("BUILD_REPOSITORY_URI", "github.com/foo/bar")

		p, _ := NewOrchestratorSpecificConfigProvider()

		assert.False(t, p.IsPullRequest())
		assert.Equal(t, "feat/test-azure", p.GetBranch())
		assert.Equal(t, "https://pogo.foobar/_build/results?buildId=42", p.GetBuildUrl())
		assert.Equal(t, "abcdef42713", p.GetCommit())
		assert.Equal(t, "github.com/foo/bar", p.GetRepoUrl())
		assert.Equal(t, "Azure", p.OrchestratorType())
	})

	t.Run("PR", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Setenv("SYSTEM_PULLREQUEST_SOURCEBRANCH", "feat/test-azure")
		os.Setenv("SYSTEM_PULLREQUEST_TARGETBRANCH", "main")
		os.Setenv("SYSTEM_PULLREQUEST_PULLREQUESTID", "42")
		os.Setenv("BUILD_REASON", "PullRequest")

		p := AzureDevOpsConfigProvider{}
		c := p.GetPullRequestConfig()

		assert.True(t, p.IsPullRequest())
		assert.Equal(t, "feat/test-azure", c.Branch)
		assert.Equal(t, "main", c.Base)
		assert.Equal(t, "42", c.Key)
	})

	t.Run("PR - Branch Policy", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Setenv("SYSTEM_PULLREQUEST_SOURCEBRANCH", "feat/test-azure")
		os.Setenv("SYSTEM_PULLREQUEST_TARGETBRANCH", "main")
		os.Setenv("SYSTEM_PULLREQUEST_PULLREQUESTID", "123456789")
		os.Setenv("SYSTEM_PULLREQUEST_PULLREQUESTNUMBER", "42")
		os.Setenv("BUILD_REASON", "PullRequest")

		p := AzureDevOpsConfigProvider{}
		c := p.GetPullRequestConfig()

		assert.True(t, p.IsPullRequest())
		assert.Equal(t, "feat/test-azure", c.Branch)
		assert.Equal(t, "main", c.Base)
		assert.Equal(t, "42", c.Key)
	})

	t.Run("Azure DevOps - false", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()

		os.Setenv("AZURE_HTTP_USER_AGENT", "false")

		o := DetectOrchestrator()

		assert.Equal(t, Orchestrator(Unknown), o)
	})

	t.Run("env variables", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Setenv("SYSTEM_COLLECTIONURI", "https://dev.azure.com/fabrikamfiber/")
		os.Setenv("SYSTEM_TEAMPROJECTID", "123a4567-ab1c-12a1-1234-123456ab7890")
		os.Setenv("BUILD_BUILDID", "42")
		os.Setenv("AGENT_VERSION", "2.193.0")
		os.Setenv("BUILD_BUILDNUMBER", "20220318.16")
		os.Setenv("BUILD_REPOSITORY_NAME", "repo-org/repo-name")

		p := AzureDevOpsConfigProvider{}

		assert.Equal(t, "https://dev.azure.com/fabrikamfiber/", p.getSystemCollectionURI())
		assert.Equal(t, "123a4567-ab1c-12a1-1234-123456ab7890", p.getTeamProjectId())
		assert.Equal(t, "42", p.getBuildId())          // Don't confuse getBuildId and GetBuildId!
		assert.Equal(t, "20220318.16", p.GetBuildId()) // buildNumber is used in the UI
		assert.Equal(t, "2.193.0", p.OrchestratorVersion())
		assert.Equal(t, "repo-org/repo-name", p.GetJobName())

	})
}

func TestAzureDevOpsConfigProvider_GetPipelineStartTime(t *testing.T) {
	type fields struct {
		client  http.Client
		options http.ClientOptions
	}

	tests := []struct {
		name           string
		fields         fields
		apiInformation map[string]interface{}
		want           time.Time
	}{
		{
			name: "Retrieve correct time",
			fields: fields{
				client:  http.Client{},
				options: http.ClientOptions{},
			},
			apiInformation: map[string]interface{}{"startTime": "2022-03-18T12:30:42.0Z"},
			want:           time.Date(2022, time.March, 18, 12, 30, 42, 0, time.UTC),
		},
		{
			name: "Empty apiInformation",
			fields: fields{
				client:  http.Client{},
				options: http.ClientOptions{},
			},
			apiInformation: map[string]interface{}{},
			want:           time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "apiInformation does not contain key",
			fields: fields{
				client:  http.Client{},
				options: http.ClientOptions{},
			},
			apiInformation: map[string]interface{}{"Somekey": "somevalue"},
			want:           time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "apiInformation contains malformed date",
			fields: fields{
				client:  http.Client{},
				options: http.ClientOptions{},
			},
			apiInformation: map[string]interface{}{"startTime": "2022-03/18 12:30:42.0Z"},
			want:           time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AzureDevOpsConfigProvider{
				client:  tt.fields.client,
				options: tt.fields.options,
			}
			apiInformation = tt.apiInformation
			pipelineStartTime := a.GetPipelineStartTime()
			assert.Equalf(t, tt.want, pipelineStartTime, "GetPipelineStartTime()")
		})
	}
}

func TestAzureDevOpsConfigProvider_GetBuildStatus(t *testing.T) {

	tests := []struct {
		name   string
		want   string
		envVar string
	}{
		{
			name:   "Success",
			envVar: "Succeeded",
			want:   "SUCCESS",
		},
		{
			name:   "aborted",
			envVar: "Canceled",
			want:   "ABORTED",
		},
		{
			name:   "failure",
			envVar: "failed",
			want:   "FAILURE",
		},
		{
			name:   "other",
			envVar: "some other status",
			want:   "FAILURE",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer resetEnv(os.Environ())
			os.Clearenv()
			os.Setenv("AGENT_JOBSTATUS", tt.envVar)
			a := &AzureDevOpsConfigProvider{}

			assert.Equalf(t, tt.want, a.GetBuildStatus(), "GetBuildStatus()")
		})
	}
}
