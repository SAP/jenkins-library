package cmd

import (
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func gCTSTest(config gCTSTestOptions, telemetryData *telemetry.CustomData) {
	var qualityChecksConfig gctsExecuteABAPQualityChecksOptions = gctsExecuteABAPQualityChecksOptions(config)

	gctsExecuteABAPQualityChecks(qualityChecksConfig, telemetryData)
}
