//go:build unit
// +build unit

package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/cmd/mocks"
	"github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type ghRCMock struct {
	createErr         error
	latestRelease     *github.RepositoryRelease
	release           *github.RepositoryRelease
	delErr            error
	delID             int64
	delOwner          string
	delRepo           string
	listErr           error
	listID            int64
	listOwner         string
	listReleaseAssets []*github.ReleaseAsset
	listRepo          string
	listOpts          *github.ListOptions
	latestStatusCode  int
	latestErr         error
	uploadID          int64
	uploadOpts        *github.UploadOptions
	uploadOwner       string
	uploadRepo        string
}

func (g *ghRCMock) CreateRelease(ctx context.Context, owner string, repo string, release *github.RepositoryRelease) (*github.RepositoryRelease, *github.Response, error) {
	g.release = release
	return release, nil, g.createErr
}

func (g *ghRCMock) DeleteReleaseAsset(ctx context.Context, owner string, repo string, id int64) (*github.Response, error) {
	g.delOwner = owner
	g.delRepo = repo
	g.delID = id
	return nil, g.delErr
}

func (g *ghRCMock) GetLatestRelease(ctx context.Context, owner string, repo string) (*github.RepositoryRelease, *github.Response, error) {
	hc := http.Response{StatusCode: 200}
	if g.latestStatusCode != 0 {
		hc.StatusCode = g.latestStatusCode
	}

	if len(owner) == 0 {
		return g.latestRelease, nil, g.latestErr
	}

	ghResp := github.Response{Response: &hc}
	return g.latestRelease, &ghResp, g.latestErr
}

func (g *ghRCMock) ListReleaseAssets(ctx context.Context, owner string, repo string, id int64, opt *github.ListOptions) ([]*github.ReleaseAsset, *github.Response, error) {
	g.listID = id
	g.listOwner = owner
	g.listRepo = repo
	g.listOpts = opt
	return g.listReleaseAssets, nil, g.listErr
}

func (g *ghRCMock) UploadReleaseAsset(ctx context.Context, owner string, repo string, id int64, opt *github.UploadOptions, file *os.File) (*github.ReleaseAsset, *github.Response, error) {
	g.uploadID = id
	g.uploadOwner = owner
	g.uploadRepo = repo
	g.uploadOpts = opt
	return nil, nil, nil
}

type ghICMock struct {
	issues        []*github.Issue
	response      github.Response
	lastPublished time.Time
	owner         string
	repo          string
	options       *github.IssueListByRepoOptions
}

func (g *ghICMock) ListByRepo(ctx context.Context, owner string, repo string, opt *github.IssueListByRepoOptions) ([]*github.Issue, *github.Response, error) {
	g.owner = owner
	g.repo = repo
	g.options = opt
	g.lastPublished = opt.Since
	return g.issues, &g.response, nil
}

