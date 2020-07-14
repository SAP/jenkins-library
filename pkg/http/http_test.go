package http

import (
	"bytes"
	"encoding/xml"
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
	"github.com/stretchr/testify/mock"
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
	opts := ClientOptions{TransportTimeout: 10, MaxRequestDuration: 5, Username: "TestUser", Password: "TestPassword", Token: "TestToken", Logger: log.Entry().WithField("package", "github.com/SAP/jenkins-library/pkg/http")}
	c.SetOptions(opts)

	assert.Equal(t, opts.TransportTimeout, c.transportTimeout)
	assert.Equal(t, opts.MaxRequestDuration, c.maxRequestDuration)
	assert.Equal(t, opts.Username, c.username)
	assert.Equal(t, opts.Password, c.password)
	assert.Equal(t, opts.Token, c.token)
}

func TestApplyDefaults(t *testing.T) {
	tt := []struct {
		client   Client
		expected Client
	}{
		{client: Client{}, expected: Client{transportTimeout: 3 * time.Minute, maxRequestDuration: 0, logger: log.Entry().WithField("package", "SAP/jenkins-library/pkg/http")}},
		{client: Client{transportTimeout: 10, maxRequestDuration: 5}, expected: Client{transportTimeout: 10, maxRequestDuration: 5, logger: log.Entry().WithField("package", "SAP/jenkins-library/pkg/http")}},
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

func TestTransportTimout(t *testing.T) {
	t.Run("timeout works on transport level", func(t *testing.T) {
		// init
		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Sleep for longer than the configured timeout
			time.Sleep(2 * time.Second)
		}))
		defer svr.Close()

		client := Client{transportTimeout: 1 * time.Second}
		buffer := bytes.Buffer{}

		// test
		_, err := client.SendRequest(http.MethodGet, svr.URL, &buffer, nil, nil)
		// assert
		if assert.Error(t, err, "expected request to fail") {
			assert.Contains(t, err.Error(), "timeout awaiting response headers")
		}
	})
	t.Run("timeout is not hit on transport level", func(t *testing.T) {
		// init
		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Sleep for less than the configured timeout
			time.Sleep(1 * time.Second)
		}))
		defer svr.Close()

		client := Client{transportTimeout: 2 * time.Second}
		buffer := bytes.Buffer{}
		// test
		_, err := client.SendRequest(http.MethodGet, svr.URL, &buffer, nil, nil)
		// assert
		assert.NoError(t, err)
	})
}

func TestParseHTTPResponseBodyJSON(t *testing.T) {

	type myJSONStruct struct {
		FullName string `json:"full_name"`
		Name     string `json:"name"`
		Owner    struct {
			Login string `json:"login"`
		} `json:"owner"`
	}

	t.Run("parse JSON successfull", func(t *testing.T) {

		json := `{"name":"Test Name","full_name":"test full name","owner":{"login": "octocat"}}`
		// create a new reader with that JSON
		r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
		httpResponse := http.Response{
			StatusCode: 200,
			Body:       r,
		}

		var response myJSONStruct
		err := ParseHTTPResponseBodyJSON(&httpResponse, &response)

		if assert.NoError(t, err) {

			t.Run("check correct parsing", func(t *testing.T) {
				assert.Equal(t, myJSONStruct(myJSONStruct{FullName: "test full name", Name: "Test Name", Owner: struct {
					Login string "json:\"login\""
				}{Login: "octocat"}}), response)
			})

		}
	})

	t.Run("http response is nil", func(t *testing.T) {

		var response myJSONStruct
		err := ParseHTTPResponseBodyJSON(nil, &response)

		t.Run("check error", func(t *testing.T) {
			assert.EqualError(t, err, "cannot parse HTTP response with value <nil>")
		})

	})

	t.Run("wrong JSON formatting", func(t *testing.T) {

		json := `{"name":"Test Name","full_name":"test full name";"owner":{"login": "octocat"}}`
		r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
		httpResponse := http.Response{
			StatusCode: 200,
			Body:       r,
		}

		var response myJSONStruct
		err := ParseHTTPResponseBodyJSON(&httpResponse, &response)
		println(response.FullName)

		t.Run("check error", func(t *testing.T) {
			assert.EqualError(t, err, "HTTP response body could not be parsed as JSON: {\"name\":\"Test Name\",\"full_name\":\"test full name\";\"owner\":{\"login\": \"octocat\"}}: invalid character ';' after object key:value pair")
		})

	})

	t.Run("IO read error", func(t *testing.T) {

		mockReadCloser := mockReadCloser{}
		// if Read is called, it will return error
		mockReadCloser.On("Read", mock.AnythingOfType("[]uint8")).Return(0, fmt.Errorf("error reading"))
		// if Close is called, it will return error
		mockReadCloser.On("Close").Return(fmt.Errorf("error closing"))

		httpResponse := http.Response{
			StatusCode: 200,
			Body:       &mockReadCloser,
		}

		var response myJSONStruct
		err := ParseHTTPResponseBodyJSON(&httpResponse, &response)

		t.Run("check error", func(t *testing.T) {
			assert.EqualError(t, err, "HTTP response body could not be read: error reading")
		})

	})

}

