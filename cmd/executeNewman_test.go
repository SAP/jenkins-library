package cmd

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

type executeNewmanMockUtils struct {
	errorOnGlob bool
	filesToFind []string
}

func newExecuteNewmanMockUtils() executeNewmanMockUtils {
	return executeNewmanMockUtils{
		filesToFind: []string{"localFile.txt", "2localFile.txt"},
	}
}

func TestRunExecuteNewman(t *testing.T) {
	t.Parallel()

	allFineConfig := executeNewmanOptions{
		NewmanCollection: "localFile.txt",
		NewmanRunCommand: "runcommand",
	}

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init

		utils := newExecuteNewmanMockUtils()

		// test
		err := runExecuteNewman(&allFineConfig, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error on file search", func(t *testing.T) {
		t.Parallel()
		// init

		utils := newExecuteNewmanMockUtils()
		utils.filesToFind = nil

		// test
		err := runExecuteNewman(&allFineConfig, nil, utils)

		// assert
		assert.EqualError(t, err, "no collection found with pattern 'localFile.txt'")
	})

	t.Run("no newman file", func(t *testing.T) {
		t.Parallel()
		// init

		utils := newExecuteNewmanMockUtils()
		utils.errorOnGlob = true

		// test
		err := runExecuteNewman(&allFineConfig, nil, utils)

		// assert
		assert.EqualError(t, err, "Could not execute global search for 'localFile.txt': error on Glob")
	})
}

func TestDefineCollectionDisplayName(t *testing.T) {
	t.Parallel()

	t.Run("normal path", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join("dir1", "dir2", "fancyFile.txt")
		result := defineCollectionDisplayName(path)
		assert.Equal(t, "dir1_dir2_fancyFile", result)
	})

	t.Run("directory", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join("dir1", "dir2", "dir3")
		result := defineCollectionDisplayName(path)
		assert.Equal(t, "dir1_dir2_dir3", result)
	})

	t.Run("empty path", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(".")
		result := defineCollectionDisplayName(path)
		assert.Equal(t, "", result)
	})
}

func TestResolveTemplate(t *testing.T) {
	t.Parallel()

	t.Run("nothing to replace", func(t *testing.T) {
		t.Parallel()

		config := executeNewmanOptions{NewmanRunCommand: "this is my fancy command"}

		cmd, err := resolveTemplate(&config, "collectionsDisplayName")
		assert.NoError(t, err)
		assert.Equal(t, "this is my fancy command", cmd)
	})

	t.Run("replace display name", func(t *testing.T) {
		t.Parallel()

		config := executeNewmanOptions{NewmanRunCommand: "this is my fancy command {{.CollectionDisplayName}}"}

		cmd, err := resolveTemplate(&config, "theDisplayName")
		assert.NoError(t, err)
		assert.Equal(t, "this is my fancy command theDisplayName", cmd)
	})

	t.Run("error when parameter cannot be resolved", func(t *testing.T) {
		t.Parallel()

		config := executeNewmanOptions{NewmanRunCommand: "this is my fancy command {{.collectionDisplayName}}"}

		_, err := resolveTemplate(&config, "theDisplayName")
		assert.EqualError(t, err, "error on executing template: template: template:1:27: executing \"template\" at <.collectionDisplayName>: can't evaluate field collectionDisplayName in type cmd.TemplateConfig")
	})

	t.Run("error when template cannot be parsed", func(t *testing.T) {
		t.Parallel()

		config := executeNewmanOptions{NewmanRunCommand: "this is my fancy command {{.collectionDisplayName}"}

		_, err := resolveTemplate(&config, "theDisplayName")
		assert.EqualError(t, err, "could not parse newman command template: template: template:1: unexpected \"}\" in operand")
	})
}

func (e executeNewmanMockUtils) Glob(string) (matches []string, err error) {
	if e.errorOnGlob {
		return nil, fmt.Errorf("error on Glob")
	}

	return e.filesToFind, nil
}
