package cmd

import (
	"bytes"
	"encoding/json"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	jsonpatch "github.com/evanphx/json-patch"
)

func jsonApplyPatch(config jsonApplyPatchOptions, telemetryData *telemetry.CustomData) {
	err := runJsonApplyPatch(&config, &piperutils.Files{})
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runJsonApplyPatch(config *jsonApplyPatchOptions, fileUtils piperutils.FileUtils) error {
	schema, err := fileUtils.FileRead(config.Input)
	if err != nil {
		return err
	}

	patchFile, err := fileUtils.FileRead(config.Patch)
	if err != nil {
		return err
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
		// Ignore error and just use original result.
		formattedJson = patchedSchema
	}

	err = fileUtils.FileWrite(config.Output, formattedJson, 0644)
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
