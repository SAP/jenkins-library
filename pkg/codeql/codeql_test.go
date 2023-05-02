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
		codeqlScanAuditInstance := NewCodeqlScanAuditInstance("", "", "", "", []string{})
		codeScanning, err := getVulnerabilitiesFromClient(ctx, &ghCodeqlScanningMock, "ref", &codeqlScanAuditInstance)
		assert.NoError(t, err)
		assert.Equal(t, 3, codeScanning.Total)
		assert.Equal(t, 1, codeScanning.Audited)
	})

	t.Run("Error", func(t *testing.T) {
		ghCodeqlScanningErrorMock := githubCodeqlScanningErrorMock{}
		codeqlScanAuditInstance := NewCodeqlScanAuditInstance("", "", "", "", []string{})
		_, err := getVulnerabilitiesFromClient(ctx, &ghCodeqlScanningErrorMock, "ref", &codeqlScanAuditInstance)
		assert.Error(t, err)
	})
}
