package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type FileUtilsMock struct {
	*mock.FilesMock
	copiedFiles []string
}

func (f *FileUtilsMock) FileExists(path string) (bool, error) {
	return path == "dummy.mtar" || path == ".xs_session", nil
}

func (f *FileUtilsMock) Copy(src, dest string) (int64, error) {
	f.copiedFiles = append(f.copiedFiles, fmt.Sprintf("%s->%s", src, dest))
	return 0, nil
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

	s := mock.ShellMockRunner{}

	var removedFiles []string

	cpeOut := xsDeployCommonPipelineEnvironment{}
	fileUtilsMock := FileUtilsMock{}

	fRemove := func(path string) error {
		removedFiles = append(removedFiles, path)
		return nil
	}

	var stdout string

	t.Run("Standard deploy succeeds", func(t *testing.T) {

		defer func() {
			fileUtilsMock.copiedFiles = nil
			removedFiles = nil
			s.Calls = nil
			stdout = ""
		}()

		rStdout, wStdout := io.Pipe()

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			buf := new(bytes.Buffer)
			if _, err := io.Copy(buf, rStdout); err != nil {
				log.Entry().Warning("failed to copy buffer")
			}
			stdout = buf.String()
			wg.Done()
		}()

		e := runXsDeploy(myXsDeployOptions, &cpeOut, &s, &fileUtilsMock, fRemove, wStdout)

		wStdout.Close()
		wg.Wait()

		assert.NoError(t, e)

		t.Run("Standard checks", func(t *testing.T) {
			// Contains --> we do not check for the shebang
			assert.Contains(t, s.Calls[0], "xs login -a https://example.org:12345 -u me -p 'secretPassword' -o myOrg -s mySpace --skip-ssl-validation")
			assert.Contains(t, s.Calls[1], "xs deploy dummy.mtar --dummy-deploy-opts")
			assert.Contains(t, s.Calls[2], "xs logout")
			assert.Len(t, s.Calls, 3)

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
		})

		t.Run("Password not exposed", func(t *testing.T) {
			assert.NotEmpty(t, stdout)
			assert.NotContains(t, stdout, myXsDeployOptions.Password)
		})
	})

	t.Run("Standard deploy fails, deployable missing", func(t *testing.T) {

		defer func() {
			fileUtilsMock.copiedFiles = nil
			removedFiles = nil
			s.Calls = nil
		}()

		oldMtaPath := myXsDeployOptions.MtaPath

		defer func() {
			myXsDeployOptions.MtaPath = oldMtaPath
		}()

		// this file is not denoted in the file exists mock
		myXsDeployOptions.MtaPath = "doesNotExist"

		e := runXsDeploy(myXsDeployOptions, &cpeOut, &s, &fileUtilsMock, fRemove, io.Discard)
		assert.EqualError(t, e, "Deployable 'doesNotExist' does not exist")
	})

	t.Run("Standard deploy fails, action provided", func(t *testing.T) {

		defer func() {
			fileUtilsMock.copiedFiles = nil
			removedFiles = nil
			s.Calls = nil
		}()

		myXsDeployOptions.Action = "RETRY"
		defer func() {
			myXsDeployOptions.Action = "NONE"
		}()

		e := runXsDeploy(myXsDeployOptions, &cpeOut, &s, &fileUtilsMock, fRemove, io.Discard)
		assert.EqualError(t, e, "Cannot perform action 'RETRY' in mode 'DEPLOY'. Only action 'NONE' is allowed.")
	})

	t.Run("Standard deploy fails, error from underlying process", func(t *testing.T) {

		defer func() {
			fileUtilsMock.copiedFiles = nil
			removedFiles = nil
			s.Calls = nil
			s.ShouldFailOnCommand = nil
		}()

		s.ShouldFailOnCommand = map[string]error{"#!/bin/bash\nxs login -a https://example.org:12345 -u me -p 'secretPassword' -o myOrg -s mySpace --skip-ssl-validation\n": errors.New("Error from underlying process")}

		e := runXsDeploy(myXsDeployOptions, &cpeOut, &s, &fileUtilsMock, fRemove, io.Discard)
		assert.EqualError(t, e, "Error from underlying process")
	})

	t.Run("BG deploy succeeds", func(t *testing.T) {

		defer func() {
			fileUtilsMock.copiedFiles = nil
			removedFiles = nil
			s.Calls = nil
			s.StdoutReturn = make(map[string]string)
		}()

		s.StdoutReturn = make(map[string]string)
		s.StdoutReturn[".*xs bg-deploy.*"] = "Use \"xs bg-deploy -i 1234 -a resume\" to resume the process.\n"

		oldMode := myXsDeployOptions.Mode

		defer func() {
			myXsDeployOptions.Mode = oldMode
		}()

		myXsDeployOptions.Mode = "BG_DEPLOY"

		e := runXsDeploy(myXsDeployOptions, &cpeOut, &s, &fileUtilsMock, fRemove, io.Discard)

		if assert.NoError(t, e) {
			if assert.Len(t, s.Calls, 2) { // There are two entries --> no logout in this case.
				assert.Contains(t, s.Calls[0], "xs login")
				assert.Contains(t, s.Calls[1], "xs bg-deploy dummy.mtar --dummy-deploy-opts")
			}
		}
	})

	t.Run("BG deploy fails, missing operationID", func(t *testing.T) {

		s.StdoutReturn = make(map[string]string)
		s.StdoutReturn[".*bg_deploy.*"] = "There is no operationID ...\n"
		defer func() {
			fileUtilsMock.copiedFiles = nil
			removedFiles = nil
			s.Calls = nil
			s.StdoutReturn = make(map[string]string)
		}()

		oldMode := myXsDeployOptions.Mode

		defer func() {
			myXsDeployOptions.Mode = oldMode
		}()

		myXsDeployOptions.Mode = "BG_DEPLOY"

		e := runXsDeploy(myXsDeployOptions, &cpeOut, &s, &fileUtilsMock, fRemove, io.Discard)
		assert.EqualError(t, e, "No operationID found")
	})

	t.Run("BG deploy abort succeeds", func(t *testing.T) {

		defer func() {
			fileUtilsMock.copiedFiles = nil
			removedFiles = nil
			s.Calls = nil
		}()

		oldMode := myXsDeployOptions.Mode
		oldAction := myXsDeployOptions.Action

		defer func() {
			myXsDeployOptions.Mode = oldMode
			myXsDeployOptions.Action = oldAction
			myXsDeployOptions.OperationID = ""
		}()

		myXsDeployOptions.Mode = "BG_DEPLOY"
		myXsDeployOptions.Action = "ABORT"
		myXsDeployOptions.OperationID = "12345"

		e := runXsDeploy(myXsDeployOptions, &cpeOut, &s, &fileUtilsMock, fRemove, io.Discard)

		if assert.NoError(t, e) {
			if assert.Len(t, s.Calls, 2) { // There is no login --> we have two calls
				assert.Contains(t, s.Calls[0], "xs bg-deploy -i 12345 -a abort")
				assert.Contains(t, s.Calls[1], "xs logout")
			}

		}
	})

	t.Run("BG deploy abort fails due to missing operationId", func(t *testing.T) {

		defer func() {
			fileUtilsMock.copiedFiles = nil
			removedFiles = nil
			s.Calls = nil
		}()

		oldMode := myXsDeployOptions.Mode
		oldAction := myXsDeployOptions.Action

		defer func() {
			myXsDeployOptions.Mode = oldMode
			myXsDeployOptions.Action = oldAction
		}()

		myXsDeployOptions.Mode = "BG_DEPLOY"
		myXsDeployOptions.Action = "ABORT"

		e := runXsDeploy(myXsDeployOptions, &cpeOut, &s, &fileUtilsMock, fRemove, io.Discard)
		assert.EqualError(t, e, "OperationID was not provided. This is required for action 'ABORT'.")
	})
}

func TestRetrieveOperationID(t *testing.T) {
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
