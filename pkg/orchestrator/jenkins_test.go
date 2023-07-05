//go:build unit
// +build unit

package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/pkg/errors"

	"net/http"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestJenkins(t *testing.T) {
	t.Run("BranchBuild", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Setenv("JENKINS_URL", "FOO BAR BAZ")
		os.Setenv("BUILD_URL", "https://jaas.url/job/foo/job/bar/job/main/1234/")
		os.Setenv("BRANCH_NAME", "main")
		os.Setenv("GIT_COMMIT", "abcdef42713")
		os.Setenv("GIT_URL", "github.com/foo/bar")

		p, _ := NewOrchestratorSpecificConfigProvider()

		assert.False(t, p.IsPullRequest())
		assert.Equal(t, "https://jaas.url/job/foo/job/bar/job/main/1234/", p.GetBuildURL())
		assert.Equal(t, "main", p.GetBranch())
		assert.Equal(t, "refs/heads/main", p.GetReference())
		assert.Equal(t, "abcdef42713", p.GetCommit())
		assert.Equal(t, "github.com/foo/bar", p.GetRepoURL())
		assert.Equal(t, "Jenkins", p.OrchestratorType())
	})

	t.Run("PR", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Setenv("BRANCH_NAME", "PR-42")
		os.Setenv("CHANGE_BRANCH", "feat/test-jenkins")
		os.Setenv("CHANGE_TARGET", "main")
		os.Setenv("CHANGE_ID", "42")

		p := JenkinsConfigProvider{}
		c := p.GetPullRequestConfig()

		assert.True(t, p.IsPullRequest())
		assert.Equal(t, "refs/pull/42/head", p.GetReference())
		assert.Equal(t, "feat/test-jenkins", c.Branch)
		assert.Equal(t, "main", c.Base)
		assert.Equal(t, "42", c.Key)
	})

	t.Run("env variables", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Setenv("JENKINS_HOME", "/var/lib/jenkins")
		os.Setenv("BUILD_ID", "1234")
		os.Setenv("JOB_URL", "https://jaas.url/job/foo/job/bar/job/main")
		os.Setenv("JENKINS_VERSION", "42")
		os.Setenv("JOB_NAME", "foo/bar/BRANCH")
		os.Setenv("STAGE_NAME", "Promote")
		os.Setenv("BUILD_URL", "https://jaas.url/job/foo/job/bar/job/main/1234/")
		os.Setenv("STAGE_NAME", "Promote")

		p := JenkinsConfigProvider{}

		assert.Equal(t, "/var/lib/jenkins", p.getJenkinsHome())
		assert.Equal(t, "1234", p.GetBuildID())
		assert.Equal(t, "https://jaas.url/job/foo/job/bar/job/main", p.GetJobURL())
		assert.Equal(t, "42", p.OrchestratorVersion())
		assert.Equal(t, "Jenkins", p.OrchestratorType())
		assert.Equal(t, "foo/bar/BRANCH", p.GetJobName())
		assert.Equal(t, "Promote", p.GetStageName())
		assert.Equal(t, "https://jaas.url/job/foo/job/bar/job/main/1234/", p.GetBuildURL())

	})
}

