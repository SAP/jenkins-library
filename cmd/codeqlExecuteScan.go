package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/SAP/jenkins-library/pkg/codeql"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/toolrecord"
	"github.com/pkg/errors"
)

type codeqlExecuteScanUtils interface {
	command.ExecRunner

	piperutils.FileUtils
}

type RepoInfo struct {
	serverUrl string
	repo      string
	commitId  string
	ref       string
	owner     string
}

type codeqlExecuteScanUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

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

func getGitRepoInfo(repoUri string, repoInfo *RepoInfo) error {
	if repoUri == "" {
		return errors.New("repository param is not set or it cannot be auto populated")
	}

	pat := regexp.MustCompile(`^(https|git):\/\/([\S]+:[\S]+@)?([^\/:]+)[\/:]([^\/:]+\/[\S]+)$`)
	matches := pat.FindAllStringSubmatch(repoUri, -1)
	if len(matches) > 0 {
		match := matches[0]
		repoInfo.serverUrl = "https://" + match[3]
		repoData := strings.Split(strings.TrimSuffix(match[4], ".git"), "/")
		if len(repoData) != 2 {
			return fmt.Errorf("Invalid repository %s", repoUri)
		}

		repoInfo.owner = repoData[0]
		repoInfo.repo = repoData[1]
		return nil
	}

	return fmt.Errorf("Invalid repository %s", repoUri)
}

