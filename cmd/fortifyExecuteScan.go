package cmd

import (
	"fmt"
	"regexp"
	"time"

	"github.com/SAP/jenkins-library/pkg/fortify"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func fortifyExecuteScan(config fortifyExecuteScanOptions, telemetryData *telemetry.CustomData, influx *fortifyExecuteScanInflux) error {
	sys := fortify.NewSystemInstance(config.ServerURL, config.APIEndpoint, config.AuthToken, time.Second*30)
	return runFortifyScan(config, sys, telemetryData, influx)
}

func runFortifyScan(config fortifyExecuteScanOptions, sys fortify.System, telemetryData *telemetry.CustomData, influx *fortifyExecuteScanInflux) error {
	log.Entry().Debugf("Running Fortify scan against SSC at %v", config.ServerURL)
	gav, err := piperutils.GetMavenGAV(config.BuildDescriptorFile)
	if err != nil {
		log.Entry().Warnf("Unable to load project coordinates from descriptor %v: %v", config.BuildDescriptorFile, err)
	}
	fortifyProjectName, fortifyProjectVersion := determineProjectCoordinates(config, gav)
	project, err := sys.GetProjectByName(fortifyProjectName)
	if err != nil {
		log.Entry().Fatalf("Failed to load project %v: %v", fortifyProjectName, err)
	}
	projectVersion, err := sys.GetProjectVersionDetailsByProjectIDAndVersionName(project.ID, fortifyProjectVersion)
	if err != nil {
		log.Entry().Fatalf("Failed to load project version %v: %v", fortifyProjectVersion, err)
	}
	if len(config.PullRequestName) > 0 {
		fortifyProjectVersion = config.PullRequestName
		projectVersion, err := sys.LookupOrCreateProjectVersionDetailsForPullRequest(project.ID, projectVersion, fortifyProjectVersion)
		if err != nil {
			log.Entry().Fatalf("Failed to lookup / create project version for pull request %v: %v", fortifyProjectVersion, err)
		}
		log.Entry().Debugf("Looked up / created project version with ID %v for PR %v", projectVersion.ID, fortifyProjectVersion)
	} else {
		prID := determinePullRequestMerge(config)
		if len(prID) > 0 {
			log.Entry().Debugf("Determined PR identifier %v for merge check", prID)
			err = sys.MergeProjectVersionStateOfPRIntoMaster(config.FprDownloadEndpoint, config.FprUploadEndpoint, project.ID, projectVersion.ID, fmt.Sprintf("PR-%v", prID))
			if err != nil {
				log.Entry().Fatalf("Failed to merge project version state for pull request %v: %v", fortifyProjectVersion, err)
			}
		}
	}

	log.Entry().Debugf("Scanning and uploading to project %v with version %v and projectVersionId %v", fortifyProjectName, fortifyProjectVersion, projectVersion.ID)

	// TODO create pip /maven command

	// Trigger scan via cmd

	return nil
}

func determinePullRequestMerge(config fortifyExecuteScanOptions) string {
	log.Entry().Debugf("Retrieved commit message %v", config.CommitMessage)
	r, _ := regexp.Compile(config.PullRequestMessageRegex)
	matches := r.FindSubmatch([]byte(config.CommitMessage))
	if matches != nil && len(matches) > 1 {
		return string(matches[config.PullRequestMessageRegexGroup])
	}
	return ""
}

func determineProjectCoordinates(config fortifyExecuteScanOptions, gav *piperutils.MavenGAV) (string, string) {
	projectName, err := piperutils.ExecuteTemplate(config.ProjectName, *gav)
	if err != nil {
		log.Entry().Warnf("Unable to resolve fortify project name %v", err)
	}
	projectVersion, err := piperutils.ExecuteTemplate(config.ProjectVersion, *gav)
	if err != nil {
		log.Entry().Warnf("Unable to resolve fortify project version %v", err)
	}
	return projectName, projectVersion
}
