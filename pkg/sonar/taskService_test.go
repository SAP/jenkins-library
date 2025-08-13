package sonar

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	sonargo "github.com/magicsong/sonargo/sonar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
)

func TestGetTask(t *testing.T) {
	testURL := "https://example.org"
	t.Run("success", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		sender := &piperhttp.Client{}
		sender.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		// add response handler
		httpmock.RegisterResponder(http.MethodGet, testURL+"/api/"+EndpointCeTask+"", httpmock.NewStringResponder(http.StatusOK, responseCeTaskSuccess))
		// create service instance
		serviceUnderTest := NewTaskService(testURL, mock.Anything, mock.Anything, sender)
		// test
		result, response, err := serviceUnderTest.GetTask(&sonargo.CeTaskOption{Id: mock.Anything})
		// assert
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.NotEmpty(t, response)
		assert.Equal(t, 1, httpmock.GetTotalCallCount(), "unexpected number of requests")
	})
	t.Run("request error", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		sender := &piperhttp.Client{}
		sender.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		// add response handler
		httpmock.RegisterResponder(http.MethodGet, testURL+"/api/"+EndpointCeTask+"", httpmock.NewErrorResponder(errors.New("internal server error")))
		// create service instance
		serviceUnderTest := NewTaskService(testURL, mock.Anything, mock.Anything, sender)
		// test
		result, response, err := serviceUnderTest.GetTask(&sonargo.CeTaskOption{Id: mock.Anything})
		// assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "internal server error")
		assert.Empty(t, result)
		assert.Empty(t, response)
		assert.Equal(t, 1, httpmock.GetTotalCallCount(), "unexpected number of requests")
	})
	t.Run("server error", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		sender := &piperhttp.Client{}
		sender.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		// add response handler
		httpmock.RegisterResponder(http.MethodGet, testURL+"/api/"+EndpointCeTask+"", httpmock.NewStringResponder(http.StatusNotFound, responseCeTaskError))
		// create service instance
		serviceUnderTest := NewTaskService(testURL, mock.Anything, mock.Anything, sender)
		// test
		result, response, err := serviceUnderTest.GetTask(&sonargo.CeTaskOption{Id: mock.Anything})
		// assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "No activity found for task ")
		assert.Empty(t, result)
		assert.NotEmpty(t, response)
		assert.Equal(t, 1, httpmock.GetTotalCallCount(), "unexpected number of requests")
	})
}

func TestWaitForTask(t *testing.T) {
	testURL := "https://example.org"
	t.Run("success", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		sender := &piperhttp.Client{}
		sender.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		// add response handler

		httpmock.RegisterResponder(http.MethodGet, testURL+"/api/"+EndpointCeTask+"", httpmock.ResponderFromMultipleResponses(
			[]*http.Response{
				httpmock.NewStringResponse(http.StatusOK, responseCeTaskPending),
				httpmock.NewStringResponse(http.StatusOK, responseCeTaskProcessing),
				httpmock.NewStringResponse(http.StatusOK, responseCeTaskSuccess),
			},
		))
		// create service instance
		serviceUnderTest := NewTaskService(testURL, mock.Anything, mock.Anything, sender)
		serviceUnderTest.PollInterval = time.Millisecond
		// test
		err := serviceUnderTest.WaitForTask()
		// assert
		assert.NoError(t, err)
		assert.Equal(t, 3, httpmock.GetTotalCallCount(), "unexpected number of requests")
	})
	t.Run("failure", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		sender := &piperhttp.Client{}
		sender.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		// add response handler

		httpmock.RegisterResponder(http.MethodGet, testURL+"/api/"+EndpointCeTask+"", httpmock.ResponderFromMultipleResponses(
			[]*http.Response{
				httpmock.NewStringResponse(http.StatusOK, responseCeTaskPending),
				httpmock.NewStringResponse(http.StatusOK, responseCeTaskProcessing),
				httpmock.NewStringResponse(http.StatusNotFound, responseCeTaskFailure),
			},
		))
		// create service instance
		serviceUnderTest := NewTaskService(testURL, mock.Anything, mock.Anything, sender)
		serviceUnderTest.PollInterval = time.Millisecond
		// test
		err := serviceUnderTest.WaitForTask()
		// assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "status: FAILED")
		assert.Equal(t, 3, httpmock.GetTotalCallCount(), "unexpected number of requests")
	})
}

