package cmd

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTrGitGetChangeDocumentID(t *testing.T) {
	t.Parallel()

	t.Run("good", func(t *testing.T) {
		t.Parallel()

		t.Run("getChangeDocumentID", func(t *testing.T) {
			configMock := newCdIDConfigMock()

			id, err := getChangeDocumentID(configMock.config, &transportRequestUtilsMock{cdID: "56781234"})

			if assert.NoError(t, err) {
				assert.Equal(t, id, "56781234")
			}
		})
		t.Run("runTransportRequestDocIDFromGit", func(t *testing.T) {
			configMock := newCdIDConfigMock()
			cpe := &transportRequestDocIDFromGitCommonPipelineEnvironment{}

			err := runTransportRequestDocIDFromGit(configMock.config, nil, &transportRequestUtilsMock{cdID: "56781234"}, cpe)

			if assert.NoError(t, err) {
				assert.Equal(t, cpe.custom.changeDocumentID, "56781234")
			}
		})

	})
	t.Run("bad", func(t *testing.T) {
		t.Parallel()

		t.Run("runTransportRequestDocIDFromGit", func(t *testing.T) {
			configMock := newCdIDConfigMock()
			cpe := &transportRequestDocIDFromGitCommonPipelineEnvironment{}

			err := runTransportRequestDocIDFromGit(configMock.config, nil, &transportRequestUtilsMock{err: errors.New("fail")}, cpe)

			assert.EqualError(t, err, "fail")
		})

	})
}

type cdIDConfigMock struct {
	config *transportRequestDocIDFromGitOptions
}

func newCdIDConfigMock() *cdIDConfigMock {
	return &cdIDConfigMock{
		config: &transportRequestDocIDFromGitOptions{
			GitFrom:             "origin/master",
			GitTo:               "HEAD",
			ChangeDocumentLabel: "ChangeDocument",
		},
	}
}
