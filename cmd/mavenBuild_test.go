package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMavenBuild(t *testing.T) {
	t.Run("mavenBuild should install the artifact", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{}

		err := runMavenBuild(&config, nil, &mockedUtils)

		assert.Nil(t, err)
		assert.Equal(t, mockedUtils.Calls[0].Exec, "mvn")
		assert.Contains(t, mockedUtils.Calls[0].Params, "install")
	})

	t.Run("mavenBuild should skip integration tests", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()
		mockedUtils.AddFile("integration-tests/pom.xml", []byte{})

		config := mavenBuildOptions{}

		err := runMavenBuild(&config, nil, &mockedUtils)

		assert.Nil(t, err)
		assert.Equal(t, mockedUtils.Calls[0].Exec, "mvn")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-pl", "!integration-tests")
	})

	t.Run("mavenBuild should flatten", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{Flatten: true}

		err := runMavenBuild(&config, nil, &mockedUtils)

		assert.Nil(t, err)
		assert.Contains(t, mockedUtils.Calls[0].Params, "flatten:flatten")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-Dflatten.mode=resolveCiFriendliesOnly")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-DupdatePomFile=true")
	})

	t.Run("mavenBuild should run only verify", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{Verify: true}

		err := runMavenBuild(&config, nil, &mockedUtils)

		assert.Nil(t, err)
		assert.Contains(t, mockedUtils.Calls[0].Params, "verify")
		assert.NotContains(t, mockedUtils.Calls[0].Params, "install")
	})

	t.Run("mavenBuild should createBOM", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{CreateBOM: true}

		err := runMavenBuild(&config, nil, &mockedUtils)

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

}
