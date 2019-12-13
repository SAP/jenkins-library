package http

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSendRequest(t *testing.T) {
	var passedHeaders = map[string][]string{}
	passedCookies := []*http.Cookie{}
	var passedUsername string
	var passedPassword string
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		passedHeaders = map[string][]string{}
		if req.Header != nil {
			for name, headers := range req.Header {
				passedHeaders[name] = headers
			}
		}
		passedCookies = req.Cookies()
		passedUsername, passedPassword, _ = req.BasicAuth()

		rw.Write([]byte("OK"))
	}))
	// Close the server when test finishes
	defer server.Close()

	tt := []struct {
		client   Client
		method   string
		body     io.Reader
		header   http.Header
		cookies  []*http.Cookie
		expected string
	}{
		{client: Client{}, method: "GET", expected: "OK"},
		{client: Client{}, method: "GET", header: map[string][]string{"Testheader": []string{"Test1", "Test2"}}, expected: "OK"},
		{client: Client{}, cookies: []*http.Cookie{{Name: "TestCookie1", Value: "TestValue1"}, {Name: "TestCookie2", Value: "TestValue2"}}, method: "GET", expected: "OK"},
		{client: Client{username: "TestUser", password: "TestPwd"}, method: "GET", expected: "OK"},
	}

	for key, test := range tt {
		t.Run(fmt.Sprintf("Row %v", key+1), func(t *testing.T) {
			response, err := test.client.SendRequest("GET", server.URL, test.body, test.header, test.cookies)
			assert.NoError(t, err, "Error occured but none expected")
			content, err := ioutil.ReadAll(response.Body)
			assert.Equal(t, test.expected, string(content), "Returned content incorrect")
			response.Body.Close()

			for k, h := range test.header {
				assert.Containsf(t, passedHeaders, k, "Header %v not contained", k)
				assert.Equalf(t, h, passedHeaders[k], "Header %v contains different value")
			}

			if len(test.cookies) > 0 {
				assert.Equal(t, test.cookies, passedCookies, "Passed cookies not correct")
			}

			if len(test.client.username) > 0 {
				assert.Equal(t, test.client.username, passedUsername)
			}

			if len(test.client.password) > 0 {
				assert.Equal(t, test.client.password, passedPassword)
			}
		})
	}
}

func TestSetOptions(t *testing.T) {
	c := Client{}
	opts := ClientOptions{Timeout: 10, Username: "TestUser", Password: "TestPassword", Token: "TestToken"}
	c.SetOptions(opts)

	assert.Equal(t, opts.Timeout, c.timeout)
	assert.Equal(t, opts.Username, c.username)
	assert.Equal(t, opts.Password, c.password)
	assert.Equal(t, opts.Token, c.token)
}

func TestApplyDefaults(t *testing.T) {
	tt := []struct {
		client   Client
		expected Client
	}{
		{client: Client{}, expected: Client{timeout: time.Second * 10}},
		{client: Client{timeout: 10}, expected: Client{timeout: 10}},
	}

	for k, v := range tt {
		v.client.applyDefaults()
		assert.Equal(t, v.expected, v.client, fmt.Sprintf("Run %v failed", k))
	}
}
