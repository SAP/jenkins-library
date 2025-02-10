//go:build unit
// +build unit

package cmd

import (
	"testing"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
	"github.com/SAP/jenkins-library/pkg/mock"
	github "github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/assert"
)

func TestGetChunk(t *testing.T) {
	tests := []struct {
		name           string
		chunkSize      int
		largeString    string
		expectedChunks []string
	}{
		{
			name: "large string",
			largeString: `The quick
brown fox jumps
over
the lazy dog
`,
			chunkSize:      12,
			expectedChunks: []string{"The quick\nbr", "own fox jump", "s\nover\nthe l", "azy dog\n"},
		},
		{
			name:           "small string",
			largeString:    `small`,
			chunkSize:      12,
			expectedChunks: []string{"small"},
		},
		{
			name:           "exact size",
			largeString:    `exact size12`,
			chunkSize:      12,
			expectedChunks: []string{"exact size12"},
		},
		{
			name:           "empty string",
			largeString:    ``,
			chunkSize:      12,
			expectedChunks: []string{""},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			chunks := getChunks([]rune(test.largeString), test.chunkSize)
			assert.ElementsMatch(t, test.expectedChunks, chunks)
		})
	}
}

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
			ChunkSize:  100,
		}
		options := piperGithub.CreateIssueOptions{}
		resultChunks := []string{}
		createIssue := func(options *piperGithub.CreateIssueOptions) (*github.Issue, error) {
			resultChunks = append(resultChunks, string(options.Body))
			return nil, nil
		}

		// test
		err := runGithubCreateIssue(&config, nil, &options, &filesMock, createIssue)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, config.Token, options.Token)
		assert.Equal(t, config.APIURL, options.APIURL)
		assert.Equal(t, config.Owner, options.Owner)
		assert.Equal(t, config.Repository, options.Repository)
		assert.Equal(t, config.Title, options.Title)
		assert.Equal(t, config.Assignees, options.Assignees)
		assert.Equal(t, config.UpdateExisting, options.UpdateExisting)
		assert.ElementsMatch(t, resultChunks, []string{string(config.Body)})
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
			ChunkSize:    100,
		}
		options := piperGithub.CreateIssueOptions{}
		resultChunks := []string{}
		createIssue := func(options *piperGithub.CreateIssueOptions) (*github.Issue, error) {
			resultChunks = append(resultChunks, string(options.Body))
			return nil, nil
		}
		// test
		err := runGithubCreateIssue(&config, nil, &options, &filesMock, createIssue)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, config.Token, options.Token)
		assert.Equal(t, config.APIURL, options.APIURL)
		assert.Equal(t, config.Owner, options.Owner)
		assert.Equal(t, config.Repository, options.Repository)
		assert.Equal(t, config.Title, options.Title)
		assert.Equal(t, config.Assignees, options.Assignees)
		assert.Equal(t, config.UpdateExisting, options.UpdateExisting)
		assert.ElementsMatch(t, resultChunks, []string{"Test markdown"})
	})

	t.Run("Error - missing issue body", func(t *testing.T) {
		// init
		filesMock := mock.FilesMock{}
		config := githubCreateIssueOptions{ChunkSize: 100}
		options := piperGithub.CreateIssueOptions{}
		resultChunks := []string{}
		createIssue := func(options *piperGithub.CreateIssueOptions) (*github.Issue, error) {
			resultChunks = append(resultChunks, string(options.Body))
			return nil, nil
		}
		// test
		err := runGithubCreateIssue(&config, nil, &options, &filesMock, createIssue)

		// assert
		assert.EqualError(t, err, "either parameter `body` or parameter `bodyFilePath` is required")
	})
}
