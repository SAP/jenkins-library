package ans

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	authHeaderKey = "Authorization"

	infoSeverity    = "INFO"
	noticeSeverity  = "NOTICE"
	warningSeverity = "WARNING"
	errorSeverity   = "ERROR"
	fatalSeverity   = "FATAL"

	exceptionCategory    = "EXCEPTION"
	alertCategory        = "ALERT"
	notificationCategory = "NOTIFICATION"
)

// Event structure of the SAP Alert Notification Service
type Event struct {
	EventType      string                 `json:"eventType,omitempty"`
	EventTimestamp int64                  `json:"eventTimestamp,omitempty"`
	Severity       string                 `json:"severity,omitempty"`
	Category       string                 `json:"category,omitempty"`
	Subject        string                 `json:"subject,omitempty"`
	Body           string                 `json:"body,omitempty"`
	Priority       int                    `json:"priority,omitempty"`
	Tags           map[string]interface{} `json:"tags,omitempty"`
	Resource       *Resource              `json:"resource,omitempty"`
}

// Resource structure of the SAP Alert Notification Service Event
type Resource struct {
	ResourceName     string                 `json:"resourceName,omitempty"`
	ResourceType     string                 `json:"resourceType,omitempty"`
	ResourceInstance string                 `json:"resourceInstance,omitempty"`
	Tags             map[string]interface{} `json:"tags,omitempty"`
}

// MergeWithJSON unmarshalls an ANS Event JSON string and merges it with the existing receiver Event
func (event *Event) MergeWithJSON(eventJSON []byte) (err error) {
	err = json.Unmarshal(eventJSON, &event)
	if err != nil {
		err = errors.Wrapf(err, "error unmarshalling ANS event from JSON string %q", eventJSON)
		return
	}
	return
}

// TranslateLogrusLogLevel takes the logrus log level and translates it to an ANS severity ans category string
func TranslateLogrusLogLevel(level logrus.Level) (severity, category string) {
	severity = infoSeverity
	category = notificationCategory
	switch level {
	case logrus.InfoLevel:
		severity = infoSeverity
		category = notificationCategory
	case logrus.DebugLevel:
		severity = infoSeverity
		category = notificationCategory
	case logrus.WarnLevel:
		severity = warningSeverity
		category = alertCategory
	case logrus.ErrorLevel:
		severity = errorSeverity
		category = exceptionCategory
	case logrus.FatalLevel:
		severity = fatalSeverity
		category = exceptionCategory
	case logrus.PanicLevel:
		severity = fatalSeverity
		category = exceptionCategory
	}
	return
}
