package fortify

import (
	"strings"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	ff "github.com/piper-validation/fortify-client-go/fortify"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
)

func TestGetProjectByName(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/projects" && req.URL.RawQuery == "fulltextsearch=true&q=name%3Dpython-test-sven" {
			rw.Write([]byte(
				`{"count": 0, "data": [{"createdBy": "string", "creationDate": "2020-02-10T21:22:11.506+0000", "description": "string",
"id": 4711, "issueTemplateId": "string", "name": "python-test-sven"}, "errorCode": 0, "links": {}, "message": "string",
"responseCode": 200, "stackTrace": "string", "successCount": 0}`))
			return
		}
		if req.URL.Path == "/projectVersions" && req.URL.RawQuery == "fulltextsearch=true&q=name%3D0" {
			rw.Write([]byte(`{
				"data": [
				  {
					"id": 666,
					"name": "0"
				  }
				],
				"errorCode": 0,
				"responseCode": 200
			  }`))
			return
		}
	}))
	// Close the server when test finishes
	defer server.Close()

	parts := strings.Split(server.URL, "://")

	t.Run("test success", func(t *testing.T) {
		client := ff.NewHTTPClientWithConfig(strfmt.Default, &ff.TransportConfig{
			Host:     parts[1],
			Schemes:  []string{parts[0]},
			BasePath: ""},
		)
		sys := NewSystemInstanceForClient(client, "test2456", 60*time.Second)

		result, err := sys.GetProjectByName("python-test-sven")
		assert.NoError(t, err, "GetProjectByName call not successful")
		assert.Equal(t, "python-test-sven", strings.ToLower(*result.Name), "Expected to receive python-test-sven")

		result2, err := sys.GetProjectVersionDetailsByNameAndProjectID(result.ID, "0")
		assert.NoError(t, err, "GetProjectVersionDetailsByNameAndProjectID call not successful")
		assert.Equal(t, "0", *result2.Name, "Expected project version with different name")
	})
}
