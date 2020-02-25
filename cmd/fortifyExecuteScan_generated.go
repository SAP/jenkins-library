package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/spf13/cobra"
)

type fortifyExecuteScanOptions struct {
	MvnCustomArgs                 string `json:"mvnCustomArgs,omitempty"`
	PythonRequirementsFile        string `json:"pythonRequirementsFile,omitempty"`
	PythonVersion                 string `json:"pythonVersion,omitempty"`
	UploadResults                 bool   `json:"uploadResults,omitempty"`
	BuildDescriptorFile           string `json:"buildDescriptorFile,omitempty"`
	CommitID                      string `json:"commitId,omitempty"`
	Repository                    string `json:"repository,omitempty"`
	Memory                        string `json:"memory,omitempty"`
	UpdateRulePack                bool   `json:"updateRulePack,omitempty"`
	PythonExcludes                string `json:"pythonExcludes,omitempty"`
	FortifyReportDownloadEndpoint string `json:"fortifyReportDownloadEndpoint,omitempty"`
	PollingMinutes                int    `json:"pollingMinutes,omitempty"`
	QuickScan                     bool   `json:"quickScan,omitempty"`
	Translate                     string `json:"translate,omitempty"`
	FortifyAPIEndpoint            string `json:"fortifyApiEndpoint,omitempty"`
	ReportType                    string `json:"reportType,omitempty"`
	GitTreeish                    string `json:"gitTreeish,omitempty"`
	XMakeJobName                  string `json:"xMakeJobName,omitempty"`
	PythonAdditionalPath          string `json:"pythonAdditionalPath,omitempty"`
	ArtifactURL                   string `json:"artifactUrl,omitempty"`
	ConsiderSuspicious            bool   `json:"considerSuspicious,omitempty"`
	FortifyFprUploadEndpoint      string `json:"fortifyFprUploadEndpoint,omitempty"`
	FortifyProjectName            string `json:"fortifyProjectName,omitempty"`
	PythonIncludes                string `json:"pythonIncludes,omitempty"`
	Reporting                     bool   `json:"reporting,omitempty"`
	FortifyServerURL              string `json:"fortifyServerUrl,omitempty"`
	BuildDescriptorExcludeList    string `json:"buildDescriptorExcludeList,omitempty"`
	PullRequestMessageRegexGroup  int    `json:"pullRequestMessageRegexGroup,omitempty"`
	DeltaMinutes                  int    `json:"deltaMinutes,omitempty"`
	SpotCheckMinimum              int    `json:"spotCheckMinimum,omitempty"`
	FortifyFprDownloadEndpoint    string `json:"fortifyFprDownloadEndpoint,omitempty"`
	FortifyProjectVersion         string `json:"fortifyProjectVersion,omitempty"`
	PythonInstallCommand          string `json:"pythonInstallCommand,omitempty"`
	Environment                   string `json:"environment,omitempty"`
	PullRequestName               string `json:"pullRequestName,omitempty"`
	NameVersionMapping            string `json:"nameVersionMapping,omitempty"`
	PullRequestMessageRegex       string `json:"pullRequestMessageRegex,omitempty"`
	XMakeServer                   string `json:"xMakeServer,omitempty"`
	ScanType                      string `json:"scanType,omitempty"`
}

type fortifyExecuteScanInflux struct {
	fortify_data struct {
		fields struct {
			projectName       string
			projectVersion    string
			violations        string
			corporateTotal    string
			corporateAudited  string
			auditAllTotal     string
			auditAllAudited   string
			spotChecksTotal   string
			spotChecksAudited string
			suspicious        string
			exploitable       string
			suppressed        string
		}
		tags struct {
		}
	}
}

