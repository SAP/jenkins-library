package cmd

import (
	"github.com/SAP/jenkins-library/pkg/influx"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func influxWriteData(config influxWriteDataOptions, telemetryData *telemetry.CustomData) {

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := writeData(&config, telemetryData)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func writeData(config *influxWriteDataOptions, telemetryData *telemetry.CustomData) error {
	log.Entry().Info("influxWriteData step")

	client, err := influx.NewClient(config.InfluxVersion, config.ServerURL, config.AuthToken, config.Organization, config.Bucket)
	if err != nil {
		return err
	}
	err = client.WriteMetrics(config.DataMap, config.DataMapTags)
	log.Entry().Info("Metrics have been written successfully")
	return err
}
