package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/influx"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func influxWriteData(config influxWriteDataOptions, _ *telemetry.CustomData) {
	influxClient := influxdb2.NewClient(config.ServerURL, config.AuthToken)
	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := writeData(&config, influxClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func writeData(config *influxWriteDataOptions, influxClient influxdb2.Client) error {
	log.Entry().Info("influxWriteData step")

	client := influx.NewClient(influxClient, config.Organization, config.Bucket)
	var dataMap map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(config.DataMap), &dataMap); err != nil {
		return fmt.Errorf("Failed to unmarshal dataMap: %v", err)
	}
	var dataMapTags map[string]map[string]string
	if config.DataMapTags != "" {
		if err := json.Unmarshal([]byte(config.DataMapTags), &dataMapTags); err != nil {
			return fmt.Errorf("Failed to unmarshal dataMapTags: %v", err)
		}
	}
	if err := client.WriteMetrics(dataMap, dataMapTags); err != nil {
		return err
	}
	log.Entry().Info("Metrics have been written successfully")
	return nil
}
