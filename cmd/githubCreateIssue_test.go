package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"

	"github.com/stretchr/testify/assert"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
)

func TestTransformConfig(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		// init
		filesMock := mock.FilesMock{}
		config := githubCreateIssueOptions{
			Owner:      "TEST",
			Repository: "test",
			Body:       "This is my test body",
			Title:      "This is my title",
			Assignees:  []string{"userIdOne", "userIdTwo"},
		}
		options := piperGithub.CreateIssueOptions{}

		// test
		err := transformConfig(&config, &options, filesMock.FileRead)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, config.Token, options.Token)
		assert.Equal(t, config.APIURL, options.APIURL)
		assert.Equal(t, config.Owner, options.Owner)
		assert.Equal(t, config.Repository, options.Repository)
		assert.Equal(t, config.Body, options.Body)
		assert.Equal(t, config.Title, options.Title)
		assert.Equal(t, config.Assignees, options.Assignees)
		assert.Equal(t, config.UpdateExisting, options.UpdateExisting)
	})

	t.Run("Success bodyFilePath", func(t *testing.T) {
		// init
		filesMock := mock.FilesMock{}
		filesMock.AddFile("test.md", []byte("Test markdown"))
		config := githubCreateIssueOptions{
			Owner:        "TEST",
			Repository:   "test",
			BodyFilePath: "test.md",
			Title:        "This is my title",
			Assignees:    []string{"userIdOne", "userIdTwo"},
		}
		options := piperGithub.CreateIssueOptions{}

		// test
		err := transformConfig(&config, &options, filesMock.FileRead)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, config.Token, options.Token)
		assert.Equal(t, config.APIURL, options.APIURL)
		assert.Equal(t, config.Owner, options.Owner)
		assert.Equal(t, config.Repository, options.Repository)
		assert.Equal(t, config.Body, options.Body)
		assert.Equal(t, config.Title, options.Title)
		assert.Equal(t, config.Assignees, options.Assignees)
		assert.Equal(t, config.UpdateExisting, options.UpdateExisting)
	})

	t.Run("Error - missing issue body", func(t *testing.T) {
		// init
		filesMock := mock.FilesMock{}
		config := githubCreateIssueOptions{}
		options := piperGithub.CreateIssueOptions{}

		// test
		err := transformConfig(&config, &options, filesMock.FileRead)

		// assert
		assert.EqualError(t, err, "either parameter `body` or parameter `bodyFilePath` is required")
	})
}
