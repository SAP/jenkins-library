//go:build unit
// +build unit

package maven

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestSettings(t *testing.T) {

	defer func() {
		getenv = os.Getenv
	}()

	getenv = func(name string) string {
		if name == "M2_HOME" {
			return "/usr/share/maven"
		} else if name == "HOME" {
			return "/home/me"
		}
		return ""
	}

	t.Run("Settings file source location not provided", func(t *testing.T) {

		utilsMock := newSettingsDownloadTestUtilsBundle()

		err := downloadAndCopySettingsFile("", "foo", utilsMock)

		assert.EqualError(t, err, "Settings file source location not provided")
	})

	t.Run("Settings file destination location not provided", func(t *testing.T) {

		utilsMock := newSettingsDownloadTestUtilsBundle()

		err := downloadAndCopySettingsFile("/opt/sap/maven/global-settings.xml", "", utilsMock)

		assert.EqualError(t, err, "Settings file destination location not provided")
	})

	t.Run("Retrieve settings files", func(t *testing.T) {

		utilsMock := newSettingsDownloadTestUtilsBundle()

		utilsMock.AddFile("/opt/sap/maven/global-settings.xml", []byte(""))
		utilsMock.AddFile("/opt/sap/maven/project-settings.xml", []byte(""))

		err := DownloadAndCopySettingsFiles("/opt/sap/maven/global-settings.xml", "/opt/sap/maven/project-settings.xml", utilsMock)

		if assert.NoError(t, err) {
			assert.True(t, utilsMock.HasCopiedFile("/opt/sap/maven/global-settings.xml", "/usr/share/maven/conf/settings.xml"))
			assert.True(t, utilsMock.HasCopiedFile("/opt/sap/maven/project-settings.xml", "/home/me/.m2/settings.xml"))
		}

		assert.Empty(t, utilsMock.downloadedFiles)
	})

	t.Run("Retrieve settings file via http", func(t *testing.T) {

		utilsMock := newSettingsDownloadTestUtilsBundle()

		err := downloadAndCopySettingsFile("https://example.org/maven/global-settings.xml", "/usr/share/maven/conf/settings.xml", utilsMock)

		if assert.NoError(t, err) {
			assert.Equal(t, "/usr/share/maven/conf/settings.xml", utilsMock.downloadedFiles["https://example.org/maven/global-settings.xml"])
		}
	})

	t.Run("Retrieve settings file via http - received error from downloader", func(t *testing.T) {

		utilsMock := newSettingsDownloadTestUtilsBundle()
		utilsMock.expectedError = fmt.Errorf("Download failed")

		err := downloadAndCopySettingsFile("https://example.org/maven/global-settings.xml", "/usr/share/maven/conf/settings.xml", utilsMock)

		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "failed to download maven settings from URL")
		}
	})

	t.Run("Retrieve project settings file - file not found", func(t *testing.T) {

		utilsMock := newSettingsDownloadTestUtilsBundle()

		err := downloadAndCopySettingsFile("/opt/sap/maven/project-settings.xml", "/home/me/.m2/settings.xml", utilsMock)

		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "cannot copy '/opt/sap/maven/project-settings.xml': file does not exist")
		}
	})

	t.Run("create new Project settings file", func(t *testing.T) {

		utilsMock := newSettingsDownloadTestUtilsBundle()

		projectSettingsFilePath, err := CreateNewProjectSettingsXML("dummyRepoId", "dummyRepoUser", "dummyRepoPassword", utilsMock)
		if assert.NoError(t, err) {
			projectSettingsContent, _ := utilsMock.FileRead(projectSettingsFilePath)
			var projectSettings Settings

			err = xml.Unmarshal(projectSettingsContent, &projectSettings)

			if assert.NoError(t, err) {
				assert.Equal(t, projectSettings.Servers.ServerType[0].ID, "dummyRepoId")
				assert.Equal(t, projectSettings.Servers.ServerType[0].Username, "dummyRepoUser")
				assert.Equal(t, projectSettings.Servers.ServerType[0].ID, "dummyRepoId")
			}

		}
	})

	t.Run("update server tag in existing settings file", func(t *testing.T) {

		utilsMock := newSettingsDownloadTestUtilsBundle()
		var projectSettings Settings
		projectSettings.Xsi = "http://www.w3.org/2001/XMLSchema-instance"
		projectSettings.SchemaLocation = "http://maven.apache.org/SETTINGS/1.0.0 https://maven.apache.org/xsd/settings-1.0.0.xsd"
		projectSettings.Servers.ServerType = []Server{
			{
				ID:       "dummyRepoId1",
				Username: "dummyRepoUser1",
				Password: "dummyRepoId1",
			},
		}
		settingsXml, err := xml.MarshalIndent(projectSettings, "", "    ")
		settingsXmlString := string(settingsXml)
		Replacer := strings.NewReplacer("&#xA;", "", "&#x9;", "")
		settingsXmlString = Replacer.Replace(settingsXmlString)

		xmlstring := []byte(xml.Header + settingsXmlString)

		utilsMock.FileWrite(".pipeline/mavenProjectSettings", xmlstring, 0777)

		projectSettingsFilePath, err := UpdateProjectSettingsXML(".pipeline/mavenProjectSettings", "dummyRepoId2", "dummyRepoUser2", "dummyRepoPassword2", utilsMock)
		if assert.NoError(t, err) {
			projectSettingsContent, _ := utilsMock.FileRead(projectSettingsFilePath)
			var projectSettings Settings

			err = xml.Unmarshal(projectSettingsContent, &projectSettings)

			if assert.NoError(t, err) {
				assert.Equal(t, projectSettings.Servers.ServerType[1].ID, "dummyRepoId2")
				assert.Equal(t, projectSettings.Servers.ServerType[1].Username, "dummyRepoUser2")
				assert.Equal(t, projectSettings.Servers.ServerType[1].ID, "dummyRepoId2")
			}

		}
	})

	t.Run("update server tag in existing settings file - invalid settings.xml", func(t *testing.T) {

		utilsMock := newSettingsDownloadTestUtilsBundle()
		xmlstring := []byte("well this is obviously invalid")
		utilsMock.FileWrite(".pipeline/mavenProjectSettings", xmlstring, 0777)

		_, err := UpdateProjectSettingsXML(".pipeline/mavenProjectSettings", "dummyRepoId2", "dummyRepoUser2", "dummyRepoPassword2", utilsMock)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "failed to unmarshal settings xml file")
		}
	})

	t.Run("update active profile tag in existing settings file", func(t *testing.T) {

		utilsMock := newSettingsDownloadTestUtilsBundle()
		var projectSettings Settings
		projectSettings.Xsi = "http://www.w3.org/2001/XMLSchema-instance"
		projectSettings.SchemaLocation = "http://maven.apache.org/SETTINGS/1.0.0 https://maven.apache.org/xsd/settings-1.0.0.xsd"
		projectSettings.ActiveProfiles.ActiveProfile = []string{"dummyProfile"}
		settingsXml, err := xml.MarshalIndent(projectSettings, "", "    ")
		settingsXmlString := string(settingsXml)
		Replacer := strings.NewReplacer("&#xA;", "", "&#x9;", "")
		settingsXmlString = Replacer.Replace(settingsXmlString)

		xmlstring := []byte(xml.Header + settingsXmlString)

		destination, _ := getGlobalSettingsFileDest()

		utilsMock.FileWrite("/usr/share/maven/conf/settings.xml", xmlstring, 0777)

		err = UpdateActiveProfileInSettingsXML([]string{"newProfile"}, utilsMock)

		if assert.NoError(t, err) {
			projectSettingsContent, _ := utilsMock.FileRead(destination)
			var projectSettings Settings

			err = xml.Unmarshal(projectSettingsContent, &projectSettings)

			if assert.NoError(t, err) {
				assert.Equal(t, projectSettings.ActiveProfiles.ActiveProfile[0], "newProfile")
			}

		}
	})
}

func newSettingsDownloadTestUtilsBundle() *settingsDownloadTestUtils {
	utilsBundle := settingsDownloadTestUtils{
		FilesMock: &mock.FilesMock{},
	}
	return &utilsBundle
}

type settingsDownloadTestUtils struct {
	*mock.FilesMock
	expectedError   error
	downloadedFiles map[string]string // src, dest
}

func (c *settingsDownloadTestUtils) SetOptions(options piperhttp.ClientOptions) {
}

func (c *settingsDownloadTestUtils) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {

	if c.expectedError != nil {
		return c.expectedError
	}

	if c.downloadedFiles == nil {
		c.downloadedFiles = make(map[string]string)
	}
	c.downloadedFiles[url] = filename
	return nil
}
