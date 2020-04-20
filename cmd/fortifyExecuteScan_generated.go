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
	AuthToken                       string `json:"authToken,omitempty"`
	AutoCreate                      bool   `json:"autoCreate,omitempty"`
	MvnCustomArgs                   string `json:"mvnCustomArgs,omitempty"`
	ModulePath                      string `json:"modulePath,omitempty"`
	PythonRequirementsFile          string `json:"pythonRequirementsFile,omitempty"`
	AutodetectClasspath             bool   `json:"autodetectClasspath,omitempty"`
	PythonRequirementsInstallSuffix string `json:"pythonRequirementsInstallSuffix,omitempty"`
	PythonVersion                   string `json:"pythonVersion,omitempty"`
	UploadResults                   bool   `json:"uploadResults,omitempty"`
	BuildDescriptorFile             string `json:"buildDescriptorFile,omitempty"`
	CommitID                        string `json:"commitId,omitempty"`
	CommitMessage                   string `json:"commitMessage,omitempty"`
	RepoURL                         string `json:"repoUrl,omitempty"`
	Repository                      string `json:"repository,omitempty"`
	Memory                          string `json:"memory,omitempty"`
	UpdateRulePack                  bool   `json:"updateRulePack,omitempty"`
	PythonExcludes                  string `json:"pythonExcludes,omitempty"`
	ReportDownloadEndpoint          string `json:"reportDownloadEndpoint,omitempty"`
	PollingMinutes                  int    `json:"pollingMinutes,omitempty"`
	QuickScan                       bool   `json:"quickScan,omitempty"`
	Translate                       string `json:"translate,omitempty"`
	APIEndpoint                     string `json:"apiEndpoint,omitempty"`
	ReportType                      string `json:"reportType,omitempty"`
	PythonAdditionalPath            string `json:"pythonAdditionalPath,omitempty"`
	ArtifactURL                     string `json:"artifactUrl,omitempty"`
	ConsiderSuspicious              bool   `json:"considerSuspicious,omitempty"`
	FprUploadEndpoint               string `json:"fprUploadEndpoint,omitempty"`
	ProjectName                     string `json:"projectName,omitempty"`
	PythonIncludes                  string `json:"pythonIncludes,omitempty"`
	Reporting                       bool   `json:"reporting,omitempty"`
	ServerURL                       string `json:"serverUrl,omitempty"`
	BuildDescriptorExcludeList      string `json:"buildDescriptorExcludeList,omitempty"`
	PullRequestMessageRegexGroup    int    `json:"pullRequestMessageRegexGroup,omitempty"`
	DeltaMinutes                    int    `json:"deltaMinutes,omitempty"`
	SpotCheckMinimum                int    `json:"spotCheckMinimum,omitempty"`
	FprDownloadEndpoint             string `json:"fprDownloadEndpoint,omitempty"`
	ProjectVersion                  string `json:"projectVersion,omitempty"`
	ProjectVersioningScheme         string `json:"projectVersioningScheme,omitempty"`
	PythonInstallCommand            string `json:"pythonInstallCommand,omitempty"`
	ReportTemplateID                int    `json:"reportTemplateId,omitempty"`
	FilterSetTitle                  string `json:"filterSetTitle,omitempty"`
	PullRequestName                 string `json:"pullRequestName,omitempty"`
	NameVersionMapping              string `json:"nameVersionMapping,omitempty"`
	PullRequestMessageRegex         string `json:"pullRequestMessageRegex,omitempty"`
	ScanType                        string `json:"scanType,omitempty"`
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

