package log

import (
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/ans"
	"github.com/SAP/jenkins-library/pkg/xsuaa"
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
func NewANSHook(serviceKey, correlationID, eventTemplateFilePath string) ANSHook {
	ansServiceKey, err := ans.UnmarshallServiceKeyJSON(serviceKey)
	if err != nil {
		Entry().WithField("stepName", "ANS").Warnf("cannot initialize ans due to faulty serviceKey json: %v", err)
	}
	event := ans.Event{
		EventType: "Piper",
		Tags:      map[string]interface{}{"ans:correlationId": correlationID},
		Resource: &ans.Resource{
			ResourceType: "Piper",
			ResourceName: "Pipeline",
		},
	}
	if len(eventTemplateFilePath) > 0 {
		eventTemplate, err := ioutil.ReadFile(eventTemplateFilePath)
		if err != nil {
			Entry().WithField("stepName", "ANS").Warnf("provided ANS event template file with path '%s' could not be read: %v", eventTemplateFilePath, err)
		}
		err = event.MergeWithJSON(eventTemplate)
		if err != nil {
			Entry().WithField("stepName", "ANS").Warnf("provided ANS event template '%s' could not be unmarshalled: %v", eventTemplate, err)
		}
	}
	x := xsuaa.XSUAA{
		OAuthURL:     ansServiceKey.OauthUrl,
		ClientID:     ansServiceKey.ClientId,
		ClientSecret: ansServiceKey.ClientSecret,
	}
	h := ANSHook{
		client: ans.ANS{XSUAA: x, URL: ansServiceKey.Url},
		event:  event,
	}
	return h
}

// Levels returns the supported log level of the hook.
func (ansHook *ANSHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.InfoLevel, logrus.DebugLevel, logrus.WarnLevel, logrus.ErrorLevel, logrus.PanicLevel, logrus.FatalLevel}
}

// Fire creates a new event from the logrus and sends an event to the ANS backend
func (ansHook *ANSHook) Fire(entry *logrus.Entry) error {
	logLevel := entry.Level
	event, err := copyEvent(ansHook.event)
	if err != nil {
		return err
	}

	event.EventTimestamp = entry.Time.Unix()
	if event.Subject == "" {
		event.Subject = fmt.Sprint(entry.Data["stepName"])
	}
	if strings.HasPrefix(entry.Message, "fatal error") {
		logLevel = logrus.FatalLevel
	}
	event.Body = entry.Message
	for k, v := range entry.Data {
		if k == "error" {
			logLevel = logrus.ErrorLevel
		}
		event.Tags[k] = v
	}
	event.Severity, event.Category = ans.TranslateLogrusLogLevel(logLevel)
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
