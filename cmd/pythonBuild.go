1package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/buildsettings"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/python"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

const (
	cycloneDxVersion       = "6.1.1"
	cycloneDxSchemaVersion = "1.4"
	stepName               = "pythonBuild"
)

type pythonBuildUtils interface {
	command.ExecRunner
	FileExists(filename string) (bool, error)
	piperutils.FileUtils
}

type pythonBuildUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func newPythonBuildUtils() pythonBuildUtils {
	utils := pythonBuildUtilsBundle{
		Command: &command.Command{
			StepName: stepName,
		},
		Files: &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func pythonBuild(config pythonBuildOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *pythonBuildCommonPipelineEnvironment) {
	utils := newPythonBuildUtils()

	err := runPythonBuild(&config, telemetryData, utils, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runPythonBuild(config *pythonBuildOptions, telemetryData *telemetry.CustomData, utils pythonBuildUtils, commonPipelineEnvironment *pythonBuildCommonPipelineEnvironment) error {
	if exitHandler, err := python.CreateVirtualEnvironment(utils.RunExecutable, utils.RemoveAll, config.VirtualEnvironmentName); err != nil {
		return err
	} else {
		log.DeferExitHandler(exitHandler)
		defer exitHandler()
	}

	if err := python.BuildWithSetupPy(utils.RunExecutable, config.VirtualEnvironmentName, config.BuildFlags, config.SetupFlags); err != nil {
		return err
	}

	if config.CreateBOM {
		if err := python.CreateBOM(utils.RunExecutable, utils.FileExists, config.VirtualEnvironmentName, config.RequirementsFilePath, cycloneDxVersion, cycloneDxSchemaVersion); err != nil {
			return fmt.Errorf("BOM creation failed: %w", err)
		}
	}

	if info, err := createBuildSettingsInfo(config); err != nil {
		return err
	} else {
		commonPipelineEnvironment.custom.buildSettingsInfo = info
	}

	if config.Publish {
		if err := python.PublishPackage(
			utils.RunExecutable,
			config.VirtualEnvironmentName,
			config.TargetRepositoryURL,
			config.TargetRepositoryUser,
			config.TargetRepositoryPassword,
		); err != nil {
			return fmt.Errorf("failed to publish: %w", err)
		}
	}

	// After build, rename all dist/* files with underscores to dashes in the base name
	// This was introduced to fix setuptools  renaming from "-" to "_"
	const distDir = "dist" // move to config or ?
	renameArtifactsInDist(distDir)

	return nil
}

func renameArtifactsInDist(distDir string) {
    files, err := os.ReadDir(distDir)
    if err != nil {
        log.Entry().Warnf("Could not read dist directory for artifact renaming: %v", err)
        return
    }

    for _, f := range files {
        oldName := f.Name()
        // Only process .tar.gz artifacts with at least one underscore
        if !strings.Contains(oldName, "_") || !strings.HasSuffix(oldName, ".tar.gz") {
            continue
        }

        // Remove .tar.gz extension for processing
        base := strings.TrimSuffix(oldName, ".tar.gz")
        lastUnderscore := strings.LastIndex(base, "_")
        if lastUnderscore == -1 {
            continue
        }
        namePart := base[:lastUnderscore]
        versionPart := base[lastUnderscore+1:]
        // Replace underscores with dashes only in the version part
        newVersionPart := strings.ReplaceAll(versionPart, "_", "-")
        newName := namePart + "_" + newVersionPart + ".tar.gz"
        if newName == oldName {
            continue
        }
        oldPath := filepath.Join(distDir, oldName)
        newPath := filepath.Join(distDir, newName)
        if err := os.Rename(oldPath, newPath); err != nil {
            log.Entry().Warnf("Failed to rename artifact %s to %s: %v", oldName, newName, err)
        } else {
            log.Entry().Infof("Renamed artifact %s to %s", oldName, newName)
        }
    }
}

// TODO: extract to common place
func createBuildSettingsInfo(config *pythonBuildOptions) (string, error) {
	log.Entry().Debugf("creating build settings information...")
	dockerImage, err := GetDockerImageValue(stepName)
	if err != nil {
		return "", err
	}

	pythonConfig := buildsettings.BuildOptions{
		CreateBOM:         config.CreateBOM,
		Publish:           config.Publish,
		BuildSettingsInfo: config.BuildSettingsInfo,
		DockerImage:       dockerImage,
	}
	buildSettingsInfo, err := buildsettings.CreateBuildSettingsInfo(&pythonConfig, stepName)
	if err != nil {
		return "", err
	}
	return buildSettingsInfo, nil
}
