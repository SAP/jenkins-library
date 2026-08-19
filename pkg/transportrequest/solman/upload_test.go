//go:build unit
// +build unit

package solman

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestSolmanUpload(t *testing.T) {

	f := &mock.FilesMock{}
	f.AddFile("myDeployable.xxx", []byte(""))

	defaultUploadAction := UploadAction{}
	defaultUploadAction.WithConnection(
		Connection{
			Endpoint: "https://example.org/solman",
			User:     "me",
			Password: "******",
		})
	defaultUploadAction.WithChangeDocumentID("123456")
	defaultUploadAction.WithTransportRequestID("000K11111111")
	defaultUploadAction.WithApplicationID("MY_APP")
	defaultUploadAction.WithFile("myDeployable.xxx")
	defaultUploadAction.WithCMOpts([]string{"-Dmyprop1=abc", "-Dmyprop2=def"})

	t.Run("Deployable does not exist", func(t *testing.T) {

		uploadActionFileMissing := defaultUploadAction
		uploadActionFileMissing.WithFile("myMissingDeployable.xxx")
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
		uploadAction.WithConnection(
			Connection{
				Endpoint: "",
				User:     "me",
				Password: "******",
			},
		)
		uploadAction.WithTransportRequestID("")
		err := uploadAction.Perform(f, e)
		if assert.Error(t, err) {
			// we should not rely on the order of the missing parameters
			assert.Contains(t, err.Error(), "cannot upload artifact 'myDeployable.xxx': the following parameters are not available")
			assert.Contains(t, err.Error(), "Connection.Endpoint")
			assert.Contains(t, err.Error(), "TransportRequestID")
		}
	})

	t.Run("Deploy command returns with return code not equal zero", func(t *testing.T) {

		e := &mock.ExecMockRunner{}
		e.ExitCode = 1

		err := defaultUploadAction.Perform(f, e)

		assert.EqualError(t, err, "cannot upload artifact 'myDeployable.xxx': upload command returned with exit code '1'")
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
