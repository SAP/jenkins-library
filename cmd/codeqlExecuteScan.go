package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/codeql"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

type codeqlExecuteScanUtils interface {
	command.ExecRunner

	piperutils.FileUtils

	DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error
}

type codeqlExecuteScanUtilsBundle struct {
	*command.Command
	*piperutils.Files
	*piperhttp.Client
}

func newCodeqlExecuteScanUtils() codeqlExecuteScanUtils {
	utils := codeqlExecuteScanUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
		Client:  &piperhttp.Client{},
	}

	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func codeqlExecuteScan(config codeqlExecuteScanOptions, telemetryData *telemetry.CustomData, influx *codeqlExecuteScanInflux) {
	utils := newCodeqlExecuteScanUtils()

	influx.step_data.fields.codeql = false

	reports, err := runCodeqlExecuteScan(&config, telemetryData, utils, influx)
	piperutils.PersistReportsAndLinks("codeqlExecuteScan", "./", utils, reports, nil)

	if err != nil {
		log.Entry().WithError(err).Fatal("Codeql scan failed")
	}
	influx.step_data.fields.codeql = true
}

func appendCodeqlQuery(cmd []string, codeqlQuery string) []string {
	if len(codeqlQuery) > 0 {
		cmd = append(cmd, codeqlQuery)
	}

	return cmd
}

func execute(utils codeqlExecuteScanUtils, cmd []string, isVerbose bool) error {
	if isVerbose {
		cmd = append(cmd, "-v")
	}

	return utils.RunExecutable("codeql", cmd...)
}

func getLangFromBuildTool(buildTool string) string {
	switch buildTool {
	case "maven":
		return "java"
	case "pip":
		return "python"
	case "npm":
		return "javascript"
	case "yarn":
		return "javascript"
	case "golang":
		return "go"
	default:
		return ""
	}
}

func getToken(config *codeqlExecuteScanOptions) (bool, string) {
	if len(config.GithubToken) > 0 {
		return true, config.GithubToken
	}

	envVal, isEnvGithubToken := os.LookupEnv("GITHUB_TOKEN")
	if isEnvGithubToken {
		return true, envVal
	}

	return false, ""
}

func printCodeqlImageVersion() {
	codeqlVersion, err := os.ReadFile("/etc/image-version")
	if err != nil {
		log.Entry().Infof("CodeQL image version: unknown")
	} else {
		log.Entry().Infof("CodeQL image version: %s", string(codeqlVersion))
	}
}

