package project_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/cnbutils/project"
	"github.com/SAP/jenkins-library/pkg/mock"
)

func TestResolvePath(t *testing.T) {
	t.Run("project descriptor and no path is maintained, it is located in the root", func(t *testing.T) {
		utils := &cnbutils.MockUtils{
			FilesMock: &mock.FilesMock{},
		}
		utils.AddFile("project.toml", []byte(""))

		location, err := project.ResolvePath("project.toml", "", utils)
		require.NoError(t, err)

		assert.Equal(t, "project.toml", location)
	})

	t.Run("project descriptor and path is is a file, it is located in the root", func(t *testing.T) {
		utils := &cnbutils.MockUtils{
			FilesMock: &mock.FilesMock{},
		}
		utils.CurrentDir = "/workdir"
		utils.AddFile("project.toml", []byte(""))
		utils.AddFile("test/file.zip", []byte(""))

		location, err := project.ResolvePath("project.toml", "test/file.zip", utils)
		require.NoError(t, err)

		assert.Equal(t, "/workdir/project.toml", location)
	})

	t.Run("project descriptor and path is is a dir, it is located in the path", func(t *testing.T) {
		utils := &cnbutils.MockUtils{
			FilesMock: &mock.FilesMock{},
		}
		utils.AddFile("test/project.toml", []byte(""))

		location, err := project.ResolvePath("project.toml", "test", utils)
		require.NoError(t, err)

		assert.Equal(t, filepath.Join("test", "project.toml"), location)
	})

	t.Run("project descriptor does not exist, empty string", func(t *testing.T) {
		utils := &cnbutils.MockUtils{
			FilesMock: &mock.FilesMock{},
		}
		location, err := project.ResolvePath("project.toml", "", utils)
		require.NoError(t, err)

		assert.Equal(t, "", location)
	})
}
