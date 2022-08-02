package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

type codeqlExecuteScanUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)
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

	pat := regexp.MustCompile(`^(https|git)(:\/\/|@)([^\/:]+)[\/:]([^\/:]+\/[^.]+)(.git)*$`)
	matches := pat.FindAllStringSubmatch(repoUri, -1)
	if len(matches) > 0 {
		match := matches[0]
		repoInfo.serverUrl = "https://" + match[3]
		repoInfo.repo = match[4]
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

	cmd = append(cmd, "--language="+language)
	if len(config.Language) > 0 {
		cmd = append(cmd, "--language="+config.Language)
	}

	//codeql has an autobuilder which tries to build the project based on specified programming language
	if len(config.BuildCommand) > 0 {
		cmd = append(cmd, "--command="+config.BuildCommand)
	}

	err := execute(utils, cmd, GeneralConfig.Verbose)
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
	cmd = codeqlQuery(cmd, config.QuerySuite)
	err = execute(utils, cmd, GeneralConfig.Verbose)
	if err != nil {
		log.Entry().Error("failed running command codeql database analyze for sarif generation")
		return err
	}

	reports = append(reports, piperutils.Path{Target: fmt.Sprintf("%vtarget/codeqlReport.sarif", config.ModulePath)})

	cmd = nil
	cmd = append(cmd, "database", "analyze", "--format=csv", fmt.Sprintf("--output=%vtarget/codeqlReport.csv", config.ModulePath), config.Database)
	cmd = codeqlQuery(cmd, config.QuerySuite)
	err = execute(utils, cmd, GeneralConfig.Verbose)
	if err != nil {
		log.Entry().Error("failed running command codeql database analyze for csv generation")
		return err
	}

	reports = append(reports, piperutils.Path{Target: fmt.Sprintf("%vtarget/codeqlReport.csv", config.ModulePath)})

	piperutils.PersistReportsAndLinks("codeqlExecuteScan", "./", reports, nil)

	err = uploadResults(config, utils)
	if err != nil {
		log.Entry().Error("failed to upload results")
		return err
	}

	return nil
}
