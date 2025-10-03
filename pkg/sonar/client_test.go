package sonar

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	testURL := "https://example.org/api/"
	t.Run("", func(t *testing.T) {
		// init
		requester := Requester{
			Host:     testURL,
			Username: mock.Anything,
		}
		// test
		request, err := requester.create(http.MethodGet, mock.Anything, &IssuesSearchOption{P: "42"})
		// assert
		assert.NoError(t, err)
		assert.Empty(t, request.URL.Opaque)
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, "https", request.URL.Scheme)
		assert.Equal(t, "example.org", request.URL.Host)
		assert.Equal(t, "/api/"+mock.Anything, request.URL.Path)
		assert.Contains(t, request.Header.Get("Authorization"), "Basic ")
	})
}

func TestNewAPIClient(t *testing.T) {
	tests := []struct {
		name string
		host string
		want string
	}{
		{name: mock.Anything, want: "https://example.org/api/", host: "https://example.org"},
		{name: mock.Anything, want: "https://example.org/api/", host: "https://example.org/"},
		{name: mock.Anything, want: "https://example.org/api/", host: "https://example.org/api"},
		{name: mock.Anything, want: "https://example.org/api/", host: "https://example.org/api/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAPIClient(tt.host, mock.Anything, nil)
			assert.Equal(t, tt.want, got.Host)
		})
	}
}
