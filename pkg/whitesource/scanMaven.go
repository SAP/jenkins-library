package whitesource

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"io"
	"path/filepath"
	"strings"
)

// MavenScanOptions contains parameters needed during the scan.
type MavenScanOptions struct {
	// ScanType defines the type of scan. Can be "maven" or "mta" for scanning with Maven.
	ScanType    string
	OrgToken    string
	UserToken   string
	ProductName string
	// ProjectName is an optional name for an "aggregator" project.
	// All scanned maven modules will be reflected in the aggregate project.
	ProjectName                string
	BuildDescriptorExcludeList []string
	// PomPath is the path to root build descriptor file.
	PomPath string
	// M2Path is the path to the local maven repository.
	M2Path string
	// GlobalSettingsFile is an optional path to a global maven settings file.
	GlobalSettingsFile string
	// ProjectSettingsFile is an optional path to a local maven settings file.
	ProjectSettingsFile string
}

type mavenUtils interface {
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(executable string, params ...string) error

	FileExists(path string) (bool, error)
	FileRead(path string) ([]byte, error)
}

// ExecuteMavenScan constructs maven parameters from the given configuration, and executes the maven goal
// "org.whitesource:whitesource-maven-plugin:19.5.1:update".
func (s *Scan) ExecuteMavenScan(config *MavenScanOptions, utils mavenUtils) error {
	log.Entry().Infof("Using Whitesource scan for Maven project")
	pomPath := config.PomPath
	if pomPath == "" {
		pomPath = "pom.xml"
	}
	return s.ExecuteMavenScanForPomFile(config, utils, pomPath)
}

// ExecuteMavenScanForPomFile constructs maven parameters from the given configuration, and executes the maven goal
// "org.whitesource:whitesource-maven-plugin:19.5.1:update" for the given pom file.
func (s *Scan) ExecuteMavenScanForPomFile(config *MavenScanOptions, utils mavenUtils, pomPath string) error {
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

func (s *Scan) generateMavenWhitesourceDefines(config *MavenScanOptions) []string {
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

func generateMavenWhitesourceFlags(config *MavenScanOptions, utils mavenUtils) (flags []string, excludes []string) {
	excludes = config.BuildDescriptorExcludeList
	if len(excludes) == 0 {
		excludes = []string{
			filepath.Join("unit-tests", "pom.xml"),
			filepath.Join("integration-tests", "pom.xml"),
			filepath.Join("performance-tests", "pom.xml"),
		}
	}
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

func (s *Scan) appendModulesThatWillBeScanned(utils mavenUtils, excludes []string) error {
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
