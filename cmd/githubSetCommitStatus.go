package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/google/go-github/v68/github"

	"github.com/pkg/errors"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
)

type gitHubCommitStatusRepositoriesService interface {
	CreateStatus(ctx context.Context, owner, repo, ref string, status *github.RepoStatus) (*github.RepoStatus, *github.Response, error)
}

func githubSetCommitStatus(config githubSetCommitStatusOptions, telemetryData *telemetry.CustomData) {
	// TODO provide parameter for trusted certs
	ctx, client, err := piperGithub.NewClientBuilder(config.Token, config.APIURL).Build()
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to get GitHub client")
	}

	err = runGithubSetCommitStatus(ctx, &config, telemetryData, client.Repositories)
	if err != nil {
		log.Entry().WithError(err).Fatal("GitHub status update failed")
	}
}

func runGithubSetCommitStatus(ctx context.Context, config *githubSetCommitStatusOptions, telemetryData *telemetry.CustomData, ghRepositoriesService gitHubCommitStatusRepositoriesService) error {
	status := github.RepoStatus{Context: &config.Context, Description: &config.Description, State: &config.Status, TargetURL: &config.TargetURL}
	_, _, err := ghRepositoriesService.CreateStatus(ctx, config.Owner, config.Repository, config.CommitID, &status)
	if err != nil {
		if strings.Contains(fmt.Sprint(err), "No commit found for SHA") {
			log.SetErrorCategory(log.ErrorCustom)
		}
		return errors.Wrapf(err, "failed to set status '%v' on commitId '%v'", config.Status, config.CommitID)
	}
	return nil
}
