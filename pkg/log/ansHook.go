package log

import (
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/ans"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"strings"
)

// ANSHook is used to set the hook features for the logrus hook
type ANSHook struct {
	client ans.Client
	event  ans.Event
}

// NewANSHook creates a new ANS hook for logrus
func NewANSHook(config ans.Configuration, correlationID string) (hook ANSHook, err error) {
	return newANSHook(config, correlationID, &ans.ANS{})
}

func newANSHook(config ans.Configuration, correlationID string, client ans.Client) (hook ANSHook, err error) {
	ansServiceKey, err := ans.UnmarshallServiceKeyJSON(config.ServiceKey)
	if err != nil {
		err = errors.Wrap(err, "cannot initialize SAP Alert Notification Service due to faulty serviceKey json")
		return
	}
	client.SetOptions(ansServiceKey)

	event := ans.Event{
		EventType: "Piper",
		Tags:      map[string]interface{}{"ans:correlationId": correlationID, "ans:sourceEventId": correlationID},
		Resource: &ans.Resource{
			ResourceType: "Pipeline",
			ResourceName: "Pipeline",
		},
	}
	if len(config.EventTemplateFilePath) > 0 {
		eventTemplateString, err := ioutil.ReadFile(config.EventTemplateFilePath)
		if err != nil {
			Entry().WithField("stepName", "ANS").Warnf("provided SAP Alert Notification Service event template file with path '%s' could not be read: %v", config.EventTemplateFilePath, err)
		} else {
			err = event.MergeWithJSON(eventTemplateString)
			if err != nil {
				Entry().WithField("stepName", "ANS").Warnf("provided SAP Alert Notification Service event template '%s' could not be unmarshalled: %v", eventTemplateString, err)
			}
		}
	}
	if len(config.EventTemplate) > 0 {
		if err = event.MergeWithJSON([]byte(config.EventTemplate)); err != nil {
			Entry().WithField("stepName", "ANS").Warnf("provided SAP Alert Notification Service event template '%s' could not be unmarshalled: %v", config.EventTemplate, err)
		}
	}
	h := ANSHook{
		client: client,
		event:  event,
	}
	err = h.client.CheckCorrectSetup()
	if err != nil {
		return
	}
	return h, nil
}

// Levels returns the supported log level of the hook.
func (ansHook *ANSHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.WarnLevel, logrus.ErrorLevel, logrus.PanicLevel, logrus.FatalLevel}
}

// Fire creates a new event from the logrus and sends an event to the ANS backend
func (ansHook *ANSHook) Fire(entry *logrus.Entry) error {
	if len(strings.TrimSpace(entry.Message)) == 0 {
		return nil
	}
	event, err := copyEvent(ansHook.event)
	if err != nil {
		return err
	}

	logLevel := entry.Level
	for k, v := range entry.Data {
		event.Tags[k] = v
	}
	if errorCategory := GetErrorCategory().String(); errorCategory != "undefined" {
		event.Tags["errorCategory"] = errorCategory
	}

	event.EventTimestamp = entry.Time.Unix()
	if event.Subject == "" {
		event.Subject = fmt.Sprint(entry.Data["stepName"])
	}
	event.Body = entry.Message
	event.SetSeverityAndCategory(logLevel)
	event.Tags["logLevel"] = logLevel.String()

	err = ansHook.client.Send(event)
	if err != nil {
		return err
	}
	return nil
}

func copyEvent(source ans.Event) (destination ans.Event, err error) {
	sourceJSON, err := json.Marshal(source)
	if err != nil {
		return
	}
	err = destination.MergeWithJSON(sourceJSON)
	return
}
