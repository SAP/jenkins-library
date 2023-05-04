//go:build unit
// +build unit

package cmd

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLastSuccessfullCommitSuccess(t *testing.T) {

	config := gctsRollbackOptions{
		Host:                      "http://testHost.com:50000",
		Client:                    "000",
		Repository:                "testRepo",
		GithubPersonalAccessToken: "3a09064f3029f5a304d69987ef8f95d1dfa6da44",
	}

	t.Run("return last successful commit", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 200, ResponseBody: `
		{
			"state": "success",
			"statuses": [
			    {
				"url": "https://github.com/repos/testUser/testRepo/statuses/c316a4af470991f9a3ca51a12c44354e72729e3d",
				"avatar_url": "https://github.com/avatars/u/50615?",
				"id": 81586547,
				"node_id": "MDEzOlN0YXR1c0iUdnRleHQ4MTU4NjkyNg==",
				"state": "success",
				"description": "This commit looks good",
				"target_url": "https://jenkins.instance.com/job/jobName/job/test/job/master/41/display/redirect",
				"context": "continuous-integration/jenkins/branch",
				"created_at": "2020-04-24T07:25:59Z",
				"updated_at": "2020-04-24T07:25:59Z"
			    }
			],
			"sha": "c316a4af470991f9a3ca51a12c44354e72729e3d",
			"total_count": 1,
			"repository": {
			    "id": 348933,
			    "node_id": "MDEwOlJlcG9zaXRdkgjzNDg3NDM=",
			    "name": "testRepo",
			    "full_name": "testUser/testRepo",
			    "private": true,
			    "owner": {
				"login": "testUser",
				"id": 50613,
				"node_id": "MDQ6VXKdigUwNjEz",
				"avatar_url": "https://github.com/avatars/u/50653?",
				"gravatar_id": "",
				"url": "https://github.com/users/testUser",
				"html_url": "https://github.com/testUser",
				"followers_url": "https://github.com/users/testUser/followers",
				"following_url": "https://github.com/users/testUser/following{/other_user}",
				"gists_url": "https://github.com/users/testUser/gists{/gist_id}",
				"starred_url": "https://github.com/users/testUser/starred{/owner}{/repo}",
				"subscriptions_url": "https://github.com/users/testUser/subscriptions",
				"organizations_url": "https://github.com/users/testUser/orgs",
				"repos_url": "https://github.com/users/testUser/repos",
				"events_url": "https://github.com/users/testUser/events{/privacy}",
				"received_events_url": "https://github.com/users/testUser/received_events",
				"type": "User",
				"site_admin": false
			    },
			    "html_url": "https://github.com/testUser/testRepo",
			    "description": "testing go lib",
			    "fork": false,
			    "url": "https://github.com/repos/testUser/testRepo",
			    "forks_url": "https://github.com/repos/testUser/testRepo/forks",
			    "keys_url": "https://github.com/repos/testUser/testRepo/keys{/key_id}",
			    "collaborators_url": "https://github.com/repos/testUser/testRepo/collaborators{/collaborator}",
			    "teams_url": "https://github.com/repos/testUser/testRepo/teams",
			    "hooks_url": "https://github.com/repos/testUser/testRepo/hooks",
			    "issue_events_url": "https://github.com/repos/testUser/testRepo/issues/events{/number}",
			    "events_url": "https://github.com/repos/testUser/testRepo/events",
			    "assignees_url": "https://github.com/repos/testUser/testRepo/assignees{/user}",
			    "branches_url": "https://github.com/repos/testUser/testRepo/branches{/branch}",
			    "tags_url": "https://github.com/repos/testUser/testRepo/tags",
			    "blobs_url": "https://github.com/repos/testUser/testRepo/git/blobs{/sha}",
			    "git_tags_url": "https://github.com/repos/testUser/testRepo/git/tags{/sha}",
			    "git_refs_url": "https://github.com/repos/testUser/testRepo/git/refs{/sha}",
			    "trees_url": "https://github.com/repos/testUser/testRepo/git/trees{/sha}",
			    "statuses_url": "https://github.com/repos/testUser/testRepo/statuses/{sha}",
			    "languages_url": "https://github.com/repos/testUser/testRepo/languages",
			    "stargazers_url": "https://github.com/repos/testUser/testRepo/stargazers",
			    "contributors_url": "https://github.com/repos/testUser/testRepo/contributors",
			    "subscribers_url": "https://github.com/repos/testUser/testRepo/subscribers",
			    "subscription_url": "https://github.com/repos/testUser/testRepo/subscription",
			    "commits_url": "https://github.com/repos/testUser/testRepo/commits{/sha}",
			    "git_commits_url": "https://github.com/repos/testUser/testRepo/git/commits{/sha}",
			    "comments_url": "https://github.com/repos/testUser/testRepo/comments{/number}",
			    "issue_comment_url": "https://github.com/repos/testUser/testRepo/issues/comments{/number}",
			    "contents_url": "https://github.com/repos/testUser/testRepo/contents/{+path}",
			    "compare_url": "https://github.com/repos/testUser/testRepo/compare/{base}...{head}",
			    "merges_url": "https://github.com/repos/testUser/testRepo/merges",
			    "archive_url": "https://github.com/repos/testUser/testRepo/{archive_format}{/ref}",
			    "downloads_url": "https://github.com/repos/testUser/testRepo/downloads",
			    "issues_url": "https://github.com/repos/testUser/testRepo/issues{/number}",
			    "pulls_url": "https://github.com/repos/testUser/testRepo/pulls{/number}",
			    "milestones_url": "https://github.com/repos/testUser/testRepo/milestones{/number}",
			    "notifications_url": "https://github.com/repos/testUser/testRepo/notifications{?since,all,participating}",
			    "labels_url": "https://github.com/repos/testUser/testRepo/labels{/name}",
			    "releases_url": "https://github.com/repos/testUser/testRepo/releases{/id}",
			    "deployments_url": "https://github.com/repos/testUser/testRepo/deployments"
			},
			"commit_url": "https://github.com/repos/testUser/testRepo/commits/c316a4af470991f9a3ca51a12c44354e72729e3d",
			"url": "https://github.com/repos/testUser/testRepo/commits/c316a4af470991f9a3ca51a12c44354e72729e3d/status"
		    }
		`}

		parsedURL, _ := url.Parse("https://github.com/testUser/testRepo")
		commitList := []string{"c316a4af470991f9a3ca51a12c44354e72729e3d", "579ba54ef025377319b6bb2e01f034c4f9b72026", "6bf1e447581c41ba421a3cce45495d369e99c8e5", "7852bb9be857a9439e1a3674f830b121d3fbd7c4"}

		successfullCommit, err := getLastSuccessfullCommit(&config, nil, &httpClient, parsedURL, commitList)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "https://api.github.com/repos/testUser/testRepo/commits/c316a4af470991f9a3ca51a12c44354e72729e3d/status", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

			t.Run("check token", func(t *testing.T) {
				assert.Equal(t, "Bearer 3a09064f3029f5a304d69987ef8f95d1dfa6da44", httpClient.Options.Token)
			})

			t.Run("check commit list", func(t *testing.T) {
				assert.Equal(t, "c316a4af470991f9a3ca51a12c44354e72729e3d", successfullCommit)
			})
		}
	})
}

