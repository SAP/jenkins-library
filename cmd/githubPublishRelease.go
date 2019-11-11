package cmd

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/google/go-github/v28/github"
	"github.com/pkg/errors"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
)

type githubRepoClient interface {
	CreateRelease(ctx context.Context, owner string, repo string, release *github.RepositoryRelease) (*github.RepositoryRelease, *github.Response, error)
	DeleteReleaseAsset(ctx context.Context, owner string, repo string, id int64) (*github.Response, error)
	GetLatestRelease(ctx context.Context, owner string, repo string) (*github.RepositoryRelease, *github.Response, error)
	ListReleaseAssets(ctx context.Context, owner string, repo string, id int64, opt *github.ListOptions) ([]*github.ReleaseAsset, *github.Response, error)
	UploadReleaseAsset(ctx context.Context, owner string, repo string, id int64, opt *github.UploadOptions, file *os.File) (*github.ReleaseAsset, *github.Response, error)
}

type githubIssueClient interface {
	ListByRepo(ctx context.Context, owner string, repo string, opt *github.IssueListByRepoOptions) ([]*github.Issue, *github.Response, error)
}

func githubPublishRelease(myGithubPublishReleaseOptions githubPublishReleaseOptions) error {
	ctx, client, err := piperGithub.NewClient(myGithubPublishReleaseOptions.Token, myGithubPublishReleaseOptions.APIURL, myGithubPublishReleaseOptions.UploadURL)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to get GitHub client.")
	}

	err = runGithubPublishRelease(ctx, &myGithubPublishReleaseOptions, client.Repositories, client.Issues)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to publish GitHub release.")
	}

	return nil
}

func runGithubPublishRelease(ctx context.Context, myGithubPublishReleaseOptions *githubPublishReleaseOptions, ghRepoClient githubRepoClient, ghIssueClient githubIssueClient) error {

	var publishedAt github.Timestamp

	lastRelease, resp, err := ghRepoClient.GetLatestRelease(ctx, myGithubPublishReleaseOptions.Owner, myGithubPublishReleaseOptions.Repository)
	if err != nil {
		if resp.StatusCode == 404 {
			//no previous release found -> first release
			myGithubPublishReleaseOptions.AddDeltaToLastRelease = false
			log.Entry().Debug("This is the first release.")
		} else {
			return errors.Wrap(err, "Error occured when retrieving latest GitHub release.")
		}
	}
	publishedAt = lastRelease.GetPublishedAt()
	log.Entry().Debugf("Previous GitHub release published: '%v'", publishedAt)

	//updating assets only supported on latest release
	if myGithubPublishReleaseOptions.UpdateAsset && myGithubPublishReleaseOptions.Version == "latest" {
		return uploadReleaseAsset(ctx, lastRelease.GetID(), myGithubPublishReleaseOptions, ghRepoClient)
	}

	releaseBody := ""

	if len(myGithubPublishReleaseOptions.ReleaseBodyHeader) > 0 {
		releaseBody += myGithubPublishReleaseOptions.ReleaseBodyHeader + "\n"
	}

	if myGithubPublishReleaseOptions.AddClosedIssues {
		releaseBody += getClosedIssuesText(ctx, publishedAt, myGithubPublishReleaseOptions, ghIssueClient)
	}

	if myGithubPublishReleaseOptions.AddDeltaToLastRelease {
		releaseBody += getReleaseDeltaText(myGithubPublishReleaseOptions, lastRelease)
	}

	release := github.RepositoryRelease{
		TagName:         &myGithubPublishReleaseOptions.Version,
		TargetCommitish: &myGithubPublishReleaseOptions.Commitish,
		Name:            &myGithubPublishReleaseOptions.Version,
		Body:            &releaseBody,
	}

	createdRelease, _, err := ghRepoClient.CreateRelease(ctx, myGithubPublishReleaseOptions.Owner, myGithubPublishReleaseOptions.Repository, &release)
	if err != nil {
		return errors.Wrapf(err, "Creation of release '%v' failed", *release.TagName)
	}
	log.Entry().Infof("Release %v created on %v/%v", *createdRelease.TagName, myGithubPublishReleaseOptions.Owner, myGithubPublishReleaseOptions.Repository)

	if len(myGithubPublishReleaseOptions.AssetPath) > 0 {
		return uploadReleaseAsset(ctx, createdRelease.GetID(), myGithubPublishReleaseOptions, ghRepoClient)
	}

	return nil
}

