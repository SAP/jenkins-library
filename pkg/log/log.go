package log

import (
	"fmt"
	"io"
	"strings"

	"github.com/sirupsen/logrus"
)

//PiperLogFormatter is the custom formatter of piper
type PiperLogFormatter struct {
	logrus.TextFormatter
	logFormat string
}

const (
	logFormatPlain         = "plain"
	logFormatDefault       = "default"
	logFormatWithTimestamp = "timestamp"
)

//Format the log message
func (formatter *PiperLogFormatter) Format(entry *logrus.Entry) (bytes []byte, err error) {
	message := ""

	stepName := entry.Data["stepName"]
	if stepName == nil {
		stepName = "(noStepName)"
	}

	switch formatter.logFormat {
	case logFormatDefault:
		message = fmt.Sprintf("%-5s %-6s - %s\n", entry.Level, stepName, entry.Message)
	case logFormatWithTimestamp:
		message = fmt.Sprintf("%s %-5s %-6s - %s\n", entry.Time.Format("15:04:05"), entry.Level, stepName, entry.Message)
	case logFormatPlain:
		message = entry.Message + "\n"
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

// SetFormatter specifies the log format to use for piper's output
func SetFormatter(logFormat string) {
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

func RegisterSecret(secret string) {
	if len(secret) > 0 {
		secrets = append(secrets, secret)
	}
}
