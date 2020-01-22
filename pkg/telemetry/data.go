package telemetry

import (
	"encoding/json"
	"net/url"
)

// BaseData ...
type BaseData struct {
	ActionName string `json:"action_name"`
	EventType  string `json:"event_type"`
	SiteID     string `json:"idsite"`
	URL        string `json:"url"`
	//GitOwner          string `json:"e_a"` // first custom field name is indeed e_a, not e_1
	//GitRepository     string `json:"e_2"`
	StepName        string `json:"e_3"`
	PipelineURLSha1 string `json:"e_4"` // defaults to env.JOB_URl
	BuildURLSha1    string `json:"e_5"` // defaults to env.BUILD_URL
	//GitPathSha1       string `json:"e_6"`
	//GitOwnerSha1      string `json:"e_7"`
	//GitRepositorySha1 string `json:"e_8"`
	//JobName           string `json:"e_9"`
	StageName string `json:"e_10"`
}

var baseData BaseData

// BaseMetaData ...
type BaseMetaData struct {
	//GitOwnerLabel          string `json:"custom_1"`
	//GitRepositoryLabel     string `json:"custom_2"`
	StepNameLabel        string `json:"custom_3"`
	PipelineURLSha1Label string `json:"custom_4"`
	BuildURLSha1Label    string `json:"custom_5"`
	//GitPathSha1Label       string `json:"custom_6"`
	//GitOwnerSha1Label      string `json:"custom_7"`
	//GitRepositorySha1Label string `json:"custom_8"`
	//JobNameLabel           string `json:"custom_9"`
	StageNameLabel string `json:"custom_10"`
}

var baseMetaData BaseMetaData = BaseMetaData{
	//GitOwnerLabel:          "owner",
	//GitRepositoryLabel:     "repository",
	StepNameLabel:        "stepName",
	PipelineURLSha1Label: "",
	BuildURLSha1Label:    "",
	//GitPathSha1Label:       "gitpathsha1",
	//GitOwnerSha1Label:      "",
	//GitRepositorySha1Label: "",
	//JobNameLabel:           "jobName",
	StageNameLabel: "stageName",
}

// CustomData ...
type CustomData struct {
	Custom1Label string `json:"custom_11"`
	Custom2Label string `json:"custom_12"`
	Custom3Label string `json:"custom_13"`
	Custom4Label string `json:"custom_14"`
	Custom5Label string `json:"custom_15"`
	Custom1      string `json:"e_11"`
	Custom2      string `json:"e_12"`
	Custom3      string `json:"e_13"`
	Custom4      string `json:"e_14"`
	Custom5      string `json:"e_15"`
	//Custom30      string `json:"e_30"`
	//Custom30Label string `json:"custom_30"`
}

// Data ...
type Data struct {
	BaseData
	BaseMetaData
	CustomData
}

func (d *Data) toMap() (result map[string]string) {
	jsonObj, _ := json.Marshal(d)
	json.Unmarshal(jsonObj, &result)
	return
}

func (d *Data) toPayloadString() string {
	parameters := url.Values{}

	for key, value := range d.toMap() {
		if len(value) > 0 {
			parameters.Add(key, value)
		}
	}
	//TODO: Remove labels for empty fields?

	return parameters.Encode()
}
