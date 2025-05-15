//go:build unit
// +build unit

package fortify

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	ff "github.com/piper-validation/fortify-client-go/fortify"
	"github.com/piper-validation/fortify-client-go/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
)

func spinUpServer(f func(http.ResponseWriter, *http.Request)) (*SystemInstance, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(f))

	parts := strings.Split(server.URL, "://")
	client := ff.NewHTTPClientWithConfig(strfmt.Default, &ff.TransportConfig{
		Host:     parts[1],
		Schemes:  []string{parts[0]},
		BasePath: ""},
	)

	httpClient := &piperHttp.Client{}
	httpClientOptions := piperHttp.ClientOptions{Token: "test2456", TransportTimeout: 60 * time.Second}
	httpClient.SetOptions(httpClientOptions)

	sys := NewSystemInstanceForClient(client, httpClient, server.URL, "test2456", 60*time.Second)
	return sys, server
}

func TestCreateTransportConfig(t *testing.T) {
	t.Run("Valid URL", func(t *testing.T) {
		config := createTransportConfig("http://some.fortify.host.com/ssc", "/api/v2")
		assert.Equal(t, []string{"http"}, config.Schemes)
		assert.Equal(t, "some.fortify.host.com", config.Host)
		assert.Equal(t, "ssc/api/v2", config.BasePath)
	})
	t.Run("Slashes are trimmed", func(t *testing.T) {
		config := createTransportConfig("http://some.fortify.host.com/ssc//", "//api/v2/")
		assert.Equal(t, []string{"http"}, config.Schemes)
		assert.Equal(t, "some.fortify.host.com", config.Host)
		assert.Equal(t, "ssc/api/v2", config.BasePath)
	})
	t.Run("URL missing scheme results in no error", func(t *testing.T) {
		config := createTransportConfig("some.fortify.host.com/ssc", "api/v1")
		assert.Equal(t, []string{"https"}, config.Schemes)
		assert.Equal(t, "some.fortify.host.com", config.Host)
		assert.Equal(t, "ssc/api/v1", config.BasePath)
	})
	t.Run("URL with more than one slash is accepted", func(t *testing.T) {
		config := createTransportConfig("https://some.fortify.host.com/some/path/ssc", "api/v1")
		assert.Equal(t, []string{"https"}, config.Schemes)
		assert.Equal(t, "some.fortify.host.com", config.Host)
		assert.Equal(t, "some/path/ssc/api/v1", config.BasePath)
	})
}

func TestNewSystemInstance(t *testing.T) {
	t.Run("fields are initialized", func(t *testing.T) {
		sys := NewSystemInstance("https://some.fortify.host.com/ssc", "api/v1", "akjhskjhks", "", 10*time.Second)
		assert.IsType(t, ff.Fortify{}, *sys.client, "Expected to get a Fortify client instance")
		assert.IsType(t, piperHttp.Client{}, *sys.httpClient, "Expected to get a HTTP client instance")
		assert.IsType(t, logrus.Entry{}, *sys.logger, "Expected to get a logrus entry instance")
		assert.Equal(t, 10*time.Second, sys.timeout, "Expected different timeout value")
		assert.Equal(t, "akjhskjhks", sys.token, "Expected different token value")
		assert.Equal(t, "https://some.fortify.host.com/ssc", sys.serverURL)
	})
	t.Run("SSC URL is trimmed", func(t *testing.T) {
		sys := NewSystemInstance("https://some.fortify.host.com/ssc/", "api/v1", "akjhskjhks", "", 10*time.Second)
		assert.Equal(t, "https://some.fortify.host.com/ssc", sys.serverURL)
	})
}

func TestGetProjectByName(t *testing.T) {
	// Start a local HTTP server
	autocreateCalled := false
	commitCalled := false
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/projects" && req.URL.RawQuery == "q=name%3A%22python-test%22" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(
				`{"data": [{"_href": "https://fortify/ssc/api/v1/projects/4711","createdBy": "someUser","name": "python-test",
				"description": "","id": 4711,"creationDate": "2018-12-03T06:29:38.197+0000","issueTemplateId": "dasdasdasdsadasdasdasdasdas"}],
				"count": 1,"responseCode": 200,"links": {"last": {"href": "https://fortify/ssc/api/v1/projects?q=name%A3python-test&start=0"},
				"first": {"href": "https://fortify/ssc/api/v1/projects?q=name%3Apython-test&start=0"}}}`))
			return
		}
		if req.URL.Path == "/projects" && req.URL.RawQuery == "q=name%3A%22python+with+space+test%22" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(
				`{"data": [{"_href": "https://fortify/ssc/api/v1/projects/4711","createdBy": "someUser","name": "python with space test",
				"description": "","id": 4711,"creationDate": "2018-12-03T06:29:38.197+0000","issueTemplateId": "dasdasdasdsadasdasdasdasdas"}],
				"count": 1,"responseCode": 200,"links": {"last": {"href": "https://fortify/ssc/api/v1/projects?q=name%3A%22python+with+space+test%22&start=0"},
				"first": {"href": "https://fortify/ssc/api/v1/projects?q=name%3A%22python+with+space+test%22&start=0"}}}`))
			return
		}
		if req.URL.Path == "/projects" && req.URL.RawQuery == "q=name%3A%22python-empty%22" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(
				`{"data": [],"count": 0,"responseCode": 404,"links": {}}`))
			return
		}
		if req.URL.Path == "/projects" && req.URL.RawQuery == "q=name%3A%22python-error%22" {
			rw.WriteHeader(400)
			return
		}
		if req.URL.Path == "/projectVersions" && req.Method == "POST" {
			autocreateCalled = true
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.WriteHeader(201)
			rw.Write([]byte(
				`{"data":{"latestScanId":null,"serverVersion":17.2,"tracesOutOfDate":false,"attachmentsOutOfDate":false,"description":"",
				"project":{"id":815,"name":"autocreate","description":"","creationDate":"2018-12-03T06:29:38.197+0000","createdBy":"someUser",
				"issueTemplateId":"dasdasdasdsadasdasdasdasdas"},"sourceBasePath":null,"mode":"BASIC","masterAttrGuid":"sddasdasda","obfuscatedId":null,
				"id":10172,"customTagValuesAutoApply":null,"issueTemplateId":"dasdasdasdsadasdasdasdasdas","loadProperties":null,"predictionPolicy":null,
				"bugTrackerPluginId":null,"owner":"admin","_href":"https://fortify/ssc/api/v1/projectVersions/10172",
				"committed":true,"bugTrackerEnabled":false,"active":true,"snapshotOutOfDate":false,"issueTemplateModifiedTime":1578411924701,
				"securityGroup":null,"creationDate":"2018-02-09T16:59:41.297+0000","refreshRequired":false,"issueTemplateName":"someTemplate",
				"migrationVersion":null,"createdBy":"admin","name":"0","siteId":null,"staleIssueTemplate":false,"autoPredict":null,
				"currentState":{"id":10172,"committed":true,"attentionRequired":false,"analysisResultsExist":true,"auditEnabled":true,
				"lastFprUploadDate":"2018-02-09T16:59:53.497+0000","extraMessage":null,"analysisUploadEnabled":true,"batchBugSubmissionExists":false,
				"hasCustomIssues":false,"metricEvaluationDate":"2018-03-10T00:02:45.553+0000","deltaPeriod":7,"issueCountDelta":0,"percentAuditedDelta":0.0,
				"criticalPriorityIssueCountDelta":0,"percentCriticalPriorityIssuesAuditedDelta":0.0},"assignedIssuesCount":0,"status":null},
				"count":1,"responseCode":200,"links":{"last":{"href":"https://fortify/ssc/api/v1/projects/815/versions?start=0"},
				"first":{"href":"https://fortify/ssc/api/v1/projects/815/versions?start=0"}}}`))
			return
		}
		if req.URL.Path == "/projectVersions/10172" {
			commitCalled = true
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(
				`{"data": {"_href": "https://fortify/ssc/api/v1/projects/815", "committed": true,"createdBy": "someUser","name": "autocreate",
				"description": "","id": 815,"creationDate": "2018-12-03T06:29:38.197+0000","issueTemplateId": "dasdasdasdsadasdasdasdasdas"},
				"count": 1,"responseCode": 200,"links": {"last": {"href": "https://fortify/ssc/api/v1/projects?q=name%3Apython-test&start=0"},
				"first": {"href": ""}}}`))
			return
		}
	})
	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		result, err := sys.GetProjectByName("python-test", false, "")
		assert.NoError(t, err, "GetProjectByName call not successful")
		assert.Equal(t, "python-test", strings.ToLower(*result.Name), "Expected to get python-test")
	})

	t.Run("test space", func(t *testing.T) {
		result, err := sys.GetProjectByName("python with space test", false, "")
		assert.NoError(t, err, "GetProjectByName call not successful")
		assert.Equal(t, "python with space test", strings.ToLower(*result.Name), "Expected to get python with space test")
	})

	t.Run("test empty", func(t *testing.T) {
		_, err := sys.GetProjectByName("python-empty", false, "")
		assert.Error(t, err, "Expected error but got success")
	})

	t.Run("test error", func(t *testing.T) {
		_, err := sys.GetProjectByName("python-error", false, "")
		assert.Error(t, err, "Expected error but got success")
	})

	t.Run("test auto create success", func(t *testing.T) {
		result, err := sys.GetProjectByName("autocreate", true, "123456")
		assert.NoError(t, err, "GetProjectByName call not successful")
		assert.Equal(t, true, autocreateCalled, "Expected autocreation function to be called but wasn't")
		assert.Equal(t, true, commitCalled, "Expected commit function to be called but wasn't")
		assert.Equal(t, "autocreate", strings.ToLower(*result.Name), "Expected to get autocreate project")
	})
}

