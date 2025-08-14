package log

import (
	"fmt"
	"os"
	"strings"

	"github.com/SAP/jenkins-library/pkg/ans"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ANSHook is used to set the hook features for the logrus hook
type ANSHook struct {
	client        ans.Client
	eventTemplate ans.Event
	firing        bool
}

// Levels returns the supported log level of the hook.
func (ansHook *ANSHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.WarnLevel, logrus.ErrorLevel, logrus.PanicLevel, logrus.FatalLevel}
}

// Fire creates a new event from the logrus and sends an event to the ANS backend
func (ansHook *ANSHook) Fire(entry *logrus.Entry) (err error) {
	if ansHook.firing {
		return fmt.Errorf("ANS hook has already been fired")
	}
	ansHook.firing = true
	defer func() { ansHook.firing = false }()

	if len(strings.TrimSpace(entry.Message)) == 0 {
		return
	}
	var event ans.Event
	if event, err = ansHook.eventTemplate.Copy(); err != nil {
		return
	}

	logLevel := entry.Level
	event.SetSeverityAndCategory(logLevel)
	var stepName string
	if entry.Data["stepName"] != nil {
		stepName = fmt.Sprint(entry.Data["stepName"])
	} else {
		stepName = "n/a"
	}
	event.Tags["cicd:stepName"] = stepName
	if errorCategory := GetErrorCategory().String(); errorCategory != "undefined" {
		event.Tags["cicd:errorCategory"] = errorCategory
	}

	event.EventTimestamp = entry.Time.Unix()
	if event.Subject == "" {
		event.Subject = fmt.Sprintf("Step '%s' sends '%s'", stepName, event.Severity)
	}
	event.Body = entry.Message
	event.Tags["cicd:logLevel"] = logLevel.String()

	return ansHook.client.Send(event)
}

type registrationUtil interface {
	ans.Client
	registerHook(hook *ANSHook)
}

type registrationUtilImpl struct {
	ans.Client
}

func (u *registrationUtilImpl) registerHook(hook *ANSHook) {
	RegisterHook(hook)
}

func (u *registrationUtilImpl) registerSecret(secret string) {
	RegisterSecret(secret)
}

// RegisterANSHookIfConfigured creates a new ANS hook for logrus if it is configured and registers it
func RegisterANSHookIfConfigured(correlationID string) error {
	return registerANSHookIfConfigured(correlationID, &registrationUtilImpl{Client: &ans.ANS{}})
}

func registerANSHookIfConfigured(correlationID string, util registrationUtil) error {
	ansServiceKeyJSON := os.Getenv("PIPER_ansHookServiceKey")
	if len(ansServiceKeyJSON) == 0 {
		return nil
	}

	ansServiceKey, err := ans.UnmarshallServiceKeyJSON(ansServiceKeyJSON)
	if err != nil {
		return errors.Wrap(err, "cannot initialize SAP Alert Notification Service due to faulty serviceKey json")
	}
	RegisterSecret(ansServiceKey.ClientSecret)

	util.SetServiceKey(ansServiceKey)
	if err = util.CheckCorrectSetup(); err != nil {
		return errors.Wrap(err, "check http request to SAP Alert Notification Service failed; not setting up the ANS hook")
	}

	eventTemplate, err := setupEventTemplate(os.Getenv("PIPER_ansEventTemplate"), correlationID)
	if err != nil {
		return err
	}
	util.registerHook(&ANSHook{
		client:        util,
		eventTemplate: eventTemplate,
	})
	return nil
}

func setupEventTemplate(customerEventTemplate, correlationID string) (ans.Event, error) {
	event := ans.Event{
		EventType: "Piper",
		Tags:      map[string]interface{}{"ans:correlationId": correlationID, "ans:sourceEventId": correlationID},
		Resource: &ans.Resource{
			ResourceType: "Pipeline",
			ResourceName: "Pipeline",
		},
	}

	if len(customerEventTemplate) > 0 {
		if err := event.MergeWithJSON([]byte(customerEventTemplate)); err != nil {
			return ans.Event{}, errors.Wrapf(err, "provided SAP Alert Notification Service event template '%s' could not be unmarshalled", customerEventTemplate)
		}
	}
	if len(event.Severity) > 0 {
		Entry().WithField("stepName", "ANS").Warnf("event severity set to '%s' will be overwritten according to the log level", event.Severity)
		event.Severity = ""
	}
	if len(event.Category) > 0 {
		Entry().WithField("stepName", "ANS").Warnf("event category set to '%s' will be overwritten according to the log level", event.Category)
		event.Category = ""
	}
	if err := event.Validate(); err != nil {
		return ans.Event{}, errors.Wrap(err, "did not initialize SAP Alert Notification Service due to faulty event template json")
	}
	return event, nil
}
