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

// InitLogger sets the stepName field and sets log level with respect to verbose flag.
func InitLogger(stepName string, verbose bool) error {
	if verbose {
		//Logger().Debugf("logging set to level: %s", level)
		logrus.SetLevel(logrus.DebugLevel)
	}
	logger = Entry().WithField("stepName", stepName)
	return nil
}

// WriterWithErrorLevel returns a writer that logs with ERROR level.
func WriterWithErrorLevel() *io.PipeWriter {
	return Entry().WriterLevel(logrus.ErrorLevel)
}
