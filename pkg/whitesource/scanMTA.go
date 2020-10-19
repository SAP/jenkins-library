package whitesource

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
)

// ExecuteMTAScan executes a scan for the Java part with maven, and performs a scan for each NPM module.
func (s *Scan) ExecuteMTAScan(config *ScanOptions, utils Utils) error {
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
