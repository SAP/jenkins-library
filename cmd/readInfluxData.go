package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/spf13/cobra"
)

const (
	influxDataPath = ".pipeline/influx/"
)

// ReadPipelineEnv reads the commonPipelineEnvironment from disk and outputs it as JSON
func ReadInfluxData() *cobra.Command {
	return &cobra.Command{
		Use:   "readInfluxData",
		Short: "Reads the Influx data written by the steps from disk and outputs it as JSON",
		PreRun: func(cmd *cobra.Command, args []string) {
			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)
		},

		Run: func(cmd *cobra.Command, args []string) {
			fileUtils := piperutils.Files{}
			err := runReadInfluxData(&fileUtils, os.Stdout)
			if err != nil {
				log.Entry().Fatalf("error when reading Influx data: %v", err)
			}
		},
	}
}

func runReadInfluxData(fileUtils piperutils.FileUtils, out io.Writer) error {
	files, err := fileUtils.Glob(fmt.Sprintf("%v**", influxDataPath))
	if err != nil {
		return fmt.Errorf("failed to find Influx data files: %w", err)
	}

	influxData := struct {
		Fields map[string]map[string]interface{} `json:"fields,omitempty"`
		Tags   map[string]map[string]interface{} `json:"tags,omitempty"`
	}{
		Fields: map[string]map[string]interface{}{},
		Tags:   map[string]map[string]interface{}{},
	}

	for _, fileName := range files {
		parts := strings.Split(strings.TrimPrefix(fileName, influxDataPath), "/")
		if len(parts) != 3 {
			log.Entry().Infof("skipping to read %v", fileName)
			continue
		}
		theMeasurement := parts[0]
		theType := parts[1]
		theName := parts[2]
		var theValue interface{}

		fileContent, err := fileUtils.FileRead(fileName)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		ext := filepath.Ext(fileName)
		if ext == ".json" {
			theName = strings.TrimSuffix(theName, ".json")
			if err := json.Unmarshal(fileContent, &theValue); err != nil {
				return fmt.Errorf("failed to unmarshal json content of influx data file %v: %w", fileName, err)
			}
		} else {
			switch string(fileContent) {
			case "true":
				theValue = true
			case "false":
				theValue = false
			default:
				theValue = string(fileContent)
			}
		}

		if theType == "fields" {
			if influxData.Fields[theMeasurement] == nil {
				influxData.Fields[theMeasurement] = map[string]interface{}{}
			}
			influxData.Fields[theMeasurement][theName] = theValue
		} else if theType == "tags" {
			if influxData.Tags[theMeasurement] == nil {
				influxData.Tags[theMeasurement] = map[string]interface{}{}
			}
			influxData.Tags[theMeasurement][theName] = theValue
		}
	}

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "\t")
	if err := encoder.Encode(influxData); err != nil {
		return err
	}

	return nil
}
