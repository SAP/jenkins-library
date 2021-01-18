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
		ChangeDocumentId:   "123456",
		TransportRequestId: "000K11111111",
		ApplicationID:      "MY_APP",
		File:               "myDeployable.xxx",
		CMOpts:             []string{"-Dmyprop=abc"},
	}

	t.Run("Deployable does not exist", func(t *testing.T) {

		uploadActionFileMissing := defaultUploadAction
		uploadActionFileMissing.File = "myMissingDeployable.xxx"
		e := &mock.ExecMockRunner{}

		err := uploadActionFileMissing.Perform(f, e)

		assert.EqualError(t, err, "File 'myMissingDeployable.xxx' does not exist.")
	})

	t.Run("Some deploy parameters are missing", func(t *testing.T) {

		uploadActionMissingParameters := defaultUploadAction
		uploadActionMissingParameters.Connection.Endpoint = ""
		uploadActionMissingParameters.ChangeDocumentId = ""
		e := &mock.ExecMockRunner{}

		err := uploadActionMissingParameters.Perform(f, e)

		assert.EqualError(t, err, "Cannot perform artifact upload. The following parameters are not available [Connection.Endpoint ChangeDocumentId]")
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
			assert.Equal(t, []string{"-Dmyprop=abc"}, e.Env)
		}
	})

	t.Run("Deploy command returns with return code not equal zero", func(t *testing.T) {

		e := &mock.ExecMockRunner{}
		e.ExitCode = 1

		err := defaultUploadAction.Perform(f, e)

		assert.EqualError(t, err, "Cannot upload 'myDeployable.xxx': Upload command returned with exit code '1'")
	})

	t.Run("Deploy command cannot be executed", func(t *testing.T) {

		e := &mock.ExecMockRunner{
			ShouldFailOnCommand: map[string]error{
				"cmclient.*": fmt.Errorf("Cannot execute upload command"),
			},
		}

		err := defaultUploadAction.Perform(f, e)

		assert.EqualError(t, err, "Cannot upload 'myDeployable.xxx': Cannot execute upload command")
	})

}

func TestFindEmptyStringsInConfig(t *testing.T) {
	uploadAction := UploadAction{
		Connection: Connection{
			Endpoint: "<set>",
			User:     "",
			Password: "<set>",
		},
		ChangeDocumentId:   "",
		TransportRequestId: "<set>",
		ApplicationID:      "<set>",
		File:               "<set>",
		CMOpts:             []string{},
	}
	emptyStrings, err := FindEmptyStrings(uploadAction)
	if assert.NoError(t, err) {
		assert.Len(t, emptyStrings, 2)
		assert.Subset(t, emptyStrings, []string{"Connection.User", "ChangeDocumentId"})
	}
}
