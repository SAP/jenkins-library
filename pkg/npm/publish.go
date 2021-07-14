package npm

import (
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/log"
	CredentialUtils "github.com/SAP/jenkins-library/pkg/piperutils"
	FileUtils "github.com/SAP/jenkins-library/pkg/piperutils"
)

// PublishAllPackages executes npm publish for all package.json files defined in packageJSONFiles list
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

	npmignore := NewNPMIgnore(filepath.Dir(packageJSON))
	if exists, err := FileUtils.FileExists(npmignore.filepath); exists {
		if err != nil {
			return errors.Wrapf(err, "failed to check for existing %s file", npmignore.filepath)
		}
		log.Entry().Debugf("loading existing %s file", npmignore.filepath)
		if err = npmignore.Load(); err != nil {
			return errors.Wrapf(err, "failed to read existing %s file", npmignore.filepath)
		}
	} else {
		log.Entry().Debug("creating .npmignore file")
	}
	log.Entry().Debug("adding **/piper")
	npmignore.Add("**/piper")
	log.Entry().Debug("adding **/sap-piper")
	npmignore.Add("**/sap-piper")
	// update .npmrc
	if err := npmignore.Write(); err != nil {
		return errors.Wrapf(err, "failed to update %s file", npmignore.filepath)
	}

	if len(registry) > 0 {
		npmrc := NewNPMRC(filepath.Dir(packageJSON))
		// check existing .npmrc file
		if exists, err := FileUtils.FileExists(npmrc.filepath); exists {
			if err != nil {
				return errors.Wrapf(err, "failed to check for existing %s file", npmrc.filepath)
			}
			log.Entry().Debugf("loading existing %s file", npmrc.filepath)
			if err = npmrc.Load(); err != nil {
				return errors.Wrapf(err, "failed to read existing %s file", npmrc.filepath)
			}
		} else {
			log.Entry().Debug("creating .npmrc file")
		}
		// set registry
		log.Entry().Debugf("adding registry %s", registry)
		npmrc.Set("registry", registry)
		// set registry auth
		if len(username) > 0 && len(password) > 0 {
			log.Entry().Debug("adding registry credentials")
			npmrc.Set("_auth", CredentialUtils.EncodeUsernamePassword(username, password))
			npmrc.Set("always-auth", "true")
		}
		// update .npmrc
		if err := npmrc.Write(); err != nil {
			return errors.Wrapf(err, "failed to update %s file", npmrc.filepath)
		}
	} else {
		log.Entry().Debug("no registry provided")
	}

	err := execRunner.RunExecutable("npm", "publish", filepath.Dir(packageJSON))
	if err != nil {
		return err
	}
	return nil
}
