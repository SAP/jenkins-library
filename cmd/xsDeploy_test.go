package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/mock"
	sliceUtils "github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

type FileUtilsMock struct {
	copiedFiles       []string
	fileThrowingError []string
	existingFiles     []string
	writtenFiles      map[string][]byte
}

func (f *FileUtilsMock) FileExists(path string) (bool, error) {
	if sliceUtils.ContainsString(f.fileThrowingError, path) {
		return false, errors.New("error on FileExists for " + path)
	}
	return sliceUtils.ContainsString(f.existingFiles, path), nil
}

func (f *FileUtilsMock) Copy(src, dest string) (int64, error) {
	f.copiedFiles = append(f.copiedFiles, fmt.Sprintf("%s->%s", src, dest))
	return 0, nil
}

func (f *FileUtilsMock) FileRead(string) ([]byte, error) {
	return []byte{}, nil
}

func (f *FileUtilsMock) FileWrite(path string, content []byte, _ os.FileMode) error {
	if len(f.writtenFiles) == 0 {
		f.writtenFiles = map[string][]byte{}
	}
	f.writtenFiles[path] = content
	return nil
}

func (f *FileUtilsMock) MkdirAll(string, os.FileMode) error {
	return nil
}

func (f *FileUtilsMock) Chmod(string, os.FileMode) error {
	return errors.New("not implemented. Func is only present in order to fulfill the interface contract. Needs to be adjusted in case it gets used")
}

func (f *FileUtilsMock) Abs(string) (string, error) {
	return "", errors.New("not implemented. Func is only present in order to fulfill the interface contract. Needs to be adjusted in case it gets used")
}

func (f *FileUtilsMock) Glob(string) (matches []string, err error) {
	return nil, errors.New("not implemented. Func is only present in order to fulfill the interface contract. Needs to be adjusted in case it gets used")
}

func (f *FileUtilsMock) TempDir(dir, pattern string) (directoryName string, err error) {
	return ioutil.TempDir(dir, pattern)
}

type xsDeployUtilsMock struct {
	envVariables map[string]string
}

func (x *xsDeployUtilsMock) Getenv(key string) string {
	if x.envVariables[key] != "" {
		return x.envVariables[key]
	}
	return os.Getenv(key)
}