func (i *fortifyExecuteScanInflux) persist(path, resourceName string) {
	measurementContent := []struct {
		measurement string
		valType     string
		name        string
		value       string
	}{
		{valType: config.InfluxField, measurement: "fortify_data", name: "projectName", value: i.fortify_data.fields.projectName},
		{valType: config.InfluxField, measurement: "fortify_data", name: "projectVersion", value: i.fortify_data.fields.projectVersion},
		{valType: config.InfluxField, measurement: "fortify_data", name: "violations", value: i.fortify_data.fields.violations},
		{valType: config.InfluxField, measurement: "fortify_data", name: "corporateTotal", value: i.fortify_data.fields.corporateTotal},
		{valType: config.InfluxField, measurement: "fortify_data", name: "corporateAudited", value: i.fortify_data.fields.corporateAudited},
		{valType: config.InfluxField, measurement: "fortify_data", name: "auditAllTotal", value: i.fortify_data.fields.auditAllTotal},
		{valType: config.InfluxField, measurement: "fortify_data", name: "auditAllAudited", value: i.fortify_data.fields.auditAllAudited},
		{valType: config.InfluxField, measurement: "fortify_data", name: "spotChecksTotal", value: i.fortify_data.fields.spotChecksTotal},
		{valType: config.InfluxField, measurement: "fortify_data", name: "spotChecksAudited", value: i.fortify_data.fields.spotChecksAudited},
		{valType: config.InfluxField, measurement: "fortify_data", name: "suspicious", value: i.fortify_data.fields.suspicious},
		{valType: config.InfluxField, measurement: "fortify_data", name: "exploitable", value: i.fortify_data.fields.exploitable},
		{valType: config.InfluxField, measurement: "fortify_data", name: "suppressed", value: i.fortify_data.fields.suppressed},
	}

	errCount := 0
	for _, metric := range measurementContent {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(metric.measurement, fmt.Sprintf("%vs", metric.valType), metric.name), metric.value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting influx environment.")
			errCount++
		}
	}
	if errCount > 0 {
		os.Exit(1)
	}
}

// FortifyExecuteScanCommand This step executes a Fortify scan on the specified project to perform static code analysis and check the source code for security flaws.
func FortifyExecuteScanCommand() *cobra.Command {
	metadata := fortifyExecuteScanMetadata()
	var stepConfig fortifyExecuteScanOptions
	var startTime time.Time
	var influx fortifyExecuteScanInflux

	var createFortifyExecuteScanCmd = &cobra.Command{
		Use:   "fortifyExecuteScan",
		Short: "This step executes a Fortify scan on the specified project to perform static code analysis and check the source code for security flaws.",
		Long: `This step executes a Fortify scan on the specified project to perform static code analysis and check the source code for security flaws.
The Fortify step supports two ways of running a scan. It can either trigger the xMake OD job if the ` + "`" + `environment` + "`" + ` parameter is set to value
` + "`" + `'xMake'` + "`" + ` or it can trigger a scan locally on your Jenkins within a docker container if the parameter value is set to ` + "`" + `'docker'` + "`" + `.

!!! error
    When moving your project to Fortify template 18.10 or creating a new one on this template version please make sure to generate a new access token
    for the Fortify backend. Unfortunately the template version change forced us to use new APIs and the updated permission set is not reflected to existing tokens!

!!! warning
    To scan MTA projects via ` + "`" + `scanType: 'mta'` + "`" + ` you will have to switch to ` + "`" + `environment: 'docker'` + "`" + `. xMake does NOT support scanning MTA structures with Fortify.

!!! warning
    The ` + "`" + `ci_pipeline` + "`" + ` default Fortify user has been deprecated and deleted mid of March 2018. Please check the updated Prerequisites section below for the
    changes you need to perform to your configuration.

For more details related to Fortify please consult the related [JAM group](https://jam4.sapjam.com/groups/hwsYB62safobfg6sX9QrYW/overview_page/W8SJMLfSSBMgWcHS4NQbt7)`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			startTime = time.Now()
			log.SetStepName("fortifyExecuteScan")
			log.SetVerbose(GeneralConfig.Verbose)
			return PrepareConfig(cmd, &metadata, "fortifyExecuteScan", &stepConfig, config.OpenPiperFile)
		},
		Run: func(cmd *cobra.Command, args []string) {
			telemetryData := telemetry.CustomData{}
			telemetryData.ErrorCode = "1"
			handler := func() {
				influx.persist(GeneralConfig.EnvRootPath, "influx")
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetry.Send(&telemetryData)
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, "fortifyExecuteScan")
			fortifyExecuteScan(stepConfig, &telemetryData, &influx)
			telemetryData.ErrorCode = "0"
		},
	}

	addFortifyExecuteScanFlags(createFortifyExecuteScanCmd, &stepConfig)
	return createFortifyExecuteScanCmd
}

