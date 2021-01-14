package whitesource

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/magiconair/properties"
	"github.com/pkg/errors"
)

// ConfigOption defines a dedicated WhiteSource config which can be enforced if required
type ConfigOption struct {
	Name          string
	Value         interface{}
	OmitIfPresent string
	Force         bool
}

// ConfigOptions contains a list of config options (ConfigOption)
type ConfigOptions []ConfigOption

// RewriteUAConfigurationFile updates the user's Unified Agent configuration with configuration which should be enforced or just eases the overall configuration
// It then returns the path to the file containing the updated configuration
func (s *ScanOptions) RewriteUAConfigurationFile(utils Utils) (string, error) {

	// read config from inputFilePath or config.whitesource.configFilePath
	uaConfig, err := properties.LoadFile(s.ConfigFilePath, properties.UTF8)
	uaConfigMap := map[string]string{}
	if err != nil {
		log.Entry().Warningf("Failed to load configuration file '%v'. Creating a configuration file from scratch.", s.ConfigFilePath)
	} else {
		uaConfigMap = uaConfig.Map()
	}

	cOptions := ConfigOptions{}
	cOptions.addGeneralDefaults(s)
	cOptions.addBuildToolDefaults(s.BuildTool)

	newConfigMap := cOptions.updateConfig(&uaConfigMap)
	newConfig := properties.LoadMap(newConfigMap)

	now := time.Now().Format("20060102150405")
	newConfigFilePath := fmt.Sprintf("%v.%v", s.ConfigFilePath, now)

	var configContent bytes.Buffer
	_, err = newConfig.Write(&configContent, properties.UTF8)
	if err != nil {
		return "", errors.Wrap(err, "failed to write properties")
	}

	err = utils.FileWrite(newConfigFilePath, configContent.Bytes(), 0666)
	if err != nil {
		return "", errors.Wrap(err, "failed to write file")
	}

	return newConfigFilePath, nil
}

func (c *ConfigOptions) updateConfig(originalConfig *map[string]string) map[string]string {
	newConfig := map[string]string{}
	for k, v := range *originalConfig {
		newConfig[k] = v
	}

	for _, cOpt := range *c {
		//omit default if value present
		var dependentValue string
		if len(cOpt.OmitIfPresent) > 0 {
			dependentValue = newConfig[cOpt.OmitIfPresent]
		}
		if len(dependentValue) == 0 && (cOpt.Force || len(newConfig[cOpt.Name]) == 0) {
			newConfig[cOpt.Name] = fmt.Sprint(cOpt.Value)
		}
	}
	return newConfig
}

func (c *ConfigOptions) addGeneralDefaults(config *ScanOptions) {
	cOptions := ConfigOptions{}
	if strings.HasPrefix(config.ProductName, "DIST - ") {
		cOptions = append(cOptions, []ConfigOption{
			{Name: "checkPolicies", Value: false, Force: true},
			{Name: "forceCheckAllDependencies", Value: false, Force: true},
		}...)
	} else {
		cOptions = append(cOptions, []ConfigOption{
			{Name: "checkPolicies", Value: true, Force: true},
			{Name: "forceCheckAllDependencies", Value: true, Force: true},
		}...)
	}

	if config.Verbose {
		cOptions = append(cOptions, []ConfigOption{
			{Name: "log.level", Value: "debug"},
			{Name: "log.files.level", Value: "debug"},
		}...)
	}

	if len(config.Excludes) > 0 {
		cOptions = append(cOptions, ConfigOption{Name: "excludes", Value: config.Excludes, Force: true})
	}

	if len(config.Includes) > 0 {
		cOptions = append(cOptions, ConfigOption{Name: "includes", Value: config.Includes, Force: true})
	}

	//ToDo: handle m2 path - previously done in autoGenerateWhitesourceConfig()
	if config.BuildTool == "maven" && len(config.M2Path) > 0 {
		cOptions = append(cOptions, ConfigOption{Name: "maven.m2RepositoryPath", Value: config.M2Path, Force: true})
	}

	cOptions = append(cOptions, []ConfigOption{
		{Name: "apiKey", Value: config.OrgToken, Force: true},
		{Name: "productName", Value: config.ProductName, Force: true},
		{Name: "productVersion", Value: config.ProductVersion, Force: true},
		{Name: "projectName", Value: config.ProjectName, Force: true},
		{Name: "projectVersion", Value: config.ProductVersion, Force: true},
		{Name: "productToken", Value: config.ProductToken, OmitIfPresent: "projectToken", Force: true},
		{Name: "userKey", Value: config.UserToken, Force: true},
		{Name: "forceUpdate", Value: true, Force: true},
		{Name: "offline", Value: false, Force: true},
		{Name: "ignoreSourceFiles", Value: true, Force: true},
		{Name: "resolveAllDependencies", Value: false, Force: true},
		{Name: "failErrorLevel", Value: "ALL", Force: true},
		{Name: "case.sensitive.glob", Value: false},
		{Name: "followSymbolicLinks", Value: true},
	}...)

	for _, cOpt := range cOptions {
		*c = append(*c, cOpt)
	}
}

