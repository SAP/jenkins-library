package maven

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var getenv = os.Getenv

// SettingsDownloadUtils defines an interface for downloading files.
type SettingsDownloadUtils interface {
	DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error
}

// FileUtils defines the external file-related functionality needed by this package.
type FileUtils interface {
	FileExists(filename string) (bool, error)
	Copy(src, dest string) (int64, error)
	MkdirAll(path string, perm os.FileMode) error
	Glob(pattern string) (matches []string, err error)
}

// DownloadAndGetMavenParameters downloads the global or project settings file if the strings contain URLs.
// It then constructs the arguments that need to be passed to maven in order to point to use these settings files.
func DownloadAndGetMavenParameters(globalSettingsFile string, projectSettingsFile string, fileUtils FileUtils, httpClient SettingsDownloadUtils) ([]string, error) {
	mavenArgs := []string{}
	if len(globalSettingsFile) > 0 {
		globalSettingsFileName, err := downloadSettingsIfURL(globalSettingsFile, ".pipeline/mavenGlobalSettings.xml", fileUtils, httpClient, false)
		if err != nil {
			return nil, err
		}
		mavenArgs = append(mavenArgs, "--global-settings", globalSettingsFileName)
	} else {

		log.Entry().Debugf("Global settings file not provided via configuration.")
	}

	if len(projectSettingsFile) > 0 {
		projectSettingsFileName, err := downloadSettingsIfURL(projectSettingsFile, ".pipeline/mavenProjectSettings.xml", fileUtils, httpClient, false)
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
func DownloadAndCopySettingsFiles(globalSettingsFile string, projectSettingsFile string, fileUtils FileUtils, httpClient SettingsDownloadUtils) error {
	if len(projectSettingsFile) > 0 {
		destination, err := getProjectSettingsFileDest()
		if err != nil {
			return err
		}

		if err := downloadAndCopySettingsFile(projectSettingsFile, destination, fileUtils, httpClient); err != nil {
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
		if err := downloadAndCopySettingsFile(globalSettingsFile, destination, fileUtils, httpClient); err != nil {
			return err
		}
	} else {

		log.Entry().Debugf("Global settings file not provided via configuration.")
	}

	return nil
}

func downloadAndCopySettingsFile(src string, dest string, fileUtils FileUtils, httpClient SettingsDownloadUtils) error {
	if len(src) == 0 {
		return fmt.Errorf("Settings file source location not provided")
	}

	if len(dest) == 0 {
		return fmt.Errorf("Settings file destination location not provided")
	}

	log.Entry().Debugf("Copying file \"%s\" to \"%s\"", src, dest)

	if strings.HasPrefix(src, "http:") || strings.HasPrefix(src, "https:") {
		err := downloadSettingsFromURL(src, dest, fileUtils, httpClient, true)
		if err != nil {
			return err
		}
	} else {

		// for sake os symmetry it would be better to use a file protocol prefix here (file:)

		parent := filepath.Dir(dest)

		parentFolderExists, err := fileUtils.FileExists(parent)

		if err != nil {
			return err
		}

		if !parentFolderExists {
			if err = fileUtils.MkdirAll(parent, 0775); err != nil {
				return err
			}
		}

		if _, err := fileUtils.Copy(src, dest); err != nil {
			return err
		}
	}

	return nil
}

func downloadSettingsIfURL(settingsFileOption, settingsFile string, fileUtils FileUtils, httpClient SettingsDownloadUtils, overwrite bool) (string, error) {
	result := settingsFileOption
	if strings.HasPrefix(settingsFileOption, "http:") || strings.HasPrefix(settingsFileOption, "https:") {
		err := downloadSettingsFromURL(settingsFileOption, settingsFile, fileUtils, httpClient, overwrite)
		if err != nil {
			return "", err
		}
		result = settingsFile
	}
	return result, nil
}

func downloadSettingsFromURL(url, filename string, fileUtils FileUtils, httpClient SettingsDownloadUtils, overwrite bool) error {
	exists, _ := fileUtils.FileExists(filename)
	if exists && !overwrite {
		log.Entry().Infof("Not downloading maven settings file, because it already exists at '%s'", filename)
		return nil
	}
	err := httpClient.DownloadFile(url, filename, nil, nil)
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
	return m2Home + "/conf/settings.xml", nil
}

func getProjectSettingsFileDest() (string, error) {
	home, err := getEnvironmentVariable("HOME")
	if err != nil {
		return "", err
	}
	return home + "/.m2/settings.xml", nil
}

func getEnvironmentVariable(name string) (string, error) {

	envVar := getenv(name)

	if len(envVar) == 0 {
		return "", fmt.Errorf("Environment variable \"%s\" not set or empty", name)
	}

	return envVar, nil
}
