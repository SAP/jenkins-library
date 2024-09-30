package gcp

//func TestPublish(t *testing.T) {
//	t.Run("success", func(t *testing.T) {
//		// init
//		projectID := "PROJECT_NUMBER"
//		topic := "TOPIC"
//		token := "TOKEN"
//		data := []byte(mock.Anything)
//
//		apiurl := fmt.Sprintf(api_url, projectID, topic)
//
//		mockResponse := map[string]interface{}{
//			"messageIds": []string{"10721501285371497"},
//		}
//
//		// mock
//		httpmock.Activate()
//		defer httpmock.DeactivateAndReset()
//		httpmock.RegisterResponder(http.MethodPost, apiurl,
//			func(req *http.Request) (*http.Response, error) {
//				assert.Contains(t, req.Header, "Authorization")
//				assert.Equal(t, req.Header.Get("Authorization"), "Bearer TOKEN")
//				assert.Contains(t, req.Header, "Content-Type")
//				assert.Equal(t, req.Header.Get("Content-Type"), "application/json")
//				return httpmock.NewJsonResponse(http.StatusOK, mockResponse)
//			},
//		)
//
//		// test
//		err := Publish(projectID, topic, token, mock.Anything, data)
//		// asserts
//		assert.NoError(t, err)
//	})
//}
