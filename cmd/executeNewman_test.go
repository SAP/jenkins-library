package cmd

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"strings"
	"testing"
)

type executeNewmanMockUtils struct {
	errorOnGlob                 bool
	errorOnNewmanInstall        bool
	errorOnRunShell             bool
	errorOnFinalScriptExecution bool
	errorOnRunExecutable        bool
	errorOnLoggingNode          bool
	errorOnLoggingNpm           bool
	executedExecutable          string
	executedParams              []string
	executedShell               string
	executedScript              string
	filesToFind                 []string
}

func newExecuteNewmanMockUtils() executeNewmanMockUtils {
	return executeNewmanMockUtils{
		filesToFind: []string{"localFile.txt", "2localFile.txt"},
	}
}

func TestRunExecuteNewman(t *testing.T) {
	t.Parallel()

	allFineConfig := executeNewmanOptions{
		NewmanCollection:  "localFile.txt",
		NewmanRunCommand:  "runcommand",
		CfAppsWithSecrets: false,
	}

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init

		utils := newExecuteNewmanMockUtils()

		// test
		err := runExecuteNewman(&allFineConfig, &utils)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, "/bin/sh", utils.executedShell)
		assert.Equal(t, "PATH=\\$PATH:~/.npm-global/bin newman runcommand --suppress-exit-code", utils.executedScript)
	})

	t.Run("happy path with fail on error", func(t *testing.T) {
		t.Parallel()
		// init

		utils := newExecuteNewmanMockUtils()
		fineConfig := allFineConfig
		fineConfig.FailOnError = true

		// test
		err := runExecuteNewman(&fineConfig, &utils)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, "/bin/sh", utils.executedShell)
		assert.Equal(t, "PATH=\\$PATH:~/.npm-global/bin newman runcommand", utils.executedScript)
	})

	t.Run("error on shell execution", func(t *testing.T) {
		t.Parallel()
		// init

		utils := newExecuteNewmanMockUtils()
		utils.errorOnFinalScriptExecution = true

		// test
		err := runExecuteNewman(&allFineConfig, &utils)

		// assert
		assert.EqualError(t, err, "The execution of the newman tests failed, see the log for details.: error on newman execution")
	})

	t.Run("error on newman installation", func(t *testing.T) {
		t.Parallel()
		// init

		utils := newExecuteNewmanMockUtils()
		utils.errorOnNewmanInstall = true

		// test
		err := runExecuteNewman(&allFineConfig, &utils)

		// assert
		assert.EqualError(t, err, "error installing newman: error on newman install")
	})

	t.Run("error on npm version logging", func(t *testing.T) {
		t.Parallel()
		// init

		utils := newExecuteNewmanMockUtils()
		utils.errorOnLoggingNpm = true

		// test
		err := runExecuteNewman(&allFineConfig, &utils)

		// assert
		assert.EqualError(t, err, "error logging npm version: error on RunExecutable")
	})

	t.Run("error on template resolution", func(t *testing.T) {
		t.Parallel()
		// init

		utils := newExecuteNewmanMockUtils()
		config := allFineConfig
		config.NewmanRunCommand = "this is my erroneous command {{.collectionDisplayName}"

		// test
		err := runExecuteNewman(&config, &utils)

		// assert
		assert.EqualError(t, err, "could not parse newman command template: template: template:1: unexpected \"}\" in operand")
	})

	t.Run("error on file search", func(t *testing.T) {
		t.Parallel()
		// init

		utils := newExecuteNewmanMockUtils()
		utils.filesToFind = nil

		// test
		err := runExecuteNewman(&allFineConfig, &utils)

		// assert
		assert.EqualError(t, err, "no collection found with pattern 'localFile.txt'")
	})

	t.Run("no newman file", func(t *testing.T) {
		t.Parallel()
		// init

		utils := newExecuteNewmanMockUtils()
		utils.errorOnGlob = true

		// test
		err := runExecuteNewman(&allFineConfig, &utils)

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

	t.Run("replace config Verbose", func(t *testing.T) {
		t.Parallel()

		config := executeNewmanOptions{
			NewmanRunCommand: "this is my fancy command {{.Config.Verbose}}",
			Verbose:          false,
		}

		cmd, err := resolveTemplate(&config, "theDisplayName")
		assert.NoError(t, err)
		assert.Equal(t, "this is my fancy command false", cmd)
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

func TestInstallNewman(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		utils := newExecuteNewmanMockUtils()

		err := installNewman("command", &utils)
		assert.NoError(t, err)
		assert.Equal(t, "/bin/sh", utils.executedShell)
		assert.Equal(t, "NPM_CONFIG_PREFIX=~/.npm-global command", utils.executedScript)
	})

	t.Run("error on run shell", func(t *testing.T) {
		t.Parallel()
		utils := newExecuteNewmanMockUtils()
		utils.errorOnRunShell = true

		err := installNewman("command", &utils)
		assert.EqualError(t, err, "error installing newman: error on RunShell")
	})
}

func TestLogVersions(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		utils := newExecuteNewmanMockUtils()

		err := logVersions(&utils)
		assert.NoError(t, err)
		assert.Equal(t, "npm", utils.executedExecutable)
		assert.Equal(t, "--version", utils.executedParams[0])
	})

	t.Run("error in node execution", func(t *testing.T) {
		utils := newExecuteNewmanMockUtils()
		utils.errorOnLoggingNode = true

		err := logVersions(&utils)
		assert.EqualError(t, err, "error logging node version: error on RunExecutable")
	})

	t.Run("error in npm execution", func(t *testing.T) {
		utils := newExecuteNewmanMockUtils()
		utils.errorOnLoggingNpm = true

		err := logVersions(&utils)
		assert.EqualError(t, err, "error logging npm version: error on RunExecutable")
		assert.Equal(t, "node", utils.executedExecutable)
		assert.Equal(t, "--version", utils.executedParams[0])
	})
}

func (e *executeNewmanMockUtils) Glob(string) (matches []string, err error) {
	if e.errorOnGlob {
		return nil, fmt.Errorf("error on Glob")
	}

	return e.filesToFind, nil
}

func (e *executeNewmanMockUtils) RunShell(shell, script string) error {
	if e.errorOnRunShell {
		return fmt.Errorf("error on RunShell")
	}
	if e.errorOnNewmanInstall && strings.Contains(script, "NPM_CONFIG_PREFIX=~/.npm-global") {
		return fmt.Errorf("error on newman install")
	}
	if e.errorOnFinalScriptExecution && strings.Contains(script, "PATH=\\$PATH:~/.npm-global/bin newman") {
		return fmt.Errorf("error on newman execution")
	}

	e.executedShell = shell
	e.executedScript = script
	return nil
}

func (e *executeNewmanMockUtils) RunExecutable(executable string, params ...string) error {
	if e.errorOnRunExecutable {
		return fmt.Errorf("error on RunExecutable")
	}
	if e.errorOnLoggingNode && executable == "node" && params[0] == "--version" {
		return fmt.Errorf("error on RunExecutable")
	}
	if e.errorOnLoggingNpm && executable == "npm" && params[0] == "--version" {
		return fmt.Errorf("error on RunExecutable")
	}

	e.executedExecutable = executable
	e.executedParams = params
	return nil
}
