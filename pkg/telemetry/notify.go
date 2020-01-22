package telemetry

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/sirupsen/logrus"
)

// WARNING ...
const WARNING = "WARNING"

// ERROR ...
const ERROR = "ERROR"

// Notify ...
func Notify(level, message string) {
	data := CustomData{}
	SendTelemetry(&data)

	notification := log.Entry().WithField("type", "notification")

	switch level {
	case WARNING:
		notification.Warning(message)
	case ERROR:
		notification.Fatal(message)
	}
}

// Fire ...
func (t *BaseData) Fire(entry *logrus.Entry) error {
	telemetryData := CustomData{}
	SendTelemetry(&telemetryData)
	return nil
}

// Levels ...
func (t *BaseData) Levels() (levels []logrus.Level) {
	levels = append(levels, logrus.ErrorLevel)
	levels = append(levels, logrus.FatalLevel)
	//levels = append(levels, logrus.PanicLevel)
	return
}