The Fortify step triggers a scan locally on your Jenkins within a docker container so finally you have to supply a docker image with a Fortify SCA
and Java plus Maven or alternatively Python installed into it for being able to perform any scans.`,
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
	cmd.Flags().StringVar(&stepConfig.AuthToken, "authToken", os.Getenv("PIPER_authToken"), "The FortifyToken to use for authentication")
	cmd.Flags().BoolVar(&stepConfig.AutoCreate, "autoCreate", false, "Whether Fortify project and project version shall be implicitly auto created in case they cannot be found in the backend")
	cmd.Flags().StringVar(&stepConfig.MvnCustomArgs, "mvnCustomArgs", ``, "Allows providing additional Maven command line parameters")
	cmd.Flags().StringVar(&stepConfig.ModulePath, "modulePath", `./`, "Allows providing the path for the module to scan")
	cmd.Flags().StringVar(&stepConfig.PythonRequirementsFile, "pythonRequirementsFile", os.Getenv("PIPER_pythonRequirementsFile"), "The requirements file used in `scanType: 'pip'` to populate the build environment with the necessary dependencies")
	cmd.Flags().BoolVar(&stepConfig.AutodetectClasspath, "autodetectClasspath", true, "Whether the classpath is automatically determined via build tool i.e. maven or pip or not at all")
	cmd.Flags().StringVar(&stepConfig.PythonRequirementsInstallSuffix, "pythonRequirementsInstallSuffix", os.Getenv("PIPER_pythonRequirementsInstallSuffix"), "The suffix for the command used to install the requirements file in `scanType: 'pip'` to populate the build environment with the necessary dependencies")
	cmd.Flags().StringVar(&stepConfig.PythonVersion, "pythonVersion", `python3`, "Python version to be used in `scanType: 'pip'`")
	cmd.Flags().BoolVar(&stepConfig.UploadResults, "uploadResults", true, "Whether results shall be uploaded or not")
	cmd.Flags().StringVar(&stepConfig.BuildDescriptorFile, "buildDescriptorFile", os.Getenv("PIPER_buildDescriptorFile"), "Path to the build descriptor file addressing the module/folder to be scanned. Defaults are for scanType=`maven`: `./pom.xml`, scanType=`pip`: `./setup.py`, scanType=`mta`: determined automatically")
	cmd.Flags().StringVar(&stepConfig.CommitID, "commitId", os.Getenv("PIPER_commitId"), "Set the Git commit ID for identifing artifacts throughout the scan.")
	cmd.Flags().StringVar(&stepConfig.CommitMessage, "commitMessage", os.Getenv("PIPER_commitMessage"), "Set the Git commit message for identifing pull request merges throughout the scan.")
	cmd.Flags().StringVar(&stepConfig.RepoURL, "repoUrl", os.Getenv("PIPER_repoUrl"), "Set the source code repository URL for identifing sources of the scan.")
	cmd.Flags().StringVar(&stepConfig.Repository, "repository", os.Getenv("PIPER_repository"), "Set the GitHub repository for identifing artifacts throughout the scan.")
	cmd.Flags().StringVar(&stepConfig.Memory, "memory", `-Xmx4G -Xms512M`, "The amount of memory granted to the translate/scan executions")
	cmd.Flags().BoolVar(&stepConfig.UpdateRulePack, "updateRulePack", true, "Whether the rule pack shall be updated and pulled from Fortify SSC before scanning or not")
	cmd.Flags().StringVar(&stepConfig.PythonExcludes, "pythonExcludes", `-exclude ./**/tests/**/*;./**/setup.py`, "The excludes pattern used in `scanType: 'pip'` for excluding specific .py files i.e. tests")
	cmd.Flags().StringVar(&stepConfig.ReportDownloadEndpoint, "reportDownloadEndpoint", `/transfer/reportDownload.html`, "Fortify SSC endpoint for Report downloads")
	cmd.Flags().IntVar(&stepConfig.PollingMinutes, "pollingMinutes", 30, "The number of minutes for which an uploaded FPR artifact's status is being polled to finish queuing/processing, if exceeded polling will be stopped and an error will be thrown")
	cmd.Flags().BoolVar(&stepConfig.QuickScan, "quickScan", false, "Whether a quick scan should be performed, please consult the related Fortify documentation on JAM on the impact of this setting")
	cmd.Flags().StringVar(&stepConfig.Translate, "translate", os.Getenv("PIPER_translate"), "JSON string of list of maps with required key `'src'`, and optional keys `'exclude'`, `'libDirs'`, `'aspnetcore'`, and `'dotNetCoreVersion'`")
	cmd.Flags().StringVar(&stepConfig.APIEndpoint, "apiEndpoint", `/api/v1`, "Fortify SSC endpoint used for uploading the scan results and checking the audit state")
	cmd.Flags().StringVar(&stepConfig.ReportType, "reportType", `PDF`, "The type of report to be generated")
	cmd.Flags().StringVar(&stepConfig.PythonAdditionalPath, "pythonAdditionalPath", `./lib`, "The addional path which can be used in `scanType: 'pip'` for customization purposes")
	cmd.Flags().StringVar(&stepConfig.ArtifactURL, "artifactUrl", os.Getenv("PIPER_artifactUrl"), "Path/Url pointing to an additional artifact repository for resolution of additional artifacts during the build")
	cmd.Flags().BoolVar(&stepConfig.ConsiderSuspicious, "considerSuspicious", true, "Whether suspicious issues should trigger the check to fail or not")
	cmd.Flags().StringVar(&stepConfig.FprUploadEndpoint, "fprUploadEndpoint", `/upload/resultFileUpload.html`, "Fortify SSC endpoint for FPR uploads")
	cmd.Flags().StringVar(&stepConfig.ProjectName, "projectName", `{{list .GroupID .ArtifactID | join "-" | trimAll "-"}}`, "The project used for reporting results in SSC")
	cmd.Flags().StringVar(&stepConfig.PythonIncludes, "pythonIncludes", `./**/*`, "The includes pattern used in `scanType: 'pip'` for including .py files")
	cmd.Flags().BoolVar(&stepConfig.Reporting, "reporting", false, "Influences whether a report is generated or not")
	cmd.Flags().StringVar(&stepConfig.ServerURL, "serverUrl", os.Getenv("PIPER_serverUrl"), "Fortify SSC Url to be used for accessing the APIs")
	cmd.Flags().StringVar(&stepConfig.BuildDescriptorExcludeList, "buildDescriptorExcludeList", `[]`, "Build descriptor files to exclude modules from being scanned")
	cmd.Flags().IntVar(&stepConfig.PullRequestMessageRegexGroup, "pullRequestMessageRegexGroup", 1, "The group number for extracting the pull request id in `pullRequestMessageRegex`")
	cmd.Flags().IntVar(&stepConfig.DeltaMinutes, "deltaMinutes", 5, "The number of minutes for which an uploaded FPR artifact is considered to be recent and healthy, if exceeded an error will be thrown")
	cmd.Flags().IntVar(&stepConfig.SpotCheckMinimum, "spotCheckMinimum", 1, "The minimum number of issues that must be audited per category in the `Spot Checks of each Category` folder to avoid an error being thrown")
	cmd.Flags().StringVar(&stepConfig.FprDownloadEndpoint, "fprDownloadEndpoint", `/download/currentStateFprDownload.html`, "Fortify SSC endpoint  for FPR downloads")
	cmd.Flags().StringVar(&stepConfig.ProjectVersion, "projectVersion", `{{(split "." (split "-" .Version)._0)._0}}`, "The project version used for reporting results in SSC")
	cmd.Flags().StringVar(&stepConfig.ProjectVersioningScheme, "projectVersioningScheme", `major`, "The project versioning scheme used for creating the version to report results in SSC, can be one of `'major'`, `'semantic'`, `'full'`, `'text'`")
	cmd.Flags().StringVar(&stepConfig.PythonInstallCommand, "pythonInstallCommand", `{{.Pip}} install --user .`, "Additional install command that can be run when `scanType: 'pip'` is used which allows further customizing the execution environment of the scan")
	cmd.Flags().IntVar(&stepConfig.ReportTemplateID, "reportTemplateId", 18, "Report template ID to be used for generating the Fortify report")
	cmd.Flags().StringVar(&stepConfig.FilterSetTitle, "filterSetTitle", `SAP`, "Title of the filter set to use for analysing the results")
	cmd.Flags().StringVar(&stepConfig.PullRequestName, "pullRequestName", os.Getenv("PIPER_pullRequestName"), "The name of the pull request branch which will trigger creation of a new version in Fortify SSC based on the master branch version")
	cmd.Flags().StringVar(&stepConfig.NameVersionMapping, "nameVersionMapping", os.Getenv("PIPER_nameVersionMapping"), "Allows modifying associated project name and version in `scanType: 'mta'` with a map of lists where the map's key is the path to the build descriptor file and the list value contains project name as first, and project version as second parameter, those may be `null` to force only overwriting one parameter")
	cmd.Flags().StringVar(&stepConfig.PullRequestMessageRegex, "pullRequestMessageRegex", `.*Merge pull request #(\\d+) from.*`, "Regex used to identify the PR-XXX reference within the merge commit message")
	cmd.Flags().StringVar(&stepConfig.ScanType, "scanType", `maven`, "Scan type used for the step which can be `'maven'`, `'pip'`")

	cmd.MarkFlagRequired("authToken")
}

