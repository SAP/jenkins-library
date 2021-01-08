package whitesource

import (
	"bytes"
	"fmt"
	"io/ioutil"
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

// Config specifies WhiteSource-specific configuration used for configuring the Unified Agent
type Config struct {
	BuildTool      string
	ConfigFilePath string
	OrgToken       string
	ProductName    string
	ProductToken   string
	ProductVersion string
	ProjectName    string
	UserKey        string
	Verbose        bool
}

// RewriteUAConfigurationFile updates the user's Unified Agent configuration with configuration which should be enforced or just eases the overall configuration
// It then returns the path to the file containing the updated configuration
func (c *Config) RewriteUAConfigurationFile() (string, error) {

	// read config from inputFilePath or config.whitesource.configFilePath
	uaConfig, err := properties.LoadFile(c.ConfigFilePath, properties.UTF8)
	uaConfigMap := map[string]string{}
	if err != nil {
		log.Entry().Warningf("Failed to load configuration file '%v'. Creating a configuration file from scratch.", c.ConfigFilePath)
	} else {
		uaConfigMap = uaConfig.Map()
	}

	cOptions := ConfigOptions{}
	cOptions.addGeneralDefaults(c)
	cOptions.addBuildToolDefaults(c.BuildTool)

	newConfigMap := cOptions.updateConfig(&uaConfigMap)
	newConfig := properties.LoadMap(newConfigMap)

	now := time.Now().Format("20060102150405")
	newConfigFilePath := fmt.Sprintf("%v.%v", c.ConfigFilePath, now)

	var configContent bytes.Buffer
	_, err = newConfig.Write(&configContent, properties.UTF8)
	if err != nil {
		return "", errors.Wrap(err, "failed to write properties")
	}
	err = ioutil.WriteFile(newConfigFilePath, configContent.Bytes(), 0666)
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

func (c *ConfigOptions) addGeneralDefaults(config *Config) {
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

	cOptions = append(cOptions, []ConfigOption{
		{Name: "apiKey", Value: config.OrgToken, Force: true},
		{Name: "productName", Value: config.ProductName, Force: true},
		{Name: "productVersion", Value: config.ProductVersion, Force: true},
		{Name: "projectName", Value: config.ProjectName, Force: true},
		{Name: "projectVersion", Value: config.ProductVersion, Force: true},
		{Name: "productToken", Value: config.ProductToken, OmitIfPresent: "projectToken", Force: true},
		{Name: "userKey", Value: config.UserKey, Force: true},
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
			//ToDo: check value! Was /.*.tar/ in groovy
			{Name: "docker.includes", Value: "/.*.tar/", Force: true},
			{Name: "ignoreSourceFiles", Value: true, Force: true},
			{Name: "python.resolveGlobalPackages", Value: true, Force: false},
			{Name: "resolveAllDependencies", Value: true, Force: false},
			{Name: "updateType", Value: "OVERRIDE", Force: true},
		},
		"dub": {
			{Name: "includes", Value: "**/*.d **/*.di"},
		},
		//ToDo: rename to gomod?
		"golang": {
			{Name: "go.resolveDependencies", Value: true, Force: true},
			{Name: "go.ignoreSourceFiles", Value: true, Force: true},
			{Name: "go.collectDependenciesAtRuntime", Value: false},
			{Name: "go.dependencyManager", Value: "dep"},
			{Name: "includes", Value: "**/*.lock"},
			{Name: "excludes", Value: "**/*sources.jar **/*javadoc.jar"},
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

	log.Entry().Infof("Configuration for buildTool: '%v' is not yet hardened, please do a quality assessment of your scan results.", buildTool)
	return fmt.Errorf("configuration not hardened")
}
