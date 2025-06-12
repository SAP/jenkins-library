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
		fatalMessage := entry.Message
		
		// Prioritize the actual error over generic messages like "step execution failed"
		var displayMsg string
		if errorMsg != "" && errorMsg != "<nil>" {
			// Use the actual error message
			displayMsg = errorMsg
		} else if fatalMessage != "" {
			// Fallback to the fatal message
			displayMsg = fatalMessage
		} else {
			displayMsg = "Unknown error"
		}
		
		// Clean up common verbose error patterns for better readability
		if strings.Contains(displayMsg, "cmd.Run() failed: exit status") && strings.Contains(displayMsg, "running command") {
			// Extract the command that failed for cleaner display
			parts := strings.Split(displayMsg, "running command")
			if len(parts) > 1 {
				cmdPart := strings.Split(parts[1], "failed:")[0]
				cmdPart = strings.Trim(cmdPart, " '")
				displayMsg = fmt.Sprintf("Command '%s' failed", cmdPart)
			}
		}
		
		// Log as GitHub Actions error with step context
		Entry().Errorf("❌ %s: %s", stepName, displayMsg)
		
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
