package telemetry

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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

func TestInitialise(t *testing.T) {
	t.Run("with disabled telemetry", func(t *testing.T) {
		// init
		client = nil
		// test
		Initialize(true, "testStep")
		// assert
		assert.Equal(t, nil, client)
		assert.Equal(t, BaseData{}, baseData)
	})

	t.Run("", func(t *testing.T) {
		// init
		client = nil
		// test
		Initialize(false, "testStep")
		// assert
		assert.NotEqual(t, nil, client)
		assert.Equal(t, "testStep", baseData.StepName)
	})
}
func TestSend(t *testing.T) {
	t.Run("with disabled telemetry", func(t *testing.T) {
		// init
		mock = clientMock{}
		client = &mock
		disabled = true
		// test
		Send(&CustomData{})
		// assert
		assert.Equal(t, 0, len(mock.httpMethod))
		assert.Equal(t, 0, len(mock.urlsCalled))
	})

	t.Run("", func(t *testing.T) {
		// init
		mock = clientMock{}
		client = &mock
		disabled = false
		baseData = BaseData{
			ActionName: "testAction",
		}
		// test
		Send(&CustomData{
			Custom1:      "test",
			Custom1Label: "label",
		})
		// assert
		assert.Equal(t, "GET", mock.httpMethod)
		assert.Contains(t, mock.urlsCalled, baseURL)
		assert.Contains(t, mock.urlsCalled, "custom26=label")
		assert.Contains(t, mock.urlsCalled, "e_26=test")
		assert.Contains(t, mock.urlsCalled, "action_name=testAction")
	})
}
func TestEnvVars(t *testing.T) {
	t.Run("without values", func(t *testing.T) {
		// init
		client = nil
		// test
		Initialize(false, "testStep")
		// assert
		assert.Equal(t, "n/a", baseData.PipelineURLHash)
		assert.Equal(t, "n/a", baseData.BuildURLHash)
	})

	t.Run("", func(t *testing.T) {
		// init
		os.Setenv("JOB_URL", "someValue")
		os.Setenv("BUILD_URL", "someValue")
		client = nil
		// test
		Initialize(false, "testStep")
		// assert
		assert.Equal(t, "c1353b55ce4db511684b8a3b7b5c4b3d99ee9dec", baseData.PipelineURLHash)
		assert.Equal(t, "c1353b55ce4db511684b8a3b7b5c4b3d99ee9dec", baseData.BuildURLHash)
		// cleanup
		os.Unsetenv("JOB_URL")
		os.Unsetenv("BUILD_URL")
	})
}

func TestGetData(t *testing.T) {
	type args struct {
		customData *CustomData
	}
	tests := []struct {
		name string
		args args
		want Data
	}{
		{
			"Test",
			args{customData: &CustomData{
				Duration:      "100",
				ErrorCode:     "0",
				ErrorCategory: "Undefined",
			},
			},
			Data{
				BaseData: BaseData{
					ActionName:      "",
					EventType:       "",
					SiteID:          "",
					URL:             "",
					StepName:        "",
					StageName:       "",
					PipelineURLHash: "",
					BuildURLHash:    "",
				},
				BaseMetaData: BaseMetaData{
					StepNameLabel:        "stepName",
					StageNameLabel:       "stageName",
					PipelineURLHashLabel: "pipelineUrlHash",
					BuildURLHashLabel:    "buildUrlHash",
					DurationLabel:        "duration",
					ExitCodeLabel:        "exitCode",
					ErrorCategoryLabel:   "errorCategory",
				},
				CustomData: CustomData{
					Duration:      "100",
					ErrorCode:     "0",
					ErrorCategory: "Undefined",
					Custom1Label:  "",
					Custom2Label:  "",
					Custom3Label:  "",
					Custom4Label:  "",
					Custom5Label:  "",
					Custom1:       "",
					Custom2:       "",
					Custom3:       "",
					Custom4:       "",
					Custom5:       "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseData = BaseData{
				ActionName:      "",
				EventType:       "",
				SiteID:          "",
				URL:             "",
				StepName:        "",
				StageName:       "",
				PipelineURLHash: "",
				BuildURLHash:    "",
			}
			if got := GetData(tt.args.customData); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetData() = %v, want %v", got, tt.want)
			}
		})
	}
}
