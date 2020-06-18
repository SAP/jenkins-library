package cmd

import (
	"bytes"
	"encoding/json"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/evanphx/json-patch"
)

func schemaPatch(config schemaPatchOptions, telemetryData *telemetry.CustomData) {
	err := runSchemaPatch(&config, &piperutils.Files{})
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runSchemaPatch(config *schemaPatchOptions, fileUtils piperutils.FileUtils) error {
	schemaFile, err := fileUtils.FileRead(config.Schema)
	if err != nil {
		return nil
	}
	schema := []byte(string(schemaFile))

	patchFile, err := fileUtils.FileRead(config.Patch)
	if err != nil {
		return nil
	}
	patcher, err := jsonpatch.DecodePatch(patchFile)
	if err != nil {
		return err
	}

	patchedSchema, err := patcher.Apply(schema)
	if err != nil {
		return err
	}

	formattedJson, err := formatJson(patchedSchema)
	if err != nil {
		formattedJson = patchedSchema
	}

	err = fileUtils.FileWrite(config.Output, formattedJson, 0700)
	if err != nil {
		return err
	}
	return nil
}

func formatJson(input []byte) ([]byte, error) {
	var output bytes.Buffer
	err := json.Indent(&output, input, "", "    ")
	if err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