func TestGetProjectVersionDetailsByProjectIDAndVersionName(t *testing.T) {
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
				"bugTrackerPluginId":null,"owner":"admin","_href":"https://fortify/ssc/api/v1/projectVersions/10172",
				"committed":true,"bugTrackerEnabled":false,"active":true,"snapshotOutOfDate":false,"issueTemplateModifiedTime":1578411924701,
				"securityGroup":null,"creationDate":"2018-02-09T16:59:41.297+0000","refreshRequired":false,"issueTemplateName":"someTemplate",
				"migrationVersion":null,"createdBy":"admin","name":"0","siteId":null,"staleIssueTemplate":false,"autoPredict":null,
				"currentState":{"id":10172,"committed":true,"attentionRequired":false,"analysisResultsExist":true,"auditEnabled":true,
				"lastFprUploadDate":"2018-02-09T16:59:53.497+0000","extraMessage":null,"analysisUploadEnabled":true,"batchBugSubmissionExists":false,
				"hasCustomIssues":false,"metricEvaluationDate":"2018-03-10T00:02:45.553+0000","deltaPeriod":7,"issueCountDelta":0,"percentAuditedDelta":0.0,
				"criticalPriorityIssueCountDelta":0,"percentCriticalPriorityIssuesAuditedDelta":0.0},"assignedIssuesCount":0,"status":null}],
				"count":1,"responseCode":200,"links":{"last":{"href":"https://fortify/ssc/api/v1/projects/4711/versions?start=0"},
				"first":{"href":"https://fortify/ssc/api/v1/projects/4711/versions?start=0"}}}`))
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
		if req.URL.Path == "/projectVersions" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.WriteHeader(201)
			rw.Write([]byte(
				`{"data":{"latestScanId":null,"serverVersion":17.2,"tracesOutOfDate":false,"attachmentsOutOfDate":false,"description":"",
				"project":{"id":815,"name":"autocreate","description":"","creationDate":"2018-12-03T06:29:38.197+0000","createdBy":"someUser",
				"issueTemplateId":"dasdasdasdsadasdasdasdasdas"},"sourceBasePath":null,"mode":"BASIC","masterAttrGuid":"sddasdasda","obfuscatedId":null,
				"id":10172,"customTagValuesAutoApply":null,"issueTemplateId":"dasdasdasdsadasdasdasdasdas","loadProperties":null,"predictionPolicy":null,
				"bugTrackerPluginId":null,"owner":"admin","_href":"https://fortify/ssc/api/v1/projectVersions/10172",
				"committed":true,"bugTrackerEnabled":false,"active":true,"snapshotOutOfDate":false,"issueTemplateModifiedTime":1578411924701,
				"securityGroup":null,"creationDate":"2018-02-09T16:59:41.297+0000","refreshRequired":false,"issueTemplateName":"someTemplate",
				"migrationVersion":null,"createdBy":"admin","name":"0","siteId":null,"staleIssueTemplate":false,"autoPredict":null,
				"currentState":{"id":10172,"committed":true,"attentionRequired":false,"analysisResultsExist":true,"auditEnabled":true,
				"lastFprUploadDate":"2018-02-09T16:59:53.497+0000","extraMessage":null,"analysisUploadEnabled":true,"batchBugSubmissionExists":false,
				"hasCustomIssues":false,"metricEvaluationDate":"2018-03-10T00:02:45.553+0000","deltaPeriod":7,"issueCountDelta":0,"percentAuditedDelta":0.0,
				"criticalPriorityIssueCountDelta":0,"percentCriticalPriorityIssuesAuditedDelta":0.0},"assignedIssuesCount":0,"status":null},
				"count":1,"responseCode":200,"links":{"last":{"href":"https://fortify/ssc/api/v1/projects/815/versions?start=0"},
				"first":{"href":"https://fortify/ssc/api/v1/projects/815/versions?start=0"}}}`))
			return
		}
		if req.URL.Path == "/projectVersions/0" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(
				`{"data":{"latestScanId":null,"serverVersion":17.2,"tracesOutOfDate":false,"attachmentsOutOfDate":false,"description":"",
				"project":{"id":815,"name":"autocreate","description":"","creationDate":"2018-12-03T06:29:38.197+0000","createdBy":"someUser",
				"issueTemplateId":"dasdasdasdsadasdasdasdasdas"},"sourceBasePath":null,"mode":"BASIC","masterAttrGuid":"sddasdasda","obfuscatedId":null,
				"id":10172,"customTagValuesAutoApply":null,"issueTemplateId":"dasdasdasdsadasdasdasdasdas","loadProperties":null,"predictionPolicy":null,
				"bugTrackerPluginId":null,"owner":"admin","_href":"https://fortify/ssc/api/v1/projectVersions/10172",
				"committed":true,"bugTrackerEnabled":false,"active":true,"snapshotOutOfDate":false,"issueTemplateModifiedTime":1578411924701,
				"securityGroup":null,"creationDate":"2018-02-09T16:59:41.297+0000","refreshRequired":false,"issueTemplateName":"someTemplate",
				"migrationVersion":null,"createdBy":"admin","name":"0","siteId":null,"staleIssueTemplate":false,"autoPredict":null,
				"currentState":{"id":10172,"committed":true,"attentionRequired":false,"analysisResultsExist":true,"auditEnabled":true,
				"lastFprUploadDate":"2018-02-09T16:59:53.497+0000","extraMessage":null,"analysisUploadEnabled":true,"batchBugSubmissionExists":false,
				"hasCustomIssues":false,"metricEvaluationDate":"2018-03-10T00:02:45.553+0000","deltaPeriod":7,"issueCountDelta":0,"percentAuditedDelta":0.0,
				"criticalPriorityIssueCountDelta":0,"percentCriticalPriorityIssuesAuditedDelta":0.0},"assignedIssuesCount":0,"status":null},
				"count":1,"responseCode":200,"links":{"last":{"href":"https://fortify/ssc/api/v1/projects/815/versions?start=0"},
				"first":{"href":"https://fortify/ssc/api/v1/projects/815/versions?start=0"}}}`))
			return
		}
		if req.URL.Path == "/projects/8888/versions" && req.URL.RawQuery == "q=name%3A%221%22" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(`{"data":[{"id":9910,"project":{"id":8888,"name":"test","description":"Created by Go script","creationDate":"2022-06-24T04:44:12.344+0000",
				"createdBy":"jajajajja","issueTemplateId":"asxasca-asff-b57aedaf41"},"name":"1","description":"","createdBy":"afsafa","creationDate":"2021-07-17T04:09:17.909+0000",
				"sourceBasePath":null,"committed":true,"issueTemplateId":"asdawffbcad88eb041","issueTemplateName":"aiudfnwofn","loadProperties":null,"staleIssueTemplate":false,
				"snapshotOutOfDate":false,"refreshRequired":false,"attachmentsOutOfDate":false,"migrationVersion":null,"masterAttrGuid":"akjnfkjsnfkj686b","tracesOutOfDate":false,
				"issueTemplateModifiedTime":1556502937909,"active":true,"obfuscatedId":null,"owner":"","serverVersion":21.2,"siteId":null,"latestScanId":null,"mode":"BASIC",
				"currentState":{"id":9910,"committed":true,"attentionRequired":false,"analysisResultsExist":false,"auditEnabled":false,"lastFprUploadDate":null,"extraMessage":null,
				"analysisUploadEnabled":true,"batchBugSubmissionExists":false,"hasCustomIssues":false,"metricEvaluationDate":null,"deltaPeriod":7,"issueCountDelta":0,"percentAuditedDelta":0,
				"criticalPriorityIssueCountDelta":0,"percentCriticalPriorityIssuesAuditedDelta":0},"bugTrackerPluginId":null,"bugTrackerEnabled":false,"securityGroup":null,"status":null,
				"assignedIssuesCount":0,"customTagValuesAutoApply":null,"autoPredict":null,"predictionPolicy":null,"_href":"https://fortify/ssc/api/v1/projectVersions/8888"}],"count":1,
				"responseCode":200,"links":{"last":{"href":"https://fortify/ssc/api/v1/projects/8888/versions?q=name%3A1&start=0"},"first":{"href":"https://fortify/ssc/api/v1/projects/8888/versions?q=name%3A1&start=0"}}}`))
			return
		}
	})

	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		result, err := sys.GetProjectVersionDetailsByProjectIDAndVersionName(4711, "0", false, "")
		assert.NoError(t, err, "GetProjectVersionDetailsByNameAndProjectID call not successful")
		assert.Equal(t, "0", *result.Name, "Expected to get project version with different name")
	})

	t.Run("test empty", func(t *testing.T) {
		_, err := sys.GetProjectVersionDetailsByProjectIDAndVersionName(777, "python-empty", false, "")
		assert.Error(t, err, "Expected error but got success")
	})

	t.Run("test HTTP error", func(t *testing.T) {
		_, err := sys.GetProjectVersionDetailsByProjectIDAndVersionName(999, "python-http-error", false, "")
		assert.Error(t, err, "Expected error but got success")
	})

	t.Run("test auto create success", func(t *testing.T) {
		result, err := sys.GetProjectVersionDetailsByProjectIDAndVersionName(815, "0", true, "autocreate")
		assert.NoError(t, err, "GetProjectVersionDetailsByNameAndProjectID call not successful")
		assert.Equal(t, "0", *result.Name, "Expected to get project version with different name")
		assert.Equal(t, "autocreate", *result.Project.Name, "Expected to get project with different name")
	})
	t.Run("test filter projectVersion", func(t *testing.T) {
		result, err := sys.GetProjectVersionDetailsByProjectIDAndVersionName(8888, "1", true, "autocreate")
		assert.NoError(t, err, "GetProjectVersionDetailsByNameAndProjectID call not successful")
		assert.Equal(t, "1", *result.Name, "Expected to get exact project version")
	})
}

func TestGetProjectVersionAttributesByProjectVersionID(t *testing.T) {
	// Start a local HTTP server
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/projectVersions/4711/attributes" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(
				`{"data": [{"_href": "https://fortify/ssc/api/v1/projectVersions/4711/attributes/4712","attributeDefinitionId": 31,
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
		result, err := sys.GetProjectVersionAttributesByProjectVersionID(4711)
		assert.NoError(t, err, "GetProjectVersionAttributesByProjectVersionID call not successful")
		assert.Equal(t, "abcd", *result[0].Value, "Expected to get attribute with different value")
		assert.Equal(t, int64(4712), result[0].ID, "Expected to get attribute with different id")
	})

	t.Run("test empty", func(t *testing.T) {
		result, err := sys.GetProjectVersionAttributesByProjectVersionID(777)
		assert.NoError(t, err, "GetProjectVersionAttributesByID call not successful")
		assert.Equal(t, 0, len(result), "Expected to not get any attributes")
	})

	t.Run("test HTTP error", func(t *testing.T) {
		_, err := sys.GetProjectVersionAttributesByProjectVersionID(999)
		assert.Error(t, err, "Expected error but got success")
	})
}

