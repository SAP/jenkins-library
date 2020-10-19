package whitesource

import (
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"io/ioutil"
	"os"
	"path/filepath"
)

const whiteSourceConfig = "whitesource.config.json"

func setValueAndLogChange(config map[string]interface{}, key string, value interface{}) {
	oldValue, exists := config[key]
	if exists && oldValue != value {
		log.Entry().Infof("overwriting '%s' in %s: %v -> %v", key, whiteSourceConfig, oldValue, value)
	}
	config[key] = value
}

func setValueOmitIfPresent(config map[string]interface{}, key, omitIfPresent string, value interface{}) {
	_, exists := config[omitIfPresent]
	if exists {
		return
	}
	setValueAndLogChange(config, key, value)
}

// writeWhitesourceConfigJSON creates or merges the file whitesource.config.json in the current
// directory from the given NPMScanOptions.
func (s *Scan) writeWhitesourceConfigJSON(config *ScanOptions, utils Utils, devDep, ignoreLsErrors bool) error {
	var npmConfig = make(map[string]interface{})

	exists, _ := utils.FileExists(whiteSourceConfig)
	if exists {
		fileContents, err := utils.FileRead(whiteSourceConfig)
		if err != nil {
			return fmt.Errorf("file '%s' already exists, but could not be read: %w", whiteSourceConfig, err)
		}
		err = json.Unmarshal(fileContents, &npmConfig)
		if err != nil {
			return fmt.Errorf("file '%s' already exists, but could not be parsed: %w", whiteSourceConfig, err)
		}
		log.Entry().Infof("The file '%s' already exists in the project. Changed config details will be logged.",
			whiteSourceConfig)
	}

	npmConfig["apiKey"] = config.OrgToken
	npmConfig["userKey"] = config.UserToken
	setValueAndLogChange(npmConfig, "checkPolicies", true)
	setValueAndLogChange(npmConfig, "productName", config.ProductName)
	setValueAndLogChange(npmConfig, "productVer", s.ProductVersion)
	setValueOmitIfPresent(npmConfig, "productToken", "projectToken", config.ProductToken)
	if config.ProjectName != "" {
		// In case there are other modules (i.e. maven modules in MTA projects),
		// or more than one NPM module, setting the project name will lead to
		// overwriting any previous scan results with the one from this module!
		// If this is not provided, the WhiteSource project name will be generated
		// from "name" in package.json plus " - " plus productVersion.
		setValueAndLogChange(npmConfig, "projectName", config.ProjectName)
	}
	setValueAndLogChange(npmConfig, "devDep", devDep)
	setValueAndLogChange(npmConfig, "ignoreNpmLsErrors", ignoreLsErrors)

	jsonBuffer, err := json.Marshal(npmConfig)
	if err != nil {
		return fmt.Errorf("failed to generate '%s': %w", whiteSourceConfig, err)
	}

	err = utils.FileWrite(whiteSourceConfig, jsonBuffer, 0644)
	if err != nil {
		return fmt.Errorf("failed to write '%s': %w", whiteSourceConfig, err)
	}
	return nil
}

// ExecuteNpmScan iterates over all found npm modules and performs a scan in each one.
func (s *Scan) ExecuteNpmScan(config *ScanOptions, utils Utils) error {
	modules, err := utils.FindPackageJSONFiles(config)
	if err != nil {
		return fmt.Errorf("failed to find package.json files with excludes: %w", err)
	}
	if len(modules) == 0 {
		return fmt.Errorf("found no NPM modules to scan. Configured excludes: %v",
			config.BuildDescriptorExcludeList)
	}
	for _, module := range modules {
		err := s.executeNpmScanForModule(module, config, utils)
		if err != nil {
			return fmt.Errorf("failed to scan NPM module '%s': %w", module, err)
		}
	}
	return nil
}

