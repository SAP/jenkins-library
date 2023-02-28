package orchestrator

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestGitHubActions(t *testing.T) {
	t.Run("BranchBuild", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Unsetenv("GITHUB_HEAD_REF")
		os.Setenv("GITHUB_ACTIONS", "true")
		os.Setenv("GITHUB_REF", "refs/heads/feat/test-gh-actions")
		os.Setenv("GITHUB_RUN_ID", "42")
		os.Setenv("GITHUB_SHA", "abcdef42713")
		os.Setenv("GITHUB_SERVER_URL", "github.com")
		os.Setenv("GITHUB_REPOSITORY", "foo/bar")

		p, _ := NewOrchestratorSpecificConfigProvider()

		assert.False(t, p.IsPullRequest())
		assert.Equal(t, "github.com/foo/bar/actions/runs/42", p.GetBuildURL())
		assert.Equal(t, "feat/test-gh-actions", p.GetBranch())
		assert.Equal(t, "refs/heads/feat/test-gh-actions", p.GetReference())
		assert.Equal(t, "abcdef42713", p.GetCommit())
		assert.Equal(t, "github.com/foo/bar", p.GetRepoURL())
		assert.Equal(t, "GitHubActions", p.OrchestratorType())
	})

	t.Run("PR", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Setenv("GITHUB_HEAD_REF", "feat/test-gh-actions")
		os.Setenv("GITHUB_BASE_REF", "main")
		os.Setenv("GITHUB_REF", "refs/pull/42/merge")

		p := GitHubActionsConfigProvider{}
		c := p.GetPullRequestConfig()

		assert.True(t, p.IsPullRequest())
		assert.Equal(t, "feat/test-gh-actions", c.Branch)
		assert.Equal(t, "main", c.Base)
		assert.Equal(t, "42", c.Key)
	})

	t.Run("Test get logs - success", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Unsetenv("GITHUB_HEAD_REF")
		os.Setenv("GITHUB_ACTIONS", "true")
		os.Setenv("GITHUB_REF_NAME", "feat/test-gh-actions")
		os.Setenv("GITHUB_REF", "refs/heads/feat/test-gh-actions")
		os.Setenv("GITHUB_RUN_ID", "42")
		os.Setenv("GITHUB_SHA", "abcdef42713")
		os.Setenv("GITHUB_REPOSITORY", "foo/bar")
		os.Setenv("GITHUB_URL", "https://github.com/")
		p := func() OrchestratorSpecificConfigProviding {
			g := GitHubActionsConfigProvider{}
			g.client = piperHttp.Client{}
			g.client.SetOptions(piperHttp.ClientOptions{
				MaxRequestDuration:        5 * time.Second,
				Password:                  "TOKEN",
				TransportSkipVerification: true,
				UseDefaultTransport:       true, // need to use default transport for http mock
				MaxRetries:                -1,
			})
			return &g
		}()
		stagesID := StagesID{
			Jobs: []Job{
				{ID: 123},
				{ID: 124},
				{ID: 125},
			},
		}
		logs := []string{
			"log_record1\n",
			"log_record2\n",
		}
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, "https://api.github.com/repos/foo/bar/actions/runs/42/jobs",
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewJsonResponse(200, stagesID)
			},
		)
		httpmock.RegisterResponder(http.MethodGet, "https://api.github.com/repos/foo/bar/actions/jobs/123/logs",
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(200, string(logs[0])), nil
			},
		)
		httpmock.RegisterResponder(http.MethodGet, "https://api.github.com/repos/foo/bar/actions/jobs/124/logs",
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(200, string(logs[1])), nil
			},
		)

		actual, err := p.GetLog()

		assert.NoError(t, err)
		assert.Equal(t, strings.Join(logs, ""), string(actual))
	})

	t.Run("Test get logs - error: failed to get stages ID", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Unsetenv("GITHUB_HEAD_REF")
		os.Setenv("GITHUB_ACTIONS", "true")
		os.Setenv("GITHUB_REF_NAME", "feat/test-gh-actions")
		os.Setenv("GITHUB_REF", "refs/heads/feat/test-gh-actions")
		os.Setenv("GITHUB_RUN_ID", "42")
		os.Setenv("GITHUB_SHA", "abcdef42713")
		os.Setenv("GITHUB_REPOSITORY", "foo/bar")
		os.Setenv("GITHUB_URL", "https://github.com/")
		p := func() OrchestratorSpecificConfigProviding {
			g := GitHubActionsConfigProvider{}
			g.client = piperHttp.Client{}
			g.client.SetOptions(piperHttp.ClientOptions{
				MaxRequestDuration:        5 * time.Second,
				Password:                  "TOKEN",
				TransportSkipVerification: true,
				UseDefaultTransport:       true, // need to use default transport for http mock
				MaxRetries:                -1,
			})
			return &g
		}()
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, "https://api.github.com/repos/foo/bar/actions/runs/42/jobs",
			func(req *http.Request) (*http.Response, error) {
				return nil, fmt.Errorf("err")
			},
		)
		actual, err := p.GetLog()

		assert.Nil(t, actual)
		assert.EqualError(t, err, "failed to get API data: HTTP request to https://api.github.com/repos/foo/bar/actions/runs/42/jobs failed with error: HTTP GET request to https://api.github.com/repos/foo/bar/actions/runs/42/jobs failed: Get \"https://api.github.com/repos/foo/bar/actions/runs/42/jobs\": err")
	})

	t.Run("Test get logs - failed to get logs", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Unsetenv("GITHUB_HEAD_REF")
		os.Setenv("GITHUB_ACTIONS", "true")
		os.Setenv("GITHUB_REF_NAME", "feat/test-gh-actions")
		os.Setenv("GITHUB_REF", "refs/heads/feat/test-gh-actions")
		os.Setenv("GITHUB_RUN_ID", "42")
		os.Setenv("GITHUB_SHA", "abcdef42713")
		os.Setenv("GITHUB_REPOSITORY", "foo/bar")
		os.Setenv("GITHUB_URL", "https://github.com/")
		p := func() OrchestratorSpecificConfigProviding {
			g := GitHubActionsConfigProvider{}
			g.client = piperHttp.Client{}
			g.client.SetOptions(piperHttp.ClientOptions{
				MaxRequestDuration:        5 * time.Second,
				Password:                  "TOKEN",
				TransportSkipVerification: true,
				UseDefaultTransport:       true, // need to use default transport for http mock
				MaxRetries:                -1,
			})
			return &g
		}()
		stagesID := StagesID{
			Jobs: []Job{
				{ID: 123},
				{ID: 124},
				{ID: 125},
			},
		}
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, "https://api.github.com/repos/foo/bar/actions/runs/42/jobs",
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewJsonResponse(200, stagesID)
			},
		)
		httpmock.RegisterResponder(http.MethodGet, "https://api.github.com/repos/foo/bar/actions/jobs/123/logs",
			func(req *http.Request) (*http.Response, error) {
				return nil, fmt.Errorf("err")
			},
		)

		actual, err := p.GetLog()

		assert.Nil(t, actual)
		assert.EqualError(t, err, "failed to get logs: failed to get API data: HTTP request to https://api.github.com/repos/foo/bar/actions/jobs/124/logs failed with error: HTTP GET request to https://api.github.com/repos/foo/bar/actions/jobs/124/logs failed: Get \"https://api.github.com/repos/foo/bar/actions/jobs/124/logs\": no responder found")
	})
}
