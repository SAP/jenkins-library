package cmd

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
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

func githubPublishRelease(config githubPublishReleaseOptions, telemetryData *telemetry.CustomData) {
	ctx, client, err := piperGithub.NewClient(config.Token, config.APIURL, config.UploadURL)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to get GitHub client.")
	}

	err = runGithubPublishRelease(ctx, &config, client.Repositories, client.Issues)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to publish GitHub release.")
	}
}

func runGithubPublishRelease(ctx context.Context, config *githubPublishReleaseOptions, ghRepoClient githubRepoClient, ghIssueClient githubIssueClient) error {

	var publishedAt github.Timestamp

	lastRelease, resp, err := ghRepoClient.GetLatestRelease(ctx, config.Owner, config.Repository)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			//no previous release found -> first release
			config.AddDeltaToLastRelease = false
			log.Entry().Debug("This is the first release.")
		} else {
			return errors.Wrapf(err, "Error occured when retrieving latest GitHub release (%v/%v)", config.Owner, config.Repository)
		}
	}
	publishedAt = lastRelease.GetPublishedAt()
	log.Entry().Debugf("Previous GitHub release published: '%v'", publishedAt)

	//updating assets only supported on latest release
	if len(config.AssetPath) > 0 && config.Version == "latest" {
		return uploadReleaseAsset(ctx, lastRelease.GetID(), config, ghRepoClient)
	}

	releaseBody := ""

	if len(config.ReleaseBodyHeader) > 0 {
		releaseBody += config.ReleaseBodyHeader + "\n"
	}

	if config.AddClosedIssues {
		releaseBody += getClosedIssuesText(ctx, publishedAt, config, ghIssueClient)
	}

	if config.AddDeltaToLastRelease {
		releaseBody += getReleaseDeltaText(config, lastRelease)
	}

	release := github.RepositoryRelease{
		TagName:         &config.Version,
		TargetCommitish: &config.Commitish,
		Name:            &config.Version,
		Body:            &releaseBody,
	}

	createdRelease, _, err := ghRepoClient.CreateRelease(ctx, config.Owner, config.Repository, &release)
	if err != nil {
		return errors.Wrapf(err, "Creation of release '%v' failed", *release.TagName)
	}
	log.Entry().Infof("Release %v created on %v/%v", *createdRelease.TagName, config.Owner, config.Repository)

	if len(config.AssetPath) > 0 {
		return uploadReleaseAsset(ctx, createdRelease.GetID(), config, ghRepoClient)
	}

	return nil
}

func getClosedIssuesText(ctx context.Context, publishedAt github.Timestamp, config *githubPublishReleaseOptions, ghIssueClient githubIssueClient) string {
	closedIssuesText := ""

	options := github.IssueListByRepoOptions{
		State:     "closed",
		Direction: "asc",
		Since:     publishedAt.Time,
	}
	if len(config.Labels) > 0 {
		options.Labels = config.Labels
	}
	ghIssues, _, err := ghIssueClient.ListByRepo(ctx, config.Owner, config.Repository, &options)
	if err != nil {
		log.Entry().WithError(err).Error("Failed to get GitHub issues.")
	}

	prTexts := []string{"**List of closed pull-requests since last release**"}
	issueTexts := []string{"**List of closed issues since last release**"}

	for _, issue := range ghIssues {
		if issue.IsPullRequest() && !isExcluded(issue, config.ExcludeLabels) {
			prTexts = append(prTexts, fmt.Sprintf("[#%v](%v): %v", issue.GetNumber(), issue.GetHTMLURL(), issue.GetTitle()))
			log.Entry().Debugf("Added PR #%v to release", issue.GetNumber())
		} else if !issue.IsPullRequest() && !isExcluded(issue, config.ExcludeLabels) {
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

func getReleaseDeltaText(config *githubPublishReleaseOptions, lastRelease *github.RepositoryRelease) string {
	releaseDeltaText := ""

	//add delta link to previous release
	releaseDeltaText += "\n**Changes**\n"
	releaseDeltaText += fmt.Sprintf(
		"[%v...%v](%v/%v/%v/compare/%v...%v)\n",
		lastRelease.GetTagName(),
		config.Version,
		config.ServerURL,
		config.Owner,
		config.Repository,
		lastRelease.GetTagName(), config.Version,
	)

	return releaseDeltaText
}

func uploadReleaseAsset(ctx context.Context, releaseID int64, config *githubPublishReleaseOptions, ghRepoClient githubRepoClient) error {

	assets, _, err := ghRepoClient.ListReleaseAssets(ctx, config.Owner, config.Repository, releaseID, &github.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "Failed to get list of release assets.")
	}
	var assetID int64
	for _, a := range assets {
		if a.GetName() == filepath.Base(config.AssetPath) {
			assetID = a.GetID()
			break
		}
	}
	if assetID != 0 {
		//asset needs to be deleted first since API does not allow for replacement
		_, err := ghRepoClient.DeleteReleaseAsset(ctx, config.Owner, config.Repository, assetID)
		if err != nil {
			return errors.Wrap(err, "Failed to delete release asset.")
		}
	}

	mediaType := mime.TypeByExtension(filepath.Ext(config.AssetPath))
	if mediaType == "" {
		mediaType = "application/octet-stream"
	}
	log.Entry().Debugf("Using mediaType '%v'", mediaType)

	name := filepath.Base(config.AssetPath)
	log.Entry().Debugf("Using file name '%v'", name)

	opts := github.UploadOptions{
		Name:      name,
		MediaType: mediaType,
	}
	file, err := os.Open(config.AssetPath)
	defer file.Close()
	if err != nil {
		return errors.Wrapf(err, "Failed to load release asset '%v'", config.AssetPath)
	}

	log.Entry().Info("Starting to upload release asset.")
	asset, _, err := ghRepoClient.UploadReleaseAsset(ctx, config.Owner, config.Repository, releaseID, &opts, file)
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
