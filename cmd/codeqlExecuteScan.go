package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

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

	err := runCodeqlExecuteScan(&config, telemetryData, utils)
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
		repoInfo.repo = strings.TrimSuffix(match[4], ".git")
		return nil
	}

	return fmt.Errorf("Invalid repository %s", repoUri)
}

func uploadResults(config *codeqlExecuteScanOptions, utils codeqlExecuteScanUtils) error {
	if config.UploadResults {
		if len(config.GithubToken) == 0 {
			return errors.New("failed running upload-results as github token was not specified")
		}

		if config.CommitID == "NA" {
			return errors.New("failed running upload-results as gitCommitId is not available")
		}

		var repoInfo RepoInfo
		err := getGitRepoInfo(config.Repository, &repoInfo)
		if err != nil {
			log.Entry().Error(err)
		}
		repoInfo.ref = config.AnalyzedRef
		repoInfo.commitId = config.CommitID

		provider, err := orchestrator.NewOrchestratorSpecificConfigProvider()
		if err != nil {
			log.Entry().Error(err)
		} else {
			if repoInfo.ref == "" {
				repoInfo.ref = provider.GetReference()
			}

			if repoInfo.commitId == "" {
				repoInfo.commitId = provider.GetCommit()
			}

			if repoInfo.serverUrl == "" {
				err = getGitRepoInfo(provider.GetRepoURL(), &repoInfo)
				if err != nil {
					log.Entry().Error(err)
				}
			}
		}

		cmd := []string{"github", "upload-results", "--sarif=" + fmt.Sprintf("%vtarget/codeqlReport.sarif", config.ModulePath), "-a=" + config.GithubToken}

		if repoInfo.commitId != "" {
			cmd = append(cmd, "--commit="+repoInfo.commitId)
		}

		if repoInfo.serverUrl != "" {
			cmd = append(cmd, "--github-url="+repoInfo.serverUrl)
		}

		if repoInfo.repo != "" {
			cmd = append(cmd, "--repository="+repoInfo.repo)
		}

		if repoInfo.ref != "" {
			cmd = append(cmd, "--ref="+repoInfo.ref)
		}

		//if no git pramas are passed(commitId, reference, serverUrl, repository), then codeql tries to auto populate it based on git information of the checkout repository.
		//It also depends on the orchestrator. Some orchestrator keep git information and some not.
		err = execute(utils, cmd, GeneralConfig.Verbose)
		if err != nil {
			log.Entry().Error("failed to upload sarif results")
			return err
		}
	}

	return nil
}

func runCodeqlExecuteScan(config *codeqlExecuteScanOptions, telemetryData *telemetry.CustomData, utils codeqlExecuteScanUtils) error {
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
			return fmt.Errorf("as the buildTool is custom. please atleast specify the language parameter")
		} else {
			return fmt.Errorf("the step could not recognize the specified buildTool %s. please specify valid buildtool", config.BuildTool)
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
		return err
	}

	err = os.MkdirAll(fmt.Sprintf("%vtarget", config.ModulePath), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	cmd = nil
	cmd = append(cmd, "database", "analyze", "--format=sarif-latest", fmt.Sprintf("--output=%vtarget/codeqlReport.sarif", config.ModulePath), config.Database)
	cmd = append(cmd, getRamAndThreadsFromConfig(config)...)
	cmd = codeqlQuery(cmd, config.QuerySuite)
	err = execute(utils, cmd, GeneralConfig.Verbose)
	if err != nil {
		log.Entry().Error("failed running command codeql database analyze for sarif generation")
		return err
	}

	reports = append(reports, piperutils.Path{Target: fmt.Sprintf("%vtarget/codeqlReport.sarif", config.ModulePath)})

	cmd = nil
	cmd = append(cmd, "database", "analyze", "--format=csv", fmt.Sprintf("--output=%vtarget/codeqlReport.csv", config.ModulePath), config.Database)
	cmd = append(cmd, getRamAndThreadsFromConfig(config)...)
	cmd = codeqlQuery(cmd, config.QuerySuite)
	err = execute(utils, cmd, GeneralConfig.Verbose)
	if err != nil {
		log.Entry().Error("failed running command codeql database analyze for csv generation")
		return err
	}

	reports = append(reports, piperutils.Path{Target: fmt.Sprintf("%vtarget/codeqlReport.csv", config.ModulePath)})
	err = uploadResults(config, utils)
	if err != nil {
		log.Entry().Error("failed to upload results")
		return err
	}

	// create toolrecord file
	toolRecordFileName, err := createToolRecordCodeql(utils, "./", *config)
	if err != nil {
		// do not fail until the framework is well established
		log.Entry().Warning("TR_CODEQL: Failed to create toolrecord file ...", err)
	} else {
		reports = append(reports, piperutils.Path{Target: toolRecordFileName})
	}

	piperutils.PersistReportsAndLinks("codeqlExecuteScan", "./", utils, reports, nil)

	return nil
}