func TestGetLastSuccessfullCommitFailure(t *testing.T) {

	config := gctsRollbackOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
	}

	t.Run("a http error occurred", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 500}

		parsedURL, _ := url.Parse("https://github.com/testUser/testRepo")
		commitList := []string{"c316a4af470991f9a3ca51a12c44354e72729e3d", "579ba54ef025377319b6bb2e01f034c4f9b72026", "6bf1e447581c41ba421a3cce45495d369e99c8e5", "7852bb9be857a9439e1a3674f830b121d3fbd7c4"}

		_, err := getLastSuccessfullCommit(&config, nil, &httpClient, parsedURL, commitList)

		assert.EqualError(t, err, "a http error occurred")
	})
}
func TestGetCommitsSuccess(t *testing.T) {

	config := gctsRollbackOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
	}

	t.Run("return list of commits", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 200, ResponseBody: `
		{
			"commits": [
			    {
				"id": "c316a4af470991f9a3ca51a12c44354e72729e3d",
				"author": "Test User",
				"authorMail": "test.user@example.com",
				"message": "test",
				"description": "test\n",
				"date": "2020-04-24 09:07:50"
			    },
			    {
				"id": "579ba54ef025377319b6bb2e01f034c4f9b72026",
				"author": "Test User",
				"authorMail": "test.user@example.com",
				"message": "add major feature",
				"description": "add major feature\n",
				"date": "2020-04-24 09:01:49"
			    },
			    {
				"id": "6bf1e447581c41ba421a3cce45495d369e99c8e5",
				"author": "Test User",
				"authorMail": "test.user@example.com",
				"message": "minor fix",
				"description": "minor fix\n",
				"date": "2020-04-24 08:56:58"
			    },
			    {
				"id": "7852bb9be857a9439e1a3674f830b121d3fbd7c4",
				"author": "Test User",
				"authorMail": "test.user@example.com",
				"message": "updated",
				"description": "updated\n",
				"date": "2020-04-24 08:56:51"
			    }
			]
		}
		`}

		commitList, err := getCommits(&config, nil, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo/getCommit?sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

			t.Run("check commit list", func(t *testing.T) {
				assert.Equal(t, []string{"c316a4af470991f9a3ca51a12c44354e72729e3d", "579ba54ef025377319b6bb2e01f034c4f9b72026", "6bf1e447581c41ba421a3cce45495d369e99c8e5", "7852bb9be857a9439e1a3674f830b121d3fbd7c4"}, commitList)
			})

		}

	})
}
func TestGetCommitsFailure(t *testing.T) {

	config := gctsRollbackOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
	}

	t.Run("a http error occurred", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 500, ResponseBody: `
		}
		"log": [
			{
			  "time": 20180606130524,
			  "user": "JENKINS",
			  "section": "REPOSITORY_FACTORY",
			  "action": "CREATE_REPOSITORY",
			  "severity": "INFO",
			  "message": "Start action CREATE_REPOSITORY review",
			  "code": "GCTS.API.410"
			}
		      ],
		      "errorLog": [
			{
			  "time": 20180606130524,
			  "user": "JENKINS",
			  "section": "REPOSITORY_FACTORY",
			  "action": "CREATE_REPOSITORY",
			  "severity": "INFO",
			  "message": "Start action CREATE_REPOSITORY review",
			  "code": "GCTS.API.410"
			}
		      ],
		      "exception": {
			"message": "repository_not_found",
			"description": "Repository not found",
			"code": 404
		      }
		}
		`}

		_, err := getCommits(&config, nil, &httpClient)

		assert.EqualError(t, err, "a http error occurred")
	})
}
func TestGetRepoInfoSuccess(t *testing.T) {

	config := gctsRollbackOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
	}

	t.Run("return struct of repository information", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 200, ResponseBody: `
		{
			"result": {
			    "rid": "testRepo",
			    "name": "testRepo",
			    "role": "SOURCE",
			    "type": "GIT",
			    "vsid": "BCH",
			    "status": "READY",
			    "branch": "master",
			    "url": "https://github.com/testUser/testRepo",
			    "createdBy": "testUser",
			    "createdDate": "2020-04-27",
			    "config": [
				{
				    "key": "CLIENT_VCS_CONNTYPE",
				    "value": "ssl",
				    "category": "CONNECTION"
				},
				{
				    "key": "CLIENT_VCS_URI",
				    "value": "https://github.com/testUser/testRepo",
				    "category": "CONNECTION"
				}
			    ],
			    "objects": 2,
			    "currentCommit": "c316a4af470991f9a3ca51a12c44354e72729e3d",
			    "connection": "ssl"
			}
		}
		`}

		repoInfo, err := getRepoInfo(&config, nil, &httpClient)

		repoInfoExpected := &getRepoInfoResponseBody{
			Result: struct {
				Rid           string `json:"rid"`
				Name          string `json:"name"`
				Role          string `json:"role"`
				Type          string `json:"type"`
				Vsid          string `json:"vsid"`
				Status        string `json:"status"`
				Branch        string `json:"branch"`
				URL           string `json:"url"`
				Version       string `json:"version"`
				Objects       int    `json:"objects"`
				CurrentCommit string `json:"currentCommit"`
				Connection    string `json:"connection"`
				Config        []struct {
					Key   string `json:"key"`
					Value string `json:"value"`
				} `json:"config"`
			}{
				Rid:           "testRepo",
				Name:          "testRepo",
				Role:          "SOURCE",
				Type:          "GIT",
				Vsid:          "BCH",
				Status:        "READY",
				Branch:        "master",
				URL:           "https://github.com/testUser/testRepo",
				Version:       "",
				Objects:       2,
				CurrentCommit: "c316a4af470991f9a3ca51a12c44354e72729e3d",
				Connection:    "ssl",
				Config: []struct {
					Key   string `json:"key"`
					Value string `json:"value"`
				}{
					{"CLIENT_VCS_CONNTYPE", "ssl"},
					{"CLIENT_VCS_URI", "https://github.com/testUser/testRepo"},
				},
			},
		}

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo?sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

			t.Run("check commit list", func(t *testing.T) {
				assert.Equal(t, repoInfoExpected, repoInfo)
			})
		}
	})
}

