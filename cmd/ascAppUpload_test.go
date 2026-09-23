//go:build unit
// +build unit

package cmd

import (
	"errors"
	"testing"

	"github.com/SAP/jenkins-library/pkg/asc"
	"github.com/SAP/jenkins-library/pkg/mock"

	"github.com/stretchr/testify/assert"
)

type ascAppUploadMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newAscAppUploadTestsUtils() ascAppUploadMockUtils {
	utils := ascAppUploadMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

type ascSystemMock struct {
	deployRequest  asc.DeployRequest
	deployAppError error
}

func (sys *ascSystemMock) DeployApp(request asc.DeployRequest) error {
	sys.deployRequest = request
	return sys.deployAppError
}

func TestRunAscAppUpload(t *testing.T) {
	t.Parallel()

	t.Run("successful deploy", func(t *testing.T) {
		t.Parallel()
		// init
		config := ascAppUploadOptions{
			FilePath:           "./sample-app.ipa",
			JamfTargetSystem:   "test",
			BundleID:           "com.sap.sample",
			ReleaseAppVersion:  "3.1.15",
			ReleaseDescription: "test test",
			ReleaseDate:        "2026-09-17",
			ReleaseVisible:     true,
		}

		utils := newAscAppUploadTestsUtils()
		utils.AddFile("sample-app.ipa", []byte("dummy content"))

		ascClient := &ascSystemMock{}

		// test
		err := runAscAppUpload(&config, nil, utils, ascClient)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, "com.sap.sample", ascClient.deployRequest.BundleID)
		assert.Equal(t, "./sample-app.ipa", ascClient.deployRequest.FilePath)
		assert.Equal(t, "3.1.15", ascClient.deployRequest.Version)
		assert.Equal(t, "test test", ascClient.deployRequest.Description)
		assert.Equal(t, "2026-09-17", ascClient.deployRequest.ReleaseDate)
		assert.True(t, ascClient.deployRequest.Visible)
		assert.Equal(t, "test", ascClient.deployRequest.TargetSystem)
	})

	t.Run("error if jamfTargetSystem is not set", func(t *testing.T) {
		t.Parallel()
		// init
		config := ascAppUploadOptions{
			FilePath: "./sample-app.ipa",
			BundleID: "com.sap.sample",
		}

		utils := newAscAppUploadTestsUtils()

		ascClient := &ascSystemMock{}

		// test
		err := runAscAppUpload(&config, nil, utils, ascClient)

		// assert
		assert.EqualError(t, err, "jamfTargetSystem must be set")
	})

	t.Run("error during deploy", func(t *testing.T) {
		t.Parallel()
		// init
		config := ascAppUploadOptions{
			FilePath:         "./sample-app.ipa",
			JamfTargetSystem: "test",
			BundleID:         "com.sap.sample",
		}

		utils := newAscAppUploadTestsUtils()

		errorMessage := "Error while deploying app"

		ascClient := &ascSystemMock{
			deployAppError: errors.New(errorMessage),
		}

		// test
		err := runAscAppUpload(&config, nil, utils, ascClient)

		// assert
		assert.EqualError(t, err, "failed to deploy app: "+errorMessage)
	})
}
