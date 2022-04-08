package cmd

import (
	"testing"

	piperconf "github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/stretchr/testify/assert"
)

func TestMavenBuild(t *testing.T) {
	configOptions.openFile = piperconf.OpenPiperFile

	cpe := mavenBuildCommonPipelineEnvironment{}

	t.Run("mavenBuild should install the artifact", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		assert.Equal(t, mockedUtils.Calls[0].Exec, "mvn")
		assert.Contains(t, mockedUtils.Calls[0].Params, "install")
	})

	t.Run("mavenBuild should populate the cpe artifacts", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()
		mockedUtils.AddFile("pom.xml", []byte{})
		mockedUtils.AddFile("target/artifacts-test.jar", []byte{})
		mockedUtils.AddFile("target/artifacts-test.war", []byte{})
		mockedUtils.AddFile("target/artifacts-test-classes.jar", []byte{})

		mockedUtils.StdoutReturn = map[string]string{
			"mvn .*project.build.finalName": "artifacts-test",
		}

		config := mavenBuildOptions{}
		cpe := mavenBuildCommonPipelineEnvironment{}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		assert.ElementsMatch(t, cpe.artifacts, []piperenv.Artifact{{
			Kind: "java:jar",
			Path: "target/artifacts-test.jar",
		}, {
			Kind: "java:war",
			Path: "target/artifacts-test.war",
		}, {
			Kind: "java:classes-jar",
			Path: "target/artifacts-test-classes.jar",
		}})
	})

	t.Run("mavenBuild should skip integration tests", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()
		mockedUtils.AddFile("integration-tests/pom.xml", []byte{})

		config := mavenBuildOptions{}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		assert.Equal(t, mockedUtils.Calls[0].Exec, "mvn")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-pl", "!integration-tests")
	})

	t.Run("mavenBuild should flatten", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{Flatten: true}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		assert.Contains(t, mockedUtils.Calls[0].Params, "flatten:flatten")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-Dflatten.mode=resolveCiFriendliesOnly")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-DupdatePomFile=true")
	})

	t.Run("mavenBuild should run only verify", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{Verify: true}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		assert.Contains(t, mockedUtils.Calls[0].Params, "verify")
		assert.NotContains(t, mockedUtils.Calls[0].Params, "install")
	})

	t.Run("mavenBuild should createBOM", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{CreateBOM: true}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		assert.Contains(t, mockedUtils.Calls[0].Params, "org.cyclonedx:cyclonedx-maven-plugin:makeAggregateBom")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-DschemaVersion=1.2")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-DincludeBomSerialNumber=true")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-DincludeCompileScope=true")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-DincludeProvidedScope=true")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-DincludeRuntimeScope=true")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-DincludeSystemScope=true")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-DincludeTestScope=false")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-DincludeLicenseText=false")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-DoutputFormat=xml")
	})

	t.Run("mavenBuild include install and deploy when publish is true", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{Publish: true, Verify: false}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		assert.Contains(t, mockedUtils.Calls[0].Params, "install")
		assert.NotContains(t, mockedUtils.Calls[0].Params, "verify")
		assert.Contains(t, mockedUtils.Calls[1].Params, "deploy")

	})

	t.Run("mavenBuild with deploy must skip build, install and test", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{Publish: true, Verify: false}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		assert.Contains(t, mockedUtils.Calls[1].Params, "-Dmaven.main.skip=true")
		assert.Contains(t, mockedUtils.Calls[1].Params, "-Dmaven.test.skip=true")
		assert.Contains(t, mockedUtils.Calls[1].Params, "-Dmaven.install.skip=true")

	})

	t.Run("mavenBuild with deploy must include alt repo id and url when passed as parameter", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{Publish: true, Verify: false, AltDeploymentRepositoryID: "ID", AltDeploymentRepositoryURL: "http://sampleRepo.com"}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		assert.Contains(t, mockedUtils.Calls[1].Params, "-DaltDeploymentRepository=ID::default::http://sampleRepo.com")
	})

	t.Run("mavenBuild accepts profiles", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{Profiles: []string{"profile1", "profile2"}}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		assert.Contains(t, mockedUtils.Calls[0].Params, "--activate-profiles")
		assert.Contains(t, mockedUtils.Calls[0].Params, "profile1,profile2")
	})

}