func TestParseHTTPResponseBodyXML(t *testing.T) {

	type myXMLStruct struct {
		XMLName xml.Name `xml:"service"`
		Text    string   `xml:",chardata"`
		App     string   `xml:"app,attr"`
		Atom    string   `xml:"atom,attr"`
	}

	t.Run("parse XML successfull", func(t *testing.T) {

		myXML := `
		<?xml version="1.0" encoding="utf-8"?>
		<app:service xmlns:app="http://www.w3.org/2007/app" xmlns:atom="http://www.w3.org/2005/Atom"/>
		`
		// create a new reader with that xml
		r := ioutil.NopCloser(bytes.NewReader([]byte(myXML)))
		httpResponse := http.Response{
			StatusCode: 200,
			Body:       r,
		}

		var response myXMLStruct
		err := ParseHTTPResponseBodyXML(&httpResponse, &response)

		if assert.NoError(t, err) {

			t.Run("check correct parsing", func(t *testing.T) {
				// assert.Equal(t, "<?xml version=\"1.0\" encoding=\"utf-8\"?><app:service xmlns:app=\"http://www.w3.org/2007/app\" xmlns:atom=\"http://www.w3.org/2005/Atom\"/>", response)
				assert.Equal(t, myXMLStruct(myXMLStruct{XMLName: xml.Name{Space: "http://www.w3.org/2007/app", Local: "service"}, Text: "", App: "http://www.w3.org/2007/app", Atom: "http://www.w3.org/2005/Atom"}), response)
			})

		}
	})

	t.Run("http response is nil", func(t *testing.T) {

		var response myXMLStruct
		err := ParseHTTPResponseBodyXML(nil, &response)

		t.Run("check error", func(t *testing.T) {
			assert.EqualError(t, err, "cannot parse HTTP response with value <nil>")
		})

	})

	t.Run("wrong XML formatting", func(t *testing.T) {

		myXML := `
		<?xml version="1.0" encoding="utf-8"?>
		<app:service xmlns:app=http://www.w3.org/2007/app" xmlns:atom="http://www.w3.org/2005/Atom"/>
		`
		r := ioutil.NopCloser(bytes.NewReader([]byte(myXML)))
		httpResponse := http.Response{
			StatusCode: 200,
			Body:       r,
		}

		var response myXMLStruct
		err := ParseHTTPResponseBodyXML(&httpResponse, &response)

		t.Run("check error", func(t *testing.T) {
			assert.EqualError(t, err, "HTTP response body could not be parsed as XML: \n\t\t<?xml version=\"1.0\" encoding=\"utf-8\"?>\n\t\t<app:service xmlns:app=http://www.w3.org/2007/app\" xmlns:atom=\"http://www.w3.org/2005/Atom\"/>\n\t\t: XML syntax error on line 3: unquoted or missing attribute value in element")
		})

	})

	t.Run("IO read error", func(t *testing.T) {

		mockReadCloser := mockReadCloser{}
		// if Read is called, it will return error
		mockReadCloser.On("Read", mock.AnythingOfType("[]uint8")).Return(0, fmt.Errorf("error reading"))
		// if Close is called, it will return error
		mockReadCloser.On("Close").Return(fmt.Errorf("error closing"))

		httpResponse := http.Response{
			StatusCode: 200,
			Body:       &mockReadCloser,
		}

		var response myXMLStruct
		err := ParseHTTPResponseBodyXML(&httpResponse, &response)

		t.Run("check error", func(t *testing.T) {
			assert.EqualError(t, err, "HTTP response body could not be read: error reading")
		})

	})

}

type mockReadCloser struct {
	mock.Mock
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *mockReadCloser) Close() error {
	args := m.Called()
	return args.Error(0)
}
