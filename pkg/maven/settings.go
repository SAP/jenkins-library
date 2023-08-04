package maven

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
)

var getenv = os.Getenv

// SettingsDownloadUtils defines an interface for downloading and storing maven settings files.
type SettingsDownloadUtils interface {
	DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error
	FileExists(filename string) (bool, error)
	Copy(src, dest string) (int64, error)
	MkdirAll(path string, perm os.FileMode) error
	FileWrite(path string, content []byte, perm os.FileMode) error
	FileRead(path string) ([]byte, error)
}

// DownloadAndGetMavenParameters downloads the global or project settings file if the strings contain URLs.
// It then constructs the arguments that need to be passed to maven in order to point to use these settings files.
func DownloadAndGetMavenParameters(globalSettingsFile string, projectSettingsFile string, utils SettingsDownloadUtils) ([]string, error) {
	mavenArgs := []string{}
	if len(globalSettingsFile) > 0 {
		globalSettingsFileName, err := downloadSettingsIfURL(globalSettingsFile, ".pipeline/mavenGlobalSettings.xml", utils, false)
		if err != nil {
			return nil, err
		}
		mavenArgs = append(mavenArgs, "--global-settings", globalSettingsFileName)
	} else {

		log.Entry().Debugf("Global settings file not provided via configuration.")
	}

	if len(projectSettingsFile) > 0 {
		projectSettingsFileName, err := downloadSettingsIfURL(projectSettingsFile, ".pipeline/mavenProjectSettings.xml", utils, false)
		if err != nil {
			return nil, err
		}
		mavenArgs = append(mavenArgs, "--settings", projectSettingsFileName)
	} else {

		log.Entry().Debugf("Project settings file not provided via configuration.")
	}
	return mavenArgs, nil
}

// DownloadAndCopySettingsFiles downloads the global or project settings file if the strings contain URLs.
// It copies the given files to either the locations specified in the environment variables M2_HOME and HOME
// or the default locations where maven expects them.
func DownloadAndCopySettingsFiles(globalSettingsFile string, projectSettingsFile string, utils SettingsDownloadUtils) error {
	if len(projectSettingsFile) > 0 {
		destination, err := getProjectSettingsFileDest()
		if err != nil {
			return err
		}

		if err := downloadAndCopySettingsFile(projectSettingsFile, destination, utils); err != nil {
			return err
		}
	} else {

		log.Entry().Debugf("Project settings file not provided via configuration.")
	}

	if len(globalSettingsFile) > 0 {
		destination, err := getGlobalSettingsFileDest()
		if err != nil {
			return err
		}
		if err := downloadAndCopySettingsFile(globalSettingsFile, destination, utils); err != nil {
			return err
		}
	} else {

		log.Entry().Debugf("Global settings file not provided via configuration.")
	}

	return nil
}

func UpdateActiveProfileInSettingsXML(newActiveProfiles []string, utils SettingsDownloadUtils) error {
	settingsFile, err := getGlobalSettingsFileDest()
	if err != nil {
		return err
	}

	settingsXMLContent, err := utils.FileRead(settingsFile)
	if err != nil {
		return fmt.Errorf("error reading global settings xml file at %v , continuing without active profile update", settingsFile)
	}

	var projectSettings Settings
	err = xml.Unmarshal([]byte(settingsXMLContent), &projectSettings)

	if err != nil {
		return fmt.Errorf("failed to unmarshal settings xml file '%v': %w", settingsFile, err)
	}

	if len(projectSettings.ActiveProfiles.ActiveProfile) == 0 {
		log.Entry().Warnf("no active profile found to replace in settings xml %v , continuing without file edit", settingsFile)
	} else {
		projectSettings.Xsi = "http://www.w3.org/2001/XMLSchema-instance"
		projectSettings.SchemaLocation = "http://maven.apache.org/SETTINGS/1.0.0 https://maven.apache.org/xsd/settings-1.0.0.xsd"

		projectSettings.ActiveProfiles.ActiveProfile = []string{}
		projectSettings.ActiveProfiles.ActiveProfile = append(projectSettings.ActiveProfiles.ActiveProfile, newActiveProfiles...)

		settingsXml, err := xml.MarshalIndent(projectSettings, "", "    ")
		if err != nil {
			return fmt.Errorf("failed to marshal maven project settings xml: %w", err)
		}

		settingsXmlString := string(settingsXml)
		Replacer := strings.NewReplacer("&#xA;", "", "&#x9;", "")
		settingsXmlString = Replacer.Replace(settingsXmlString)
		xmlstring := []byte(xml.Header + settingsXmlString)

		err = utils.FileWrite(settingsFile, xmlstring, 0777)

		if err != nil {
			return fmt.Errorf("failed to write maven Settings during <activeProfile> update xml: %w", err)
		}
		log.Entry().Infof("Successfully updated <acitveProfile> details in maven settings file : '%s'", settingsFile)

	}
	return nil
}

