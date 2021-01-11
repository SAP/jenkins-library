package whitesource

import (
	"fmt"
	"os"
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

	//ToDo: Download Docker/container image if required

	// ToDo: What to do with auto generation? Completely replace with confighelper?
	// Auto generate a config file based on the working directory's contents.
	// TODO/NOTE: Currently this scans the UA jar file as a dependency since it is downloaded beforehand
	//if err := autoGenerateWhitesourceConfig(config, utils); err != nil {
	//	return err
	//}

	// ToDo: check if this is required
	if err := s.AppendScannedProject(s.AggregateProjectName); err != nil {
		return err
	}

	configPath, err := config.RewriteUAConfigurationFile()
	if err != nil {
		return err
	}

	//ToDo: remove parameters which are added to UA config via RewriteUAConfigurationFile()
	err = utils.RunExecutable(javaPath, "-jar", config.AgentFileName, "-d", ".", "-c", configPath,
		"-apiKey", config.OrgToken, "-userKey", config.UserToken, "-project", s.AggregateProjectName,
		"-product", config.ProductName, "-productVersion", s.ProductVersion)

	removeJre(javaPath, utils)

	if err != nil {
		//ToDo: check exit code and categorize error?
		return errors.Wrapf(err, "failed to execute WhiteSource scan")
	}
	return nil
}

// downloadAgent downloads the unified agent jar file if one does not exist
func downloadAgent(config *ScanOptions, utils Utils) error {
	agentFile := config.AgentFileName
	exists, err := utils.FileExists(agentFile)
	if err != nil {
		return fmt.Errorf("could not check whether the file '%s' exists: %w", agentFile, err)
	}
	if !exists {
		err := utils.DownloadFile(config.AgentDownloadURL, agentFile, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to download unified agent from URL '%s' to file '%s': %w",
				config.AgentDownloadURL, agentFile, err)
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
			errors.Wrapf(err, "failed to download unified agent from URL '%s'", config.JreDownloadURL)
		}

		// ToDo: replace tar call with go library call
		// ToDo: extract to subdirectory to make deletion & exclusion easier
		err = utils.MkdirAll(jvmDir, 0755)

		err = utils.RunExecutable("tar", fmt.Sprintf("--directory=%v", jvmDir), "--strip-components=1", "-xzf", jvmTarGz)
		if err != nil {
			errors.Wrapf(err, "failed to extract %v", jvmTarGz)
		}
		log.Entry().Info("Java successfully installed")
		javaPath = filepath.Join(jvmDir, "bin", "java")
	}
	return javaPath, nil
}

func removeJre(javaPath string, utils Utils) {
	if javaPath == "java" {
		return
	}
	if err := utils.RemoveAll(jvmDir); err != nil {
		log.Entry().Warning("Failed to remove downloaded and extracted jvm")
	}
	log.Entry().Infof("Java successfully removed from %v", jvmDir)
	if err := utils.FileRemove(jvmTarGz); err != nil {
		log.Entry().Warningf("Failed to remove downloaded %v", jvmTarGz)
	}
	log.Entry().Infof("%v successfully removed", jvmTarGz)
}

// autoGenerateWhitesourceConfig
// Auto generate a config file based on the current directory structure, renames it to user specified configFilePath
// Generated file name will be 'wss-generated-file.config'
func autoGenerateWhitesourceConfig(config *ScanOptions, utils Utils) error {
	// TODO: Should we rely on -detect, or set the parameters manually?
	if err := utils.RunExecutable("java", "-jar", config.AgentFileName, "-d", ".", "-detect"); err != nil {
		return err
	}

	// Rename generated config file to config.ConfigFilePath parameter
	if err := utils.FileRename("wss-generated-file.config", config.ConfigFilePath); err != nil {
		return err
	}

	// Append aggregateModules=true parameter to config file (consolidates multi-module projects into one)
	f, err := utils.FileOpen(config.ConfigFilePath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	// Append additional config parameters to prevent multiple projects being generated
	m2Path := config.M2Path
	if m2Path == "" {
		m2Path = ".m2"
	}
	cfg := fmt.Sprintf("\ngradle.aggregateModules=true\nmaven.aggregateModules=true\ngradle.localRepositoryPath=.gradle\nmaven.m2RepositoryPath=%s\nexcludes=%s",
		m2Path,
		config.Excludes)
	if _, err = f.WriteString(cfg); err != nil {
		return err
	}

	// archiveExtractionDepth=0
	if err := utils.RunExecutable("sed", "-ir", `s/^[#]*\s*archiveExtractionDepth=.*/archiveExtractionDepth=0/`,
		config.ConfigFilePath); err != nil {
		return err
	}

	// config.Includes defaults to "**/*.java **/*.jar **/*.py **/*.go **/*.js **/*.ts"
	regex := fmt.Sprintf(`s/^[#]*\s*includes=.*/includes="%s"/`, config.Includes)
	if err := utils.RunExecutable("sed", "-ir", regex, config.ConfigFilePath); err != nil {
		return err
	}

	return nil
}
