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
		// ToDo: If step is called x times, and it fails multiple times the error is overwritten
	}
	filePath := filepath.Join(f.Path, fileName)
	errDetails, _ := json.Marshal(&details)
	Entry().Infof("fatal error: errorDetails{correlationId:\"%v\",stepName:\"%v\",category:\"%v\",error:\"%v\",result:\"%v\",message:\"%v\"}",
		details["correlationId"], details["stepName"], details["category"], details["error"], details["result"], details["message"])

	_, err := ioutil.ReadFile(filePath)
	if err != nil {
		// do not overwrite file in case it already exists
		// this helps to report the first error which occured - instead of the last one
		ioutil.WriteFile(filePath, errDetails, 0666)
	}

	return nil
}

// ErrorDetails struct holds information about errors of the step
type ErrorDetails struct {
	Message       string
	Error         string
	Category      string
	Result        string
	CorrelationId string
	StepName      string
}

// GetErrorsJson reads errorDetails.json files from the CPE and returns an ErrorDetails struct.
func GetErrorsJson() ([]ErrorDetails, error) {
	fileName := "errorDetails.json"
	path, err := os.Getwd()
	if err != nil {
		Entry().Error("can not get current working dir")
		return []ErrorDetails{}, err
	}

	pathCPE := path + "/.pipeline/commonPipelineEnvironment"
	matches, err := filepath.Glob(pathCPE + "/*" + fileName)
	if err != nil {
		Entry().Error("could not search filepath for *errorDetails.json files")
		return []ErrorDetails{}, err
	}
	if len(matches) == 0 {
		Entry().Debug("no errors found, returning empty errorDetails")
		return []ErrorDetails{}, nil
	}
	Entry().Debugf("found the following errorDetails files: %v", matches)

	var errorDetails []ErrorDetails
	Entry().Debugf("Found %v files", matches)

	for _, v := range matches {
		errorDetail, err := readErrorJson(v)
		if err != nil {
			Entry().Errorf("could not read error details for file %v", v)
			errorDetail = ErrorDetails{}
		}
		errorDetails = append(errorDetails, errorDetail)

	}
	return errorDetails, nil
}

func readErrorJson(filePath string) (ErrorDetails, error) {
	errorDetails := ErrorDetails{}
	jsonFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		Entry().Errorf("could not read file from path: %v", filePath)
		return ErrorDetails{}, err
	}
	err = json.Unmarshal(jsonFile, &errorDetails)
	if err != nil {
		Entry().Error("could not unmarshal error details")
		return ErrorDetails{}, err
	}
	return errorDetails, nil
}
