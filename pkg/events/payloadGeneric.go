package events

import (
	"encoding/json"
	"maps"

	"github.com/SAP/jenkins-library/pkg/log"
)

type PayloadGeneric struct {
	JSONData string
}

func (e *PayloadGeneric) Merge(otherJSONData string) {
	if otherJSONData == "" {
		return
	}
	if e.JSONData == "" {
		e.JSONData = otherJSONData
		return
	}

	// read other data
	var otherJSONObj map[string]interface{}
	err := json.Unmarshal([]byte(otherJSONData), &otherJSONObj)
	if err != nil {
		log.Entry().WithError(err).Error("Failed to unmarshal additional data")
		return
	}
	// read existing data
	var newDataObj map[string]interface{}
	err = json.Unmarshal([]byte(e.JSONData), &newDataObj)
	if err != nil {
		log.Entry().WithError(err).Error("Failed to unmarshal existing event data")
		return
	}
	// merge
	maps.Copy(newDataObj, otherJSONObj)
	// write back
	jsonBytes, err := json.Marshal(newDataObj)
	if err != nil {
		log.Entry().WithError(err).Error("Failed to marshal merged event data")
		return
	}
	e.JSONData = string(jsonBytes)
}

func (e *PayloadGeneric) ToJSON() string {
	return e.JSONData
}
