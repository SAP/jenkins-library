//go:build unit

package project

import (
	"net/http"
	"testing"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
)

func TestParseDescriptor(t *testing.T) {
	t.Run("parses the project.toml file v01", func(t *testing.T) {
		projectToml := `[project]
id = "io.buildpacks.my-app"
version = "0.1"

[build]
include = [
	"cmd/",
	"go.mod",
	"go.sum",
	"*.go"
]

[[build.env]]
name = "VAR1"
value = "VAL1"

[[build.env]]
name = "VAR2"
value = "VAL2"

[[build.env]]
name = "EMPTY"
value = ""

[[build.pre.group]]
id = "paketo-buildpacks/java"
version = "5.9.1"

[[build.post.group]]
id = "paketo-buildpacks/java"
version = "5.9.1"

[[build.buildpacks]]
id = "paketo-buildpacks/java"
version = "5.9.1"

[[build.buildpacks]]
id = "paketo-buildpacks/nodejs"
`
		utils := &cnbutils.MockUtils{
			FilesMock: &mock.FilesMock{},
		}

		fakeJavaResponse := "{\"latest\":{\"version\":\"1.1.1\",\"namespace\":\"test\",\"name\":\"test\",\"description\":\"\",\"homepage\":\"\",\"licenses\":null,\"stacks\":[\"test\",\"test\"],\"id\":\"test\"},\"versions\":[{\"version\":\"5.9.1\",\"_link\":\"https://test-java/5.9.1\"}]}"
		fakeNodeJsResponse := "{\"latest\":{\"version\":\"1.1.1\",\"namespace\":\"test\",\"name\":\"test\",\"description\":\"\",\"homepage\":\"\",\"licenses\":null,\"stacks\":[\"test\",\"test\"],\"id\":\"test\"},\"versions\":[{\"version\":\"1.1.1\",\"_link\":\"https://test-nodejs/1.1.1\"}]}"

		utils.AddFile("project.toml", []byte(projectToml))
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, "https://registry.buildpacks.io/api/v1/buildpacks/paketo-buildpacks/java", httpmock.NewStringResponder(200, fakeJavaResponse))
		httpmock.RegisterResponder(http.MethodGet, "https://registry.buildpacks.io/api/v1/buildpacks/paketo-buildpacks/nodejs", httpmock.NewStringResponder(200, fakeNodeJsResponse))

		httpmock.RegisterResponder(http.MethodGet, "https://test-java/5.9.1", httpmock.NewStringResponder(200, "{\"addr\": \"index.docker.io/test-java@5.9.1\"}"))
		httpmock.RegisterResponder(http.MethodGet, "https://test-nodejs/1.1.1", httpmock.NewStringResponder(200, "{\"addr\": \"index.docker.io/test-nodejs@1.1.1\"}"))
		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		descriptor, err := ParseDescriptor("project.toml", utils, client)

		assert.NoError(t, err)
		assert.Equal(t, descriptor.EnvVars["VAR1"], "VAL1")
		assert.Equal(t, descriptor.EnvVars["VAR2"], "VAL2")
		assert.Equal(t, descriptor.EnvVars["EMPTY"], "")

		assert.Equal(t, descriptor.ProjectID, "io.buildpacks.my-app")

		assert.Contains(t, descriptor.Buildpacks, "index.docker.io/test-java@5.9.1")
		assert.Contains(t, descriptor.Buildpacks, "index.docker.io/test-nodejs@1.1.1")

		assert.Contains(t, descriptor.PreBuildpacks, "index.docker.io/test-java@5.9.1")
		assert.Contains(t, descriptor.PostBuildpacks, "index.docker.io/test-java@5.9.1")

		assert.NotNil(t, descriptor.Include)

		t3 := descriptor.Include.MatchesPath("cmd/cobra.go")
		assert.True(t, t3)

		t4 := descriptor.Include.MatchesPath("pkg/test/main.go")
		assert.True(t, t4)

		t5 := descriptor.Include.MatchesPath("Makefile")
		assert.False(t, t5)
	})

	t.Run("parses the project.toml file v02", func(t *testing.T) {
		projectToml := `[_]
id = "io.buildpacks.my-app"
version = "0.1"
schema-version = "0.2"

[io.buildpacks]
include = [
	"cmd/",
	"go.mod",
	"go.sum",
	"*.go"
]

[[io.buildpacks.build.env]]
name = "VAR1"
value = "VAL1"

[[io.buildpacks.build.env]]
name = "VAR2"
value = "VAL2"

[[io.buildpacks.build.env]]
name = "EMPTY"
value = ""

[[io.buildpacks.pre.group]]
id = "paketo-buildpacks/java"
version = "5.9.1"

[[io.buildpacks.post.group]]
id = "paketo-buildpacks/java"
version = "5.9.1"

[[io.buildpacks.group]]
id = "paketo-buildpacks/java"
version = "5.9.1"

[[io.buildpacks.group]]
id = "paketo-buildpacks/nodejs"
`
		utils := &cnbutils.MockUtils{
			FilesMock: &mock.FilesMock{},
		}

		fakeJavaResponse := "{\"latest\":{\"version\":\"1.1.1\",\"namespace\":\"test\",\"name\":\"test\",\"description\":\"\",\"homepage\":\"\",\"licenses\":null,\"stacks\":[\"test\",\"test\"],\"id\":\"test\"},\"versions\":[{\"version\":\"5.9.1\",\"_link\":\"https://test-java/5.9.1\"}]}"
		fakeNodeJsResponse := "{\"latest\":{\"version\":\"1.1.1\",\"namespace\":\"test\",\"name\":\"test\",\"description\":\"\",\"homepage\":\"\",\"licenses\":null,\"stacks\":[\"test\",\"test\"],\"id\":\"test\"},\"versions\":[{\"version\":\"1.1.1\",\"_link\":\"https://test-nodejs/1.1.1\"}]}"

		utils.AddFile("project.toml", []byte(projectToml))
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, "https://registry.buildpacks.io/api/v1/buildpacks/paketo-buildpacks/java", httpmock.NewStringResponder(200, fakeJavaResponse))
		httpmock.RegisterResponder(http.MethodGet, "https://registry.buildpacks.io/api/v1/buildpacks/paketo-buildpacks/nodejs", httpmock.NewStringResponder(200, fakeNodeJsResponse))

		httpmock.RegisterResponder(http.MethodGet, "https://test-java/5.9.1", httpmock.NewStringResponder(200, "{\"addr\": \"index.docker.io/test-java@5.9.1\"}"))
		httpmock.RegisterResponder(http.MethodGet, "https://test-nodejs/1.1.1", httpmock.NewStringResponder(200, "{\"addr\": \"index.docker.io/test-nodejs@1.1.1\"}"))
		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		descriptor, err := ParseDescriptor("project.toml", utils, client)

		assert.NoError(t, err)
		assert.Equal(t, "VAL1", descriptor.EnvVars["VAR1"])
		assert.Equal(t, "VAL2", descriptor.EnvVars["VAR2"])
		assert.Equal(t, "", descriptor.EnvVars["EMPTY"])

		assert.Equal(t, "io.buildpacks.my-app", descriptor.ProjectID)

		assert.Contains(t, descriptor.Buildpacks, "index.docker.io/test-java@5.9.1")
		assert.Contains(t, descriptor.Buildpacks, "index.docker.io/test-nodejs@1.1.1")
		assert.Contains(t, descriptor.PreBuildpacks, "index.docker.io/test-java@5.9.1")
		assert.Contains(t, descriptor.PostBuildpacks, "index.docker.io/test-java@5.9.1")

		assert.NotNil(t, descriptor.Include)

		t3 := descriptor.Include.MatchesPath("cmd/cobra.go")
		assert.True(t, t3)

		t4 := descriptor.Include.MatchesPath("pkg/test/main.go")
		assert.True(t, t4)

		t5 := descriptor.Include.MatchesPath("Makefile")
		assert.False(t, t5)
	})

	t.Run("fails with inline buildpack", func(t *testing.T) {
		projectToml := `[project]
id = "io.buildpacks.my-app"
version = "0.1"

[[build.buildpacks]]
id = "test/inline"
	[build.buildpacks.script]
	api = "0.5"
	shell = "/bin/bash"
	inline = "date"
`
		utils := &cnbutils.MockUtils{
			FilesMock: &mock.FilesMock{},
		}

		utils.AddFile("project.toml", []byte(projectToml))

		_, err := ParseDescriptor("project.toml", utils, &piperhttp.Client{})

		assert.Error(t, err)
		assert.Equal(t, "inline buildpacks are not supported", err.Error())
	})

	t.Run("fails with both exclude and include specified", func(t *testing.T) {
		projectToml := `[project]
id = "io.buildpacks.my-app"
version = "0.1"

[build]
include = [
	"test"
]

exclude = [
	"test"
]
`

		utils := &cnbutils.MockUtils{
			FilesMock: &mock.FilesMock{},
		}
		utils.AddFile("project.toml", []byte(projectToml))

		_, err := ParseDescriptor("project.toml", utils, &piperhttp.Client{})

		assert.Error(t, err)
		assert.Equal(t, "project descriptor options 'exclude' and 'include' are mutually exclusive", err.Error())
	})

	t.Run("fails with file not found", func(t *testing.T) {
		utils := &cnbutils.MockUtils{
			FilesMock: &mock.FilesMock{},
		}

		_, err := ParseDescriptor("project.toml", utils, &piperhttp.Client{})

		assert.Error(t, err)
		assert.Equal(t, "could not read 'project.toml'", err.Error())
	})

	t.Run("fails to parse corrupted project.toml", func(t *testing.T) {
		projectToml := "test123"
		utils := &cnbutils.MockUtils{
			FilesMock: &mock.FilesMock{},
		}
		utils.AddFile("project.toml", []byte(projectToml))
		_, err := ParseDescriptor("project.toml", utils, &piperhttp.Client{})

		assert.Error(t, err)
		assert.Equal(t, "parsing schema version: toml: line 1: unexpected EOF; expected key separator '='", err.Error())
	})
}
