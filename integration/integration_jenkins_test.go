//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestJenkinsIntegration ./integration/...

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

func TestJenkinsIntegrationTriggerJob(t *testing.T) {
	t.Skip("no Jenkins instance for testing available yet")
	//TODO: check if testcontainers can be used
	// init
	ctx := context.Background()
	// ctx = context.WithValue(ctx, "debug", true)

	// os.Setenv("PIPER_INTEGRATION_JENKINS_USER_NAME", "")
	// os.Setenv("PIPER_INTEGRATION_JENKINS_TOKEN", "")
	// os.Setenv("PIPER_INTEGRATION_JENKINS_HOST", "")
	// os.Setenv("PIPER_INTEGRATION_JENKINS_JOB_NAME", "")

	host := os.Getenv("PIPER_INTEGRATION_JENKINS_HOST")
	user := os.Getenv("PIPER_INTEGRATION_JENKINS_USER_NAME")
	token := os.Getenv("PIPER_INTEGRATION_JENKINS_TOKEN")
	jobName := os.Getenv("PIPER_INTEGRATION_JENKINS_JOB_NAME")
	require.NotEmpty(t, host, "Jenkins host url is missing")
	require.NotEmpty(t, user, "Jenkins user name is missing")
	require.NotEmpty(t, token, "Jenkins token is missing")
	require.NotEmpty(t, jobName, "Jenkins job name is missing")

	jenx, err := jenkins.Instance(ctx, http.DefaultClient, host, user, token)
	require.NotNil(t, jenx, "could not connect to Jenkins instance")
	require.NoError(t, err)
	// test
	job, getJobErr := jenkins.GetJob(ctx, jenx, jobName)
	build, triggerJobErr := jenkins.TriggerJob(ctx, jenx, job, nil)
	// asserts
	assert.NoError(t, getJobErr)
	assert.NoError(t, triggerJobErr)
	assert.NotNil(t, build)
	assert.True(t, build.IsRunning(ctx))
}