func TestJenkinsConfigProvider_GetPipelineStartTime(t *testing.T) {
	type fields struct {
		client  piperhttp.Client
		options piperhttp.ClientOptions
	}
	tests := []struct {
		name                    string
		fields                  fields
		want                    time.Time
		wantHTTPErr             bool
		wantHTTPStatusCodeError bool
		wantHTTPJSONParseError  bool
	}{
		{
			name:                    "Retrieve correct time",
			want:                    time.Date(2022, time.March, 21, 22, 30, 0, 0, time.UTC),
			wantHTTPErr:             false,
			wantHTTPStatusCodeError: false,
		},
		{
			name:                    "ParseHTTPResponseBodyJSON error",
			want:                    time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
			wantHTTPErr:             false,
			wantHTTPStatusCodeError: false,
		},
		{
			name:                    "GetRequest fails",
			want:                    time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
			wantHTTPErr:             true,
			wantHTTPStatusCodeError: false,
		},
		{
			name:                    "response code != 200 http.StatusNoContent",
			want:                    time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
			wantHTTPErr:             false,
			wantHTTPStatusCodeError: true,
		},
		{
			name:                    "parseResponseBodyJson fails",
			want:                    time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
			wantHTTPErr:             false,
			wantHTTPStatusCodeError: false,
			wantHTTPJSONParseError:  true,
		},
	}

	j := &JenkinsConfigProvider{
		client: piperhttp.Client{},
	}
	j.client.SetOptions(piperhttp.ClientOptions{
		MaxRequestDuration:        5 * time.Second,
		Token:                     "TOKEN",
		TransportSkipVerification: true,
		UseDefaultTransport:       true,
		MaxRetries:                -1,
	})
	httpmock.Activate()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer resetEnv(os.Environ())
			os.Clearenv()
			buildURl := "https://jaas.url/job/foo/job/bar/job/main/1234/"
			os.Setenv("BUILD_URL", buildURl)

			fakeUrl := buildURl + "api/json"
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
						return httpmock.NewJsonResponse(200, "timestamp:asdffd")
					}
					return httpmock.NewStringResponse(200, "{\"timestamp\":1647901800932,\"url\":\"https://jaas.url/view/piperpipelines/job/foo/job/bar/job/main/3731/\"}"), nil
				},
			)

			assert.Equalf(t, tt.want, j.GetPipelineStartTime(), "GetPipelineStartTime()")
		})
	}
}

func TestJenkinsConfigProvider_GetBuildStatus(t *testing.T) {
	apiSuccess := []byte(`{ "queueId":376475,
				"result":"SUCCESS",
				"timestamp":1647946800925
				}`)
	apiFailure := []byte(`{ "queueId":376475,
				"result":"FAILURE",
				"timestamp":1647946800925
				}`)
	apiAborted := []byte(`{ "queueId":376475,
				"result":"ABORTED",
				"timestamp":1647946800925
				}`)

	apiOTHER := []byte(`{ "queueId":376475,
				"result":"SOMETHING",
				"timestamp":1647946800925
				}`)

	tests := []struct {
		name           string
		want           string
		apiInformation []byte
	}{
		{
			name:           "SUCCESS",
			apiInformation: apiSuccess,
			want:           "SUCCESS",
		},
		{
			name:           "ABORTED",
			apiInformation: apiAborted,
			want:           "ABORTED",
		},
		{
			name:           "FAILURE",
			apiInformation: apiFailure,
			want:           "FAILURE",
		},
		{
			name:           "Unknown FAILURE",
			apiInformation: apiOTHER,
			want:           "FAILURE",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var apiInformation map[string]interface{}
			err := json.Unmarshal(tt.apiInformation, &apiInformation)
			if err != nil {
				t.Fatal("could not parse json:", err)
			}
			j := &JenkinsConfigProvider{
				apiInformation: apiInformation,
			}
			assert.Equalf(t, tt.want, j.GetBuildStatus(), "GetBuildStatus()")
		})
	}
}

