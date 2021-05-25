package jenkins

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/jenkins/mocks"
	"github.com/bndr/gojenkins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTriggerJob(t *testing.T) {
	jobName := "ContinuousDelivery/piper-library"
	jobID := strings.ReplaceAll(jobName, "/", "/job/")
	jobParameters := map[string]string{}

	t.Run("error - task not started", func(t *testing.T) {
		// init
		queueID := int64(0)
		jenkins := &mocks.Jenkins{}
		jenkins.
			On("BuildJob", context.Background(), jobID, map[string]string{}).
			Return(queueID, fmt.Errorf(mock.Anything))
		// test
		build, err := TriggerJob(jenkins, jobName, jobParameters)
		// asserts
		jenkins.AssertExpectations(t)
		assert.EqualError(t, err, mock.Anything)
		assert.Nil(t, build)
	})
	t.Run("error - task already queued", func(t *testing.T) {
		// init
		queueID := int64(0)
		jenkins := &mocks.Jenkins{}
		jenkins.
			On("BuildJob", context.Background(), jobID, map[string]string{}).
			Return(queueID, nil)
		// test
		build, err := TriggerJob(jenkins, jobName, jobParameters)
		// asserts
		jenkins.AssertExpectations(t)
		assert.EqualError(t, err, "unable to queue build")
		assert.Nil(t, build)
	})
	t.Run("error - task not queued", func(t *testing.T) {
		// init
		queueID := int64(43)
		jenkins := &mocks.Jenkins{}
		jenkins.Test(t)
		jenkins.
			On("BuildJob", context.Background(), jobID, map[string]string{}).
			Return(queueID, nil).
			On("GetBuildFromQueueID", context.Background(), queueID).
			Return(nil, fmt.Errorf(mock.Anything))
		// test
		build, err := TriggerJob(jenkins, jobName, jobParameters)
		// asserts
		jenkins.AssertExpectations(t)
		assert.EqualError(t, err, mock.Anything)
		assert.Nil(t, build)
	})
	t.Run("success", func(t *testing.T) {
		// init
		queueID := int64(43)
		jenkins := &mocks.Jenkins{}
		jenkins.Test(t)
		jenkins.
			On("BuildJob", context.Background(), jobID, map[string]string{}).
			Return(queueID, nil).
			On("GetBuildFromQueueID", context.Background(), queueID).
			Return(&gojenkins.Build{}, nil)
		// test
		build, err := TriggerJob(jenkins, jobName, jobParameters)
		// asserts
		jenkins.AssertExpectations(t)
		assert.NoError(t, err)
		assert.NotNil(t, build)
	})
}