func initGitInfo(config *codeqlExecuteScanOptions) RepoInfo {
	var repoInfo RepoInfo
	err := getGitRepoInfo(config.Repository, &repoInfo)
	if err != nil {
		log.Entry().Error(err)
	}
	repoInfo.ref = config.AnalyzedRef
	repoInfo.commitId = config.CommitID

	provider, err := orchestrator.NewOrchestratorSpecificConfigProvider()
	if err != nil {
		log.Entry().Warn("No orchestrator found. We assume piper is running locally.")
	} else {
		if repoInfo.ref == "" {
			repoInfo.ref = provider.GetReference()
		}

		if repoInfo.commitId == "" || repoInfo.commitId == "NA" {
			repoInfo.commitId = provider.GetCommit()
		}

		if repoInfo.serverUrl == "" {
			err = getGitRepoInfo(provider.GetRepoURL(), &repoInfo)
			if err != nil {
				log.Entry().Error(err)
			}
		}
	}

	return repoInfo
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

func uploadResults(config *codeqlExecuteScanOptions, repoInfo RepoInfo, token string, utils codeqlExecuteScanUtils) error {
	cmd := []string{"github", "upload-results", "--sarif=" + filepath.Join(config.ModulePath, "target", "codeqlReport.sarif")}

	if config.GithubToken != "" {
		cmd = append(cmd, "-a="+token)
	}

	if repoInfo.commitId != "" {
		cmd = append(cmd, "--commit="+repoInfo.commitId)
	}

	if repoInfo.serverUrl != "" {
		cmd = append(cmd, "--github-url="+repoInfo.serverUrl)
	}

	if repoInfo.repo != "" {
		cmd = append(cmd, "--repository="+(repoInfo.owner+"/"+repoInfo.repo))
	}

	if repoInfo.ref != "" {
		cmd = append(cmd, "--ref="+repoInfo.ref)
	}

	//if no git pramas are passed(commitId, reference, serverUrl, repository), then codeql tries to auto populate it based on git information of the checkout repository.
	//It also depends on the orchestrator. Some orchestrator keep git information and some not.
	err := execute(utils, cmd, GeneralConfig.Verbose)
	if err != nil {
		log.Entry().Error("failed to upload sarif results")
		return err
	}

	return nil
}

func runCodeqlExecuteScan(config *codeqlExecuteScanOptions, telemetryData *telemetry.CustomData, utils codeqlExecuteScanUtils) ([]piperutils.Path, error) {
	codeqlVersion, err := os.ReadFile("/etc/image-version")
	if err != nil {
		log.Entry().Infof("CodeQL image version: unknown")
	} else {
		log.Entry().Infof("CodeQL image version: %s", string(codeqlVersion))
	}

	var reports []piperutils.Path
	cmd := []string{"database", "create", config.Database, "--overwrite", "--source-root", config.ModulePath}

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
		cmd = append(cmd, "--command="+config.BuildCommand)
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

	repoInfo := initGitInfo(config)
	repoUrl := fmt.Sprintf("%s/%s/%s", repoInfo.serverUrl, repoInfo.owner, repoInfo.repo)
	repoReference, err := buildRepoReference(repoUrl, repoInfo.ref)
	repoCodeqlScanUrl := fmt.Sprintf("%s/security/code-scanning?query=is:open+ref:%s", repoUrl, repoInfo.ref)

	if !config.UploadResults {
		log.Entry().Warn("The sarif results will not be uploaded to the repository and compliance report will not be generated as uploadResults is set to false.")
	} else {
		hasToken, token := getToken(config)
		if !hasToken {
			return reports, errors.New("failed running upload-results as githubToken was not specified")
		}

		err = uploadResults(config, repoInfo, token, utils)
		if err != nil {

			return reports, err
		}

		codeqlScanAuditInstance := codeql.NewCodeqlScanAuditInstance(config.GithubAPIURL, repoInfo.owner, repoInfo.repo, token, []string{})
		scanResults, err := codeqlScanAuditInstance.GetVulnerabilities(repoInfo.ref)
		if err != nil {
			return reports, errors.Wrap(err, "failed to get scan results")
		}

		unaudited := (scanResults.Total - scanResults.Audited)
		if unaudited > config.VulnerabilityThresholdTotal {
			msg := fmt.Sprintf("Your repository %v with ref %v is not compliant. Total unaudited issues are %v which is greater than the VulnerabilityThresholdTotal count %v", repoUrl, repoInfo.ref, unaudited, config.VulnerabilityThresholdTotal)
			if config.CheckForCompliance {

				return reports, errors.Errorf(msg)
			}

			log.Entry().Warning(msg)
		}

		codeqlAudit := codeql.CodeqlAudit{ToolName: "codeql", RepositoryUrl: repoUrl, CodeScanningLink: repoCodeqlScanUrl, RepositoryReferenceUrl: repoReference, ScanResults: scanResults}
		paths, err := codeql.WriteJSONReport(codeqlAudit, config.ModulePath)
		if err != nil {
			return reports, errors.Wrap(err, "failed to write json compliance report")
		}

		reports = append(reports, paths...)
	}

	toolRecordFileName, err := createAndPersistToolRecord(utils, repoInfo, repoReference, repoUrl, repoCodeqlScanUrl)
	if err != nil {
		log.Entry().Warning("TR_CODEQL: Failed to create toolrecord file ...", err)
	} else {
		reports = append(reports, piperutils.Path{Target: toolRecordFileName})
	}

	return reports, nil
}

func createAndPersistToolRecord(utils codeqlExecuteScanUtils, repoInfo RepoInfo, repoReference string, repoUrl string, repoCodeqlScanUrl string) (string, error) {
	toolRecord, err := createToolRecordCodeql(utils, repoInfo, repoReference, repoUrl, repoCodeqlScanUrl)
	if err != nil {
		return "", err
	}

	toolRecordFileName, err := persistToolRecord(toolRecord)
	if err != nil {
		return "", err
	}

	return toolRecordFileName, nil
}

func createToolRecordCodeql(utils codeqlExecuteScanUtils, repoInfo RepoInfo, repoUrl string, repoReference string, repoCodeqlScanUrl string) (*toolrecord.Toolrecord, error) {
	record := toolrecord.New(utils, "./", "codeql", repoInfo.serverUrl)

	if repoInfo.serverUrl == "" {
		return record, errors.New("Repository not set")
	}

	if repoInfo.commitId == "" || repoInfo.commitId == "NA" {
		return record, errors.New("CommitId not set")
	}

	if repoInfo.ref == "" {
		return record, errors.New("Analyzed Reference not set")
	}

	record.DisplayName = fmt.Sprintf("%s %s - %s %s", repoInfo.owner, repoInfo.repo, repoInfo.ref, repoInfo.commitId)
	record.DisplayURL = fmt.Sprintf("%s/security/code-scanning?query=is:open+ref:%s", repoUrl, repoInfo.ref)

	err := record.AddKeyData("repository",
		fmt.Sprintf("%s/%s", repoInfo.owner, repoInfo.repo),
		fmt.Sprintf("%s %s", repoInfo.owner, repoInfo.repo),
		repoUrl)
	if err != nil {
		return record, err
	}

	err = record.AddKeyData("repositoryReference",
		repoInfo.ref,
		fmt.Sprintf("%s - %s", repoInfo.repo, repoInfo.ref),
		repoReference)
	if err != nil {
		return record, err
	}

	err = record.AddKeyData("scanResult",
		fmt.Sprintf("%s/%s", repoInfo.ref, repoInfo.commitId),
		fmt.Sprintf("%s %s - %s %s", repoInfo.owner, repoInfo.repo, repoInfo.ref, repoInfo.commitId),
		fmt.Sprintf("%s/security/code-scanning?query=is:open+ref:%s", repoUrl, repoInfo.ref))
	if err != nil {
		return record, err
	}

	return record, nil
}

func buildRepoReference(repository, analyzedRef string) (string, error) {
	ref := strings.Split(analyzedRef, "/")
	if len(ref) < 3 {
		return "", errors.New(fmt.Sprintf("Wrong analyzedRef format: %s", analyzedRef))
	}
	if strings.Contains(analyzedRef, "pull") {
		if len(ref) < 4 {
			return "", errors.New(fmt.Sprintf("Wrong analyzedRef format: %s", analyzedRef))
		}
		return fmt.Sprintf("%s/pull/%s", repository, ref[2]), nil
	}
	return fmt.Sprintf("%s/tree/%s", repository, ref[2]), nil
}

func persistToolRecord(toolRecord *toolrecord.Toolrecord) (string, error) {
	err := toolRecord.Persist()
	if err != nil {
		return "", err
	}
	return toolRecord.GetFileName(), nil
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
