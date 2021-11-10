package sonar

import (
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
)

func TestComponentService(t *testing.T) {
	testURL := "https://example.org"
	t.Run("success", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		sender := &piperhttp.Client{}
		sender.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		// add response handler
		httpmock.RegisterResponder(http.MethodGet, testURL+"/api/"+EndpointMeasuresComponent+"", httpmock.NewStringResponder(http.StatusOK, responseCoverage))
		// create service instance
		serviceUnderTest := NewMeasuresComponentService(testURL, mock.Anything, mock.Anything, mock.Anything, sender)
		// test
		cov, err := serviceUnderTest.GetCoverage()
		// assert
		assert.NoError(t, err)
		assert.Equal(t, float32(81), cov.BranchCoverage)
		assert.Equal(t, 1, httpmock.GetTotalCallCount(), "unexpected number of requests")
	})
	t.Run("invalid metric value", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		sender := &piperhttp.Client{}
		sender.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		// add response handler
		httpmock.RegisterResponder(http.MethodGet, testURL+"/api/"+EndpointMeasuresComponent+"", httpmock.NewStringResponder(http.StatusOK, responseCoverageInvalidValue))
		// create service instance
		serviceUnderTest := NewMeasuresComponentService(testURL, mock.Anything, mock.Anything, mock.Anything, sender)
		// test
		cov, err := serviceUnderTest.GetCoverage()
		// assert
		assert.Error(t, err)
		assert.Nil(t, cov)
		assert.Equal(t, 1, httpmock.GetTotalCallCount(), "unexpected number of requests")
	})
}

const responseCoverage = `{
  "component": {
    "key": "com.sap.piper.test",
    "name": "com.sap.piper.test",
    "qualifier": "TRK",
    "measures": [
      {
        "metric": "line_coverage",
        "value": "80.4",
        "bestValue": false
      },
      {
        "metric": "branch_coverage",
        "value": "81.0",
        "bestValue": false
      },
      {
        "metric": "coverage",
        "value": "80.7",
        "bestValue": false
      },
      {
        "metric": "extra_valie",
        "value": "42.7",
        "bestValue": false
      }
    ]
  }
}`

const responseCoverageInvalidValue = `{
	"component": {
	  "key": "com.sap.piper.test",
	  "name": "com.sap.piper.test",
	  "qualifier": "TRK",
	  "measures": [
		{
		  "metric": "line_coverage",
		  "value": "xyz",
		  "bestValue": false
		}
	  ]
	}
  }`
