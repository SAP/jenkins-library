package pact

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type pactUtilsMock struct {
	*mock.ExecMockRunner
	*mock.FilesMock
	calledMethod string
	calledUrl string
	user string
	pwd string
	body io.Reader
	httpSendErrors map[string]error
	httpResponseStatusCode int
	httpResponseContent string
	pbLinks LatestPactsForProviderTagResp
}

func (p *pactUtilsMock) SendRequest(method string, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	p.calledMethod = method
	p.calledUrl = url
	p.body = body
	var b []byte
	if p.httpResponseContent == "" {
		b, _ = json.Marshal(p.pbLinks)
	} else {
		b = []byte(p.httpResponseContent)
	}

	if p.httpResponseStatusCode == 0 {
		p.httpResponseStatusCode = http.StatusOK
	}

	resp := http.Response{
		StatusCode: p.httpResponseStatusCode,
		Body: io.NopCloser(bytes.NewBuffer(b)),
	}
	return &resp, p.httpSendErrors[url]
}

func (p *pactUtilsMock) SetOptions(options piperhttp.ClientOptions) {
	p.user = options.Username
	p.pwd = options.Password
}

func NewPactUtilsMock() *pactUtilsMock{
	utils := pactUtilsMock{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return &utils
}

func TestNewPactBrokerClient(t *testing.T) {
	t.Parallel()

	b := NewPactBrokerClient("testHost", "testUser", "testPassword")
	assert.Equal(t, "testHost", b.hostname)
	assert.Equal(t, "testUser", b.brokerUser)
	assert.Equal(t, "testPassword", b.brokerPass)
}

func TestLatestPactsForProviderByTag(t *testing.T) {
	t.Parallel()

	t.Run("success - default", func(t *testing.T){
		c := NewPactBrokerClient("testhost", "testuser", "testpassword")
		mockUtils := NewPactUtilsMock()
		testLinks := LatestPactsForProviderTagResp{
			Links: Links{
				PBPacts: []Link{
					{HRef: "https://link.1", Title: "contract 1", Name: "1"},
					{HRef: "https://link.2", Title: "contract 2", Name: "2"},
				},
			},
		}
		mockUtils.pbLinks = testLinks
		links, err := c.LatestPactsForProviderByTag("testProvider", "tag", mockUtils)

		assert.NoError(t, err)
		assert.Equal(t, testLinks, *links)
		assert.Equal(t, "https://testhost/pacts/provider/testProvider/latest/tag", mockUtils.calledUrl)
		assert.Equal(t, "testuser", mockUtils.user)
		assert.Equal(t, "testpassword", mockUtils.pwd)
		assert.Equal(t, http.MethodGet, mockUtils.calledMethod)
	})

	t.Run("success - no tests", func(t *testing.T){
		c := NewPactBrokerClient("testhost", "testuser", "testpassworf")
		mockUtils := NewPactUtilsMock()
		_, err := c.LatestPactsForProviderByTag("testProvider", "tag", mockUtils)

		assert.NoError(t, err)
	})

	t.Run("failure - not found", func(t *testing.T){
		c := NewPactBrokerClient("testhost", "testuser", "testpassworf")
		mockUtils := NewPactUtilsMock()
		mockUtils.httpSendErrors = map[string]error{"https://testhost/pacts/provider/testProvider/latest/tag": ErrNotFound}
		_, err := c.LatestPactsForProviderByTag("testProvider", "tag", mockUtils)

		assert.EqualError(t, err, fmt.Sprint(ErrNotFound))
	})

	t.Run("failure - get tests", func(t *testing.T){
		c := NewPactBrokerClient("testhost", "testuser", "testpassworf")
		mockUtils := NewPactUtilsMock()
		mockUtils.httpSendErrors = map[string]error{"https://testhost/pacts/provider/testProvider/latest/tag": fmt.Errorf("send failure")}
		_, err := c.LatestPactsForProviderByTag("testProvider", "tag", mockUtils)

		assert.EqualError(t, err, "send failure")
	})

	t.Run("failure - handle json", func(t *testing.T){
		c := NewPactBrokerClient("testhost", "testuser", "testpassworf")
		mockUtils := NewPactUtilsMock()
		mockUtils.httpResponseContent = "{"
		_, err := c.LatestPactsForProviderByTag("testProvider", "tag", mockUtils)

		assert.Contains(t, fmt.Sprint(err), "failed to unmarshal response:")
	})
}

func TestDownloadPactContract(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T){
		c := NewPactBrokerClient("testhost", "testuser", "testpassword")
		mockUtils := NewPactUtilsMock()
		mockUtils.httpResponseContent = `{"dummy":"content"}`
		pactContract, err := c.DownloadPactContract("http://the.url", mockUtils)

		assert.NoError(t, err)
		assert.Equal(t, []byte(mockUtils.httpResponseContent), pactContract)
		assert.Equal(t, "http://the.url", mockUtils.calledUrl)
		assert.Equal(t, "testuser", mockUtils.user)
		assert.Equal(t, "testpassword", mockUtils.pwd)
		assert.Equal(t, http.MethodGet, mockUtils.calledMethod)

	})

	t.Run("failure", func(t *testing.T){
		c := NewPactBrokerClient("testhost", "testuser", "testpassword")
		mockUtils := NewPactUtilsMock()
		mockUtils.httpSendErrors = map[string]error{"http://the.url": fmt.Errorf("send failure")}
		_, err := c.DownloadPactContract("http://the.url", mockUtils)

		assert.EqualError(t, err, "send failure")
	})
}
