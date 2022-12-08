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
	chunks, err := getBody(config, ioutil.ReadFile)
	if err != nil {
		return err
	}
	transformConfig(config, &options, chunks[0])
	err = piperGithub.CreateIssue(&options)
	if err != nil {
		return err
	}
	if len(chunks) > 0 {
		for _, v := range chunks[1:] {
			options.Body = []byte(v)
			options.UpdateExisting = true
			err = piperGithub.CreateIssue(&options)
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
			return nil, errors.Wrapf(err, "failed to read file '%v'", config.BodyFilePath)
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
	for i := 0; i < length; {
		to := length
		if to > i+chunkSize {
			to = i + chunkSize
		} else {
			chunks = append(chunks, string(value[i:to]))
			break
		}

		for j := to - 1; j > i; j-- {
			if value[j] == '\n' {
				to = j
				break
			}
		}
		fmt.Printf("to %v  i %v", to, i)
		chunks = append(chunks, string(value[i:to]))
		i = to
	}
	return chunks
}
