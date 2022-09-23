package pact

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
)

// EnsureDir ensures the specified directory does not already exist before creating it.
func EnsureDir(folderPath string, utils Utils) error {
	if exists, serr := utils.DirExists(folderPath); !exists || serr != nil {
		if merr := utils.MkdirAll(folderPath, 0777); merr != nil {
			return merr
		}
	}
	return nil
}

// EnsureValidDir ensures the directory path passed in as an argument is in the correct format /Path/To/Dir/
// The function will format and return a valid path if it is not currently correct.
func EnsureValidDir(folderPath string) string {
	// Ensures path to pacts is in correct format by extracting just the directory path if a file path is given
	if strings.HasSuffix(folderPath, ".json") {
		folderPath = filepath.Dir(folderPath) + "/"
	}

	if !strings.HasSuffix(folderPath, "/") {
		folderPath = folderPath + "/"
	}

	return folderPath
}

// MustGetenv ensures the environment variable passed in as an argument exists, else it logs a fatal error
/*
func MustGetenv(key string) string {
	var value string
	var exist bool
	if value, exist = os.LookupEnv(key); !exist || value == "" {
		log.Fatalf("lack env %s", key)
	}
	return value
}
*/

// ReadAndUnmarshalFile reads in a file and unmarshals into the interface passed in as an argument.
// spec must be a reference variable because there is no return value
func ReadAndUnmarshalFile(file string, spec interface{}, utils Utils) error {
	byteValue, err := utils.ReadFile(file)
	if err != nil {
		return err
	}

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'spec' which we defined above
	if err := json.Unmarshal(byteValue, spec); err != nil {
		return fmt.Errorf("failed to unmarshal %v: %w", file, err)
	}

	return nil
}

// sendRequest is a wrapper for sending http request
func sendRequest(method, url, username, password string, body io.Reader, utils Utils) ([]byte, error) {
	utils.SetOptions(piperhttp.ClientOptions{Username: username, Password: password})
	resp, err := utils.SendRequest(http.MethodGet, url, body, nil, nil)

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	byteResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return byteResp, nil
}