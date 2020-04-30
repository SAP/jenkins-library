package maven

import (
	"errors"
	"fmt"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"os"
	"path/filepath"
	"strings"
)

var getenv = os.Getenv

// SettingsFileType ...
type SettingsFileType int

const (
	// GlobalSettingsFile ...
	GlobalSettingsFile SettingsFileType = iota
	// ProjectSettingsFile ...
	ProjectSettingsFile
)

// GetSettingsFile ...
func GetSettingsFile(settingsFileType SettingsFileType, src string, fileUtils piperutils.FileUtils, httpClient piperhttp.Downloader) error {

	var dest string
	var err error

	switch settingsFileType {
	case GlobalSettingsFile:
		dest, err = getGlobalSettingsFileDest()
	case ProjectSettingsFile:
		dest, err = getProjectSettingsFileDest()
	default:
		return errors.New("Invalid SettingsFileType")
	}

	if err != nil {
		return err
	}

	if len(src) == 0 {
		return fmt.Errorf("Settings file source location not provided")
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