func getClosedIssuesText(ctx context.Context, publishedAt github.Timestamp, myGithubPublishReleaseOptions *githubPublishReleaseOptions, ghIssueClient githubIssueClient) string {
	closedIssuesText := ""

	options := github.IssueListByRepoOptions{
		State:     "closed",
		Direction: "asc",
		Since:     publishedAt.Time,
	}
	if len(myGithubPublishReleaseOptions.Labels) > 0 {
		options.Labels = myGithubPublishReleaseOptions.Labels
	}
	ghIssues, _, err := ghIssueClient.ListByRepo(ctx, myGithubPublishReleaseOptions.Owner, myGithubPublishReleaseOptions.Repository, &options)
	if err != nil {
		log.Entry().WithError(err).Error("Failed to get GitHub issues.")
	}

	prTexts := []string{"**List of closed pull-requests since last release**"}
	issueTexts := []string{"**List of closed issues since last release**"}

	for _, issue := range ghIssues {
		if issue.IsPullRequest() && !isExcluded(issue, myGithubPublishReleaseOptions.ExcludeLabels) {
			prTexts = append(prTexts, fmt.Sprintf("[#%v](%v): %v", issue.GetNumber(), issue.GetHTMLURL(), issue.GetTitle()))
			log.Entry().Debugf("Added PR #%v to release", issue.GetNumber())
		} else if !issue.IsPullRequest() && !isExcluded(issue, myGithubPublishReleaseOptions.ExcludeLabels) {
			issueTexts = append(issueTexts, fmt.Sprintf("[#%v](%v): %v", issue.GetNumber(), issue.GetHTMLURL(), issue.GetTitle()))
			log.Entry().Debugf("Added Issue #%v to release", issue.GetNumber())
		}
	}

	if len(prTexts) > 1 {
		closedIssuesText += "\n" + strings.Join(prTexts, "\n") + "\n"
	}

	if len(issueTexts) > 1 {
		closedIssuesText += "\n" + strings.Join(issueTexts, "\n") + "\n"
	}
	return closedIssuesText
}

func getReleaseDeltaText(myGithubPublishReleaseOptions *githubPublishReleaseOptions, lastRelease *github.RepositoryRelease) string {
	releaseDeltaText := ""

	//add delta link to previous release
	releaseDeltaText += "\n**Changes**\n"
	releaseDeltaText += fmt.Sprintf(
		"[%v...%v](%v/%v/%v/compare/%v...%v)\n",
		lastRelease.GetTagName(),
		myGithubPublishReleaseOptions.Version,
		myGithubPublishReleaseOptions.ServerURL,
		myGithubPublishReleaseOptions.Owner,
		myGithubPublishReleaseOptions.Repository,
		lastRelease.GetTagName(), myGithubPublishReleaseOptions.Version,
	)

	return releaseDeltaText
}

func uploadReleaseAsset(ctx context.Context, releaseID int64, myGithubPublishReleaseOptions *githubPublishReleaseOptions, ghRepoClient githubRepoClient) error {

	assets, _, err := ghRepoClient.ListReleaseAssets(ctx, myGithubPublishReleaseOptions.Owner, myGithubPublishReleaseOptions.Repository, releaseID, &github.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "Failed to get list of release assets.")
	}
	var assetID int64
	for _, a := range assets {
		if a.GetName() == filepath.Base(myGithubPublishReleaseOptions.AssetPath) {
			assetID = a.GetID()
			break
		}
	}
	if assetID != 0 {
		//asset needs to be deleted first since API does not allow for replacement
		_, err := ghRepoClient.DeleteReleaseAsset(ctx, myGithubPublishReleaseOptions.Owner, myGithubPublishReleaseOptions.Repository, assetID)
		if err != nil {
			return errors.Wrap(err, "Failed to delete release asset.")
		}
	}

	mediaType := mime.TypeByExtension(filepath.Ext(myGithubPublishReleaseOptions.AssetPath))
	if mediaType == "" {
		mediaType = "application/octet-stream"
	}
	log.Entry().Debugf("Using mediaType '%v'", mediaType)

	name := filepath.Base(myGithubPublishReleaseOptions.AssetPath)
	log.Entry().Debugf("Using file name '%v'", name)

	opts := github.UploadOptions{
		Name:      name,
		MediaType: mediaType,
	}
	file, err := os.Open(myGithubPublishReleaseOptions.AssetPath)
	defer file.Close()
	if err != nil {
		return errors.Wrapf(err, "Failed to load release asset '%v'", myGithubPublishReleaseOptions.AssetPath)
	}

	log.Entry().Info("Starting to upload release asset.")
	asset, _, err := ghRepoClient.UploadReleaseAsset(ctx, myGithubPublishReleaseOptions.Owner, myGithubPublishReleaseOptions.Repository, releaseID, &opts, file)
	if err != nil {
		return errors.Wrap(err, "Failed to upload release asset.")
	}
	log.Entry().Infof("Done uploading asset '%v'.", asset.GetURL())

	return nil
}

func isExcluded(issue *github.Issue, excludeLabels []string) bool {
	//issue.Labels[0].GetName()
	for _, ex := range excludeLabels {
		for _, l := range issue.Labels {
			if ex == l.GetName() {
				return true
			}
		}
	}
	return false
}