func TestGetRepoInfoFailure(t *testing.T) {

	config := gctsRollbackOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
	}

	t.Run("a http error occurred", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 500, ResponseBody: `
		}
		"log": [
			{
			  "time": 20180606130524,
			  "user": "JENKINS",
			  "section": "REPOSITORY_FACTORY",
			  "action": "CREATE_REPOSITORY",
			  "severity": "INFO",
			  "message": "Start action CREATE_REPOSITORY review",
			  "code": "GCTS.API.410"
			}
		      ],
		      "errorLog": [
			{
			  "time": 20180606130524,
			  "user": "JENKINS",
			  "section": "REPOSITORY_FACTORY",
			  "action": "CREATE_REPOSITORY",
			  "severity": "INFO",
			  "message": "Start action CREATE_REPOSITORY review",
			  "code": "GCTS.API.410"
			}
		      ],
		      "exception": {
			"message": "repository_not_found",
			"description": "Repository not found",
			"code": 404
		      }
		}
		`}

		_, err := getRepoInfo(&config, nil, &httpClient)

		assert.EqualError(t, err, "a http error occurred")
	})
}
func TestGetRepoHistorySuccess(t *testing.T) {

	config := gctsRollbackOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
	}

	t.Run("return struct of repository history", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 200, ResponseBody: `
		{
			"result": [
			  {
			    "rid": "com.example",
			    "checkoutTime": 20180606130524,
			    "fromCommit": "f1cdb6a032c1d8187c0990b51e94e8d8bb9898b2",
			    "toCommit": "f1cdb6a032c1d8187c0990b51e94e8d8bb9898b2",
			    "caller": "JOHNDOE",
			    "request": "SIDK1234567",
			    "type": "PULL"
			  }
			]
		}
		`}

		repoHistory, err := getRepoHistory(&config, nil, &httpClient)

		repoHistoryExpected := &getRepoHistoryResponseBody{
			Result: []struct {
				Rid          string `json:"rid"`
				CheckoutTime int64  `json:"checkoutTime"`
				FromCommit   string `json:"fromCommit"`
				ToCommit     string `json:"toCommit"`
				Caller       string `json:"caller"`
				Request      string `json:"request"`
				Type         string `json:"type"`
			}{
				{
					Rid:          "com.example",
					CheckoutTime: 20180606130524,
					FromCommit:   "f1cdb6a032c1d8187c0990b51e94e8d8bb9898b2",
					ToCommit:     "f1cdb6a032c1d8187c0990b51e94e8d8bb9898b2",
					Caller:       "JOHNDOE",
					Request:      "SIDK1234567",
					Type:         "PULL",
				},
			},
		}

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo/getHistory?sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

			t.Run("check commit list", func(t *testing.T) {
				assert.Equal(t, repoHistoryExpected, repoHistory)
			})
		}
	})
}

func TestGetRepoHistoryFailure(t *testing.T) {

	config := gctsRollbackOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
	}

	t.Run("a http error occurred", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 500, ResponseBody: `
		}
		"log": [
			{
			  "time": 20180606130524,
			  "user": "JENKINS",
			  "section": "REPOSITORY_FACTORY",
			  "action": "CREATE_REPOSITORY",
			  "severity": "INFO",
			  "message": "Start action CREATE_REPOSITORY review",
			  "code": "GCTS.API.410"
			}
		      ],
		      "errorLog": [
			{
			  "time": 20180606130524,
			  "user": "JENKINS",
			  "section": "REPOSITORY_FACTORY",
			  "action": "CREATE_REPOSITORY",
			  "severity": "INFO",
			  "message": "Start action CREATE_REPOSITORY review",
			  "code": "GCTS.API.410"
			}
		      ],
		      "exception": {
			"message": "repository_not_found",
			"description": "Repository not found",
			"code": 404
		      }
		}
		`}

		_, err := getRepoHistory(&config, nil, &httpClient)

		assert.EqualError(t, err, "a http error occurred")
	})
}
