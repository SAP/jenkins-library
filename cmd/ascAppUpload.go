package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/SAP/jenkins-library/pkg/asc"
	"github.com/SAP/jenkins-library/pkg/command"
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
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

	ascClient, err := asc.NewSystemInstance(client, config.ServerURL, config.AppToken, time.Duration(config.Timeout)*time.Second)
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

	log.Entry().Infof("Deploy %v to ASC & Jamf (BundleID %v)", config.FilePath, config.BundleID)

	err := ascClient.DeployApp(asc.DeployRequest{
		BundleID:     config.BundleID,
		FilePath:     config.FilePath,
		Version:      config.ReleaseAppVersion,
		Description:  config.ReleaseDescription,
		ReleaseDate:  config.ReleaseDate,
		Visible:      config.ReleaseVisible,
		TargetSystem: config.JamfTargetSystem,
		User:         config.User,
	})
	if err != nil {
		log.SetErrorCategory(log.ErrorService)
		return fmt.Errorf("failed to deploy app: %w", err)
	}

	log.Entry().Infof("Successfully deployed %v to ASC (BundleID %v) & Jamf (system %v)", config.FilePath, config.BundleID, config.JamfTargetSystem)

	return nil
}
