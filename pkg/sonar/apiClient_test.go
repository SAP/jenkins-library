package sonar

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
)

func TestMalwareScanTests(t *testing.T) {

	t.Run("No malware, no encrypted content", func(t *testing.T) {
		url := "https://fake/api"
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		log.SetVerbose(true)

		// NewBasicAuthClient(mock.Anything, mock.Anything, mock.Anything, httpmock
		httpmock.RegisterRegexpResponder("GET", regexp.MustCompile(fmt.Sprintf("%s/.*", url)),
			// httpmock.RegisterResponder("GET", url+"/issues/search",
			httpmock.NewStringResponder(http.StatusOK, `{
    "total": 33,
    "p": 1,
    "ps": 1,
    "paging": {
        "pageIndex": 1,
        "pageSize": 1,
        "total": 34
    },
    "effortTotal": 39,
    "debtTotal": 39,
    "issues": [
        {
            "key": "AW7dg-l9Of9WKKBDmJJV",
            "rule": "javascript:S1131",
            "severity": "MINOR",
            "component": "Piper-Validation/NPM:karma.conf.js",
            "project": "Piper-Validation/NPM",
            "line": 53,
            "hash": "cea5bf6de7e4b3893bc76cc384a86592",
            "textRange": {
                "startLine": 53,
                "endLine": 53,
                "startOffset": 0,
                "endOffset": 40
            },
            "flows": [],
            "status": "OPEN",
            "message": "Remove the useless trailing whitespaces at the end of this line.",
            "effort": "1min",
            "debt": "1min",
            "assignee": "d065687",
            "author": "christopher.fenner@sap.com",
            "tags": [
                "convention"
            ],
            "creationDate": "2019-07-29T11:44:29+0000",
            "updateDate": "2019-12-06T23:20:33+0000",
            "type": "CODE_SMELL",
            "organization": "default-organization",
            "fromHotspot": false
        }
    ],
    "components": [
        {
            "organization": "default-organization",
            "key": "Piper-Validation/NPM",
            "uuid": "AW7dg-Y1v4pDRYwyZFuL",
            "enabled": true,
            "qualifier": "TRK",
            "name": "Piper-Validation: NPM",
            "longName": "Piper-Validation: NPM"
        },
        {
            "organization": "default-organization",
            "key": "Piper-Validation/NPM:karma.conf.js",
            "uuid": "AW7dg-isOf9WKKBDmJIu",
            "enabled": true,
            "qualifier": "FIL",
            "name": "karma.conf.js",
            "longName": "karma.conf.js",
            "path": "karma.conf.js"
        }
    ],
    "facets": []
}`))
		// httpmock.NewStringResponder(200, `[{"id": 1, "name": "My Great Article"}]`))

		httpmock.GetTotalCallCount()
		sender := &piperhttp.Client{}
		sender.SetOptions(piperhttp.ClientOptions{TransportSkipVerification: true})

		issues := IssueService{
			Host:       url,
			Token:      "dcf72846f813f84648b5255beda2599d0a38b4ab",
			Project:    "Piper-Validation/NPM",
			HTTPClient: sender,
		}
		result, err := issues.GetNumberOfMinorIssues()

		httpmock.GetTotalCallCount()
		info := httpmock.GetCallCountInfo()
		assert.Equal(t, "", info)
		assert.NoError(t, err)
		assert.Equal(t, 33, result)
		// info["GET https://api.mybiz.com/articles"]
		// httpClient := httpmock{StatusCode: 200, ResponseBody: "{\"malwareDetected\":false,\"encryptedContentDetected\":false,\"scanSize\":298782,\"mimeType\":\"application/octet-stream\",\"SHA256\":\"96ca802fbd54d31903f1115a1d95590c685160637d9262bd340ab30d0f817e85\"}"}

		// error := runMalwareScan(&malwareScanConfig, nil, nil, &httpClient)

		// if assert.NoError(t, error) {

		// 	t.Run("check url", func(t *testing.T) {
		// 		assert.Equal(t, "https://example.org/malwarescanner/scan", httpClient.URL)
		// 	})

		// 	t.Run("check method", func(t *testing.T) {
		// 		assert.Equal(t, "POST", httpClient.Method)
		// 	})

		// 	t.Run("check user", func(t *testing.T) {
		// 		assert.Equal(t, "me", httpClient.Options.Username)
		// 	})

		// 	t.Run("check password", func(t *testing.T) {
		// 		assert.Equal(t, "********", httpClient.Options.Password)
		// 	})
		// }
	})
}
