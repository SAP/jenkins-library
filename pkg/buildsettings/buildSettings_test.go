package buildsettings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateBuildSettingsInfo(t *testing.T) {

	t.Run("test build settings cpe with no previous existing values", func(t *testing.T) {
		testTableConfig := []struct {
			config    BuildOptions
			buildTool string
			expected  string
		}{
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
		}

		for _, testCase := range testTableConfig {
			builSettings, err := CreateBuildSettingsInfo(&testCase.config, testCase.buildTool)
			assert.Nil(t, err)
			assert.Equal(t, builSettings, testCase.expected)
		}
	})

}
