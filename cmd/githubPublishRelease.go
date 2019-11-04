package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v28/github"
	"github.com/pkg/errors"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
)

type githubRepoClient interface {
	GetLatestRelease(ctx context.Context, owner string, repo string) (*github.RepositoryRelease, *github.Response, error)
	CreateRelease(ctx context.Context, owner string, repo string, release *github.RepositoryRelease) (*github.RepositoryRelease, *github.Response, error)
}

type githubIssueClient interface {
	ListByRepo(ctx context.Context, owner string, repo string, opt *github.IssueListByRepoOptions) ([]*github.Issue, *github.Response, error)
}

func githubPublishRelease(myGithubPublishReleaseOptions githubPublishReleaseOptions) error {
	ctx, client, err := piperGithub.NewClient(myGithubPublishReleaseOptions.GithubToken, myGithubPublishReleaseOptions.GithubAPIURL, myGithubPublishReleaseOptions.GithubAPIURL)
	if err != nil {
		return err
	}

	err = runGithubPublishRelease(ctx, &myGithubPublishReleaseOptions, client.Repositories, client.Issues)
	if err != nil {
		return err
	}

	return nil
}

func runGithubPublishRelease(ctx context.Context, myGithubPublishReleaseOptions *githubPublishReleaseOptions, ghRepoClient githubRepoClient, ghIssueClient githubIssueClient) error {

	var publishedAt github.Timestamp
	lastRelease, resp, err := ghRepoClient.GetLatestRelease(ctx, myGithubPublishReleaseOptions.GithubOrg, myGithubPublishReleaseOptions.GithubRepo)
	if err != nil {
		if resp.StatusCode == 404 {
			//first release
			myGithubPublishReleaseOptions.AddDeltaToLastRelease = false
			publishedAt = lastRelease.GetPublishedAt()
		} else {
			return errors.Wrap(err, "Error occured when retrieving latest GitHub releass")
		}
	}

	releaseBody := ""

	if len(myGithubPublishReleaseOptions.ReleaseBodyHeader) > 0 {
		releaseBody += myGithubPublishReleaseOptions.ReleaseBodyHeader + "<br /\n>"
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

	//create release
	createdRelease, _, err := ghRepoClient.CreateRelease(ctx, myGithubPublishReleaseOptions.GithubOrg, myGithubPublishReleaseOptions.GithubRepo, &release)
	if err != nil {
		return errors.Wrapf(err, "creation of release '%v' failed", release.TagName)
	}

	// todo switch to logging
	fmt.Printf("Release %v created on %v/%v", *createdRelease.TagName, myGithubPublishReleaseOptions.GithubOrg, myGithubPublishReleaseOptions.GithubRepo)

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
	ghIssues, _, err := ghIssueClient.ListByRepo(ctx, myGithubPublishReleaseOptions.GithubOrg, myGithubPublishReleaseOptions.GithubRepo, &options)
	if err != nil {
		//log error
	}

	prTexts := []string{"<br />**List of closed pull-requests since last release**"}
	issueTexts := []string{"<br />**List of closed issues since last release**"}

	for _, issue := range ghIssues {
		if issue.IsPullRequest() && !isExcluded(issue, myGithubPublishReleaseOptions.ExcludeLabels) {
			prTexts = append(prTexts, fmt.Sprintf("[#%v](%v): %v", issue.GetNumber(), issue.GetHTMLURL(), issue.GetTitle()))
		} else if !issue.IsPullRequest() && !isExcluded(issue, myGithubPublishReleaseOptions.ExcludeLabels) {
			issueTexts = append(issueTexts, fmt.Sprintf("[#%v](%v): %v", issue.GetNumber(), issue.GetHTMLURL(), issue.GetTitle()))
		}
	}

	if len(prTexts) > 1 {
		closedIssuesText += strings.Join(prTexts, "\n") + "\n"
	}

	if len(issueTexts) > 1 {
		closedIssuesText += strings.Join(issueTexts, "\n") + "\n"
	}
	return closedIssuesText
}

func getReleaseDeltaText(myGithubPublishReleaseOptions *githubPublishReleaseOptions, lastRelease *github.RepositoryRelease) string {
	releaseDeltaText := ""

	//add delta link to previous release
	releaseDeltaText += "<br />**Changes**<br />"
	releaseDeltaText += fmt.Sprintf(
		"[%v...%v](%v/%v/%v/compare/%v...%v) <br />",
		lastRelease.GetTagName(),
		myGithubPublishReleaseOptions.Version,
		myGithubPublishReleaseOptions.GithubServerURL,
		myGithubPublishReleaseOptions.GithubOrg,
		myGithubPublishReleaseOptions.GithubRepo,
		lastRelease.GetTagName(), myGithubPublishReleaseOptions.Version,
	)

	return releaseDeltaText
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
