package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/codeql"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

type codeqlExecuteScanUtils interface {
	command.ExecRunner

	piperutils.FileUtils
}

type codeqlExecuteScanUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

const sarifUploadComplete = "complete"
const sarifUploadFailed = "failed"

func newCodeqlExecuteScanUtils() codeqlExecuteScanUtils {
	utils := codeqlExecuteScanUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}

	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func codeqlExecuteScan(config codeqlExecuteScanOptions, telemetryData *telemetry.CustomData) {

	utils := newCodeqlExecuteScanUtils()

	reports, err := runCodeqlExecuteScan(&config, telemetryData, utils)
	piperutils.PersistReportsAndLinks("codeqlExecuteScan", "./", utils, reports, nil)

	if err != nil {
		log.Entry().WithError(err).Fatal("Codeql scan failed")
	}
}

func codeqlQuery(cmd []string, codeqlQuery string) []string {
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

func getGitRepoInfo(repoUri string, repoInfo *codeql.RepoInfo) error {
	if repoUri == "" {
		return errors.New("repository param is not set or it cannot be auto populated")
	}

	pat := regexp.MustCompile(`^(https:\/\/|git@)([\S]+:[\S]+@)?([^\/:]+)[\/:]([^\/:]+\/[\S]+)$`)
	matches := pat.FindAllStringSubmatch(repoUri, -1)
	if len(matches) > 0 {
		match := matches[0]
		repoInfo.ServerUrl = "https://" + match[3]
		repoData := strings.Split(strings.TrimSuffix(match[4], ".git"), "/")
		if len(repoData) != 2 {
			return fmt.Errorf("Invalid repository %s", repoUri)
		}

		repoInfo.Owner = repoData[0]
		repoInfo.Repo = repoData[1]
		return nil
	}

	return fmt.Errorf("Invalid repository %s", repoUri)
}

func initGitInfo(config *codeqlExecuteScanOptions) (codeql.RepoInfo, error) {
	var repoInfo codeql.RepoInfo
	err := getGitRepoInfo(config.Repository, &repoInfo)
	if err != nil {
		log.Entry().Error(err)
	}

	repoInfo.Ref = config.AnalyzedRef
	repoInfo.CommitId = config.CommitID

	provider, err := orchestrator.NewOrchestratorSpecificConfigProvider()
	if err != nil {
		log.Entry().Warn("No orchestrator found. We assume piper is running locally.")
	} else {
		if repoInfo.Ref == "" {
			repoInfo.Ref = provider.GetReference()
		}

		if repoInfo.CommitId == "" || repoInfo.CommitId == "NA" {
			repoInfo.CommitId = provider.GetCommit()
		}

		if repoInfo.ServerUrl == "" {
			err = getGitRepoInfo(provider.GetRepoURL(), &repoInfo)
			if err != nil {
				log.Entry().Error(err)
			}
		}
	}
	if len(config.TargetGithubRepoURL) > 0 {
		if strings.Contains(repoInfo.ServerUrl, "github") {
			log.Entry().Errorf("TargetGithubRepoURL should not be set as the source repo is on github.")
			return repoInfo, errors.New("TargetGithubRepoURL should not be set as the source repo is on github.")
		}
		err := getGitRepoInfo(config.TargetGithubRepoURL, &repoInfo)
		if err != nil {
			log.Entry().Error(err)
			return repoInfo, err
		}
		if len(config.TargetGithubBranchName) > 0 {
			repoInfo.Ref = config.TargetGithubBranchName
			if len(strings.Split(config.TargetGithubBranchName, "/")) < 3 {
				repoInfo.Ref = "refs/heads/" + config.TargetGithubBranchName
			}
		}
	}

	return repoInfo, nil
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

func uploadResults(config *codeqlExecuteScanOptions, repoInfo codeql.RepoInfo, token string, utils codeqlExecuteScanUtils) (string, error) {
	cmd := []string{"github", "upload-results", "--sarif=" + filepath.Join(config.ModulePath, "target", "codeqlReport.sarif")}

	if config.GithubToken != "" {
		cmd = append(cmd, "-a="+token)
	}

	if repoInfo.CommitId != "" {
		cmd = append(cmd, "--commit="+repoInfo.CommitId)
	}

	if repoInfo.ServerUrl != "" {
		cmd = append(cmd, "--github-url="+repoInfo.ServerUrl)
	}

	if repoInfo.Repo != "" {
		cmd = append(cmd, "--repository="+(repoInfo.Owner+"/"+repoInfo.Repo))
	}

	if repoInfo.Ref != "" {
		cmd = append(cmd, "--ref="+repoInfo.Ref)
	}

	//if no git params are passed(commitId, reference, serverUrl, repository), then codeql tries to auto populate it based on git information of the checkout repository.
	//It also depends on the orchestrator. Some orchestrator keep git information and some not.

	var bufferOut, bufferErr bytes.Buffer
	utils.Stdout(&bufferOut)
	defer utils.Stdout(log.Writer())
	utils.Stderr(&bufferErr)
	defer utils.Stderr(log.Writer())

	err := execute(utils, cmd, GeneralConfig.Verbose)
	if err != nil {
		e := bufferErr.String()
		log.Entry().Error(e)
		if strings.Contains(e, "Unauthorized") {
			log.Entry().Error("Either your Github Token is invalid or you use both Vault and Jenkins credentials where your Vault credentials are invalid, to use your Jenkins credentials try setting 'skipVault:true'")
		}
		log.Entry().Error("failed to upload sarif results")
		return "", err
	}

	url := bufferOut.String()
	return strings.TrimSpace(url), nil
}

func waitSarifUploaded(config *codeqlExecuteScanOptions, codeqlSarifUploader codeql.CodeqlSarifUploader) error {
	maxRetries := config.SarifCheckMaxRetries
	retryInterval := time.Duration(config.SarifCheckRetryInterval) * time.Second

	log.Entry().Info("waiting for the SARIF to upload")
	i := 1
	for {
		sarifStatus, err := codeqlSarifUploader.GetSarifStatus()
		if err != nil {
			return err
		}
		log.Entry().Infof("the SARIF processing status: %s", sarifStatus.ProcessingStatus)
		if sarifStatus.ProcessingStatus == sarifUploadComplete {
			return nil
		}
		if sarifStatus.ProcessingStatus == sarifUploadFailed {
			for e := range sarifStatus.Errors {
				log.Entry().Error(e)
			}
			return errors.New("failed to upload sarif file")
		}
		if i <= maxRetries {
			log.Entry().Infof("still waiting for the SARIF to upload: retrying in %d seconds... (retry %d/%d)", config.SarifCheckRetryInterval, i, maxRetries)
			time.Sleep(retryInterval)
			i++
			continue
		}
		return errors.New("failed to check sarif uploading status: max retries reached")
	}
}

func runCodeqlExecuteScan(config *codeqlExecuteScanOptions, telemetryData *telemetry.CustomData, utils codeqlExecuteScanUtils) ([]piperutils.Path, error) {
	codeqlVersion, err := os.ReadFile("/etc/image-version")
	if err != nil {
		log.Entry().Infof("CodeQL image version: unknown")
	} else {
		log.Entry().Infof("CodeQL image version: %s", string(codeqlVersion))
	}

	var reports []piperutils.Path
	cmd := []string{"database", "create", config.Database, "--overwrite", "--source-root", ".", "--working-dir", config.ModulePath}

	language := getLangFromBuildTool(config.BuildTool)

	if len(language) == 0 && len(config.Language) == 0 {
		if config.BuildTool == "custom" {
			return reports, fmt.Errorf("as the buildTool is custom. please specify the language parameter")
		} else {
			return reports, fmt.Errorf("the step could not recognize the specified buildTool %s. please specify valid buildtool", config.BuildTool)
		}
	}
	if len(language) > 0 {
		cmd = append(cmd, "--language="+language)
	} else {
		cmd = append(cmd, "--language="+config.Language)
	}

	cmd = append(cmd, getRamAndThreadsFromConfig(config)...)

	//codeql has an autobuilder which tries to build the project based on specified programming language
	if len(config.BuildCommand) > 0 {
		buildCmd := config.BuildCommand
		if len(config.ProjectSettingsFile) > 0 && config.BuildTool == "maven" {
			buildCmd = fmt.Sprintf("%s --settings=%s", buildCmd, config.ProjectSettingsFile)
		}
		if len(config.GlobalSettingsFile) > 0 && config.BuildTool == "maven" {
			buildCmd = fmt.Sprintf("%s --global-settings=%s", buildCmd, config.GlobalSettingsFile)
		}
		cmd = append(cmd, "--command="+buildCmd)
	}

	err = execute(utils, cmd, GeneralConfig.Verbose)
	if err != nil {
		log.Entry().Error("failed running command codeql database create")
		return reports, err
	}

	err = os.MkdirAll(filepath.Join(config.ModulePath, "target"), os.ModePerm)
	if err != nil {
		return reports, fmt.Errorf("failed to create directory: %w", err)
	}

	cmd = nil
	cmd = append(cmd, "database", "analyze", "--format=sarif-latest", fmt.Sprintf("--output=%v", filepath.Join(config.ModulePath, "target", "codeqlReport.sarif")), config.Database)
	cmd = append(cmd, getRamAndThreadsFromConfig(config)...)
	cmd = codeqlQuery(cmd, config.QuerySuite)
	err = execute(utils, cmd, GeneralConfig.Verbose)
	if err != nil {
		log.Entry().Error("failed running command codeql database analyze for sarif generation")
		return reports, err
	}

	reports = append(reports, piperutils.Path{Target: filepath.Join(config.ModulePath, "target", "codeqlReport.sarif")})

	cmd = nil
	cmd = append(cmd, "database", "analyze", "--format=csv", fmt.Sprintf("--output=%v", filepath.Join(config.ModulePath, "target", "codeqlReport.csv")), config.Database)
	cmd = append(cmd, getRamAndThreadsFromConfig(config)...)
	cmd = codeqlQuery(cmd, config.QuerySuite)
	err = execute(utils, cmd, GeneralConfig.Verbose)
	if err != nil {
		log.Entry().Error("failed running command codeql database analyze for csv generation")
		return reports, err
	}

	reports = append(reports, piperutils.Path{Target: filepath.Join(config.ModulePath, "target", "codeqlReport.csv")})

	repoInfo, err := initGitInfo(config)
	if err != nil {
		return reports, err
	}
	repoUrl := fmt.Sprintf("%s/%s/%s", repoInfo.ServerUrl, repoInfo.Owner, repoInfo.Repo)
	repoReference, err := codeql.BuildRepoReference(repoUrl, repoInfo.Ref)
	repoCodeqlScanUrl := fmt.Sprintf("%s/security/code-scanning?query=is:open+ref:%s", repoUrl, repoInfo.Ref)

	if len(config.TargetGithubRepoURL) > 0 {
		hasToken, token := getToken(config)
		if !hasToken {
			return reports, errors.New("failed running upload db sources to GitHub as githubToken was not specified")
		}
		repoUploader, err := codeql.NewGitUploaderInstance(
			token,
			repoInfo.Ref,
			config.Database,
			repoInfo.CommitId,
			config.Repository,
			config.TargetGithubRepoURL,
		)
		if err != nil {
			return reports, err
		}
		targetCommitId, err := repoUploader.UploadProjectToGithub()
		if err != nil {
			return reports, errors.Wrap(err, "failed uploading db sources from non-GitHub SCM to GitHub")
		}
		repoInfo.CommitId = targetCommitId
	}

	if !config.UploadResults {
		log.Entry().Warn("The sarif results will not be uploaded to the repository and compliance report will not be generated as uploadResults is set to false.")
	} else {
		hasToken, token := getToken(config)
		if !hasToken {
			return reports, errors.New("failed running upload-results as githubToken was not specified")
		}

		sarifUrl, err := uploadResults(config, repoInfo, token, utils)
		if err != nil {
			return reports, err
		}
		codeqlSarifUploader := codeql.NewCodeqlSarifUploaderInstance(sarifUrl, token)
		err = waitSarifUploaded(config, &codeqlSarifUploader)
		if err != nil {
			return reports, errors.Wrap(err, "failed to upload sarif")
		}

		codeqlScanAuditInstance := codeql.NewCodeqlScanAuditInstance(repoInfo.ServerUrl, repoInfo.Owner, repoInfo.Repo, token, []string{})
		scanResults, err := codeqlScanAuditInstance.GetVulnerabilities(repoInfo.Ref)
		if err != nil {
			return reports, errors.Wrap(err, "failed to get scan results")
		}

		codeqlAudit := codeql.CodeqlAudit{ToolName: "codeql", RepositoryUrl: repoUrl, CodeScanningLink: repoCodeqlScanUrl, RepositoryReferenceUrl: repoReference, QuerySuite: config.QuerySuite, ScanResults: scanResults}
		paths, err := codeql.WriteJSONReport(codeqlAudit, config.ModulePath)
		if err != nil {
			return reports, errors.Wrap(err, "failed to write json compliance report")
		}
		reports = append(reports, paths...)

		if config.CheckForCompliance {
			for _, scanResult := range scanResults {
				unaudited := scanResult.Total - scanResult.Audited
				if unaudited > config.VulnerabilityThresholdTotal {
					msg := fmt.Sprintf("Your repository %v with ref %v is not compliant. Total unaudited issues are %v which is greater than the VulnerabilityThresholdTotal count %v", repoUrl, repoInfo.Ref, unaudited, config.VulnerabilityThresholdTotal)
					return reports, errors.Errorf(msg)
				}
			}
		}
	}

	toolRecordFileName, err := codeql.CreateAndPersistToolRecord(utils, repoInfo, repoReference, repoUrl, config.ModulePath)
	if err != nil {
		log.Entry().Warning("TR_CODEQL: Failed to create toolrecord file ...", err)
	} else {
		reports = append(reports, piperutils.Path{Target: toolRecordFileName})
	}

	return reports, nil
}

func getRamAndThreadsFromConfig(config *codeqlExecuteScanOptions) []string {
	params := make([]string, 0, 2)
	if len(config.Threads) > 0 {
		params = append(params, "--threads="+config.Threads)
	}
	if len(config.Ram) > 0 {
		params = append(params, "--ram="+config.Ram)
	}
	return params
}
