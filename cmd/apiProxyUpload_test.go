package cmd

import (
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestRunApiProxyUpload(t *testing.T) {
	t.Parallel()

	t.Run("Successfull Api Proxy Create Test", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		path := filepath.Join("tempDir", "apiproxy.zip")
		filesMock.AddFile(path, []byte("dummy content"))
		exists, err := filesMock.FileExists(path)
		if err != nil {
			t.Fatal("Failed to create temporary file")
		}
		assert.True(t, exists)
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`

		config := apiProxyUploadOptions{
			APIServiceKey: apiServiceKey,
			FilePath:      path,
		}

		httpClient := httpMockCpis{CPIFunction: "ApiProxyUpload", ResponseBody: ``, TestType: "ApiProxyUploadPositiveCase"}

		err = runApiProxyUpload(&config, nil, &filesMock, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "https://demo/apiportal/api/1.0/Transport.svc/APIProxies", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "POST", httpClient.Method)
			})
		}
	})

	t.Run("Failed case of API Proxy Create Test", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		path := filepath.Join("tempDir", "apiproxy.zip")
		filesMock.AddFile(path, []byte("dummy content"))
		exists, err := filesMock.FileExists(path)
		assert.NoError(t, err)
		assert.True(t, exists)
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`

		config := apiProxyUploadOptions{
			APIServiceKey: apiServiceKey,
			FilePath:      path,
		}

		httpClient := httpMockCpis{CPIFunction: "ApiProxyArtifactFail", ResponseBody: ``, TestType: "NegativeApiProxyArtifactUploadResBody"}

		uploadErr := runApiProxyUpload(&config, nil, &filesMock, &httpClient)
		assert.EqualError(t, uploadErr, "HTTP POST request to https://demo/apiportal/api/1.0/Transport.svc/APIProxies failed with error: : Service not Found")
	})

	t.Run("file not exist", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`

		config := apiProxyUploadOptions{
			APIServiceKey: apiServiceKey,
			FilePath:      "",
		}

		httpClient := httpMockCpis{CPIFunction: "ApiProxyArtifactFail", ResponseBody: ``, TestType: "NegativeApiProxyArtifactUploadResBody"}

		uploadErr := runApiProxyUpload(&config, nil, &filesMock, &httpClient)
		assert.EqualError(t, uploadErr, "Error reading file: could not read ''")
	})

	t.Run("file not zip", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		path := filepath.Join("tempDir", "apiproxy.pptx")
		filesMock.AddFile(path, []byte("dummy content"))
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`

		config := apiProxyUploadOptions{
			APIServiceKey: apiServiceKey,
			FilePath:      path,
		}

		httpClient := httpMockCpis{CPIFunction: "ApiProxyArtifactFail", ResponseBody: ``, TestType: "NegativeApiProxyArtifactUploadResBody"}

		uploadErr := runApiProxyUpload(&config, nil, &filesMock, &httpClient)
		assert.EqualError(t, uploadErr, "not valid zip archive")
	})
}