func runCodeqlExecuteScan(config *codeqlExecuteScanOptions, telemetryData *telemetry.CustomData, utils codeqlExecuteScanUtils, influx *codeqlExecuteScanInflux) ([]piperutils.Path, error) {
	printCodeqlImageVersion()

	var reports []piperutils.Path

	dbCreateCustomFlags := codeql.ParseCustomFlags(config.DatabaseCreateFlags)
	err := runDatabaseCreate(config, dbCreateCustomFlags, utils)
	if err != nil {
		log.Entry().WithError(err).Error("failed to create codeql database")
		return reports, err
	}

	err = os.MkdirAll(filepath.Join(config.ModulePath, "target"), os.ModePerm)
	if err != nil {
		log.Entry().WithError(err).Error("failed to create output directory for reports")
		return reports, err
	}

	dbAnalyzeCustomFlags := codeql.ParseCustomFlags(config.DatabaseAnalyzeFlags)
	scanReports, err := runDatabaseAnalyze(config, dbAnalyzeCustomFlags, utils)
	if err != nil {
		log.Entry().WithError(err).Error("failed to analyze codeql database")
		return reports, err
	}
	reports = append(reports, scanReports...)

	repoInfo, err := codeql.GetRepoInfo(config.Repository, config.AnalyzedRef, config.CommitID,
		config.TargetGithubRepoURL, config.TargetGithubBranchName)
	if err != nil {
		log.Entry().WithError(err).Error("failed to get repository info")
		return reports, err
	}

	if len(config.TargetGithubRepoURL) > 0 {
		err = uploadProjectToGitHub(config, repoInfo)
		if err != nil {
			log.Entry().WithError(err).Error("failed to upload project to Github")
			return reports, err
		}
	}

	var scanResults []codeql.CodeqlFindings
	if !config.UploadResults {
		log.Entry().Warn("The sarif results will not be uploaded to the repository and compliance report will not be generated as uploadResults is set to false.")
	} else {
		log.Entry().Infof("The sarif results will be uploaded to the repository %s", repoInfo.FullUrl)

		hasToken, token := getToken(config)
		if !hasToken {
			return reports, fmt.Errorf("failed running upload-results as githubToken was not specified")
		}

		err = uploadSarifResults(config, token, repoInfo, utils)
		if err != nil {
			log.Entry().WithError(err).Error("failed to upload sarif results")
			return reports, err
		}

		codeqlScanAuditInstance := codeql.NewCodeqlScanAuditInstance(repoInfo.ServerUrl, repoInfo.Owner, repoInfo.Repo, token, []string{})
		scanResults, err = codeqlScanAuditInstance.GetVulnerabilities(repoInfo.AnalyzedRef)
		if err != nil {
			log.Entry().WithError(err).Error("failed to get vulnerabilities")
			return reports, err
		}

		codeqlAudit := codeql.CodeqlAudit{
			ToolName:               "codeql",
			RepositoryUrl:          repoInfo.FullUrl,
			CodeScanningLink:       repoInfo.ScanUrl,
			RepositoryReferenceUrl: repoInfo.FullRef,
			QuerySuite:             config.QuerySuite,
			ScanResults:            scanResults,
		}
		paths, err := codeql.WriteJSONReport(codeqlAudit, config.ModulePath)
		if err != nil {
			log.Entry().WithError(err).Error("failed to write json compliance report")
			return reports, err
		}
		reports = append(reports, paths...)

		if config.CheckForCompliance {
			err = checkForCompliance(scanResults, config, repoInfo)
			if err != nil {
				return reports, err
			}
		}
	}

	addDataToInfluxDB(repoInfo, config.QuerySuite, scanResults, influx)

	toolRecordFileName, err := codeql.CreateAndPersistToolRecord(utils, repoInfo, config.ModulePath)
	if err != nil {
		log.Entry().Warning("TR_CODEQL: Failed to create toolrecord file ...", err)
	} else {
		reports = append(reports, piperutils.Path{Target: toolRecordFileName})
	}

	return reports, nil
}

func runDatabaseCreate(config *codeqlExecuteScanOptions, customFlags map[string]string, utils codeqlExecuteScanUtils) error {
	cmd, err := prepareCmdForDatabaseCreate(customFlags, config, utils)
	if err != nil {
		log.Entry().Error("failed to prepare command for codeql database create")
		return err
	}
	if err = execute(utils, cmd, GeneralConfig.Verbose); err != nil {
		log.Entry().Error("failed running command codeql database create")
		return err
	}
	return nil
}

func runDatabaseAnalyze(config *codeqlExecuteScanOptions, customFlags map[string]string, utils codeqlExecuteScanUtils) ([]piperutils.Path, error) {
	sarifReport, err := executeAnalysis("sarif-latest", "codeqlReport.sarif", customFlags, config, utils)
	if err != nil {
		return nil, err
	}
	csvReport, err := executeAnalysis("csv", "codeqlReport.csv", customFlags, config, utils)
	if err != nil {
		return nil, err
	}
	return append(sarifReport, csvReport...), nil
}

func runGithubUploadResults(config *codeqlExecuteScanOptions, repoInfo *codeql.RepoInfo, token string, utils codeqlExecuteScanUtils) (string, error) {
	cmd := prepareCmdForUploadResults(config, repoInfo, token)

	var bufferOut, bufferErr bytes.Buffer
	utils.Stdout(&bufferOut)
	defer utils.Stdout(log.Writer())
	utils.Stderr(&bufferErr)
	defer utils.Stderr(log.Writer())

	if err := execute(utils, cmd, GeneralConfig.Verbose); err != nil {
		e := bufferErr.String()
		log.Entry().Error(e)
		if strings.Contains(e, "Unauthorized") {
			log.Entry().Error("Either your Github Token is invalid or you use both Vault and Jenkins credentials where your Vault credentials are invalid, to use your Jenkins credentials try setting 'skipVault:true'")
		}
		return "", err
	}

	url := strings.TrimSpace(bufferOut.String())
	return url, nil
}