func TestSetProjectVersionAttributesByProjectVersionID(t *testing.T) {
	// Start a local HTTP server
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/projectVersions/4711/attributes" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			bodyBytes, _ := io.ReadAll(req.Body)
			bodyString := string(bodyBytes)
			response := `{"data": `
			response += bodyString
			response += `,"count": 1,"responseCode": 200}`
			rw.WriteHeader(200)
			rw.Write([]byte(response))
			return
		}
	})

	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		value := "abcd"
		defID := int64(18)
		attributes := []*models.Attribute{{ID: 4712, Value: &value, AttributeDefinitionID: &defID}}
		result, err := sys.SetProjectVersionAttributesByProjectVersionID(4711, attributes)
		assert.NoError(t, err, "SetProjectVersionAttributesByProjectVersionID call not successful")
		assert.Equal(t, 1, len(result), "Expected to get slice with different amount of values")
		assert.Equal(t, "abcd", *result[0].Value, "Expected to get attribute with different value")
		assert.Equal(t, int64(4712), result[0].ID, "Expected to get attribute with different id")
	})
}

func TestCreateProjectVersion(t *testing.T) {
	// Start a local HTTP server
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/projectVersions" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			bodyBytes, _ := io.ReadAll(req.Body)
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
			bodyBytes, _ := io.ReadAll(req.Body)
			bodyContent = string(bodyBytes)
			rw.Write([]byte(
				`{"data":[{"latestScanId":null,"serverVersion":17.2,"tracesOutOfDate":false,"attachmentsOutOfDate":false,"description":"",
				"project":{"id":4711,"name":"python-test","description":"","creationDate":"2018-12-03T06:29:38.197+0000","createdBy":"someUser",
				"issueTemplateId":"dasdasdasdsadasdasdasdasdas"},"sourceBasePath":null,"mode":"BASIC","masterAttrGuid":"sddasdasda","obfuscatedId":null,
				"id":10172,"customTagValuesAutoApply":null,"issueTemplateId":"dasdasdasdsadasdasdasdasdas","loadProperties":null,"predictionPolicy":null,
				"bugTrackerPluginId":null,"owner":"admin","_href":"https://fortify/ssc/api/v1/projectVersions/10172",
				"committed":true,"bugTrackerEnabled":false,"active":true,"snapshotOutOfDate":false,"issueTemplateModifiedTime":1578411924701,
				"securityGroup":null,"creationDate":"2018-02-09T16:59:41.297+0000","refreshRequired":false,"issueTemplateName":"someTemplate",
				"migrationVersion":null,"createdBy":"admin","name":"0","siteId":null,"staleIssueTemplate":false,"autoPredict":null,
				"currentState":{"id":10172,"committed":true,"attentionRequired":false,"analysisResultsExist":true,"auditEnabled":true,
				"lastFprUploadDate":"2018-02-09T16:59:53.497+0000","extraMessage":null,"analysisUploadEnabled":true,"batchBugSubmissionExists":false,
				"hasCustomIssues":false,"metricEvaluationDate":"2018-03-10T00:02:45.553+0000","deltaPeriod":7,"issueCountDelta":0,"percentAuditedDelta":0.0,
				"criticalPriorityIssueCountDelta":0,"percentCriticalPriorityIssuesAuditedDelta":0.0},"assignedIssuesCount":0,"status":null}],
				"count":1,"responseCode":200,"links":{"last":{"href":"https://fortify/ssc/api/v1/projects/4711/versions?start=0"},
				"first":{"href":"https://fortify/ssc/api/v1/projects/4711/versions?start=0"}}}`))
			return
		}
	})
	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		expected := `{"copyAnalysisProcessingRules":true,"copyBugTrackerConfiguration":true,"copyCustomTags":true,"previousProjectVersionId":10172,"projectVersionId":10173}
`
		err := sys.ProjectVersionCopyFromPartial(10172, 10173)
		assert.NoError(t, err, "ProjectVersionCopyFromPartial call not successful")
		assert.Equal(t, expected, bodyContent, "Different request content expected")
	})
}

func TestProjectVersionCopyCurrentState(t *testing.T) {
	// Start a local HTTP server
	bodyContent := ""
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/projectVersions/action/copyCurrentState" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			bodyBytes, _ := io.ReadAll(req.Body)
			bodyContent = string(bodyBytes)
			rw.Write([]byte(
				`{"data":[{"latestScanId":null,"serverVersion":17.2,"tracesOutOfDate":false,"attachmentsOutOfDate":false,"description":"",
				"project":{"id":4711,"name":"python-test","description":"","creationDate":"2018-12-03T06:29:38.197+0000","createdBy":"someUser",
				"issueTemplateId":"dasdasdasdsadasdasdasdasdas"},"sourceBasePath":null,"mode":"BASIC","masterAttrGuid":"sddasdasda","obfuscatedId":null,
				"id":10172,"customTagValuesAutoApply":null,"issueTemplateId":"dasdasdasdsadasdasdasdasdas","loadProperties":null,"predictionPolicy":null,
				"bugTrackerPluginId":null,"owner":"admin","_href":"https://fortify/ssc/api/v1/projectVersions/10172",
				"committed":true,"bugTrackerEnabled":false,"active":true,"snapshotOutOfDate":false,"issueTemplateModifiedTime":1578411924701,
				"securityGroup":null,"creationDate":"2018-02-09T16:59:41.297+0000","refreshRequired":false,"issueTemplateName":"someTemplate",
				"migrationVersion":null,"createdBy":"admin","name":"0","siteId":null,"staleIssueTemplate":false,"autoPredict":null,
				"currentState":{"id":10172,"committed":true,"attentionRequired":false,"analysisResultsExist":true,"auditEnabled":true,
				"lastFprUploadDate":"2018-02-09T16:59:53.497+0000","extraMessage":null,"analysisUploadEnabled":true,"batchBugSubmissionExists":false,
				"hasCustomIssues":false,"metricEvaluationDate":"2018-03-10T00:02:45.553+0000","deltaPeriod":7,"issueCountDelta":0,"percentAuditedDelta":0.0,
				"criticalPriorityIssueCountDelta":0,"percentCriticalPriorityIssuesAuditedDelta":0.0},"assignedIssuesCount":0,"status":null}],
				"count":1,"responseCode":200,"links":{"last":{"href":"https://fortify/ssc/api/v1/projects/4711/versions?start=0"},
				"first":{"href":"https://fortify/ssc/api/v1/projects/4711/versions?start=0"}}}`))
			return
		}
	})
	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		expected := `{"previousProjectVersionId":10172,"projectVersionId":10173}
`
		err := sys.ProjectVersionCopyCurrentState(10172, 10173)
		assert.NoError(t, err, "ProjectVersionCopyCurrentState call not successful")
		assert.Equal(t, expected, bodyContent, "Different request content expected")
	})
}

