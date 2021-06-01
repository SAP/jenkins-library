// +build integration
// can be execute with go test -tags=integration ./integration/...

package main

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SAP/jenkins-library/pkg/jenkins"
)

func TestTriggerJob(t *testing.T) {
	t.Skip("no Jenkins instance for testing available yet")
	// init
	ctx := context.Background()
	// ctx = context.WithValue(ctx, "debug", true)

	host := os.Getenv("PIPER_INTEGRATION_JENKINS_HOST")
	user := os.Getenv("PIPER_INTEGRATION_JENKINS_USER_NAME")
	token := os.Getenv("PIPER_INTEGRATION_JENKINS_TOKEN")
	job := os.Getenv("PIPER_INTEGRATION_JENKINS_JOB_NAME")
	require.NotEmpty(t, host, "Jenkins host url is missing")
	require.NotEmpty(t, user, "Jenkins user name is missing")
	require.NotEmpty(t, token, "Jenkins token is missing")
	require.NotEmpty(t, job, "Jenkins job name is missing")

	jenx, err := jenkins.Instance(ctx, http.DefaultClient, host, user, token)
	require.NotNil(t, jenx, "could not connect to Jenkins instance")
	require.NoError(t, err)
	// test
	build, err := jenkins.TriggerJob(ctx, jenx, job, nil)
	// asserts
	assert.NoError(t, err)
	assert.NotNil(t, build)
	assert.True(t, build.IsRunning(ctx))
}
