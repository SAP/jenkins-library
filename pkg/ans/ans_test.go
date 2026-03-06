//go:build unit
// +build unit

package ans

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SAP/jenkins-library/pkg/xsuaa"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func (e *Examinee) initRun(onRequest func(rw http.ResponseWriter, req *http.Request)) {
	e.init()
	e.onRequest = onRequest
}

func TestANS_Send(t *testing.T) {
	examinee := Examinee{}
	defer examinee.finish()

	eventDefault := Event{EventType: "my event", EventTimestamp: 1647526655}

	t.Run("good", func(t *testing.T) {
		t.Run("pass request attributes", func(t *testing.T) {
			examinee.initRun(func(rw http.ResponseWriter, req *http.Request) {
				assert.Equal(t, http.MethodPost, req.Method, "Mismatch in requested method")
				assert.Equal(t, "/cf/producer/v1/resource-events", req.URL.Path, "Mismatch in requested path")
				assert.Equal(t, "bearer 1234", req.Header.Get(authHeaderKey), "Mismatch in requested auth header")
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"), "Mismatch in requested content type header")
			})
			examinee.ans.Send(eventDefault)
		})
		t.Run("pass request attribute event", func(t *testing.T) {
			examinee.initRun(func(rw http.ResponseWriter, req *http.Request) {
				eventBody, _ := io.ReadAll(req.Body)
				event := &Event{}
				json.Unmarshal(eventBody, event)
				assert.Equal(t, eventDefault, *event, "Mismatch in requested event body")
			})
			examinee.ans.Send(eventDefault)
		})
		t.Run("on status 202", func(t *testing.T) {
			examinee.initRun(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusAccepted)
			})
			err := examinee.ans.Send(eventDefault)
			require.NoError(t, err, "No error expected.")
		})
	})

	t.Run("bad", func(t *testing.T) {
		t.Run("on status 400", func(t *testing.T) {
			examinee.initRun(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte("an error occurred"))
			})
			err := examinee.ans.Send(eventDefault)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "Did not get expected status code 202")
		})
	})
}

func TestANS_CheckCorrectSetup(t *testing.T) {
	examinee := Examinee{}
	defer examinee.finish()

	t.Run("good", func(t *testing.T) {
		t.Run("pass request attributes", func(t *testing.T) {
			examinee.initRun(func(rw http.ResponseWriter, req *http.Request) {
				assert.Equal(t, http.MethodGet, req.Method, "Mismatch in requested method")
				assert.Equal(t, "/cf/consumer/v1/matched-events", req.URL.Path, "Mismatch in requested path")
				assert.Equal(t, "bearer 1234", req.Header.Get(authHeaderKey), "Mismatch in requested auth header")
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"), "Mismatch in requested content type header")
			})
			examinee.ans.CheckCorrectSetup()
		})
		t.Run("on status 200", func(t *testing.T) {
			examinee.initRun(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			})
			err := examinee.ans.CheckCorrectSetup()
			require.NoError(t, err, "No error expected.")
		})
	})

	t.Run("bad", func(t *testing.T) {
		t.Run("on status 400", func(t *testing.T) {
			examinee.initRun(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte("an error occurred"))
			})
			err := examinee.ans.CheckCorrectSetup()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "Did not get expected status code 200")
		})
	})
}

func TestANS_UnmarshallServiceKey(t *testing.T) {
	t.Parallel()

	serviceKeyJSONDefault := `{"url": "https://my.test.backend", "client_id": "myTestClientID", "client_secret": "super secret", "oauth_url": "https://my.test.oauth.provider"}`
	serviceKeyDefault := ServiceKey{Url: "https://my.test.backend", ClientId: "myTestClientID", ClientSecret: "super secret", OauthUrl: "https://my.test.oauth.provider"}

	t.Run("good", func(t *testing.T) {
		t.Run("Proper event JSON yields correct event", func(t *testing.T) {
			serviceKey, err := UnmarshallServiceKeyJSON(serviceKeyJSONDefault)
			require.NoError(t, err, "No error expected.")
			assert.Equal(t, serviceKeyDefault, serviceKey, "Got the wrong ans serviceKey")
		})
	})

	t.Run("bad", func(t *testing.T) {
		t.Run("JSON key data is an invalid string", func(t *testing.T) {
			serviceKeyDesc := `invalid descriptor`
			_, err := UnmarshallServiceKeyJSON(serviceKeyDesc)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "error unmarshalling ANS serviceKey")
		})
	})
}

func TestANS_readResponseBody(t *testing.T) {
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
			} else {
				require.NoError(t, err, "No error expected")
				assert.Equal(t, tt.want, got, "Did not receive expected body")
			}
		})
	}
}

func TestANS_SetServiceKey(t *testing.T) {
	t.Run("ServiceKey sets ANS fields", func(t *testing.T) {
		gotANS := &ANS{}
		serviceKey := ServiceKey{Url: "https://my.test.backend", ClientId: "myTestClientID", ClientSecret: "super secret", OauthUrl: "https://my.test.oauth.provider"}
		gotANS.SetServiceKey(serviceKey)
		wantANS := &ANS{
			XSUAA: xsuaa.XSUAA{
				OAuthURL:     "https://my.test.oauth.provider",
				ClientID:     "myTestClientID",
				ClientSecret: "super secret",
			},
			URL: "https://my.test.backend",
		}
		assert.Equal(t, wantANS, gotANS)
	})
}