func TestProjectVersionCopyPermissions(t *testing.T) {
	// Start a local HTTP server
	bodyContent := ""
	referenceContent := `[{"displayName":"some user","email":"some.one@test.com","entityName":"some_user","firstName":"some","id":589,"lastName":"user","type":"User"}]
`
	response := `{"data": `
	response += referenceContent
	response += `,"count": 1,"responseCode": 200}`
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/projectVersions/10172/authEntities" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(response))
			return
		}
		if req.URL.Path == "/projectVersions/10173/authEntities" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			bodyBytes, _ := io.ReadAll(req.Body)
			bodyContent = string(bodyBytes)
			rw.Write([]byte(response))
			return
		}
	})
	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		err := sys.ProjectVersionCopyPermissions(10172, 10173)
		assert.NoError(t, err, "ProjectVersionCopyPermissions call not successful")
		assert.Equal(t, referenceContent, bodyContent, "Different request content expected")
	})
}

func TestCommitProjectVersion(t *testing.T) {
	// Start a local HTTP server
	bodyContent := ""
	referenceContent := `{"active":null,"bugTrackerEnabled":null,"bugTrackerPluginId":null,"committed":true,"createdBy":null,"creationDate":null,"description":null,"issueTemplateId":null,"issueTemplateModifiedTime":null,"issueTemplateName":null,"latestScanId":null,"masterAttrGuid":null,"name":null,"owner":null,"serverVersion":null,"snapshotOutOfDate":null,"staleIssueTemplate":null}
`
	response := `{"data": `
	response += referenceContent
	response += `,"count": 1,"responseCode": 200}`
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/projectVersions/10172" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			bodyBytes, _ := io.ReadAll(req.Body)
			bodyContent = string(bodyBytes)
			rw.Write([]byte(response))
			return
		}
	})
	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		result, err := sys.CommitProjectVersion(10172)
		assert.NoError(t, err, "CommitProjectVersion call not successful")
		assert.Equal(t, true, *result.Committed, "Different result content expected")
		assert.Equal(t, referenceContent, bodyContent, "Different request content expected")
	})
}

func TestInactivateProjectVersion(t *testing.T) {
	// Start a local HTTP server
	bodyContent := ""
	referenceContent := `{"active":false,"bugTrackerEnabled":null,"bugTrackerPluginId":null,"committed":true,"createdBy":null,"creationDate":null,"description":null,"issueTemplateId":null,"issueTemplateModifiedTime":null,"issueTemplateName":null,"latestScanId":null,"masterAttrGuid":null,"name":null,"owner":null,"serverVersion":null,"snapshotOutOfDate":null,"staleIssueTemplate":null}
`
	response := `{"data": `
	response += referenceContent
	response += `,"count": 1,"responseCode": 200}`
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/projectVersions/10172" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			bodyBytes, _ := io.ReadAll(req.Body)
			bodyContent = string(bodyBytes)
			rw.Write([]byte(response))
			return
		}
	})
	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		result, err := sys.inactivateProjectVersion(10172)
		assert.NoError(t, err, "InactivateProjectVersion call not successful")
		assert.Equal(t, true, *result.Committed, "Different result content expected")
		assert.Equal(t, false, *result.Active, "Different result content expected")
		assert.Equal(t, referenceContent, bodyContent, "Different request content expected")
	})
}

func TestGetArtifactsOfProjectVersion(t *testing.T) {
	// Start a local HTTP server
	response := `{"data": [{"artifactType": "FPR","fileName": "df54e2ade34c4f6aaddf35679dd87a21.tmp","approvalDate": null,"messageCount": 0,
		"scanErrorsCount": 0,"uploadIP": "10.238.8.48","allowApprove": false,"allowPurge": false,"lastScanDate": "2019-11-26T22:37:52.000+0000",
		"fileURL": null,"id": 56,"purged": false,"webInspectStatus": "NONE","inModifyingStatus": false,"originalFileName": "result.fpr",
		"allowDelete": true,"scaStatus": "PROCESSED","indexed": true,"runtimeStatus": "NONE","userName": "some_user","versionNumber": null,
		"otherStatus": "NOT_EXIST","uploadDate": "2019-11-26T22:38:11.813+0000","approvalComment": null,"approvalUsername": null,"fileSize": 984703,
		"messages": "","auditUpdated": false,"status": "PROCESS_COMPLETE"}],"count": 1,"responseCode": 200}`
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/projectVersions/10172/artifacts" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(response))
			return
		}
	})
	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		result, err := sys.GetArtifactsOfProjectVersion(10172)
		assert.NoError(t, err, "GetArtifactsOfProjectVersion call not successful")
		assert.Equal(t, 1, len(result), "Different result content expected")
		assert.Equal(t, int64(56), result[0].ID, "Different result content expected")
	})
}

func TestGetFilterSetOfProjectVersionByTitle(t *testing.T) {
	// Start a local HTTP server
	response := `{"data":[{"defaultFilterSet":true,"folders":[
	{"id":1,"guid":"4711","name":"Corporate Security Requirements","color":"000000"},
	{"id":2,"guid":"4712","name":"Audit All","color":"ff0000"},
	{"id":3,"guid":"4713","name":"Spot Checks of Each Category","color":"ff8000"},
	{"id":4,"guid":"4714","name":"Optional","color":"808080"}],"description":"",
	"guid":"666","title":"Special"}],"count":1,"responseCode":200}}`
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/projectVersions/10172/filterSets" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(response))
			return
		}
	})
	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		result, err := sys.GetFilterSetOfProjectVersionByTitle(10172, "Special")
		assert.NoError(t, err, "GetFilterSetOfProjectVersionByTitle call not successful")
		assert.Equal(t, "Special", result.Title, "Different result content expected")
	})

	t.Run("test default", func(t *testing.T) {
		result, err := sys.GetFilterSetOfProjectVersionByTitle(10172, "")
		assert.NoError(t, err, "GetFilterSetOfProjectVersionByTitle call not successful")
		assert.Equal(t, "Special", result.Title, "Different result content expected")
	})
}

func TestGetIssueFilterSelectorOfProjectVersionByName(t *testing.T) {
	// Start a local HTTP server
	response := `{"data":{"groupBySet": [{"entityType": "CUSTOMTAG","guid": "adsffghjkl","displayName": "Analysis",
	"value": "87f2364f-dcd4-49e6-861d-f8d3f351686b","description": ""},{"entityType": "ISSUE","guid": "lkjhgfd",
	"displayName": "Category","value": "11111111-1111-1111-1111-111111111165","description": ""}],"filterBySet":[{
	"entityType": "CUSTOMTAG","filterSelectorType": "LIST","guid": "87f2364f-dcd4-49e6-861d-f8d3f351686b","displayName": "Analysis",
	"value": "87f2364f-dcd4-49e6-861d-f8d3f351686b","description": "The analysis tag must be set.",
	"selectorOptions": []},{"entityType": "FOLDER","filterSelectorType": "LIST","guid": "userAssignment","displayName": "Folder",
	"value": "FOLDER","description": "","selectorOptions": []}]},"responseCode":200}}`
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/projectVersions/10172/issueSelectorSet" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(response))
			return
		}
	})
	// Close the server when test finishes
	defer server.Close()

	t.Run("test success one", func(t *testing.T) {
		result, err := sys.GetIssueFilterSelectorOfProjectVersionByName(10172, []string{"Analysis"}, nil)
		assert.NoError(t, err, "GetIssueFilterSelectorOfProjectVersionByName call not successful")
		assert.NotNil(t, result, "Expected non nil value")
		assert.Equal(t, 1, len(result.FilterBySet), "Different result expected")
		assert.Equal(t, 1, len(result.GroupBySet), "Different result expected")
	})

	t.Run("test success several", func(t *testing.T) {
		result, err := sys.GetIssueFilterSelectorOfProjectVersionByName(10172, []string{"Analysis", "Folder"}, nil)
		assert.NoError(t, err, "GetIssueFilterSelectorOfProjectVersionByName call not successful")
		assert.NotNil(t, result, "Expected non nil value")
		assert.Equal(t, 2, len(result.FilterBySet), "Different result expected")
		assert.Equal(t, 1, len(result.GroupBySet), "Different result expected")
	})

	t.Run("test empty", func(t *testing.T) {
		result, err := sys.GetIssueFilterSelectorOfProjectVersionByName(10172, []string{"Some", "Other"}, nil)
		assert.NoError(t, err, "GetIssueFilterSelectorOfProjectVersionByName call not successful")
		assert.NotNil(t, result, "Expected non nil value")
		assert.Equal(t, 0, len(result.FilterBySet), "Different result expected")
		assert.Equal(t, 0, len(result.GroupBySet), "Different result expected")
	})
}

