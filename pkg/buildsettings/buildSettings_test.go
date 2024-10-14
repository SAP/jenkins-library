package buildsettings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateBuildSettingsInfo(t *testing.T) {

	t.Run("test build settings cpe with no previous and existing values", func(t *testing.T) {
		testTableConfig := []struct {
			config    BuildOptions
			buildTool string
			expected  string
		}{
			{
				config:    BuildOptions{CreateBOM: true},
				buildTool: "golangBuild",
				expected:  "{\"golangBuild\":[{\"createBOM\":true}]}",
			},
			{
				config:    BuildOptions{DockerImage: "golang:latest"},
				buildTool: "golangBuild",
				expected:  "{\"golangBuild\":[{\"dockerImage\":\"golang:latest\"}]}",
			},
			{
				config:    BuildOptions{CreateBOM: true, DockerImage: "gradle:latest"},
				buildTool: "gradleExecuteBuild",
				expected:  "{\"gradleExecuteBuild\":[{\"createBOM\":true,\"dockerImage\":\"gradle:latest\"}]}",
			},
			{
				config:    BuildOptions{Publish: true},
				buildTool: "helmExecute",
				expected:  "{\"helmExecute\":[{\"publish\":true}]}",
			},
			{
				config:    BuildOptions{Publish: true},
				buildTool: "kanikoExecute",
				expected:  "{\"kanikoExecute\":[{\"publish\":true}]}",
			},
			{
				config:    BuildOptions{Profiles: []string{"profile1", "profile2"}, CreateBOM: true},
				buildTool: "mavenBuild",
				expected:  "{\"mavenBuild\":[{\"profiles\":[\"profile1\",\"profile2\"],\"createBOM\":true}]}",
			},
			{
				config:    BuildOptions{Profiles: []string{"profile1", "profile2"}, CreateBOM: true, BuildSettingsInfo: "{\"mavenBuild\":[{\"createBOM\":true}]}"},
				buildTool: "mavenBuild",
				expected:  "{\"mavenBuild\":[{\"createBOM\":true},{\"profiles\":[\"profile1\",\"profile2\"],\"createBOM\":true}]}",
			},
			{
				config:    BuildOptions{Profiles: []string{"release.build"}, Publish: true, GlobalSettingsFile: "http://nexus.test:8081/nexus/"},
				buildTool: "mtaBuild",
				expected:  "{\"mtaBuild\":[{\"profiles\":[\"release.build\"],\"publish\":true,\"globalSettingsFile\":\"http://nexus.test:8081/nexus/\"}]}",
			},
			{
				config:    BuildOptions{CreateBOM: true},
				buildTool: "pythonBuild",
				expected:  "{\"pythonBuild\":[{\"createBOM\":true}]}",
			},
			{
				config:    BuildOptions{CreateBOM: true},
				buildTool: "npmExecuteScripts",
				expected:  "{\"npmExecuteScripts\":[{\"createBOM\":true}]}",
			},
			{
				config:    BuildOptions{DockerImage: "builder:latest"},
				buildTool: "cnbBuild",
				expected:  "{\"cnbBuild\":[{\"dockerImage\":\"builder:latest\"}]}",
			},
		}

		for _, testCase := range testTableConfig {
			buildSettings, err := CreateBuildSettingsInfo(&testCase.config, testCase.buildTool)
			assert.Nil(t, err)
			assert.Equal(t, testCase.expected, buildSettings)
		}
	})

}
