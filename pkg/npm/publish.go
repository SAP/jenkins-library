package npm

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
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
		return fmt.Errorf("error reading package scope from %s: %w", packageJSON, err)
	}

	npmignore := NewNPMIgnore(filepath.Dir(packageJSON))
	if exists, err := exec.Utils.FileExists(npmignore.filepath); exists {
		if err != nil {
			return fmt.Errorf("failed to check for existing %s file: %w", npmignore.filepath, err)
		}
		log.Entry().Debugf("loading existing %s file", npmignore.filepath)
		if err = npmignore.Load(); err != nil {
			return fmt.Errorf("failed to read existing %s file: %w", npmignore.filepath, err)
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
		return fmt.Errorf("failed to update %s file: %w", npmignore.filepath, err)
	}

	// update .piperNpmrc
	if len(registry) > 0 {
		// check existing .npmrc file
		if exists, err := exec.Utils.FileExists(npmrc.filepath); exists {
			if err != nil {
				return fmt.Errorf("failed to check for existing %s file: %w", npmrc.filepath, err)
			}
			log.Entry().Debugf("loading existing %s file", npmrc.filepath)
			if err = npmrc.Load(); err != nil {
				return fmt.Errorf("failed to read existing %s file: %w", npmrc.filepath, err)
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
			npmrc.Set(fmt.Sprintf("%s:%s", strings.TrimPrefix(registry, "https:"), "_auth"), piperutils.EncodeUsernamePassword(username, password))
			npmrc.Set("always-auth", "true")
		}
		// update .npmrc
		if err := npmrc.Write(); err != nil {
			return fmt.Errorf("failed to update %s file: %w", npmrc.filepath, err)
		}
	} else {
		log.Entry().Debug("no registry provided")
	}

	// Read version to check if it's a prerelease
	version, err := exec.readPackageVersion(packageJSON)
	if err != nil {
		return fmt.Errorf("failed to read package version from %s: %w", packageJSON, err)
	}

	tag := publishTag
	if tag == "" && isPrerelease(version) {
		tag = "prerelease"
		log.Entry().Infof("No publish tag provided, using '%s' based on version %s", tag, version)
	}

	if packBeforePublish {
		// change directory in package json file , since npm pack will run only for that packages
		if err = exec.Utils.Chdir(filepath.Dir(packageJSON)); err != nil {
			return fmt.Errorf("failed to change into directory for executing script: %w", err)
		}

		if err = execRunner.RunExecutable("npm", "pack"); err != nil {
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
		publishArgs := []string{"publish", "--tarball", tarballFilePath, "--userconfig", ".piperNpmrc", "--registry", registry}
		if tag != "" {
			publishArgs = append(publishArgs, "--tag", tag)
		}

		if err = execRunner.RunExecutable("npm", publishArgs...); err != nil {
			return fmt.Errorf("failed publishing artifact: %w", err)
		}

		if projectNpmrcExists {
			// undo the renaming ot the .npmrc to keep the workspace like before
			if err = exec.Utils.FileRename(projectNpmrc+".tmp", projectNpmrc); err != nil {
				log.Entry().Warnf("unable to rename the .npmrc file : %v", err)
			}
		}

		if err = exec.Utils.Chdir(oldWorkingDirectory); err != nil {
			return fmt.Errorf("failed to change back into original directory: %w", err)
		}
	} else {
		// Build publish command with --tag for prerelease versions (required by npm 11+)
		publishArgs := []string{"publish", "--userconfig", npmrc.filepath, "--registry", registry}
		if tag != "" {
			publishArgs = append(publishArgs, "--tag", tag)
		}

		if err = execRunner.RunExecutable("npm", publishArgs...); err != nil {
			return fmt.Errorf("failed publishing artifact: %w", err)
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

func (exec *Execute) readPackage(packageJSON string) (*npmMinimalPackageDescriptor, error) {
	b, err := exec.Utils.FileRead(packageJSON)
	if err != nil {
		return nil, err
	}

	var pd npmMinimalPackageDescriptor
	if err = json.Unmarshal(b, &pd); err != nil {
		return nil, err
	}

	return &pd, nil
}

func (exec *Execute) readPackageScope(packageJSON string) (string, error) {
	pd, err := exec.readPackage(packageJSON)
	if err != nil {
		return "", err
	}
	return pd.Scope(), nil
}

// readPackageVersion reads the version from package.json
func (exec *Execute) readPackageVersion(packageJSON string) (string, error) {
	pd, err := exec.readPackage(packageJSON)
	if err != nil {
		return "", err
	}
	if pd == nil {
		return "", fmt.Errorf("version not found in package descriptor %s", packageJSON)
	}

	return pd.Version, nil
}

// isPrerelease checks if a version string is a prerelease version
// According to semver spec, a prerelease version is indicated by appending a hyphen
// and a series of dot separated identifiers (e.g., 1.0.0-alpha, 1.0.0-beta.1, 0.0.2-20251029013231)
func isPrerelease(version string) bool {
	// Remove build metadata (anything after +) as it doesn't affect prerelease status
	version = strings.Split(version, "+")[0]

	// Check if there's a hyphen after the version numbers
	// This indicates a prerelease version per semver specification
	return strings.Contains(version, "-")
}
