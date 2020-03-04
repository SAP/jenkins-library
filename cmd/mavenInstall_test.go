package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMavenInstall(t *testing.T) {
	t.Run("mavenInstall should install the artifact", func(t *testing.T) {
		execMockRunner := mock.ExecMockRunner{}

		config := mavenInstallOptions{}

		err := runMavenInstall(&config, nil, &execMockRunner)

		assert.Nil(t, err)
		assert.Equal(t, execMockRunner.Calls[0].Exec, "mvn")
		assert.Contains(t, execMockRunner.Calls[0].Params, "install")
	})

	t.Run("mavenInstall should flatten", func(t *testing.T) {
		execMockRunner := mock.ExecMockRunner{}

		config := mavenInstallOptions{Flatten: true}

		err := runMavenInstall(&config, nil, &execMockRunner)

		assert.Nil(t, err)
		assert.Contains(t, execMockRunner.Calls[0].Params, "flatten:flatten")
		assert.Contains(t, execMockRunner.Calls[0].Params, "-Dflatten.mode=resolveCiFriendliesOnly")
		assert.Contains(t, execMockRunner.Calls[0].Params, "-DupdatePomFile=true")
	})


}

