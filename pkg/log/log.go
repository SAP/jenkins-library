package log

import (
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

func Logger() *logrus.Entry {
	if err := initLogger(); err != nil {
		logrus.Warnf("error initializing logrus %v", err)
	}
	return logger
}

func initLogger() error {
	if logger == nil {
		logger = logrus.WithField("library", "piper-lib-os")
	}
	return nil
}

func InitLogger(stepName string, verbose bool) error {
	initLogger()
	if verbose {
		//Logger().Debugf("logging set to level: %s", level)
		logrus.SetLevel(logrus.DebugLevel)
	}
	logger = logger.WithField("stepName", stepName)
	return nil
}