func TestRunGithubPublishRelease(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - first release & no body", func(t *testing.T) {
		ghIssueClient := ghICMock{}
		ghRepoClient := ghRCMock{
			latestStatusCode: 404,
			latestErr:        fmt.Errorf("not found"),
		}

		myGithubPublishReleaseOptions := githubPublishReleaseOptions{
			AddDeltaToLastRelease: true,
			Commitish:             "master",
			Owner:                 "TEST",
			PreRelease:            true,
			Repository:            "test",
			ServerURL:             "https://github.com",
			ReleaseBodyHeader:     "Header",
			Version:               "1.0",
		}
		err := runGithubPublishRelease(ctx, &myGithubPublishReleaseOptions, &ghRepoClient, &ghIssueClient)
		assert.NoError(t, err, "Error occurred but none expected.")

		assert.Equal(t, "Header\n", ghRepoClient.release.GetBody())
		assert.Equal(t, true, ghRepoClient.release.GetPrerelease())
		assert.Equal(t, "1.0", ghRepoClient.release.GetTagName())
	})

	t.Run("Success - first release with tag prefix set & no body", func(t *testing.T) {
		ghIssueClient := ghICMock{}
		ghRepoClient := ghRCMock{
			latestStatusCode: 404,
			latestErr:        fmt.Errorf("not found"),
		}

		myGithubPublishReleaseOptions := githubPublishReleaseOptions{
			AddDeltaToLastRelease: true,
			Commitish:             "master",
			Owner:                 "TEST",
			PreRelease:            true,
			Repository:            "test",
			ServerURL:             "https://github.com",
			ReleaseBodyHeader:     "Header",
			Version:               "1.0",
			TagPrefix:             "v",
		}
		err := runGithubPublishRelease(ctx, &myGithubPublishReleaseOptions, &ghRepoClient, &ghIssueClient)
		assert.NoError(t, err, "Error occurred but none expected.")

		assert.Equal(t, "Header\n", ghRepoClient.release.GetBody())
		assert.Equal(t, true, ghRepoClient.release.GetPrerelease())
		assert.Equal(t, "v1.0", ghRepoClient.release.GetTagName())
	})

	t.Run("Success - subsequent releases & with body", func(t *testing.T) {
		lastTag := "1.0"
		lastPublishedAt := github.Timestamp{Time: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)}
		ghRepoClient := ghRCMock{
			createErr: nil,
			latestRelease: &github.RepositoryRelease{
				TagName:     &lastTag,
				PublishedAt: &lastPublishedAt,
			},
		}
		prHTMLURL := "https://github.com/TEST/test/pull/1"
		prTitle := "Pull"
		prNo := 1

		issHTMLURL := "https://github.com/TEST/test/issues/2"
		issTitle := "Issue"
		issNo := 2

		ghIssueClient := ghICMock{
			issues: []*github.Issue{
				{Number: &prNo, Title: &prTitle, HTMLURL: &prHTMLURL, PullRequestLinks: &github.PullRequestLinks{URL: &prHTMLURL}},
				{Number: &issNo, Title: &issTitle, HTMLURL: &issHTMLURL},
			},
		}
		myGithubPublishReleaseOptions := githubPublishReleaseOptions{
			AddClosedIssues:       true,
			AddDeltaToLastRelease: true,
			Commitish:             "master",
			Owner:                 "TEST",
			Repository:            "test",
			ServerURL:             "https://github.com",
			ReleaseBodyHeader:     "Header",
			Version:               "1.1",
		}
		err := runGithubPublishRelease(ctx, &myGithubPublishReleaseOptions, &ghRepoClient, &ghIssueClient)

		assert.NoError(t, err, "Error occurred but none expected.")

		assert.Equal(t, "Header\n\n**List of closed pull-requests since last release**\n[#1](https://github.com/TEST/test/pull/1): Pull\n\n**List of closed issues since last release**\n[#2](https://github.com/TEST/test/issues/2): Issue\n\n**Changes**\n[1.0...1.1](https://github.com/TEST/test/compare/1.0...1.1)\n", ghRepoClient.release.GetBody())
		assert.Equal(t, "1.1", ghRepoClient.release.GetName())
		assert.Equal(t, "1.1", ghRepoClient.release.GetTagName())
		assert.Equal(t, "master", ghRepoClient.release.GetTargetCommitish())

		assert.Equal(t, lastPublishedAt.Time, ghIssueClient.lastPublished)
	})

	t.Run("Success - update asset", func(t *testing.T) {
		var releaseID int64 = 1
		ghIssueClient := ghICMock{}
		ghRepoClient := ghRCMock{
			latestRelease: &github.RepositoryRelease{
				ID: &releaseID,
			},
		}

		myGithubPublishReleaseOptions := githubPublishReleaseOptions{
			AssetPath: filepath.Join("testdata", t.Name()+"_test.txt"),
			Version:   "latest",
		}

		err := runGithubPublishRelease(ctx, &myGithubPublishReleaseOptions, &ghRepoClient, &ghIssueClient)

		assert.NoError(t, err, "Error occurred but none expected.")

		assert.Nil(t, ghRepoClient.release)

		assert.Equal(t, releaseID, ghRepoClient.listID)
		assert.Equal(t, releaseID, ghRepoClient.uploadID)
	})

	t.Run("Error - get release", func(t *testing.T) {
		ghIssueClient := ghICMock{}
		ghRepoClient := ghRCMock{
			latestErr: fmt.Errorf("Latest release error"),
		}
		myGithubPublishReleaseOptions := githubPublishReleaseOptions{
			Owner:      "TEST",
			Repository: "test",
		}
		err := runGithubPublishRelease(ctx, &myGithubPublishReleaseOptions, &ghRepoClient, &ghIssueClient)

		assert.Equal(t, "Error occurred when retrieving latest GitHub release (TEST/test): Latest release error", fmt.Sprint(err))
	})

	t.Run("Error - get release no response", func(t *testing.T) {
		ghIssueClient := ghICMock{}
		ghRepoClient := ghRCMock{
			latestErr: fmt.Errorf("Latest release error, no response"),
		}
		myGithubPublishReleaseOptions := githubPublishReleaseOptions{
			Owner:      "",
			Repository: "test",
		}
		err := runGithubPublishRelease(ctx, &myGithubPublishReleaseOptions, &ghRepoClient, &ghIssueClient)

		assert.Equal(t, "Error occurred when retrieving latest GitHub release (/test): Latest release error, no response", fmt.Sprint(err))
	})

	t.Run("Error - create release", func(t *testing.T) {
		ghIssueClient := ghICMock{}
		ghRepoClient := ghRCMock{
			createErr: fmt.Errorf("Create release error"),
		}
		myGithubPublishReleaseOptions := githubPublishReleaseOptions{
			Version: "1.0",
		}
		err := runGithubPublishRelease(ctx, &myGithubPublishReleaseOptions, &ghRepoClient, &ghIssueClient)

		assert.Equal(t, "Creation of release '1.0' failed: Create release error", fmt.Sprint(err))
	})
}

