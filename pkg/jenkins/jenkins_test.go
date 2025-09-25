package jenkins

import (
	"context"
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/jenkins/mocks"
	"github.com/bndr/gojenkins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInterfaceCompatibility(t *testing.T) {
	var _ Jenkins = new(gojenkins.Jenkins)
	var _ Build = new(gojenkins.Build)
}

func TestTriggerJob(t *testing.T) {
	ctx := context.Background()
	jobParameters := map[string]string{}

	t.Run("error - job not updated", func(t *testing.T) {
		// init
		jenkins := &mocks.Jenkins{}
		jenkins.Test(t)
		job := &mocks.Job{}
		job.Test(t)
		job.
			On("Poll", ctx).Return(404, fmt.Errorf("%s", mock.Anything))
		// test
		build, err := TriggerJob(ctx, jenkins, job, jobParameters)
		// asserts
		job.AssertExpectations(t)
		jenkins.AssertExpectations(t)
		assert.EqualError(t, err, fmt.Sprintf("failed to load job: %s", mock.Anything))
		assert.Nil(t, build)
	})
	t.Run("error - task not started", func(t *testing.T) {
		// init
		queueID := int64(0)
		jenkins := &mocks.Jenkins{}
		jenkins.Test(t)
		job := &mocks.Job{}
		job.Test(t)
		job.
			On("Poll", ctx).Return(200, nil).
			On("InvokeSimple", ctx, map[string]string{}).
			Return(queueID, fmt.Errorf("%s", mock.Anything))
		// test
		build, err := TriggerJob(ctx, jenkins, job, jobParameters)
		// asserts
		job.AssertExpectations(t)
		jenkins.AssertExpectations(t)
		assert.EqualError(t, err, mock.Anything)
		assert.Nil(t, build)
	})
	t.Run("error - task already queued", func(t *testing.T) {
		// init
		queueID := int64(0)
		jenkins := &mocks.Jenkins{}
		jenkins.Test(t)
		job := &mocks.Job{}
		job.Test(t)
		job.
			On("Poll", ctx).Return(200, nil).
			On("InvokeSimple", ctx, jobParameters).
			Return(queueID, nil)
		// test
		build, err := TriggerJob(ctx, jenkins, job, jobParameters)
		// asserts
		job.AssertExpectations(t)
		jenkins.AssertExpectations(t)
		assert.EqualError(t, err, "unable to queue build")
		assert.Nil(t, build)
	})
	t.Run("error - task not queued", func(t *testing.T) {
		// init
		queueID := int64(43)
		jenkins := &mocks.Jenkins{}
		jenkins.Test(t)
		job := &mocks.Job{}
		job.Test(t)
		job.
			On("Poll", ctx).Return(200, nil).
			On("InvokeSimple", ctx, jobParameters).
			Return(queueID, nil).
			On("GetJob").
			Return(&gojenkins.Job{})
		jenkins.
			On("GetBuildFromQueueID", ctx, mock.Anything, queueID).
			Return(nil, fmt.Errorf("%s", mock.Anything))
		// test
		build, err := TriggerJob(ctx, jenkins, job, jobParameters)
		// asserts
		job.AssertExpectations(t)
		jenkins.AssertExpectations(t)
		assert.EqualError(t, err, mock.Anything)
		assert.Nil(t, build)
	})
	t.Run("success", func(t *testing.T) {
		// init
		queueID := int64(43)
		jenkins := &mocks.Jenkins{}
		jenkins.Test(t)
		job := &mocks.Job{}
		job.Test(t)
		job.
			On("Poll", ctx).Return(200, nil).
			On("InvokeSimple", ctx, jobParameters).
			Return(queueID, nil).
			On("GetJob").
			Return(&gojenkins.Job{})
		jenkins.
			On("GetBuildFromQueueID", ctx, mock.Anything, queueID).
			Return(&gojenkins.Build{}, nil)
		// test
		build, err := TriggerJob(ctx, jenkins, job, jobParameters)
		// asserts
		jenkins.AssertExpectations(t)
		assert.NoError(t, err)
		assert.NotNil(t, build)
	})
}
