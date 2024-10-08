package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/pkg/errors"

	b64 "encoding/base64"

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
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	SetEnv(env []string)
	RunExecutable(e string, p ...string) error

	FileExists(path string) (bool, error)
	FileRead(path string) ([]byte, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
	FileRemove(path string) error
	DirExists(path string) (bool, error)
	Glob(pattern string) (matches []string, err error)
	Copy(src, dest string) (int64, error)
	MkdirAll(path string, perm os.FileMode) error

	DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error

	UsesMta() bool
	UsesMaven() bool
	UsesNpm() bool

	getEnvParameter(path, name string) string
	evaluate(options *maven.EvaluateOptions, expression string) (string, error)
}

type utilsBundle struct {
	*piperutils.ProjectStructure
	*piperutils.Files
	*command.Command
	*piperhttp.Client
}

func newUtilsBundle() *utilsBundle {
	utils := utilsBundle{
		ProjectStructure: &piperutils.ProjectStructure{},
		Files:            &piperutils.Files{},
		Command:          &command.Command{},
		Client:           &piperhttp.Client{},
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func (u *utilsBundle) FileWrite(filePath string, content []byte, perm os.FileMode) error {
	parent := filepath.Dir(filePath)
	if parent != "" {
		err := u.Files.MkdirAll(parent, 0775)
		if err != nil {
			return err
		}
	}
	return u.Files.FileWrite(filePath, content, perm)
}

func (u *utilsBundle) getEnvParameter(path, name string) string {
	return piperenv.GetParameter(path, name)
}

func (u *utilsBundle) evaluate(options *maven.EvaluateOptions, expression string) (string, error) {
	return maven.Evaluate(options, expression, u)
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
	performMavenUpload := len(options.MavenRepository) > 0
	performNpmUpload := len(options.NpmRepository) > 0

	if !performMavenUpload && !performNpmUpload {
		if options.Format == "" {
			return fmt.Errorf("none of the parameters 'mavenRepository' and 'npmRepository' are configured, or 'format' should be set if the 'url' already contains the repository ID")
		}
		if options.Format == "maven" {
			performMavenUpload = true
		} else if options.Format == "npm" {
			performNpmUpload = true
		}
	}

	err := uploader.SetRepoURL(options.Url, options.Version, options.MavenRepository, options.NpmRepository)
	if err != nil {
		return err
	}

	if utils.UsesNpm() && performNpmUpload {
		log.Entry().Info("NPM project structure detected")
		err = uploadNpmArtifacts(utils, uploader, options)
	} else {
		log.Entry().Info("Skipping npm upload because either no package json was found or NpmRepository option is not provided.")
	}
	if err != nil {
		return err
	}

	if performMavenUpload {
		if utils.UsesMta() {
			log.Entry().Info("MTA project structure detected")
			return uploadMTA(utils, uploader, options)
		} else if utils.UsesMaven() {
			log.Entry().Info("Maven project structure detected")
			return uploadMaven(utils, uploader, options)
		}
	} else {
		log.Entry().Info("Skipping maven and mta upload because mavenRepository option is not provided.")
	}

	return nil
}

func uploadNpmArtifacts(utils nexusUploadUtils, uploader nexus.Uploader, options *nexusUploadOptions) error {
	environment := []string{"npm_config_registry=" + uploader.GetNexusURLProtocol() + "://" + uploader.GetNpmRepoURL(), "npm_config_email=project-piper@no-reply.com"}
	if options.Username != "" && options.Password != "" {
		auth := b64.StdEncoding.EncodeToString([]byte(options.Username + ":" + options.Password))
		environment = append(environment, "npm_config__auth="+auth)
	} else {
		log.Entry().Info("No credentials provided for npm upload, trying to upload anonymously.")
	}
	utils.SetEnv(environment)
	err := utils.RunExecutable("npm", "publish")
	return err
}

func uploadMTA(utils nexusUploadUtils, uploader nexus.Uploader, options *nexusUploadOptions) error {
	if options.GroupID == "" {
		return fmt.Errorf("the 'groupId' parameter needs to be provided for MTA projects")
	}
	var mtaPath string
	exists, _ := utils.FileExists("mta.yaml")
	if exists {
		mtaPath = "mta.yaml"
		// Give this file precedence, but it would be even better if
		// ProjectStructure could be asked for the mta file it detected.
	} else {
		// This will fail anyway if the file doesn't exist
		mtaPath = "mta.yml"
	}
	mtaInfo, err := getInfoFromMtaFile(utils, mtaPath)
	if err == nil {
		if options.ArtifactID != "" {
			mtaInfo.ID = options.ArtifactID
		}
		err = uploader.SetInfo(options.GroupID, mtaInfo.ID, mtaInfo.Version)
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

func getInfoFromMtaFile(utils nexusUploadUtils, filePath string) (*mtaYaml, error) {
	mtaYamlContent, err := utils.FileRead(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read from required project descriptor file '%s'",
			filePath)
	}
	return getInfoFromMtaYaml(mtaYamlContent, filePath)
}

func getInfoFromMtaYaml(mtaYamlContent []byte, filePath string) (*mtaYaml, error) {
	var mtaYaml mtaYaml
	err := yaml.Unmarshal(mtaYamlContent, &mtaYaml)
	if err != nil {
		// Eat the original error as it is unhelpful and confusingly mentions JSON, while the
		// user thinks it should parse YAML (it is transposed by the implementation).
		return nil, fmt.Errorf("failed to parse contents of the project descriptor file '%s'",
			filePath)
	}
	return &mtaYaml, nil
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
	if options.Username == "" || options.Password == "" {
		return "", nil
	}

	err := utils.FileWrite(settingsPath, []byte(nexusMavenSettings), os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to write maven settings file to '%s': %w", settingsPath, err)
	}

	log.Entry().Debugf("Writing nexus credentials to environment")
	utils.SetEnv([]string{"NEXUS_username=" + options.Username, "NEXUS_password=" + options.Password})

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

const deployGoal = "org.apache.maven.plugins:maven-deploy-plugin:2.8.2:deploy-file"

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
	defines = append(defines, "-Durl="+uploader.GetNexusURLProtocol()+"://"+uploader.GetMavenRepoURL())
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
		defer func() { _ = utils.FileRemove(settingsFile) }()
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

	err = uploadArtifactsBundle(d, generatePOM, mavenOptions, utils)
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
	utils nexusUploadUtils) error {
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
	_, err := maven.Execute(&mavenOptions, utils)
	return err
}

func addArtifact(utils nexusUploadUtils, uploader nexus.Uploader, filePath, classifier, fileType string) error {
	exists, _ := utils.FileExists(filePath)
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
	pomFiles, _ := utils.Glob("**/pom.xml")
	if len(pomFiles) == 0 {
		return errPomNotFound
	}

	for _, pomFile := range pomFiles {
		parentDir := filepath.Dir(pomFile)
		if parentDir == "integration-tests" || parentDir == "unit-tests" {
			continue
		}
		err := uploadMavenArtifacts(utils, uploader, options, parentDir, filepath.Join(parentDir, "target"))
		if err != nil {
			return err
		}
	}
	return nil
}

func uploadMavenArtifacts(utils nexusUploadUtils, uploader nexus.Uploader, options *nexusUploadOptions,
	pomPath, targetFolder string) error {
	pomFile := composeFilePath(pomPath, "pom", "xml")

	evaluateOptions := &maven.EvaluateOptions{
		PomPath:            pomFile,
		GlobalSettingsFile: options.GlobalSettingsFile,
		M2Path:             options.M2Path,
	}

	packaging, _ := utils.evaluate(evaluateOptions, "project.packaging")
	if packaging == "" {
		packaging = "jar"
	}
	if packaging != "pom" {
		// Ignore this module if there is no 'target' folder
		hasTarget, _ := utils.DirExists(targetFolder)
		if !hasTarget {
			log.Entry().Warnf("Ignoring module '%s' as it has no 'target' folder", pomPath)
			return nil
		}
	}
	groupID, _ := utils.evaluate(evaluateOptions, "project.groupId")
	if groupID == "" {
		groupID = options.GroupID
	}
	artifactID, err := utils.evaluate(evaluateOptions, "project.artifactId")
	var artifactsVersion string
	if err == nil {
		artifactsVersion, err = utils.evaluate(evaluateOptions, "project.version")
	}
	if err == nil {
		err = uploader.SetInfo(groupID, artifactID, artifactsVersion)
	}
	var finalBuildName string
	if err == nil {
		finalBuildName, _ = utils.evaluate(evaluateOptions, "project.build.finalName")
		if finalBuildName == "" {
			// Fallback to composing final build name, see http://maven.apache.org/pom.html#BaseBuild_Element
			finalBuildName = artifactID + "-" + artifactsVersion
		}
	}
	if err == nil {
		err = addArtifact(utils, uploader, pomFile, "", "pom")
	}
	if err == nil && packaging != "pom" {
		err = addMavenTargetArtifacts(utils, uploader, pomFile, targetFolder, finalBuildName, packaging)
	}
	if err == nil {
		err = uploadArtifacts(utils, uploader, options, true)
	}
	return err
}

func addMavenTargetArtifacts(utils nexusUploadUtils, uploader nexus.Uploader, pomFile, targetFolder, finalBuildName, packaging string) error {
	fileTypes := []string{packaging}
	if packaging != "jar" {
		// Try to find additional artifacts with a classifier
		fileTypes = append(fileTypes, "jar")
	}

	for _, fileType := range fileTypes {
		pattern := targetFolder + "/*." + fileType
		matches, _ := utils.Glob(pattern)
		if len(matches) == 0 && fileType == packaging {
			return fmt.Errorf("target artifact not found for packaging '%s'", packaging)
		}
		log.Entry().Debugf("Glob matches for %s: %s", pattern, strings.Join(matches, ", "))

		prefix := filepath.Join(targetFolder, finalBuildName) + "-"
		suffix := "." + fileType
		for _, filename := range matches {
			classifier := ""
			temp := filename
			if strings.HasPrefix(temp, prefix) && strings.HasSuffix(temp, suffix) {
				temp = strings.TrimPrefix(temp, prefix)
				temp = strings.TrimSuffix(temp, suffix)
				classifier = temp
			}
			err := addArtifact(utils, uploader, filename, classifier, fileType)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func composeFilePath(folder, name, extension string) string {
	fileName := name + "." + extension
	return filepath.Join(folder, fileName)
}
