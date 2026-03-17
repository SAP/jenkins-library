//go:build unit
// +build unit

package cmd

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"errors"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type executedExecutables struct {
	executable string
	params     []string
}

type newmanExecuteMockUtils struct {
	// *mock.ExecMockRunner
	// *mock.FilesMock
	errorOnGlob            bool
	errorOnNewmanInstall   bool
	errorOnRunShell        bool
	errorOnNewmanExecution bool
	errorOnLoggingNode     bool
	errorOnLoggingNpm      bool
	executedExecutables    []executedExecutables
	filesToFind            []string
	commandIndex           int
}

func newNewmanExecuteMockUtils() newmanExecuteMockUtils {
	return newmanExecuteMockUtils{
		filesToFind: []string{"localFile.json", "localFile2.json"},
	}
}

func TestRunNewmanExecute(t *testing.T) {
	t.Parallel()

	allFineConfig := newmanExecuteOptions{
		NewmanCollection:     "**.json",
		NewmanEnvironment:    "env.json",
		NewmanGlobals:        "globals.json",
		NewmanInstallCommand: "npm install newman --global --quiet",
		NewmanRunCommand:     "run {{.NewmanCollection}} --environment {{.Config.NewmanEnvironment}} --globals {{.Config.NewmanGlobals}} --reporters junit,html --reporter-junit-export target/newman/TEST-{{.CollectionDisplayName}}.xml --reporter-html-export target/newman/TEST-{{.CollectionDisplayName}}.html",
	}

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init

		utils := newNewmanExecuteMockUtils()

		// test
		err := runNewmanExecute(&allFineConfig, &utils)

		// assert
		assert.NoError(t, err)
		assert.Contains(t, utils.executedExecutables, executedExecutables{executable: "node", params: []string{"--version"}})
		assert.Contains(t, utils.executedExecutables, executedExecutables{executable: "npm", params: []string{"--version"}})
		assert.Contains(t, utils.executedExecutables, executedExecutables{executable: "npm", params: []string{"install", "newman", "--global", "--quiet", "--prefix=~/.npm-global"}})
		assert.Contains(t, utils.executedExecutables, executedExecutables{executable: filepath.FromSlash("/home/node/.npm-global/bin/newman"), params: []string{"run", "localFile.json", "--environment", "env.json", "--globals", "globals.json", "--reporters", "junit,html", "--reporter-junit-export", "target/newman/TEST-localFile.xml", "--reporter-html-export", "target/newman/TEST-localFile.html", "--suppress-exit-code"}})
		assert.Contains(t, utils.executedExecutables, executedExecutables{executable: filepath.FromSlash("/home/node/.npm-global/bin/newman"), params: []string{"run", "localFile2.json", "--environment", "env.json", "--globals", "globals.json", "--reporters", "junit,html", "--reporter-junit-export", "target/newman/TEST-localFile2.xml", "--reporter-html-export", "target/newman/TEST-localFile2.html", "--suppress-exit-code"}})
	})

	t.Run("happy path with fail on error", func(t *testing.T) {
		t.Parallel()
		// init

		utils := newNewmanExecuteMockUtils()
		fineConfig := allFineConfig
		fineConfig.FailOnError = true

		// test
		err := runNewmanExecute(&fineConfig, &utils)

		// assert
		assert.NoError(t, err)
		assert.Contains(t, utils.executedExecutables, executedExecutables{executable: "node", params: []string{"--version"}})
		assert.Contains(t, utils.executedExecutables, executedExecutables{executable: "npm", params: []string{"--version"}})
		assert.Contains(t, utils.executedExecutables, executedExecutables{executable: "npm", params: []string{"install", "newman", "--global", "--quiet", "--prefix=~/.npm-global"}})
		assert.Contains(t, utils.executedExecutables, executedExecutables{executable: filepath.FromSlash("/home/node/.npm-global/bin/newman"), params: []string{"run", "localFile.json", "--environment", "env.json", "--globals", "globals.json", "--reporters", "junit,html", "--reporter-junit-export", "target/newman/TEST-localFile.xml", "--reporter-html-export", "target/newman/TEST-localFile.html"}})
		assert.Contains(t, utils.executedExecutables, executedExecutables{executable: filepath.FromSlash("/home/node/.npm-global/bin/newman"), params: []string{"run", "localFile2.json", "--environment", "env.json", "--globals", "globals.json", "--reporters", "junit,html", "--reporter-junit-export", "target/newman/TEST-localFile2.xml", "--reporter-html-export", "target/newman/TEST-localFile2.html"}})
	})

	t.Run("error on newman execution", func(t *testing.T) {
		t.Parallel()
		// init

		utils := newNewmanExecuteMockUtils()
		utils.errorOnNewmanExecution = true

		// test
		err := runNewmanExecute(&allFineConfig, &utils)

		// assert
		assert.EqualError(t, err, "The execution of the newman tests failed, see the log for details.: error on newman execution")
	})

	t.Run("error on newman installation", func(t *testing.T) {
		t.Parallel()
		// init

		utils := newNewmanExecuteMockUtils()
		utils.errorOnNewmanInstall = true

		// test
		err := runNewmanExecute(&allFineConfig, &utils)

		// assert
		assert.EqualError(t, err, "error installing newman: error on newman install")
	})

	t.Run("error on npm version logging", func(t *testing.T) {
		t.Parallel()
		// init

		utils := newNewmanExecuteMockUtils()
		utils.errorOnLoggingNpm = true

		// test
		err := runNewmanExecute(&allFineConfig, &utils)

		// assert
		assert.EqualError(t, err, "error logging npm version: error on RunExecutable")
	})

	t.Run("error on template resolution", func(t *testing.T) {
		t.Parallel()
		// init

		utils := newNewmanExecuteMockUtils()
		config := allFineConfig
		config.NewmanRunCommand = "this is my erroneous command {{.collectionDisplayName}"

		// test
		err := runNewmanExecute(&config, &utils)

		// assert
		assert.EqualError(t, err, "could not parse newman command template: template: template:1: bad character U+007D '}'")
	})

	t.Run("error on file search", func(t *testing.T) {
		t.Parallel()
		// init

		utils := newNewmanExecuteMockUtils()
		utils.filesToFind = nil

		// test
		err := runNewmanExecute(&allFineConfig, &utils)

		// assert
		assert.EqualError(t, err, "no collection found with pattern '**.json'")
	})

	t.Run("no newman file", func(t *testing.T) {
		t.Parallel()
		// init

		utils := newNewmanExecuteMockUtils()
		utils.errorOnGlob = true

		// test
		err := runNewmanExecute(&allFineConfig, &utils)

		// assert
		assert.EqualError(t, err, "Could not execute global search for '**.json': error on Glob")
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

	t.Run("directory with dot prefix", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(".dir1", "dir2", "dir3", "file.json")
		result := defineCollectionDisplayName(path)
		assert.Equal(t, "dir1_dir2_dir3_file", result)
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

		// config := newmanExecuteOptions{NewmanRunCommand: "this is my fancy command"}
		config := newmanExecuteOptions{RunOptions: []string{"this", "is", "my", "fancy", "command"}}

		cmd, err := resolveTemplate(&config, "collectionsDisplayName")
		assert.NoError(t, err)
		assert.Equal(t, []string{"this", "is", "my", "fancy", "command"}, cmd)
	})

	t.Run("replace display name", func(t *testing.T) {
		t.Parallel()

		config := newmanExecuteOptions{RunOptions: []string{"this", "is", "my", "fancy", "command", "{{.CollectionDisplayName}}"}}

		cmd, err := resolveTemplate(&config, "theDisplayName")
		assert.NoError(t, err)
		assert.Equal(t, []string{"this", "is", "my", "fancy", "command", "theDisplayName"}, cmd)
	})

	t.Run("get environment variable", func(t *testing.T) {
		t.Parallel()

		temporaryEnvVarName := uuid.New().String()
		os.Setenv(temporaryEnvVarName, "myEnvVar")
		defer os.Unsetenv(temporaryEnvVarName)
		config := newmanExecuteOptions{RunOptions: []string{"this", "is", "my", "fancy", "command", "with", "--env-var", "{{getenv \"" + temporaryEnvVarName + "\"}}"}}

		cmd, err := resolveTemplate(&config, "collectionsDisplayName")
		assert.NoError(t, err)
		assert.Equal(t, []string{"this", "is", "my", "fancy", "command", "with", "--env-var", "myEnvVar"}, cmd)
	})

	t.Run("error when parameter cannot be resolved", func(t *testing.T) {
		t.Parallel()

		config := newmanExecuteOptions{RunOptions: []string{"this", "is", "my", "fancy", "command", "{{.collectionDisplayName}}"}}

		_, err := resolveTemplate(&config, "theDisplayName")
		assert.EqualError(t, err, "error on executing template: template: template:1:2: executing \"template\" at <.collectionDisplayName>: can't evaluate field collectionDisplayName in type cmd.TemplateConfig")
	})

	t.Run("error when template cannot be parsed", func(t *testing.T) {
		t.Parallel()

		config := newmanExecuteOptions{RunOptions: []string{"this", "is", "my", "fancy", "command", "{{.collectionDisplayName}"}}

		_, err := resolveTemplate(&config, "theDisplayName")
		assert.EqualError(t, err, "could not parse newman command template: template: template:1: bad character U+007D '}'")
	})
}

