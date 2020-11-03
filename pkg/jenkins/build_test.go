package jenkins

import (
	"fmt"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/jenkins/mocks"
	"github.com/bndr/gojenkins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWaitForBuildToFinish(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// init
		build := &mocks.Build{}
		build.
			On("IsRunning").Return(true).Once().
			On("IsRunning").Return(false)
		// test
		WaitForBuildToFinish(build, time.Millisecond)
		// asserts
		build.AssertExpectations(t)
	})
}

func TestFetchBuildArtifact(t *testing.T) {
	fileName := "artifactFile.xml"

	t.Run("success", func(t *testing.T) {
		// init
		build := &mocks.Build{}
		build.On("IsRunning").Return(false)
		build.On("GetArtifacts").Return(
			[]gojenkins.Artifact{
				gojenkins.Artifact{FileName: mock.Anything},
				gojenkins.Artifact{FileName: fileName},
			},
		)
		// test
		artifact, err := FetchBuildArtifact(build, fileName)
		// asserts
		build.AssertExpectations(t)
		assert.NoError(t, err)
		assert.Equal(t, fileName, artifact.FileName())
	})
	t.Run("error - job running", func(t *testing.T) {
		// init
		build := &mocks.Build{}
		build.On("IsRunning").Return(true)
		// test
		_, err := FetchBuildArtifact(build, fileName)
		// asserts
		build.AssertExpectations(t)
		assert.EqualError(t, err, "Failed to fetch artifact: Job is still running")
	})
	t.Run("error - no artifacts", func(t *testing.T) {
		// init
		build := &mocks.Build{}
		build.On("IsRunning").Return(false)
		build.On("GetArtifacts").Return([]gojenkins.Artifact{})
		// test
		_, err := FetchBuildArtifact(build, fileName)
		// asserts
		build.AssertExpectations(t)
		assert.EqualError(t, err, fmt.Sprintf("Failed to fetch artifact: Artifact '%s' not found", fileName))
	})
	t.Run("error - artifact not found", func(t *testing.T) {
		// init
		build := &mocks.Build{}
		build.On("IsRunning").Return(false)
		build.On("GetArtifacts").Return([]gojenkins.Artifact{gojenkins.Artifact{FileName: mock.Anything}})
		// test
		_, err := FetchBuildArtifact(build, fileName)
		// asserts
		build.AssertExpectations(t)
		assert.EqualError(t, err, fmt.Sprintf("Failed to fetch artifact: Artifact '%s' not found", fileName))
	})
}
