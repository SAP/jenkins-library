package buildpacks

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type Telemetry struct {
	customData *telemetry.CustomData
	runImages  []string
}

func NewTelemetry(customData *telemetry.CustomData) *Telemetry {
	return &Telemetry{
		customData: customData,
	}
}

func (d *Telemetry) WithBuilder(builder string) {
	d.customData.CnbBuilder = builder
}

func (d *Telemetry) WithBuildTool(buildTool string) {
	d.customData.BuildTool = buildTool
}

func (d *Telemetry) WithRunImage(runImage string) {
	if d.customData.CnbRunImage == "" {
		d.customData.CnbRunImage = runImage
	} else {
		d.customData.CnbRunImage += fmt.Sprintf(",%s", runImage)
	}
}