func (c *ConfigOptions) addBuildToolDefaults(buildTool string) error {
	buildToolDefaults := map[string]ConfigOptions{
		"docker": {
			{Name: "docker.scanImages", Value: true, Force: true},
			{Name: "docker.scanTarFiles", Value: true, Force: true},
			{Name: "docker.includes", Value: ".*.tar", Force: true},
			{Name: "ignoreSourceFiles", Value: true, Force: true},
			{Name: "python.resolveGlobalPackages", Value: true, Force: false},
			{Name: "resolveAllDependencies", Value: true, Force: false},
			{Name: "updateType", Value: "OVERRIDE", Force: true},
		},
		"dub": {
			{Name: "includes", Value: "**/*.d **/*.di"},
		},
		//ToDo: rename to go?
		//ToDo: switch to gomod as dependency manager
		"golang": {
			{Name: "go.resolveDependencies", Value: true, Force: true},
			{Name: "go.ignoreSourceFiles", Value: true, Force: true},
			{Name: "go.collectDependenciesAtRuntime", Value: false},
			{Name: "go.dependencyManager", Value: "modules"},
		},
		"gradle": {
			{Name: "gradle.localRepositoryPath", Value: ".gradle", Force: false},
		},
		"maven": {
			{Name: "updateEmptyProject", Value: true, Force: true},
			{Name: "maven.resolveDependencies", Value: true, Force: true},
			{Name: "maven.ignoreSourceFiles", Value: true, Force: true},
			{Name: "maven.aggregateModules", Value: false, Force: true},
			{Name: "maven.ignoredScopes", Value: "test provided"},
			{Name: "maven.ignorePomModules", Value: false},
			{Name: "maven.runPreStep", Value: true},
			{Name: "maven.projectNameFromDependencyFile", Value: true},
			{Name: "includes", Value: "**/*.jar"},
			{Name: "excludes", Value: "**/*sources.jar **/*javadoc.jar"},
		},
		"npm": {
			{Name: "npm.resolveDependencies", Value: true, Force: true},
			{Name: "npm.ignoreSourceFiles", Value: true, Force: true},
			{Name: "npm.ignoreNpmLsErrors", Value: true},
			{Name: "npm.failOnNpmLsErrors", Value: false},
			{Name: "npm.runPreStep", Value: true},
			{Name: "npm.projectNameFromDependencyFile", Value: true},
			{Name: "npm.resolveLockFile", Value: true},
		},
		"pip": {
			{Name: "python.resolveDependencies", Value: true, Force: true},
			{Name: "python.ignoreSourceFiles", Value: true, Force: true},
			{Name: "python.ignorePipInstallErrors", Value: false},
			{Name: "python.installVirtualenv", Value: true},
			{Name: "python.resolveHierarchyTree", Value: true},
			{Name: "python.requirementsFileIncludes", Value: "requirements.txt"},
			{Name: "python.resolveSetupPyFiles", Value: true},
			{Name: "python.runPipenvPreStep", Value: true},
			{Name: "python.pipenvDevDependencies", Value: true},
			{Name: "python.IgnorePipenvInstallErrors", Value: false},
			{Name: "includes", Value: "**/*.py **/*.txt"},
			{Name: "excludes", Value: "**/*sources.jar **/*javadoc.jar"},
		},
		"sbt": {
			{Name: "sbt.resolveDependencies", Value: true, Force: true},
			{Name: "sbt.ignoreSourceFiles", Value: true, Force: true},
			{Name: "sbt.aggregateModules", Value: false, Force: true},
			{Name: "sbt.runPreStep", Value: true},
			{Name: "includes", Value: "**/*.jar"},
			{Name: "excludes", Value: "**/*sources.jar **/*javadoc.jar"},
		},
	}
	if config := buildToolDefaults[buildTool]; config != nil {
		for _, cOpt := range config {
			*c = append(*c, cOpt)
		}
		return nil
	}

	//ToDo: Do we want to auto generate the config via autoGenerateWhitesourceConfig() here?
	// -> try to load original config file -> if not available generate?

	log.Entry().Infof("Configuration for buildTool: '%v' is not yet hardened, please do a quality assessment of your scan results.", buildTool)
	return fmt.Errorf("configuration not hardened")
}

//ToDo: Check if we want to optionally allow auto generation for unknown projects
func autoGenerateWhitesourceConfig(config *ScanOptions, utils Utils) error {
	// TODO: Should we rely on -detect, or set the parameters manually?
	if err := utils.RunExecutable("java", "-jar", config.AgentFileName, "-d", ".", "-detect"); err != nil {
		return err
	}

	// Rename generated config file to config.ConfigFilePath parameter
	if err := utils.FileRename("wss-generated-file.config", config.ConfigFilePath); err != nil {
		return err
	}
	return nil
}
