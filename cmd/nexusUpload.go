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
	err := uploader.SetRepoURL(options.Url, options.Version, options.Repository)
	if err != nil {
		return err
	}
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
	var mtaPath string
	exists, _ := utils.fileExists("mta.yaml")
	if exists {
		mtaPath = "mta.yaml"
		// Give this file precedence, but it would be even better if
		// ProjectStructure could be asked for the mta file it detected.
	} else {
		// This will fail anyway if the file doesn't exist
		mtaPath = "mta.yml"
	}
	version, err := getVersionFromMtaFile(utils, mtaPath)
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
		err = uploader.SetInfo(options.GroupID, artifactID, version)
		if err == nexus.ErrEmptyVersion {
			err = fmt.Errorf("the project descriptor file 'mta.yaml' has an invalid version: %w", err)
		}
	}
	if err == nil {
		err = addArtifact(utils, uploader, mtaPath, "", "yaml")
	}
	if err == nil {
		mtarFilePath := utils.getEnvParameter(".pipeline/commonPipelineEnvironment", "mtarFilePath")
		log.Entry().Debugf("mtar file path: '%s'", mtarFilePath)
		err = addArtifact(utils, uploader, mtarFilePath, "", "mtar")
	}
	if err == nil {
		err = uploadArtifacts(utils, uploader, options, false)
	}
	return err
}

type mtaYaml struct {
	ID      string `json:"ID"`
	Version string `json:"version"`
}

func getVersionFromMtaFile(utils nexusUploadUtils, filePath string) (string, error) {
	mtaYamlContent, err := utils.fileRead(filePath)
	if err != nil {
		return "", fmt.Errorf("could not read from required project descriptor file '%s'",
			filePath)
	}
	return getVersionFromMtaYaml(mtaYamlContent, filePath)
}

func getVersionFromMtaYaml(mtaYamlContent []byte, filePath string) (string, error) {
	var mtaYaml mtaYaml
	err := yaml.Unmarshal(mtaYamlContent, &mtaYaml)
	if err != nil {
		// Eat the original error as it is unhelpful and confusingly mentions JSON, while the
		// user thinks it should parse YAML (it is transposed by the implementation).
		return "", fmt.Errorf("failed to parse contents of the project descriptor file '%s'",
			filePath)
	}
	return mtaYaml.Version, nil
}

func createMavenExecuteOptions(options *nexusUploadOptions) maven.ExecuteOptions {
	mavenOptions := maven.ExecuteOptions{
		ReturnStdout:       false,
		M2Path:             options.M2Path,
		GlobalSettingsFile: options.GlobalSettingsFile,
	}
	return mavenOptions
}

const settingsServerID = "artifact.deployment.nexus"

const nexusMavenSettings = `<settings xmlns="http://maven.apache.org/SETTINGS/1.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://maven.apache.org/SETTINGS/1.0.0 http://maven.apache.org/xsd/settings-1.0.0.xsd">
	<servers>
		<server>
			<id>artifact.deployment.nexus</id>
			<username>${env.NEXUS_username}</username>
			<password>${env.NEXUS_password}</password>
		</server>
	</servers>
</settings>
`

const settingsPath = ".pipeline/nexusMavenSettings.xml"

func setupNexusCredentialsSettingsFile(utils nexusUploadUtils, options *nexusUploadOptions,
	mavenOptions *maven.ExecuteOptions) (string, error) {
	if options.User == "" || options.Password == "" {
		return "", nil
	}

	err := utils.fileWrite(settingsPath, []byte(nexusMavenSettings), os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to write maven settings file to '%s': %w", settingsPath, err)
	}

	log.Entry().Debugf("Writing nexus credentials to environment")
	utils.getExecRunner().SetEnv([]string{"NEXUS_username=" + options.User, "NEXUS_password=" + options.Password})

	mavenOptions.ProjectSettingsFile = settingsPath
	mavenOptions.Defines = append(mavenOptions.Defines, "-DrepositoryId="+settingsServerID)
	return settingsPath, nil
}

type artifactDefines struct {
	file        string
	packaging   string
	files       string
	classifiers string
	types       string
}

const deployGoal = "org.apache.maven.plugins:maven-deploy-plugin:2.7:deploy-file"

