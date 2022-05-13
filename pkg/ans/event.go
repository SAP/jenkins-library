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
	if err = json.Unmarshal(eventJSON, &event); err != nil {
		err = errors.Wrapf(err, "error unmarshalling ANS event from JSON string %q", eventJSON)
	}
	return
}

// SetSeverityAndCategory takes the logrus log level and sets the corresponding ANS severity and category string
func (event *Event) SetSeverityAndCategory(level logrus.Level) {
	switch level {
	case logrus.InfoLevel:
		event.Severity = infoSeverity
		event.Category = notificationCategory
	case logrus.DebugLevel:
		event.Severity = infoSeverity
		event.Category = notificationCategory
	case logrus.WarnLevel:
		event.Severity = warningSeverity
		event.Category = alertCategory
	case logrus.ErrorLevel:
		event.Severity = errorSeverity
		event.Category = exceptionCategory
	case logrus.FatalLevel:
		event.Severity = fatalSeverity
		event.Category = exceptionCategory
	case logrus.PanicLevel:
		event.Severity = fatalSeverity
		event.Category = exceptionCategory
	}
}
