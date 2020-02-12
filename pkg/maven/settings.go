package maven

import (
	"errors"
	"fmt"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"io/ioutil"
	"net/http"
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
	ProjectSettingsFile SettingsFileType = iota
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

	if len(src) > 0 {

		log.Entry().Debugf("Copying file \"%s\" to \"%s\"", src, dest)

		if strings.HasPrefix(src, "http:") || strings.HasPrefix(src, "https:") {

			if err := httpClient.DownloadFile(src, dest, nil, nil); err != nil {
				return err
			}
		} else {

			// for sake os symetry it would be better to use a file protocol prefix here (file:)

			parent := filepath.Dir(dest)

			exists, err := fileUtils.FileExists(parent)

			if err != nil {
				return err
			}

			if !exists {
				if err = fileUtils.MkdirAll(parent, 0775); err != nil {
					return err
				}
			}

			exists, err = fileUtils.FileExists(src)

			if err != nil {
				return err
			}

			if !exists {
				return fmt.Errorf("File \"%s\" not found", src)
			}

			fmt.Printf("Copying file '%s' to '%s'", src, dest)
			if _, err := fileUtils.Copy(src, dest); err != nil {
				return err
			}
		}
	}
	log.Entry().Debugf("File \"%s\" copied to \"%s\"", src, dest)
	return nil
}

// getSettingsFileFromURL ...
func getSettingsFileFromURL(url, file string, fileUtils piperutils.FileUtils, httpClient piperhttp.Sender) error {

	var e error

	//CHECK:
	// - how does this work with a proxy inbetween?
	// - how does this work with http 302 (relocated) --> curl -L
	response, e := httpClient.SendRequest(http.MethodGet, url, nil, nil, nil)
	if e != nil {
		return e
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("Got %d reponse from download attempt for \"%s\"", response.StatusCode, url)
	}

	body, e := ioutil.ReadAll(response.Body)
	if e != nil {
		return e
	}

	e = fileUtils.FileWrite(file, body, 0644)
	if e != nil {
		return e
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