func CreateNewProjectSettingsXML(altDeploymentRepositoryID string, altDeploymentRepositoryUser string, altDeploymentRepositoryPassword string, utils SettingsDownloadUtils) (string, error) {
	settingsXML := Settings{
		XMLName:        xml.Name{Local: "settings"},
		Xsi:            "http://www.w3.org/2001/XMLSchema-instance",
		SchemaLocation: "http://maven.apache.org/SETTINGS/1.0.0 https://maven.apache.org/xsd/settings-1.0.0.xsd",
		Servers: ServersType{
			ServerType: []Server{
				{
					ID:       altDeploymentRepositoryID,
					Username: altDeploymentRepositoryUser,
					Password: altDeploymentRepositoryPassword,
				},
			},
		},
	}

	xmlstring, err := xml.MarshalIndent(settingsXML, "", "    ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Settings.xml: %w", err)
	}

	xmlstring = []byte(xml.Header + string(xmlstring))

	err = utils.FileWrite(".pipeline/mavenProjectSettings.xml", xmlstring, 0777)
	if err != nil {
		return "", fmt.Errorf("failed to write maven Project Settings xml: %w", err)
	}

	log.Entry().Infof("Successfully created maven project settings with <server> details at .pipeline/mavenProjectSettings.xml")

	return ".pipeline/mavenProjectSettings.xml", nil

}

func UpdateProjectSettingsXML(projectSettingsFile string, altDeploymentRepositoryID string, altDeploymentRepositoryUser string, altDeploymentRepositoryPassword string, utils SettingsDownloadUtils) (string, error) {
	projectSettingsFileDestination := ".pipeline/mavenProjectSettings"
	var err error
	if exists, _ := utils.FileExists(projectSettingsFile); exists {
		projectSettingsFileDestination = projectSettingsFile
		err = addServerTagtoProjectSettingsXML(projectSettingsFile, altDeploymentRepositoryID, altDeploymentRepositoryUser, altDeploymentRepositoryPassword, utils)
	} else {
		err = addServerTagtoProjectSettingsXML(".pipeline/mavenProjectSettings", altDeploymentRepositoryID, altDeploymentRepositoryUser, altDeploymentRepositoryPassword, utils)
	}

	if err != nil {
		return "", fmt.Errorf("failed to unmarshal settings xml file '%v': %w", projectSettingsFile, err)
	}
	return projectSettingsFileDestination, nil

}

