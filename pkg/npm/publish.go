package npm

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/versioning"
)

type RCManager interface {
	SetRegistry(registry, username, password, scope string) error
	GetFilePath() string
}

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
func (exec *Execute) PublishAllPackages(packageJSONFiles []string, registry, username, password string, packBeforePublish bool, buildCoordinates *[]versioning.Coordinates) error {
	for _, packageJSON := range packageJSONFiles {
		log.Entry().Infof("triggering publish for %s", packageJSON)

		fileExists, err := exec.Utils.FileExists(packageJSON)
		if err != nil {
			return fmt.Errorf("cannot check if '%s' exists: %w", packageJSON, err)
		}
		if !fileExists {
			return fmt.Errorf("package.json file '%s' not found: %w", packageJSON, err)
		}

		err = exec.publish(packageJSON, registry, username, password, packBeforePublish, buildCoordinates)
		if err != nil {
			return err
		}
	}
	return nil
}

// publish executes npm publish for package.json
func (exec *Execute) publish(packageJSON, registry, username, password string, packBeforePublish bool, buildCoordinates *[]versioning.Coordinates) error {
	oldWorkingDirectory, err := exec.Utils.Getwd()

	scope, err := exec.readPackageScope(packageJSON)
	if err != nil {
		return fmt.Errorf("failed to read package scope: %w", err)
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
	npmignore.Add("**/bom*.xml")

	log.Entry().Debugf("adding piper npmrc file %v", exec.Tool.RC.GetFilePath())
	npmignore.Add(exec.Tool.RC.GetFilePath())

	if err := npmignore.Write(); err != nil {
		return fmt.Errorf("failed to update %s file: %w", npmignore.filepath, err)
	}

	if err := exec.Tool.RC.SetRegistry(registry, username, password, scope); err != nil {
		return fmt.Errorf("failed to configure registry: %w", err)
	}

	if packBeforePublish {
		// change directory in package json file , since npm pack will run only for that packages
		if err := exec.Utils.Chdir(filepath.Dir(packageJSON)); err != nil {
			return fmt.Errorf("failed to change into directory for executing script: %w", err)
		}

		if err := exec.Tool.Pack(); err != nil {
			return fmt.Errorf("failed to run %s pack: %w", exec.Tool.Name, err)
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

		// TODO: handle for npm
		// if a user has a .npmrc file and if it has a scope (e.g @sap to download scoped dependencies)
		// if the package to be published also has the same scope (@sap) then npm gets confused
		// and tries to publish to the scope that comes from the npmrc file
		// and is not the desired publish since we want to publish to the other registry (from .piperNpmrc)
		// file and not to the one mentioned in the users npmrc file
		// to solve this we rename the users npmrc file before publish, the original npmrc is already
		// packaged in the tarball and hence renaming it before publish should not have an effect
		// projectNpmrc := filepath.Join(".", ".npmrc")
		// projectNpmrcExists, _ := exec.Utils.FileExists(projectNpmrc)

		// if projectNpmrcExists {
		// 	// rename the .npmrc file since it interferes with publish
		// 	err = exec.Utils.FileRename(projectNpmrc, projectNpmrc+".tmp")
		// 	if err != nil {
		// 		return fmt.Errorf("error when renaming current .npmrc file : %w", err)
		// 	}
		// }

		if err := exec.Tool.Publish("--tarball", tarballFilePath); err != nil {
			return fmt.Errorf("failed to run %s publish: %w", exec.Tool.Name, err)
		}

		// TODO: handle for npm
		// if projectNpmrcExists {
		// 	// undo the renaming ot the .npmrc to keep the workspace like before
		// 	err = exec.Utils.FileRename(projectNpmrc+".tmp", projectNpmrc)
		// 	if err != nil {
		// 		log.Entry().Warnf("unable to rename the .npmrc file : %v", err)
		// 	}
		// }

		if err := exec.Utils.Chdir(oldWorkingDirectory); err != nil {
			return fmt.Errorf("failed to change back into original directory: %w", err)
		}
	} else {
		if err := exec.Tool.Publish("--registry", registry); err != nil {
			return fmt.Errorf("failed to run %s publish: %w", exec.Tool.Name, err)
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
			coordinate.BuildPath = filepath.Dir(packageJSON)
			coordinate.URL = registry
			coordinate.Packaging = "tgz"
			coordinate.PURL = piperutils.GetPurl(filepath.Join(filepath.Dir(packageJSON), npmBomFilename))

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