func TestGetClosedIssuesText(t *testing.T) {
	ctx := context.Background()
	publishedAt := github.Timestamp{Time: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)}

	t.Run("No issues", func(t *testing.T) {
		ghIssueClient := ghICMock{}
		myGithubPublishReleaseOptions := githubPublishReleaseOptions{
			Version: "1.0",
		}

		res := getClosedIssuesText(ctx, publishedAt, &myGithubPublishReleaseOptions, &ghIssueClient)

		assert.Equal(t, "", res)
	})

	t.Run("All issues", func(t *testing.T) {
		ctx := context.Background()
		publishedAt := github.Timestamp{Time: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)}

		prHTMLURL := []string{"https://github.com/TEST/test/pull/1", "https://github.com/TEST/test/pull/2"}
		prTitle := []string{"Pull1", "Pull2"}
		prNo := []int{1, 2}

		issHTMLURL := []string{"https://github.com/TEST/test/issues/3", "https://github.com/TEST/test/issues/4"}
		issTitle := []string{"Issue3", "Issue4"}
		issNo := []int{3, 4}

		ghIssueClient := ghICMock{
			issues: []*github.Issue{
				{Number: &prNo[0], Title: &prTitle[0], HTMLURL: &prHTMLURL[0], PullRequestLinks: &github.PullRequestLinks{URL: &prHTMLURL[0]}},
				{Number: &prNo[1], Title: &prTitle[1], HTMLURL: &prHTMLURL[1], PullRequestLinks: &github.PullRequestLinks{URL: &prHTMLURL[1]}},
				{Number: &issNo[0], Title: &issTitle[0], HTMLURL: &issHTMLURL[0]},
				{Number: &issNo[1], Title: &issTitle[1], HTMLURL: &issHTMLURL[1]},
			},
		}

		myGithubPublishReleaseOptions := githubPublishReleaseOptions{
			Owner:      "TEST",
			Repository: "test",
		}

		res := getClosedIssuesText(ctx, publishedAt, &myGithubPublishReleaseOptions, &ghIssueClient)

		assert.Equal(t, "\n**List of closed pull-requests since last release**\n[#1](https://github.com/TEST/test/pull/1): Pull1\n[#2](https://github.com/TEST/test/pull/2): Pull2\n\n**List of closed issues since last release**\n[#3](https://github.com/TEST/test/issues/3): Issue3\n[#4](https://github.com/TEST/test/issues/4): Issue4\n", res)
		assert.Equal(t, "TEST", ghIssueClient.owner, "Owner not properly passed")
		assert.Equal(t, "test", ghIssueClient.repo, "Repo not properly passed")
		assert.Equal(t, "closed", ghIssueClient.options.State, "Issue state not properly passed")
		assert.Equal(t, "asc", ghIssueClient.options.Direction, "Sort direction not properly passed")
		assert.Equal(t, publishedAt.Time, ghIssueClient.options.Since, "PublishedAt not properly passed")
	})
}

