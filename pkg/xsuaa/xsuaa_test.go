package xsuaa

import (
	"encoding/base64"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestXSUAA_GetBearerToken(t *testing.T) {
	type (
		fields struct {
			ClientID     string
			ClientSecret string
		}
		want struct {
			authToken AuthToken
			errRegex  string
		}
		response struct {
			statusCode int
			bodyText   string
		}
	)
	tests := []struct {
		name         string
		fields       fields
		oauthUrlPath string
		want         want
		response     response
	}{
		{
			name: "Straight forward",
			fields: fields{
				ClientID:     "myClientID",
				ClientSecret: "secret",
			},
			want: want{
				authToken: AuthToken{
					TokenType:   "bearer",
					AccessToken: "1234",
					ExpiresIn:   9876,
				}},
			response: response{
				bodyText: `{"access_token": "1234", "expires_in": 9876, "token_type": "bearer"}`,
			},
		},
		{
			name: "No expiring duration",
			fields: fields{
				ClientID:     "myClientID",
				ClientSecret: "secret",
			},
			want: want{
				authToken: AuthToken{
					TokenType:   "bearer",
					AccessToken: "1234",
				}},
			response: response{
				bodyText: `{"access_token": "1234", "token_type": "bearer"}`,
			},
		},
		{
			name: "OAuth Url with path",
			fields: fields{
				ClientID:     "myClientID",
				ClientSecret: "secret",
			},
			oauthUrlPath: "/oauth/token?grant_type=client_credentials",
			want: want{
				authToken: AuthToken{
					TokenType:   "bearer",
					AccessToken: "1234",
					ExpiresIn:   9876,
				}},
			response: response{
				bodyText: `{"access_token": "1234", "expires_in": 9876, "token_type": "bearer"}`,
			},
		},
		{
			name: "No token type",
			fields: fields{
				ClientID:     "myClientID",
				ClientSecret: "secret",
			},
			want: want{
				authToken: AuthToken{
					TokenType:   "bearer",
					AccessToken: "1234",
					ExpiresIn:   9876,
				}},
			response: response{
				bodyText: `{"access_token": "1234", "expires_in": 9876}`,
			},
		},
		{
			name: "HTTP error",
			fields: fields{
				ClientID:     "myClientID",
				ClientSecret: "secret",
			},
			want: want{errRegex: `fetching an access token failed: HTTP GET request to .*/oauth/token\?grant_type=client_credentials&response_type=token ` +
				`failed: expected response code 200, got '401', response body: '{"error": "unauthorized"}'`},
			response: response{
				statusCode: 401,
				bodyText:   `{"error": "unauthorized"}`,
			},
		},
		{
			name: "Wrong response code",
			want: want{errRegex: `expected response code 200, got '201', response body: '{"success": "created"}'`},
			response: response{
				statusCode: 201,
				bodyText:   `{"success": "created"}`,
			},
		},
		{
			name: "No 'access_token' field in json response",
			want: want{errRegex: `expected authToken field 'access_token' in json response: got response body: '{"authToken": "1234"}'`},
			response: response{
				bodyText: `{"authToken": "1234"}`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestedUrlPath string
			var requestedAuthHeader string
			// Start a local HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				requestedUrlPath = req.URL.String()
				if tt.response.statusCode != 0 {
					rw.WriteHeader(tt.response.statusCode)
				}
				requestedAuthHeader = req.Header.Get(authHeaderKey)
				rw.Write([]byte(tt.response.bodyText))
			}))
			// Close the server when test finishes
			defer server.Close()

			oauthUrl := server.URL + tt.oauthUrlPath
			x := &XSUAA{
				OAuthURL:     oauthUrl,
				ClientID:     tt.fields.ClientID,
				ClientSecret: tt.fields.ClientSecret,
			}
			gotToken, err := x.GetBearerToken()
			if tt.want.errRegex != "" {
				require.Error(t, err, "Error expected")
				assert.Regexp(t, tt.want.errRegex, err.Error())
				return
			}
			require.NoError(t, err, "No error expected")
			assert.Equal(t, tt.want.authToken.TokenType, gotToken.TokenType, "Did not receive expected token type.")
			assert.Equal(t, tt.want.authToken.AccessToken, gotToken.AccessToken, "Did not receive expected access token.")
			if tt.want.authToken.ExpiresIn == 0 {
				assert.Equal(t, time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
					gotToken.ExpiresAt, "ExpiresAt should be date zero")
			} else {
				assert.NotEqual(t, time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
					gotToken.ExpiresAt, "ExpiresAt should be proper date")
			}
			wantUrlPath := "/oauth/token?grant_type=client_credentials&response_type=token"
			assert.Equal(t, wantUrlPath, requestedUrlPath)
			wantAuth := tt.fields.ClientID + ":" + tt.fields.ClientSecret
			assert.Equal(t, "Basic "+base64.StdEncoding.EncodeToString([]byte(wantAuth)), requestedAuthHeader)
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

func TestXSUAA_SetAuthHeaderIfNotPresent(t *testing.T) {
	type (
		fields struct {
			ClientID        string
			ClientSecret    string
			CachedAuthToken AuthToken
		}
		args struct {
			authHeader string
		}
		want struct {
			token    string
			errRegex string
		}
		response struct {
			statusCode int
			bodyText   string
		}
	)
	tests := []struct {
		name     string
		fields   fields
		args     args
		want     want
		response response
	}{
		{
			name: "Straight forward",
			fields: fields{
				ClientID:     "myClientID",
				ClientSecret: "secret",
			},
			want: want{token: "bearer 1234"},
			response: response{
				bodyText: `{"access_token": "1234", "expires_in": 9876, "token_type": "bearer"}`,
			},
		},
		{
			name: "Error case",
			fields: fields{
				ClientID:     "myClientID",
				ClientSecret: "secret",
			},
			want: want{errRegex: `fetching an access token failed: HTTP GET request to .*/oauth/token\?grant_type=client_credentials&response_type=token ` +
				`failed: expected response code 200, got '401', response body: '{"error": "unauthorized"}'`},
			response: response{
				statusCode: 401,
				bodyText:   `{"error": "unauthorized"}`,
			},
		},
		{
			name: "Missing field parameter",
			fields: fields{
				ClientID: "myClientID",
			},
			want: want{errRegex: `OAuthURL, ClientID and ClientSecret have to be set on the xsuaa instance`},
			response: response{
				statusCode: 401,
				bodyText:   `{"error": "unauthorized"}`,
			},
		},
		{
			name: "Different token type",
			fields: fields{
				ClientID:     "myClientID",
				ClientSecret: "secret",
			},
			want: want{token: "jwt 1234"},
			response: response{
				bodyText: `{"access_token": "1234", "expires_in": 9876, "token_type": "jwt"}`,
			},
		},
		{
			name: "Auth authHeader already set",
			fields: fields{
				ClientID:     "myClientID",
				ClientSecret: "secret",
			},
			args: args{authHeader: "basic eW91aGF2ZXRvb211Y2g6dGltZQ=="},
			want: want{token: "basic eW91aGF2ZXRvb211Y2g6dGltZQ=="},
			response: response{
				bodyText: `{"access_token": "1234", "expires_in": 9876, "token_type": "jwt"}`,
			},
		},
		{
			name: "Valid token skips getting a new one",
			fields: fields{
				ClientID:     "myClientID",
				ClientSecret: "secret",
				CachedAuthToken: AuthToken{
					TokenType:   "bearer",
					AccessToken: "4321",
					ExpiresAt:   time.Now().Add(43200 * time.Second),
				},
			},
			want: want{token: "bearer 4321"},
		},
		{
			name: "Token about to expire",
			fields: fields{
				ClientID:     "myClientID",
				ClientSecret: "secret",
				CachedAuthToken: AuthToken{
					TokenType:   "junk",
					AccessToken: "4321",
					ExpiresAt:   time.Now().Add(100 * time.Second),
				},
			},
			want: want{token: "bearer 1234"},
			response: response{
				bodyText: `{"access_token": "1234", "expires_in": 9876, "token_type": "bearer"}`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start a local HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if tt.response.statusCode != 0 {
					rw.WriteHeader(tt.response.statusCode)
				}
				rw.Write([]byte(tt.response.bodyText))
			}))
			// Close the server when test finishes
			defer server.Close()

			x := &XSUAA{
				OAuthURL:        server.URL,
				ClientID:        tt.fields.ClientID,
				ClientSecret:    tt.fields.ClientSecret,
				CachedAuthToken: tt.fields.CachedAuthToken,
			}
			header := make(http.Header)
			if len(tt.args.authHeader) > 0 {
				header.Add(authHeaderKey, tt.args.authHeader)
			}
			err := x.SetAuthHeaderIfNotPresent(&header)
			if tt.want.errRegex != "" {
				require.Error(t, err, "Error expected")
				assert.Regexp(t, tt.want.errRegex, err.Error(), "")
				return
			}
			require.NoError(t, err, "No error expected")
			assert.Equal(t, tt.want.token, header.Get("Authorization"))
		})
	}
}

func Test_setExpireTime(t *testing.T) {
	t.Run("Straight forward", func(t *testing.T) {
		dummyTime := time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC)
		got := setExpireTime(dummyTime, time.Duration(43200))
		want := time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, got, want, "Time should have increased by 12 hours")
	})
}
