package npm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/log"
	CredentialUtils "github.com/SAP/jenkins-library/pkg/piperutils"
	FileUtils "github.com/SAP/jenkins-library/pkg/piperutils"
)

// PublishAllPackages executes npm publish for all package.json files defined in packageJSONFiles list
func (exec *Execute) PublishAllPackages(packageJSONFiles []string, registry, username, password string, packBeforePublish bool) error {
	for _, packageJSON := range packageJSONFiles {
		fileExists, err := exec.Utils.FileExists(packageJSON)
		if err != nil {
			return fmt.Errorf("cannot check if '%s' exists: %w", packageJSON, err)
		}
		if !fileExists {
			return fmt.Errorf("package.json file '%s' not found: %w", packageJSON, err)
		}

		err = exec.publish(packageJSON, registry, username, password, packBeforePublish)
		if err != nil {
			return err
		}
	}
	return nil
}

// publish executes npm publish for package.json
func (exec *Execute) publish(packageJSON, registry, username, password string, packBeforePublish bool) error {
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

	npmrc := NewNPMRC(filepath.Dir(packageJSON))

	log.Entry().Debugf("adding piper npmrc file %v", npmrc.filepath)
	npmignore.Add(npmrc.filepath)

	if err := npmignore.Write(); err != nil {
		return errors.Wrapf(err, "failed to update %s file", npmignore.filepath)
	}

	// update .piperNpmrc
	if len(registry) > 0 {
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
			log.Entry().Debugf("creating new npmrc file at %s", npmrc.filepath)
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

	if packBeforePublish {
		tmpDirectory := getTempDirForNpmTarBall()
		//defer os.RemoveAll(tmpDirectory)

		err := execRunner.RunExecutable("npm", "pack", "--pack-destination", tmpDirectory)
		if err != nil {
			return err
		}

		_, err = FileUtils.Copy(npmrc.filepath, filepath.Join(tmpDirectory, ".piperNpmrc"))
		if err != nil {
			return fmt.Errorf("error copying piperNpmrc file from %v to %v with error: %w",
				npmrc.filepath, filepath.Join(tmpDirectory, ".piperNpmrc"), err)
		}

		tarballFileName := ""
		err = filepath.Walk(tmpDirectory, func(path string, info os.FileInfo, err error) error {
			if filepath.Ext(path) == ".tgz" {
				tarballFileName = filepath.Base(path)
				log.Entry().Debugf("found tarball file at %v", tarballFileName)
			}
			return nil
		})
		if err != nil {
			return err
		}

		err = os.Rename(filepath.Join(filepath.Dir(packageJSON), ".npmrc"), filepath.Join(filepath.Dir(packageJSON), ".tmpNpmrc"))
		if err != nil {
			return fmt.Errorf("error when renaming current .npmrc file : %w", err)
		}

		err = os.Chdir(tmpDirectory)
		if err != nil {
			return fmt.Errorf("error when changing directory to %v with error : %w", tmpDirectory, err)
		}

		err = execRunner.RunExecutable("npm", "publish", "--tarball", tarballFileName, "--userconfig", ".piperNpmrc", "--registry", registry)
		if err != nil {
			return err
		}

		log.Entry().Debugf("renaming original .npmrc files from %v to %v", filepath.Join(filepath.Dir(packageJSON), ".tmpNpmrc"), filepath.Join(filepath.Dir(packageJSON), ".npmrc"))
		err = os.Rename(filepath.Join(filepath.Dir(packageJSON), ".tmpNpmrc"), filepath.Join(filepath.Dir(packageJSON), ".npmrc"))
		if err != nil {
			return fmt.Errorf("error when renaming current .npmrc file : %w", err)
		}
	} else {
		err := execRunner.RunExecutable("npm", "publish", "--userconfig", npmrc.filepath, "--registry", registry)
		if err != nil {
			return err
		}
	}

	return nil
}

func getTempDirForNpmTarBall() string {
	tmpFolder, err := ioutil.TempDir(".", "temp-")
	if err != nil {
		log.Entry().WithError(err).WithField("path", tmpFolder).Debug("Creating temp directory failed")
	}
	return tmpFolder
}