// retrieve step metadata
func fortifyExecuteScanMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "authToken",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "autoCreate",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "mvnCustomArgs",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "modulePath",
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
						Name:        "autodetectClasspath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "pythonRequirementsInstallSuffix",
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
						Name:        "commitMessage",
						ResourceRef: []config.ResourceReference{{Name: "commonPipelineEnvironment", Param: "git/commitMessage"}},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "repoUrl",
						ResourceRef: []config.ResourceReference{{Name: "commonPipelineEnvironment", Param: "gitHttpsUrl"}},
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
						Name:        "reportDownloadEndpoint",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "fortifyReportDownloadEndpoint"}},
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
						Name:        "apiEndpoint",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "fortifyApiEndpoint"}},
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
						Name:        "fprUploadEndpoint",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "fortifyFprUploadEndpoint"}},
					},
					{
						Name:        "projectName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "fortifyProjectName"}},
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
						Name:        "serverUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "fortifyServerUrl"}},
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
						Name:        "fprDownloadEndpoint",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "fortifyFprDownloadEndpoint"}},
					},
					{
						Name:        "projectVersion",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "fortifyProjectVersion"}},
					},
					{
						Name:        "projectVersioningScheme",
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
						Name:        "reportTemplateId",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "int",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "filterSetTitle",
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
