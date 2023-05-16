//go:build unit
// +build unit

package cnbutils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/stretchr/testify/assert"
)

func TestGetImageName(t *testing.T) {
	t.Parallel()

	t.Run("Registry without protocol will add https", func(t *testing.T) {
		t.Parallel()

		targetImage, err := cnbutils.GetTargetImage("registry", "image", "tag", "", "")

		assert.NoError(t, err)
		assert.Equal(t, "https", targetImage.ContainerRegistry.Scheme)
		assert.Equal(t, "registry", targetImage.ContainerRegistry.Host)
	})

	t.Run("Registry with protocol will keep it", func(t *testing.T) {
		t.Parallel()

		targetImage, err := cnbutils.GetTargetImage("http://registry", "image", "tag", "", "")

		assert.NoError(t, err)
		assert.Equal(t, "http", targetImage.ContainerRegistry.Scheme)
		assert.Equal(t, "registry", targetImage.ContainerRegistry.Host)
	})

	t.Run("Image name is taken from the configuration", func(t *testing.T) {
		t.Parallel()

		targetImage, err := cnbutils.GetTargetImage("http://registry", "image", "tag", "", "")

		assert.NoError(t, err)
		assert.Equal(t, "image", targetImage.ContainerImageName)
		assert.Equal(t, "tag", targetImage.ContainerImageTag)
	})

	t.Run("Image name is taken from project.toml", func(t *testing.T) {
		t.Parallel()

		targetImage, err := cnbutils.GetTargetImage("http://registry", "", "tag", "project-id.0", "")

		assert.NoError(t, err)
		assert.Equal(t, "project-id-0", targetImage.ContainerImageName)
		assert.Equal(t, "tag", targetImage.ContainerImageTag)
	})

	t.Run("Image name is taken from git repo", func(t *testing.T) {
		t.Parallel()

		tmpdir := t.TempDir()

		err := os.MkdirAll(filepath.Join(tmpdir, "commonPipelineEnvironment", "git"), os.ModePerm)
		assert.NoError(t, err)

		err = os.WriteFile(filepath.Join(tmpdir, "commonPipelineEnvironment", "git", "repository"), []byte("repo-name"), os.ModePerm)
		assert.NoError(t, err)

		targetImage, err := cnbutils.GetTargetImage("http://registry", "", "tag", "", tmpdir)
		assert.NoError(t, err)
		assert.Equal(t, "repo-name", targetImage.ContainerImageName)
		assert.Equal(t, "tag", targetImage.ContainerImageTag)
	})

	t.Run("Image name is taken from github repo", func(t *testing.T) {
		t.Parallel()

		tmpdir := t.TempDir()

		err := os.MkdirAll(filepath.Join(tmpdir, "commonPipelineEnvironment", "github"), os.ModePerm)
		assert.NoError(t, err)

		err = os.WriteFile(filepath.Join(tmpdir, "commonPipelineEnvironment", "github", "repository"), []byte("repo-name"), os.ModePerm)
		assert.NoError(t, err)

		targetImage, err := cnbutils.GetTargetImage("http://registry", "", "tag", "", tmpdir)
		assert.NoError(t, err)
		assert.Equal(t, "repo-name", targetImage.ContainerImageName)
		assert.Equal(t, "tag", targetImage.ContainerImageTag)
	})

	t.Run("throws an error if unable to find image name", func(t *testing.T) {
		t.Parallel()

		_, err := cnbutils.GetTargetImage("http://registry", "", "tag", "", "")

		assert.Error(t, err)
		assert.Equal(t, "failed to derive default for image name", err.Error())
	})
}
