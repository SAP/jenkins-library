package telemetry

import (
	"encoding/json"
	"net/url"
)

// BaseData ...
type BaseData struct {
	ActionName        string `json:"actionName"`
	EventType         string `json:"eventType"`
	SiteID            string `json:"idsite"`
	URL               string `json:"url"`
	GitOwner          string `json:"e_a"` // first custom field name is indeed e_a, not e_1
	GitRepository     string `json:"e_2"`
	StepName          string `json:"e_3"`
	PipelineURLSha1   string `json:"e_4"` // defaults to env.JOB_URl
	BuildURLSha1      string `json:"e_5"` // defaults to env.BUILD_URL
	GitPathSha1       string `json:"e_6"`
	GitOwnerSha1      string `json:"e_7"`
	GitRepositorySha1 string `json:"e_8"`
	JobName           string `json:"e_9"`
	StageName         string `json:"e_10"`
}

// BaseMetaData
type BaseMetaData struct {
	GitOwnerLabel          string `json:"custom_1"`
	GitRepositoryLabel     string `json:"custom_2"`
	StepNameLabel          string `json:"custom_3"`
	PipelineURLSha1Label   string `json:"custom_4"`
	BuildURLSha1Label      string `json:"custom_5"`
	GitPathSha1Label       string `json:"custom_6"`
	GitOwnerSha1Label      string `json:"custom_7"`
	GitRepositorySha1Label string `json:"custom_8"`
	JobNameLabel           string `json:"custom_9"`
	StageNameLabel         string `json:"custom_10"`
	BuildToolLabel         string `json:"custom_11"`
	ScanTypeLabel          string `json:"custom_24"`
}

// CustomData ...
type CustomData struct {
	BuildTool string `json:"e_11"`
	// ...
	ScanType      string `json:"e_24"`
	Custom25      string `json:"e_25"`
	Custom26      string `json:"e_26"`
	Custom27      string `json:"e_27"`
	Custom28      string `json:"e_28"`
	Custom29      string `json:"e_29"`
	Custom30      string `json:"e_30"`
	custom25Label string `json:"custom_25"`
	custom26Label string `json:"custom_26"`
	custom27Label string `json:"custom_27"`
	custom28Label string `json:"custom_28"`
	custom29Label string `json:"custom_29"`
	Custom30Label string `json:"custom_30"`
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
