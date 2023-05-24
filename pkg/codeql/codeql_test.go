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
	alerts := []*github.Alert{{State: &openState}, {State: &openState}, {State: &closedState}}
	return alerts, nil, nil
}

func (g *githubCodeqlScanningMock) ListAnalysesForRepo(ctx context.Context, owner, repo string, opts *github.AnalysesListOptions) ([]*github.ScanningAnalysis, *github.Response, error) {
	analysis := []*github.ScanningAnalysis{{ResultsCount: 3}}
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
		codeqlScanAuditInstance := NewCodeqlScanAuditInstance("", "", "", "", []string{})
		codeScanning, err := getVulnerabilitiesFromClient(ctx, &ghCodeqlScanningMock, "ref", &codeqlScanAuditInstance, totalAlerts)
		assert.NoError(t, err)
		assert.Equal(t, 3, codeScanning.Total)
		assert.Equal(t, 1, codeScanning.Audited)
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

func TestgetTotalAnalysesFromClient(t *testing.T) {
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
