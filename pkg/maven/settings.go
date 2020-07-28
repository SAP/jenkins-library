package maven

import (
	"fmt"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"os"
	"path/filepath"
	"strings"
)

var getenv = os.Getenv

func DownloadAndCopySettingsFiles(globalSettingsFile string, projectSettingsFile string, fileUtils piperutils.FileUtils, httpClient piperhttp.Downloader) error {
	if len(projectSettingsFile) > 0 {
		destination, err := getProjectSettingsFileDest()
		if err != nil {
			return err
		}

		if err := GetSettingsFile(projectSettingsFile, destination, fileUtils, httpClient); err != nil {
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
		if err := GetSettingsFile(globalSettingsFile, destination, fileUtils, httpClient); err != nil {
			return err
		}
	} else {

		log.Entry().Debugf("Global settings file not provided via configuration.")
	}

	return nil
}

// GetSettingsFile ...
func GetSettingsFile(src string, dest string, fileUtils piperutils.FileUtils, httpClient piperhttp.Downloader) error {
	if len(src) == 0 {
		return fmt.Errorf("Settings file source location not provided")
	}

	if len(dest) == 0 {
		return fmt.Errorf("Settings file destination location not provided")
	}

	log.Entry().Debugf("Copying file \"%s\" to \"%s\"", src, dest)

	if strings.HasPrefix(src, "http:") || strings.HasPrefix(src, "https:") {

		if err := httpClient.DownloadFile(src, dest, nil, nil); err != nil {
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
