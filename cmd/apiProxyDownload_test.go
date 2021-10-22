package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type apiProxyDownloadMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func TestRunApiProxyDownload(t *testing.T) {
	t.Parallel()

	t.Run("Successfull Download of API Proxy", func(t *testing.T) {
		tempDir, tmpErr := ioutil.TempDir("", "")
		defer os.RemoveAll(tempDir) // clean up
		assert.NoError(t, tmpErr, "Error when creating temp dir")
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := apiProxyDownloadOptions{
			APIServiceKey: apiServiceKey,
			APIProxyName:  "flow1",
			DownloadPath:  tempDir,
		}
		httpClient := httpMockCpis{CPIFunction: "APIProxyDownload", ResponseBody: ``, TestType: "PositiveAndGetetIntegrationArtifactDownloadResBody"}
		err := runApiProxyDownload(&config, nil, &httpClient)
		absolutePath := filepath.Join(tempDir, "flow1.zip")
		assert.DirExists(t, tempDir)
		if assert.NoError(t, err) {
			assert.Equal(t, fileExists(absolutePath), true)
			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "https://demo/apiportal/api/1.0/Transport.svc/APIProxies?name=flow1", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})
		}
	})

	t.Run("Failed case of api proxy artifact Download", func(t *testing.T) {
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := apiProxyDownloadOptions{
			APIServiceKey: apiServiceKey,
			APIProxyName:  "proxy1",
			DownloadPath:  "tmp",
		}
		httpClient := httpMockCpis{CPIFunction: "APIProxyDownloadFailure", ResponseBody: ``, TestType: "Negative"}
		err := runApiProxyDownload(&config, nil, &httpClient)
		assert.EqualError(t, err, "HTTP GET request to https://demo/apiportal/api/1.0/Transport.svc/APIProxies?name=proxy1 failed with error: Service not Found")
	})
}
