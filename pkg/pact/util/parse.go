package util

import (
	"encoding/json"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/piperutils"
)

// ReadAndUnmarshalFile reads in a file and unmarshals into the interface passed in as an argument.
// spec must be a reference variable because there is no return value
func ReadAndUnmarshalFile(file string, spec interface{}, utils piperutils.FileUtils) error {

	// Open and read jsonFile
	byteValue, err := utils.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read and open file '%v': %w", file, err)
	}

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'spec' which we defined above
	if err := json.Unmarshal(byteValue, spec); err != nil {
		return fmt.Errorf("failed to parse json file '%v': %w", file, err)
	}

	return nil
}
