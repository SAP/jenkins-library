package log

import (
	"time"

	"github.com/sirupsen/logrus"
)

// logCollectorHook provides a logrus hook which stores details about any logs of each step.
// This is helpful in order to transfer the logs and error details to an logging system as Splunk
// and by that make it possible to have a better investigation of any error.

type CollectorHook struct {
	CorrelationID string
	Messages      []Message
	// TODO: log Levels?
}

// Levels returns the supported log level of the hook.
func (f *CollectorHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.InfoLevel, logrus.DebugLevel, logrus.WarnLevel, logrus.ErrorLevel, logrus.PanicLevel, logrus.FatalLevel}
}

type Message struct {
	Time    time.Time    `json:"time,omitempty"`
	Level   logrus.Level `json:"level,omitempty"`
	Message string       `json:"message,omitempty"`
	Data    interface{}  `json:"data,omitempty"`
}

// Fire creates a new event from the logrus and stores it in the SplunkHook object
func (f *CollectorHook) Fire(entry *logrus.Entry) error {
	message := Message{
		Time:    entry.Time,
		Data:    entry.Data,
		Message: entry.Message,
	}
	f.Messages = append(f.Messages, message)
	return nil
}