func createToolRecordCodeql(utils codeqlExecuteScanUtils, workspace string, config codeqlExecuteScanOptions) (string, error) {
	repoURL := strings.TrimSuffix(config.Repository, ".git")
	toolInstance, orgName, repoName, err := parseRepositoryURL(repoURL)
	if err != nil {
		return "", err
	}
	record := toolrecord.New(utils, workspace, "codeql", toolInstance)
	record.DisplayName = fmt.Sprintf("%s %s - %s %s", orgName, repoName, config.AnalyzedRef, config.CommitID)
	record.DisplayURL = fmt.Sprintf("%s/security/code-scanning?query=is:open+ref:%s", repoURL, config.AnalyzedRef)
	// Repository
	err = record.AddKeyData("repository",
		fmt.Sprintf("%s/%s", orgName, repoName),
		fmt.Sprintf("%s %s", orgName, repoName),
		config.Repository)
	if err != nil {
		return "", err
	}
	// Repository Reference
	repoReference, err := buildRepoReference(repoURL, config.AnalyzedRef)
	if err != nil {
		log.Entry().WithError(err).Warn("Failed to build repository reference")
	}
	err = record.AddKeyData("repositoryReference",
		config.AnalyzedRef,
		fmt.Sprintf("%s - %s", repoName, config.AnalyzedRef),
		repoReference)
	if err != nil {
		return "", err
	}
	// Scan Results
	err = record.AddKeyData("scanResult",
		fmt.Sprintf("%s/%s", config.AnalyzedRef, config.CommitID),
		fmt.Sprintf("%s %s - %s %s", orgName, repoName, config.AnalyzedRef, config.CommitID),
		fmt.Sprintf("%s/security/code-scanning?query=is:open+ref:%s", repoURL, config.AnalyzedRef))
	if err != nil {
		return "", err
	}
	err = record.Persist()
	if err != nil {
		return "", err
	}
	return record.GetFileName(), nil
}

func parseRepositoryURL(repository string) (toolInstance, orgName, repoName string, err error) {
	if repository == "" {
		err = errors.New("Repository param is not set")
		return
	}
	fullRepo := strings.TrimSuffix(repository, ".git")
	// regexp for toolInstance
	re := regexp.MustCompile(`^[a-zA-Z0-9]+://[a-zA-Z0-9-_.]+/`)
	matchedHost := re.FindAllString(fullRepo, -1)
	if len(matchedHost) == 0 {
		err = errors.New("Unable to parse tool instance from repository url")
		return
	}
	orgRepoNames := strings.Split(strings.TrimPrefix(fullRepo, matchedHost[0]), "/")
	if len(orgRepoNames) < 2 {
		err = errors.New("Unable to parse organization and repo names from repository url")
		return
	}

	toolInstance = strings.Trim(matchedHost[0], "/")
	orgName = orgRepoNames[0]
	repoName = orgRepoNames[1]
	return
}

func buildRepoReference(repository, analyzedRef string) (string, error) {
	if repository == "" || analyzedRef == "" {
		return "", errors.New("Repository or analyzedRef param is not set")
	}
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