func TestReduceIssueFilterSelectorSet(t *testing.T) {
	sys, _ := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {})
	name1 := "Special"
	name2 := "Other"
	guid := "FOLDER"
	options := []*models.SelectorOption{{GUID: "1234567", DisplayName: "Test"}, {GUID: "1234568", DisplayName: "Test2"}}
	filterSet := models.IssueFilterSelectorSet{FilterBySet: []*models.IssueFilterSelector{}, GroupBySet: []*models.IssueSelector{}}
	filterSet.FilterBySet = append(filterSet.FilterBySet, &models.IssueFilterSelector{DisplayName: name1, SelectorOptions: options})
	filterSet.FilterBySet = append(filterSet.FilterBySet, &models.IssueFilterSelector{DisplayName: name2})
	filterSet.GroupBySet = append(filterSet.GroupBySet, &models.IssueSelector{DisplayName: &name2, GUID: &guid})
	reducedFilterSet := sys.ReduceIssueFilterSelectorSet(&filterSet, []string{"Special"}, []string{"Test"})
	assert.Equal(t, 1, len(reducedFilterSet.FilterBySet), "Different result expected")
	assert.Equal(t, 1, len(reducedFilterSet.FilterBySet[0].SelectorOptions), "Different result expected")
	assert.Equal(t, "Test", reducedFilterSet.FilterBySet[0].SelectorOptions[0].DisplayName, "Different result expected")
	assert.Equal(t, 0, len(reducedFilterSet.GroupBySet), "Different result expected")
}

func TestGetProjectIssuesByIDAndFilterSetGroupedBySelector(t *testing.T) {
	// Start a local HTTP server
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/projectVersions/10172/filterSets" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(`{"data":[{"defaultFilterSet":true,"folders":[
				{"id":1,"guid":"4711","name":"Corporate Security Requirements","color":"000000"},
				{"id":2,"guid":"4712","name":"Audit All","color":"ff0000"},
				{"id":3,"guid":"4713","name":"Spot Checks of Each Category","color":"ff8000"},
				{"id":4,"guid":"4714","name":"Optional","color":"808080"}],"description":"",
				"guid":"666","title":"Special"}],"count":1,"responseCode":200}}`))
			return
		}
		if req.URL.Path == "/projectVersions/10172/issueGroups" {
			assert.Equal(t, "filterset=666&groupingtype=FOLDER&showsuppressed=true", req.URL.RawQuery)
			return
		}
		rw.WriteHeader(400)
	})
	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		name := "Special"
		guid := "FOLDER"
		filterSet := models.IssueFilterSelectorSet{FilterBySet: []*models.IssueFilterSelector{}, GroupBySet: []*models.IssueSelector{}}
		filterSet.FilterBySet = append(filterSet.FilterBySet, &models.IssueFilterSelector{DisplayName: name})
		filterSet.GroupBySet = append(filterSet.GroupBySet, &models.IssueSelector{DisplayName: &name, GUID: &guid})
		_, err := sys.GetProjectIssuesByIDAndFilterSetGroupedBySelector(10172, "", "666", &filterSet)
		assert.NoError(t, err, "GetProjectIssuesByIDAndFilterSetGroupedByFolder call not successful")
	})
}

func TestGetProjectIssuesByIDAndFilterSetGroupedByCategory(t *testing.T) {
	// Start a local HTTP server
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/projectVersions/10172/filterSets" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(`{"data":[{"defaultFilterSet":true,"folders":[
				{"id":1,"guid":"4711","name":"Corporate Security Requirements","color":"000000"},
				{"id":2,"guid":"4712","name":"Audit All","color":"ff0000"},
				{"id":3,"guid":"4713","name":"Spot Checks of Each Category","color":"ff8000"},
				{"id":4,"guid":"4714","name":"Optional","color":"808080"}],"description":"",
				"guid":"666","title":"Special"}],"count":1,"responseCode":200}}`))
			return
		}
		if req.URL.Path == "/projectVersions/10172/issueGroups" {
			assert.Equal(t, "filter=4713&filterset=666&groupingtype=11111111-1111-1111-1111-111111111165&showsuppressed=true", req.URL.RawQuery)
			return
		}
		rw.WriteHeader(400)
	})
	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		name := "Special"
		guid := "11111111-1111-1111-1111-111111111165"
		filterSet := models.IssueFilterSelectorSet{FilterBySet: []*models.IssueFilterSelector{}, GroupBySet: []*models.IssueSelector{}}
		filterSet.FilterBySet = append(filterSet.FilterBySet, &models.IssueFilterSelector{DisplayName: name})
		filterSet.GroupBySet = append(filterSet.GroupBySet, &models.IssueSelector{DisplayName: &name, GUID: &guid})
		_, err := sys.GetProjectIssuesByIDAndFilterSetGroupedBySelector(10172, "4713", "666", &filterSet)
		assert.NoError(t, err, "GetProjectIssuesByIDAndFilterSetGroupedByCategory call not successful")
	})
}

func TestGetProjectIssuesByIDAndFilterSetGroupedByAnalysis(t *testing.T) {
	// Start a local HTTP server
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/projectVersions/10172/filterSets" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(`{"data":[{"defaultFilterSet":true,"folders":[
				{"id":1,"guid":"4711","name":"Corporate Security Requirements","color":"000000"},
				{"id":2,"guid":"4712","name":"Audit All","color":"ff0000"},
				{"id":3,"guid":"4713","name":"Spot Checks of Each Category","color":"ff8000"},
				{"id":4,"guid":"4714","name":"Optional","color":"808080"}],"description":"",
				"guid":"666","title":"Special"}],"count":1,"responseCode":200}}`))
			return
		}
		if req.URL.Path == "/projectVersions/10172/issueGroups" {
			assert.Equal(t, "filterset=666&groupingtype=87f2364f-dcd4-49e6-861d-f8d3f351686b&showsuppressed=true", req.URL.RawQuery)
			return
		}
		rw.WriteHeader(400)
	})
	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		name := "Special"
		guid := "87f2364f-dcd4-49e6-861d-f8d3f351686b"
		filterSet := models.IssueFilterSelectorSet{FilterBySet: []*models.IssueFilterSelector{}, GroupBySet: []*models.IssueSelector{}}
		filterSet.FilterBySet = append(filterSet.FilterBySet, &models.IssueFilterSelector{DisplayName: name})
		filterSet.GroupBySet = append(filterSet.GroupBySet, &models.IssueSelector{DisplayName: &name, GUID: &guid})
		_, err := sys.GetProjectIssuesByIDAndFilterSetGroupedBySelector(10172, "", "666", &filterSet)
		assert.NoError(t, err, "GetProjectIssuesByIDAndFilterSetGroupedByAnalysis call not successful")
	})
}

func TestGetIssueStatisticsOfProjectVersion(t *testing.T) {
	// Start a local HTTP server
	response := `{"data": [{"filterSetId": 3887,"hiddenCount": 0,"suppressedDisplayableCount": 0,"suppressedCount": 11,"hiddenDisplayableCount": 0,"projectVersionId": 10172,
				"removedDisplayableCount": 0,"removedCount": 747}],"count": 1,"responseCode": 200}`
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/projectVersions/10172/issueStatistics" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(response))
			return
		}
	})
	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		result, err := sys.GetIssueStatisticsOfProjectVersion(10172)
		assert.NoError(t, err, "GetArtifactsOfProjectVersion call not successful")
		assert.Equal(t, 1, len(result), "Different result content expected")
		assert.Equal(t, int64(10172), *result[0].ProjectVersionID, "Different result content expected")
		assert.Equal(t, int32(11), *result[0].SuppressedCount, "Different result content expected")
	})
}

func TestGenerateQGateReport(t *testing.T) {
	// Start a local HTTP server
	data := ""
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/reports" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			bodyBytes, _ := io.ReadAll(req.Body)
			data = string(bodyBytes)
			response := `{"data": `
			response += data
			response += `,"responseCode": 201}`
			rw.WriteHeader(201)
			rw.Write([]byte(response))
			return
		}
	})
	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		result, err := sys.GenerateQGateReport(2837, 17540, 18, "Fortify", "develop", "PDF")
		assert.NoError(t, err, "GetArtifactsOfProjectVersion call not successful")
		assert.Equal(t, int64(2837), result.Projects[0].ID, "Different result content expected")
		assert.Equal(t, int64(17540), result.Projects[0].Versions[0].ID, "Different result content expected")
		assert.Equal(t, int64(18), *result.ReportDefinitionID, "Different result content expected")
	})
}

func TestGetReportDetails(t *testing.T) {
	// Start a local HTTP server
	response := `{"data": {"id":999,"name":"FortifyReport","note":"","type":"PORTFOLIO","reportDefinitionId":18,"format":"PDF",
	"projects":[{"id":2837,"name":"Fortify","versions":[{"id":17540,"name":"develop"}]}],"projectVersionDisplayName":"develop",
	"inputReportParameters":[{"name":"Q-gate-report","identifier":"projectVersionId","paramValue":17540,"type":"SINGLE_PROJECT"}]},"count": 1,"responseCode": 200}`
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/reports/999" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(response))
			return
		}
	})
	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		result, err := sys.GetReportDetails(999)
		assert.NoError(t, err, "GetReportDetails call not successful")
		assert.Equal(t, int64(999), result.ID, "Different result content expected")
	})
}

func TestGetFileToken(t *testing.T) {
	// Start a local HTTP server
	bodyContent := ""
	reference := `{"fileTokenType":"TOKEN_TYPE"}
