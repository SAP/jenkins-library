package jenkins

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/jenkins/mocks"
	"github.com/bndr/gojenkins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWaitForBuildToFinish(t *testing.T) {
	ctx := context.Background()
	t.Run("success", func(t *testing.T) {
		// init
		build := &mocks.Build{}
		build.On("Poll", ctx).Return(200, nil)
		build.
			On("IsRunning", ctx).Return(true).Once().
			On("IsRunning", ctx).Return(false)
		// test
		WaitForBuildToFinish(ctx, build, time.Millisecond)
		// asserts
		build.AssertExpectations(t)
	})
}

func TestFetchBuildArtifact(t *testing.T) {
	ctx := context.Background()
	fileName := "artifactFile.xml"

	t.Run("success", func(t *testing.T) {
		// init
		build := &mocks.Build{}
		build.On("IsRunning", ctx).Return(false)
		build.On("GetArtifacts").Return(
			[]gojenkins.Artifact{
				{FileName: mock.Anything},
				{FileName: fileName},
			},
		)
		// test
		artifact, err := FetchBuildArtifact(ctx, build, fileName)
		// asserts
		build.AssertExpectations(t)
		assert.NoError(t, err)
		assert.Equal(t, fileName, artifact.FileName())
	})
	t.Run("error - job running", func(t *testing.T) {
		// init
		build := &mocks.Build{}
		build.On("IsRunning", ctx).Return(true)
		// test
		_, err := FetchBuildArtifact(ctx, build, fileName)
		// asserts
		build.AssertExpectations(t)
		assert.EqualError(t, err, "failed to fetch artifact: Job is still running")
	})
	t.Run("error - no artifacts", func(t *testing.T) {
		// init
		build := &mocks.Build{}
		build.On("IsRunning", ctx).Return(false)
		build.On("GetArtifacts").Return([]gojenkins.Artifact{})
		// test
		_, err := FetchBuildArtifact(ctx, build, fileName)
		// asserts
		build.AssertExpectations(t)
		assert.EqualError(t, err, fmt.Sprintf("failed to fetch artifact: Artifact '%s' not found", fileName))
	})
	t.Run("error - artifact not found", func(t *testing.T) {
		// init
		build := &mocks.Build{}
		build.On("IsRunning", ctx).Return(false)
		build.On("GetArtifacts").Return([]gojenkins.Artifact{{FileName: mock.Anything}})
		// test
		_, err := FetchBuildArtifact(ctx, build, fileName)
		// asserts
		build.AssertExpectations(t)
		assert.EqualError(t, err, fmt.Sprintf("failed to fetch artifact: Artifact '%s' not found", fileName))
	})
}