func uploadArtifacts(utils nexusUploadUtils, uploader nexus.Uploader, options *nexusUploadOptions,
	generatePOM bool) error {
	if uploader.GetGroupID() == "" {
		return fmt.Errorf("no group ID was provided, or could be established from project files")
	}

	artifacts := uploader.GetArtifacts()
	if len(artifacts) == 0 {
		return errors.New("no artifacts to upload")
	}

	var defines []string
	defines = append(defines, "-Durl=http://"+uploader.GetRepoURL())
	defines = append(defines, "-DgroupId="+uploader.GetGroupID())
	defines = append(defines, "-Dversion="+uploader.GetArtifactsVersion())
	defines = append(defines, "-DartifactId="+uploader.GetArtifactsID())

	mavenOptions := createMavenExecuteOptions(options)
	mavenOptions.Goals = []string{deployGoal}
	mavenOptions.Defines = defines

	settingsFile, err := setupNexusCredentialsSettingsFile(utils, options, &mavenOptions)
	if err != nil {
		return fmt.Errorf("writing credential settings for maven failed: %w", err)
	}
	if settingsFile != "" {
		defer utils.fileRemove(settingsFile)
	}

	// iterate over the artifact descriptions, the first one is the main artifact, the following ones are
	// sub-artifacts.
	var d artifactDefines
	for i, artifact := range artifacts {
		if i == 0 {
			d.file = artifact.File
			d.packaging = artifact.Type
		} else {
			// Note: It is important to append the comma, even when the list is empty
			// or the appended item is empty. So classifiers could end up like ",,classes".
			// This is needed to match the third classifier "classes" to the third sub-artifact.
			d.files = appendItemToString(d.files, artifact.File, i == 1)
			d.classifiers = appendItemToString(d.classifiers, artifact.Classifier, i == 1)
			d.types = appendItemToString(d.types, artifact.Type, i == 1)
		}
	}

	err = uploadArtifactsBundle(d, generatePOM, mavenOptions, utils.getExecRunner())
	if err != nil {
		return fmt.Errorf("uploading artifacts for ID '%s' failed: %w", uploader.GetArtifactsID(), err)
	}
	uploader.Clear()
	return nil
}

// appendItemToString appends a comma this is not the first item, regardless of whether
// list or item are empty.
func appendItemToString(list, item string, first bool) string {
	if !first {
		list += ","
	}
	return list + item
}

func uploadArtifactsBundle(d artifactDefines, generatePOM bool, mavenOptions maven.ExecuteOptions,
	execRunner execRunner) error {
	if d.file == "" {
		return fmt.Errorf("no file specified")
	}

	var defines []string

	defines = append(defines, "-Dfile="+d.file)
	defines = append(defines, "-Dpackaging="+d.packaging)
	if !generatePOM {
		defines = append(defines, "-DgeneratePom=false")
	}

	if len(d.files) > 0 {
		defines = append(defines, "-Dfiles="+d.files)
		defines = append(defines, "-Dclassifiers="+d.classifiers)
		defines = append(defines, "-Dtypes="+d.types)
	}

	mavenOptions.Defines = append(mavenOptions.Defines, defines...)
	_, err := maven.Execute(&mavenOptions, execRunner)
	return err
}

func addArtifact(utils nexusUploadUtils, uploader nexus.Uploader, filePath, classifier, fileType string) error {
	exists, _ := utils.fileExists(filePath)
	if !exists {
		return fmt.Errorf("artifact file not found '%s'", filePath)
	}
	artifact := nexus.ArtifactDescription{
		File:       filePath,
		Type:       fileType,
		Classifier: classifier,
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
		err = nil
	}
	return err
}

func uploadMavenArtifacts(utils nexusUploadUtils, uploader nexus.Uploader, options *nexusUploadOptions,
	pomPath, targetFolder, additionalClassifiers string) error {
	pomFile := composeFilePath(pomPath, "pom", "xml")
	exists, _ := utils.fileExists(pomFile)
	if !exists {
		return errPomNotFound
	}
	groupID, _ := utils.evaluate(pomFile, "project.groupId")
	if groupID == "" {
		groupID = options.GroupID
	}
	artifactID, err := utils.evaluate(pomFile, "project.artifactId")
	var artifactsVersion string
	if err == nil {
		artifactsVersion, err = utils.evaluate(pomFile, "project.version")
	}
	if err == nil {
		err = uploader.SetInfo(groupID, artifactID, artifactsVersion)
	}
	var finalBuildName string
	if err == nil {
		finalBuildName, _ = utils.evaluate(pomFile, "project.build.finalName")
		if finalBuildName == "" {
			// Fallback to using artifactID as base-name of artifact files.
			finalBuildName = artifactID
		}
	}
	if err == nil {
		err = addArtifact(utils, uploader, pomFile, "", "pom")
	}
	if err == nil {
		err = addMavenTargetArtifact(utils, uploader, pomFile, targetFolder, finalBuildName)
	}
	if err == nil {
		err = addMavenTargetSubArtifacts(utils, uploader, additionalClassifiers, targetFolder, finalBuildName)
	}
	if err == nil {
		err = uploadArtifacts(utils, uploader, options, true)
	}
	return err
}

func addMavenTargetArtifact(utils nexusUploadUtils, uploader nexus.Uploader,
	pomFile, targetFolder, finalBuildName string) error {
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
	filePath := composeFilePath(targetFolder, finalBuildName, packaging)
	return addArtifact(utils, uploader, filePath, "", packaging)
}

func addMavenTargetSubArtifacts(utils nexusUploadUtils, uploader nexus.Uploader,
	additionalClassifiers, targetFolder, finalBuildName string) error {
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
		filePath := composeFilePath(targetFolder, finalBuildName+"-"+classifier.Classifier, classifier.FileType)
		err = addArtifact(utils, uploader, filePath, classifier.Classifier, classifier.FileType)
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
