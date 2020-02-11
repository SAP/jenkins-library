package fortify

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	ff "github.com/piper-validation/fortify-client-go/fortify"
	"github.com/piper-validation/fortify-client-go/models"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
)

func spinUpServer(f func(http.ResponseWriter, *http.Request)) (*SystemInstance, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(f))

	parts := strings.Split(server.URL, "://")
	client := ff.NewHTTPClientWithConfig(strfmt.Default, &ff.TransportConfig{
		Host:     parts[1],
		Schemes:  []string{parts[0]},
		BasePath: ""},
	)

	sys := NewSystemInstanceForClient(client, "test2456", 60*time.Second)
	return sys, server
}

func TestGetProjectByName(t *testing.T) {
	// Start a local HTTP server
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
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
		if req.URL.Path == "/projects" && req.URL.RawQuery == "fulltextsearch=true&q=name%3Dpython-empty" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(
				`{"data": [],"count": 0,"responseCode": 404,"links": {}}`))
			return
		}
		if req.URL.Path == "/projects" && req.URL.RawQuery == "fulltextsearch=true&q=name%3Dpython-error" {
			rw.WriteHeader(400)
			return
		}
	})
	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		result, err := sys.GetProjectByName("python-test")
		assert.NoError(t, err, "GetProjectByName call not successful")
		assert.Equal(t, "python-test", strings.ToLower(*result.Name), "Expected to get python-test")
	})

	t.Run("test empty", func(t *testing.T) {
		_, err := sys.GetProjectByName("python-empty")
		assert.Error(t, err, "Expected error but got success")
	})

	t.Run("test error", func(t *testing.T) {
		_, err := sys.GetProjectByName("python-error")
		assert.Error(t, err, "Expected error but got success")
	})
}

func TestGetProjectVersionDetailsByNameAndProjectID(t *testing.T) {
	// Start a local HTTP server
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
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
		if req.URL.Path == "/projects/777/versions" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(
				`{"data": [],"count": 0,"responseCode": 404,"links": {}}`))
			return
		}
		if req.URL.Path == "/projects/999/versions" {
			rw.WriteHeader(500)
			return
		}
	})

	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		result, err := sys.GetProjectVersionDetailsByNameAndProjectID(4711, "0")
		assert.NoError(t, err, "GetProjectVersionDetailsByNameAndProjectID call not successful")
		assert.Equal(t, "0", *result.Name, "Expected to get project version with different name")
	})

	t.Run("test empty", func(t *testing.T) {
		_, err := sys.GetProjectVersionDetailsByNameAndProjectID(777, "python-empty")
		assert.Error(t, err, "Expected error but got success")
	})

	t.Run("test HTTP error", func(t *testing.T) {
		_, err := sys.GetProjectVersionDetailsByNameAndProjectID(999, "python-http-error")
		assert.Error(t, err, "Expected error but got success")
	})
}

func TestGetProjectVersionAttributesByID(t *testing.T) {
	// Start a local HTTP server
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/projectVersions/4711/attributes" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(
				`{"data": [{"_href": "https://fortify.mo.sap.corp/ssc/api/v1/projectVersions/4711/attributes/4712","attributeDefinitionId": 31,
				"values": null,"guid": "gdgfdgfdgfdgfd","id": 4712,"value": "abcd"}],"count": 8,"responseCode": 200}`))
			return
		}
		if req.URL.Path == "/projectVersions/777/attributes" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(
				`{"data": [],"count": 0,"responseCode": 404,"links": {}}`))
			return
		}
		if req.URL.Path == "/projectVersions/999/attributes" {
			rw.WriteHeader(500)
			return
		}
	})

	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		result, err := sys.GetProjectVersionAttributesByID(4711)
		assert.NoError(t, err, "GetProjectVersionAttributesByID call not successful")
		assert.Equal(t, "abcd", *result[0].Value, "Expected to get attribute with different value")
		assert.Equal(t, int64(4712), result[0].ID, "Expected to get attribute with different id")
	})

	t.Run("test empty", func(t *testing.T) {
		result, err := sys.GetProjectVersionAttributesByID(777)
		assert.NoError(t, err, "GetProjectVersionAttributesByID call not successful")
		assert.Equal(t, 0, len(result), "Expected to not get any attributes")
	})

	t.Run("test HTTP error", func(t *testing.T) {
		_, err := sys.GetProjectVersionAttributesByID(999)
		assert.Error(t, err, "Expected error but got success")
	})
}