func TestJenkinsConfigProvider_GetBuildReason(t *testing.T) {
	apiJsonSchedule := []byte(`{
				"_class": "org.jenkinsci.plugins.workflow.job.WorkflowRun",
				"actions": [{
						"_class": "hudson.model.CauseAction",
						"causes": [{
							"_class": "hudson.triggers.TimerTrigger$TimerTriggerCause",
							"shortDescription": "Started by timer"
						}]
					},
					{
						"_class": "jenkins.metrics.impl.TimeInQueueAction",
						"blockedDurationMillis": "0"
					}
				]
				}`)

	apiJSONManual := []byte(`{
				"_class": "org.jenkinsci.plugins.workflow.job.WorkflowRun",
				"actions": [{
						"_class": "hudson.model.CauseAction",
						"causes": [{
							"_class": "hudson.model.Cause$UserIdCause",
							"shortDescription": "Started by user John Doe",
							"userId": "i12345",
							"userName": "John Doe"
						}]
					},
					{
						"_class": "jenkins.metrics.impl.TimeInQueueAction",
						"blockedDurationMillis": "0"
					}
				]
				}`)

	apiJSONPullRequest := []byte(`{
				"_class": "org.jenkinsci.plugins.workflow.job.WorkflowRun",
				"actions": [ {
					    "_class": "hudson.model.CauseAction",
					    "causes": [
						{
						    "_class": "jenkins.branch.BranchEventCause",
						    "shortDescription": "Pull request #1511 opened"
						}
					    ]
					}]
				}`)

	apiJSONResourceTrigger := []byte(`{
				"_class": "org.jenkinsci.plugins.workflow.job.WorkflowRun",
				"actions": [ {
					    "_class": "hudson.model.CauseAction",
					    "causes": [
							{
							    "_class": "org.jenkinsci.plugins.workflow.support.steps.build.BuildUpstreamCause",
							    "shortDescription": "Started by upstream project \"dummy/dummy/PR-1234\" build number 42",
							    "upstreamBuild": 24,
							    "upstreamProject": "dummy/dummy/PR-1234",
							    "upstreamUrl": "job/dummy/job/dummy/job/PR-1234/"
							}
						    ]
					}]
				}`)

	apiJSONUnknown := []byte(`{
				"_class": "org.jenkinsci.plugins.workflow.job.WorkflowRun",
				"actions": [{
						"_class": "hudson.model.CauseAction",
						"causes": [{
							"_class": "hudson.model.RandomThingHere",
							"shortDescription": "Something"
						}]
					},
					{
						"_class": "jenkins.metrics.impl.TimeInQueueAction",
						"blockedDurationMillis": "0"
					}
				]
				}`)

	tests := []struct {
		name           string
		apiInformation []byte
		want           string
	}{
		{
			name:           "Manual trigger",
			apiInformation: apiJSONManual,
			want:           "Manual",
		},
		{
			name:           "Scheduled trigger",
			apiInformation: apiJsonSchedule,
			want:           "Schedule",
		},
		{
			name:           "PullRequest trigger",
			apiInformation: apiJSONPullRequest,
			want:           "PullRequest",
		},
		{
			name:           "ResourceTrigger trigger",
			apiInformation: apiJSONResourceTrigger,
			want:           "ResourceTrigger",
		},
		{
			name:           "Unknown",
			apiInformation: apiJSONUnknown,
			want:           "Unknown",
		},
		{
			name:           "Empty api",
			apiInformation: []byte(`{}`),
			want:           "Unknown",
		},
		{
			name: "Empty action api",
			apiInformation: []byte(`{
				"actions": [{}]
			}`),
			want: "Unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var apiInformation map[string]interface{}
			err := json.Unmarshal(tt.apiInformation, &apiInformation)
			if err != nil {
				t.Fatal("could not parse json:", err)
			}
			j := &JenkinsConfigProvider{apiInformation: apiInformation}

			assert.Equalf(t, tt.want, j.GetBuildReason(), "GetBuildReason()")
		})
	}
}

func TestJenkinsConfigProvider_getAPIInformation(t *testing.T) {

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
			j := &JenkinsConfigProvider{
				apiInformation: tt.apiInformation,
			}
			j.client.SetOptions(piperhttp.ClientOptions{
				MaxRequestDuration:        5 * time.Second,
				Token:                     "TOKEN",
				TransportSkipVerification: true,
				UseDefaultTransport:       true, // need to use default transport for http mock
				MaxRetries:                -1,
			})

			defer resetEnv(os.Environ())
			os.Clearenv()
			os.Setenv("BUILD_URL", "https://jaas.url/job/foo/job/bar/job/main/1234/")

			fakeUrl := "https://jaas.url/job/foo/job/bar/job/main/1234/api/json"
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
			j.fetchAPIInformation()
			assert.Equal(t, tt.wantAPIInformation, j.apiInformation)
		})
	}
}

func TestJenkinsConfigProvider_GetLog(t *testing.T) {

	tests := []struct {
		name                    string
		want                    []byte
		wantErr                 assert.ErrorAssertionFunc
		wantHTTPErr             bool
		wantHTTPStatusCodeError bool
	}{
		{
			name:    "Successfully got log file",
			want:    []byte("Success!"),
			wantErr: assert.NoError,
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
			j := &JenkinsConfigProvider{}
			j.client.SetOptions(piperhttp.ClientOptions{
				MaxRequestDuration:        5 * time.Second,
				Token:                     "TOKEN",
				TransportSkipVerification: true,
				UseDefaultTransport:       true, // need to use default transport for http mock
				MaxRetries:                -1,
			})

			defer resetEnv(os.Environ())
			os.Clearenv()
			os.Setenv("BUILD_URL", "https://jaas.url/job/foo/job/bar/job/main/1234/")

			fakeUrl := "https://jaas.url/job/foo/job/bar/job/main/1234/consoleText"
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
					return httpmock.NewStringResponse(200, "Success!"), nil
				},
			)

			got, err := j.GetLog()
			if !tt.wantErr(t, err, fmt.Sprintf("GetLog()")) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetLog()")
		})
	}
}

