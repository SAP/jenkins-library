//go:build unit
// +build unit

package sonar

import (
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
)

func TestIssueService(t *testing.T) {
	testURL := "https://example.org"
	t.Run("success", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		sender := &piperhttp.Client{}
		sender.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		// add response handler
		httpmock.RegisterResponder(http.MethodGet, testURL+"/api/"+EndpointIssuesSearch+"", httpmock.NewStringResponder(http.StatusOK, responseIssueSearchCritical))
		// create service instance
		serviceUnderTest := NewIssuesService(testURL, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, sender)
		// test
		count, _, err := serviceUnderTest.GetNumberOfBlockerIssues()
		// assert
		assert.NoError(t, err)
		assert.Equal(t, 111, count)
		assert.Equal(t, 1, httpmock.GetTotalCallCount(), "unexpected number of requests")
	})
	t.Run("error", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		sender := &piperhttp.Client{}
		sender.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		// add response handler
		httpmock.RegisterResponder(http.MethodGet, testURL+"/api/"+EndpointIssuesSearch+"", httpmock.NewStringResponder(http.StatusNotFound, responseIssueSearchError))
		// create service instance
		serviceUnderTest := NewIssuesService(testURL, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, sender)
		// test
		count, _, err := serviceUnderTest.GetNumberOfCriticalIssues()
		// assert
		assert.Error(t, err)
		assert.Equal(t, -1, count)
		assert.Equal(t, 1, httpmock.GetTotalCallCount(), "unexpected number of requests")
	})
	t.Run("", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		sender := &piperhttp.Client{}
		sender.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		// add response handler
		httpmock.RegisterResponder(http.MethodGet, testURL+"/api/"+EndpointIssuesSearch+"", httpmock.NewStringResponder(http.StatusOK, responseIssueSearchCritical))
		// create service instance
		serviceUnderTest := NewIssuesService(testURL, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, sender)
		// test
		countMajor, _, err := serviceUnderTest.GetNumberOfMajorIssues()
		countMinor, _, err := serviceUnderTest.GetNumberOfMinorIssues()
		countInfo, _, err := serviceUnderTest.GetNumberOfInfoIssues()
		// assert
		assert.NoError(t, err)
		assert.Equal(t, 111, countMajor)
		assert.Equal(t, 111, countMinor)
		assert.Equal(t, 111, countInfo)
		assert.Equal(t, 3, httpmock.GetTotalCallCount(), "unexpected number of requests")
	})
}

const responseIssueSearchError = `{
  "errors": [
    {
      "msg": "At least one of the following parameters must be specified: organization, projects, projectKeys (deprecated), componentKeys, componentUuids (deprecated), assignees, issues"
    }
  ]
}`

const responseIssueSearchCritical = `{
  "total": 111,
  "p": 1,
  "ps": 1,
  "paging": {
    "pageIndex": 1,
    "pageSize": 1,
    "total": 111
  },
  "effortTotal": 1176,
  "debtTotal": 1176,
  "issues": [
    {
      "key": "AXW3MmCVOYWf3_DBLGvL",
      "rule": "go:S3776",
      "severity": "CRITICAL",
      "component": "SAP_jenkins-library:cmd/fortifyExecuteScan.go",
      "project": "SAP_jenkins-library",
      "line": 647,
      "hash": "a154a51bdb1502a2ac057a348d08e7f6",
      "textRange": {
        "startLine": 647,
        "endLine": 647,
        "startOffset": 5,
        "endOffset": 23
      },
      "flows": [
        {
          "locations": [
            {
              "component": "SAP_jenkins-library:cmd/fortifyExecuteScan.go",
              "textRange": {
                "startLine": 651,
                "endLine": 651,
                "startOffset": 1,
                "endOffset": 3
              },
              "msg": "+1"
            }
          ]
        }
      ],
      "status": "OPEN",
      "message": "Refactor this method to reduce its Cognitive Complexity from 16 to the 15 allowed.",
      "effort": "6min",
      "debt": "6min",
      "assignee": "CCFenner@github",
      "author": "33484802+olivernocon@users.noreply.github.com",
      "tags": [],
      "creationDate": "2020-11-11T11:06:04+0100",
      "updateDate": "2020-11-11T11:06:04+0100",
      "type": "CODE_SMELL",
      "organization": "sap-1"
    }
  ],
  "components": [
    {
      "organization": "sap-1",
      "key": "SAP_jenkins-library:cmd/fortifyExecuteScan.go",
      "uuid": "AXVKXJIlrkwsFznOfAie",
      "enabled": true,
      "qualifier": "FIL",
      "name": "fortifyExecuteScan.go",
      "longName": "cmd/fortifyExecuteScan.go",
      "path": "cmd/fortifyExecuteScan.go"
    },
    {
      "organization": "sap-1",
      "key": "SAP_jenkins-library",
      "uuid": "AXVFg_8dh6o1O3pu_MCx",
      "enabled": true,
      "qualifier": "TRK",
      "name": "jenkins-library",
      "longName": "jenkins-library"
    }
  ],
  "organizations": [
    {
      "key": "sap-1",
      "name": "SAP"
    }
  ],
  "facets": []
}`
