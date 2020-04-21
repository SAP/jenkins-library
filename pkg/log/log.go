package log

import (
	"github.com/sirupsen/logrus"
	"strings"
)

type RemoveSecretFormatterDecorator struct {
	logrus.TextFormatter
}

func (formatter *RemoveSecretFormatterDecorator) Format(entry *logrus.Entry) (bytes []byte, err error) {
	formattedMessage, err := formatter.TextFormatter.Format(entry)

	if err != nil {
		return nil, err
	}

	message := string(formattedMessage)

	for _, secret := range secrets {
		message = strings.Replace(message, secret, "****", -1)
	}

	return []byte(message), nil
}

type RemoveSecretFormatterPlainMessageDecorator struct {
	logrus.TextFormatter
}

func (formatter *RemoveSecretFormatterPlainMessageDecorator) Format(entry *logrus.Entry) (bytes []byte, err error) {
	message := entry.Message + "\n"

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

func SetFormatter(messageOnlyLogFormat bool) {
	if messageOnlyLogFormat {
		logger.Logger.SetFormatter(&RemoveSecretFormatterPlainMessageDecorator{})
	} else {
		logger.Logger.SetFormatter(&RemoveSecretFormatterDecorator{})
	}
}

// SetStepName sets the stepName field.
func SetStepName(stepName string) {
	logger = Entry().WithField("stepName", stepName)
}

// DeferExitHandler registers a logrus exit handler to allow cleanup activities.
func DeferExitHandler(handler func()) {
	logrus.DeferExitHandler(handler)
}

func RegisterSecret(secret string) {
	if len(secret) > 0 {
		secrets = append(secrets, secret)
	}
}
