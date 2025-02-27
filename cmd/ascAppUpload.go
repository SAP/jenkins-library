package cmd

import (
	"github.com/SAP/jenkins-library/pkg/asc"
	"github.com/SAP/jenkins-library/pkg/command"
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

type ascAppUploadUtils interface {
	command.ExecRunner
}

type ascAppUploadUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func newAscAppUploadUtils() ascAppUploadUtils {
	utils := ascAppUploadUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func ascAppUpload(config ascAppUploadOptions, telemetryData *telemetry.CustomData) {
	utils := newAscAppUploadUtils()
	client := &piperHttp.Client{}

	ascClient, err := asc.NewSystemInstance(client, config.ServerURL, config.AppToken)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Failed to create ASC client talking to URL %v", config.ServerURL)
	} else {
		err = runAscAppUpload(&config, telemetryData, utils, ascClient)
	}

	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAscAppUpload(config *ascAppUploadOptions, telemetryData *telemetry.CustomData, utils ascAppUploadUtils, ascClient asc.System) error {

	if len(config.JamfTargetSystem) == 0 {
		return errors.New("jamfTargetSystem must be set")
	}

	log.Entry().Infof("Collect data to create new release in ASC")

	app, err := ascClient.GetAppById(config.AppID)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrap(err, "failed to get app information")
	}

	log.Entry().Debugf("Found App with name %v", app.AppName)

	log.Entry().Infof("Create release for %v in ASC (AppID %v)", app.AppName, app.AppId)

	releaseResponse, err := ascClient.CreateRelease(app.AppId, config.ReleaseAppVersion, config.ReleaseDescription, config.ReleaseDate, config.ReleaseVisible)

	if err != nil {
		log.SetErrorCategory(log.ErrorService)
		return errors.Wrap(err, "failed to create release")
	}

	if releaseResponse.Status != "success" {
		log.SetErrorCategory(log.ErrorService)
		return errors.New(releaseResponse.Message)
	}

	log.Entry().Infof("Collect data to upload app to ASC & Jamf")

	jamfAppInformationResponse, err := ascClient.GetJamfAppInfo(app.BundleId, config.JamfTargetSystem)
	if err != nil {
		log.SetErrorCategory(log.ErrorService)
		return errors.Wrap(err, "failed to get jamf app info")
	}

	jamfAppId := jamfAppInformationResponse.MobileDeviceApplication.General.Id

	if jamfAppId == 0 {
		return errors.Errorf("failed to get jamf app id")
	}

	log.Entry().Debugf("Got Jamf info for app %v, jamfId: %v", app.AppName, jamfAppId)

	log.Entry().Infof("Upload ipa %v to ASC & Jamf", config.FilePath)

	err = ascClient.UploadIpa(config.FilePath, jamfAppId, config.JamfTargetSystem, app.BundleId, releaseResponse.Data)
	if err != nil {
		log.SetErrorCategory(log.ErrorService)
		return errors.Wrap(err, "failed to upload ipa")
	}

	log.Entry().Infof("Successfully uploaded %v to ASC (AppId %v) & Jamf (Id %v)", config.FilePath, app.AppId, jamfAppId)

	return nil
}
