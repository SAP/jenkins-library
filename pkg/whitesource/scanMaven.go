package whitesource

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"path/filepath"
	"strings"
)

// ExecuteMavenScan constructs maven parameters from the given configuration, and executes the maven goal
// "org.whitesource:whitesource-maven-plugin:19.5.1:update".
func (s *Scan) ExecuteMavenScan(config *ScanOptions, utils Utils) error {
	log.Entry().Infof("Using Whitesource scan for Maven project")
	pomPath := config.PomPath
	if pomPath == "" {
		pomPath = "pom.xml"
	}
	return s.ExecuteMavenScanForPomFile(config, utils, pomPath)
}

// ExecuteMavenScanForPomFile constructs maven parameters from the given configuration, and executes the maven goal
// "org.whitesource:whitesource-maven-plugin:19.5.1:update" for the given pom file.
func (s *Scan) ExecuteMavenScanForPomFile(config *ScanOptions, utils Utils, pomPath string) error {
	pomExists, _ := utils.FileExists(pomPath)
	if !pomExists {
		return fmt.Errorf("for scanning with type '%s', the file '%s' must exist in the project root",
			config.ScanType, pomPath)
	}

	defines := s.generateMavenWhitesourceDefines(config)
	flags, excludes := generateMavenWhitesourceFlags(config, utils)
	err := s.appendModulesThatWillBeScanned(utils, excludes)
	if err != nil {
		return fmt.Errorf("failed to determine maven modules which will be scanned: %w", err)
	}

	_, err = maven.Execute(&maven.ExecuteOptions{
		PomPath:             pomPath,
		M2Path:              config.M2Path,
		GlobalSettingsFile:  config.GlobalSettingsFile,
		ProjectSettingsFile: config.ProjectSettingsFile,
		Defines:             defines,
		Flags:               flags,
		Goals:               []string{"org.whitesource:whitesource-maven-plugin:19.5.1:update"},
	}, utils)

	return err
}

func (s *Scan) generateMavenWhitesourceDefines(config *ScanOptions) []string {
	defines := []string{
		"-Dorg.whitesource.orgToken=" + config.OrgToken,
		"-Dorg.whitesource.product=" + config.ProductName,
		"-Dorg.whitesource.checkPolicies=true",
		"-Dorg.whitesource.failOnError=true",
	}

	// Aggregate all modules into one WhiteSource project, if user specified the 'projectName' parameter.
	if config.ProjectName != "" {
		defines = append(defines, "-Dorg.whitesource.aggregateProjectName="+config.ProjectName)
		defines = append(defines, "-Dorg.whitesource.aggregateModules=true")
	}

	if config.UserToken != "" {
		defines = append(defines, "-Dorg.whitesource.userKey="+config.UserToken)
	}

	if s.ProductVersion != "" {
		defines = append(defines, "-Dorg.whitesource.productVersion="+s.ProductVersion)
	}

	return defines
}

func generateMavenWhitesourceFlags(config *ScanOptions, utils Utils) (flags []string, excludes []string) {
	excludes = config.BuildDescriptorExcludeList
	// From the documentation, these are file paths to a module's pom.xml.
	// For MTA projects, we want to support mixing paths to package.json files and pom.xml files.
	for _, exclude := range excludes {
		if !strings.HasSuffix(exclude, "pom.xml") {
			continue
		}
		exists, _ := utils.FileExists(exclude)
		if !exists {
			continue
		}
		moduleName := filepath.Dir(exclude)
		if moduleName != "" {
			flags = append(flags, "-pl", "!"+moduleName)
		}
	}
	return flags, excludes
}

func (s *Scan) appendModulesThatWillBeScanned(utils Utils, excludes []string) error {
	return maven.VisitAllMavenModules(".", utils, excludes, func(info maven.ModuleInfo) error {
		project := info.Project
		if project.Packaging != "pom" {
			if project.ArtifactID == "" {
				return fmt.Errorf("artifactId missing from '%s'", info.PomXMLPath)
			}

			err := s.AppendScannedProject(project.ArtifactID)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
