//go:build unit

package orchestrator

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/jarcoal/httpmock"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func TestAzure(t *testing.T) {
	t.Run("Azure - BranchBuild", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Setenv("AZURE_HTTP_USER_AGENT", "FOO BAR BAZ")
		os.Setenv("BUILD_SOURCEBRANCH", "refs/heads/feat/test-azure")
		os.Setenv("SYSTEM_TEAMFOUNDATIONCOLLECTIONURI", "https://pogo.sap/")
		os.Setenv("SYSTEM_TEAMPROJECT", "foo")
		os.Setenv("BUILD_BUILDID", "42")
		os.Setenv("BUILD_SOURCEVERSION", "abcdef42713")
		os.Setenv("BUILD_REPOSITORY_URI", "github.com/foo/bar")
		os.Setenv("SYSTEM_DEFINITIONNAME", "bar")
		os.Setenv("SYSTEM_DEFINITIONID", "1234")
		p, _ := GetOrchestratorConfigProvider(nil)

		assert.False(t, p.IsPullRequest())
		assert.Equal(t, "feat/test-azure", p.Branch())
		assert.Equal(t, "refs/heads/feat/test-azure", p.GitReference())
		assert.Equal(t, "https://pogo.sap/foo/bar/_build/results?buildId=42", p.BuildURL())
		assert.Equal(t, "abcdef42713", p.CommitSHA())
		assert.Equal(t, "github.com/foo/bar", p.RepoURL())
		assert.Equal(t, "Azure", p.OrchestratorType())
		assert.Equal(t, "https://pogo.sap/foo/bar/_build?definitionId=1234", p.JobURL())
	})

	t.Run("PR", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Setenv("SYSTEM_PULLREQUEST_SOURCEBRANCH", "feat/test-azure")
		os.Setenv("SYSTEM_PULLREQUEST_TARGETBRANCH", "main")
		os.Setenv("SYSTEM_PULLREQUEST_PULLREQUESTID", "42")
		os.Setenv("BUILD_REASON", "PullRequest")

		p := azureDevopsConfigProvider{}
		c := p.PullRequestConfig()

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

		p := azureDevopsConfigProvider{}
		c := p.PullRequestConfig()

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

		p := azureDevopsConfigProvider{}

		assert.Equal(t, "https://dev.azure.com/fabrikamfiber/", p.getSystemCollectionURI())
		assert.Equal(t, "123a4567-ab1c-12a1-1234-123456ab7890", p.getTeamProjectID())
		assert.Equal(t, "42", p.getAzureBuildID())  // Don't confuse getAzureBuildID and provider.BuildID!
		assert.Equal(t, "20220318.16", p.BuildID()) // buildNumber is used in the UI
		assert.Equal(t, "2.193.0", p.OrchestratorVersion())
		assert.Equal(t, "repo-org/repo-name", p.JobName())

	})
}

func TestAzureDevOpsConfigProvider_GetPipelineStartTime(t *testing.T) {

	tests := []struct {
		name           string
		apiInformation map[string]interface{}
		want           time.Time
	}{
		{
			name:           "Retrieve correct time",
			apiInformation: map[string]interface{}{"startTime": "2022-03-18T12:30:42.0Z"},
			want:           time.Date(2022, time.March, 18, 12, 30, 42, 0, time.UTC),
		},
		{
			name:           "Empty apiInformation",
			apiInformation: map[string]interface{}{},
			want:           time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:           "apiInformation does not contain key",
			apiInformation: map[string]interface{}{"someKey": "someValue"},
			want:           time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:           "apiInformation contains malformed date",
			apiInformation: map[string]interface{}{"startTime": "2022-03/18 12:30:42.0Z"},
			want:           time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &azureDevopsConfigProvider{}
			a.apiInformation = tt.apiInformation
			pipelineStartTime := a.PipelineStartTime()
			assert.Equalf(t, tt.want, pipelineStartTime, "PipelineStartTime()")
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
			a := &azureDevopsConfigProvider{}

			assert.Equalf(t, tt.want, a.BuildStatus(), "BuildStatus()")
		})
	}
}

