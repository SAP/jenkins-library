package npm

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/log"
	CredentialUtils "github.com/SAP/jenkins-library/pkg/piperutils"
)

type npmMinimalPackageDescriptor struct {
	Name    string `json:version`
	Version string `json:version`
}

func (pd *npmMinimalPackageDescriptor) Scope() string {
	r := regexp.MustCompile(`^(?:(?P<scope>@[^\/]+)\/)?(?P<package>.+)$`)

	matches := r.FindStringSubmatch(pd.Name)

	if len(matches) == 0 {
		return ""
	}

	return matches[1]
}

// PublishAllPackages executes npm publish for all package.json files defined in packageJSONFiles list
func (exec *Execute) PublishAllPackages(packageJSONFiles []string, registry, username, password string, packBeforePublish bool) error {
	for _, packageJSON := range packageJSONFiles {
		log.Entry().Infof("triggering publish for %s", packageJSON)

		fileExists, err := exec.Utils.FileExists(packageJSON)
		if err != nil {
			return fmt.Errorf("cannot check if '%s' exists: %w", packageJSON, err)
		}
		if !fileExists {
			return fmt.Errorf("package.json file '%s' not found: %w", packageJSON, err)
		}

		err = exec.publish(packageJSON, registry, username, password, packBeforePublish, packageJSONFiles)
		if err != nil {
			return err
		}
	}
	return nil
}

// publish executes npm publish for package.json
func (exec *Execute) publish(packageJSON, registry, username, password string, packBeforePublish bool, packageJSONFiles []string) error {
	execRunner := exec.Utils.GetExecRunner()

	scope, err := exec.readPackageScope(packageJSON)

	if err != nil {
		return errors.Wrapf(err, "error reading package scope from %s", packageJSON)
	}

	npmignore := NewNPMIgnore(filepath.Dir(packageJSON))
	if exists, err := exec.Utils.FileExists(npmignore.filepath); exists {
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
		if exists, err := exec.Utils.FileExists(npmrc.filepath); exists {
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

		if len(scope) > 0 {
			npmrc.Set(fmt.Sprintf("%s:registry", scope), registry)
		}

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
		tmpDirectory, err := exec.Utils.TempDir(".", "temp-")

		if err != nil {
			return errors.Wrap(err, "creating temp directory failed")
		}

		defer exec.Utils.RemoveAll(tmpDirectory)

		err = execRunner.RunExecutable("npm", "pack", "--pack-destination", tmpDirectory)
		if err != nil {
			return err
		}

		_, err = exec.Utils.Copy(npmrc.filepath, filepath.Join(tmpDirectory, ".piperNpmrc"))
		if err != nil {
			return fmt.Errorf("error copying piperNpmrc file from %v to %v with error: %w",
				npmrc.filepath, filepath.Join(tmpDirectory, ".piperNpmrc"), err)
		}

		tarballs, err := exec.Utils.Glob(filepath.Join(tmpDirectory, "*.tgz"))

		if err != nil {
			return err
		}

		if len(tarballs) != 1 {
			return fmt.Errorf("found more tarballs than expected: %v", tarballs)
		}

		tarballFilePath, err := exec.Utils.Abs(tarballs[0])

		if err != nil {
			return err
		}

		// all existing npmrc file may interfear with publish hence rename them
		if err = exec.renameExistingNpmrcFiles(packageJSONFiles, true); err != nil {
			return err
		}

		err = execRunner.RunExecutable("npm", "publish", "--tarball", tarballFilePath, "--userconfig", filepath.Join(tmpDirectory, ".piperNpmrc"), "--registry", registry)
		if err != nil {
			return errors.Wrap(err, "failed publishing artifact")
		}

		// rename the all renamed npmrc file to original name since they would need to be packed in further publish , specially in cf deploy cases
		if err = exec.renameExistingNpmrcFiles(packageJSONFiles, false); err != nil {
			return err
		}

	} else {
		err := execRunner.RunExecutable("npm", "publish", "--userconfig", npmrc.filepath, "--registry", registry)
		if err != nil {
			return errors.Wrap(err, "failed publishing artifact")
		}
	}

	return nil
}

func (exec *Execute) renameExistingNpmrcFiles(packageJSONFiles []string, fromOriginalToTmp bool) error {

	for _, packageJSON := range packageJSONFiles {

		orginalFileName := filepath.Join(filepath.Dir(packageJSON), ".npmrc")
		toBeReplaced := filepath.Join(filepath.Dir(packageJSON), ".npmrc.tmp")

		if !fromOriginalToTmp {
			orginalFileName = filepath.Join(filepath.Dir(packageJSON), ".npmrc.tmp")
			toBeReplaced = filepath.Join(filepath.Dir(packageJSON), ".npmrc")
		}

		projectNpmrcExists, _ := exec.Utils.FileExists(orginalFileName)

		if projectNpmrcExists {
			// rename the .npmrc file since it interferes with publish
			err := exec.Utils.FileRename(orginalFileName, toBeReplaced)
			if err != nil {
				return fmt.Errorf("error when renaming current .npmrc from %v to %v : %w", orginalFileName, toBeReplaced, err)
			}
		}
	}

	return nil
}

func (exec *Execute) readPackageScope(packageJSON string) (string, error) {
	b, err := exec.Utils.FileRead(packageJSON)

	if err != nil {
		return "", err
	}

	var pd npmMinimalPackageDescriptor

	json.Unmarshal(b, &pd)

	return pd.Scope(), nil
}
