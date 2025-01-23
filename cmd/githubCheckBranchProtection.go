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

type gitHubBranchProtectionRepositoriesService interface {
	GetBranchProtection(ctx context.Context, owner, repo, branch string) (*github.Protection, *github.Response, error)
}

func githubCheckBranchProtection(config githubCheckBranchProtectionOptions, telemetryData *telemetry.CustomData) {
	// TODO provide parameter for trusted certs
	ctx, client, err := piperGithub.NewClientBuilder(config.Token, config.APIURL).Build()
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to get GitHub client")
	}

	err = runGithubCheckBranchProtection(ctx, &config, telemetryData, client.Repositories)
	if err != nil {
		log.Entry().WithError(err).Fatal("GitHub branch protection check failed")
	}
}

func runGithubCheckBranchProtection(ctx context.Context, config *githubCheckBranchProtectionOptions, telemetryData *telemetry.CustomData, ghRepositoriesService gitHubBranchProtectionRepositoriesService) error {
	ghProtection, _, err := ghRepositoriesService.GetBranchProtection(ctx, config.Owner, config.Repository, config.Branch)
	if err != nil {
		return errors.Wrap(err, "failed to read branch protection information")
	}

	// validate required status checks
	for _, check := range config.RequiredChecks {
		var found bool
		foundContexts := []string{}
		if requiredStatusChecks := ghProtection.GetRequiredStatusChecks(); requiredStatusChecks != nil && requiredStatusChecks.Contexts != nil {
			foundContexts = *requiredStatusChecks.Contexts
		}
		for _, context := range foundContexts {
			if check == context {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("required status check '%v' not found among '%v' in branch protection configuration", check, strings.Join(foundContexts, ","))
		}
	}

	// validate that admins are enforced in checks
	if config.RequireEnforceAdmins && !ghProtection.GetEnforceAdmins().Enabled {
		return fmt.Errorf("admins are not enforced in branch protection configuration")
	}

	// validate number of mandatory reviewers
	if config.RequiredApprovingReviewCount > 0 && ghProtection.GetRequiredPullRequestReviews().RequiredApprovingReviewCount < config.RequiredApprovingReviewCount {
		return fmt.Errorf("not enough mandatory reviewers in branch protection configuration, expected at least %v, got %v", config.RequiredApprovingReviewCount, ghProtection.GetRequiredPullRequestReviews().RequiredApprovingReviewCount)
	}

	return nil
}
