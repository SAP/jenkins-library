//go:build unit
// +build unit

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"

	"github.com/SAP/jenkins-library/pkg/config"
)

var cpe mavenBuildCommonPipelineEnvironment

func TestMavenBuild(t *testing.T) {
	SetConfigOptions(ConfigCommandOptions{
		OpenFile: config.OpenPiperFile,
	})

	t.Run("mavenBuild should install the artifact", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		options := mavenBuildOptions{}

		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)
		expectedParams := []string{"install"}

		assert.Nil(t, err)
		if assert.Equal(t, 1, len(mockedUtils.Calls), "Expected one maven invocation for the main build") {
			assert.Equal(t, "mvn", mockedUtils.Calls[0].Exec)
			assert.Contains(t, mockedUtils.Calls[0].Params, expectedParams[0], "Call should contain install goal")
		}
	})

	t.Run("mavenBuild accepts profiles", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		options := mavenBuildOptions{Profiles: []string{"profile1", "profile2"}}

		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 1, len(mockedUtils.Calls), "Expected one maven invocation for the main build") {
			assert.Contains(t, mockedUtils.Calls[0].Params, "--activate-profiles")
			assert.Contains(t, mockedUtils.Calls[0].Params, "profile1,profile2")
		}
	})

	t.Run("mavenBuild should create BOM", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		options := mavenBuildOptions{CreateBOM: true}

		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two Maven invocations (default + makeAggregateBom)") {
			assert.Equal(t, "mvn", mockedUtils.Calls[1].Exec)
			assert.Contains(t, mockedUtils.Calls[0].Params, mvnCycloneDXPackage+":makeAggregateBom")
			assert.Contains(t, mockedUtils.Calls[0].Params, "-DoutputName=bom-maven")
		}
	})

	t.Run("mavenBuild include install and deploy when publish is true", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		options := mavenBuildOptions{Publish: true, Verify: false, AltDeploymentRepositoryID: "ID", AltDeploymentRepositoryURL: "http://sampleRepo.com", AltDeploymentRepositoryUser: "user", AltDeploymentRepositoryPassword: "pass"}

		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two Maven invocations (main and deploy)") {
			assert.Contains(t, mockedUtils.Calls[0].Params, "install")
			assert.NotContains(t, mockedUtils.Calls[0].Params, "verify")
			assert.Contains(t, mockedUtils.Calls[1].Params, "deploy")
		}
	})

	t.Run("mavenBuild with deploy must skip build, install and test", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		options := mavenBuildOptions{Publish: true, Verify: false, DeployFlags: []string{"-Dmaven.main.skip=true", "-Dmaven.test.skip=true", "-Dmaven.install.skip=true"}, AltDeploymentRepositoryID: "ID", AltDeploymentRepositoryURL: "http://sampleRepo.com", AltDeploymentRepositoryUser: "user", AltDeploymentRepositoryPassword: "pass"}

		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two Maven invocations (main and deploy)") {
			assert.Contains(t, mockedUtils.Calls[1].Params, "-Dmaven.main.skip=true")
			assert.Contains(t, mockedUtils.Calls[1].Params, "-Dmaven.test.skip=true")
			assert.Contains(t, mockedUtils.Calls[1].Params, "-Dmaven.install.skip=true")
		}
	})

	t.Run("mavenBuild with deploy must include alt repo id and url when passed as parameter", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		options := mavenBuildOptions{Publish: true, Verify: false, AltDeploymentRepositoryID: "ID", AltDeploymentRepositoryURL: "http://sampleRepo.com", AltDeploymentRepositoryUser: "user", AltDeploymentRepositoryPassword: "pass"}

		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two Maven invocations (main and deploy)") {
			assert.Contains(t, mockedUtils.Calls[1].Params, "-DaltDeploymentRepository=ID::default::http://sampleRepo.com")
		}
	})

	t.Run("mavenBuild should not create build artifacts metadata when CreateBuildArtifactsMetadata is false and Publish is true", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()
		mockedUtils.AddFile("pom.xml", []byte{})
		options := mavenBuildOptions{CreateBuildArtifactsMetadata: false, Publish: true, AltDeploymentRepositoryID: "ID", AltDeploymentRepositoryURL: "http://sampleRepo.com", AltDeploymentRepositoryUser: "user", AltDeploymentRepositoryPassword: "pass"}
		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)
		assert.Nil(t, err)
		assert.Equal(t, mockedUtils.Calls[0].Exec, "mvn")
		assert.Contains(t, mockedUtils.Calls[0].Params, "install")
		assert.Empty(t, cpe.custom.mavenBuildArtifacts)
	})

	t.Run("mavenBuild should not create build artifacts metadata when CreateBuildArtifactsMetadata is true and Publish is false", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()
		mockedUtils.AddFile("pom.xml", []byte{})
		options := mavenBuildOptions{CreateBuildArtifactsMetadata: true, Publish: false}
		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)
		assert.Nil(t, err)
		assert.Equal(t, mockedUtils.Calls[0].Exec, "mvn")
		assert.Empty(t, cpe.custom.mavenBuildArtifacts)
	})
}

