package npm

import (
	"fmt"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/pkg/errors"
)

// PublishAllPackages executes npm or yarn Install for all package.json fileUtils defined in packageJSONFiles
func (exec *Execute) PublishAllPackages(packageJSONFiles []string, registry, username, password string) error {
	for _, packageJSON := range packageJSONFiles {
		fileExists, err := exec.Utils.FileExists(packageJSON)
		if err != nil {
			return fmt.Errorf("cannot check if '%s' exists: %w", packageJSON, err)
		}
		if !fileExists {
			return fmt.Errorf("package.json file '%s' not found: %w", packageJSON, err)
		}

		err = exec.publish(packageJSON, registry, username, password)
		if err != nil {
			return err
		}
	}
	return nil
}

// publish executes npm publish for package.json
func (exec *Execute) publish(packageJSON, registry, username, password string) error {
	execRunner := exec.Utils.GetExecRunner()

	if len(registry) > 0 {
		log.Entry().Info("Registry provided, creating .npmrc file!")
		npmrc := NewNPMRC(filepath.Dir(packageJSON))

		exists, err := piperutils.FileExists(npmrc.path)
		if err != nil {
			return errors.Wrapf(err, "failed to read existing %s file", npmrc.path)
		}
		if exists {
			npmrc.Load()
		}
		log.Entry().Debugf("content: %s", npmrc.Print())

		npmrc.Set("registry", registry)
		log.Entry().Debugf("content: %s", npmrc.Print())

		if len(username) > 0 && len(password) > 0 {
			npmrc.SetAuth(registry, username, password)
			log.Entry().Debugf("content: %s", npmrc.Print())
		}

		err = npmrc.Write()
		if err != nil {
			return err
		}
	} else {
		log.Entry().Info("No registry provided!")
	}

	err := execRunner.RunExecutable("npm", "publish --dry-run", filepath.Dir(packageJSON))
	if err != nil {
		return err
	}
	return nil
}