func executeAnalysis(format, reportName string, customFlags map[string]string, config *codeqlExecuteScanOptions, utils codeqlExecuteScanUtils) ([]piperutils.Path, error) {
	moduleTargetPath := filepath.Join(config.ModulePath, "target")
	report := filepath.Join(moduleTargetPath, reportName)
	cmd, err := prepareCmdForDatabaseAnalyze(customFlags, config, format, report)
	if err != nil {
		log.Entry().Errorf("failed to prepare command for codeql database analyze (format=%s)", format)
		return nil, err
	}
	if err = execute(utils, cmd, GeneralConfig.Verbose); err != nil {
		log.Entry().Errorf("failed running command codeql database analyze for %s generation", format)
		return nil, err
	}
	return []piperutils.Path{
		{Target: report},
	}, nil
}

func prepareCmdForDatabaseCreate(customFlags map[string]string, config *codeqlExecuteScanOptions, utils codeqlExecuteScanUtils) ([]string, error) {
	cmd := []string{"database", "create", config.Database}
	cmd = codeql.AppendFlagIfNotSetByUser(cmd, []string{"--overwrite", "--no-overwrite"}, []string{"--overwrite"}, customFlags)
	cmd = codeql.AppendFlagIfNotSetByUser(cmd, []string{"--source-root", "-s"}, []string{"--source-root", "."}, customFlags)
	cmd = codeql.AppendFlagIfNotSetByUser(cmd, []string{"--working-dir"}, []string{"--working-dir", config.ModulePath}, customFlags)

	if !codeql.IsFlagSetByUser(customFlags, []string{"--language", "-l"}) {
		language := getLangFromBuildTool(config.BuildTool)
		if len(language) == 0 && len(config.Language) == 0 {
			if config.BuildTool == "custom" {
				return nil, fmt.Errorf("as the buildTool is custom. please specify the language parameter")
			} else {
				return nil, fmt.Errorf("the step could not recognize the specified buildTool %s. please specify valid buildtool", config.BuildTool)
			}
		}
		if len(language) > 0 {
			cmd = append(cmd, "--language="+language)
		} else {
			cmd = append(cmd, "--language="+config.Language)
		}
	}

	cmd = codeql.AppendThreadsAndRam(cmd, config.Threads, config.Ram, customFlags)

	if len(config.BuildCommand) > 0 && !codeql.IsFlagSetByUser(customFlags, []string{"--command", "-c"}) {
		buildCmd := config.BuildCommand
		buildCmd = buildCmd + getMavenSettings(buildCmd, config, utils)
		cmd = append(cmd, "--command="+buildCmd)
	}

	if codeql.IsFlagSetByUser(customFlags, []string{"--command", "-c"}) {
		updateCmdFlag(config, customFlags, utils)
	}
	cmd = codeql.AppendCustomFlags(cmd, customFlags)

	return cmd, nil
}

func prepareCmdForDatabaseAnalyze(customFlags map[string]string, config *codeqlExecuteScanOptions, format, reportName string) ([]string, error) {
	cmd := []string{"database", "analyze", "--format=" + format, "--output=" + reportName, config.Database}
	cmd = codeql.AppendThreadsAndRam(cmd, config.Threads, config.Ram, customFlags)
	cmd = codeql.AppendCustomFlags(cmd, customFlags)
	cmd = appendCodeqlQuery(cmd, config.QuerySuite)
	return cmd, nil
}

func prepareCmdForUploadResults(config *codeqlExecuteScanOptions, repoInfo *codeql.RepoInfo, token string) []string {
	cmd := []string{"github", "upload-results", "--sarif=" + filepath.Join(config.ModulePath, "target", "codeqlReport.sarif")}

	//if no git params are passed(commitId, reference, serverUrl, repository), then codeql tries to auto populate it based on git information of the checkout repository.
	//It also depends on the orchestrator. Some orchestrator keep git information and some not.

	if token != "" {
		cmd = append(cmd, "-a="+token)
	}

	if repoInfo.CommitId != "" {
		cmd = append(cmd, "--commit="+repoInfo.CommitId)
	}

	if repoInfo.ServerUrl != "" {
		cmd = append(cmd, "--github-url="+repoInfo.ServerUrl)
	}

	if repoInfo.Repo != "" && repoInfo.Owner != "" {
		cmd = append(cmd, "--repository="+(repoInfo.Owner+"/"+repoInfo.Repo))
	}

	if repoInfo.AnalyzedRef != "" {
		cmd = append(cmd, "--ref="+repoInfo.AnalyzedRef)
	}
	return cmd
}

