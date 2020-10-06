package cloudfoundry

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"regexp"
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

// GetVarsFileOptions Returns a string array containing valid var file options,
// e.g.: --vars myVars.yml
// The options array contains all vars files which could be resolved in the file system.
// In case some vars files cannot be found, the missing files are reported
// via the error which is in this case a VarsFilesNotFoundError. In that case the options
// array contains nevertheless the options for all existing files.
func GetVarsFileOptions(varsFiles []string) ([]string, error) {
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

// GetVarsOptions Returns the vars as valid var option string slice
// InvalidVars are reported via error.
func GetVarsOptions(vars []string) ([]string, error) {
	invalidVars := []string{}
	varOptions := []string{}
	var err error
	for _, v := range vars {
		valid, e := validateVar(v)
		if e != nil {
			return []string{}, fmt.Errorf("Cannot validate var '%s': %w", v, e)
		}
		if !valid {
			invalidVars = append(invalidVars, v)
			continue
		}
		varOptions = append(varOptions, "--var", v)
	}

	if len(invalidVars) > 0 {
		return []string{}, fmt.Errorf("Invalid vars: %v", invalidVars)
	}
	return varOptions, err
}

func validateVar(v string) (bool, error) {
	return regexp.MatchString(`\S{1,}=\S{1,}`, v)
}