func TestDeploy(t *testing.T) {
	myXsDeployOptions := xsDeployOptions{
		APIURL:                "https://example.org:12345",
		Username:              "me",
		Password:              "secretPassword",
		Org:                   "myOrg",
		Space:                 "mySpace",
		LoginOpts:             "--skip-ssl-validation",
		DeployOpts:            "--dummy-deploy-opts",
		XsSessionFile:         ".xs_session",
		Mode:                  "DEPLOY",
		Action:                "NONE",
		MtaPath:               "dummy.mtar",
		OperationIDLogPattern: `^.*xs bg-deploy -i (.*) -a.*$`,
	}

	cpeOut := xsDeployCommonPipelineEnvironment{}

	t.Run("Standard deploy succeeds", func(t *testing.T) {
		t.Parallel()

		var stdout string

		rStdout, wStdout := io.Pipe()

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			buf := new(bytes.Buffer)
			_, _ = io.Copy(buf, rStdout)
			stdout = buf.String()
			wg.Done()
		}()

		var removedFiles []string

		shellMockRunner := mock.ShellMockRunner{}
		fileUtilsMock := FileUtilsMock{
			existingFiles: []string{"dummy.mtar", ".xs_session"},
		}
		e := runXsDeploy(myXsDeployOptions, &cpeOut, &shellMockRunner, &fileUtilsMock, removeFilesFuncBuilder(&removedFiles), wStdout, &xsDeployUtilsMock{})

		_ = wStdout.Close()
		wg.Wait()
		_ = rStdout.Close()

		assert.NoError(t, e)

		// Contains --> we do not check for the shebang
		assert.Len(t, shellMockRunner.Calls, 3)
		assert.Contains(t, shellMockRunner.Calls[0], "xs login -a https://example.org:12345 -u me -p 'secretPassword' -o myOrg -s mySpace --skip-ssl-validation")
		assert.Contains(t, shellMockRunner.Calls[1], "xs deploy dummy.mtar --dummy-deploy-opts")
		assert.Contains(t, shellMockRunner.Calls[2], "xs logout")

		// xs session file needs to be removed at end during a normal deployment
		assert.Len(t, removedFiles, 1)
		assert.Contains(t, removedFiles, ".xs_session")

		assert.Len(t, fileUtilsMock.copiedFiles, 2)
		// We copy the xs session file to the workspace in order to be able to use the file later.
		// This happens directly after login
		// We copy the xs session file from the workspace to the home folder in order to be able to
		// use that file. This is important in case we rely on a login which happened e
		assert.Contains(t, fileUtilsMock.copiedFiles[0], "/.xs_session->.xs_session")
		assert.Contains(t, fileUtilsMock.copiedFiles[1], ".xs_session->")
		assert.Contains(t, fileUtilsMock.copiedFiles[1], "/.xs_session")

		// no password exposed
		assert.NotEmpty(t, stdout)
		assert.NotContains(t, stdout, myXsDeployOptions.Password)
	})

	t.Run("error on file remove", func(t *testing.T) {
		t.Parallel()

		remove := func(path string) error {
			return errors.New("error removing file " + path)
		}
		fileUtilsMock := FileUtilsMock{
			existingFiles: []string{"dummy.mtar", ".xs_session"},
		}
		e := runXsDeploy(myXsDeployOptions, &cpeOut, &mock.ShellMockRunner{}, &fileUtilsMock, remove, ioutil.Discard, &xsDeployUtilsMock{})

		assert.EqualError(t, e, "error removing file .xs_session")
	})

	t.Run("error on logout", func(t *testing.T) {
		t.Parallel()

		shellMock := mock.ShellMockRunner{}
		shellMock.ShouldFailOnCommand = map[string]error{}
		shellMock.ShouldFailOnCommand["#!/bin/bash\nxs logout"] = errors.New("error on logout")

		var removedFiles []string

		fileUtilsMock := FileUtilsMock{
			existingFiles: []string{"dummy.mtar", ".xs_session"},
		}
		e := runXsDeploy(myXsDeployOptions, &cpeOut, &shellMock, &fileUtilsMock, removeFilesFuncBuilder(&removedFiles), ioutil.Discard, &xsDeployUtilsMock{})

		assert.EqualError(t, e, "error on logout")
	})

	t.Run("handle log but no log files exists", func(t *testing.T) {
		t.Parallel()

		fileUtilsMock := FileUtilsMock{
			existingFiles: []string{"dummy.mtar", ".xs_session"},
		}

		tempDir, err := fileUtilsMock.TempDir(".", "logTest")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		err = os.Mkdir(filepath.Join(tempDir, ".xs_logs"), 0777)
		assert.NoError(t, err)

		shellMock := mock.ShellMockRunner{}
		shellMock.ShouldFailOnCommand = map[string]error{}
		shellMock.ShouldFailOnCommand["#!/bin/bash\nxs logout"] = errors.New("error on logout")

		var removedFiles []string

		deployUtilsMock := xsDeployUtilsMock{
			envVariables: map[string]string{
				"HOME": tempDir,
			},
		}
		e := runXsDeploy(myXsDeployOptions, &cpeOut, &shellMock, &fileUtilsMock, removeFilesFuncBuilder(&removedFiles), ioutil.Discard, &deployUtilsMock)

		assert.EqualError(t, e, "error on logout")
	})

	t.Run("handle log with file", func(t *testing.T) {
		t.Parallel()

		fileUtilsMock := FileUtilsMock{
			existingFiles: []string{"dummy.mtar", ".xs_session"},
		}

		tempDir, err := fileUtilsMock.TempDir(".", "logTest")
		assert.NoError(t, err)
		defer func() {
			err = os.RemoveAll(tempDir)
			assert.NoError(t, err)
		}()

		logDirectory := filepath.Join(tempDir, ".xs_logs")

		err = os.Mkdir(logDirectory, 0777)
		assert.NoError(t, err)
		err = ioutil.WriteFile(filepath.Join(logDirectory, "logs1.txt"), []byte("Hello World. These are logs."), 0777)
		assert.NoError(t, err)

		shellMock := mock.ShellMockRunner{}
		shellMock.ShouldFailOnCommand = map[string]error{}
		shellMock.ShouldFailOnCommand["#!/bin/bash\nxs logout"] = errors.New("error on logout")

		var removedFiles []string

		deployUtilsMock := xsDeployUtilsMock{
			envVariables: map[string]string{
				"HOME": tempDir,
			},
		}
		e := runXsDeploy(myXsDeployOptions, &cpeOut, &shellMock, &fileUtilsMock, removeFilesFuncBuilder(&removedFiles), ioutil.Discard, &deployUtilsMock)

		assert.EqualError(t, e, "error on logout")
	})

	t.Run("error on file read dummy.mtar", func(t *testing.T) {
		t.Parallel()

		fileUtils := FileUtilsMock{
			existingFiles: []string{"dummy.mtar", ".xs_session"},
		}
		fileUtils.fileThrowingError = []string{"dummy.mtar"}
		e := runXsDeploy(myXsDeployOptions, &cpeOut, &mock.ShellMockRunner{}, &fileUtils, removeFilesFuncBuilder(&[]string{}), ioutil.Discard, &xsDeployUtilsMock{})

		assert.EqualError(t, e, "error on FileExists for dummy.mtar")
	})

	t.Run("error on file read xs_session", func(t *testing.T) {
		t.Parallel()

		fileUtils := FileUtilsMock{
			existingFiles: []string{"dummy.mtar", ".xs_session"},
		}
		fileUtils.fileThrowingError = []string{".xs_session"}
		e := runXsDeploy(myXsDeployOptions, &cpeOut, &mock.ShellMockRunner{}, &fileUtils, removeFilesFuncBuilder(&[]string{}), ioutil.Discard, &xsDeployUtilsMock{})

		assert.EqualError(t, e, "error on FileExists for .xs_session")
	})

	t.Run("xs_session does not exist", func(t *testing.T) {
		t.Parallel()

		fileUtils := FileUtilsMock{
			existingFiles: []string{"dummy.mtar", ".xs_session"},
		}
		fileUtils.existingFiles = []string{"dummy.mtar"}
		e := runXsDeploy(myXsDeployOptions, &cpeOut, &mock.ShellMockRunner{}, &fileUtils, removeFilesFuncBuilder(&[]string{}), ioutil.Discard, &xsDeployUtilsMock{})

		assert.EqualError(t, e, "xs session file does not exist (.xs_session)")
	})

	t.Run("invalid deploy mode", func(t *testing.T) {
		t.Parallel()

		options := myXsDeployOptions
		options.Mode = "ERROR"
		fileUtilsMock := FileUtilsMock{
			existingFiles: []string{"dummy.mtar", ".xs_session"},
		}
		e := runXsDeploy(options, &cpeOut, &mock.ShellMockRunner{}, &fileUtilsMock, removeFilesFuncBuilder(&[]string{}), ioutil.Discard, &xsDeployUtilsMock{})

		assert.EqualError(t, e, "Extracting mode failed: 'ERROR': Unknown DeployMode: 'ERROR'")
	})

	t.Run("no deploy mode", func(t *testing.T) {
		t.Parallel()

		options := myXsDeployOptions
		options.Mode = "NONE"
		shellMockRunner := mock.ShellMockRunner{}
		fileUtilsMock := FileUtilsMock{
			existingFiles: []string{"dummy.mtar", ".xs_session"},
		}
		e := runXsDeploy(options, &cpeOut, &shellMockRunner, &fileUtilsMock, removeFilesFuncBuilder(&[]string{}), ioutil.Discard, &xsDeployUtilsMock{})

		assert.NoError(t, e)
		assert.Len(t, shellMockRunner.Calls, 0)
	})

	t.Run("invalid action", func(t *testing.T) {
		t.Parallel()

		options := myXsDeployOptions
		options.Action = "INVALID"
		fileUtilsMock := FileUtilsMock{
			existingFiles: []string{"dummy.mtar", ".xs_session"},
		}
		e := runXsDeploy(options, &cpeOut, &mock.ShellMockRunner{}, &fileUtilsMock, removeFilesFuncBuilder(&[]string{}), ioutil.Discard, &xsDeployUtilsMock{})

		assert.EqualError(t, e, "Extracting action failed: 'INVALID': Unknown Action: 'INVALID'")
	})

	t.Run("Invalid deploy command", func(t *testing.T) {
		t.Parallel()
		_, err := NoDeploy.GetDeployCommand()
		assert.EqualError(t, err, "Invalid deploy mode: 'NONE'.")
	})

	t.Run("Standard deploy fails, deployable missing", func(t *testing.T) {
		t.Parallel()

		testOptions := myXsDeployOptions
		// this file is not denoted in the file exists mock
		testOptions.MtaPath = "doesNotExist"

		fileUtilsMock := FileUtilsMock{
			existingFiles: []string{"dummy.mtar", ".xs_session"},
		}
		e := runXsDeploy(testOptions, &cpeOut, &mock.ShellMockRunner{}, &fileUtilsMock, removeFilesFuncBuilder(&[]string{}), ioutil.Discard, &xsDeployUtilsMock{})
		assert.EqualError(t, e, "Deployable 'doesNotExist' does not exist")
	})

	t.Run("Standard deploy fails, action provided", func(t *testing.T) {
		t.Parallel()

		testOptions := myXsDeployOptions
		testOptions.Action = "RETRY"

		fileUtilsMock := FileUtilsMock{
			existingFiles: []string{"dummy.mtar", ".xs_session"},
		}
		e := runXsDeploy(testOptions, &cpeOut, &mock.ShellMockRunner{}, &fileUtilsMock, removeFilesFuncBuilder(&[]string{}), ioutil.Discard, &xsDeployUtilsMock{})
		assert.EqualError(t, e, "Cannot perform action 'RETRY' in mode 'DEPLOY'. Only action 'NONE' is allowed.")
	})

	t.Run("Standard deploy fails, error from underlying process", func(t *testing.T) {
		t.Parallel()

		mockRunner := mock.ShellMockRunner{}
		mockRunner.ShouldFailOnCommand = map[string]error{"#!/bin/bash\nxs login -a https://example.org:12345 -u me -p 'secretPassword' -o myOrg -s mySpace --skip-ssl-validation\n": errors.New("error from underlying process")}

		fileUtilsMock := FileUtilsMock{
			existingFiles: []string{"dummy.mtar", ".xs_session"},
		}
		e := runXsDeploy(myXsDeployOptions, &cpeOut, &mockRunner, &fileUtilsMock, removeFilesFuncBuilder(&[]string{}), ioutil.Discard, &xsDeployUtilsMock{})
		assert.EqualError(t, e, "error from underlying process")
	})

	t.Run("BG deploy succeeds", func(t *testing.T) {
		t.Parallel()

		shellMockRunner := mock.ShellMockRunner{}
		shellMockRunner.StdoutReturn = make(map[string]string)
		shellMockRunner.StdoutReturn[".*xs bg-deploy.*"] = "Use \"xs bg-deploy -i 1234 -a resume\" to resume the process.\n"

		testOptions := myXsDeployOptions
		testOptions.Mode = "BG_DEPLOY"

		fileUtilsMock := FileUtilsMock{
			existingFiles: []string{"dummy.mtar", ".xs_session"},
		}
		e := runXsDeploy(testOptions, &cpeOut, &shellMockRunner, &fileUtilsMock, removeFilesFuncBuilder(&[]string{}), ioutil.Discard, &xsDeployUtilsMock{})

		if assert.NoError(t, e) {
			if assert.Len(t, (shellMockRunner).Calls, 2) { // There are two entries --> no logout in this case.
				assert.Contains(t, shellMockRunner.Calls[0], "xs login")
				assert.Contains(t, shellMockRunner.Calls[1], "xs bg-deploy dummy.mtar --dummy-deploy-opts")
			}
		}
	})

	t.Run("BG deploy fails, missing operationID", func(t *testing.T) {
		t.Parallel()

		shellMockRunner := mock.ShellMockRunner{}
		shellMockRunner.StdoutReturn = make(map[string]string)
		shellMockRunner.StdoutReturn[".*bg_deploy.*"] = "There is no operationID ...\n"

		testOptions := myXsDeployOptions
		testOptions.Mode = "BG_DEPLOY"

		fileUtilsMock := FileUtilsMock{
			existingFiles: []string{"dummy.mtar", ".xs_session"},
		}
		e := runXsDeploy(testOptions, &cpeOut, &shellMockRunner, &fileUtilsMock, removeFilesFuncBuilder(&[]string{}), ioutil.Discard, &xsDeployUtilsMock{})
		assert.EqualError(t, e, "No operationID found")
	})

	t.Run("BG deploy abort succeeds", func(t *testing.T) {
		t.Parallel()

		testOptions := myXsDeployOptions
		testOptions.Mode = "BG_DEPLOY"
		testOptions.Action = "ABORT"
		testOptions.OperationID = "12345"

		shellMockRunner := mock.ShellMockRunner{}
		fileUtilsMock := FileUtilsMock{
			existingFiles: []string{"dummy.mtar", ".xs_session"},
		}
		e := runXsDeploy(testOptions, &cpeOut, &shellMockRunner, &fileUtilsMock, removeFilesFuncBuilder(&[]string{}), ioutil.Discard, &xsDeployUtilsMock{})

		if assert.NoError(t, e) {
			if assert.Len(t, shellMockRunner.Calls, 2) { // There is no login --> we have two calls
				assert.Contains(t, shellMockRunner.Calls[0], "xs bg-deploy -i 12345 -a abort")
				assert.Contains(t, shellMockRunner.Calls[1], "xs logout")
			}

		}
	})

	t.Run("BG deploy resume succeeds", func(t *testing.T) {
		t.Parallel()

		testOptions := myXsDeployOptions
		testOptions.Mode = "BG_DEPLOY"
		testOptions.Action = "RESUME"
		testOptions.OperationID = "12345"

		shellMockRunner := mock.ShellMockRunner{}
		fileUtilsMock := FileUtilsMock{
			existingFiles: []string{"dummy.mtar", ".xs_session"},
		}
		e := runXsDeploy(testOptions, &cpeOut, &shellMockRunner, &fileUtilsMock, removeFilesFuncBuilder(&[]string{}), ioutil.Discard, &xsDeployUtilsMock{})

		if assert.NoError(t, e) {
			if assert.Len(t, shellMockRunner.Calls, 2) { // There is no login --> we have two calls
				assert.Contains(t, shellMockRunner.Calls[0], "xs bg-deploy -i 12345 -a resume")
				assert.Contains(t, shellMockRunner.Calls[1], "xs logout")
			}

		}
	})

	t.Run("BG deploy abort fails due to missing operationId", func(t *testing.T) {
		t.Parallel()

		testOptions := myXsDeployOptions
		testOptions.Mode = "BG_DEPLOY"
		testOptions.Action = "ABORT"

		fileUtilsMock := FileUtilsMock{
			existingFiles: []string{"dummy.mtar", ".xs_session"},
		}
		e := runXsDeploy(testOptions, &cpeOut, &mock.ShellMockRunner{}, &fileUtilsMock, removeFilesFuncBuilder(&[]string{}), ioutil.Discard, &xsDeployUtilsMock{})
		assert.EqualError(t, e, "OperationID was not provided. This is required for action 'ABORT'.")
	})
}