func addServerTagtoProjectSettingsXML(projectSettingsFile string, altDeploymentRepositoryID string, altDeploymentRepositoryUser string, altDeploymentRepositoryPassword string, utils SettingsDownloadUtils) error {
	var projectSettings Settings
	settingsXMLContent, err := utils.FileRead(projectSettingsFile)
	if err != nil {
		return fmt.Errorf("failed to read file '%v': %w", projectSettingsFile, err)
	}

	err = xml.Unmarshal([]byte(settingsXMLContent), &projectSettings)
	if err != nil {
		return fmt.Errorf("failed to unmarshal settings xml file '%v': %w", projectSettingsFile, err)
	}

	if len(projectSettings.Servers.ServerType) == 0 {
		projectSettings.Xsi = "http://www.w3.org/2001/XMLSchema-instance"
		projectSettings.SchemaLocation = "http://maven.apache.org/SETTINGS/1.0.0 https://maven.apache.org/xsd/settings-1.0.0.xsd"
		projectSettings.Servers.ServerType = []Server{
			{
				ID:       altDeploymentRepositoryID,
				Username: altDeploymentRepositoryUser,
				Password: altDeploymentRepositoryPassword,
			},
		}
	} else if len(projectSettings.Servers.ServerType) > 0 { // if <server> tag is present then add the staging server tag
		stagingServer := Server{
			ID:       altDeploymentRepositoryID,
			Username: altDeploymentRepositoryUser,
			Password: altDeploymentRepositoryPassword,
		}
		projectSettings.Servers.ServerType = append(projectSettings.Servers.ServerType, stagingServer)
	}

	settingsXml, err := xml.MarshalIndent(projectSettings, "", "    ")
	if err != nil {
		fmt.Errorf("failed to marshal maven project settings xml: %w", err)
	}
	settingsXmlString := string(settingsXml)
	Replacer := strings.NewReplacer("&#xA;", "", "&#x9;", "")
	settingsXmlString = Replacer.Replace(settingsXmlString)

	xmlstring := []byte(xml.Header + settingsXmlString)

	err = utils.FileWrite(projectSettingsFile, xmlstring, 0777)
	if err != nil {
		fmt.Errorf("failed to write maven Settings xml: %w", err)
	}
	log.Entry().Infof("Successfully updated <server> details in maven project settings file : '%s'", projectSettingsFile)

	return nil
}

func downloadAndCopySettingsFile(src string, dest string, utils SettingsDownloadUtils) error {
	if len(src) == 0 {
		return fmt.Errorf("Settings file source location not provided")
	}

	if len(dest) == 0 {
		return fmt.Errorf("Settings file destination location not provided")
	}

	log.Entry().Debugf("Copying file \"%s\" to \"%s\"", src, dest)

	if strings.HasPrefix(src, "http:") || strings.HasPrefix(src, "https:") {
		err := downloadSettingsFromURL(src, dest, utils, true)
		if err != nil {
			return err
		}
	} else {

		// for sake os symmetry it would be better to use a file protocol prefix here (file:)

		parent := filepath.Dir(dest)

		parentFolderExists, err := utils.FileExists(parent)

		if err != nil {
			return err
		}

		if !parentFolderExists {
			if err = utils.MkdirAll(parent, 0775); err != nil {
				return err
			}
		}

		if _, err := utils.Copy(src, dest); err != nil {
			return err
		}
	}

	return nil
}

func downloadSettingsIfURL(settingsFileOption, settingsFile string, utils SettingsDownloadUtils, overwrite bool) (string, error) {
	result := settingsFileOption
	if strings.HasPrefix(settingsFileOption, "http:") || strings.HasPrefix(settingsFileOption, "https:") {
		err := downloadSettingsFromURL(settingsFileOption, settingsFile, utils, overwrite)
		if err != nil {
			return "", err
		}
		result = settingsFile
	}
	return result, nil
}

func downloadSettingsFromURL(url, filename string, utils SettingsDownloadUtils, overwrite bool) error {
	exists, _ := utils.FileExists(filename)
	if exists && !overwrite {
		log.Entry().Infof("Not downloading maven settings file, because it already exists at '%s'", filename)
		return nil
	}
	err := utils.DownloadFile(url, filename, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to download maven settings from URL '%s' to file '%s': %w",
			url, filename, err)
	}
	return nil
}

func getGlobalSettingsFileDest() (string, error) {

	m2Home, err := getEnvironmentVariable("M2_HOME")
	if err != nil {
		return "", err
	}
	return filepath.Join(m2Home, "conf", "settings.xml"), nil
}

func getProjectSettingsFileDest() (string, error) {
	home, err := getEnvironmentVariable("HOME")
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".m2", "settings.xml"), nil
}

func getEnvironmentVariable(name string) (string, error) {

	envVar := getenv(name)

	if len(envVar) == 0 {
		return "", fmt.Errorf("Environment variable \"%s\" not set or empty", name)
	}

	return envVar, nil
}
