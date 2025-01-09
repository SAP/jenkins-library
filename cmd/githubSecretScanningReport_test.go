package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/google/go-github/v45/github"

	piperMock "github.com/SAP/jenkins-library/pkg/mock"
)

var fakeGithubURL = "https://github.local/myorg/myrepository"

type githubSecretScanningServiceMock struct {
	mock.Mock
}

func (gssm *githubSecretScanningServiceMock) GetRepo(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error) {
	args := gssm.Called(ctx, owner, repo)
	return args.Get(0).(*github.Repository), args.Get(1).(*github.Response), args.Error(2)
}

func (gssm *githubSecretScanningServiceMock) ListAlertsForRepo(ctx context.Context, owner, repo string, opts *github.SecretScanningAlertListOptions) ([]*github.SecretScanningAlert, *github.Response, error) {
	args := gssm.Called(ctx, owner, repo, opts)
	return args.Get(0).([]*github.SecretScanningAlert), args.Get(1).(*github.Response), args.Error(2)
}

type githubSecretScanningMockUtils struct {
	*piperMock.ExecMockRunner
	*piperMock.FilesMock
	*githubSecretScanningServiceMock
}

func newGithubSecretScanTestsUtils() githubSecretScanningMockUtils {
	utils := githubSecretScanningMockUtils{
		ExecMockRunner:                  &piperMock.ExecMockRunner{},
		FilesMock:                       &piperMock.FilesMock{},
		githubSecretScanningServiceMock: &githubSecretScanningServiceMock{},
	}
	return utils
}

func TestRunGithubSecretScanningReport(t *testing.T) {
	t.Parallel()

	t.Run("happy path - no findings in repository", func(t *testing.T) {
		config := githubSecretScanningReportOptions{
			Owner:      "myorg",
			Repository: "myrepository",
		}

		ctx := context.Background()

		utils := newGithubSecretScanTestsUtils()

		utils.On("GetRepo", ctx, config.Owner, config.Repository).Return(&github.Repository{HTMLURL: &fakeGithubURL}, &github.Response{}, nil)
		utils.On("ListAlertsForRepo", ctx, config.Owner, config.Repository, mock.Anything).Return([]*github.SecretScanningAlert{}, &github.Response{}, nil)

		// test
		err := runGithubSecretScanningReport(ctx, &config, nil, utils)

		// assert
		assert.NoError(t, err)

		if reportRWC, err := utils.Open("github-secretscanning.report.json"); assert.NoError(t, err) {
			defer reportRWC.Close()

			var report githubSecretScanningReportType
			if err = json.NewDecoder(reportRWC).Decode(&report); assert.NoError(t, err) {
				assert.Equal(t, "GitHubSecretScanning", report.ToolName)
				assert.Equal(t, "https://github.local/myorg/myrepository", report.RepositoryURL)
				assert.Equal(t, "https://github.local/myorg/myrepository/security/secret-scanning", report.SecretScanningURL)

				if assert.Len(t, report.Findings, 1) {
					assert.Equal(t, "Audit All", report.Findings[0].ClassificationName)
					assert.Equal(t, 0, report.Findings[0].Total)
					assert.Equal(t, 0, report.Findings[0].Audited)
				}
			}
		}
	})

	t.Run("happy path - repo has findings", func(t *testing.T) {
		config := githubSecretScanningReportOptions{
			Owner:      "myorg",
			Repository: "myrepository",
		}

		ctx := context.Background()

		utils := newGithubSecretScanTestsUtils()

		resolved := "resolved"

		utils.On("GetRepo", ctx, config.Owner, config.Repository).Return(&github.Repository{HTMLURL: &fakeGithubURL}, &github.Response{}, nil)
		utils.On("ListAlertsForRepo", ctx, config.Owner, config.Repository, mock.Anything).Return([]*github.SecretScanningAlert{
			&github.SecretScanningAlert{State: &resolved},
			&github.SecretScanningAlert{},
		}, &github.Response{}, nil)

		// test
		err := runGithubSecretScanningReport(ctx, &config, nil, utils)

		// assert
		assert.NoError(t, err)

		if reportRWC, err := utils.Open("github-secretscanning.report.json"); assert.NoError(t, err) {
			defer reportRWC.Close()

			var report githubSecretScanningReportType
			if err = json.NewDecoder(reportRWC).Decode(&report); assert.NoError(t, err) {
				assert.Equal(t, "GitHubSecretScanning", report.ToolName)
				assert.Equal(t, "https://github.local/myorg/myrepository", report.RepositoryURL)
				assert.Equal(t, "https://github.local/myorg/myrepository/security/secret-scanning", report.SecretScanningURL)

				if assert.Len(t, report.Findings, 1) {
					assert.Equal(t, "Audit All", report.Findings[0].ClassificationName)
					assert.Equal(t, 2, report.Findings[0].Total)
					assert.Equal(t, 1, report.Findings[0].Audited)
				}
			}
		}
	})

	t.Run("error path - simulate repo not found", func(t *testing.T) {
		config := githubSecretScanningReportOptions{
			Owner:      "myorg",
			Repository: "myrepository",
		}
		ctx := context.Background()
		utils := newGithubSecretScanTestsUtils()

		utils.On("GetRepo", ctx, config.Owner, config.Repository).Return(&github.Repository{}, &github.Response{}, fmt.Errorf("not found"))

		// test
		err := runGithubSecretScanningReport(ctx, &config, nil, utils)

		// assert
		assert.EqualError(t, err, "couldn't generate the github secret scanning report: not found")
	})
}
