package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunApiKeyValueMapUpload(t *testing.T) {
	t.Parallel()

	t.Run("Successfull Api Key Value Map Create Test", func(t *testing.T) {
		// init
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := apiKeyValueMapUploadOptions{
			APIServiceKey:   apiServiceKey,
			Key:             "demo",
			Value:           "name",
			KeyValueMapName: "demoMap",
		}
		httpClient := httpMockCpis{CPIFunction: "ApiKeyValueMapUpload", ResponseBody: ``, TestType: "PositiveCase"}
		// test
		err := runApiKeyValueMapUpload(&config, nil, &httpClient)
		// assert
		if assert.NoError(t, err) {
			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "https://demo/apiportal/api/1.0/Management.svc/KeyMapEntries", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "POST", httpClient.Method)
			})
		}
	})

	t.Run("Failed case of API Key Value Map Create Test", func(t *testing.T) {
		// init
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := apiKeyValueMapUploadOptions{
			APIServiceKey:   apiServiceKey,
			Key:             "demo",
			Value:           "name",
			KeyValueMapName: "demoMap",
		}

		httpClient := httpMockCpis{CPIFunction: "ApiKeyValueMapUpload", ResponseBody: ``, TestType: "Negative"}

		// test
		err := runApiKeyValueMapUpload(&config, nil, &httpClient)

		// assert
		assert.EqualError(t, err, "HTTP \"POST\" request to \"https://demo/apiportal/api/1.0/Management.svc/KeyMapEntries\" failed with error: 401 Unauthorized")
	})

	t.Run("Test API Key Value Map payload", func(t *testing.T) {
		// init
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := apiKeyValueMapUploadOptions{
			APIServiceKey:   apiServiceKey,
			Key:             "demo",
			Value:           "name",
			KeyValueMapName: "demoMap",
		}

		// test
		payload, err := createJSONPayload(&config)
		// assert
		assert.NoError(t, err)
		assert.NotNil(t, payload)
	})

	t.Run("Http Response not accepted Test case", func(t *testing.T) {
		// init
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := apiKeyValueMapUploadOptions{
			APIServiceKey:   apiServiceKey,
			Key:             "demo",
			Value:           "name",
			KeyValueMapName: "demoMap",
		}

		httpClient := httpMockCpis{CPIFunction: "ApiKeyValueMapUpload", ResponseBody: ``, TestType: "HttpResponseNotAccepted"}

		// test
		err := runApiKeyValueMapUpload(&config, nil, &httpClient)

		// assert
		assert.EqualError(t, err, "Failed to upload API key value map artefact, Response Status code: 202")
	})

	t.Run("Http Response not accepted Test case", func(t *testing.T) {
		// init
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := apiKeyValueMapUploadOptions{
			APIServiceKey:   apiServiceKey,
			Key:             "demo",
			Value:           "name",
			KeyValueMapName: "demoMap",
		}

		httpClient := httpMockCpis{CPIFunction: "ApiKeyValueMapUpload", ResponseBody: ``, TestType: "NilHttpResponse"}

		// test
		err := runApiKeyValueMapUpload(&config, nil, &httpClient)

		// assert
		assert.EqualError(t, err, "HTTP \"POST\" request to \"https://demo/apiportal/api/1.0/Management.svc/KeyMapEntries\" failed with error: invalid payalod")
	})

}
