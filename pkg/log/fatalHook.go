package log

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	details["time"] = entry.Time

	fileName := "errorDetails.json"
	if details["stepName"] != nil {
		fileName = fmt.Sprintf("%v_%v", fmt.Sprint(details["stepName"]), fileName)
		// ToDo: If step is called x times, and it fails multiple times the error is overwritten
	}
	filePath := filepath.Join(f.Path, fileName)
	errDetails, _ := json.Marshal(&details)
	
	// Provide cleaner error output for GitHub Actions
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		// Extract meaningful error message from the error details
		errorMsg := fmt.Sprint(details["error"])
		stepName := fmt.Sprint(details["stepName"])
		
		// Create a clean, human-readable error summary
		if strings.Contains(errorMsg, "cmd.Run() failed: exit status") {
			// For command failures, focus on the actual command error
			var cleanMsg string
			if strings.Contains(errorMsg, "running command") {
				// Extract just the command that failed
				parts := strings.Split(errorMsg, "running command")
				if len(parts) > 1 {
					cmdPart := strings.Split(parts[1], "failed:")[0]
					cmdPart = strings.Trim(cmdPart, " '")
					cleanMsg = fmt.Sprintf("Command failed: %s", cmdPart)
				}
			}
			if cleanMsg == "" {
				cleanMsg = "Command execution failed"
			}
			Entry().Errorf("❌ %s: %s", stepName, cleanMsg)
		} else {
			// For other errors, show a clean version
			Entry().Errorf("❌ %s: %s", stepName, errorMsg)
		}
		
		// Still log the detailed JSON for debugging, but with debug level
		Entry().Debugf("fatal error: errorDetails%v", string(errDetails))
	} else {
		// Original behavior for non-GitHub Actions environments
		// Logging information needed for error reporting -  do not modify.
		Entry().Infof("fatal error: errorDetails%v", string(errDetails))
	}
	
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
