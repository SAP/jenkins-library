package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"errors"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

const piperDbName string = "piper_step_db.db"
const piperReportName string = "findings.csv"

type credentialdiggerUtils interface {
	command.ExecRunner
	piperutils.FileUtils
}

type credentialdiggerUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func newCDUtils() credentialdiggerUtils {
	utils := credentialdiggerUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func credentialdiggerScan(config credentialdiggerScanOptions, telemetryData *telemetry.CustomData) error {
	utils := newCDUtils()
	// 0: Get attributes from orchestrator
	provider, prov_err := orchestrator.GetOrchestratorConfigProvider(nil)
	if prov_err != nil {
		log.Entry().WithError(prov_err).Error(
			"Unable to load orchestrator specific configuration.")
	}
	if config.Repository == "" {
		// Get current repository from orchestrator
		log.Entry().Debug("Repository URL not defined in step configuration. Try get it from orchestrators")
		repoUrlOrchestrator := provider.RepoURL()
		if repoUrlOrchestrator == "n/a" {
			// Jenkins configuration error
			configError := errors.New(fmt.Sprintf("Unknown repository URL %s", repoUrlOrchestrator))
			log.Entry().WithError(configError).Error(
				"Repository URL n/a. Please verify git plugin is installed.")
			// The repository to scan was not identified. Return an error
			return configError
		}
		config.Repository = repoUrlOrchestrator
		log.Entry().Debug("Use current repository: ", repoUrlOrchestrator)
	}
	if provider.IsPullRequest() {
		// set the pr number
		config.PrNumber, _ = strconv.Atoi(provider.PullRequestConfig().Key)
		log.Entry().Debug("Scan the current pull request: number ", config.PrNumber)
	}

	// 1: Add rules
	log.Entry().Info("Load rules")
	err := credentialdiggerAddRules(&config, telemetryData, utils)
	if err != nil {
		log.Entry().Error("Failed running credentialdigger add_rules")
		return err
	}
	log.Entry().Info("Rules added")

	// 2: Scan the repository
	// Choose between scan-pr, scan-snapshot, and full-scan (with this priority
	// order)
	switch {
	case config.PrNumber != 0: // int type is not nillable in golang
		log.Entry().Debug("Scan PR")
		// if a PrNumber is declared, run scan_pr
		err = credentialdiggerScanPR(&config, telemetryData, utils) // scan PR with CD
	case config.Snapshot != "":
		log.Entry().Debug("Scan snapshot")
		// if a Snapshot is declared, run scan_snapshot
		err = credentialdiggerScanSnapshot(&config, telemetryData, utils) // scan Snapshot with CD
	default:
		// The default case is the normal full scan
		log.Entry().Debug("Full scan repo")
		err = credentialdiggerFullScan(&config, telemetryData, utils) // full scan with CD
	}
	// err is an error exit number when there are findings
	if err == nil {
		log.Entry().Info("No leaks found in this repo with scan")
		// Even if there are no leaks, the user may still want to export all
		// the discoveries (param exportAll set to true)
	}

	// 3: Get discoveries
	err = credentialdiggerGetDiscoveries(&config, telemetryData, utils)
	if err != nil {
		// The exit number is the number of discoveries exported
		// Therefore, this error is not relevant, if raised
		log.Entry().Warn("There are findings to review")
	} else {
		// There are no discoveries exported, so no need to generate the
		// artifact
		return nil
	}

	// 4: Export report in workspace
	reports := []piperutils.Path{}
	reports = append(reports, piperutils.Path{Target: fmt.Sprintf("%v", piperReportName)})
	piperutils.PersistReportsAndLinks("credentialdiggerScan", "./", utils, reports, nil)

	return nil
}

func executeCredentialDiggerProcess(utils credentialdiggerUtils, args []string) error {
	return utils.RunExecutable("credentialdigger", args...)
}

// hasConfigurationFile checks if the given file exists
func hasRulesFile(file string, utils credentialdiggerUtils) bool {
	exists, err := utils.FileExists(file)
	if err != nil {
		log.Entry().WithError(err).Error()
	}
	return exists
}

func credentialdiggerAddRules(config *credentialdiggerScanOptions, telemetryData *telemetry.CustomData, service credentialdiggerUtils) error {
	// Credentialdigger home can be changed with local forks (e.g., for local piper runs)
	cdHome := "/credential-digger-ui" // cdHome path as in docker container
	if cdh := os.Getenv("CREDENTIALDIGGER_HOME"); cdh != "" {
		cdHome = cdh
	}
	log.Entry().Debug("Use credentialdigger home ", cdHome)
	// Set the rule file to the standard ruleset shipped within credential
	// digger container
	ruleFile := filepath.Join(cdHome, "backend", "rules.yml")

	if config.RulesDownloadURL != "" {
		// Download custom rule file from this URL
		log.Entry().Debugf("Download custom ruleset from %v", config.RulesDownloadURL)
		dlClient := piperhttp.Client{}
		ruleFile := filepath.Join(cdHome, "backend", "custom-rules.yml")
		dlClient.DownloadFile(config.RulesDownloadURL, ruleFile, nil, nil)
		log.Entry().Info("Download and use remote rules")
	} else {
		log.Entry().Debug("Use a local ruleset")
		// Use rules defined in stashed file
		if hasRulesFile(config.RulesFile, service) {
			log.Entry().WithField("file", config.RulesFile).Info(
				"Use stashed rules file from repository")
			ruleFile = config.RulesFile
		} else {
			log.Entry().Info("Use standard pre-defined rules")
		}

	}
	cmd_list := []string{"add_rules", "--sqlite", piperDbName, ruleFile}
	return executeCredentialDiggerProcess(service, cmd_list)
}

func credentialdiggerGetDiscoveries(config *credentialdiggerScanOptions, telemetryData *telemetry.CustomData, service credentialdiggerUtils) error {
	log.Entry().Info("Get discoveries")
	cmd_list := []string{"get_discoveries", config.Repository, "--sqlite", piperDbName,
		"--save", piperReportName}
	// Export all the discoveries or export only new ones
	if !config.ExportAll {
		cmd_list = append(cmd_list, "--state", "new")
	} else {
		log.Entry().Info("Export all discoveries")
	}
	err := executeCredentialDiggerProcess(service, cmd_list)
	if err != nil {
		log.Entry().Warn("Report generated")
		return err
	}
	log.Entry().Info("Scan complete with no potential leaks")
	return nil
}

func credentialdiggerBuildCommonArgs(config *credentialdiggerScanOptions) []string {
	/*Some arguments are the same for all the scan flavors. Build them here
	* not to duplicate code.*/
	scan_args := []string{}
	// Repository url and sqlite db (always mandatory)
	scan_args = append(scan_args, config.Repository, "--sqlite", piperDbName)
	//git token is not mandatory for base credential digger tool, but in
	//piper it is
	scan_args = append(scan_args, "--git_token", config.Token)
	//debug
	if config.Debug {
		log.Entry().Debug("Run the scan in debug mode")
		scan_args = append(scan_args, "--debug")
	}
	//models
	if len(config.Models) > 0 {
		log.Entry().Debugf("Enable models %v", config.Models)
		scan_args = append(scan_args, "--models")
		scan_args = append(scan_args, config.Models...)
	}

	return scan_args
}

func credentialdiggerScanSnapshot(config *credentialdiggerScanOptions, telemetryData *telemetry.CustomData, service credentialdiggerUtils) error {
	log.Entry().Infof(
		"Scan Snapshot %v from repo %v", config.Snapshot, config.Repository)
	cmd_list := []string{"scan_snapshot",
		"--snapshot", config.Snapshot}
	cmd_list = append(cmd_list, credentialdiggerBuildCommonArgs(config)...)
	leaks := executeCredentialDiggerProcess(service, cmd_list)
	if leaks != nil {
		log.Entry().Warn("The scan found potential leaks in this Snapshot")
		return leaks
	} else {
		log.Entry().Info("No leaks found")
		return nil
	}
}

func credentialdiggerScanPR(config *credentialdiggerScanOptions, telemetryData *telemetry.CustomData, service credentialdiggerUtils) error {
	log.Entry().Infof(
		"Scan PR %v from repo %v", config.PrNumber, config.Repository)
	cmd_list := []string{"scan_pr",
		"--pr", strconv.Itoa(config.PrNumber),
		"--api_endpoint", config.APIURL}
	cmd_list = append(cmd_list, credentialdiggerBuildCommonArgs(config)...)
	leaks := executeCredentialDiggerProcess(service, cmd_list)
	if leaks != nil {
		log.Entry().Warn("The scan found potential leaks in this PR")
		return leaks
	} else {
		log.Entry().Info("No leaks found")
		return nil
	}
}

func credentialdiggerFullScan(config *credentialdiggerScanOptions, telemetryData *telemetry.CustomData, service credentialdiggerUtils) error {
	log.Entry().Infof("Full scan of repository %v", config.Repository)
	cmd_list := []string{"scan"}
	cmd_list = append(cmd_list, credentialdiggerBuildCommonArgs(config)...)
	leaks := executeCredentialDiggerProcess(service, cmd_list)
	if leaks != nil {
		log.Entry().Warn("The scan found potential leaks")
		log.Entry().Warnf("%v potential leaks found", leaks)
		return leaks
	} else {
		log.Entry().Info("No leaks found")
		return nil
	}
}
