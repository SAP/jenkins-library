package ans

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	validate *validator.Validate
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
	EventType      string         `json:"eventType,omitempty"`
	EventTimestamp int64          `json:"eventTimestamp,omitempty" validate:"omitempty,min=0"`
	Severity       string         `json:"severity,omitempty" validate:"omitempty,oneof=INFO NOTICE WARNING ERROR FATAL"`
	Category       string         `json:"category,omitempty" validate:"omitempty,oneof=EXCEPTION ALERT NOTIFICATION"`
	Subject        string         `json:"subject,omitempty"`
	Body           string         `json:"body,omitempty"`
	Priority       int            `json:"priority,omitempty" validate:"omitempty,min=1,max=1000"`
	Tags           map[string]any `json:"tags,omitempty"`
	Resource       *Resource      `json:"resource,omitempty"`
}

// Resource structure of the SAP Alert Notification Service Event
type Resource struct {
	ResourceName     string         `json:"resourceName,omitempty"`
	ResourceType     string         `json:"resourceType,omitempty"`
	ResourceInstance string         `json:"resourceInstance,omitempty"`
	Tags             map[string]any `json:"tags,omitempty"`
}

// MergeWithJSON unmarshalls an ANS Event JSON string and merges it with the existing receiver Event
func (event *Event) MergeWithJSON(eventJSON []byte) (err error) {
	if err = strictUnmarshal(eventJSON, &event); err != nil {
		return fmt.Errorf("error unmarshalling ANS event from JSON string %q: %w", eventJSON, err)
	}
	return
}

// Validate will validate the Event according to the 'validate' tags in the struct
func (event *Event) Validate() (err error) {
	validate = validator.New()

	if err = validate.Struct(event); err != nil {
		errs := err.(validator.ValidationErrors)
		err = fmt.Errorf("event JSON failed the validation")
		for _, fe := range errs {
			err = fmt.Errorf("%s: %w", translateFieldError(fe), err)
		}
	}
	return
}

// Copy will copy an ANS Event
func (event *Event) Copy() (destination Event, err error) {
	var sourceJSON []byte
	if sourceJSON, err = json.Marshal(event); err != nil {
		return
	}
	err = destination.MergeWithJSON(sourceJSON)
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

func strictUnmarshal(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

var enPrinter = message.NewPrinter(language.English)

func translateFieldError(fe validator.FieldError) string {
	field := fe.Field()
	param := fe.Param()
	switch fe.Tag() {
	case "oneof":
		values := strings.ReplaceAll(param, " ", " ")
		return fmt.Sprintf("%s must be one of [%s]", field, values)
	case "min":
		n, err := strconv.ParseInt(param, 10, 64)
		if err != nil {
			return fmt.Sprintf("%s must be %s or greater", field, param)
		}
		return fmt.Sprintf("%s must be %s or greater", field, enPrinter.Sprintf("%d", n))
	case "max":
		n, err := strconv.ParseInt(param, 10, 64)
		if err != nil {
			return fmt.Sprintf("%s must be %s or less", field, param)
		}
		return fmt.Sprintf("%s must be %s or less", field, enPrinter.Sprintf("%d", n))
	default:
		return fe.Error()
	}
}
