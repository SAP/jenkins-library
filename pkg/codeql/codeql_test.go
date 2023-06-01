//go:build unit
// +build unit

package codeql

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-github/v45/github"
	"github.com/stretchr/testify/assert"
)

type githubCodeqlScanningMock struct {
}

func (g *githubCodeqlScanningMock) ListAlertsForRepo(ctx context.Context, owner, repo string, opts *github.AlertListOptions) ([]*github.Alert, *github.Response, error) {
	openState := "open"
	closedState := "closed"
	alerts := []*github.Alert{}

	if repo == "testRepo1" {
		alerts = append(alerts, &github.Alert{State: &openState})
		alerts = append(alerts, &github.Alert{State: &openState})
		alerts = append(alerts, &github.Alert{State: &closedState})
	}

	if repo == "testRepo2" {
		if opts.Page == 1 {
			for i := 0; i < 50; i++ {
				alerts = append(alerts, &github.Alert{State: &openState})
			}
			for i := 0; i < 50; i++ {
				alerts = append(alerts, &github.Alert{State: &closedState})
			}
		}

		if opts.Page == 2 {
			for i := 0; i < 10; i++ {
				alerts = append(alerts, &github.Alert{State: &openState})
			}
			for i := 0; i < 30; i++ {
				alerts = append(alerts, &github.Alert{State: &closedState})
			}
		}
	}

	return alerts, nil, nil
}

func (g *githubCodeqlScanningMock) ListAnalysesForRepo(ctx context.Context, owner, repo string, opts *github.AnalysesListOptions) ([]*github.ScanningAnalysis, *github.Response, error) {
	resultsCount := 3
	analysis := []*github.ScanningAnalysis{{ResultsCount: &resultsCount}}
	return analysis, nil, nil
}

type githubCodeqlScanningErrorMock struct {
}

func (g *githubCodeqlScanningErrorMock) ListAlertsForRepo(ctx context.Context, owner, repo string, opts *github.AlertListOptions) ([]*github.Alert, *github.Response, error) {
	return []*github.Alert{}, nil, errors.New("Some error")
}

func (g *githubCodeqlScanningErrorMock) ListAnalysesForRepo(ctx context.Context, owner, repo string, opts *github.AnalysesListOptions) ([]*github.ScanningAnalysis, *github.Response, error) {
	return []*github.ScanningAnalysis{}, nil, errors.New("Some error")
}

func TestGetVulnerabilitiesFromClient(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	t.Run("Success", func(t *testing.T) {
		ghCodeqlScanningMock := githubCodeqlScanningMock{}
		totalAlerts := 3
		codeqlScanAuditInstance := NewCodeqlScanAuditInstance("", "", "testRepo1", "", []string{})
		codeScanning, err := getVulnerabilitiesFromClient(ctx, &ghCodeqlScanningMock, "ref", &codeqlScanAuditInstance, totalAlerts)
		assert.NoError(t, err)
		assert.Equal(t, 3, codeScanning.Total)
		assert.Equal(t, 1, codeScanning.Audited)
	})

	t.Run("Success with pagination results", func(t *testing.T) {
		ghCodeqlScanningMock := githubCodeqlScanningMock{}
		totalAlerts := 120
		codeqlScanAuditInstance := NewCodeqlScanAuditInstance("", "", "testRepo2", "", []string{})
		codeScanning, err := getVulnerabilitiesFromClient(ctx, &ghCodeqlScanningMock, "ref", &codeqlScanAuditInstance, totalAlerts)
		assert.NoError(t, err)
		assert.Equal(t, 120, codeScanning.Total)
		assert.Equal(t, 80, codeScanning.Audited)
	})

	t.Run("Error", func(t *testing.T) {
		ghCodeqlScanningErrorMock := githubCodeqlScanningErrorMock{}
		totalAlerts := 3
		codeqlScanAuditInstance := NewCodeqlScanAuditInstance("", "", "", "", []string{})
		_, err := getVulnerabilitiesFromClient(ctx, &ghCodeqlScanningErrorMock, "ref", &codeqlScanAuditInstance, totalAlerts)
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

func TestGetTotalAnalysesFromClient(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	t.Run("Success", func(t *testing.T) {
		ghCodeqlScanningMock := githubCodeqlScanningMock{}
		codeqlScanAuditInstance := NewCodeqlScanAuditInstance("", "", "", "", []string{})
		total, err := getTotalAlertsFromClient(ctx, &ghCodeqlScanningMock, "ref", &codeqlScanAuditInstance)
		assert.NoError(t, err)
		assert.Equal(t, 3, total)
	})

	t.Run("Error", func(t *testing.T) {
		ghCodeqlScanningErrorMock := githubCodeqlScanningErrorMock{}
		codeqlScanAuditInstance := NewCodeqlScanAuditInstance("", "", "", "", []string{})
		_, err := getTotalAlertsFromClient(ctx, &ghCodeqlScanningErrorMock, "ref", &codeqlScanAuditInstance)
		assert.Error(t, err)
	})
}
