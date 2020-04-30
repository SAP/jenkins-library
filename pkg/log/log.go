package log

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

//PiperLogFormatter is the custom formatter of piper
type PiperLogFormatter struct {
	logrus.TextFormatter
	messageOnlyLogFormat bool
	withTimeStamp        bool
	defaultFormat        bool
}

const (
	logFormatPlain         = "plain"
	logFormatDefault       = "default"
	logFormatWithTimestamp = "timestamp"
)

//Format the log message
func (formatter *PiperLogFormatter) Format(entry *logrus.Entry) (bytes []byte, err error) {
	message := ""

	// experimental: align level with underlying tool (like maven or npm)
	if strings.Contains(entry.Message, "ERROR") || strings.Contains(entry.Message, "ERR!") {
		entry.Level = logrus.ErrorLevel
	}
	if strings.Contains(entry.Message, "WARN") {
		entry.Level = logrus.WarnLevel
	}

	if formatter.messageOnlyLogFormat {
		message = entry.Message + "\n"
	} else if formatter.withTimeStamp {
		message = fmt.Sprintf("%s %-5s %-6s - %s\n", entry.Time.Format("15:04:05"), entry.Level, entry.Data["stepName"], entry.Message)

	} else if formatter.defaultFormat {
		message = fmt.Sprintf("%-5s %-6s - %s\n", entry.Level, entry.Data["stepName"], entry.Message)

	} else {
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
var logger *logrus.Entry
var secrets []string

// Entry returns the logger entry or creates one if none is present.
func Entry() *logrus.Entry {
	if logger == nil {
		logger = logrus.WithField("library", LibraryRepository)
		logger.Logger.SetFormatter(&PiperLogFormatter{})
	}

	return logger
}

// SetVerbose sets the log level with respect to verbose flag.
func SetVerbose(verbose bool) {
	if verbose {
		//Logger().Debugf("logging set to level: %s", level)
		logrus.SetLevel(logrus.DebugLevel)
	}
}

// SetFormatter specifies whether to log only hte message
func SetFormatter(logFormat string) {
	Entry().Logger.SetFormatter(&PiperLogFormatter{
		messageOnlyLogFormat: logFormat == logFormatPlain,
		withTimeStamp:        logFormat == logFormatWithTimestamp,
		defaultFormat:        logFormat == logFormatDefault})
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

func RegisterSecret(secret string) {
	if len(secret) > 0 {
		secrets = append(secrets, secret)
	}
}
