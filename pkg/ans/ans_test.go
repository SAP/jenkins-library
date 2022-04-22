package ans

import (
	"encoding/json"
	"github.com/SAP/jenkins-library/pkg/xsuaa"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Examinee struct {
	xsuaa     *xsuaa.XSUAA
	server    *httptest.Server
	ans       *ANS
	onRequest func(rw http.ResponseWriter, req *http.Request)
}

func (e *Examinee) request(rw http.ResponseWriter, req *http.Request) {
	e.onRequest(rw, req)
}

func (e *Examinee) finish() {
	if e.server != nil {
		e.server.Close()
		e.server = nil
	}
}

func (e *Examinee) init() {
	if e.xsuaa == nil {
		e.xsuaa = &xsuaa.XSUAA{
			OAuthURL:     "https://my.test.oauth.provider",
			ClientID:     "myTestClientID",
			ClientSecret: "super secret",
			CachedAuthToken: xsuaa.AuthToken{
				TokenType:   "bearer",
				AccessToken: "1234",
				ExpiresIn:   12345,
			},
		}
	}
	if e.server == nil {
		e.server = httptest.NewServer(http.HandlerFunc(e.request))
	}
	if e.ans == nil {
		e.ans = &ANS{XSUAA: *e.xsuaa, URL: e.server.URL}
	}
}

func (e *Examinee) execute(event Event, onRequest func(rw http.ResponseWriter, req *http.Request)) error {
	e.init()
	e.onRequest = onRequest
	return e.ans.Send(event)
}

func TestANS_Send(t *testing.T) {
	examinee := Examinee{}
	defer examinee.finish()
	examinee.init()

	eventDefault := Event{EventType: "my event", EventTimestamp: 1647526655}

	t.Run("good", func(t *testing.T) {
		t.Run("pass request attributes", func(t *testing.T) {
			examinee.execute(eventDefault, func(rw http.ResponseWriter, req *http.Request) {
				assert.Equal(t, http.MethodPost, req.Method, "Mismatch in requested method")
				assert.Equal(t, "/cf/producer/v1/resource-events", req.URL.Path, "Mismatch in requested path")
				assert.Equal(t, "bearer 1234", req.Header.Get(authHeaderKey), "Mismatch in requested auth header")
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"), "Mismatch in requested content type header")
			})
		})
		t.Run("pass request attribute event", func(t *testing.T) {
			examinee.execute(eventDefault, func(rw http.ResponseWriter, req *http.Request) {
				eventBody, _ := ioutil.ReadAll(req.Body)
				event := &Event{}
				json.Unmarshal(eventBody, event)
				assert.Equal(t, eventDefault, *event, "Mismatch in requested event body")
			})
		})
		t.Run("on status 202", func(t *testing.T) {
			err := examinee.execute(eventDefault, func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusAccepted)
			})
			require.NoError(t, err, "No error expected.")
		})
	})

	t.Run("bad", func(t *testing.T) {
		t.Run("on status 400", func(t *testing.T) {
			err := examinee.execute(eventDefault, func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte("an error occurred"))
			})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "Did not get expected status code 202")
		})
	})
}

func TestUnmarshallServiceKey(t *testing.T) {
	tests := []struct {
		name              string
		serviceKeyJSON    string
		wantAnsServiceKey ServiceKey
		wantErr           bool
	}{
		{
			name: "Proper event JSON yields correct event",
			serviceKeyJSON: `{
						"url": "https://my.test.backend",
						"client_id": "myTestClientID",
						"client_secret": "super secret",
						"oauth_url": "https://my.test.oauth.provider"
					}`,
			wantAnsServiceKey: ServiceKey{
				Url:          "https://my.test.backend",
				ClientId:     "myTestClientID",
				ClientSecret: "super secret",
				OauthUrl:     "https://my.test.oauth.provider",
			},
			wantErr: false,
		},
		{
			name:           "Faulty JSON yields error",
			serviceKeyJSON: `bli-da-blup`,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAnsServiceKey, err := UnmarshallServiceKeyJSON(tt.serviceKeyJSON)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshallServiceKeyJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.wantAnsServiceKey, gotAnsServiceKey, "Got the wrong ans serviceKey")
		})
	}
}

func Test_readResponseBody(t *testing.T) {
	tests := []struct {
		name        string
		response    *http.Response
		want        []byte
		wantErrText string
	}{
		{
			name:     "Straight forward",
			response: httpmock.NewStringResponse(200, "test string"),
			want:     []byte("test string"),
		},
		{
			name:        "No response error",
			wantErrText: "did not retrieve an HTTP response",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readResponseBody(tt.response)
			if tt.wantErrText != "" {
				require.Error(t, err, "Error expected")
				assert.EqualError(t, err, tt.wantErrText, "Error is not equal")
				return
			}
			require.NoError(t, err, "No error expected")
			assert.Equal(t, tt.want, got, "Did not receive expected body")
		})
	}
}