func TestJenkinsConfigProvider_InitOrchestratorProvider(t *testing.T) {

	tests := []struct {
		name           string
		settings       *OrchestratorSettings
		apiInformation map[string]interface{}
	}{
		{
			name:     "Init, test empty apiInformation",
			settings: &OrchestratorSettings{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JenkinsConfigProvider{}
			j.InitOrchestratorProvider(tt.settings)
			var expected map[string]interface{}
			assert.Equal(t, j.apiInformation, expected)
		})
	}
}

func TestJenkinsConfigProvider_GetChangeSet(t *testing.T) {

	changeSetTwo := []byte(`{
"displayName": "#531",
"duration": 424269,
"changeSets": [
        {
            "_class": "hudson.plugins.git.GitChangeSetList",
            "items": [
                {
                    "_class": "hudson.plugins.git.GitChangeSet",
                    "commitId": "987654321",
                    "timestamp": 1655057520000
                },
		{
                    "_class": "hudson.plugins.git.GitChangeSet",
                    "commitId": "123456789",
                    "timestamp": 1656057520000
                }
            ],
            "kind": "git"
        }
    ]
}`)

	changeSetMultiple := []byte(`{
"displayName": "#531",
"duration": 424269,
"changeSets": [
    {
        "_class": "hudson.plugins.git.GitChangeSetList",
        "items": [
            {
                "_class": "hudson.plugins.git.GitChangeSet",
                "commitId": "987654321",
                "timestamp": 1655057520000
            },
            {
                "_class": "hudson.plugins.git.GitChangeSet",
                "commitId": "123456789",
                "timestamp": 1656057520000
            }
        ],
        "kind": "git"
    },
    {
        "_class": "hudson.plugins.git.GitChangeSetList",
        "items": [
            {
                "_class": "hudson.plugins.git.GitChangeSet",
                "commitId": "456789123",
                "timestamp": 1659948036000
            },
            {
                "_class": "hudson.plugins.git.GitChangeSet",
                "commitId": "654717777",
                "timestamp": 1660053494000
            }
        ],
        "kind": "git"
    }
]
}`)

	changeSetEmpty := []byte(`{
"displayName": "#531",
"duration": 424269,
"changeSets": []
}`)
	changeSetNotAvailable := []byte(`{
"displayName": "#531",
"duration": 424269
}`)
	tests := []struct {
		name          string
		want          []ChangeSet
		testChangeSet []byte
	}{
		{
			name: "success",
			want: []ChangeSet{
				{CommitId: "987654321", Timestamp: "1655057520000"},
				{CommitId: "123456789", Timestamp: "1656057520000"},
			},
			testChangeSet: changeSetTwo,
		},
		{
			name: "success multiple",
			want: []ChangeSet{
				{CommitId: "987654321", Timestamp: "1655057520000"},
				{CommitId: "123456789", Timestamp: "1656057520000"},
				{CommitId: "456789123", Timestamp: "1659948036000"},
				{CommitId: "654717777", Timestamp: "1660053494000"},
			},
			testChangeSet: changeSetMultiple,
		},
		{
			name:          "failure - changeSet empty",
			want:          []ChangeSet(nil),
			testChangeSet: changeSetEmpty,
		},
		{
			name:          "failure - no changeSet found",
			want:          []ChangeSet(nil),
			testChangeSet: changeSetNotAvailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var apiInformation map[string]interface{}
			err := json.Unmarshal(tt.testChangeSet, &apiInformation)
			if err != nil {
				t.Fatal("could not parse json:", err)
			}
			j := &JenkinsConfigProvider{apiInformation: apiInformation}
			assert.Equalf(t, tt.want, j.GetChangeSet(), "GetChangeSet()")
		})
	}
}
