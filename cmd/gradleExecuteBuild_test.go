package cmd

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperenv"
)

const moduleFileContent = `{"variants": [{"name": "apiElements","files": [{"name": "gradle-1.2.3-12234567890-plain.jar"}]}]}`

type gradleExecuteBuildMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
	Filepath
}

type isDirEntryMock func() bool

func (d isDirEntryMock) Name() string {
	panic("not implemented")
}

func (d isDirEntryMock) IsDir() bool {
	return d()
}

func (d isDirEntryMock) Type() fs.FileMode {
	panic("not implemented")
}
func (d isDirEntryMock) Info() (fs.FileInfo, error) {
	panic("not implemented")
}

func TestRunGradleExecuteBuild(t *testing.T) {
	pipelineEnv := &gradleExecuteBuildCommonPipelineEnvironment{}

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

		err := runGradleExecuteBuild(options, nil, utils, pipelineEnv)
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

		err := runGradleExecuteBuild(options, nil, utils, pipelineEnv)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(utils.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"build", "-p", "path/to"}}, utils.Calls[0])
	})

	t.Run("success case - build with flags", func(t *testing.T) {
		utils := gradleExecuteBuildMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{},
			FilesMock:      &mock.FilesMock{},
		}
		utils.FilesMock.AddFile("path/to/build.gradle", []byte{})
		options := &gradleExecuteBuildOptions{
			Path:       "path/to",
			Task:       "build",
			BuildFlags: []string{"clean", "build", "-x", "test"},
			UseWrapper: false,
		}

		err := runGradleExecuteBuild(options, nil, utils, pipelineEnv)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(utils.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"clean", "build", "-x", "test", "-p", "path/to"}}, utils.Calls[0])
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

		err := runGradleExecuteBuild(options, nil, utils, pipelineEnv)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(utils.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"tasks", "-p", "path/to"}}, utils.Calls[0])
		assert.Equal(t, mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "gradle", Params: []string{"cyclonedxBom", "-p", "path/to", "--init-script", "initScript.gradle.tmp"}}, utils.Calls[1])
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"build", "-p", "path/to"}}, utils.Calls[2])
		assert.True(t, utils.HasWrittenFile("initScript.gradle.tmp"))
		assert.True(t, utils.HasRemovedFile("initScript.gradle.tmp"))
	})

	t.Run("success case - publishing of artifacts", func(t *testing.T) {
		var walkDir WalkDirFunc = func(root string, fn fs.WalkDirFunc) error {
			var dirMock isDirEntryMock = func() bool {
				return false
			}
			return fn(filepath.Join("test_subproject_path", "build", "publications", "maven", "module.json"), dirMock, nil)
		}
		utils := gradleExecuteBuildMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{},
			FilesMock:      &mock.FilesMock{},
			Filepath:       walkDir,
		}
		utils.FilesMock.AddFile("path/to/build.gradle", []byte{})
		utils.FilesMock.AddFile(filepath.Join("test_subproject_path", "build", "publications", "maven", "module.json"), []byte(moduleFileContent))
		options := &gradleExecuteBuildOptions{
			Path:       "path/to",
			Task:       "build",
			UseWrapper: false,
			Publish:    true,
		}

		err := runGradleExecuteBuild(options, nil, utils, pipelineEnv)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(utils.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"build", "-p", "path/to"}}, utils.Calls[0])
		assert.Equal(t, mock.ExecCall{Exec: "gradle", Params: []string{"tasks", "-p", "path/to"}}, utils.Calls[1])
		assert.Equal(t, mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "gradle", Params: []string{"publish", "-p", "path/to", "--init-script", "initScript.gradle.tmp"}}, utils.Calls[2])
		assert.Equal(t, "gradle-1.2.3-12234567890-plain.jar", pipelineEnv.custom.artifacts[0].Name)
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

		err := runGradleExecuteBuild(options, nil, utils, pipelineEnv)
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

		err := runGradleExecuteBuild(options, nil, utils, pipelineEnv)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to build")
	})

	t.Run("failed case - build with flags", func(t *testing.T) {
		utils := gradleExecuteBuildMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{
				ShouldFailOnCommand: map[string]error{"gradle clean build -x test -p path/to": errors.New("failed to build with flags")},
			},
			FilesMock: &mock.FilesMock{},
		}
		utils.FilesMock.AddFile("path/to/build.gradle", []byte{})
		options := &gradleExecuteBuildOptions{
			Path:       "path/to",
			Task:       "build",
			BuildFlags: []string{"clean", "build", "-x", "test"},
			UseWrapper: false,
		}

		err := runGradleExecuteBuild(options, nil, utils, pipelineEnv)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to build with flags")
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

		err := runGradleExecuteBuild(options, nil, utils, pipelineEnv)
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

		err := runGradleExecuteBuild(options, nil, utils, pipelineEnv)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to publish artifacts")
	})
}

func TestGetPublishedArtifactsNames(t *testing.T) {
	tt := []struct {
		name              string
		utils             gradleExecuteBuildMockUtils
		moduleFile        string
		moduleFileContent string
		expectedResult    piperenv.Artifacts
		expectedErr       error
	}{
		{
			name: "failed to check file existence",
			utils: gradleExecuteBuildMockUtils{
				ExecMockRunner: &mock.ExecMockRunner{},
				FilesMock: &mock.FilesMock{
					FileExistsErrors: map[string]error{"module.json": fmt.Errorf("err")},
				},
			},
			moduleFile:        "module.json",
			moduleFileContent: "",
			expectedErr:       fmt.Errorf("failed to check existence of the file 'module.json': err"),
		}, {
			name: "failed to get file",
			utils: gradleExecuteBuildMockUtils{
				ExecMockRunner: &mock.ExecMockRunner{},
				FilesMock:      &mock.FilesMock{},
			},
			moduleFile:        "",
			moduleFileContent: "",
			expectedErr:       fmt.Errorf("failed to get '': file does not exist"),
		}, {
			name: "failed to read file",
			utils: gradleExecuteBuildMockUtils{
				ExecMockRunner: &mock.ExecMockRunner{},
				FilesMock: &mock.FilesMock{
					FileReadErrors: map[string]error{"module.json": fmt.Errorf("err")},
				},
			},
			moduleFile:        "module.json",
			moduleFileContent: "",
			expectedErr:       fmt.Errorf("failed to read 'module.json': err"),
		}, {
			name: "failed to unmarshal file",
			utils: gradleExecuteBuildMockUtils{
				ExecMockRunner: &mock.ExecMockRunner{},
				FilesMock:      &mock.FilesMock{},
			},
			moduleFile:        "module.json",
			moduleFileContent: "",
			expectedErr:       fmt.Errorf("failed to unmarshal 'module.json': unexpected end of JSON input"),
		}, {
			name: "success - get name of published artifact",
			utils: gradleExecuteBuildMockUtils{
				ExecMockRunner: &mock.ExecMockRunner{},
				FilesMock:      &mock.FilesMock{},
			},
			moduleFile:        "module.json",
			moduleFileContent: moduleFileContent,
			expectedResult:    piperenv.Artifacts{piperenv.Artifact{Name: "gradle-1.2.3-12234567890-plain.jar"}},
			expectedErr:       nil,
		},
	}

	for _, test := range tt {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if test.moduleFile != "" {
				test.utils.FilesMock.AddFile(test.moduleFile, []byte(test.moduleFileContent))
			}
			artifacts, err := getPublishedArtifactsNames(test.moduleFile, test.utils)
			assert.Equal(t, test.expectedResult, artifacts)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}
