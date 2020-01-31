package cmd

import (
	"os"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"

	"github.com/spf13/cobra"
)

type detectExecuteScanOptions struct {
	APIToken       string   `json:"apiToken,omitempty"`
	CodeLocation   string   `json:"codeLocation,omitempty"`
	ProjectName    string   `json:"projectName,omitempty"`
	ProjectVersion string   `json:"projectVersion,omitempty"`
	Scanners       []string `json:"scanners,omitempty"`
	ScanPaths      []string `json:"scanPaths,omitempty"`
	ScanProperties []string `json:"scanProperties,omitempty"`
	ServerURL      string   `json:"serverUrl,omitempty"`
}

var myDetectExecuteScanOptions detectExecuteScanOptions

// DetectExecuteScanCommand Executes Synopsis Detect scan
func DetectExecuteScanCommand() *cobra.Command {
	metadata := detectExecuteScanMetadata()

	var createDetectExecuteScanCmd = &cobra.Command{
		Use:   "detectExecuteScan",
		Short: "Executes Synopsis Detect scan",
		Long:  `This step executes [Synopsis Detect](https://synopsys.atlassian.net/wiki/spaces/INTDOCS/pages/62423113/Synopsys+Detect) scans.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			log.SetStepName("detectExecuteScan")
			log.SetVerbose(GeneralConfig.Verbose)
			return PrepareConfig(cmd, &metadata, "detectExecuteScan", &myDetectExecuteScanOptions, config.OpenPiperFile)
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			telemetry.Initialize(GeneralConfig.NoTelemetry, "detectExecuteScan")
			telemetry.Send(&telemetry.CustomData{})
			return detectExecuteScan(myDetectExecuteScanOptions)
		},
	}

	addDetectExecuteScanFlags(createDetectExecuteScanCmd)
	return createDetectExecuteScanCmd
}

func addDetectExecuteScanFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&myDetectExecuteScanOptions.APIToken, "apiToken", os.Getenv("PIPER_apiToken"), "Api token to be used for connectivity with Synopsis Detect server.")
	cmd.Flags().StringVar(&myDetectExecuteScanOptions.CodeLocation, "codeLocation", os.Getenv("PIPER_codeLocation"), "An override for the name Detect will use for the scan file it creates.")
	cmd.Flags().StringVar(&myDetectExecuteScanOptions.ProjectName, "projectName", os.Getenv("PIPER_projectName"), "Name of the Synopsis Detect (formerly BlackDuck) project.")
	cmd.Flags().StringVar(&myDetectExecuteScanOptions.ProjectVersion, "projectVersion", os.Getenv("PIPER_projectVersion"), "Version of the Synopsis Detect (formerly BlackDuck) project.")
	cmd.Flags().StringSliceVar(&myDetectExecuteScanOptions.Scanners, "scanners", []string{"signature"}, "List of scanners to be used for Synopsis Detect (formerly BlackDuck) scan.")
	cmd.Flags().StringSliceVar(&myDetectExecuteScanOptions.ScanPaths, "scanPaths", []string{"."}, "List of paths which should be scanned by the Synopsis Detect (formerly BlackDuck) scan.")
	cmd.Flags().StringSliceVar(&myDetectExecuteScanOptions.ScanProperties, "scanProperties", []string{"--blackduck.signature.scanner.memory=4096", "--blackduck.timeout=6000", "--blackduck.trust.cert=true", "--detect.policy.check.fail.on.severities=BLOCKER,CRITICAL,MAJOR", "--detect.report.timeout=4800", "--logging.level.com.synopsys.integration=DEBUG"}, "Properties passed to the Synopsis Detect (formerly BlackDuck) scan. You can find details in the [Synopsis Detect documentation](https://synopsys.atlassian.net/wiki/spaces/INTDOCS/pages/622846/Using+Synopsys+Detect+Properties)")
	cmd.Flags().StringVar(&myDetectExecuteScanOptions.ServerURL, "serverUrl", os.Getenv("PIPER_serverUrl"), "Server url to the Synopsis Detect (formerly BlackDuck) Server.")

	cmd.MarkFlagRequired("apiToken")
	cmd.MarkFlagRequired("projectName")
	cmd.MarkFlagRequired("projectVersion")
}

// retrieve step metadata
func detectExecuteScanMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "apiToken",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "detect/apiToken"}},
					},
					{
						Name:        "codeLocation",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "projectName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "detect/projectName"}},
					},
					{
						Name:        "projectVersion",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "detect/projectVersion"}},
					},
					{
						Name:        "scanners",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "detect/scanners"}},
					},
					{
						Name:        "scanPaths",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "detect/scanPaths"}},
					},
					{
						Name:        "scanProperties",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "detect/scanProperties"}},
					},
					{
						Name:        "serverUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "detect/serverUrl"}},
					},
				},
			},
		},
	}
	return theMetaData
}
