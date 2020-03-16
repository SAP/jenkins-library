package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/nexus"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/ghodss/yaml"
)

// nexusUploadUtils defines an interface for utility functionality used from external packages,
// so it can be easily mocked for testing.
type nexusUploadUtils interface {
	fileExists(path string) (bool, error)
	fileRead(path string) ([]byte, error)
	fileWrite(path string, content []byte, perm os.FileMode) error
	fileRemove(path string)
	usesMta() bool
	usesMaven() bool
	getEnvParameter(path, name string) string
	getExecRunner() execRunner
	evaluate(pomFile, expression string) (string, error)
}

type utilsBundle struct {
	projectStructure piperutils.ProjectStructure
	fileUtils        piperutils.Files
	execRunner       *command.Command
}

func newUtilsBundle() *utilsBundle {
	return &utilsBundle{
		projectStructure: piperutils.ProjectStructure{},
		fileUtils:        piperutils.Files{},
	}
}

func (u *utilsBundle) fileExists(path string) (bool, error) {
	return u.fileUtils.FileExists(path)
}

func (u *utilsBundle) fileRead(path string) ([]byte, error) {
	return u.fileUtils.FileRead(path)
}

func (u *utilsBundle) fileWrite(filePath string, content []byte, perm os.FileMode) error {
	parent := filepath.Dir(filePath)
	if parent != "" {
		err := u.fileUtils.MkdirAll(parent, 0775)
		if err != nil {
			return err
		}
	}
	return u.fileUtils.FileWrite(filePath, content, perm)
}

func (u *utilsBundle) fileRemove(path string) {
	err := os.Remove(path)
	if err != nil {
		log.Entry().WithError(err).Warnf("Failed to remove file '%s'.", path)
	} else {
		log.Entry().Infof("Remove file '%s'", path)
	}
}

func (u *utilsBundle) usesMta() bool {
	return u.projectStructure.UsesMta()
}

func (u *utilsBundle) usesMaven() bool {
	return u.projectStructure.UsesMaven()
}

func (u *utilsBundle) getEnvParameter(path, name string) string {
	return piperenv.GetParameter(path, name)
}

func (u *utilsBundle) getExecRunner() execRunner {
	if u.execRunner == nil {
		u.execRunner = &command.Command{}
		u.execRunner.Stdout(log.Entry().Writer())
		u.execRunner.Stderr(log.Entry().Writer())
	}
	return u.execRunner
}

func (u *utilsBundle) evaluate(pomFile, expression string) (string, error) {
	return maven.Evaluate(pomFile, expression, u.getExecRunner())
}