func TestAzureDevOpsConfigProvider_getAPIInformation(t *testing.T) {
	tests := []struct {
		name                    string
		wantHTTPErr             bool
		wantHTTPStatusCodeError bool
		wantHTTPJSONParseError  bool
		apiInformation          map[string]interface{}
		wantAPIInformation      map[string]interface{}
	}{
		{
			name:               "success case",
			apiInformation:     map[string]interface{}{},
			wantAPIInformation: map[string]interface{}{"Success": "Case"},
		},
		{
			name:               "apiInformation already set",
			apiInformation:     map[string]interface{}{"API info": "set"},
			wantAPIInformation: map[string]interface{}{"API info": "set"},
		},
		{
			name:               "failed to get response",
			apiInformation:     map[string]interface{}{},
			wantHTTPErr:        true,
			wantAPIInformation: map[string]interface{}{},
		},
		{
			name:                    "response code != 200 http.StatusNoContent",
			wantHTTPStatusCodeError: true,
			apiInformation:          map[string]interface{}{},
			wantAPIInformation:      map[string]interface{}{},
		},
		{
			name:                   "parseResponseBodyJson fails",
			wantHTTPJSONParseError: true,
			apiInformation:         map[string]interface{}{},
			wantAPIInformation:     map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &azureDevopsConfigProvider{
				apiInformation: tt.apiInformation,
			}

			a.client.SetOptions(piperhttp.ClientOptions{
				MaxRequestDuration:        5 * time.Second,
				Token:                     "TOKEN",
				TransportSkipVerification: true,
				UseDefaultTransport:       true, // need to use default transport for http mock
				MaxRetries:                -1,
			})

			defer resetEnv(os.Environ())
			os.Clearenv()
			os.Setenv("SYSTEM_COLLECTIONURI", "https://dev.azure.com/fabrikamfiber/")
			os.Setenv("SYSTEM_TEAMPROJECTID", "123a4567-ab1c-12a1-1234-123456ab7890")
			os.Setenv("BUILD_BUILDID", "1234")

			fakeUrl := "https://dev.azure.com/fabrikamfiber/123a4567-ab1c-12a1-1234-123456ab7890/_apis/build/builds/1234/"
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("GET", fakeUrl,
				func(req *http.Request) (*http.Response, error) {
					if tt.wantHTTPErr {
						return nil, errors.New("this error shows up")
					}
					if tt.wantHTTPStatusCodeError {
						return &http.Response{
							Status:     "204",
							StatusCode: http.StatusNoContent,
							Request:    req,
						}, nil
					}
					if tt.wantHTTPJSONParseError {
						// Intentionally malformed JSON response
						return httpmock.NewJsonResponse(200, "timestamp:broken")
					}
					return httpmock.NewStringResponse(200, "{\"Success\":\"Case\"}"), nil
				},
			)

			a.fetchAPIInformation()
			assert.Equal(t, tt.wantAPIInformation, a.apiInformation)
		})
	}
}

func TestAzureDevOpsConfigProvider_GetLog(t *testing.T) {
	tests := []struct {
		name                    string
		want                    []byte
		wantErr                 assert.ErrorAssertionFunc
		wantHTTPErr             bool
		wantHTTPStatusCodeError bool
		wantLogCountError       bool
	}{
		{
			name:    "Successfully got log file",
			want:    []byte("Success"),
			wantErr: assert.NoError,
		},
		{
			name:              "Log count variable not available",
			want:              []byte(""),
			wantErr:           assert.NoError,
			wantLogCountError: true,
		},
		{
			name:        "HTTP error",
			want:        []byte(""),
			wantErr:     assert.Error,
			wantHTTPErr: true,
		},
		{
			name:                    "Status code error",
			want:                    []byte(""),
			wantErr:                 assert.NoError,
			wantHTTPStatusCodeError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &azureDevopsConfigProvider{}
			a.client.SetOptions(piperhttp.ClientOptions{
				MaxRequestDuration:        5 * time.Second,
				Token:                     "TOKEN",
				TransportSkipVerification: true,
				UseDefaultTransport:       true, // need to use default transport for http mock
				MaxRetries:                -1,
			})

			defer resetEnv(os.Environ())
			os.Clearenv()
			os.Setenv("SYSTEM_COLLECTIONURI", "https://dev.azure.com/fabrikamfiber/")
			os.Setenv("SYSTEM_TEAMPROJECTID", "123a4567-ab1c-12a1-1234-123456ab7890")
			os.Setenv("BUILD_BUILDID", "1234")

			fakeUrl := "https://dev.azure.com/fabrikamfiber/123a4567-ab1c-12a1-1234-123456ab7890/_apis/build/builds/1234/logs"
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("GET", fakeUrl+"/1",
				func(req *http.Request) (*http.Response, error) {
					return httpmock.NewStringResponse(200, "Success"), nil
				})
			httpmock.RegisterResponder("GET", fakeUrl,
				func(req *http.Request) (*http.Response, error) {
					if tt.wantHTTPErr {
						return nil, errors.New("this error shows up")
					}
					if tt.wantHTTPStatusCodeError {
						return &http.Response{
							Status:     "204",
							StatusCode: http.StatusNoContent,
							Request:    req,
						}, nil
					}
					if tt.wantLogCountError {
						return httpmock.NewJsonResponse(200, map[string]interface{}{
							"some": "value",
						})
					}
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"count": 1,
					})
				},
			)
			got, err := a.FullLogs()
			if !tt.wantErr(t, err, fmt.Sprintf("FullLogs()")) {
				return
			}
			assert.Equalf(t, tt.want, got, "FullLogs()")
		})
	}
}
