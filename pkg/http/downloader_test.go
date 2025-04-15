//go:build unit

package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestDownloadRequest(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(
		http.HandlerFunc(
			func(rw http.ResponseWriter, req *http.Request) { rw.Write([]byte("my fancy file content")) }))
	// Close the server when test finishes
	defer server.Close()

	client := Client{
		logger: log.Entry().WithField("package", "SAP/jenkins-library/pkg/http"),
	}

	workingDir := t.TempDir()
	targetFile := filepath.Join(workingDir, "abc/123/abc.xml")

	// function under test
	err := client.DownloadFile(server.URL, targetFile, nil, nil)
	// asserts
	assert.NoError(t, err, "Error occurred but none expected")
	assert.FileExists(t, targetFile, "File not found")
	bytes, err := os.ReadFile(targetFile)
	assert.NoError(t, err, "Error occurred but none expected")
	assert.Equal(t, "my fancy file content", string(bytes))
}