`
	response := `{"data": {"fileTokenType": "TOKEN_TYPE","token": "ZjE1OTdjZjEtMjAzNS00NTFmLThiOWItNzBkYzI0MWEzZGNj"},"responseCode": 201}`
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/fileTokens" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			bodyBytes, _ := io.ReadAll(req.Body)
			bodyContent = string(bodyBytes)
			rw.WriteHeader(201)
			rw.Write([]byte(response))
			return
		}
	})
	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		result, err := sys.getFileToken("TOKEN_TYPE")
		assert.NoError(t, err)
		assert.Equal(t, "ZjE1OTdjZjEtMjAzNS00NTFmLThiOWItNzBkYzI0MWEzZGNj", result.Token)
		assert.Equal(t, reference, bodyContent)
	})
}

func TestInvalidateFileToken(t *testing.T) {
	// Start a local HTTP server
	response := `{"responseCode": 200}`
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/fileTokens" && req.Method == "DELETE" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.WriteHeader(200)
			rw.Write([]byte(response))
			return
		}
	})
	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		err := sys.invalidateFileTokens()
		assert.NoError(t, err, "invalidateFileTokens call not successful")
	})
}

func TestUploadResultFile(t *testing.T) {
	// Start a local HTTP server
	bodyContent := ""
	getTokenCalled := false
	invalidateTokenCalled := false
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/fileTokens" && req.Method == "DELETE" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.WriteHeader(200)
			rw.Write([]byte(`{"responseCode": 200}`))
			invalidateTokenCalled = true
			return
		}
		if req.URL.Path == "/fileTokens" && req.Method == "POST" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.WriteHeader(201)
			rw.Write([]byte(`{"data": {"token": "89ee873"}, "responseCode": 201}`))
			getTokenCalled = true
			return
		}
		if req.URL.Path == "/upload/resultFileUpload.html" && req.URL.RawQuery == "mat=89ee873" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			bodyBytes, _ := io.ReadAll(req.Body)
			bodyContent = string(bodyBytes)
			rw.WriteHeader(200)
			rw.Write([]byte("OK"))
			return
		}
	})
	// Close the server when test finishes
	defer server.Close()

	testFile, err := os.CreateTemp("", "result.fpr")
	if err != nil {
		t.FailNow()
	}
	defer os.RemoveAll(testFile.Name()) // clean up

	t.Run("test success", func(t *testing.T) {
		err := sys.UploadResultFile("/upload/resultFileUpload.html", testFile.Name(), 10770)
		assert.NoError(t, err, "UploadFile call not successful")
		assert.Contains(t, bodyContent, `Content-Disposition: form-data; name="file"; filename=`, "Expected different content in request body")
		assert.Contains(t, bodyContent, `Content-Disposition: form-data; name="entityId"`, "Expected different content in request body")
		assert.Contains(t, bodyContent, `10770`, "Expected different content in request body")
		assert.Equal(t, true, getTokenCalled, "Expected GetUploadToken to be called")
		assert.Equal(t, true, invalidateTokenCalled, "Expected InvalidateFileTokens to be called")
	})
}

func TestDownloadResultFile(t *testing.T) {
	// Start a local HTTP server
	bodyContent := ""
	getTokenCalled := false
	invalidateTokenCalled := false
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/fileTokens" && req.Method == "DELETE" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.WriteHeader(200)
			rw.Write([]byte(`{"responseCode": 200}`))
			invalidateTokenCalled = true
			return
		}
		if req.URL.Path == "/fileTokens" && req.Method == "POST" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.WriteHeader(201)
			rw.Write([]byte(`{"data": {"token": "89ee873"}, "responseCode": 201}`))
			getTokenCalled = true
			return
		}
		if req.URL.Path == "/download/currentStateFprDownload.html" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			bodyContent = req.URL.RawQuery
			rw.WriteHeader(200)
			rw.Write([]byte("OK"))
			return
		}
	})
	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		data, err := sys.DownloadResultFile("/download/currentStateFprDownload.html", 10775)
		assert.NoError(t, err, "DownloadResultFile call not successful")
		assert.Equal(t, "id=10775&mat=89ee873", bodyContent, "Expected different request body")
		assert.Equal(t, []byte("OK"), data, "Expected different result")
		assert.Equal(t, true, getTokenCalled, "Expected GetUploadToken to be called")
		assert.Equal(t, true, invalidateTokenCalled, "Expected InvalidateFileTokens to be called")
	})
}

func TestDownloadReportFile(t *testing.T) {
	// Start a local HTTP server
	getTokenCalled := false
	invalidateTokenCalled := false
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/fileTokens" && req.Method == "DELETE" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.WriteHeader(200)
			rw.Write([]byte(`{"responseCode": 200}`))
			invalidateTokenCalled = true
			return
		}
		if req.URL.Path == "/fileTokens" && req.Method == "POST" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.WriteHeader(201)
			rw.Write([]byte(`{"data": {"token": "89ee873"}, "responseCode": 201}`))
			getTokenCalled = true
			return
		}
		if req.URL.Path == "/transfer/reportDownload.html" && req.URL.RawQuery == "id=10775&mat=89ee873" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.WriteHeader(200)
			rw.Write([]byte("OK"))
			return
		}
	})
	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		data, err := sys.DownloadReportFile("/transfer/reportDownload.html", 10775)
		assert.NoError(t, err, "DownloadReportFile call not successful")
		assert.Equal(t, []byte("OK"), data, "Expected different result")
		assert.Equal(t, true, getTokenCalled, "Expected GetUploadToken to be called")
		assert.Equal(t, true, invalidateTokenCalled, "Expected InvalidateFileTokens to be called")
	})
}

func TestLookupOrCreateProjectVersionDetailsForPullRequest(t *testing.T) {
	// Start a local HTTP server
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/projects/4711/versions" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.WriteHeader(200)
			rw.Write([]byte(`{"data": [], "count": 0, "responseCode": 200}`))
			return
		}
		if req.URL.Path == "/projectVersions" && req.Method == "POST" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			bodyBytes, _ := io.ReadAll(req.Body)
			bodyContent := string(bodyBytes)
			responseContent := `{"data": `
			responseContent += bodyContent
			responseContent += `,"count": 1,"responseCode": 201,"links": {}}`
			fmt.Println(responseContent)
			rw.WriteHeader(201)
			rw.Write([]byte(responseContent))
			return
		}
		if req.URL.Path == "/projectVersions/4711/attributes" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(
				`{"data": [{"_href": "https://fortify/ssc/api/v1/projectVersions/4711/attributes/4712","attributeDefinitionId": 31,
				"values": null,"guid": "gdgfdgfdgfdgfd","id": 4712,"value": "abcd"}],"count": 8,"responseCode": 200}`))
			return
		}
		if req.URL.Path == "/projectVersions/4712/attributes" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			bodyBytes, _ := io.ReadAll(req.Body)
			bodyString := string(bodyBytes)
			response := `{"data": `
			response += bodyString
			response += `,"count": 1,"responseCode": 200}`
			rw.WriteHeader(200)
			rw.Write([]byte(response))
			return
		}
		if req.URL.Path == "/projectVersions/action/copyFromPartial" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(
				`{"data":[{"latestScanId":null,"serverVersion":17.2,"tracesOutOfDate":false,"attachmentsOutOfDate":false,"description":"",
				"project":{"id":4711,"name":"python-test","description":"","creationDate":"2018-12-03T06:29:38.197+0000","createdBy":"someUser",
				"issueTemplateId":"dasdasdasdsadasdasdasdasdas"},"sourceBasePath":null,"mode":"BASIC","masterAttrGuid":"sddasdasda","obfuscatedId":null,
				"id":10172,"customTagValuesAutoApply":null,"issueTemplateId":"dasdasdasdsadasdasdasdasdas","loadProperties":null,"predictionPolicy":null,
				"bugTrackerPluginId":null,"owner":"admin","_href":"https://fortify/ssc/api/v1/projectVersions/10172",
				"committed":true,"bugTrackerEnabled":false,"active":true,"snapshotOutOfDate":false,"issueTemplateModifiedTime":1578411924701,
				"securityGroup":null,"creationDate":"2018-02-09T16:59:41.297+0000","refreshRequired":false,"issueTemplateName":"someTemplate",
				"migrationVersion":null,"createdBy":"admin","name":"0","siteId":null,"staleIssueTemplate":false,"autoPredict":null,
				"currentState":{"id":10172,"committed":true,"attentionRequired":false,"analysisResultsExist":true,"auditEnabled":true,
				"lastFprUploadDate":"2018-02-09T16:59:53.497+0000","extraMessage":null,"analysisUploadEnabled":true,"batchBugSubmissionExists":false,
				"hasCustomIssues":false,"metricEvaluationDate":"2018-03-10T00:02:45.553+0000","deltaPeriod":7,"issueCountDelta":0,"percentAuditedDelta":0.0,
				"criticalPriorityIssueCountDelta":0,"percentCriticalPriorityIssuesAuditedDelta":0.0},"assignedIssuesCount":0,"status":null}],
				"count":1,"responseCode":200,"links":{"last":{"href":"https://fortify/ssc/api/v1/projects/4711/versions?start=0"},
				"first":{"href":"https://fortify/ssc/api/v1/projects/4711/versions?start=0"}}}`))
			return
		}
		if req.URL.Path == "/projectVersions/10172" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(`{"data": {"active":null,"bugTrackerEnabled":null,"bugTrackerPluginId":null,"committed":true,"createdBy":null,"creationDate":null,"description":null,"issueTemplateId":null,"issueTemplateModifiedTime":null,"issueTemplateName":null,"latestScanId":null,"masterAttrGuid":null,"name":null,"owner":null,"serverVersion":null,"snapshotOutOfDate":null,"staleIssueTemplate":null}, "responseCode": 200}`))
			return
		}
		if req.URL.Path == "/projectVersions/action/copyCurrentState" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(
				`{"data":[{"latestScanId":null,"serverVersion":17.2,"tracesOutOfDate":false,"attachmentsOutOfDate":false,"description":"",
				"project":{"id":4711,"name":"python-test","description":"","creationDate":"2018-12-03T06:29:38.197+0000","createdBy":"someUser",
				"issueTemplateId":"dasdasdasdsadasdasdasdasdas"},"sourceBasePath":null,"mode":"BASIC","masterAttrGuid":"sddasdasda","obfuscatedId":null,
				"id":10172,"customTagValuesAutoApply":null,"issueTemplateId":"dasdasdasdsadasdasdasdasdas","loadProperties":null,"predictionPolicy":null,
				"bugTrackerPluginId":null,"owner":"admin","_href":"https://fortify/ssc/api/v1/projectVersions/10172",
				"committed":true,"bugTrackerEnabled":false,"active":true,"snapshotOutOfDate":false,"issueTemplateModifiedTime":1578411924701,
				"securityGroup":null,"creationDate":"2018-02-09T16:59:41.297+0000","refreshRequired":false,"issueTemplateName":"someTemplate",
				"migrationVersion":null,"createdBy":"admin","name":"0","siteId":null,"staleIssueTemplate":false,"autoPredict":null,
				"currentState":{"id":10172,"committed":true,"attentionRequired":false,"analysisResultsExist":true,"auditEnabled":true,
				"lastFprUploadDate":"2018-02-09T16:59:53.497+0000","extraMessage":null,"analysisUploadEnabled":true,"batchBugSubmissionExists":false,
				"hasCustomIssues":false,"metricEvaluationDate":"2018-03-10T00:02:45.553+0000","deltaPeriod":7,"issueCountDelta":0,"percentAuditedDelta":0.0,
				"criticalPriorityIssueCountDelta":0,"percentCriticalPriorityIssuesAuditedDelta":0.0},"assignedIssuesCount":0,"status":null}],
				"count":1,"responseCode":200,"links":{"last":{"href":"https://fortify/ssc/api/v1/projects/4711/versions?start=0"},
				"first":{"href":"https://fortify/ssc/api/v1/projects/4711/versions?start=0"}}}`))
			return
		}
		if req.URL.Path == "/projectVersions/10172/authEntities" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(`{"data": [{"displayName":"some user","email":"some.one@test.com","entityName":"some_user","firstName":"some","id":589,"lastName":"user","type":"User"}],"count": 1,"responseCode": 200}`))
			return
		}
		if req.URL.Path == "/projectVersions/10173/authEntities" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(`{"data": [{"displayName":"some user","email":"some.one@test.com","entityName":"some_user","firstName":"some","id":589,"lastName":"user","type":"User"}],"count": 1,"responseCode": 200}`))
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
		masterProjectVersion := models.ProjectVersion{AssignedIssuesCount: int64Value, Project: &project, Name: &name, Active: &enabled,
			Committed: &enabled, AttachmentsOutOfDate: disabled, AutoPredict: disabled, BugTrackerEnabled: &disabled,
			CustomTagValuesAutoApply: disabled, RefreshRequired: disabled, Owner: &owner, ServerVersion: &float32Value,
			SnapshotOutOfDate: &disabled, StaleIssueTemplate: &disabled, MasterAttrGUID: &masterGUID,
			LatestScanID: &int64Value, IssueTemplateName: &name, IssueTemplateModifiedTime: &int64Value,
			IssueTemplateID: &name, Description: &name, CreatedBy: &owner, BugTrackerPluginID: &name, Mode: "NONE",
			CurrentState: &projectVersionState, ID: int64Value, LoadProperties: "", CreationDate: &now,
			MigrationVersion: float32Value, ObfuscatedID: "", PredictionPolicy: "", SecurityGroup: "",
			SiteID: "", SourceBasePath: "", Status: "", TracesOutOfDate: false}
		prProjectVersion, err := sys.LookupOrCreateProjectVersionDetailsForPullRequest(4711, &masterProjectVersion, "PR-815")
		assert.NoError(t, err, "LookupOrCreateProjectVersionDetailsForPullRequest call not successful")
		assert.Equal(t, "PR-815", *prProjectVersion.Name, "Expected different result")
		assert.Equal(t, masterProjectVersion.Description, prProjectVersion.Description, "Expected different result")
		assert.Equal(t, masterProjectVersion.Active, prProjectVersion.Active, "Expected different result")
		assert.Equal(t, masterProjectVersion.Committed, prProjectVersion.Committed, "Expected different result")
		assert.Equal(t, masterProjectVersion.Project.Name, prProjectVersion.Project.Name, "Expected different result")
		assert.Equal(t, masterProjectVersion.Project.Description, prProjectVersion.Project.Description, "Expected different result")
		assert.Equal(t, masterProjectVersion.Project.ID, prProjectVersion.Project.ID, "Expected different result")
		assert.Equal(t, masterProjectVersion.IssueTemplateID, prProjectVersion.IssueTemplateID, "Expected different result")
	})
}

func TestMergeProjectVersionStateOfPRIntoMaster(t *testing.T) {
	// Start a local HTTP server
	getPRProjectVersionCalled := false
	invalidateTokenCalled := false
	getTokenCalled := false
	downloadCalled := false
	uploadCalled := false
	inactivateCalled := false
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/projects/4711/versions" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(`{"data":[{"latestScanId":null,"serverVersion":17.2,"tracesOutOfDate":false,"attachmentsOutOfDate":false,"description":"",
			"project":{"id":4711,"name":"product.some.com","description":"","creationDate":"2018-12-03T06:29:38.197+0000","createdBy":"someUser",
			"issueTemplateId":"dasdasdasdsadasdasdasdasdas"},"sourceBasePath":null,"mode":"BASIC","masterAttrGuid":"sddasdasda","obfuscatedId":null,
			"id":10172,"customTagValuesAutoApply":null,"issueTemplateId":"dasdasdasdsadasdasdasdasdas","loadProperties":null,"predictionPolicy":null,
			"bugTrackerPluginId":null,"owner":"admin","_href":"https://fortify/ssc/api/v1/projectVersions/10172",
			"committed":true,"bugTrackerEnabled":false,"active":true,"snapshotOutOfDate":false,"issueTemplateModifiedTime":1578411924701,
			"securityGroup":null,"creationDate":"2018-02-09T16:59:41.297+0000","refreshRequired":false,"issueTemplateName":"someTemplate",
			"migrationVersion":null,"createdBy":"admin","name":"PR-815","siteId":null,"staleIssueTemplate":false,"autoPredict":null,
			"currentState":{"id":10172,"committed":true,"attentionRequired":false,"analysisResultsExist":true,"auditEnabled":true,
			"lastFprUploadDate":"2018-02-09T16:59:53.497+0000","extraMessage":null,"analysisUploadEnabled":true,"batchBugSubmissionExists":false,
			"hasCustomIssues":false,"metricEvaluationDate":"2018-03-10T00:02:45.553+0000","deltaPeriod":7,"issueCountDelta":0,"percentAuditedDelta":0.0,
			"criticalPriorityIssueCountDelta":0,"percentCriticalPriorityIssuesAuditedDelta":0.0},"assignedIssuesCount":0,"status":null}],
			"count":1,"responseCode":200,"links":{"last":{"href":"https://fortify/ssc/api/v1/projects/4711/versions?start=0"},
			"first":{"href":"https://fortify/ssc/api/v1/projects/4711/versions?start=0"}}}`))
			getPRProjectVersionCalled = true
			return
		}
		if req.URL.Path == "/fileTokens" && req.Method == "DELETE" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.WriteHeader(200)
			rw.Write([]byte(`{"responseCode": 200}`))
			invalidateTokenCalled = true
			return
		}
		if req.URL.Path == "/fileTokens" && req.Method == "POST" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.WriteHeader(201)
			rw.Write([]byte(`{"data": {"token": "89ee873"}, "responseCode": 201}`))
			getTokenCalled = true
			return
		}
		if req.URL.Path == "/download/currentStateFprDownload.html" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.WriteHeader(200)
			rw.Write([]byte("OK"))
			downloadCalled = true
			return
		}
		if req.URL.Path == "/upload/resultFileUpload.html" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.WriteHeader(200)
			rw.Write([]byte("OK"))
			uploadCalled = true
			return
		}
		if req.URL.Path == "/projectVersions/10172" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(`{"data": {"active":false,"bugTrackerEnabled":null,"bugTrackerPluginId":null,"committed":true,"createdBy":null,"creationDate":null,"description":null,"issueTemplateId":null,"issueTemplateModifiedTime":null,"issueTemplateName":null,"latestScanId":null,"masterAttrGuid":null,"name":null,"owner":null,"serverVersion":null,"snapshotOutOfDate":null,"staleIssueTemplate":null}, "responseCode": 200}`))
			inactivateCalled = true
			return
		}
	})
	// Close the server when test finishes
	defer server.Close()

	t.Run("test success", func(t *testing.T) {
		err := sys.MergeProjectVersionStateOfPRIntoMaster("/download/currentStateFprDownload.html", "/upload/resultFileUpload.html", 4711, 10171, "PR-815")
		assert.NoError(t, err, "MergeProjectVersionStateOfPRIntoMaster call not successful")
		assert.Equal(t, true, getPRProjectVersionCalled, "Expected different value")
		assert.Equal(t, true, invalidateTokenCalled, "Expected different value")
		assert.Equal(t, true, getTokenCalled, "Expected different value")
		assert.Equal(t, true, downloadCalled, "Expected different value")
		assert.Equal(t, true, uploadCalled, "Expected different value")
		assert.Equal(t, true, inactivateCalled, "Expected different value")
	})
}

func TestBase64EndodePlainToken(t *testing.T) {
	t.Run("Encoded token untouched", func(t *testing.T) {
		token := "OTUzODcwNDYtNWFjOC00NTcwLTg3NWQtYTVlYzhiZDhkM2Qy"
		encodedToken := base64EndodePlainToken(token)
		assert.Equal(t, token, encodedToken)
	})
	t.Run("Unencoded token gets encoded", func(t *testing.T) {
		token := "95387046-5ac8-4570-875d-a5ec8bd8d3d2"
		encodedToken := base64EndodePlainToken(token)
		assert.Equal(t, "OTUzODcwNDYtNWFjOC00NTcwLTg3NWQtYTVlYzhiZDhkM2Qy", encodedToken)
	})
}

func TestCreateJSONReport(t *testing.T) {
	t.Run("test success", func(t *testing.T) {
		spotChecksCountByCategory := []SpotChecksAuditCount{}
		spotChecksCountByCategory = append(spotChecksCountByCategory, SpotChecksAuditCount{Audited: 3, Total: 3, Type: "J2EE Misconfiguration: Missing Error Handling"})
		spotChecksCountByCategory = append(spotChecksCountByCategory, SpotChecksAuditCount{Audited: 1, Total: 3, Type: "J2EE Bad Practices: Leftover Debug Code"})
		fortifyReportData := FortifyReportData{CorporateAudited: 30, CorporateTotal: 30, AuditAllTotal: 1, AuditAllAudited: 1, ProjectVersionID: 4999}
		jsonReport := CreateJSONReport(fortifyReportData, spotChecksCountByCategory, "https://fortify-test.com/ssc")
		assert.Equal(t, true, jsonReport.AtleastOneSpotChecksCategoryAudited)
		assert.Equal(t, true, jsonReport.IsSpotChecksPerCategoryAudited)
		assert.Equal(t, 1, jsonReport.AuditAllAudited)
		assert.Equal(t, 1, jsonReport.AuditAllTotal)
		assert.Equal(t, 30, jsonReport.CorporateAudited)
		assert.Equal(t, 30, jsonReport.CorporateTotal)
		assert.Equal(t, "https://fortify-test.com/ssc/html/ssc/version/4999", jsonReport.URL)
		assert.Equal(t, "https://fortify-test.com/ssc", jsonReport.ToolInstance)
	})

	t.Run("atleast one category spotchecks failed", func(t *testing.T) {
		spotChecksCountByCategory := []SpotChecksAuditCount{}
		spotChecksCountByCategory = append(spotChecksCountByCategory, SpotChecksAuditCount{Audited: 3, Total: 3, Type: "J2EE Misconfiguration: Missing Error Handling"})
		spotChecksCountByCategory = append(spotChecksCountByCategory, SpotChecksAuditCount{Audited: 0, Total: 1, Type: "J2EE Bad Practices: Leftover Debug Code"})
		fortifyReportData := FortifyReportData{CorporateAudited: 0, CorporateTotal: 0, AuditAllTotal: 0, AuditAllAudited: 0}
		jsonReport := CreateJSONReport(fortifyReportData, spotChecksCountByCategory, "https://fortify-test.com/ssc")
		assert.Equal(t, false, jsonReport.AtleastOneSpotChecksCategoryAudited)
		assert.Equal(t, false, jsonReport.IsSpotChecksPerCategoryAudited)
	})

	t.Run("no spot checks audited", func(t *testing.T) {
		spotChecksCountByCategory := []SpotChecksAuditCount{}
		fortifyReportData := FortifyReportData{CorporateAudited: 0, CorporateTotal: 0, AuditAllTotal: 0, AuditAllAudited: 0}
		jsonReport := CreateJSONReport(fortifyReportData, spotChecksCountByCategory, "https://fortify-test.com/ssc")
		assert.Equal(t, true, jsonReport.AtleastOneSpotChecksCategoryAudited)
		assert.Equal(t, true, jsonReport.IsSpotChecksPerCategoryAudited)
	})

	t.Run("isSpotChecksPerCategoryAudited passed spotchecks test 1", func(t *testing.T) {
		spotChecksCountByCategory := []SpotChecksAuditCount{}
		spotChecksCountByCategory = append(spotChecksCountByCategory, SpotChecksAuditCount{Audited: 10, Total: 100, Type: "J2EE Misconfiguration: Missing Error Handling"})
		fortifyReportData := FortifyReportData{CorporateAudited: 0, CorporateTotal: 0, AuditAllTotal: 0, AuditAllAudited: 0}
		jsonReport := CreateJSONReport(fortifyReportData, spotChecksCountByCategory, "https://fortify-test.com/ssc")
		assert.Equal(t, true, jsonReport.AtleastOneSpotChecksCategoryAudited)
		assert.Equal(t, true, jsonReport.IsSpotChecksPerCategoryAudited)
	})

	t.Run("isSpotChecksPerCategoryAudited failed spotchecks test 2", func(t *testing.T) {
		spotChecksCountByCategory := []SpotChecksAuditCount{}
		spotChecksCountByCategory = append(spotChecksCountByCategory, SpotChecksAuditCount{Audited: 3, Total: 100, Type: "J2EE Misconfiguration: Missing Error Handling"})
		fortifyReportData := FortifyReportData{CorporateAudited: 0, CorporateTotal: 0, AuditAllTotal: 0, AuditAllAudited: 0}
		jsonReport := CreateJSONReport(fortifyReportData, spotChecksCountByCategory, "https://fortify-test.com/ssc")
		assert.Equal(t, true, jsonReport.AtleastOneSpotChecksCategoryAudited)
		assert.Equal(t, false, jsonReport.IsSpotChecksPerCategoryAudited)
	})

	t.Run("isSpotChecksPerCategoryAudited failed spotchecks test 3", func(t *testing.T) {
		spotChecksCountByCategory := []SpotChecksAuditCount{}
		spotChecksCountByCategory = append(spotChecksCountByCategory, SpotChecksAuditCount{Audited: 9, Total: 200, Type: "J2EE Misconfiguration: Missing Error Handling"})
		fortifyReportData := FortifyReportData{CorporateAudited: 0, CorporateTotal: 0, AuditAllTotal: 0, AuditAllAudited: 0}
		jsonReport := CreateJSONReport(fortifyReportData, spotChecksCountByCategory, "https://fortify-test.com/ssc")
		assert.Equal(t, true, jsonReport.AtleastOneSpotChecksCategoryAudited)
		assert.Equal(t, false, jsonReport.IsSpotChecksPerCategoryAudited)
	})

	t.Run("isSpotChecksPerCategoryAudited passed spotchecks test 4", func(t *testing.T) {
		spotChecksCountByCategory := []SpotChecksAuditCount{}
		spotChecksCountByCategory = append(spotChecksCountByCategory, SpotChecksAuditCount{Audited: 10, Total: 200, Type: "J2EE Misconfiguration: Missing Error Handling"})
		fortifyReportData := FortifyReportData{CorporateAudited: 0, CorporateTotal: 0, AuditAllTotal: 0, AuditAllAudited: 0}
		jsonReport := CreateJSONReport(fortifyReportData, spotChecksCountByCategory, "https://fortify-test.com/ssc")
		assert.Equal(t, true, jsonReport.AtleastOneSpotChecksCategoryAudited)
		assert.Equal(t, true, jsonReport.IsSpotChecksPerCategoryAudited)
	})
}
