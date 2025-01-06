package cmd

import (
	"encoding/json"
	"time"

	"github.com/SAP/jenkins-library/pkg/ans"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func ansSendEvent(config ansSendEventOptions, telemetryData *telemetry.CustomData) {
	err := runAnsSendEvent(&config, &ans.ANS{})
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAnsSendEvent(config *ansSendEventOptions, c ans.Client) error {
	ansServiceKey, err := ans.UnmarshallServiceKeyJSON(config.AnsServiceKey)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return err
	}
	c.SetServiceKey(ansServiceKey)

	event := ans.Event{
		EventType: config.EventType,
		Severity:  config.Severity,
		Category:  config.Category,
		Subject:   config.Subject,
		Body:      config.Body,
		Priority:  config.Priority,
		Tags:      config.Tags,
		Resource: &ans.Resource{
			ResourceName:     config.ResourceName,
			ResourceType:     config.ResourceType,
			ResourceInstance: config.ResourceInstance,
			Tags:             config.ResourceTags,
		},
	}

	if GeneralConfig.Verbose {
		eventJson, _ := json.MarshalIndent(event, "", "  ")
		log.Entry().Infof("Event details: %s", eventJson)
	}

	if err = event.Validate(); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return err
	}
	// We set the time
	event.EventTimestamp = time.Now().Unix()
	if err = c.Send(event); err != nil {
		log.SetErrorCategory(log.ErrorService)
	}
	return err
}
