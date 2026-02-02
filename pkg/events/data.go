package events

import (
	"encoding/json"

	"github.com/SAP/jenkins-library/pkg/log"
)

type TaskRunEventData struct {
	TaskName  string `json:"taskName"`
	StageName string `json:"stageName"`
	Outcome   string `json:"outcome"`
}

func (data *TaskRunEventData) toJSON() string {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		log.Entry().WithError(err).Error("Failed to marshal TaskRunEventData to JSON")
		return ""
	}
	return string(jsonBytes)
}
