package cmd

import (
	"os"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
)

// Deprecated: Please use piperutils.Files{} instead
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func logWorkspaceContent() {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Entry().Errorf("Error getting current directory: %v", err)
	}
	log.Entry().Debugf("Contents of Workspace:")
	filepath.Walk(currentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Entry().Errorf("Error parsing current directory: %v", err)
		}
		mode := info.Mode()
		log.Entry().Debugf(" %s (%s)", path, mode)
		return nil
	})

}
