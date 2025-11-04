package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type validateBOMUtils interface {
	command.ExecRunner

	Glob(pattern string) (matches []string, err error)

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The validateBOMUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type validateBOMUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to validateBOMUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// validateBOMUtilsBundle and forward to the implementation of the dependency.
}

func newValidateBOMUtils() validateBOMUtils {
	utils := validateBOMUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func validateBOM(config validateBOMOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newValidateBOMUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runValidateBOM(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runValidateBOM(config *validateBOMOptions, telemetryData *telemetry.CustomData, utils validateBOMUtils) error {
	// Allow users to skip validation entirely
	if config.Skip {
		log.Entry().Info("BOM validation skipped (skip: true)")
		return nil
	}

	// Find BOM files matching the pattern
	bomFiles, err := utils.Glob(config.BomPattern)
	if err != nil {
		return fmt.Errorf("failed to find BOM files with pattern '%s': %w", config.BomPattern, err)
	}

	// Silent success if no BOM files found
	if len(bomFiles) == 0 {
		log.Entry().Infof("No BOM files found matching pattern '%s', skipping validation", config.BomPattern)
		return nil
	}

	log.Entry().Infof("Found %d BOM file(s) to validate", len(bomFiles))

	validationErrors := 0
	validatedFiles := 0

	// Validate each BOM file
	for _, bomFile := range bomFiles {
		log.Entry().Infof("Validating BOM file: %s", bomFile)

		if err := piperutils.ValidateCycloneDX14(bomFile); err != nil {
			log.Entry().WithError(err).Warnf("BOM validation failed for: %s", bomFile)
			validationErrors++

			// If configured to fail on validation errors, return immediately
			if config.FailOnValidationError {
				return fmt.Errorf("BOM validation failed for %s: %w", bomFile, err)
			}
		} else {
			log.Entry().Infof("BOM validation passed: %s", bomFile)
			validatedFiles++

			// Extract and log PURL if requested
			if config.ValidatePurl {
				purl := piperutils.GetPurl(bomFile)
				if purl != "" {
					log.Entry().Infof("BOM PURL: %s", purl)
				} else {
					log.Entry().Debugf("No PURL found in BOM file: %s", bomFile)
				}
			}
		}
	}

	// Log summary
	log.Entry().Infof("BOM validation complete: %d/%d files validated successfully", validatedFiles, len(bomFiles))

	if validationErrors > 0 && !config.FailOnValidationError {
		log.Entry().Warnf("%d BOM file(s) failed validation but failOnValidationError is false, continuing", validationErrors)
	}

	return nil
}
