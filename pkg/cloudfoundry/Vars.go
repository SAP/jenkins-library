package cloudfoundry

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

// VarsFilesNotFoundError ...
type VarsFilesNotFoundError struct {
	Message      string
	MissingFiles []string
}

func (e *VarsFilesNotFoundError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.MissingFiles)
}

var _fileUtils piperutils.FileUtils = piperutils.Files{}

// GetVarFileOptions Returns a string array containing valid var file options,
// e.g.: --vars myVars.yml
// The options array contains all vars files which could be resolved in the file system.
// In case some vars files cannot be found, the missing files are reported
// via the error which is in this case a VarsFilesNotFoundError. In that case the options
// array contains nevertheless the options for all existing files.
func GetVarFileOptions(varsFiles []string) ([]string, error) {
	varsFilesOpts := []string{}
	notFound := []string{}
	var err error
	for _, varsFile := range varsFiles {
		varsFileExists, err := _fileUtils.FileExists(varsFile)
		if err != nil {
			return nil, fmt.Errorf("Error accessing file system: %w", err)
		}
		if varsFileExists {
			varsFilesOpts = append(varsFilesOpts, "--vars-file", varsFile)
		} else {
			notFound = append(notFound, varsFile)
		}
	}
	if len(notFound) > 0 {
		err = &VarsFilesNotFoundError{
			Message:      "Some vars files could not be found",
			MissingFiles: notFound,
		}
	}
	return varsFilesOpts, err
}
