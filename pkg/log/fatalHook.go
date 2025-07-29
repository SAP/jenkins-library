package log

import (
	"encoding/json"
	"fmt"
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

	// Check for recent pattern matches first, then fall back to entry message
	if lastMatch := getLastPatternMatch(); lastMatch != "" {
		details["message"] = lastMatch
	} else {
		details["message"] = entry.Message
	}
	details["error"] = fmt.Sprint(details["error"])
	details["category"] = GetErrorCategory().String()
	details["result"] = "failure"
	details["correlationId"] = f.CorrelationID
	details["time"] = entry.Time

	fileName := "errorDetails.json"
	if details["stepName"] != nil {
		fileName = fmt.Sprintf("%v_%v", fmt.Sprint(details["stepName"]), fileName)
		// ToDo: If step is called x times, and it fails multiple times the error is overwritten
	}
	filePath := filepath.Join(f.Path, fileName)
	errDetails, _ := json.Marshal(&details)

	// Sets the fatal error details in the logging framework to be consumed in the stepTelemetryData
	SetFatalErrorDetail(errDetails)
	_, err := os.ReadFile(filePath)
	if err != nil {
		// do not overwrite file in case it already exists
		// this helps to report the first error which occurred - instead of the last one
		os.WriteFile(filePath, errDetails, 0666)
	}

	return nil
}
