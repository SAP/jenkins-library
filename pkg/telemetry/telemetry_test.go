package telemetry

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
)

type clientMock struct {
	httpMethod string
	urlsCalled string
}

func (c *clientMock) SetOptions(opts piperhttp.ClientOptions) {}

func (c *clientMock) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	c.httpMethod = method
	c.urlsCalled = url

	return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte("")))}, nil
}

var mock clientMock

func TestInitialize(t *testing.T) {

	t.Run("with disabled telemetry", func(t *testing.T) {
		telemetryClient := Telemetry{}
		// init
		client = nil
		// test
		telemetryClient.Initialize(true, "testStep")
		// assert
		assert.Equal(t, nil, client)
		assert.Equal(t, BaseData{}, telemetryClient.baseData)
	})

	t.Run("", func(t *testing.T) {
		telemetryClient := Telemetry{}
		// init
		client = nil
		// test
		telemetryClient.Initialize(false, "testStep")
		// assert
		assert.NotEqual(t, nil, client)
		assert.Equal(t, "testStep", telemetryClient.baseData.StepName)
	})
}
func TestSend(t *testing.T) {
	t.Run("with disabled telemetry", func(t *testing.T) {
		telemetryClient := Telemetry{}
		// init
		mock = clientMock{}
		client = &mock
		disabled = true
		// test
		telemetryClient.SetData(&CustomData{})
		telemetryClient.Send()
		// assert
		assert.Equal(t, 0, len(mock.httpMethod))
		assert.Equal(t, 0, len(mock.urlsCalled))
	})

	t.Run("", func(t *testing.T) {
		telemetryClient := Telemetry{}
		// init
		mock = clientMock{}
		client = &mock
		disabled = false
		telemetryClient.baseData = BaseData{
			ActionName: "testAction",
		}
		// test
		telemetryClient.SetData(&CustomData{
			Custom1:      "test",
			Custom1Label: "label",
		})
		telemetryClient.Send()
		// assert
		assert.Equal(t, "GET", mock.httpMethod)
		assert.Contains(t, mock.urlsCalled, baseURL)
		assert.Contains(t, mock.urlsCalled, "custom26=label")
		assert.Contains(t, mock.urlsCalled, "e_26=test")
		assert.Contains(t, mock.urlsCalled, "action_name=testAction")
	})
}

func TestSetData(t *testing.T) {
	type args struct {
		customData *CustomData
	}
	tests := []struct {
		name string
		args args
		want Data
	}{
		{
			name: "Test",
			args: args{customData: &CustomData{
				Duration:        "100",
				ErrorCode:       "0",
				ErrorCategory:   "Undefined",
				PiperCommitHash: "abcd12345",
			},
			},
			want: Data{
				BaseData: BaseData{
					URL:             "",
					ActionName:      "",
					EventType:       "",
					StepName:        "TestCreateDataObject",
					SiteID:          "",
					PipelineURLHash: "",
					BuildURLHash:    "",
					Orchestrator:    "Unknown",
				},
				BaseMetaData: BaseMetaData{
					StepNameLabel:        "stepName",
					StageNameLabel:       "stageName",
					PipelineURLHashLabel: "pipelineUrlHash",
					BuildURLHashLabel:    "buildUrlHash",
					DurationLabel:        "duration",
					ExitCodeLabel:        "exitCode",
					ErrorCategoryLabel:   "errorCategory",
					OrchestratorLabel:    "orchestrator",
					PiperCommitHashLabel: "piperCommitHash",
				},
				CustomData: CustomData{
					Duration:        "100",
					ErrorCode:       "0",
					ErrorCategory:   "Undefined",
					PiperCommitHash: "abcd12345",
					Custom1Label:    "",
					Custom2Label:    "",
					Custom3Label:    "",
					Custom4Label:    "",
					Custom5Label:    "",
					Custom1:         "",
					Custom2:         "",
					Custom3:         "",
					Custom4:         "",
					Custom5:         "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			telemetryClient := Telemetry{}
			telemetryClient.Initialize(false, "TestCreateDataObject")
			telemetryClient.baseData = BaseData{
				URL:             "",
				ActionName:      "",
				EventType:       "",
				StepName:        "TestCreateDataObject",
				SiteID:          "",
				PipelineURLHash: "",
				BuildURLHash:    "",
				Orchestrator:    "Unknown",
			}
			telemetryClient.baseMetaData = baseMetaData
			telemetryClient.SetData(tt.args.customData)
			fmt.Println(telemetryClient.data)
			fmt.Println(tt.want)
			if !reflect.DeepEqual(telemetryClient.data, tt.want) {
				t.Errorf("CreateDataObject() t.data= %v, want %v", telemetryClient.data, tt.want)
			}
		})
	}
}
