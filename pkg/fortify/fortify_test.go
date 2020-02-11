package fortify

import (
	"strings"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	ff "github.com/piper-validation/fortify-client-go/fortify"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
)

func TestGetProjectByName(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/projects" && req.URL.RawQuery == "fulltextsearch=true&q=name%3Dpython-test" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(
				`{"data": [{"_href": "https://fortify/ssc/api/v1/projects/4711","createdBy": "someUser","name": "python-test",
				"description": "","id": 4711,"creationDate": "2018-12-03T06:29:38.197+0000","issueTemplateId": "dasdasdasdsadasdasdasdasdas"}],
				"count": 1,"responseCode": 200,"links": {"last": {"href": "https://fortify/ssc/api/v1/projects?q=name%A3python-test&start=0"},
				"first": {"href": "https://fortify/ssc/api/v1/projects?q=name%A3python-test&start=0"}}}`))
			return
		}
		if req.URL.Path == "/projects/4711/versions" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(
				`{"data":[{"latestScanId":null,"serverVersion":17.2,"tracesOutOfDate":false,"attachmentsOutOfDate":false,"description":"",
				"project":{"id":4711,"name":"python-test","description":"","creationDate":"2018-12-03T06:29:38.197+0000","createdBy":"someUser",
				"issueTemplateId":"dasdasdasdsadasdasdasdasdas"},"sourceBasePath":null,"mode":"BASIC","masterAttrGuid":"sddasdasda","obfuscatedId":null,
				"id":10172,"customTagValuesAutoApply":null,"issueTemplateId":"dasdasdasdsadasdasdasdasdas","loadProperties":null,"predictionPolicy":null,
				"bugTrackerPluginId":null,"owner":"admin","_href":"https://fortify.mo.sap.corp/ssc/api/v1/projectVersions/10172",
				"committed":true,"bugTrackerEnabled":false,"active":true,"snapshotOutOfDate":false,"issueTemplateModifiedTime":1578411924701,
				"securityGroup":null,"creationDate":"2018-02-09T16:59:41.297+0000","refreshRequired":false,"issueTemplateName":"someTemplate",
				"migrationVersion":null,"createdBy":"admin","name":"0","siteId":null,"staleIssueTemplate":false,"autoPredict":null,
				"currentState":{"id":10172,"committed":true,"attentionRequired":false,"analysisResultsExist":true,"auditEnabled":true,
				"lastFprUploadDate":"2018-02-09T16:59:53.497+0000","extraMessage":null,"analysisUploadEnabled":true,"batchBugSubmissionExists":false,
				"hasCustomIssues":false,"metricEvaluationDate":"2018-03-10T00:02:45.553+0000","deltaPeriod":7,"issueCountDelta":0,"percentAuditedDelta":0.0,
				"criticalPriorityIssueCountDelta":0,"percentCriticalPriorityIssuesAuditedDelta":0.0},"assignedIssuesCount":0,"status":null}],
				"count":1,"responseCode":200,"links":{"last":{"href":"https://fortify.mo.sap.corp/ssc/api/v1/projects/4711/versions?start=0"},
				"first":{"href":"https://fortify.mo.sap.corp/ssc/api/v1/projects/4711/versions?start=0"}}}`))
			return
		}
	}))
	// Close the server when test finishes
	defer server.Close()

	parts := strings.Split(server.URL, "://")

	t.Run("test success", func(t *testing.T) {
		client := ff.NewHTTPClientWithConfig(strfmt.Default, &ff.TransportConfig{
			Host:     parts[1],
			Schemes:  []string{parts[0]},
			BasePath: ""},
		)
		sys := NewSystemInstanceForClient(client, "test2456", 60*time.Second)

		result, err := sys.GetProjectByName("python-test")
		assert.NoError(t, err, "GetProjectByName call not successful")
		assert.Equal(t, "python-test", strings.ToLower(*result.Name), "Expected to get python-test")

		result2, err := sys.GetProjectVersionDetailsByNameAndProjectID(result.ID, "0")
		assert.NoError(t, err, "GetProjectVersionDetailsByNameAndProjectID call not successful")
		assert.Equal(t, "0", *result2.Name, "Expected to get project version with different name")
	})
}
