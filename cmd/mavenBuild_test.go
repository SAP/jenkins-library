package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
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

}
