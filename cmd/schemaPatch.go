package cmd

import (
	"bytes"
	"encoding/json"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/evanphx/json-patch"
	"io/ioutil"
)

func schemaPatch(config schemaPatchOptions, telemetryData *telemetry.CustomData) {
	err := runSchemaPatch(&config, telemetryData, &command.Command{})
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runSchemaPatch(config *schemaPatchOptions, telemetryData *telemetry.CustomData, command execRunner) error {
	schemaFile, err := ioutil.ReadFile(config.Schema)
	if err != nil {
		return nil
	}
	schema := []byte(string(schemaFile))

	patchFile, err := ioutil.ReadFile(config.Patch)
	if err != nil {
		return nil
	}
	patch := []byte(string(patchFile))
	patcher, err := jsonpatch.DecodePatch(patch)
	if err != nil {
		return nil
	}

	patchedSchema, err := patcher.Apply(schema)
	if err != nil {
		panic(err)
	}

	formattedJson, err := jsonPrettyPrint(string(patchedSchema))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(config.Output, formattedJson, 0700)
	if err != nil {
		return err
	}
}

func jsonPrettyPrint(input string) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(input), "", "  ")
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
