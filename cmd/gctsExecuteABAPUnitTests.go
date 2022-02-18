package cmd

import (
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func gctsExecuteABAPUnitTests(config gctsExecuteABAPUnitTestsOptions, telemetryData *telemetry.CustomData) {

	var qualityChecksConfig gctsExecuteABAPQualityChecksOptions = gctsExecuteABAPQualityChecksOptions(config)

	gctsExecuteABAPQualityChecks(qualityChecksConfig, telemetryData)
}