// executeNpmScanForModule generates a configuration file whitesource.config.json with appropriate values from config,
// installs all dependencies if necessary, and executes the scan via "npx whitesource run".
func (s *Scan) executeNpmScanForModule(modulePath string, config *ScanOptions, utils Utils) error {
	log.Entry().Infof("Executing Whitesource scan for NPM module '%s'", modulePath)

	resetDir, err := utils.Getwd()
	if err != nil {
		return fmt.Errorf("failed to obtain current directory: %w", err)
	}

	dir := filepath.Dir(modulePath)
	if err := utils.Chdir(dir); err != nil {
		return fmt.Errorf("failed to change into directory '%s': %w", dir, err)
	}
	defer func() {
		err = utils.Chdir(resetDir)
		if err != nil {
			log.Entry().Errorf("Failed to reset into directory '%s': %v", resetDir, err)
		}
	}()

	if err := s.writeWhitesourceConfigJSON(config, utils, false, true); err != nil {
		return err
	}
	defer func() { _ = utils.FileRemove(whiteSourceConfig) }()

	projectName, err := getNpmProjectName(modulePath, utils)
	if err != nil {
		return err
	}

	if err := reinstallNodeModulesIfLsFails(config, utils); err != nil {
		return err
	}

	if err := s.AppendScannedProject(projectName); err != nil {
		return err
	}

	return utils.RunExecutable("npx", "whitesource", "run")
}

// getNpmProjectName tries to read a property "name" of type string from the
// package.json file in the current directory and returns an error, if this is not possible.
func getNpmProjectName(modulePath string, utils Utils) (string, error) {
	fileContents, err := utils.FileRead("package.json")
	if err != nil {
		return "", fmt.Errorf("could not read %s: %w", modulePath, err)
	}
	var packageJSON = make(map[string]interface{})
	err = json.Unmarshal(fileContents, &packageJSON)

	projectNameEntry, exists := packageJSON["name"]
	if !exists {
		return "", fmt.Errorf("the file '%s' must configure a name", modulePath)
	}

	projectName, isString := projectNameEntry.(string)
	if !isString {
		return "", fmt.Errorf("the file '%s' must configure a name", modulePath)
	}

	return projectName, nil
}

// reinstallNodeModulesIfLsFails tests running of "npm ls".
// If that fails, the node_modules directory is cleared and the file "package-lock.json" is removed.
// Then "npm install" is performed. Without this, the npm whitesource plugin will consistently hang,
// when encountering npm ls errors, even with "ignoreNpmLsErrors:true" in the configuration.
// The consequence is that what was scanned is not guaranteed to be identical to what was built & deployed.
// This hack/work-around that should be removed once scanning it consistently performed using the Unified Agent.
// A possible reason for encountering "npm ls" errors in the first place is that a different node version
// is used for whitesourceExecuteScan due to a different docker image being used compared to the build stage.
func reinstallNodeModulesIfLsFails(config *ScanOptions, utils Utils) error {
	// No need to have output from "npm ls" in the log
	utils.Stdout(ioutil.Discard)
	defer utils.Stdout(log.Writer())

	err := utils.RunExecutable("npm", "ls")
	if err == nil {
		return nil
	}
	log.Entry().Warnf("'npm ls' failed. Re-installing NPM Node Modules")
	err = utils.RemoveAll("node_modules")
	if err != nil {
		return fmt.Errorf("failed to remove node_modules directory: %w", err)
	}
	err = utils.MkdirAll("node_modules", os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to recreate node_modules directory: %w", err)
	}
	exists, _ := utils.FileExists("package-lock.json")
	if exists {
		err = utils.FileRemove("package-lock.json")
		if err != nil {
			return fmt.Errorf("failed to remove package-lock.json: %w", err)
		}
	}
	// Passing only "package.json", because we are already inside the module's directory.
	return utils.InstallAllNPMDependencies(config, []string{"package.json"})
}

// ExecuteYarnScan generates a configuration file whitesource.config.json with appropriate values from config,
// installs whitesource yarn plugin and executes the scan.
func (s *Scan) ExecuteYarnScan(config *ScanOptions, utils Utils) error {
	// To stay compatible with what the step was doing before, trigger aggregation, although
	// there is a great chance that it doesn't work with yarn the same way it doesn't with npm.
	// Maybe the yarn code-path should be removed, and only npm stays.
	config.ProjectName = s.AggregateProjectName
	if err := s.writeWhitesourceConfigJSON(config, utils, true, false); err != nil {
		return err
	}
	defer func() { _ = utils.FileRemove(whiteSourceConfig) }()
	if err := utils.RunExecutable("yarn", "global", "add", "whitesource"); err != nil {
		return err
	}
	if err := utils.RunExecutable("yarn", "install"); err != nil {
		return err
	}
	if err := utils.RunExecutable("whitesource", "yarn"); err != nil {
		return err
	}
	return nil
}
