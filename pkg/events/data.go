package events

import (
	"encoding/json"
	"maps"

	"github.com/SAP/jenkins-library/pkg/log"
)

type EventPayloadData struct {
	JSONData string
}

func (e *EventPayloadData) ToJSON() string {
	return e.JSONData
}

func (e *EventPayloadData) Merge(otherJSONData string) {
	if otherJSONData != "" {
		// read other data
		var otherJSONObj map[string]interface{}
		err := json.Unmarshal([]byte(otherJSONData), &otherJSONObj)
		if err != nil {
			log.Entry().WithError(err).Error("Failed to unmarshal additional data")
		}
		// read existing data
		var newDataObj map[string]interface{}
		err = json.Unmarshal([]byte(e.JSONData), &newDataObj)
		if err != nil {
			log.Entry().WithError(err).Error("Failed to unmarshal existing event data")
		}
		// merge
		maps.Copy(newDataObj, otherJSONObj)
		// write back
		jsonBytes, err := json.Marshal(newDataObj)
		if err != nil {
			log.Entry().WithError(err).Error("Failed to marshal merged event data")
		}
		e.JSONData = string(jsonBytes)
	}
}

type TaskRunEventPayloadData struct {
	TaskName  string `json:"taskName"`
	StageName string `json:"stageName"`
	Outcome   string `json:"outcome"`
}

func (data *TaskRunEventPayloadData) toJSON() string {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		log.Entry().WithError(err).Error("Failed to marshal TaskRunEventPayloadData to JSON")
		return ""
	}
	return string(jsonBytes)
}
