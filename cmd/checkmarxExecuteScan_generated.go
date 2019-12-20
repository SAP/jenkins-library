package cmd

import (
	"os"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/spf13/cobra"
)

type checkmarxExecuteScanOptions struct {
	FullScanCycle                 string `json:"fullScanCycle,omitempty"`
	VulnerabilityThresholdResult  string `json:"vulnerabilityThresholdResult,omitempty"`
	VulnerabilityThresholdUnit    string `json:"vulnerabilityThresholdUnit,omitempty"`
	AvoidDuplicateProjectScans    bool   `json:"avoidDuplicateProjectScans,omitempty"`
	GeneratePdfReport             bool   `json:"generatePdfReport,omitempty"`
	VulnerabilityThresholdEnabled bool   `json:"vulnerabilityThresholdEnabled,omitempty"`
	FullScansScheduled            bool   `json:"fullScansScheduled,omitempty"`
	Incremental                   bool   `json:"incremental,omitempty"`
	Preset                        string `json:"preset,omitempty"`
	CheckmarxProject              string `json:"checkmarxProject,omitempty"`
	Verbose                       string `json:"verbose,omitempty"`
	CheckmarxGroupID              string `json:"checkmarxGroupId,omitempty"`
	PullRequestName               string `json:"pullRequestName,omitempty"`
	FilterPattern                 string `json:"filterPattern,omitempty"`
	VulnerabilityThresholdLow     string `json:"vulnerabilityThresholdLow,omitempty"`
	SourceEncoding                string `json:"sourceEncoding,omitempty"`
	VulnerabilityThresholdMedium  string `json:"vulnerabilityThresholdMedium,omitempty"`
	ValidTypeScriptPresets        string `json:"validTypeScriptPresets,omitempty"`
	CheckmarxServerURL            string `json:"checkmarxServerUrl,omitempty"`
	VulnerabilityThresholdHigh    string `json:"vulnerabilityThresholdHigh,omitempty"`
	TeamName                      string `json:"teamName,omitempty"`
	EngineConfiguration           string `json:"engineConfiguration,omitempty"`
	Username                      string `json:"username,omitempty"`
	Password                      string `json:"password,omitempty"`
}

var myCheckmarxExecuteScanOptions checkmarxExecuteScanOptions
var checkmarxExecuteScanStepConfigJSON string

// CheckmarxExecuteScanCommand Checkmarx is the recommended tool for security scans of JavaScript, iOS, Swift and Ruby code.
func CheckmarxExecuteScanCommand() *cobra.Command {
	metadata := checkmarxExecuteScanMetadata()
	var createCheckmarxExecuteScanCmd = &cobra.Command{
		Use:   "checkmarxExecuteScan",
		Short: "Checkmarx is the recommended tool for security scans of JavaScript, iOS, Swift and Ruby code.",
		Long: `Checkmarx is the recommended tool for security scans of JavaScript, iOS, Swift and Ruby code.
You find further information in the [Checkmarx Jam group](https://jam4.sapjam.com/groups/about_page/1mlscAGHT38VQ4vxGfhE6u).

In addition some background information is collected on [following Wiki page](https://wiki.wdf.sap.corp/wiki/x/FOKsbg).

This step by default enforces that the SAP Q-Gate requirements for Checkmarx are met and therefore ensures that:
* No 'To Verify' High and Medium issues exist in your project
* Total number of High and Medium 'Confirmed' or 'Urgent' issues is zero
* 10% of all Low issues are 'Confirmed' or 'Not Exploitable'

For further information please also check [the review guidelines](https://jam4.sapjam.com/wiki/show/rxXj7mOf4mx84p3FSa5Q56)
and [the report generation details](https://jam4.sapjam.com/wiki/show/40zId8lTvuVnKL4m5Y1Qua?_lightbox=true).`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			log.SetStepName("checkmarxExecuteScan")
			log.SetVerbose(GeneralConfig.Verbose)
			return PrepareConfig(cmd, &metadata, "checkmarxExecuteScan", &myCheckmarxExecuteScanOptions, config.OpenPiperFile)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return checkmarxExecuteScan(myCheckmarxExecuteScanOptions)
		},
	}

	addCheckmarxExecuteScanFlags(createCheckmarxExecuteScanCmd)
	return createCheckmarxExecuteScanCmd
}

func addCheckmarxExecuteScanFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&myCheckmarxExecuteScanOptions.FullScanCycle, "fullScanCycle", "5", "Indicates how often a full scan should happen between the incremental scans when activated")
	cmd.Flags().StringVar(&myCheckmarxExecuteScanOptions.VulnerabilityThresholdResult, "vulnerabilityThresholdResult", "FAILURE", "The result of the build in case thresholds are enabled and exceeded")
	cmd.Flags().StringVar(&myCheckmarxExecuteScanOptions.VulnerabilityThresholdUnit, "vulnerabilityThresholdUnit", "percentage", "The unit for the threshold to apply.")
	cmd.Flags().BoolVar(&myCheckmarxExecuteScanOptions.AvoidDuplicateProjectScans, "avoidDuplicateProjectScans", false, "Whether duplicate scans of the same project state shall be avoided or not")
	cmd.Flags().BoolVar(&myCheckmarxExecuteScanOptions.GeneratePdfReport, "generatePdfReport", true, "Whether to generate a PDF report of the analysis results or not")
	cmd.Flags().BoolVar(&myCheckmarxExecuteScanOptions.VulnerabilityThresholdEnabled, "vulnerabilityThresholdEnabled", true, "Whether the thresholds are enabled or not. If enabled the build will be set to `vulnerabilityThresholdResult` in case a specific threshold value is exceeded")
	cmd.Flags().BoolVar(&myCheckmarxExecuteScanOptions.FullScansScheduled, "fullScansScheduled", true, "Whether full scans are to be scheduled or not. Should be used in relation with `incremental` and `fullScanCycle`")
	cmd.Flags().BoolVar(&myCheckmarxExecuteScanOptions.Incremental, "incremental", true, "Whether incremental scans are to be applied which optimizes the scan time but might reduce detection capabilities. Therefore full scans are still required from time to time and should be scheduled via `fullScansScheduled` and `fullScanCycle`")
	cmd.Flags().StringVar(&myCheckmarxExecuteScanOptions.Preset, "preset", os.Getenv("PIPER_preset"), "The preset to use for scanning, if not set explicitly the step will attempt to look up the project's setting based on the availability of `checkmarxCredentialsId`")
	cmd.Flags().StringVar(&myCheckmarxExecuteScanOptions.CheckmarxProject, "checkmarxProject", os.Getenv("PIPER_checkmarxProject"), "The name of the Checkmarx project to scan into")
	cmd.Flags().StringVar(&myCheckmarxExecuteScanOptions.Verbose, "verbose", os.Getenv("PIPER_verbose"), "Enables or disables verbose logging of the step")
	cmd.Flags().StringVar(&myCheckmarxExecuteScanOptions.CheckmarxGroupID, "checkmarxGroupId", os.Getenv("PIPER_checkmarxGroupId"), "The group ID related to your team which can be obtained via the Pipeline Syntax plugin as described in the `Details` section")
	cmd.Flags().StringVar(&myCheckmarxExecuteScanOptions.PullRequestName, "pullRequestName", os.Getenv("PIPER_pullRequestName"), "Used to supply the name for the newly created PR project branch when being used in pull request scenarios")
	cmd.Flags().StringVar(&myCheckmarxExecuteScanOptions.FilterPattern, "filterPattern", "!**/node_modules/**, !**/.xmake/**, !**/*_test.go, !**/vendor/**/*.go, **/*.html, **/*.xml, **/*.go, **/*.py, **/*.js, **/*.scala, **/*.ts", "The filter pattern used to zip the files relevant for scanning, patterns can be negated by setting an exclamation mark in front i.e. `!test/*.js` would avoid adding any javascript files located in the test directory")
	cmd.Flags().StringVar(&myCheckmarxExecuteScanOptions.VulnerabilityThresholdLow, "vulnerabilityThresholdLow", "10", "The specific threshold for low severity findings")
	cmd.Flags().StringVar(&myCheckmarxExecuteScanOptions.SourceEncoding, "sourceEncoding", os.Getenv("PIPER_sourceEncoding"), "The source encoding to be used, if not set explicitly the project's default will be used")
	cmd.Flags().StringVar(&myCheckmarxExecuteScanOptions.VulnerabilityThresholdMedium, "vulnerabilityThresholdMedium", "100", "The specific threshold for medium severity findings")
	cmd.Flags().StringVar(&myCheckmarxExecuteScanOptions.ValidTypeScriptPresets, "validTypeScriptPresets", "map[100131:SAP_Default_Typescript 100154:SAP_Default_TypeScript_JavaScript]", "")
	cmd.Flags().StringVar(&myCheckmarxExecuteScanOptions.CheckmarxServerURL, "checkmarxServerUrl", "https://cx.wdf.sap.corp:443", "The URL pointing to the root of the Checkmarx server to be used")
	cmd.Flags().StringVar(&myCheckmarxExecuteScanOptions.VulnerabilityThresholdHigh, "vulnerabilityThresholdHigh", "100", "The specific threshold for high severity findings")
	cmd.Flags().StringVar(&myCheckmarxExecuteScanOptions.TeamName, "teamName", os.Getenv("PIPER_teamName"), "The name of the team to assign newly created projects to")
	cmd.Flags().StringVar(&myCheckmarxExecuteScanOptions.EngineConfiguration, "engineConfiguration", "1", "The checkmarx engine version to be used")
	cmd.Flags().StringVar(&myCheckmarxExecuteScanOptions.Username, "username", os.Getenv("PIPER_username"), "The username to authenticate")
	cmd.Flags().StringVar(&myCheckmarxExecuteScanOptions.Password, "password", os.Getenv("PIPER_password"), "The password to authenticate")

	cmd.MarkFlagRequired("checkmarxProject")
	cmd.MarkFlagRequired("checkmarxGroupId")
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
}

// retrieve step metadata
func checkmarxExecuteScanMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:      "fullScanCycle",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "vulnerabilityThresholdResult",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "vulnerabilityThresholdUnit",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "avoidDuplicateProjectScans",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "bool",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "generatePdfReport",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "bool",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "vulnerabilityThresholdEnabled",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "bool",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "fullScansScheduled",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "bool",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "incremental",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "bool",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "preset",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "checkmarxProject",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "verbose",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "checkmarxGroupId",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "pullRequestName",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "filterPattern",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "vulnerabilityThresholdLow",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "sourceEncoding",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "vulnerabilityThresholdMedium",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "validTypeScriptPresets",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "checkmarxServerUrl",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "vulnerabilityThresholdHigh",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "teamName",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "engineConfiguration",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "username",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "password",
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
					},
				},
			},
		},
	}
	return theMetaData
}
