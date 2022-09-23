package pact

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnsureDir(t *testing.T) {
	t.Parallel()

	t.Run("success - not existing", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		
		err := EnsureDir("test/path", mockUtils)
		assert.NoError(t, err)
		exists, _ := mockUtils.DirExists("test/path")
		assert.True(t, exists)
	})

	t.Run("success - existing", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		mockUtils.AddDir("test/path")

		err := EnsureDir("test/path", mockUtils)
		assert.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		mockUtils.DirCreateErrors = map[string]error{"test/path": fmt.Errorf("create error")}
		
		err := EnsureDir("test/path", mockUtils)
		assert.EqualError(t, err, "create error")
	})
}

func TestEnsureValidDir(t *testing.T) {
	t.Parallel()

	tt := []struct{
		desc string
		path string
		expected string

	}{
		{desc: "success - default", path: "test/path/", expected: "test/path/"},
		{desc: "success - json", path: "test/path/test.json", expected: "test/path/"},
		{desc: "success - no trailing /", path: "test/path", expected: "test/path/"},
	}

	for _, test := range tt {
		t.Run(test.desc, func(t *testing.T){
			assert.Equal(t, test.expected, EnsureValidDir(test.path))
		})
	}
}

func TestReadAndUnmarshalFile(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		testFile := "path/to/test.json"
		mockUtils.AddFile(testFile, []byte(`{"test":"value"}`))

		res := map[string]string{}
		err := ReadAndUnmarshalFile(testFile, &res, mockUtils)
		assert.NoError(t, err)
		assert.Equal(t, "value", res["test"])
	})

	t.Run("failure - read file", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		testFile := "path/to/test.json"
		mockUtils.FileReadErrors = map[string]error{testFile: fmt.Errorf("read error")}

		res := map[string]string{}
		err := ReadAndUnmarshalFile(testFile, &res, mockUtils)
		assert.EqualError(t, err, "read error")
	})


	t.Run("failure - unmarshal content", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		testFile := "path/to/test.json"
		mockUtils.AddFile(testFile, []byte(`{`))

		res := map[string]string{}
		err := ReadAndUnmarshalFile(testFile, &res, mockUtils)
		assert.Contains(t, fmt.Sprint(err), "failed to unmarshal path/to/test.json")
	})
}

func TestSendRequest(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		mockUtils.httpResponseContent = "the response content"

		b := bytes.NewBuffer([]byte("the body"))
		res, err := sendRequest(http.MethodGet, "http://the.url", "testUser", "testPassword", b, mockUtils)
		assert.NoError(t, err)
		assert.Equal(t, mockUtils.httpResponseContent, string(res))
		assert.Equal(t, b, mockUtils.body)
	})

	t.Run("failure - send", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		mockUtils.httpSendErrors = map[string]error{"http://the.url": fmt.Errorf("send error")}

		b := bytes.NewBuffer([]byte{})
		_, err := sendRequest(http.MethodGet, "http://the.url", "testUser", "testPassword", b, mockUtils)
		assert.EqualError(t, err, "send error")
	})

	t.Run("failure - not found", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		mockUtils.httpResponseStatusCode = http.StatusNotFound

		b := bytes.NewBuffer([]byte{})
		_, err := sendRequest(http.MethodGet, "http://the.url", "testUser", "testPassword", b, mockUtils)
		assert.EqualError(t, err, fmt.Sprint(ErrNotFound))
	})
}



