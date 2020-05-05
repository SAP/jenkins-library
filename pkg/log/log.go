package log

import (
	"io"
	"strings"

	"github.com/sirupsen/logrus"
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

// LibraryRepository that is passed into with -ldflags
var LibraryRepository string
var logger *logrus.Entry
var secrets []string

// Entry returns the logger entry or creates one if none is present.
func Entry() *logrus.Entry {
	if logger == nil {
		logger = logrus.WithField("library", LibraryRepository)
		logger.Logger.SetFormatter(&RemoveSecretFormatterDecorator{})
	}
	return logger
}

// Writer returns an io.Writer into which a tool's output can be redirected.
func Writer() io.Writer {
	return &logrusWriter{entry: Entry()}
}

// SetVerbose sets the log level with respect to verbose flag.
func SetVerbose(verbose bool) {
	if verbose {
		//Logger().Debugf("logging set to level: %s", level)
		logrus.SetLevel(logrus.DebugLevel)
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

// RegisterHook registers a logrus hook
func RegisterHook(hook logrus.Hook) {
	logrus.AddHook(hook)
}

func RegisterSecret(secret string) {
	if len(secret) > 0 {
		secrets = append(secrets, secret)
	}
}
