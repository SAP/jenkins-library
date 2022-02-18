package project

import (
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/log"
)

// 1. If "path" is a directory, check for "${path}/${descriptor}"
// 2. If "path" is a file, check for "${PWD}/${descriptor}"
func ResolvePath(descriptor, path string, utils cnbutils.BuildUtils) (string, error) {
	isDir, err := utils.DirExists(path)
	if err != nil {
		return "", err
	}

	if isDir {
		return getFilename(path, descriptor, utils)
	}

	pwd, err := utils.Getwd()
	if err != nil {
		return "", err
	}
	return getFilename(pwd, descriptor, utils)
}

func getFilename(folder, filename string, utils cnbutils.BuildUtils) (string, error) {
	descPath := filepath.Join(folder, filename)
	exists, err := utils.FileExists(descPath)
	if err != nil {
		return "", err
	}

	if exists {
		return descPath, nil
	}

	log.Entry().Infof("Project descriptor with the path '%s' was not found", descPath)
	return "", nil
}
