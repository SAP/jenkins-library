package python

import (
	"fmt"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
)

const (
	BOMFilename = "bom-pip.xml"
)

func CreateBOM(
	executeFn func(executable string, params ...string) error,
	existsFn func(path string) (bool, error),
	virtualEnv string,
	requirementsFile string,
	cycloneDxVersion string,
	cycloneDxSchemaVersion string,
) error {
	if exists, _ := existsFn(requirementsFile); exists {
		if err := InstallRequirements(executeFn, virtualEnv, requirementsFile); err != nil {
			return err
		}
	} else {
		log.Entry().Warnf("unable to find requirements.txt file at %s , continuing SBOM generation without requirements.txt", requirementsFile)
	}

	if err := InstallCycloneDX(executeFn, virtualEnv, cycloneDxVersion); err != nil {
		return err
	}

	cycloneDxBinary := "cyclonedx-py"
	if len(virtualEnv) > 0 {
		cycloneDxBinary = filepath.Join(virtualEnv, "bin", cycloneDxBinary)
	}

	log.Entry().Debug("creating BOM")
	if err := executeFn(cycloneDxBinary,
		"env",
		"--output-file", BOMFilename,
		"--output-format", "XML",
		"--spec-version", cycloneDxSchemaVersion,
	); err != nil {
		return fmt.Errorf("failed to create BOM: %w", err)
	}
	return nil
}