func TestCreateProjectVersion(t *testing.T) {
	// Start a local HTTP server
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/projectVersions" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			bodyBytes, _ := ioutil.ReadAll(req.Body)
			bodyContent := string(bodyBytes)
			responseContent := `{"data": `
			responseContent += bodyContent
			responseContent += `,"count": 1,"responseCode": 201,"links": {}}`
			fmt.Println(responseContent)
			rw.WriteHeader(201)
			rw.Write([]byte(responseContent))
			return
		}
	})

	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		int64Value := int64(65)
		int32Value := int32(876)
		float32Value := float32(19.12)
		now := models.NewIso8601MilliDateTime()
		enabled := true
		disabled := false
		name := "Test new PV"
		owner := "someUser"
		masterGUID := "dsadaoudoiud"
		project := models.Project{CreatedBy: &owner, CreationDate: now, Description: name, ID: int64Value, IssueTemplateID: &name, Name: &name}
		projectVersionState := models.ProjectVersionState{AnalysisResultsExist: &disabled, AnalysisUploadEnabled: &disabled,
			AttentionRequired: &disabled, AuditEnabled: &enabled, BatchBugSubmissionExists: &disabled, Committed: &enabled,
			CriticalPriorityIssueCountDelta: &int32Value, DeltaPeriod: &int32Value, ExtraMessage: &name, HasCustomIssues: &disabled,
			ID: &int64Value, IssueCountDelta: &int32Value, LastFprUploadDate: &now, MetricEvaluationDate: &now, PercentAuditedDelta: &float32Value,
			PercentCriticalPriorityIssuesAuditedDelta: &float32Value}
		version := models.ProjectVersion{AssignedIssuesCount: int64Value, Project: &project, Name: &name, Active: &enabled,
			Committed: &enabled, AttachmentsOutOfDate: disabled, AutoPredict: disabled, BugTrackerEnabled: &disabled,
			CustomTagValuesAutoApply: disabled, RefreshRequired: disabled, Owner: &owner, ServerVersion: &float32Value,
			SnapshotOutOfDate: &disabled, StaleIssueTemplate: &disabled, MasterAttrGUID: &masterGUID,
			LatestScanID: &int64Value, IssueTemplateName: &name, IssueTemplateModifiedTime: &int64Value,
			IssueTemplateID: &name, Description: &name, CreatedBy: &owner, BugTrackerPluginID: &name, Mode: "NONE",
			CurrentState: &projectVersionState, ID: int64Value, LoadProperties: "", CreationDate: &now,
			MigrationVersion: float32Value, ObfuscatedID: "", PredictionPolicy: "", SecurityGroup: "",
			SiteID: "", SourceBasePath: "", Status: "", TracesOutOfDate: false}
		result, err := sys.CreateProjectVersion(&version)
		assert.NoError(t, err, "CreateProjectVersion call not successful")
		assert.Equal(t, name, *result.Name, "Expected to get PV with different value")
		assert.Equal(t, int64(65), result.ID, "Expected to get PV with different id")
	})
}

func TestProjectVersionCopyFromPartial(t *testing.T) {
	// Start a local HTTP server
	bodyContent := ""
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/projectVersions/action/copyFromPartial" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			bodyBytes, _ := ioutil.ReadAll(req.Body)
			bodyContent = string(bodyBytes)
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
	})
	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		expected := `{"copyAnalysisProcessingRules":true,"copyBugTrackerConfiguration":true,"copyCurrentStateFpr":true,"copyCustomTags":true,"previousProjectVersionId":10172,"projectVersionId":10173}
`
		err := sys.ProjectVersionCopyFromPartial(10172, 10173)
		assert.NoError(t, err, "ProjectVersionCopyFromPartial call not successful")
		assert.Equal(t, bodyContent, expected, "Different request content expected")
	})
}

func TestProjectVersionCopyCurrentState(t *testing.T) {
	// Start a local HTTP server
	bodyContent := ""
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/projectVersions/action/copyCurrentState" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			bodyBytes, _ := ioutil.ReadAll(req.Body)
			bodyContent = string(bodyBytes)
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
	})
	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		expected := `{"copyCurrentStateFpr":true,"previousProjectVersionId":10172,"projectVersionId":10173}
`
		err := sys.ProjectVersionCopyCurrentState(10172, 10173)
		assert.NoError(t, err, "ProjectVersionCopyCurrentState call not successful")
		assert.Equal(t, bodyContent, expected, "Different request content expected")
	})
}