func TestGetReleaseDeltaText(t *testing.T) {
	t.Run("test case without TagPrefix for new release", func(t *testing.T) {
		myGithubPublishReleaseOptions := githubPublishReleaseOptions{
			Owner:      "TEST",
			Repository: "test",
			ServerURL:  "https://github.com",
			Version:    "1.1",
		}
		lastTag := "1.0"
		lastRelease := github.RepositoryRelease{
			TagName: &lastTag,
		}
		res := getReleaseDeltaText(&myGithubPublishReleaseOptions, &lastRelease)
		assert.Equal(t, "\n**Changes**\n[1.0...1.1](https://github.com/TEST/test/compare/1.0...1.1)\n", res)
	})

	t.Run("test case with TagPrefix for new release", func(t *testing.T) {
		myGithubPublishReleaseOptions := githubPublishReleaseOptions{
			Owner:      "TEST",
			Repository: "test",
			ServerURL:  "https://github.com",
			Version:    "1.1",
			TagPrefix:  "release/",
		}
		lastTag := "1.0"
		lastRelease := github.RepositoryRelease{
			TagName: &lastTag,
		}
		res := getReleaseDeltaText(&myGithubPublishReleaseOptions, &lastRelease)
		assert.Equal(t, "\n**Changes**\n[1.0...release/1.1](https://github.com/TEST/test/compare/1.0...release/1.1)\n", res)
	})
}

func TestUploadReleaseAsset(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - existing asset", func(t *testing.T) {
		var releaseID int64 = 1
		assetName := "Success_-_existing_asset_test.txt"
		var assetID int64 = 11
		ghRepoClient := ghRCMock{
			latestRelease: &github.RepositoryRelease{
				ID: &releaseID,
			},
			listReleaseAssets: []*github.ReleaseAsset{
				{Name: &assetName, ID: &assetID},
			},
		}

		myGithubPublishReleaseOptions := githubPublishReleaseOptions{
			Owner:      "TEST",
			Repository: "test",
			AssetPath:  filepath.Join("testdata", t.Name()+"_test.txt"),
		}

		err := uploadReleaseAsset(ctx, releaseID, &myGithubPublishReleaseOptions, &ghRepoClient)

		assert.NoError(t, err, "Error occurred but none expected.")

		assert.Equal(t, "TEST", ghRepoClient.listOwner, "Owner not properly passed - list")
		assert.Equal(t, "test", ghRepoClient.listRepo, "Repo not properly passed - list")
		assert.Equal(t, releaseID, ghRepoClient.listID, "Relase ID not properly passed - list")

		assert.Equal(t, "TEST", ghRepoClient.delOwner, "Owner not properly passed - del")
		assert.Equal(t, "test", ghRepoClient.delRepo, "Repo not properly passed - del")
		assert.Equal(t, assetID, ghRepoClient.delID, "Relase ID not properly passed - del")

		assert.Equal(t, "TEST", ghRepoClient.uploadOwner, "Owner not properly passed - upload")
		assert.Equal(t, "test", ghRepoClient.uploadRepo, "Repo not properly passed - upload")
		assert.Equal(t, releaseID, ghRepoClient.uploadID, "Relase ID not properly passed - upload")
		assert.Equal(t, "text/plain; charset=utf-8", ghRepoClient.uploadOpts.MediaType, "Wrong MediaType passed - upload")
	})

	t.Run("Success - no asset", func(t *testing.T) {
		var releaseID int64 = 1
		assetName := "notFound"
		var assetID int64 = 11
		ghRepoClient := ghRCMock{
			latestRelease: &github.RepositoryRelease{
				ID: &releaseID,
			},
			listReleaseAssets: []*github.ReleaseAsset{
				{Name: &assetName, ID: &assetID},
			},
		}

		myGithubPublishReleaseOptions := githubPublishReleaseOptions{
			Owner:      "TEST",
			Repository: "test",
			AssetPath:  filepath.Join("testdata", t.Name()+"_test.txt"),
		}

		err := uploadReleaseAsset(ctx, releaseID, &myGithubPublishReleaseOptions, &ghRepoClient)

		assert.NoError(t, err, "Error occurred but none expected.")

		assert.Equal(t, int64(0), ghRepoClient.delID, "Relase ID should not be populated")
	})

	t.Run("Error - List Assets", func(t *testing.T) {
		var releaseID int64 = 1
		ghRepoClient := ghRCMock{
			listErr: fmt.Errorf("List Asset Error"),
		}
		myGithubPublishReleaseOptions := githubPublishReleaseOptions{}

		err := uploadReleaseAsset(ctx, releaseID, &myGithubPublishReleaseOptions, &ghRepoClient)
		assert.Equal(t, "Failed to get list of release assets.: List Asset Error", fmt.Sprint(err), "Wrong error received")
	})
}