func nexusUpload(options nexusUploadOptions, _ *telemetry.CustomData) {
	utils := newUtilsBundle()
	uploader := nexus.Upload{}

	err := runNexusUpload(utils, &uploader, &options)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runNexusUpload(utils nexusUploadUtils, uploader nexus.Uploader, options *nexusUploadOptions) error {
	if utils.usesMta() {
		log.Entry().Info("MTA project structure detected")
		return uploadMTA(utils, uploader, options)
	} else if utils.usesMaven() {
		log.Entry().Info("Maven project structure detected")
		return uploadMaven(utils, uploader, options)
	} else {
		return fmt.Errorf("unsupported project structure")
	}
}

func uploadMTA(utils nexusUploadUtils, uploader nexus.Uploader, options *nexusUploadOptions) error {
	if options.GroupID == "" {
		return fmt.Errorf("the 'groupId' parameter needs to be provided for MTA projects")
	}
	err := uploader.SetBaseURL(options.Url, options.Version, options.Repository)
	var mtaPath string
	if err == nil {
		exists, _ := utils.fileExists("mta.yaml")
		if exists {
			mtaPath = "mta.yaml"
			// Give this file precedence, but it would be even better if
			// ProjectStructure could be asked for the mta file it detected.
		} else {
			// This will fail anyway if the file doesn't exist
			mtaPath = "mta.yml"
		}
		err = setVersionFromMtaFile(utils, uploader, mtaPath)
	}
	var artifactID = options.ArtifactID
	if artifactID == "" {
		artifactID = utils.getEnvParameter(".pipeline/commonPipelineEnvironment/configuration", "artifactId")
		if artifactID == "" {
			err = fmt.Errorf("the 'artifactId' parameter was not provided and could not be retrieved from the Common Pipeline Environment")
		} else {
			log.Entry().Debugf("mtar artifact id from CPE: '%s'", artifactID)
		}
	}
	if err == nil {
		err = addArtifact(utils, uploader, mtaPath, "", "yaml", artifactID)
	}
	if err == nil {
		mtarFilePath := utils.getEnvParameter(".pipeline/commonPipelineEnvironment", "mtarFilePath")
		log.Entry().Debugf("mtar file path: '%s'", mtarFilePath)
		err = addArtifact(utils, uploader, mtarFilePath, "", "mtar", artifactID)
	}
	if err == nil {
		err = uploadArtifacts(utils, uploader, options, options.GroupID)
	}
	return err
}

type mtaYaml struct {
	ID      string `json:"ID"`
	Version string `json:"version"`
}

func setVersionFromMtaFile(utils nexusUploadUtils, uploader nexus.Uploader, filePath string) error {
	mtaYamlContent, err := utils.fileRead(filePath)
	if err != nil {
		return fmt.Errorf("could not read from required project descriptor file '%s'",
			filePath)
	}
	return setVersionFromMtaYaml(uploader, mtaYamlContent, filePath)
}

func setVersionFromMtaYaml(uploader nexus.Uploader, mtaYamlContent []byte, filePath string) error {
	var mtaYaml mtaYaml
	err := yaml.Unmarshal(mtaYamlContent, &mtaYaml)
	if err != nil {
		// Eat the original error as it is unhelpful and confusingly mentions JSON, while the
		// user thinks it should parse YAML (it is transposed by the implementation).
		return fmt.Errorf("failed to parse contents of the project descriptor file '%s'",
			filePath)
	}
	err = uploader.SetArtifactsVersion(mtaYaml.Version)
	if err != nil {
		return fmt.Errorf("the project descriptor file '%s' has an invalid version: %w",
			filePath, err)
	}
	return nil
}

func createMavenExecuteOptions(options *nexusUploadOptions) maven.ExecuteOptions {
	mavenOptions := maven.ExecuteOptions{
		ReturnStdout:       false,
		M2Path:             options.M2Path,
		GlobalSettingsFile: options.GlobalSettingsFile,
	}
	return mavenOptions
}

var settingsServerId = "artifact.deployment.nexus"
var nexusMavenSettings = `<settings xmlns="http://maven.apache.org/SETTINGS/1.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://maven.apache.org/SETTINGS/1.0.0 http://maven.apache.org/xsd/settings-1.0.0.xsd">
    <servers>
        <server>
            <id>artifact.deployment.nexus</id>
            <username>${repo.username}</username>
            <password>${repo.password}</password>
        </server>
    </servers>
</settings>
`

func setupNexusCredentialsSettingsFile(utils nexusUploadUtils, options *nexusUploadOptions,
	mavenOptions *maven.ExecuteOptions, execRunner execRunner) (string, error) {
	if options.User == "" || options.Password == "" {
		return "", nil
	}

	path := ".pipeline/nexusMavenSettings.xml"
	err := utils.fileWrite(path, []byte(nexusMavenSettings), os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to write maven settings file to '%s': %w", path, err)
	}

	log.Entry().Debugf("Writing nexus credentials to environment")
	log.Entry().Infof("Wrote maven settings to '%s", path)

	execRunner.SetEnv([]string{"NEXUS_username=" + options.User, "NEXUS_password=" + options.Password})

	mavenOptions.ProjectSettingsFile = path
	//	mavenOptions.Defines = append(mavenOptions.Defines, "-Drepo.username=$NEXUS_username")
	//	mavenOptions.Defines = append(mavenOptions.Defines, "-Drepo.password=$NEXUS_password")
	mavenOptions.Defines = append(mavenOptions.Defines, "-Drepo.username="+options.User)
	mavenOptions.Defines = append(mavenOptions.Defines, "-Drepo.password="+options.Password)
	return path, nil
}

func uploadArtifacts(utils nexusUploadUtils, uploader nexus.Uploader, options *nexusUploadOptions, groupID string) error {
	if groupID == "" {
		return fmt.Errorf("no group ID was provided, or could be established from project files")
	}

	artifacts := uploader.GetArtifacts()
	if len(artifacts) == 0 {
		return errors.New("no artifacts to upload")
	}

	var defines []string
	defines = append(defines)
	defines = append(defines, "-Durl=http://"+uploader.GetBaseURL())
	defines = append(defines, "-DgroupId="+groupID)
	defines = append(defines, "-Dversion="+uploader.GetArtifactsVersion())

	mavenOptions := createMavenExecuteOptions(options)
	mavenOptions.Goals = []string{"deploy:deploy-file"}
	mavenOptions.Defines = defines

	execRunner := utils.getExecRunner()
	settingsFile, err := setupNexusCredentialsSettingsFile(utils, options, &mavenOptions, execRunner)
	if err != nil {
		return fmt.Errorf("writing credential settings for maven failed: %w", err)
	}
	if settingsFile != "" {
		mavenOptions.Defines = append(mavenOptions.Defines, "-DrepositoryId="+settingsServerId)
		defer utils.fileRemove(settingsFile)
	}

	// iterate over the artifact descriptions and upload those with the same ID in one bundle
	artifactID := ""
	file := ""
	files := ""
	classifiers := ""
	types := ""

	for _, artifact := range artifacts {
		if artifactID != artifact.ID {
			if artifactID != "" {
				err := uploadArtifactsBundle(artifactID, file, files, classifiers, types, mavenOptions, execRunner)
				if err != nil {
					return err
				}
			}
			artifactID = artifact.ID
			file = ""
			files = ""
			classifiers = ""
			types = ""
		}
		if file == "" {
			file = artifact.File
		} else {
			if files != "" {
				files += ","
				classifiers += ","
				types += ","
			}
			files += artifact.File
			classifiers += artifact.Classifier
			types += artifact.Type
		}
	}

	if file != "" {
		return uploadArtifactsBundle(artifactID, file, files, classifiers, types, mavenOptions, execRunner)
	}
	return nil
}

func uploadArtifactsBundle(artifactID, file, files, classifiers, types string,
	mavenOptions maven.ExecuteOptions, execRunner execRunner) error {
	if artifactID == "" {
		return fmt.Errorf("no artifact ID specified")
	}
	if file == "" {
		return fmt.Errorf("no file specified")
	}

	var defines []string

	defines = append(defines, "-DartifactId="+artifactID)
	defines = append(defines, "-Dfile="+file)
	defines = append(defines, "-DgeneratePom=false")

	if len(files) > 0 {
		defines = append(defines, "-Dfiles="+files)
		defines = append(defines, "-Dclassifiers="+classifiers)
		defines = append(defines, "-Dtypes="+types)
	}

	mavenOptions.Defines = append(mavenOptions.Defines, defines...)

	_, err := maven.Execute(&mavenOptions, execRunner)
	if err != nil {
		return fmt.Errorf("uploading artifacts for ID '%s' failed: %w", artifactID, err)
	}
	return nil
}

/*
func uploadMaven(utils nexusUploadUtils, uploader nexus.Uploader, options *nexusUploadOptions) error {
	if pomExists, _ := utils.fileExists("pom.xml"); !pomExists {
		return errors.New("pom.xml not found")
	}

	err := uploader.SetBaseURL(options.Url, options.Version, options.Repository)
	if err != nil {
		return err
	}

	// Maven will look up 'settingsServerId' in the local settings file to find login credentials
	altRepository := settingsServerId + "::default::http://" + uploader.GetBaseURL()

	var defines []string
	defines = append(defines, "-Dmaven.test.skip")
	defines = append(defines, "-DaltDeploymentRepository="+altRepository)

	testModulesExcludes := maven.GetTestModulesExcludes()
	if testModulesExcludes != nil {
		defines = append(defines, testModulesExcludes...)
	}

	mavenOptions := createMavenExecuteOptions(options)
	mavenOptions.Goals = []string{"deploy"}
	mavenOptions.Defines = defines

	execRunner := utils.getExecRunner()
	settingsFile, err := setupNexusCredentialsSettingsFile(utils, options, &mavenOptions, execRunner)
	if err != nil {
		return fmt.Errorf("writing credential settings for maven failed: %w", err)
	}
	if settingsFile != "" {
		defer utils.fileRemove(settingsFile)
	}

	_, err = maven.Execute(&mavenOptions, execRunner)
	if err != nil {
		return err
	}
	return nil
}
*/

func addArtifact(utils nexusUploadUtils, uploader nexus.Uploader, filePath, classifier, fileType, id string) error {
	exists, _ := utils.fileExists(filePath)
	if !exists {
		return fmt.Errorf("artifact file not found '%s'", filePath)
	}
	artifact := nexus.ArtifactDescription{
		File:       filePath,
		Type:       fileType,
		Classifier: classifier,
		ID:         id,
	}
	return uploader.AddArtifact(artifact)
}

var errPomNotFound = errors.New("pom.xml not found")

func uploadMaven(utils nexusUploadUtils, uploader nexus.Uploader, options *nexusUploadOptions) error {
	err := uploadMavenArtifacts(utils, uploader, options, "", "target", "")
	if err != nil {
		return err
	}

	// Test if a sub-folder "application" exists and upload the artifacts from there as well.
	// This means there are built-in assumptions about the project structure (archetype),
	// that nexusUpload supports. To make this more flexible should be the scope of another PR.
	err = uploadMavenArtifacts(utils, uploader, options, "application", "application/target",
		options.AdditionalClassifiers)
	if err == errPomNotFound {
		// Ignore for missing application module
		return nil
	}
	return err
}

func uploadMavenArtifacts(utils nexusUploadUtils, uploader nexus.Uploader, options *nexusUploadOptions,
	pomPath, targetFolder, additionalClassifiers string) error {
	var err error

	pomFile := composeFilePath(pomPath, "pom", "xml")
	exists, _ := utils.fileExists(pomFile)
	if !exists {
		return errPomNotFound
	}
	groupID, err := utils.evaluate(pomFile, "project.groupId")
	if groupID == "" {
		groupID = options.GroupID
		// Reset error
		err = nil
	}
	if err == nil {
		err = uploader.SetBaseURL(options.Url, options.Version, options.Repository)
	}
	var artifactID string
	if err == nil {
		artifactID, err = utils.evaluate(pomFile, "project.artifactId")
	}
	var artifactsVersion string
	if err == nil {
		artifactsVersion, err = utils.evaluate(pomFile, "project.version")
	}
	if err == nil {
		err = uploader.SetArtifactsVersion(artifactsVersion)
	}
	if err == nil {
		err = addArtifact(utils, uploader, pomFile, "", "pom", artifactID)
	}
	if err == nil {
		err = addTargetArtifact(utils, uploader, pomFile, targetFolder, artifactID)
	}
	if err == nil {
		err = addAdditionalClassifierArtifacts(utils, uploader, additionalClassifiers, targetFolder, artifactID)
	}
	if err == nil {
		err = uploadArtifacts(utils, uploader, options, groupID)
	}
	return err
}

func addTargetArtifact(utils nexusUploadUtils, uploader nexus.Uploader, pomFile, targetFolder, artifactID string) error {
	packaging, err := utils.evaluate(pomFile, "project.packaging")
	if err != nil {
		return err
	}
	if packaging == "pom" {
		// Only pom.xml itself is the artifact
		return nil
	}
	if packaging == "" {
		packaging = "jar"
	}
	finalName, err := utils.evaluate(pomFile, "project.build.finalName")
	if err != nil || finalName == "" {
		// NOTE: The error should be ignored, and the finalName built as Maven would from artifactId and so on.
		// But it seems this expression always resolves, even if finalName is nowhere declared in the pom.xml
		return err
	}
	filePath := composeFilePath(targetFolder, finalName, packaging)
	return addArtifact(utils, uploader, filePath, "", packaging, artifactID)
}

func addAdditionalClassifierArtifacts(utils nexusUploadUtils, uploader nexus.Uploader,
	additionalClassifiers, targetFolder, artifactID string) error {
	if additionalClassifiers == "" {
		return nil
	}
	classifiers, err := getClassifiers(additionalClassifiers)
	if err != nil {
		return err
	}
	for _, classifier := range classifiers {
		if classifier.Classifier == "" || classifier.FileType == "" {
			return fmt.Errorf("invalid additional classifier description (classifier: '%s', type: '%s')",
				classifier.Classifier, classifier.FileType)
		}
		filePath := composeFilePath(targetFolder, artifactID+"-"+classifier.Classifier, classifier.FileType)
		err = addArtifact(utils, uploader, filePath, classifier.FileType, classifier.Classifier, artifactID)
		if err != nil {
			return err
		}
	}
	return nil
}

func composeFilePath(folder, name, extension string) string {
	fileName := name + "." + extension
	return filepath.Join(folder, fileName)
}

type classifierDescription struct {
	Classifier string `json:"classifier"`
	FileType   string `json:"type"`
}

func getClassifiers(classifiersAsJSON string) ([]classifierDescription, error) {
	var classifiers []classifierDescription
	err := json.Unmarshal([]byte(classifiersAsJSON), &classifiers)
	return classifiers, err
}
