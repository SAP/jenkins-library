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
	t.Run("Code Coverage: success", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		sender := &piperhttp.Client{}
		sender.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		// add response handler
		httpmock.RegisterResponder(http.MethodGet, testURL+"/api/"+EndpointMeasuresComponent+"", httpmock.NewStringResponder(http.StatusOK, responseCoverage))
		// create service instance
		serviceUnderTest := NewMeasuresComponentService(testURL, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, sender)
		// test
		cov, err := serviceUnderTest.GetCoverage()
		// assert
		assert.NoError(t, err)
		assert.Equal(t, float32(80.7), cov.Coverage)
		assert.Equal(t, float32(80.4), cov.LineCoverage)
		assert.Equal(t, 121, cov.LinesToCover)
		assert.Equal(t, 91, cov.UncoveredLines)
		assert.Equal(t, float32(81), cov.BranchCoverage)
		assert.Equal(t, 8, cov.BranchesToCover)
		assert.Equal(t, 5, cov.UncoveredBranches)
		assert.Equal(t, 1, httpmock.GetTotalCallCount(), "unexpected number of requests")
	})
	t.Run("Code Coverage: invalid metric value", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		sender := &piperhttp.Client{}
		sender.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		// add response handler
		httpmock.RegisterResponder(http.MethodGet, testURL+"/api/"+EndpointMeasuresComponent+"", httpmock.NewStringResponder(http.StatusOK, responseCoverageInvalidValue))
		// create service instance
		serviceUnderTest := NewMeasuresComponentService(testURL, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, sender)
		// test
		cov, err := serviceUnderTest.GetCoverage()
		// assert
		assert.Error(t, err)
		assert.Nil(t, cov)
		assert.Equal(t, 1, httpmock.GetTotalCallCount(), "unexpected number of requests")
	})
	t.Run("Lines Of Code: success", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		sender := &piperhttp.Client{}
		sender.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		// add response handler
		httpmock.RegisterResponder(http.MethodGet, testURL+"/api/"+EndpointMeasuresComponent+"", httpmock.NewStringResponder(http.StatusOK, responseLinesOfCode))
		// create service instance
		serviceUnderTest := NewMeasuresComponentService(testURL, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, sender)
		// test
		loc, err := serviceUnderTest.GetLinesOfCode()
		// assert
		assert.NoError(t, err)
		assert.Equal(t, 19464, loc.Total)

		for _, dist := range loc.LanguageDistribution {

			switch dist.LanguageKey {
			case "js":
				assert.Equal(t, 1504, dist.LinesOfCode)
			case "ts":
				assert.Equal(t, 16623, dist.LinesOfCode)
			case "web":
				assert.Equal(t, 1337, dist.LinesOfCode)
			}

		}

		assert.Equal(t, 1, httpmock.GetTotalCallCount(), "unexpected number of requests")
	})
	t.Run("Lines Of Code: invalid metric value", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		sender := &piperhttp.Client{}
		sender.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		// add response handler
		httpmock.RegisterResponder(http.MethodGet, testURL+"/api/"+EndpointMeasuresComponent+"", httpmock.NewStringResponder(http.StatusOK, responseLinesOfCodeInvalidValue))
		// create service instance
		serviceUnderTest := NewMeasuresComponentService(testURL, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, sender)
		// test
		loc, err := serviceUnderTest.GetLinesOfCode()
		// assert
		assert.Error(t, err)
		assert.Nil(t, loc)
		assert.Equal(t, 1, httpmock.GetTotalCallCount(), "unexpected number of requests")
	})
	t.Run("Lines Of Code: no separator", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		sender := &piperhttp.Client{}
		sender.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		// add response handler
		httpmock.RegisterResponder(http.MethodGet, testURL+"/api/"+EndpointMeasuresComponent+"", httpmock.NewStringResponder(http.StatusOK, responseLinesOfCodeNoSeparator))
		// create service instance
		serviceUnderTest := NewMeasuresComponentService(testURL, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, sender)
		// test
		loc, err := serviceUnderTest.GetLinesOfCode()
		// assert
		assert.Error(t, err)
		assert.Nil(t, loc)
		assert.Equal(t, 1, httpmock.GetTotalCallCount(), "unexpected number of requests")
	})
}

const responseCoverage = `{
	"component": {
		"key": "com.sap.piper.test",
		"name": "com.sap.piper.test",
		"qualifier": "TRK",
		"measures": [
			{ "metric": "line_coverage", "value": "80.4", "bestValue": false },
			{ "metric": "branch_coverage", "value": "81.0", "bestValue": false },
			{ "metric": "coverage", "value": "80.7", "bestValue": false },
			{ "metric": "extra_valie", "value": "42.7", "bestValue": false },
			{ "metric": "lines_to_cover", "value": "121" },
			{ "metric": "uncovered_lines", "value": "91", "bestValue": false },
			{ "metric": "conditions_to_cover", "value": "8" },
			{ "metric": "uncovered_conditions", "value": "5", "bestValue": false }
		]
	}
}`

const responseCoverageInvalidValue = `{
	"component": {
	  "key": "com.sap.piper.test",
	  "name": "com.sap.piper.test",
	  "qualifier": "TRK",
	  "measures": [
		  { "metric": "line_coverage", "value": "xyz", "bestValue": false },
		  { "metric": "uncovered_conditions", "value": "abc", "bestValue": false }
	  ]
	}
  }`

const responseLinesOfCode = `{
	"component": {
		"key": "com.sap.piper.test",
		"name": "com.sap.piper.test",
		"qualifier": "TRK",
		"measures": [
			{ "metric": "ncloc_language_distribution", "value": "js=1504;ts=16623;web=1337" },
			{ "metric": "ncloc", "value": "19464" }
		]
	}
}`

const responseLinesOfCodeInvalidValue = `{
	"component": {
	  "key": "com.sap.piper.test",
	  "name": "com.sap.piper.test",
	  "qualifier": "TRK",
	  "measures": [
		{ "metric": "ncloc_language_distribution", "value": "js=15.04;ts=16623;web=1337" },
		{ "metric": "ncloc", "value": "19464" }
	]
	}
  }`

const responseLinesOfCodeNoSeparator = `{
	"component": {
	  "key": "com.sap.piper.test",
	  "name": "com.sap.piper.test",
	  "qualifier": "TRK",
	  "measures": [
		{ "metric": "ncloc_language_distribution", "value": "js15.04" },
		{ "metric": "ncloc", "value": "19464" }
	]
	}
  }`
