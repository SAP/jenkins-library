package gcp

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPublish(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// init
		projectNumber := "PROJECT_NUMBER"
		topic := "TOPIC"
		token := "TOKEN"
		data := []byte(mock.Anything)

		apiurl := fmt.Sprintf(api_url, projectNumber, topic)

		mockResponse := map[string]interface{}{
			"messageIds": []string{"10721501285371497"},
		}

		// mock
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodPost, apiurl,
			func(req *http.Request) (*http.Response, error) {
				assert.Contains(t, req.Header, "Authorization")
				assert.Equal(t, req.Header.Get("Authorization"), "Bearer TOKEN")
				assert.Contains(t, req.Header, "Content-Type")
				assert.Equal(t, req.Header.Get("Content-Type"), "application/json")
				return httpmock.NewJsonResponse(http.StatusOK, mockResponse)
			},
		)

		// test
		err := Publish(projectNumber, topic, token, data)
		// asserts
		assert.NoError(t, err)
	})
}
