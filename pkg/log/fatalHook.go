package log

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	}
	filePath := filepath.Join(f.Path, fileName)

	_, err := ioutil.ReadFile(filePath)

	if err != nil {
		// ignore errors, since we don't want to break the logging flow
		errDetails, _ := json.Marshal(&details)
		ioutil.WriteFile(filePath, errDetails, 0666)
	}

	return nil
}
