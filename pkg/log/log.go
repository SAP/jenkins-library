package log

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// PiperLogFormatter is the custom formatter of piper
type PiperLogFormatter struct {
	logrus.TextFormatter
	logFormat string
}

const (
	logFormatPlain         = "plain"
	logFormatDefault       = "default"
	logFormatWithTimestamp = "timestamp"
	logFormatGitHubActions = "github"
)

// Format the log message
func (formatter *PiperLogFormatter) Format(entry *logrus.Entry) (bytes []byte, err error) {
	message := ""

	stepName := entry.Data["stepName"]
	if stepName == nil {
		stepName = "(noStepName)"
	}

	errorMessageSnippet := ""
	if entry.Data[logrus.ErrorKey] != nil {
		errorMessageSnippet = fmt.Sprintf(" - %s", entry.Data[logrus.ErrorKey])
	}

	// Add structured fields (excluding stepName and error which are handled separately)
	fieldsSnippet := ""
	for key, value := range entry.Data {
		if key != "stepName" && key != logrus.ErrorKey {
			if fieldsSnippet != "" {
				fieldsSnippet += " "
			}
			fieldsSnippet += fmt.Sprintf("%s=%v", key, value)
		}
	}
	if fieldsSnippet != "" {
		fieldsSnippet = " [" + fieldsSnippet + "]"
	}

	level, _ := entry.Level.MarshalText()
	levelString := string(level)
	if levelString == "warning" {
		levelString = "warn"
	}

	switch formatter.logFormat {
	case logFormatDefault:
		message = fmt.Sprintf("%-5s %-6s - %s%s%s\n", levelString, stepName, entry.Message, fieldsSnippet, errorMessageSnippet)
	case logFormatWithTimestamp:
		message = fmt.Sprintf("%s %-5s %-6s %s%s%s\n", entry.Time.Format("15:04:05"), levelString, stepName, entry.Message, fieldsSnippet, errorMessageSnippet)
	case logFormatPlain:
		message = fmt.Sprintf("%s%s%s\n", entry.Message, fieldsSnippet, errorMessageSnippet)
	case logFormatGitHubActions:
		// GitHub Actions format: use GitHub Actions logging commands and cleaner output
		switch levelString {
		case "fatal":
			message = fmt.Sprintf("::error::%s%s\n", entry.Message, errorMessageSnippet)
		case "error":
			message = fmt.Sprintf("::error::%s%s%s\n", entry.Message, fieldsSnippet, errorMessageSnippet)
		case "warn":
			message = fmt.Sprintf("::warning::%s%s%s\n", entry.Message, fieldsSnippet, errorMessageSnippet)
		case "debug":
			message = fmt.Sprintf("::debug::%s%s%s\n", entry.Message, fieldsSnippet, errorMessageSnippet)
		default: // info and others
			// For info messages, show clean output without step repetition
			message = fmt.Sprintf("%s%s%s\n", entry.Message, fieldsSnippet, errorMessageSnippet)
		}
	default:
		formattedMessage, err := formatter.TextFormatter.Format(entry)
		if err != nil {
			return nil, err
		}
		message = string(formattedMessage)
	}

	for _, secret := range secrets {
		message = strings.Replace(message, secret, "****", -1)
	}

	return []byte(message), nil
}

// LibraryRepository that is passed into with -ldflags
var LibraryRepository string
var LibraryName string
var logger *logrus.Entry
var secrets []string

// Entry returns the logger entry or creates one if none is present.
func Entry() *logrus.Entry {
	if logger == nil {
		logger = logrus.NewEntry(logrus.StandardLogger())

		// Auto-detect GitHub Actions environment and set appropriate format
		logFormat := logFormatDefault
		if isGitHubActions() {
			logFormat = logFormatGitHubActions
		}

		logger.Logger.SetFormatter(&PiperLogFormatter{logFormat: logFormat})
	}

	return logger
}

// Writer returns an io.Writer into which a tool's output can be redirected.
func Writer() io.Writer {
	return &logrusWriter{logger: Entry()}
}

// SetVerbose sets the log level with respect to verbose flag.
func SetVerbose(verbose bool) {
	if verbose {
		//Logger().Debugf("logging set to level: %s", level)
		logrus.SetLevel(logrus.DebugLevel)
	}
}

// IsVerbose returns true if DegbuLevel is enabled.
func IsVerbose() bool {
	return logrus.GetLevel() == logrus.DebugLevel
}

// SetFormatter specifies the log format to use for piper's output
func SetFormatter(logFormat string) {
	// Auto-detect GitHub Actions environment and override format if needed
	if isGitHubActions() {
		// Override any format with GitHub Actions format when running in GitHub Actions
		logFormat = logFormatGitHubActions
	}
	Entry().Logger.SetFormatter(&PiperLogFormatter{logFormat: logFormat})
}

// SetStepName sets the stepName field.
func SetStepName(stepName string) {
	logger = Entry().WithField("stepName", stepName)
}

// DeferExitHandler registers a logrus exit handler to allow cleanup activities.
func DeferExitHandler(handler func()) {
	logrus.DeferExitHandler(handler)
}

// RegisterHook registers a logrus hook
func RegisterHook(hook logrus.Hook) {
	logrus.AddHook(hook)
}

// RegisterSecret registers a value which should be masked in every log message
func RegisterSecret(secret string) {
	if len(secret) > 0 {
		secrets = append(secrets, secret)
		encoded := url.QueryEscape(secret)
		if secret != encoded {
			secrets = append(secrets, encoded)
		}
	}
}

// FIXME: The one from orchestrator/github_actions.go cannot be reused here because of circular dependency
func isGitHubActions() bool {
	envVars := []string{"GITHUB_ACTION", "GITHUB_ACTIONS"}
	for _, v := range envVars {
		if val, exists := os.LookupEnv(v); exists && len(val) > 0 && val != "false" && val != "0" {
			return true
		}
	}
	return false
}