func TestLoadRemoteRepoCertificates(t *testing.T) {
	t.Run("should find cacerts at Java 9+ path", func(t *testing.T) {
		filesMock := &mock.FilesMock{}
		execMock := &mock.ExecMockRunner{}

		javaHome := "/usr/lib/jvm/java-17"
		os.Setenv("JAVA_HOME", javaHome)
		defer os.Unsetenv("JAVA_HOME")

		java9Path := filepath.Join(javaHome, "lib", "security", "cacerts")
		filesMock.AddFile(java9Path, []byte("cacerts content"))
		filesMock.AddFile(".pipeline/mavenCaCerts", []byte{})

		var flags []string
		err := loadRemoteRepoCertificates([]string{}, nil, &flags, execMock, filesMock, "")

		assert.NoError(t, err)
	})

	t.Run("should fall back to Java 8 path", func(t *testing.T) {
		filesMock := &mock.FilesMock{}
		execMock := &mock.ExecMockRunner{}

		javaHome := "/usr/lib/jvm/java-8"
		os.Setenv("JAVA_HOME", javaHome)
		defer os.Unsetenv("JAVA_HOME")

		java8Path := filepath.Join(javaHome, "jre", "lib", "security", "cacerts")
		filesMock.AddFile(java8Path, []byte("cacerts content"))
		filesMock.AddFile(".pipeline/mavenCaCerts", []byte{})

		var flags []string
		err := loadRemoteRepoCertificates([]string{}, nil, &flags, execMock, filesMock, "")

		assert.NoError(t, err)
	})

	t.Run("should use custom javaCaCertFilePath when provided", func(t *testing.T) {
		filesMock := &mock.FilesMock{}
		execMock := &mock.ExecMockRunner{}

		customPath := "/custom/path/to/cacerts"
		filesMock.AddFile(customPath, []byte("cacerts content"))
		filesMock.AddFile(".pipeline/mavenCaCerts", []byte{})

		var flags []string
		err := loadRemoteRepoCertificates([]string{}, nil, &flags, execMock, filesMock, customPath)

		assert.NoError(t, err)
	})

	t.Run("should return nil and warn when cacerts not found", func(t *testing.T) {
		filesMock := &mock.FilesMock{}
		execMock := &mock.ExecMockRunner{}

		javaHome := "/usr/lib/jvm/java-17"
		os.Setenv("JAVA_HOME", javaHome)
		defer os.Unsetenv("JAVA_HOME")

		// Don't add any cacerts file - simulating missing file

		var flags []string
		err := loadRemoteRepoCertificates([]string{}, nil, &flags, execMock, filesMock, "")

		// Should not error - just warn and continue
		assert.NoError(t, err)
	})
}
