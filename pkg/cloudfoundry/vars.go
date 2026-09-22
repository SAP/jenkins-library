package cloudfoundry

import (
	"fmt"
	"regexp"

	"github.com/SAP/jenkins-library/pkg/piperutils"
)

var varAssignment = regexp.MustCompile(`\S+=\S+`)

// VarsFilesNotFoundError reports the variable files that were not found.
type VarsFilesNotFoundError struct {
	Message      string
	MissingFiles []string
}

func (e *VarsFilesNotFoundError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.MissingFiles)
}

// GetVarsFileOptions returns the CF CLI options for existing variable files.
// If any requested file is missing, the returned error is a
// VarsFilesNotFoundError and the options still include the files that exist.
func GetVarsFileOptions(files []string) ([]string, error) {
	options := make([]string, 0, len(files)*2)
	missing := make([]string, 0)

	for _, file := range files {
		exists, err := piperutils.Files{}.FileExists(file)
		if err != nil {
			return nil, fmt.Errorf("Error accessing file system: %w", err)
		}
		if exists {
			options = append(options, "--vars-file", file)
			continue
		}
		missing = append(missing, file)
	}

	if len(missing) == 0 {
		return options, nil
	}
	return options, &VarsFilesNotFoundError{
		Message:      "Some vars files could not be found",
		MissingFiles: missing,
	}
}

// GetVarsOptions returns the CF CLI options for variable assignments in key=value form.
func GetVarsOptions(vars []string) ([]string, error) {
	options := make([]string, 0, len(vars)*2)
	invalid := make([]string, 0)

	for _, variable := range vars {
		if !varAssignment.MatchString(variable) {
			invalid = append(invalid, variable)
			continue
		}
		options = append(options, "--var", variable)
	}

	if len(invalid) > 0 {
		return nil, fmt.Errorf("Invalid vars: %v", invalid)
	}
	return options, nil
}
