package log

import (
	"io"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

// Entry returns the logger entry or creates one if none is present.
func Entry() *logrus.Entry {
	if logger == nil {
		logger = logrus.WithField("library", "piper-lib-os")
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

// SetStepName sets the stepName field.
func SetStepName(stepName string) {
	logger = Entry().WithField("stepName", stepName)
}

// WriterWithErrorLevel returns a writer that logs with ERROR level.
func WriterWithErrorLevel() *io.PipeWriter {
	return Entry().WriterLevel(logrus.ErrorLevel)
}
