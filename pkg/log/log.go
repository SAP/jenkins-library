package log

import (
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

func Logger() *logrus.Entry {
	err := initLogger()
	if err != nil {
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
