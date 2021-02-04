package sonar

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	t.Run("", func(t *testing.T) {
		// init
		requester := Requester{
			Host:     testURL,
			Username: mock.Anything,
		}
		// test
		request, err := requester.create(http.MethodGet, endpointIssuesSearch, &IssuesSearchOption{P: "42"})
		// assert
		assert.NoError(t, err)
		assert.Empty(t, request.URL.Opaque)
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, "https", request.URL.Scheme)
		assert.Equal(t, "example.org", request.URL.Host)
		assert.Equal(t, "/api/"+endpointIssuesSearch, request.URL.Path)
		assert.Contains(t, request.Header.Get("Authorization"), "Basic ")
	})
}
