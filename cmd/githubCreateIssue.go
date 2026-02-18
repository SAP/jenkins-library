package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
	github "github.com/google/go-github/v68/github"
)

type githubCreateIssueUtils interface {
	FileRead(string) ([]byte, error)
}

func githubCreateIssue(config githubCreateIssueOptions, telemetryData *telemetry.CustomData) {
	fileUtils := &piperutils.Files{}
	options := piperGithub.CreateIssueOptions{}
	err := runGithubCreateIssue(&config, telemetryData, &options, fileUtils, piperGithub.CreateIssue)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to comment on issue")
	}
}

func runGithubCreateIssue(config *githubCreateIssueOptions, _ *telemetry.CustomData, options *piperGithub.CreateIssueOptions, utils githubCreateIssueUtils, createIssue func(*piperGithub.CreateIssueOptions) (*github.Issue, error)) error {
	chunks, err := getBody(config, utils.FileRead)
	if err != nil {
		return err
	}
	transformConfig(config, options, chunks[0])
	issue, err := createIssue(options)
	if err != nil {
		return err
	}
	if len(chunks) > 1 {
		for _, v := range chunks[1:] {
			options.Body = []byte(v)
			options.Issue = issue
			options.UpdateExisting = true
			_, err = createIssue(options)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getBody(config *githubCreateIssueOptions, readFile func(string) ([]byte, error)) ([]string, error) {
	var bodyString []rune
	if len(config.Body)+len(config.BodyFilePath) == 0 {
		return nil, fmt.Errorf("either parameter `body` or parameter `bodyFilePath` is required")
	}
	if len(config.Body) == 0 {
		issueContent, err := readFile(config.BodyFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file '%v': %w", config.BodyFilePath, err)
		}
		bodyString = []rune(string(issueContent))
	} else {
		bodyString = []rune(config.Body)
	}
	return getChunks(bodyString, config.ChunkSize), nil
}

func transformConfig(config *githubCreateIssueOptions, options *piperGithub.CreateIssueOptions, body string) {
	options.Token = config.Token
	options.APIURL = config.APIURL
	options.Owner = config.Owner
	options.Repository = config.Repository
	options.Title = config.Title
	options.Body = []byte(config.Body)
	options.Assignees = config.Assignees
	options.UpdateExisting = config.UpdateExisting
	options.Body = []byte(body)
}

func getChunks(value []rune, chunkSize int) []string {
	chunks := []string{}
	length := len(value)
	if length == 0 {
		return []string{""}
	}
	for i := 0; i < length; i += chunkSize {
		to := length
		if to > i+chunkSize {
			to = i + chunkSize
		}
		chunks = append(chunks, string(value[i:to]))
	}
	return chunks
}