func TestRetrieveOperationID(t *testing.T) {
	t.Parallel()
	operationID := retrieveOperationID(`
	Uploading 1 files:
        myFolder/dummy.mtar
	File upload finished

	Detected MTA schema version: "3.1.0"
	Detected deploy target as "myOrg mySpace"
	Detected deployed MTA with ID "my_mta" and version "0.0.1"
	Deployed MTA color: blue
	New MTA color: green
	Detected new MTA version: "0.0.1"
	Deployed MTA version: 0.0.1
	Service "xxx" is not modified and will not be updated
	Creating application "db-green" from MTA module "xx"...
	Uploading application "xx-green"...
	Staging application "xx-green"...
	Application "xx-green" staged
	Executing task "deploy" on application "xx-green"...
	Task execution status: succeeded
	Process has entered validation phase. After testing your new deployment you can resume or abort the process.
	Use "xs bg-deploy -i 1234 -a resume" to resume the process.
	Use "xs bg-deploy -i 1234 -a abort" to abort the process.
	Hint: Use the '--no-confirm' option of the bg-deploy command to skip this phase.
	`, `^.*xs bg-deploy -i (.*) -a.*$`)

	assert.Equal(t, "1234", operationID)
}

func removeFilesFuncBuilder(removedFiles *[]string) func(path string) error {
	return func(path string) error {
		*removedFiles = append(*removedFiles, path)
		return nil
	}
}