func addFortifyExecuteScanFlags(cmd *cobra.Command, stepConfig *fortifyExecuteScanOptions) {
	cmd.Flags().StringVar(&stepConfig.MvnCustomArgs, "mvnCustomArgs", "", "Allows providing additional Maven command line parameters")
	cmd.Flags().StringVar(&stepConfig.PythonRequirementsFile, "pythonRequirementsFile", os.Getenv("PIPER_pythonRequirementsFile"), "The requirements file used in `scanType: 'pip'` to populate the build environment with the necessary dependencies")
	cmd.Flags().StringVar(&stepConfig.PythonVersion, "pythonVersion", "python3", "Python version to be used in `scanType: 'pip'`")
	cmd.Flags().BoolVar(&stepConfig.UploadResults, "uploadResults", true, "Whether results shall be uploaded or not")
	cmd.Flags().StringVar(&stepConfig.BuildDescriptorFile, "buildDescriptorFile", os.Getenv("PIPER_buildDescriptorFile"), "Path to the build descriptor file addressing the module/folder to be scanned. Defaults are for scanType=`maven`: `./pom.xml`, scanType=`pip`: `./setup.py`, scanType=`mta`: determined automatically")
	cmd.Flags().StringVar(&stepConfig.CommitID, "commitId", os.Getenv("PIPER_commitId"), "Set the Git commit ID for identifing artifacts throughout the scan.")
	cmd.Flags().StringVar(&stepConfig.Repository, "repository", os.Getenv("PIPER_repository"), "Set the GitHub repository for identifing artifacts throughout the scan.")
	cmd.Flags().StringVar(&stepConfig.Memory, "memory", "-Xmx4G -Xms512M", "The amount of memory granted to the translate/scan executions")
	cmd.Flags().BoolVar(&stepConfig.UpdateRulePack, "updateRulePack", true, "Whether the rule pack shall be updated and pulled from Fortify SSC before scanning or not")
	cmd.Flags().StringVar(&stepConfig.PythonExcludes, "pythonExcludes", "-exclude ./**/test/**/*", "The excludes pattern used in `scanType: 'pip'` for excluding specific .py files i.e. tests")
	cmd.Flags().StringVar(&stepConfig.FortifyReportDownloadEndpoint, "fortifyReportDownloadEndpoint", "/transfer/reportDownload.html", "Fortify SSC endpoint for Report downloads")
	cmd.Flags().IntVar(&stepConfig.PollingMinutes, "pollingMinutes", 30, "The number of minutes for which an uploaded FPR artifact's status is being polled to finish queuing/processing, if exceeded polling will be stopped and an error will be thrown")
	cmd.Flags().BoolVar(&stepConfig.QuickScan, "quickScan", false, "Whether a quick scan should be performed, please consult the related Fortify documentation on JAM on the impact of this setting")
	cmd.Flags().StringVar(&stepConfig.Translate, "translate", os.Getenv("PIPER_translate"), "Array of maps with required key `'src'`, and optional keys `'exclude'`, `'libDirs'`, `'aspnetcore'`, and `'dotNetCoreVersion'`")
	cmd.Flags().StringVar(&stepConfig.FortifyAPIEndpoint, "fortifyApiEndpoint", "/api/v1", "Fortify SSC endpoint used for uploading the scan results and checking the audit state")
	cmd.Flags().StringVar(&stepConfig.ReportType, "reportType", "PDF", "The type of report to be generated")
	cmd.Flags().StringVar(&stepConfig.GitTreeish, "gitTreeish", os.Getenv("PIPER_gitTreeish"), "Identifies the commit/tag/branch used for scanning in `environment: 'xmake', prepoulated by the pipeline with the related commit id")
	cmd.Flags().StringVar(&stepConfig.XMakeJobName, "xMakeJobName", "${githubOrg}-${githubRepo}-OD-fortify-fortify", "The name of the job in xMake to be triggered for scanning, usually autoconstructed by the pipeline based on  the github repo and org information")
	cmd.Flags().StringVar(&stepConfig.PythonAdditionalPath, "pythonAdditionalPath", "./lib", "The addional path which can be used in `scanType: 'pip'` for customization purposes")
	cmd.Flags().StringVar(&stepConfig.ArtifactURL, "artifactUrl", os.Getenv("PIPER_artifactUrl"), "Path/Url pointing to an additional artifact repository for resolution of additional artifacts during the build")
	cmd.Flags().BoolVar(&stepConfig.ConsiderSuspicious, "considerSuspicious", true, "Whether suspicious issues should trigger the check to fail or not")
	cmd.Flags().StringVar(&stepConfig.FortifyFprUploadEndpoint, "fortifyFprUploadEndpoint", "/upload/resultFileUpload.html", "Fortify SSC endpoint for FPR uploads")
	cmd.Flags().StringVar(&stepConfig.FortifyProjectName, "fortifyProjectName", "${group}-${artifact}", "The project used for reporting results in SSC")
	cmd.Flags().StringVar(&stepConfig.PythonIncludes, "pythonIncludes", "./**/*", "The includes pattern used in `scanType: 'pip'` for including .py files")
	cmd.Flags().BoolVar(&stepConfig.Reporting, "reporting", false, "Influences whether a report is generated or not")
	cmd.Flags().StringVar(&stepConfig.FortifyServerURL, "fortifyServerUrl", "https://fortify.mo.sap.corp/ssc", "Fortify SSC Url to be used for accessing the APIs")
	cmd.Flags().StringVar(&stepConfig.BuildDescriptorExcludeList, "buildDescriptorExcludeList", "[]", "Build descriptor files to exclude modules from being scanned")
	cmd.Flags().IntVar(&stepConfig.PullRequestMessageRegexGroup, "pullRequestMessageRegexGroup", 1, "The group number for extracting the pull request id in `pullRequestMessageRegex`")
	cmd.Flags().IntVar(&stepConfig.DeltaMinutes, "deltaMinutes", 5, "The number of minutes for which an uploaded FPR artifact is considered to be recent and healthy, if exceeded an error will be thrown")
	cmd.Flags().IntVar(&stepConfig.SpotCheckMinimum, "spotCheckMinimum", 1, "The minimum number of issues that must be audited per category in the `Spot Checks of each Category` folder to avoid an error being thrown")
	cmd.Flags().StringVar(&stepConfig.FortifyFprDownloadEndpoint, "fortifyFprDownloadEndpoint", "/download/currentStateFprDownload.html", "Fortify SSC endpoint  for FPR downloads")
	cmd.Flags().StringVar(&stepConfig.FortifyProjectVersion, "fortifyProjectVersion", "${version}", "The project version used for reporting results in SSC")
	cmd.Flags().StringVar(&stepConfig.PythonInstallCommand, "pythonInstallCommand", "${pip} install --user --index-url http://nexus.wdf.sap.corp:8081/nexus/content/groups/build.snapshots.pypi/simple/ --trusted-host nexus.wdf.sap.corp .", "Additional install command that can be run when `scanType: 'pip'` is used which allows further customizing the execution environment of the scan")
	cmd.Flags().StringVar(&stepConfig.Environment, "environment", "docker", "The environment used to run the Fortify scan.")
	cmd.Flags().StringVar(&stepConfig.PullRequestName, "pullRequestName", os.Getenv("PIPER_pullRequestName"), "The name of the pull request branch which will trigger creation of a new version in Fortify SSC based on the master branch version")
	cmd.Flags().StringVar(&stepConfig.NameVersionMapping, "nameVersionMapping", os.Getenv("PIPER_nameVersionMapping"), "Allows modifying associated project name and version in `scanType: 'mta'` with a map of lists where the map's key is the path to the build descriptor file and the list value contains project name as first, and project version as second parameter, those may be `null` to force only overwriting one parameter")
	cmd.Flags().StringVar(&stepConfig.PullRequestMessageRegex, "pullRequestMessageRegex", ".*Merge pull request #(\\d+) from.*", "Regex used to identify the PR-XXX reference within the merge commit message")
	cmd.Flags().StringVar(&stepConfig.XMakeServer, "xMakeServer", "xmake-dev", "The Jenkins server URL used for triggering the xMake remote job")
	cmd.Flags().StringVar(&stepConfig.ScanType, "scanType", "maven", "Scan type used for the step which can be `'maven'`, `'pip'`")

}

