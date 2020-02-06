package http

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
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
		{client: Client{logger: log.Entry().WithField("package", "SAP/jenkins-library/pkg/http")}, method: "GET", expected: "OK"},
		{client: Client{logger: log.Entry().WithField("package", "SAP/jenkins-library/pkg/http")}, method: "GET", header: map[string][]string{"Testheader": []string{"Test1", "Test2"}}, expected: "OK"},
		{client: Client{logger: log.Entry().WithField("package", "SAP/jenkins-library/pkg/http")}, cookies: []*http.Cookie{{Name: "TestCookie1", Value: "TestValue1"}, {Name: "TestCookie2", Value: "TestValue2"}}, method: "GET", expected: "OK"},
		{client: Client{logger: log.Entry().WithField("package", "SAP/jenkins-library/pkg/http"), username: "TestUser", password: "TestPwd"}, method: "GET", expected: "OK"},
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
	opts := ClientOptions{Timeout: 10, Username: "TestUser", Password: "TestPassword", Token: "TestToken", Logger: log.Entry().WithField("package", "github.com/SAP/jenkins-library/pkg/http")}
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
		{client: Client{}, expected: Client{timeout: time.Second * 10, logger: log.Entry().WithField("package", "SAP/jenkins-library/pkg/http")}},
		{client: Client{timeout: 10}, expected: Client{timeout: 10, logger: log.Entry().WithField("package", "SAP/jenkins-library/pkg/http")}},
	}

	for k, v := range tt {
		v.client.applyDefaults()
		assert.Equal(t, v.expected, v.client, fmt.Sprintf("Run %v failed", k))
	}
}

func TestUploadRequest(t *testing.T) {
	var passedHeaders = map[string][]string{}
	passedCookies := []*http.Cookie{}
	var passedUsername string
	var passedPassword string
	var multipartFile multipart.File
	var multipartHeader *multipart.FileHeader
	var passedFileContents []byte
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
		err := req.ParseMultipartForm(4096)
		if err != nil {
			t.FailNow()
		}
		multipartFile, multipartHeader, err = req.FormFile("Field1")
		if err != nil {
			t.FailNow()
		}
		defer req.Body.Close()
		passedFileContents, err = ioutil.ReadAll(multipartFile)
		if err != nil {
			t.FailNow()
		}

		rw.Write([]byte("OK"))
	}))
	// Close the server when test finishes
	defer server.Close()

	testFile, err := ioutil.TempFile("", "testFileUpload")
	if err != nil {
		t.FailNow()
	}
	defer os.RemoveAll(testFile.Name()) // clean up

	fileContents, err := ioutil.ReadFile(testFile.Name())
	if err != nil {
		t.FailNow()
	}

	tt := []struct {
		clientOptions ClientOptions
		method        string
		body          io.Reader
		header        http.Header
		cookies       []*http.Cookie
		expected      string
	}{
		{clientOptions: ClientOptions{}, method: "PUT", expected: "OK"},
		{clientOptions: ClientOptions{}, method: "POST", expected: "OK"},
		{clientOptions: ClientOptions{}, method: "POST", header: map[string][]string{"Testheader": []string{"Test1", "Test2"}}, expected: "OK"},
		{clientOptions: ClientOptions{}, cookies: []*http.Cookie{{Name: "TestCookie1", Value: "TestValue1"}, {Name: "TestCookie2", Value: "TestValue2"}}, method: "POST", expected: "OK"},
		{clientOptions: ClientOptions{Username: "TestUser", Password: "TestPwd"}, method: "POST", expected: "OK"},
	}

	client := Client{logger: log.Entry().WithField("package", "SAP/jenkins-library/pkg/http")}
	for key, test := range tt {
		t.Run(fmt.Sprintf("UploadFile Row %v", key+1), func(t *testing.T) {
			client.SetOptions(test.clientOptions)
			response, err := client.UploadFile(server.URL, testFile.Name(), "Field1", test.header, test.cookies)
			assert.NoError(t, err, "Error occurred but none expected")
			content, err := ioutil.ReadAll(response.Body)
			assert.NoError(t, err, "Error occurred but none expected")
			assert.Equal(t, test.expected, string(content), "Returned content incorrect")
			response.Body.Close()

			assert.Equal(t, testFile.Name(), multipartHeader.Filename, "Uploaded file incorrect")
			assert.Equal(t, fileContents, passedFileContents, "Uploaded file incorrect")

			for k, h := range test.header {
				assert.Containsf(t, passedHeaders, k, "Header %v not contained", k)
				assert.Equalf(t, h, passedHeaders[k], "Header %v contains different value")
			}

			if len(test.cookies) > 0 {
				assert.Equal(t, test.cookies, passedCookies, "Passed cookies not correct")
			}

			if len(client.username) > 0 {
				assert.Equal(t, client.username, passedUsername)
			}

			if len(client.password) > 0 {
				assert.Equal(t, client.password, passedPassword)
			}
		})
		t.Run(fmt.Sprintf("UploadRequest Row %v", key+1), func(t *testing.T) {
			client.SetOptions(test.clientOptions)
			response, err := client.UploadRequest(test.method, server.URL, testFile.Name(), "Field1", test.header, test.cookies)
			assert.NoError(t, err, "Error occurred but none expected")
			content, err := ioutil.ReadAll(response.Body)
			assert.NoError(t, err, "Error occurred but none expected")
			assert.Equal(t, test.expected, string(content), "Returned content incorrect")
			response.Body.Close()

			assert.Equal(t, testFile.Name(), multipartHeader.Filename, "Uploaded file incorrect")
			assert.Equal(t, fileContents, passedFileContents, "Uploaded file incorrect")

			for k, h := range test.header {
				assert.Containsf(t, passedHeaders, k, "Header %v not contained", k)
				assert.Equalf(t, h, passedHeaders[k], "Header %v contains different value")
			}

			if len(test.cookies) > 0 {
				assert.Equal(t, test.cookies, passedCookies, "Passed cookies not correct")
			}

			if len(client.username) > 0 {
				assert.Equal(t, client.username, passedUsername)
			}

			if len(client.password) > 0 {
				assert.Equal(t, client.password, passedPassword)
			}
		})
	}
}
func TestUploadRequestWrongMethod(t *testing.T) {

	client := Client{logger: log.Entry().WithField("package", "SAP/jenkins-library/pkg/http")}
	_, err := client.UploadRequest("GET", "dummy", "testFile", "Field1", nil, nil)
	assert.Error(t, err, "No error occured but was expected")
}
