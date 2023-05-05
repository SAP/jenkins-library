//go:build unit
// +build unit

package rfc

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUploadRFC(t *testing.T) {

	defaultUploadAction := UploadAction{
		Connection: Connection{
			Endpoint: "https://example.org/rfc",
			Client:   "001",
			Instance: "DEV",
			User:     "me",
			Password: "******",
		},
		Application: Application{
			Name:        "myApp",
			Description: "The description",
			AbapPackage: "YYY",
		},
		Configuration: UploadConfig{
			AcceptUnixStyleEndOfLine: true,
			CodePage:                 "UTF-8",
			FailUploadOnWarning:      true,
			Verbose:                  true,
		},
		TransportRequestID: "123456",
		ApplicationURL:     "https://localhost:8081/myDeployable.zip",
	}

	t.Run("straight forward", func(t *testing.T) {

		exec := mock.ExecMockRunner{}

		upload := defaultUploadAction

		err := upload.Perform(&exec)

		if assert.NoError(t, err) {
			assert.Equal(t, exec.Calls, []mock.ExecCall{{Exec: "cts", Params: []string{"uploadToABAP:123456"}}})
			assert.Subset(t, []string{
				"ABAP_DEVELOPMENT_SERVER=https://example.org/rfc",
				"ABAP_DEVELOPMENT_USER=me",
				"ABAP_DEVELOPMENT_PASSWORD=******",
				"ABAP_DEVELOPMENT_INSTANCE=DEV",
				"ABAP_DEVELOPMENT_CLIENT=001",
				"ABAP_APPLICATION_NAME=myApp",
				"ABAP_APPLICATION_DESC=The description",
				"ABAP_PACKAGE=YYY",
				"ZIP_FILE_URL=https://localhost:8081/myDeployable.zip",
				"CODE_PAGE=UTF-8",
				"ABAP_ACCEPT_UNIX_STYLE_EOL=X",
				"FAIL_UPLOAD_ON_WARNING=true",
				"VERBOSE=true",
			}, exec.Env)
			assert.Len(t, exec.Env, 13)
		}
	})

	t.Run("incomplete config", func(t *testing.T) {

		exec := mock.ExecMockRunner{}

		upload := defaultUploadAction
		upload.Connection.Endpoint = ""
		upload.Application.AbapPackage = ""

		err := upload.Perform(&exec)

		if assert.Error(t, err) {
			// Don't want to rely on the order, hence not checking for the full string ...
			assert.Contains(t, err.Error(), "cannot perform artifact upload. The following parameters are not available")
			assert.Contains(t, err.Error(), "Connection.Endpoint")
			assert.Contains(t, err.Error(), "Application.AbapPackage")
		}
	})

	t.Run("invocation of cts tooling fails", func(t *testing.T) {

		t.Run("error raised", func(t *testing.T) {

			exec := mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"cts.*": fmt.Errorf("generic failure")}}

			upload := defaultUploadAction

			err := upload.Perform(&exec)

			assert.EqualError(t, err, "cannot upload artifact: generic failure")
		})

		t.Run("return code not zero", func(t *testing.T) {

			exec := mock.ExecMockRunner{ExitCode: 42}

			upload := defaultUploadAction

			err := upload.Perform(&exec)

			assert.EqualError(t, err, "cannot upload artifact: upload command returned with exit code '42'")
		})
	})
}

func TestTheAbapBool(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		assert.Equal(t, "X", toAbapBool(true))
	})
	t.Run("false", func(t *testing.T) {
		assert.Equal(t, "-", toAbapBool(false))
	})
}