// retrieve step metadata
func fortifyExecuteScanMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "mvnCustomArgs",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "pythonRequirementsFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "pythonVersion",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "uploadResults",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "buildDescriptorFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "commitId",
						ResourceRef: []config.ResourceReference{{Name: "commonPipelineEnvironment", Param: "git/commitId"}},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "repository",
						ResourceRef: []config.ResourceReference{{Name: "commonPipelineEnvironment", Param: "github/repository"}},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "memory",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "updateRulePack",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "pythonExcludes",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "fortifyReportDownloadEndpoint",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "pollingMinutes",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "int",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "quickScan",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "translate",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "fortifyApiEndpoint",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "reportType",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "gitTreeish",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "xMakeJobName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "pythonAdditionalPath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "artifactUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "considerSuspicious",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "fortifyFprUploadEndpoint",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "fortifyProjectName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "pythonIncludes",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "reporting",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "fortifyServerUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "buildDescriptorExcludeList",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "pullRequestMessageRegexGroup",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "int",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "deltaMinutes",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "int",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "spotCheckMinimum",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "int",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "fortifyFprDownloadEndpoint",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "fortifyProjectVersion",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "pythonInstallCommand",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "environment",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "pullRequestName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "nameVersionMapping",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "pullRequestMessageRegex",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "xMakeServer",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "scanType",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
				},
			},
		},
	}
	return theMetaData
}
