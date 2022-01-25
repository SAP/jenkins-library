package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
)

func githubCreateIssue(config githubCreateIssueOptions, telemetryData *telemetry.CustomData) {
	err := runGithubCreateIssue(&config, telemetryData)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to comment on issue")
	}
}

func runGithubCreateIssue(config *githubCreateIssueOptions, _ *telemetry.CustomData) error {

	options := piperGithub.CreateIssueOptions{}
	err := transformConfig(config, &options, ioutil.ReadFile)
	if err != nil {
		return err
	}

	return piperGithub.CreateIssue(&options)
}

func transformConfig(config *githubCreateIssueOptions, options *piperGithub.CreateIssueOptions, readFile func(string) ([]byte, error)) error {
	options.Token = config.Token
	options.APIURL = config.APIURL
	options.Owner = config.Owner
	options.Repository = config.Repository
	options.Title = config.Title
	options.Body = []byte(config.Body)
	options.Assignees = config.Assignees
	options.UpdateExisting = config.UpdateExisting

	if len(config.Body)+len(config.BodyFilePath) == 0 {
		return fmt.Errorf("either parameter `body` or parameter `bodyFilePath` is required")
	}
	if len(config.Body) == 0 {
		issueContent, err := readFile(config.BodyFilePath)
		if err != nil {
			return errors.Wrapf(err, "failed to read file '%v'", config.BodyFilePath)
		}
		options.Body = issueContent
	}
	return nil
}
