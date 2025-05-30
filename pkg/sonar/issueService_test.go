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
		// Severities
		var severities []Severity
		// test
		count, err := serviceUnderTest.GetNumberOfBlockerIssues(&severities)
		// assert
		assert.ElementsMatch(t, []Severity{{SeverityType: "BLOCKER", IssueType: "CODE_SMELL", IssueCount: 1}}, severities)
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
		// Severities
		var severities []Severity
		// test
		count, err := serviceUnderTest.GetNumberOfCriticalIssues(&severities)
		// assert
		assert.Error(t, err)
		assert.Equal(t, -1, count)
		assert.Equal(t, 1, httpmock.GetTotalCallCount(), "unexpected number of requests")
	})
	t.Run("multiple severities", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		sender := &piperhttp.Client{}
		sender.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		// add response handler
		httpmock.RegisterResponder(http.MethodGet, testURL+"/api/"+EndpointIssuesSearch+"", httpmock.NewStringResponder(http.StatusOK, responseIssueSearchCritical))
		// create service instance
		serviceUnderTest := NewIssuesService(testURL, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, sender)
		// Severities
		var severities []Severity
		// test
		countMajor, err := serviceUnderTest.GetNumberOfMajorIssues(&severities)
		countMinor, err := serviceUnderTest.GetNumberOfMinorIssues(&severities)
		countInfo, err := serviceUnderTest.GetNumberOfInfoIssues(&severities)
		// assert
		assert.ElementsMatch(t, []Severity{
			{SeverityType: "MAJOR", IssueType: "CODE_SMELL", IssueCount: 1},
			{SeverityType: "MINOR", IssueType: "CODE_SMELL", IssueCount: 1},
			{SeverityType: "INFO", IssueType: "CODE_SMELL", IssueCount: 1},
		}, severities)
		assert.NoError(t, err)
		assert.Equal(t, 111, countMajor)
		assert.Equal(t, 111, countMinor)
		assert.Equal(t, 111, countInfo)
		assert.Equal(t, 3, httpmock.GetTotalCallCount(), "unexpected number of requests")
	})
	t.Run("multiple issues", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		sender := &piperhttp.Client{}
		sender.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		// add response handler
		httpmock.RegisterResponder(http.MethodGet, testURL+"/api/"+EndpointIssuesSearch+"", httpmock.NewStringResponder(http.StatusOK, responseIssueSearchBug))
		// create service instance
		serviceUnderTest := NewIssuesService(testURL, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, sender)
		// Severities
		var severities []Severity
		// test
		countMajor, err := serviceUnderTest.GetNumberOfMajorIssues(&severities)
		countMinor, err := serviceUnderTest.GetNumberOfMinorIssues(&severities)
		// assert
		assert.ElementsMatch(t, []Severity{
			{SeverityType: "MAJOR", IssueType: "CODE_SMELL", IssueCount: 1},
			{SeverityType: "MAJOR", IssueType: "BUG", IssueCount: 1},
			{SeverityType: "MINOR", IssueType: "CODE_SMELL", IssueCount: 1},
			{SeverityType: "MINOR", IssueType: "BUG", IssueCount: 1},
		}, severities)
		assert.NoError(t, err)
		assert.Equal(t, 111, countMajor)
		assert.Equal(t, 111, countMinor)
		assert.Equal(t, 2, httpmock.GetTotalCallCount(), "unexpected number of requests")
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

const responseIssueSearchBug = `{
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
    },

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
      "type": "BUG",
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

func TestHotSpotService(t *testing.T) {
	testURL := "https://example.org"
	t.Run("success", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		sender := &piperhttp.Client{}
		sender.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		// add response handler
		httpmock.RegisterResponder(http.MethodGet, testURL+"/api/"+EndpointHotSpotsSearch+"", httpmock.NewStringResponder(http.StatusOK, responseHotSpotSearchMedium))
		// create service instance
		serviceUnderTest := NewIssuesService(testURL, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, sender)
		// Severities
		var hotspots []HotSpotSecurityIssue
		// test
		err := serviceUnderTest.GetHotSpotSecurityIssues(&hotspots)
		// assert
		assert.Equal(t, []HotSpotSecurityIssue{{IssueType: "MEDIUM", Count: 1}}, hotspots)
		assert.NoError(t, err)
		assert.Equal(t, 1, httpmock.GetTotalCallCount(), "unexpected number of requests")
	})
	t.Run("error", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		sender := &piperhttp.Client{}
		sender.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		// add response handler
		httpmock.RegisterResponder(http.MethodGet, testURL+"/api/"+EndpointHotSpotsSearch+"", httpmock.NewStringResponder(http.StatusNotFound, responseHotSpotSearchError))
		// create service instance
		serviceUnderTest := NewIssuesService(testURL, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, sender)
		// Severities
		var hotspots []HotSpotSecurityIssue
		// test
		err := serviceUnderTest.GetHotSpotSecurityIssues(&hotspots)
		// assert
		assert.Error(t, err)
		assert.Equal(t, 1, httpmock.GetTotalCallCount(), "unexpected number of requests")
	})
	t.Run("multiple severities", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		sender := &piperhttp.Client{}
		sender.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		// add response handler
		httpmock.RegisterResponder(http.MethodGet, testURL+"/api/"+EndpointHotSpotsSearch+"", httpmock.NewStringResponder(http.StatusOK, responseHotSpotSearchMultiple))
		// create service instance
		serviceUnderTest := NewIssuesService(testURL, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, sender)
		// Severities
		// Severities
		var hotspots []HotSpotSecurityIssue
		// test
		err := serviceUnderTest.GetHotSpotSecurityIssues(&hotspots)
		// assert
		assert.Equal(t, []HotSpotSecurityIssue{
			{IssueType: "MEDIUM", Count: 2},
			{IssueType: "LOW", Count: 1},
		}, hotspots)
		assert.NoError(t, err)
		assert.Equal(t, 1, httpmock.GetTotalCallCount(), "unexpected number of requests")
	})
}

const responseHotSpotSearchMedium = `{
  "paging": {
    "pageIndex": 1,
    "pageSize": 100,
    "total": 1
  },
  "hotspots": [
    {
      "key": "d1502ebc-941a-4262-8081-04914834d75a",
      "component": "java-camera-viewer:build/reports/configuration-cache/cbcm7qev21o4weakgfsv12gen/9cfjq3uac12x26p527bjxudhn/configuration-cache-report.html",
      "project": "java-camera-viewer",
      "securityCategory": "weak-cryptography",
      "vulnerabilityProbability": "MEDIUM",
      "status": "TO_REVIEW",
      "line": 658,
      "message": "Make sure that using this pseudorandom number generator is safe here.",
      "author": "",
      "creationDate": "2025-03-31T18:16:09+0000",
      "updateDate": "2025-04-23T11:02:47+0000",
      "textRange": {
        "startLine": 658,
        "endLine": 658,
        "startOffset": 16799,
        "endOffset": 16812
      },
      "flows": [],
      "ruleKey": "javascript:S2245",
      "messageFormattings": []
    }
  ],
  "components": [
    {
      "key": "java-camera-viewer:build/reports/configuration-cache/cbcm7qev21o4weakgfsv12gen/9cfjq3uac12x26p527bjxudhn/configuration-cache-report.html",
      "qualifier": "FIL",
      "name": "configuration-cache-report.html",
      "longName": "build/reports/configuration-cache/cbcm7qev21o4weakgfsv12gen/9cfjq3uac12x26p527bjxudhn/configuration-cache-report.html",
      "path": "build/reports/configuration-cache/cbcm7qev21o4weakgfsv12gen/9cfjq3uac12x26p527bjxudhn/configuration-cache-report.html"
    },
    {
      "key": "java-camera-viewer",
      "qualifier": "TRK",
      "name": "java-camera-viewer",
      "longName": "java-camera-viewer"
    }
  ]
}`

const responseHotSpotSearchError = `{
  "errors":[
    {
      "msg":"Project java-camera-viewer1 not found"
    }
  ]
}`

const responseHotSpotSearchMultiple = `{
  "paging": {
    "pageIndex": 1,
    "pageSize": 100,
    "total": 1
  },
  "hotspots": [
    {
      "key": "d1502ebc-941a-4262-8081-04914834d75a",
      "component": "java-camera-viewer:build/reports/configuration-cache/cbcm7qev21o4weakgfsv12gen/9cfjq3uac12x26p527bjxudhn/configuration-cache-report.html",
      "project": "java-camera-viewer",
      "securityCategory": "weak-cryptography",
      "vulnerabilityProbability": "MEDIUM",
      "status": "TO_REVIEW",
      "line": 658,
      "message": "Make sure that using this pseudorandom number generator is safe here.",
      "author": "",
      "creationDate": "2025-03-31T18:16:09+0000",
      "updateDate": "2025-04-23T11:02:47+0000",
      "textRange": {
        "startLine": 658,
        "endLine": 658,
        "startOffset": 16799,
        "endOffset": 16812
      },
      "flows": [],
      "ruleKey": "javascript:S2245",
      "messageFormattings": []
    },
    {
      "key": "d1502ebc-941a-4262-8081-04914834d75b",
      "component": "java-camera-viewer:build/reports/configuration-cache/cbcm7qev21o4weakgfsv12gen/9cfjq3uac12x26p527bjxudhn/configuration-cache-report.html",
      "project": "java-camera-viewer",
      "securityCategory": "weak-cryptography",
      "vulnerabilityProbability": "MEDIUM",
      "status": "TO_REVIEW",
      "line": 658,
      "message": "Make sure that using this pseudorandom number generator is safe here.",
      "author": "",
      "creationDate": "2025-03-31T18:16:09+0000",
      "updateDate": "2025-04-23T11:02:47+0000",
      "textRange": {
        "startLine": 658,
        "endLine": 658,
        "startOffset": 16799,
        "endOffset": 16812
      },
      "flows": [],
      "ruleKey": "javascript:S2245",
      "messageFormattings": []
    },
    {
      "key": "d1502ebc-941a-4262-8081-04914834d75c",
      "component": "java-camera-viewer:build/reports/configuration-cache/cbcm7qev21o4weakgfsv12gen/9cfjq3uac12x26p527bjxudhn/configuration-cache-report.html",
      "project": "java-camera-viewer",
      "securityCategory": "weak-cryptography",
      "vulnerabilityProbability": "LOW",
      "status": "TO_REVIEW",
      "line": 658,
      "message": "Make sure that using this pseudorandom number generator is safe here.",
      "author": "",
      "creationDate": "2025-03-31T18:16:09+0000",
      "updateDate": "2025-04-23T11:02:47+0000",
      "textRange": {
        "startLine": 658,
        "endLine": 658,
        "startOffset": 16799,
        "endOffset": 16812
      },
      "flows": [],
      "ruleKey": "javascript:S2245",
      "messageFormattings": []
    }
  ],
  "components": [
    {
      "key": "java-camera-viewer:build/reports/configuration-cache/cbcm7qev21o4weakgfsv12gen/9cfjq3uac12x26p527bjxudhn/configuration-cache-report.html",
      "qualifier": "FIL",
      "name": "configuration-cache-report.html",
      "longName": "build/reports/configuration-cache/cbcm7qev21o4weakgfsv12gen/9cfjq3uac12x26p527bjxudhn/configuration-cache-report.html",
      "path": "build/reports/configuration-cache/cbcm7qev21o4weakgfsv12gen/9cfjq3uac12x26p527bjxudhn/configuration-cache-report.html"
    },
    {
      "key": "java-camera-viewer",
      "qualifier": "TRK",
      "name": "java-camera-viewer",
      "longName": "java-camera-viewer"
    }
  ]
}`
