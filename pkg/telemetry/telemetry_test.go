package telemetry

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
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
		Initialize(false, nil, "", "testStep")
		// assert
		assert.Equal(t, nil, client)
		assert.Equal(t, BaseData{}, baseData)
	})

	t.Run("", func(t *testing.T) {
		// init
		client = nil
		// test
		Initialize(true, nil, "", "testStep")
		// assert
		assert.NotEqual(t, nil, client)
		assert.Equal(t, "testStep", baseData.StepName)
	})
}
func TestSendTelemetry(t *testing.T) {
	t.Run("with disabled telemetry", func(t *testing.T) {
		// init
		mock = clientMock{}
		client = &mock
		disabled = true
		// test
		SendTelemetry(&CustomData{})
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
		SendTelemetry(&CustomData{
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
	t.Run("", func(t *testing.T) {
		/*

			os.Getenv = func Getenv(key string) string {
				return "someValue"
			}

		*/
		// init
		client = nil
		// test
		Initialize(true, nil, "", "testStep")
		// assert
		assert.Equal(t, "someHash", baseData.PipelineURLHash)
		assert.Equal(t, "someHash", baseData.BuildURLHash)
	})
	t.Run("without values", func(t *testing.T) {
		// init
		client = nil
		// test
		Initialize(true, nil, "", "testStep")
		// assert
		assert.Equal(t, "n/a", baseData.PipelineURLHash)
		assert.Equal(t, "n/a", baseData.BuildURLHash)
	})
}