func TestLogVersions(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		utils := newNewmanExecuteMockUtils()

		err := logVersions(&utils)
		assert.NoError(t, err)
		assert.Contains(t, utils.executedExecutables, executedExecutables{executable: "npm", params: []string{"--version"}})
	})

	t.Run("error in node execution", func(t *testing.T) {
		utils := newNewmanExecuteMockUtils()
		utils.errorOnLoggingNode = true

		err := logVersions(&utils)
		assert.EqualError(t, err, "error logging node version: error on RunExecutable")
	})

	t.Run("error in npm execution", func(t *testing.T) {
		utils := newNewmanExecuteMockUtils()
		utils.errorOnLoggingNpm = true

		err := logVersions(&utils)
		assert.EqualError(t, err, "error logging npm version: error on RunExecutable")
		assert.Contains(t, utils.executedExecutables, executedExecutables{executable: "node", params: []string{"--version"}})
	})
}

func (e *newmanExecuteMockUtils) Glob(string) (matches []string, err error) {
	if e.errorOnGlob {
		return nil, errors.New("error on Glob")
	}

	return e.filesToFind, nil
}

func (e *newmanExecuteMockUtils) RunExecutable(executable string, params ...string) error {
	if e.errorOnRunShell {
		return errors.New("error on RunExecutable")
	}
	if e.errorOnLoggingNode && executable == "node" && params[0] == "--version" {
		return errors.New("error on RunExecutable")
	}
	if e.errorOnLoggingNpm && executable == "npm" && params[0] == "--version" {
		return errors.New("error on RunExecutable")
	}
	if e.errorOnNewmanExecution && strings.Contains(executable, "newman") {
		return errors.New("error on newman execution")
	}
	if e.errorOnNewmanInstall && slices.Contains(params, "install") {
		return errors.New("error on newman install")
	}

	length := len(e.executedExecutables)
	if length < e.commandIndex+1 {
		e.executedExecutables = append(e.executedExecutables, executedExecutables{})
		length++
	}

	e.executedExecutables[length-1].executable = executable
	e.executedExecutables[length-1].params = params
	e.commandIndex++

	return nil
}

func (e *newmanExecuteMockUtils) Getenv(key string) string {
	if key == "HOME" {
		return "/home/node"
	}
	return ""
}
