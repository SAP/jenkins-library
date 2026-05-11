//go:build unit

package buildsettings

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateBuildSettingsInfo(t *testing.T) {
	t.Parallel()

	t.Run("fresh path - no previous build settings", func(t *testing.T) {
		t.Parallel()
		testTableConfig := []struct {
			config    BuildOptions
			buildTool string
			expected  string
		}{
			{
				config:    BuildOptions{CreateBOM: true},
				buildTool: "golangBuild",
				expected:  `{"golangBuild":[{"createBOM":true}]}`,
			},
			{
				config:    BuildOptions{DockerImage: "golang:latest"},
				buildTool: "golangBuild",
				expected:  `{"golangBuild":[{"dockerImage":"golang:latest"}]}`,
			},
			{
				config:    BuildOptions{CreateBOM: true, DockerImage: "gradle:latest"},
				buildTool: "gradleExecuteBuild",
				expected:  `{"gradleExecuteBuild":[{"createBOM":true,"dockerImage":"gradle:latest"}]}`,
			},
			{
				config:    BuildOptions{Publish: true},
				buildTool: "helmExecute",
				expected:  `{"helmExecute":[{"publish":true}]}`,
			},
			{
				config:    BuildOptions{Publish: true},
				buildTool: "kanikoExecute",
				expected:  `{"kanikoExecute":[{"publish":true}]}`,
			},
			{
				config:    BuildOptions{Profiles: []string{"profile1", "profile2"}, CreateBOM: true},
				buildTool: "mavenBuild",
				expected:  `{"mavenBuild":[{"profiles":["profile1","profile2"],"createBOM":true}]}`,
			},
			{
				config:    BuildOptions{Profiles: []string{"release.build"}, Publish: true, GlobalSettingsFile: "http://nexus.test:8081/nexus/"},
				buildTool: "mtaBuild",
				expected:  `{"mtaBuild":[{"profiles":["release.build"],"publish":true,"globalSettingsFile":"http://nexus.test:8081/nexus/"}]}`,
			},
			{
				config:    BuildOptions{CreateBOM: true},
				buildTool: "pythonBuild",
				expected:  `{"pythonBuild":[{"createBOM":true}]}`,
			},
			{
				config:    BuildOptions{CreateBOM: true},
				buildTool: "npmExecuteScripts",
				expected:  `{"npmExecuteScripts":[{"createBOM":true}]}`,
			},
			{
				config:    BuildOptions{DockerImage: "builder:latest"},
				buildTool: "cnbBuild",
				expected:  `{"cnbBuild":[{"dockerImage":"builder:latest"}]}`,
			},
			{
				config:    BuildOptions{DockerImage: "docker:latest"},
				buildTool: "dockerBuild",
				expected:  `{"dockerBuild":[{"dockerImage":"docker:latest"}]}`,
			},
		}

		for _, testCase := range testTableConfig {
			buildSettings, err := CreateBuildSettingsInfo(&testCase.config, testCase.buildTool)
			assert.Nil(t, err)
			assert.Equal(t, testCase.expected, buildSettings)
		}
	})

	t.Run("fresh path - unsupported buildTool returns empty string without error", func(t *testing.T) {
		t.Parallel()
		config := BuildOptions{CreateBOM: true}
		result, err := CreateBuildSettingsInfo(&config, "unsupportedTool")
		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestCreateBuildSettingsInfo_MergePath(t *testing.T) {
	t.Parallel()

	t.Run("appends to existing key when buildTool already present", func(t *testing.T) {
		t.Parallel()
		config := BuildOptions{
			Profiles:          []string{"profile1", "profile2"},
			CreateBOM:         true,
			BuildSettingsInfo: `{"mavenBuild":[{"createBOM":true}]}`,
		}
		result, err := CreateBuildSettingsInfo(&config, "mavenBuild")
		assert.NoError(t, err)
		assert.Equal(t, `{"mavenBuild":[{"createBOM":true},{"profiles":["profile1","profile2"],"createBOM":true}]}`, result)
	})

	t.Run("adds new key alongside existing when buildTool not yet present", func(t *testing.T) {
		t.Parallel()
		config := BuildOptions{
			CreateBOM:         true,
			BuildSettingsInfo: `{"mavenBuild":[{"createBOM":true}]}`,
		}
		result, err := CreateBuildSettingsInfo(&config, "golangBuild")
		assert.NoError(t, err)
		assert.Contains(t, result, `"mavenBuild":[{"createBOM":true}]`)
		assert.Contains(t, result, `"golangBuild":[{"createBOM":true}]`)
	})

	t.Run("returns error on malformed BuildSettingsInfo JSON", func(t *testing.T) {
		t.Parallel()
		config := BuildOptions{
			BuildSettingsInfo: `{not-valid-json`,
		}
		_, err := CreateBuildSettingsInfo(&config, "golangBuild")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal existing build settings json")
	})

	t.Run("unsupported buildTool silently adds new key in merge path", func(t *testing.T) {
		t.Parallel()
		config := BuildOptions{
			CreateBOM:         true,
			BuildSettingsInfo: `{"mavenBuild":[{"createBOM":true}]}`,
		}
		result, err := CreateBuildSettingsInfo(&config, "unknownTool")
		assert.NoError(t, err)
		assert.Contains(t, result, `"mavenBuild"`)
		assert.Contains(t, result, `"unknownTool"`)
	})
}

// Env-override tests are sequential because os.Setenv affects the whole process.
func TestCreateBuildSettingsInfo_EnvDockerImageOverride(t *testing.T) {
	t.Run("fresh path - PIPER_dockerImage env overrides config.DockerImage", func(t *testing.T) {
		os.Setenv("PIPER_dockerImage", "override:fresh")
		defer os.Unsetenv("PIPER_dockerImage")
		config := BuildOptions{DockerImage: "original:fresh"}
		result, err := CreateBuildSettingsInfo(&config, "golangBuild")
		assert.NoError(t, err)
		assert.Equal(t, `{"golangBuild":[{"dockerImage":"override:fresh"}]}`, result)
	})

	t.Run("merge path - PIPER_dockerImage applies to appended entry only, historical entry preserved", func(t *testing.T) {
		os.Setenv("PIPER_dockerImage", "override:v2")
		defer os.Unsetenv("PIPER_dockerImage")
		config := BuildOptions{
			DockerImage:       "original:v1",
			BuildSettingsInfo: `{"golangBuild":[{"dockerImage":"original:v1"}]}`,
		}
		result, err := CreateBuildSettingsInfo(&config, "golangBuild")
		assert.NoError(t, err)
		assert.Equal(t, `{"golangBuild":[{"dockerImage":"original:v1"},{"dockerImage":"override:v2"}]}`, result)
	})
}
