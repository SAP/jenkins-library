//go:build unit
// +build unit

package cmd

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

type transportRequestUtilsMock struct {
	err  error
	trID string
	cdID string
}

func (m *transportRequestUtilsMock) FindIDInRange(label, from, to string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if strings.HasPrefix(label, "TransportRequest") {
		return m.trID, nil
	}
	if strings.HasPrefix(label, "ChangeDocument") {
		return m.cdID, nil
	}

	return "invalid", fmt.Errorf("invalid label passed: %s", label)
}

func TestTrGitGetTransportRequestID(t *testing.T) {
	t.Parallel()

	t.Run("good", func(t *testing.T) {
		t.Parallel()

		t.Run("getTransportRequestID", func(t *testing.T) {
			configMock := newTrIDConfigMock()

			id, err := getTransportRequestID(configMock.config, &transportRequestUtilsMock{trID: "43218765"})

			if assert.NoError(t, err) {
				assert.Equal(t, id, "43218765")
			}
		})
		t.Run("runTransportRequestDocIDFromGit", func(t *testing.T) {
			configMock := newTrIDConfigMock()
			cpe := &transportRequestReqIDFromGitCommonPipelineEnvironment{}

			err := runTransportRequestReqIDFromGit(configMock.config, nil, &transportRequestUtilsMock{trID: "43218765"}, cpe)

			if assert.NoError(t, err) {
				assert.Equal(t, cpe.custom.transportRequestID, "43218765")
			}
		})
	})
	t.Run("bad", func(t *testing.T) {
		t.Parallel()

		t.Run("runTransportRequestDocIDFromGit", func(t *testing.T) {
			configMock := newTrIDConfigMock()
			cpe := &transportRequestReqIDFromGitCommonPipelineEnvironment{}

			err := runTransportRequestReqIDFromGit(configMock.config, nil, &transportRequestUtilsMock{err: errors.New("fail")}, cpe)

			assert.EqualError(t, err, "fail")
		})

	})
}

type trIDConfigMock struct {
	config *transportRequestReqIDFromGitOptions
}

func newTrIDConfigMock() *trIDConfigMock {
	return &trIDConfigMock{
		config: &transportRequestReqIDFromGitOptions{
			GitFrom:               "origin/master",
			GitTo:                 "HEAD",
			TransportRequestLabel: "TransportRequest",
		},
	}
}
