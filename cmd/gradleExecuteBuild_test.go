package cmd

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/SAP/jenkins-library/pkg/mock"
)

type gradleExecuteBuildMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func TestRunGradleExecuteBuild(t *testing.T) {

	t.Run("failed case - build.gradle isn't present", func(t *testing.T) {
		utils := gradleExecuteBuildMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{},
			FilesMock:      &mock.FilesMock{},
		}
		options := &gradleExecuteBuildOptions{
			Path:       "path/to",
			Task:       "build",
			UseWrapper: false,
		}

		err := runGradleExecuteBuild(options, nil, utils)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "the specified gradle build script could not be found")
	})

	t.Run("success case - only build", func(t *testing.T) {
		utils := gradleExecuteBuildMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{},
			FilesMock:      &mock.FilesMock{},
		}
		utils.FilesMock.AddFile("path/to/build.gradle", []byte{})
		options := &gradleExecuteBuildOptions{
			Path:       "path/to",
			Task:       "build",
			UseWrapper: false,
		}

		err := runGradleExecuteBuild(options, nil, utils)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(utils.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"build", "-p", "path/to"}}, utils.Calls[0])
	})

	t.Run("success case - bom creation", func(t *testing.T) {
		utils := gradleExecuteBuildMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{},
			FilesMock:      &mock.FilesMock{},
		}
		utils.FilesMock.AddFile("path/to/build.gradle", []byte{})
		options := &gradleExecuteBuildOptions{
			Path:       "path/to",
			Task:       "build",
			UseWrapper: false,
			CreateBOM:  true,
		}

		err := runGradleExecuteBuild(options, nil, utils)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(utils.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"tasks", "-p", "path/to"}}, utils.Calls[0])
		assert.Equal(t, mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "gradle", Params: []string{"cyclonedxBom", "-p", "path/to", "--init-script", "initScript.gradle.tmp"}}, utils.Calls[1])
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"build", "-p", "path/to"}}, utils.Calls[2])
		assert.True(t, utils.HasWrittenFile("initScript.gradle.tmp"))
		assert.True(t, utils.HasRemovedFile("initScript.gradle.tmp"))
	})

	t.Run("success case - publishing of artifacts", func(t *testing.T) {
		utils := gradleExecuteBuildMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{},
			FilesMock:      &mock.FilesMock{},
		}
		utils.FilesMock.AddFile("path/to/build.gradle", []byte{})
		options := &gradleExecuteBuildOptions{
			Path:       "path/to",
			Task:       "build",
			UseWrapper: false,
			Publish:    true,
		}

		err := runGradleExecuteBuild(options, nil, utils)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(utils.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"build", "-p", "path/to"}}, utils.Calls[0])
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"tasks", "-p", "path/to"}}, utils.Calls[1])
		assert.Equal(t, mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "gradle", Params: []string{"publish", "-p", "path/to", "--init-script", "initScript.gradle.tmp"}}, utils.Calls[2])
		assert.True(t, utils.HasWrittenFile("initScript.gradle.tmp"))
		assert.True(t, utils.HasRemovedFile("initScript.gradle.tmp"))
	})

	t.Run("success case - build using wrapper", func(t *testing.T) {
		utils := gradleExecuteBuildMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{},
			FilesMock:      &mock.FilesMock{},
		}
		utils.FilesMock.AddFile("path/to/build.gradle", []byte{})
		utils.FilesMock.AddFile("gradlew", []byte{})
		options := &gradleExecuteBuildOptions{
			Path:       "path/to",
			Task:       "build",
			UseWrapper: true,
		}

		err := runGradleExecuteBuild(options, nil, utils)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(utils.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "./gradlew", Params: []string{"build", "-p", "path/to"}}, utils.Calls[0])
	})

	t.Run("failed case - build", func(t *testing.T) {
		utils := gradleExecuteBuildMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{
				ShouldFailOnCommand: map[string]error{"gradle build -p path/to": errors.New("failed to build")},
			},
			FilesMock: &mock.FilesMock{},
		}
		utils.FilesMock.AddFile("path/to/build.gradle", []byte{})
		options := &gradleExecuteBuildOptions{
			Path:       "path/to",
			Task:       "build",
			UseWrapper: false,
		}

		err := runGradleExecuteBuild(options, nil, utils)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to build")
	})

	t.Run("failed case - bom creation", func(t *testing.T) {
		utils := gradleExecuteBuildMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{
				ShouldFailOnCommand: map[string]error{"./gradlew cyclonedxBom -p path/to --init-script initScript.gradle.tmp": errors.New("failed to create bom")},
			},
			FilesMock: &mock.FilesMock{},
		}
		utils.FilesMock.AddFile("path/to/build.gradle", []byte{})
		utils.FilesMock.AddFile("gradlew", []byte{})
		options := &gradleExecuteBuildOptions{
			Path:       "path/to",
			Task:       "build",
			UseWrapper: true,
			CreateBOM:  true,
		}

		err := runGradleExecuteBuild(options, nil, utils)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create bom")
	})

	t.Run("failed case - publish artifacts", func(t *testing.T) {
		utils := gradleExecuteBuildMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{
				ShouldFailOnCommand: map[string]error{"./gradlew publish -p path/to --init-script initScript.gradle.tmp": errors.New("failed to publish artifacts")},
			},
			FilesMock: &mock.FilesMock{},
		}
		utils.FilesMock.AddFile("path/to/build.gradle", []byte{})
		utils.FilesMock.AddFile("gradlew", []byte{})
		options := &gradleExecuteBuildOptions{
			Path:       "path/to",
			Task:       "build",
			UseWrapper: true,
			Publish:    true,
		}

		err := runGradleExecuteBuild(options, nil, utils)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to publish artifacts")
	})
}
