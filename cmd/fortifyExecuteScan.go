package cmd

import (
	"github.com/SAP/jenkins-library/pkg/fortify"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func fortifyExecuteScan(config fortifyExecuteScanOptions, telemetryData *telemetry.CustomData, influx *fortifyExecuteScanInflux) error {
	log.Entry().WithField("customKey", "customValue").Info("This is how you write a log message with a custom field ...")
	return nil
}

func runFortifyScan(config fortifyExecuteScanOptions, sys fortify.System, workspace string, telemetryData *telemetry.CustomData, influx *fortifyExecuteScanInflux) error {
	return nil
}
