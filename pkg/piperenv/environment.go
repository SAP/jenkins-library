package piperenv

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
)

// This file contains functions used to read/write pipeline environment data from/to disk.
// The content of the file is the value. For the custom parameters this could for example also be a JSON representation of a more complex value.
// The file representation looks as follows:

// <pipeline env path>/

// <pipeline env path>/artifactVersion

// <pipeline env path>/git/
// <pipeline env path>/git/branch
// <pipeline env path>/git/commitId
// <pipeline env path>/git/commitMessage
// <pipeline env path>/git/repositoryUrl -> TODO: storing function(s) with ssh and https getters

// <pipeline env path>/github/
// <pipeline env path>/github/owner
// <pipeline env path>/github/repository

// <pipeline env path>/custom/
// <pipeline env path>/custom/<parameter>

// SetResourceParameter sets a resource parameter in the environment stored in the file system
func SetResourceParameter(path, resourceName, paramName, value string) error {
	paramPath := filepath.Join(path, resourceName, paramName)
	return writeToDisk(paramPath, []byte(value))
}

// GetResourceParameter reads a resource parameter from the environment stored in the file system
func GetResourceParameter(path, resourceName, paramName string) string {
	paramPath := filepath.Join(path, resourceName, paramName)
	return readFromDisk(paramPath)
}

// SetParameter sets any parameter in the pipeline environment or another environment stored in the file system
func SetParameter(path, name, value string) error {
	paramPath := filepath.Join(path, name)
	return writeToDisk(paramPath, []byte(value))
}

// GetParameter reads any parameter from the pipeline environment or another environment stored in the file system
func GetParameter(path, name string) string {
	paramPath := filepath.Join(path, name)
	return readFromDisk(paramPath)
}

func writeToDisk(filename string, data []byte) error {

	if _, err := os.Stat(filepath.Dir(filename)); os.IsNotExist(err) {
		log.Entry().Debugf("Creating directory: %v", filepath.Dir(filename))
		os.MkdirAll(filepath.Dir(filename), 0700)
	}

	//ToDo: make sure to not overwrite file but rather add another file? Create error if already existing?
	if len(data) > 0 {
		log.Entry().Debugf("Writing file to disk: %v", filename)
		return ioutil.WriteFile(filename, data, 0700)
	}
	return nil
}

func readFromDisk(filename string) string {
	//ToDo: if multiple files exist, read from latest file
	log.Entry().Debugf("Reading file from disk: %v", filename)
	v, err := ioutil.ReadFile(filename)
	val := string(v)
	if err != nil {
		val = ""
	}
	return val
}
