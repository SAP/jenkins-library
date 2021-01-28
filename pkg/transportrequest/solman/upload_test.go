package solman

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSolmanUpload(t *testing.T) {

	f := &mock.FilesMock{}
	f.AddFile("myDeployable.xxx", []byte(""))

	defaultUploadAction := UploadAction{
		Connection: Connection{
			Endpoint: "https://example.org/solman",
			User:     "me",
			Password: "******",
		},
		ChangeDocumentID:   "123456",
		TransportRequestID: "000K11111111",
		ApplicationID:      "MY_APP",
		File:               "myDeployable.xxx",
		CMOpts:             []string{"-Dmyprop1=abc", "-Dmyprop2=def"},
	}

	t.Run("Deployable does not exist", func(t *testing.T) {

		uploadActionFileMissing := defaultUploadAction
		uploadActionFileMissing.File = "myMissingDeployable.xxx"
		e := &mock.ExecMockRunner{}

		err := uploadActionFileMissing.Perform(f, e)

		assert.EqualError(t, err, "cannot upload artifact 'myMissingDeployable.xxx': file 'myMissingDeployable.xxx' does not exist")
	})

	t.Run("Straight forward", func(t *testing.T) {

		e := &mock.ExecMockRunner{}

		err := defaultUploadAction.Perform(f, e)

		if assert.NoError(t, err) {
			assert.Len(t, e.Calls, 1)
			assert.Equal(t, mock.ExecCall{
				Exec: "cmclient",
				Params: []string{
					"--endpoint", "https://example.org/solman",
					"--user", "me",
					"--password", "******",
					"--backend-type", "SOLMAN",
					"upload-file-to-transport",
					"-cID", "123456",
					"-tID", "000K11111111",
					"MY_APP",
					"myDeployable.xxx",
				},
			}, e.Calls[0])
			assert.Equal(t, []string{"CMCLIENT_OPTS=-Dmyprop1=abc -Dmyprop2=def"}, e.Env)
		}
	})

	t.Run("Missing parameters", func(t *testing.T) {
		e := &mock.ExecMockRunner{}
		uploadAction := defaultUploadAction
		uploadAction.Connection.Endpoint = ""
		uploadAction.TransportRequestID = ""
		err := uploadAction.Perform(f, e)
		if assert.Error(t, err) {
			// we should not rely on the order of the missing parameters
			assert.Contains(t, err.Error(), "cannot upload artifact 'myDeployable.xxx': the following parameters are not available")
			assert.Contains(t, err.Error(), "Connection.Endpoint")
			assert.Contains(t, err.Error(), "TransportRequestID")
		}
	})

	t.Run("DeployableMissing", func(t *testing.T) {
		f := &mock.FilesMock{}
		e := &mock.ExecMockRunner{}
		err := defaultUploadAction.Perform(f, e)
		if assert.EqualError(t, err, "cannot upload artifact 'myDeployable.xxx': file 'myDeployable.xxx' does not exist") {
		}
	})

	t.Run("Deploy command returns with return code not equal zero", func(t *testing.T) {

		e := &mock.ExecMockRunner{}
		e.ExitCode = 1

		err := defaultUploadAction.Perform(f, e)

		assert.EqualError(t, err, "cannot upload artifact 'myDeployable.xxx': Upload command returned with exit code '1'")
	})

	t.Run("Deploy command cannot be executed", func(t *testing.T) {

		e := &mock.ExecMockRunner{
			ShouldFailOnCommand: map[string]error{
				"cmclient.*": fmt.Errorf("cannot execute upload command"),
			},
		}

		err := defaultUploadAction.Perform(f, e)

		assert.EqualError(t, err, "cannot upload artifact 'myDeployable.xxx': cannot execute upload command")
	})

}
