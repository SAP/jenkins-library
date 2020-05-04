package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMavenBuild(t *testing.T) {
	t.Run("mavenBuild should install the artifact", func(t *testing.T) {
		execMockRunner := mock.ExecMockRunner{}

		mockedUtils := mock.FilesMock{}

		config := mavenBuildOptions{}

		err := runMavenBuild(&config, nil, &execMockRunner, &mockedUtils)

		assert.Nil(t, err)
		assert.Equal(t, execMockRunner.Calls[0].Exec, "mvn")
		assert.Contains(t, execMockRunner.Calls[0].Params, "install")
	})

	t.Run("mavenBuild should skip integration tests", func(t *testing.T) {
		execMockRunner := mock.ExecMockRunner{}

		mockedUtils := mock.FilesMock{}
		mockedUtils.Files = []string{"integration-tests/pom.xml"}

		config := mavenBuildOptions{}

		err := runMavenBuild(&config, nil, &execMockRunner, &mockedUtils)

		assert.Nil(t, err)
		assert.Equal(t, execMockRunner.Calls[0].Exec, "mvn")
		assert.Contains(t, execMockRunner.Calls[0].Params, "-pl", "!integration-tests")
	})

	t.Run("mavenBuild should flatten", func(t *testing.T) {
		execMockRunner := mock.ExecMockRunner{}

		mockedUtils := mock.FilesMock{}

		config := mavenBuildOptions{Flatten: true}

		err := runMavenBuild(&config, nil, &execMockRunner, &mockedUtils)

		assert.Nil(t, err)
		assert.Contains(t, execMockRunner.Calls[0].Params, "flatten:flatten")
		assert.Contains(t, execMockRunner.Calls[0].Params, "-Dflatten.mode=resolveCiFriendliesOnly")
		assert.Contains(t, execMockRunner.Calls[0].Params, "-DupdatePomFile=true")
	})

	t.Run("mavenBuild should run only verify", func(t *testing.T) {
		execMockRunner := mock.ExecMockRunner{}

		mockedUtils := mock.FilesMock{}

		config := mavenBuildOptions{Verify: true}

		err := runMavenBuild(&config, nil, &execMockRunner, &mockedUtils)

		assert.Nil(t, err)
		assert.Contains(t, execMockRunner.Calls[0].Params, "verify")
		assert.NotContains(t, execMockRunner.Calls[0].Params, "install")
	})

}