func TestUploadReleaseAssetList(t *testing.T) {
	ctx := context.Background()
	owner := "OWNER"
	repository := "REPOSITORY"
	var releaseID int64 = 1

	t.Run("Success - multiple asset", func(t *testing.T) {
		// init
		assetURL := mock.Anything
		asset1 := filepath.Join("testdata", t.Name()+"_1_test.txt")
		asset2 := filepath.Join("testdata", t.Name()+"_2_test.txt")
		assetName1 := filepath.Base(asset1)
		assetName2 := filepath.Base(asset2)
		var assetID1 int64 = 11
		var assetID2 int64 = 12
		stepConfig := githubPublishReleaseOptions{
			Owner:         owner,
			Repository:    repository,
			AssetPathList: []string{asset1, asset2},
		}
		// mocking
		ghClient := &mocks.GithubRepoClient{}
		ghClient.Test(t)
		ghClient.
			On("ListReleaseAssets", ctx, owner, repository, releaseID, mock.AnythingOfType("*github.ListOptions")).Return(
			[]*github.ReleaseAsset{
				{Name: &assetName1, ID: &assetID1, URL: &assetURL},
				{Name: &assetName2, ID: &assetID2, URL: &assetURL},
			},
			nil,
			nil,
		).
			On("DeleteReleaseAsset", ctx, owner, repository, mock.AnythingOfType("int64")).Return(
			&github.Response{Response: &http.Response{StatusCode: 200}},
			nil,
		).
			On("UploadReleaseAsset", ctx, owner, repository, releaseID, mock.AnythingOfType("*github.UploadOptions"), mock.AnythingOfType("*os.File")).Return(
			&github.ReleaseAsset{URL: &assetURL},
			&github.Response{Response: &http.Response{StatusCode: 200}},
			nil,
		)
		// test
		err := uploadReleaseAssetList(ctx, releaseID, &stepConfig, ghClient)
		// asserts
		assert.NoError(t, err)
		ghClient.AssertExpectations(t)
	})
}

func TestIsExcluded(t *testing.T) {
	l1 := "label1"
	l2 := "label2"

	tt := []struct {
		issue         *github.Issue
		excludeLabels []string
		expected      bool
	}{
		{issue: nil, excludeLabels: nil, expected: false},
		{issue: &github.Issue{}, excludeLabels: nil, expected: false},
		{issue: &github.Issue{Labels: []*github.Label{{Name: &l1}}}, excludeLabels: nil, expected: false},
		{issue: &github.Issue{Labels: []*github.Label{{Name: &l1}}}, excludeLabels: []string{"label0"}, expected: false},
		{issue: &github.Issue{Labels: []*github.Label{{Name: &l1}}}, excludeLabels: []string{"label1"}, expected: true},
		{issue: &github.Issue{Labels: []*github.Label{{Name: &l1}, {Name: &l2}}}, excludeLabels: []string{}, expected: false},
		{issue: &github.Issue{Labels: []*github.Label{{Name: &l1}, {Name: &l2}}}, excludeLabels: []string{"label1"}, expected: true},
	}

	for k, v := range tt {
		assert.Equal(t, v.expected, isExcluded(v.issue, v.excludeLabels), fmt.Sprintf("Run %v failed", k))
	}
}
