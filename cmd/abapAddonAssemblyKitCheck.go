package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type abapAddonAssemblyKitCheckUtils interface {
	command.ExecRunner
	piperhttp.Sender
}

type abapAddonAssemblyKitCheckUtilsBundle struct {
	*command.Command
	*piperhttp.Client
}

func newAbapAddonAssemblyKitCheckUtils() abapAddonAssemblyKitCheckUtils {
	utils := abapAddonAssemblyKitCheckUtilsBundle{
		Command: &command.Command{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func abapAddonAssemblyKitCheck(config abapAddonAssemblyKitCheckOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *abapAddonAssemblyKitCheckCommonPipelineEnvironment) {
	utils := newAbapAddonAssemblyKitCheckUtils()

	err := runAbapAddonAssemblyKitCheck(&config, telemetryData, utils, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitCheck(config *abapAddonAssemblyKitCheckOptions, telemetryData *telemetry.CustomData, utils abapAddonAssemblyKitCheckUtils, commonPipelineEnvironment *abapAddonAssemblyKitCheckCommonPipelineEnvironment) error {

	// log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")
	// // Example of calling methods from external dependencies directly on utils:
	// exists, err := utils.FileExists("file.txt")
	// if err != nil {
	// 	// It is good practice to set an error category.
	// 	// Most likely you want to do this at the place where enough context is known.
	// 	log.SetErrorCategory(log.ErrorConfiguration)
	// 	// Always wrap non-descriptive errors to enrich them with context for when they appear in the log:
	// 	return fmt.Errorf("failed to check for important file: %w", err)
	// }
	// if !exists {
	// 	log.SetErrorCategory(log.ErrorConfiguration)
	// 	return fmt.Errorf("cannot run without important file")
	// }

	return nil
}

type ProductVersionHeader struct {
	ProductName            string
	SemanticProductVersion string `json:"SemProductVersion"`
	ProductVersion         string
	Spslevel               string
	PatchLevel             string
	Vendor                 string
	VendorType             string
	Content                []ProductVersionContent
}

type ProductVersionContent struct {
	ProductName                      string
	SemanticProductVersion           string `json:"SemProductVersion"`
	SoftwareComponentName            string `json:"ScName"`
	SemanticSoftwareComponentVersion string `json:"SemScVersion"`
	SoftwareComponentVersion         string `json:"ScVersion"`
	SpLevel                          string
	PatchLevel                       string
	Vendor                           string
	VendorType                       string
}