func uploadSarifResults(config *codeqlExecuteScanOptions, token string, repoInfo *codeql.RepoInfo, utils codeqlExecuteScanUtils) error {
	sarifUrl, err := runGithubUploadResults(config, repoInfo, token, utils)
	if err != nil {
		return err
	}

	codeqlSarifUploader := codeql.NewCodeqlSarifUploaderInstance(sarifUrl, token)
	err = codeql.WaitSarifUploaded(config.SarifCheckMaxRetries, config.SarifCheckRetryInterval, &codeqlSarifUploader)
	if err != nil {
		return errors.Wrap(err, "failed to upload sarif")
	}
	return nil
}

func uploadProjectToGitHub(config *codeqlExecuteScanOptions, repoInfo *codeql.RepoInfo) error {
	log.Entry().Infof("DB sources for %s will be uploaded to target GitHub repo: %s", config.Repository, repoInfo.FullUrl)

	hasToken, token := getToken(config)
	if !hasToken {
		return fmt.Errorf("failed running upload db sources to GitHub as githubToken was not specified")
	}
	repoUploader, err := codeql.NewGitUploaderInstance(
		token,
		repoInfo.AnalyzedRef,
		config.Database,
		repoInfo.CommitId,
		config.Repository,
		config.TargetGithubRepoURL,
	)
	if err != nil {
		log.Entry().WithError(err).Error("failed to create github uploader")
		return err
	}
	targetCommitId, err := repoUploader.UploadProjectToGithub()
	if err != nil {
		return errors.Wrap(err, "failed uploading db sources from non-GitHub SCM to GitHub")
	}
	repoInfo.CommitId = targetCommitId
	log.Entry().Info("DB sources were successfully uploaded to target GitHub repo")

	return nil
}

func checkForCompliance(scanResults []codeql.CodeqlFindings, config *codeqlExecuteScanOptions, repoInfo *codeql.RepoInfo) error {
	for _, scanResult := range scanResults {
		if scanResult.ClassificationName == codeql.AuditAll {
			unaudited := scanResult.Total - scanResult.Audited
			if unaudited > config.VulnerabilityThresholdTotal {
				msg := fmt.Sprintf("Your repository %v with ref %v is not compliant. Total unaudited issues are %v which is greater than the VulnerabilityThresholdTotal count %v",
					repoInfo.FullUrl, repoInfo.AnalyzedRef, unaudited, config.VulnerabilityThresholdTotal)
				return errors.Errorf(msg)
			}
		}
	}
	return nil
}

func addDataToInfluxDB(repoInfo *codeql.RepoInfo, querySuite string, scanResults []codeql.CodeqlFindings, influx *codeqlExecuteScanInflux) {
	influx.codeql_data.fields.repositoryURL = repoInfo.FullUrl
	influx.codeql_data.fields.repositoryReferenceURL = repoInfo.FullRef
	influx.codeql_data.fields.codeScanningLink = repoInfo.ScanUrl
	influx.codeql_data.fields.querySuite = querySuite

	for _, sr := range scanResults {
		if sr.ClassificationName == codeql.AuditAll {
			influx.codeql_data.fields.auditAllAudited = sr.Audited
			influx.codeql_data.fields.auditAllTotal = sr.Total
		}
		if sr.ClassificationName == codeql.Optional {
			influx.codeql_data.fields.optionalAudited = sr.Audited
			influx.codeql_data.fields.optionalTotal = sr.Total
		}
	}
}

func getMavenSettings(buildCmd string, config *codeqlExecuteScanOptions, utils codeqlExecuteScanUtils) string {
	params := ""
	if len(buildCmd) > 0 && config.BuildTool == "maven" && !strings.Contains(buildCmd, "--global-settings") && !strings.Contains(buildCmd, "--settings") {
		mvnParams, err := maven.DownloadAndGetMavenParameters(config.GlobalSettingsFile, config.ProjectSettingsFile, utils)
		if err != nil {
			log.Entry().Error("failed to download and get maven parameters: ", err)
			return params
		}
		for i := 1; i < len(mvnParams); i += 2 {
			params = fmt.Sprintf("%s %s=%s", params, mvnParams[i-1], mvnParams[i])
		}
	}
	return params
}

func updateCmdFlag(config *codeqlExecuteScanOptions, customFlags map[string]string, utils codeqlExecuteScanUtils) {
	var buildCmd string
	if customFlags["--command"] != "" {
		buildCmd = customFlags["--command"]
	} else {
		buildCmd = customFlags["-c"]
	}
	buildCmd += getMavenSettings(buildCmd, config, utils)
	customFlags["--command"] = buildCmd
	delete(customFlags, "-c")
}
