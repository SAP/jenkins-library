//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestSonarIntegration ./integration/...

package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/sonar"
)

func TestSonarIntegrationIssueSearch(t *testing.T) {
	// t.Parallel()
	// init
	token := os.Getenv("PIPER_INTEGRATION_SONAR_TOKEN")
	require.NotEmpty(t, token, "SonarQube API Token is missing")
	host := "https://sonarcloud.io"
	organization := "sap-1"
	componentKey := "SAP_jenkins-library"
	options := &sonar.IssuesSearchOption{
		ComponentKeys: componentKey,
		Severities:    "MINOR",
		Resolved:      "false",
		Ps:            "1",
		Organization:  organization,
	}
	issueService := sonar.NewIssuesService(host, token, componentKey, organization, "", "", &piperhttp.Client{})
	// test
	result, _, err := issueService.SearchIssues(options)
	// assert
	assert.NoError(t, err)
	assert.NotEmpty(t, result.Components)
	//FIXME: include once implememnted
	// assert.NotEmpty(t, result.Organizations)
}

func TestSonarIntegrationMeasuresComponentSearch(t *testing.T) {
	// t.Parallel()
	// init
	token := os.Getenv("PIPER_INTEGRATION_SONAR_TOKEN")
	require.NotEmpty(t, token, "SonarQube API Token is missing")
	host := "https://sonarcloud.io"
	organization := "sap-1"
	componentKey := "SAP_jenkins-library"

	componentService := sonar.NewMeasuresComponentService(host, token, componentKey, organization, "", "", &piperhttp.Client{})
	// test
	_, err := componentService.GetCoverage()
	// assert
	assert.NoError(t, err)
}

func TestSonarIntegrationGetLinesOfCode(t *testing.T) {
	// t.Parallel()
	// init
	token := os.Getenv("PIPER_INTEGRATION_SONAR_TOKEN")
	require.NotEmpty(t, token, "SonarQube API Token is missing")
	host := "https://sonarcloud.io"
	organization := "sap-1"
	componentKey := "SAP_jenkins-library"

	componentService := sonar.NewMeasuresComponentService(host, token, componentKey, organization, "", "", &piperhttp.Client{})
	// test
	_, err := componentService.GetLinesOfCode()
	// assert
	assert.NoError(t, err)
}
