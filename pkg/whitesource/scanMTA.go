package whitesource

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"io"
	"os"
)

type mtaUtils interface {
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(executable string, params ...string) error

	Chdir(path string) error
	Getwd() (string, error)
	MkdirAll(path string, perm os.FileMode) error
	FileExists(path string) (bool, error)
	FileRead(path string) ([]byte, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
	FileRemove(path string) error
	RemoveAll(path string) error

	FindPackageJSONFiles(config *ScanOptions) ([]string, error)
	InstallAllNPMDependencies(config *ScanOptions, packageJSONFiles []string) error
}

// ExecuteMTAScan executes a scan for the Java part with maven, and performs a scan for each NPM module.
func (s *Scan) ExecuteMTAScan(config *ScanOptions, utils mtaUtils) error {
	log.Entry().Infof("Executing Whitesource scan for MTA project")
	pomExists, _ := utils.FileExists("pom.xml")
	if pomExists {
		if err := s.ExecuteMavenScanForPomFile(config, utils, "pom.xml"); err != nil {
			return err
		}
	}

	modules, err := utils.FindPackageJSONFiles(config)
	if err != nil {
		return err
	}
	if len(modules) > 0 {
		if err := s.ExecuteNpmScan(config, utils); err != nil {
			return err
		}
	}

	if !pomExists && len(modules) == 0 {
		return fmt.Errorf("neither Maven nor NPM modules found, no scan performed")
	}
	return nil
}
