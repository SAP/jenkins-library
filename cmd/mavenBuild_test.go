//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var cpe mavenBuildCommonPipelineEnvironment

func TestMavenBuild(t *testing.T) {
	t.Run("mavenBuild should install the artifact", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)
		expectedParams := []string{"install"}

		assert.Nil(t, err)
		if assert.Equal(t, 1, len(mockedUtils.Calls), "Expected one maven invocation for the main build") {
			assert.Equal(t, "mvn", mockedUtils.Calls[0].Exec)
			assert.Contains(t, mockedUtils.Calls[0].Params, expectedParams[0], "Call should contain install goal")
		}
	})

	t.Run("mavenBuild accepts profiles", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{Profiles: []string{"profile1", "profile2"}}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 1, len(mockedUtils.Calls), "Expected one maven invocation for the main build") {
			assert.Contains(t, mockedUtils.Calls[0].Params, "--activate-profiles")
			assert.Contains(t, mockedUtils.Calls[0].Params, "profile1,profile2")
		}
	})

	t.Run("mavenBuild should create BOM", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{CreateBOM: true}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two Maven invocations (default + makeAggregateBom)") {
			assert.Equal(t, "mvn", mockedUtils.Calls[1].Exec)
			assert.Contains(t, mockedUtils.Calls[0].Params, "org.cyclonedx:cyclonedx-maven-plugin:2.9.1:makeAggregateBom")
			assert.Contains(t, mockedUtils.Calls[0].Params, "-DoutputName=bom-maven")
		}
	})

	t.Run("mavenBuild include install and deploy when publish is true", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{Publish: true, Verify: false, AltDeploymentRepositoryID: "ID", AltDeploymentRepositoryURL: "http://sampleRepo.com", AltDeploymentRepositoryUser: "user", AltDeploymentRepositoryPassword: "pass"}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two Maven invocations (main and deploy)") {
			assert.Contains(t, mockedUtils.Calls[0].Params, "install")
			assert.NotContains(t, mockedUtils.Calls[0].Params, "verify")
			assert.Contains(t, mockedUtils.Calls[1].Params, "deploy")
		}
	})

	t.Run("mavenBuild with deploy must skip build, install and test", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{Publish: true, Verify: false, DeployFlags: []string{"-Dmaven.main.skip=true", "-Dmaven.test.skip=true", "-Dmaven.install.skip=true"}, AltDeploymentRepositoryID: "ID", AltDeploymentRepositoryURL: "http://sampleRepo.com", AltDeploymentRepositoryUser: "user", AltDeploymentRepositoryPassword: "pass"}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two Maven invocations (main and deploy)") {
			assert.Contains(t, mockedUtils.Calls[1].Params, "-Dmaven.main.skip=true")
			assert.Contains(t, mockedUtils.Calls[1].Params, "-Dmaven.test.skip=true")
			assert.Contains(t, mockedUtils.Calls[1].Params, "-Dmaven.install.skip=true")
		}
	})

	t.Run("mavenBuild with deploy must include alt repo id and url when passed as parameter", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{Publish: true, Verify: false, AltDeploymentRepositoryID: "ID", AltDeploymentRepositoryURL: "http://sampleRepo.com", AltDeploymentRepositoryUser: "user", AltDeploymentRepositoryPassword: "pass"}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two Maven invocations (main and deploy)") {
			assert.Contains(t, mockedUtils.Calls[1].Params, "-DaltDeploymentRepository=ID::default::http://sampleRepo.com")
		}
	})

	t.Run("mavenBuild should not create build artifacts metadata when CreateBuildArtifactsMetadata is false and Publish is true", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()
		mockedUtils.AddFile("pom.xml", []byte{})
		config := mavenBuildOptions{CreateBuildArtifactsMetadata: false, Publish: true, AltDeploymentRepositoryID: "ID", AltDeploymentRepositoryURL: "http://sampleRepo.com", AltDeploymentRepositoryUser: "user", AltDeploymentRepositoryPassword: "pass"}
		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)
		assert.Nil(t, err)
		assert.Equal(t, mockedUtils.Calls[0].Exec, "mvn")
		assert.Contains(t, mockedUtils.Calls[0].Params, "install")
		assert.Empty(t, cpe.custom.mavenBuildArtifacts)
	})

	t.Run("mavenBuild should not create build artifacts metadata when CreateBuildArtifactsMetadata is true and Publish is false", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()
		mockedUtils.AddFile("pom.xml", []byte{})
		config := mavenBuildOptions{CreateBuildArtifactsMetadata: true, Publish: false}
		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)
		assert.Nil(t, err)
		assert.Equal(t, mockedUtils.Calls[0].Exec, "mvn")
		assert.Empty(t, cpe.custom.mavenBuildArtifacts)
	})
}
