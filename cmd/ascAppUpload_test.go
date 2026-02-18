//go:build unit

package cmd

import (
	"errors"
	"fmt"
	"testing"
	"time"

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
	app                        asc.App
	appError                   error
	createReleaseResponse      asc.CreateReleaseResponse
	createReleaseResponseError error
	jamfAppInfo                asc.JamfAppInformationResponse
	jamfAppInfoError           error
	uploadIpaError             error
}

func (sys *ascSystemMock) GetAppById(appId string) (asc.App, error) {
	return sys.app, sys.appError
}

func (sys *ascSystemMock) CreateRelease(ascAppId int, version string, description string, releaseDate string, visible bool) (asc.CreateReleaseResponse, error) {
	return sys.createReleaseResponse, sys.createReleaseResponseError
}

func (sys *ascSystemMock) GetJamfAppInfo(bundleId string, jamfTargetSystem string) (asc.JamfAppInformationResponse, error) {
	return sys.jamfAppInfo, sys.jamfAppInfoError
}

func (sys *ascSystemMock) UploadIpa(path string, jamfAppId int, jamfTargetSystem string, bundleId string, ascRelease asc.Release) error {
	return sys.uploadIpaError
}

func TestRunAscAppUpload(t *testing.T) {
	t.Parallel()

	t.Run("succesfull upload", func(t *testing.T) {
		t.Parallel()
		// init
		config := ascAppUploadOptions{
			FilePath:         "./sample-app.ipa",
			JamfTargetSystem: "test",
			AppID:            "1",
		}

		utils := newAscAppUploadTestsUtils()
		utils.AddFile("sample-app.ipa", []byte("dummy content"))

		ascClient := &ascSystemMock{
			app: asc.App{
				AppId:    1,
				AppName:  "Sample App",
				BundleId: "sample.bundle.id",
				JamfId:   "1",
			},
			createReleaseResponse: asc.CreateReleaseResponse{
				Status: "success",
				Data:   asc.Release{ReleaseID: 1, AppID: 1, Version: "version", Description: "description", ReleaseDate: time.Now(), Visible: true},
			},
			jamfAppInfo: asc.JamfAppInformationResponse{
				MobileDeviceApplication: asc.JamfMobileDeviceApplication{
					General: asc.JamfMobileDeviceApplicationGeneral{
						Id: 1,
					},
				},
			},
		}
		// test
		err := runAscAppUpload(&config, nil, utils, ascClient)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error during release creation", func(t *testing.T) {
		t.Parallel()
		// init
		config := ascAppUploadOptions{
			FilePath:         "./sample-app.ipa",
			JamfTargetSystem: "test",
			AppID:            "1",
		}

		utils := newAscAppUploadTestsUtils()

		errorMessage := "Error while creating release"

		ascClient := &ascSystemMock{
			app: asc.App{
				AppId:    1,
				AppName:  "Sample App",
				BundleId: "sample.bundle.id",
				JamfId:   "1",
			},
			createReleaseResponse: asc.CreateReleaseResponse{Status: "failure", Message: errorMessage},
		}
		// test
		err := runAscAppUpload(&config, nil, utils, ascClient)

		// assert
		assert.EqualError(t, err, errorMessage)
	})

	t.Run("error while fetching jamf app info", func(t *testing.T) {
		t.Parallel()
		// init
		config := ascAppUploadOptions{
			FilePath:         "./sample-app.ipa",
			JamfTargetSystem: "test",
			AppID:            "1",
		}

		utils := newAscAppUploadTestsUtils()

		errorMessage := "Error while fetching jamf app info"

		ascClient := &ascSystemMock{
			app: asc.App{
				AppId:    1,
				AppName:  "Sample App",
				BundleId: "sample.bundle.id",
				JamfId:   "1",
			},
			createReleaseResponse: asc.CreateReleaseResponse{Status: "success", Data: asc.Release{ReleaseID: 1}},
			jamfAppInfoError:      errors.New(errorMessage),
		}
		// test
		err := runAscAppUpload(&config, nil, utils, ascClient)

		// assert
		assert.EqualError(t, err, fmt.Sprintf("failed to get jamf app info: %s", errorMessage))
	})

	t.Run("error if jamf app id is 0", func(t *testing.T) {
		t.Parallel()
		// init
		config := ascAppUploadOptions{
			FilePath:         "./sample-app.ipa",
			JamfTargetSystem: "test",
			AppID:            "1",
		}

		utils := newAscAppUploadTestsUtils()

		ascClient := &ascSystemMock{
			app: asc.App{
				AppId:    1,
				AppName:  "Sample App",
				BundleId: "sample.bundle.id",
				JamfId:   "1",
			},
			createReleaseResponse: asc.CreateReleaseResponse{Status: "success", Data: asc.Release{ReleaseID: 1}},
			jamfAppInfo: asc.JamfAppInformationResponse{
				MobileDeviceApplication: asc.JamfMobileDeviceApplication{
					General: asc.JamfMobileDeviceApplicationGeneral{
						Id: 0,
					},
				},
			},
		}
		// test
		err := runAscAppUpload(&config, nil, utils, ascClient)

		// assert
		assert.EqualError(t, err, fmt.Sprintf("failed to get jamf app id"))
	})
}
