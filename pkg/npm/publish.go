package npm

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	CredentialUtils "github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/versioning"
)

type npmMinimalPackageDescriptor struct {
	Name    string `json:"name"`
	Version string `json:"version"`
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
func (exec *Execute) PublishAllPackages(packageJSONFiles []string, registry, username, password, publishTag string, packBeforePublish bool, buildCoordinates *[]versioning.Coordinates) error {
	for _, packageJSON := range packageJSONFiles {
		log.Entry().Infof("triggering publish for %s", packageJSON)

		fileExists, err := exec.Utils.FileExists(packageJSON)
		if err != nil {
			return fmt.Errorf("cannot check if '%s' exists: %w", packageJSON, err)
		}
		if !fileExists {
			return fmt.Errorf("package.json file '%s' not found: %w", packageJSON, err)
		}

		err = exec.publish(packageJSON, registry, username, password, publishTag, packBeforePublish, buildCoordinates)
		if err != nil {
			return err
		}
	}
	return nil
}

// publish executes npm publish for package.json
func (exec *Execute) publish(packageJSON, registry, username, password, publishTag string, packBeforePublish bool, buildCoordinates *[]versioning.Coordinates) error {
	execRunner := exec.Utils.GetExecRunner()

	oldWorkingDirectory, err := exec.Utils.Getwd()

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
	// temporary installation folder used to install BOM to be ignored
	log.Entry().Debug("adding tmp to npmignore")
	npmignore.Add("tmp/")
	log.Entry().Debug("adding sboms to npmignore")
	npmignore.Add("**/bom*.{xml,json}")

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
			// See https://github.blog/changelog/2022-10-24-npm-v9-0-0-released/
			// where it states: the presence of auth related settings that are not scoped to a specific registry found in a config file
			// is no longer supported and will throw errors
			npmrc.Set(fmt.Sprintf("%s:%s", strings.TrimPrefix(registry, "https:"), "_auth"), CredentialUtils.EncodeUsernamePassword(username, password))
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
		// change directory in package json file , since npm pack will run only for that packages
		if err := exec.Utils.Chdir(filepath.Dir(packageJSON)); err != nil {
			return fmt.Errorf("failed to change into directory for executing script: %w", err)
		}

		if err := execRunner.RunExecutable("npm", "pack"); err != nil {
			return err
		}

		tarballs, err := exec.Utils.Glob(filepath.Join(".", "*.tgz"))
		if err != nil {
			return err
		}

		// we do not maintain the tarball file name and hence expect only one tarball that comes
		// from the npm pack command
		if len(tarballs) < 1 {
			return fmt.Errorf("no tarballs found")
		}
		if len(tarballs) > 1 {
			return fmt.Errorf("found more tarballs than expected: %v", tarballs)
		}

		tarballFilePath, err := exec.Utils.Abs(tarballs[0])
		if err != nil {
			return err
		}

		// if a user has a .npmrc file and if it has a scope (e.g @sap to download scoped dependencies)
		// if the package to be published also has the same scope (@sap) then npm gets confused
		// and tries to publish to the scope that comes from the npmrc file
		// and is not the desired publish since we want to publish to the other registry (from .piperNpmrc)
		// file and not to the one mentioned in the users npmrc file
		// to solve this we rename the users npmrc file before publish, the original npmrc is already
		// packaged in the tarball and hence renaming it before publish should not have an effect
		projectNpmrc := filepath.Join(".", ".npmrc")
		projectNpmrcExists, _ := exec.Utils.FileExists(projectNpmrc)

		if projectNpmrcExists {
			// rename the .npmrc file since it interferes with publish
			err = exec.Utils.FileRename(projectNpmrc, projectNpmrc+".tmp")
			if err != nil {
				return fmt.Errorf("error when renaming current .npmrc file : %w", err)
			}
		}

		// Build publish command with --tag for prerelease versions (required by npm 11+)
		// publishArgs := []string{"publish", "--tarball", tarballFilePath, "--userconfig", ".piperNpmrc", "--registry", registry, "--tag", publishTag}
		publishArgs := []string{"publish", "--tarball", tarballFilePath, "--userconfig", ".piperNpmrc", "--registry", registry}
		err = execRunner.RunExecutable("npm", publishArgs...)
		if err != nil {
			return errors.Wrap(err, "failed publishing artifact")
		}

		if projectNpmrcExists {
			// undo the renaming ot the .npmrc to keep the workspace like before
			err = exec.Utils.FileRename(projectNpmrc+".tmp", projectNpmrc)
			if err != nil {
				log.Entry().Warnf("unable to rename the .npmrc file : %v", err)
			}
		}

		if err := exec.Utils.Chdir(oldWorkingDirectory); err != nil {
			return fmt.Errorf("failed to change back into original directory: %w", err)
		}
	} else {
		// Build publish command with --tag for prerelease versions (required by npm 11+)
		// publishArgs := []string{"publish", "--userconfig", npmrc.filepath, "--registry", registry, "--tag", publishTag}
		publishArgs := []string{"publish", "--userconfig", npmrc.filepath, "--registry", registry}
		err = execRunner.RunExecutable("npm", publishArgs...)
		if err != nil {
			return errors.Wrap(err, "failed publishing artifact")
		}
	}

	options := versioning.Options{}
	var utils versioning.Utils

	artifact, err := versioning.GetArtifact("npm", packageJSON, &options, utils)
	if err != nil {
		log.Entry().Warnf("unable to get artifact metdata : %v", err)
	} else {
		coordinate, err := artifact.GetCoordinates()
		if err != nil {
			log.Entry().Warnf("unable to get artifact coordinates : %v", err)
		} else {
			component := piperutils.GetComponent(filepath.Join(filepath.Dir(packageJSON), npmBomFilename))
			coordinate.BuildPath = filepath.Dir(packageJSON)
			coordinate.URL = registry
			coordinate.Packaging = "tgz"
			coordinate.PURL = component.Purl

			*buildCoordinates = append(*buildCoordinates, coordinate)
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
