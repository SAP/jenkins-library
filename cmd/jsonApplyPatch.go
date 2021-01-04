package cmd

import (
	"bytes"
	"encoding/json"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/evanphx/json-patch"
)

type jsonApplyPatchUtils interface {
	Indent(dst *bytes.Buffer, src []byte, prefix, indent string) error
}

type jsonApplyPatchUtilsBundle struct{}

func (j jsonApplyPatchUtilsBundle) Indent(dst *bytes.Buffer, src []byte, prefix, indent string) error {
	return json.Indent(dst, src, prefix, indent)
}

func jsonApplyPatch(config jsonApplyPatchOptions, _ *telemetry.CustomData) {
	err := runJsonApplyPatch(&config, &piperutils.Files{}, jsonApplyPatchUtilsBundle{})
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runJsonApplyPatch(config *jsonApplyPatchOptions, fileUtils piperutils.FileUtils, utils jsonApplyPatchUtils) error {
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

	formattedJson, err := formatJson(patchedSchema, utils)
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

func formatJson(input []byte, utils jsonApplyPatchUtils) ([]byte, error) {
	var output bytes.Buffer
	err := utils.Indent(&output, input, "", "    ")
	if err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
