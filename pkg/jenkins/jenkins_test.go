package jenkins

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/jenkins/mocks"
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
			On("BuildJob", jobID, map[string]string{}).
			Return(queueID, fmt.Errorf(mock.Anything))
		// test
		task, err := TriggerJob(jenkins, jobName, jobParameters)
		// asserts
		jenkins.AssertExpectations(t)
		assert.EqualError(t, err, mock.Anything)
		assert.Nil(t, task)
	})
	t.Run("error - task already queued", func(t *testing.T) {
		// init
		queueID := int64(0)
		jenkins := &mocks.Jenkins{}
		jenkins.
			On("BuildJob", jobID, map[string]string{}).
			Return(queueID, nil)
		// test
		task, err := TriggerJob(jenkins, jobName, jobParameters)
		// asserts
		jenkins.AssertExpectations(t)
		assert.EqualError(t, err, "Unable to queue build")
		assert.Nil(t, task)
	})
	t.Run("error - task not queued", func(t *testing.T) {
		// init
		queueID := int64(43)
		jenkins := &mocks.Jenkins{}
		jenkins.Test(t)
		jenkins.
			On("BuildJob", jobID, map[string]string{}).
			Return(queueID, nil).
			On("GetQueueItem", queueID).
			Return(nil, fmt.Errorf(mock.Anything))
		// test
		task, err := TriggerJob(jenkins, jobName, jobParameters)
		// asserts
		jenkins.AssertExpectations(t)
		assert.EqualError(t, err, mock.Anything)
		assert.Nil(t, task)
	})
}

func TestWaitForBuildToStart(t *testing.T) {
	jobName := "ContinuousDelivery/piper-library"
	jobID := strings.ReplaceAll(jobName, "/", "/job/")

	t.Run("error - build not started", func(t *testing.T) {
		// init
		buildNumber := int64(43)
		task := &mocks.Task{}
		task.On("WaitToStart", time.Millisecond).Return(buildNumber, nil)
		jenkins := &mocks.Jenkins{}
		jenkins.
			On("GetBuild", jobID, buildNumber).
			Return(nil, fmt.Errorf("Build not started"))
		// test
		build, err := WaitForBuildToStart(jenkins, jobName, task, time.Millisecond)
		// asserts
		task.AssertExpectations(t)
		jenkins.AssertExpectations(t)
		assert.EqualError(t, err, "Build not started")
		assert.Nil(t, build)
	})
}
