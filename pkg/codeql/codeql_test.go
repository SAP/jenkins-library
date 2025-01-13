//go:build unit
// +build unit

package codeql

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/assert"
)

type githubCodeqlScanningMock struct {
}

func (g *githubCodeqlScanningMock) ListAlertsForRepo(ctx context.Context, owner, repo string, opts *github.AlertListOptions) ([]*github.Alert, *github.Response, error) {
	openState := "open"
	dismissedState := "dismissed"
	alerts := []*github.Alert{}
	response := github.Response{}
	codeqlToolName := "CodeQL"
	testToolName := "Test"

	if repo == "testRepo1" {
		alerts = append(alerts, &github.Alert{State: &openState, Tool: &github.Tool{Name: &codeqlToolName}, Rule: &github.Rule{Tags: []string{"security"}}})
		alerts = append(alerts, &github.Alert{State: &openState, Tool: &github.Tool{Name: &codeqlToolName}, Rule: &github.Rule{Tags: []string{"security"}}})
		alerts = append(alerts, &github.Alert{State: &dismissedState, Tool: &github.Tool{Name: &codeqlToolName}, Rule: &github.Rule{Tags: []string{"security"}}})
		alerts = append(alerts, &github.Alert{State: &dismissedState, Tool: &github.Tool{Name: &testToolName}, Rule: &github.Rule{Tags: []string{"security"}}})
		alerts = append(alerts, &github.Alert{State: &dismissedState, Tool: &github.Tool{Name: &codeqlToolName}, Rule: &github.Rule{Tags: []string{"useless_code"}}})
		alerts = append(alerts, &github.Alert{State: &dismissedState, Tool: &github.Tool{Name: &testToolName}, Rule: &github.Rule{Tags: []string{"useless_code"}}})
		response.NextPage = 0
	}

	if repo == "testRepo2" {
		if opts.Page == 1 {
			for i := 0; i < 50; i++ {
				alerts = append(alerts, &github.Alert{State: &openState, Tool: &github.Tool{Name: &codeqlToolName}, Rule: &github.Rule{Tags: []string{"security"}}})
				alerts = append(alerts, &github.Alert{State: &dismissedState, Tool: &github.Tool{Name: &testToolName}, Rule: &github.Rule{Tags: []string{"useless_code"}}})
			}
			for i := 0; i < 50; i++ {
				alerts = append(alerts, &github.Alert{State: &dismissedState, Tool: &github.Tool{Name: &codeqlToolName}, Rule: &github.Rule{Tags: []string{"security"}}})
				alerts = append(alerts, &github.Alert{State: &dismissedState, Tool: &github.Tool{Name: &testToolName}, Rule: &github.Rule{Tags: []string{"useless_code"}}})
			}
			response.NextPage = 2
		}

		if opts.Page == 2 {
			for i := 0; i < 10; i++ {
				alerts = append(alerts, &github.Alert{State: &openState, Tool: &github.Tool{Name: &codeqlToolName}, Rule: &github.Rule{Tags: []string{"security"}}})
				alerts = append(alerts, &github.Alert{State: &dismissedState, Tool: &github.Tool{Name: &testToolName}, Rule: &github.Rule{Tags: []string{"useless_code"}}})
			}
			for i := 0; i < 30; i++ {
				alerts = append(alerts, &github.Alert{State: &dismissedState, Tool: &github.Tool{Name: &codeqlToolName}, Rule: &github.Rule{Tags: []string{"security"}}})
				alerts = append(alerts, &github.Alert{State: &dismissedState, Tool: &github.Tool{Name: &testToolName}, Rule: &github.Rule{Tags: []string{"useless_code"}}})
			}
			response.NextPage = 0
		}
	}

	return alerts, &response, nil
}

type githubCodeqlScanningErrorMock struct {
}

func (g *githubCodeqlScanningErrorMock) ListAlertsForRepo(ctx context.Context, owner, repo string, opts *github.AlertListOptions) ([]*github.Alert, *github.Response, error) {
	return []*github.Alert{}, nil, errors.New("Some error")
}

func TestGetVulnerabilitiesFromClient(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	t.Run("Success", func(t *testing.T) {
		ghCodeqlScanningMock := githubCodeqlScanningMock{}
		codeqlScanAuditInstance := NewCodeqlScanAuditInstance("", "", "testRepo1", "", []string{})
		codeScanning, err := getVulnerabilitiesFromClient(ctx, &ghCodeqlScanningMock, "ref", &codeqlScanAuditInstance)
		assert.NoError(t, err)
		assert.NotEmpty(t, codeScanning)
		assert.Equal(t, 2, len(codeScanning))
		assert.Equal(t, 3, codeScanning[0].Total)
		assert.Equal(t, 1, codeScanning[0].Audited)
	})

	t.Run("Success with pagination results", func(t *testing.T) {
		ghCodeqlScanningMock := githubCodeqlScanningMock{}
		codeqlScanAuditInstance := NewCodeqlScanAuditInstance("", "", "testRepo2", "", []string{})
		codeScanning, err := getVulnerabilitiesFromClient(ctx, &ghCodeqlScanningMock, "ref", &codeqlScanAuditInstance)
		assert.NoError(t, err)
		assert.NotEmpty(t, codeScanning)
		assert.Equal(t, 2, len(codeScanning))
		assert.Equal(t, 140, codeScanning[0].Total)
		assert.Equal(t, 80, codeScanning[0].Audited)
	})

	t.Run("Error", func(t *testing.T) {
		ghCodeqlScanningErrorMock := githubCodeqlScanningErrorMock{}
		codeqlScanAuditInstance := NewCodeqlScanAuditInstance("", "", "", "", []string{})
		_, err := getVulnerabilitiesFromClient(ctx, &ghCodeqlScanningErrorMock, "ref", &codeqlScanAuditInstance)
		assert.Error(t, err)
	})
}

func TestGetApiUrl(t *testing.T) {
	t.Run("public url", func(t *testing.T) {
		assert.Equal(t, "https://api.github.com", getApiUrl("https://github.com"))
	})

	t.Run("enterprise github url", func(t *testing.T) {
		assert.Equal(t, "https://github.test.org/api/v3", getApiUrl("https://github.test.org"))
	})
}
