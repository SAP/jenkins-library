package log

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// FatalHook provides a logrus hook which persists details about a fatal error into the file system.
// This is helpful in order to transfer the error details to an orchestrating CI/CD system
// and by that make it possible to provide better error messages to the user.
type FatalHook struct {
	Path          string
	CorrelationID string
}

// Levels returns the supported log level of the hook.
func (f *FatalHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.FatalLevel}
}

// Fire persists the error message of the fatal error as json file into the file system.
func (f *FatalHook) Fire(entry *logrus.Entry) error {
	details := entry.Data
	if details == nil {
		details = logrus.Fields{}
	}

	details["message"] = entry.Message
	details["error"] = fmt.Sprint(details["error"])
	details["category"] = GetErrorCategory().String()
	details["result"] = "failure"
	details["correlationId"] = f.CorrelationID

	fileName := "errorDetails.json"
	if details["stepName"] != nil {
		fileName = fmt.Sprintf("%v_%v", fmt.Sprint(details["stepName"]), fileName)
		// ToDo: If step is called x times and it fails multiple times the error is overwritten
	}
	filePath := filepath.Join(f.Path, fileName)
	filePathCPE := filepath.Join(f.Path, ".pipeline", "commonPipelineEnvironment", fileName)

	errDetails, _ := json.Marshal(&details)
	ioutil.WriteFile(filePathCPE, errDetails, 0666)
	Entry().Debugf("persisted error information in %v", filePathCPE)

	_, err := ioutil.ReadFile(filePath)
	if err != nil {
		// do not overwrite file in case it already exists
		// this helps to report the first error which occured - instead of the last one
		ioutil.WriteFile(filePath, errDetails, 0666)
	}

	return nil
}

type ErrorDetails struct {
	Message       string
	Error         string
	Category      string
	Result        string
	CorrelationId string
}

func GetErrorsJson() ([]ErrorDetails, error) {
	fileName := "errorDetails.json"
	path, err := os.Getwd()
	pathCPE := path + "/.pipeline/commonPipelineEnvironment"
	if err != nil {
		fmt.Errorf("can not get current working dir")
		return []ErrorDetails{}, err
	}

	matches, err := filepath.Glob(pathCPE + "/*" + fileName)
	Entry().Debugf("found the following errorDetails files: %v", matches)
	if err != nil {
		Entry().Debugf("could not find any *errorDetails.json files")
	}

	errorDetails := []ErrorDetails{}

	if len(matches) != 0 {
		Entry().Debugf("Found %v files", matches)

		for _, v := range matches {
			errorDetail, err := readErrorJson(v)
			if err != nil {
				Entry().Debugf("could not read error details for file %v", v)
			}
			errorDetails = append(errorDetails, errorDetail)

		}
	}

	return errorDetails, nil
}

func readErrorJson(filePath string) (ErrorDetails, error) {
	//filePath := filepath.Join(path, fileName)

	errorDetails := ErrorDetails{}
	jsonFile, err := ioutil.ReadFile(filePath)

	err = json.Unmarshal(jsonFile, &errorDetails)
	if err != nil {
		fmt.Errorf("could not unmarshal error details")
		return ErrorDetails{}, err
	}
	return errorDetails, nil
}
