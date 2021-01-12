package whitesource

import (
	"fmt"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

const jvmTarGz = "jvm.tar.gz"
const jvmDir = "./jvm"

// ExecuteUAScan executes a scan with the Whitesource Unified Agent.
func (s *Scan) ExecuteUAScan(config *ScanOptions, utils Utils) error {
	// Download the unified agent jar file if one does not exist
	err := downloadAgent(config, utils)
	if err != nil {
		return err
	}

	javaPath, err := downloadJre(config, utils)
	if err != nil {
		return err
	}

	// ToDo: Download Docker/container image if required

	// ToDo: check if this is required
	if err := s.AppendScannedProject(s.AggregateProjectName); err != nil {
		return err
	}

	configPath, err := config.RewriteUAConfigurationFile(utils)
	if err != nil {
		return err
	}

	// ToDo: remove parameters which are added to UA config via RewriteUAConfigurationFile()
	// let the scanner resolve project name on its own?
	// Check: WhiteSource url needs to be passed! current service URL is API endpoint and not /agent endpoint!
	err = utils.RunExecutable(javaPath, "-jar", config.AgentFileName, "-d", ".", "-c", configPath,
		"-apiKey", config.OrgToken, "-userKey", config.UserToken, "-project", s.AggregateProjectName,
		"-product", config.ProductName, "-productVersion", s.ProductVersion, "-wss.url", config.AgentURL)

	if err := removeJre(javaPath, utils); err != nil {
		log.Entry().Warning(err)
	}

	if err != nil {
		exitCode := utils.GetExitCode()
		log.Entry().Infof("WhiteSource scan failed with exit code %v", exitCode)
		evaluateExitCode(exitCode)
		return errors.Wrapf(err, "failed to execute WhiteSource scan with exit code %v", exitCode)
	}
	return nil
}

func evaluateExitCode(exitCode int) {
	switch exitCode {
	case 255:
		log.Entry().Info("General error has occurred.")
		log.SetErrorCategory(log.ErrorUndefined)
	case 254:
		log.Entry().Info("Whitesource found one or multiple policy violations.")
		log.SetErrorCategory(log.ErrorCompliance)
	case 253:
		log.Entry().Info("The local scan client failed to execute the scan.")
		log.SetErrorCategory(log.ErrorUndefined)
	case 252:
		log.Entry().Info("There was a failure in the connection to the WhiteSource servers.")
		log.SetErrorCategory(log.ErrorInfrastructure)
	case 251:
		log.Entry().Info("The server failed to analyze the scan.")
		log.SetErrorCategory(log.ErrorService)
	case 250:
		log.Entry().Info("One of the package manager's prerequisite steps (e.g. npm install) failed.")
		log.SetErrorCategory(log.ErrorCustom)
	default:
		log.Entry().Info("Whitesource scan failed with unknown error code")
		log.SetErrorCategory(log.ErrorUndefined)
	}
}

// downloadAgent downloads the unified agent jar file if one does not exist
func downloadAgent(config *ScanOptions, utils Utils) error {
	agentFile := config.AgentFileName
	exists, err := utils.FileExists(agentFile)
	if err != nil {
		return errors.Wrapf(err, "failed to check if file '%s' exists", agentFile)
	}
	if !exists {
		err := utils.DownloadFile(config.AgentDownloadURL, agentFile, nil, nil)
		if err != nil {
			return errors.Wrapf(err, "failed to download unified agent from URL '%s' to file '%s'", config.AgentDownloadURL, agentFile)
		}
	}
	return nil
}

// downloadJre downloads the a JRE in case no java command can be executed
func downloadJre(config *ScanOptions, utils Utils) (string, error) {
	err := utils.RunExecutable("java", "--version")
	javaPath := "java"
	if err != nil {
		log.Entry().Infof("No Java installation found, downloading JVM from %v", config.JreDownloadURL)
		err := utils.DownloadFile(config.JreDownloadURL, jvmTarGz, nil, nil)
		if err != nil {
			return "", errors.Wrapf(err, "failed to download jre from URL '%s'", config.JreDownloadURL)
		}

		// ToDo: replace tar call with go library call
		err = utils.MkdirAll(jvmDir, 0755)

		err = utils.RunExecutable("tar", fmt.Sprintf("--directory=%v", jvmDir), "--strip-components=1", "-xzf", jvmTarGz)
		if err != nil {
			return "", errors.Wrapf(err, "failed to extract %v", jvmTarGz)
		}
		log.Entry().Info("Java successfully installed")
		javaPath = filepath.Join(jvmDir, "bin", "java")
	}
	return javaPath, nil
}

func removeJre(javaPath string, utils Utils) error {
	if javaPath == "java" {
		return nil
	}
	if err := utils.RemoveAll(jvmDir); err != nil {
		return fmt.Errorf("failed to remove downloaded and extracted jvm from %v", jvmDir)
	}
	log.Entry().Debugf("Java successfully removed from %v", jvmDir)
	if err := utils.FileRemove(jvmTarGz); err != nil {
		return fmt.Errorf("failed to remove downloaded %v", jvmTarGz)
	}
	log.Entry().Debugf("%v successfully removed", jvmTarGz)
	return nil
}