const responseCeTaskError = `{
    "errors": [
        {
            "msg": "No activity found for task 'AXDj5ZWQ_ZJrW2xGuBWl'"
        }
    ]
}`

const responseCeTaskPending = `{
    "task": {
        "id": "AXe5y_ZMPMpzvP5DRxw_",
        "type": "REPORT",
        "componentId": "AW8jANn5v4pDRYwyZIiM",
        "componentKey": "Piper-Validation/Golang",
        "componentName": "Piper-Validation: Golang",
        "componentQualifier": "TRK",
        "analysisId": "AXe5y_mgcqEbAZBpFc0V",
        "status": "PENDING",
        "submittedAt": "2021-02-19T10:18:07+0000",
        "submitterLogin": "CCFenner",
        "startedAt": "2021-02-19T10:18:08+0000",
        "executedAt": "2021-02-19T10:18:09+0000",
        "executionTimeMs": 551,
        "logs": false,
        "hasScannerContext": true,
        "organization": "default-organization",
        "warningCount": 1,
        "warnings": []
    }
}`

const responseCeTaskProcessing = `{
    "task": {
        "id": "AXe5y_ZMPMpzvP5DRxw_",
        "type": "REPORT",
        "componentId": "AW8jANn5v4pDRYwyZIiM",
        "componentKey": "Piper-Validation/Golang",
        "componentName": "Piper-Validation: Golang",
        "componentQualifier": "TRK",
        "analysisId": "AXe5y_mgcqEbAZBpFc0V",
        "status": "IN_PROGRESS",
        "submittedAt": "2021-02-19T10:18:07+0000",
        "submitterLogin": "CCFenner",
        "startedAt": "2021-02-19T10:18:08+0000",
        "executedAt": "2021-02-19T10:18:09+0000",
        "executionTimeMs": 551,
        "logs": false,
        "hasScannerContext": true,
        "organization": "default-organization",
        "warningCount": 1,
        "warnings": []
    }
}`

const responseCeTaskSuccess = `{
    "task": {
        "id": "AXe5y_ZMPMpzvP5DRxw_",
        "type": "REPORT",
        "componentId": "AW8jANn5v4pDRYwyZIiM",
        "componentKey": "Piper-Validation/Golang",
        "componentName": "Piper-Validation: Golang",
        "componentQualifier": "TRK",
        "analysisId": "AXe5y_mgcqEbAZBpFc0V",
        "status": "SUCCESS",
        "submittedAt": "2021-02-19T10:18:07+0000",
        "submitterLogin": "CCFenner",
        "startedAt": "2021-02-19T10:18:08+0000",
        "executedAt": "2021-02-19T10:18:09+0000",
        "executionTimeMs": 551,
        "logs": false,
        "hasScannerContext": true,
        "organization": "default-organization",
        "warningCount": 1,
        "warnings": [
            "The project key ‘Piper-Validation/Golang’ contains invalid characters. Allowed characters are alphanumeric, '-', '_', '.' and ':', with at least one non-digit. You should update the project key with the expected format."
        ]
    }
}`

const responseCeTaskFailure = `{
    "task": {
        "organization": "my-org-1",
        "id": "AVAn5RKqYwETbXvgas-I",
        "type": "REPORT",
        "componentId": "AVAn5RJmYwETbXvgas-H",
        "componentKey": "project_1",
        "componentName": "Project One",
        "componentQualifier": "TRK",
        "analysisId": "123456",
        "status": "FAILED",
        "submittedAt": "2015-10-02T11:32:15+0200",
        "startedAt": "2015-10-02T11:32:16+0200",
        "executedAt": "2015-10-02T11:32:22+0200",
        "executionTimeMs": 5286,
        "errorMessage": "Fail to extract report AVaXuGAi_te3Ldc_YItm from database",
        "logs": false,
        "hasErrorStacktrace": true,
        "errorStacktrace": "java.lang.IllegalStateException: Fail to extract report AVaXuGAi_te3Ldc_YItm from database\n\tat org.sonar.server.computation.task.projectanalysis.step.ExtractReportStep.execute(ExtractReportStep.java:50)",
        "scannerContext": "SonarQube plugins:\n\t- Git 1.0 (scmgit)\n\t- Java 3.13.1 (java)",
        "hasScannerContext": true
    }
}`
