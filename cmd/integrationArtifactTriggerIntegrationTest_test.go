package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type integrationArtifactTriggerIntegrationTestMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newIntegrationArtifactTriggerIntegrationTestTestsUtils() integrationArtifactTriggerIntegrationTestMockUtils {
	utils := integrationArtifactTriggerIntegrationTestMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunIntegrationArtifactTriggerIntegrationTest(t *testing.T) {
	t.Run("MessageBodyPath good but no ContentType (ERROR) callIFlowURL", func(t *testing.T) {
		//init
		iFlowServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := integrationArtifactTriggerIntegrationTestOptions{
			IFlowServiceKey:   iFlowServiceKey,
			IntegrationFlowID: "CPI_IFlow_Call_using_Cert",
			Platform:          "cf",
			MessageBodyPath:   "/file.txt",
			ContentType:       "",
		}

		utils := newIntegrationArtifactTriggerIntegrationTestTestsUtils()
		utils.AddFile("file.txt", []byte("dummycontent"))
		httpClient := httpMockCpis{CPIFunction: "TriggerIntegrationTest", ResponseBody: ``, TestType: "Positive"}

		//test
		err := callIFlowURL(&config, nil, utils, &httpClient, "")

		//assert
		assert.Error(t, err)
	})

	t.Run("MessageBodyPath and ContentType good but file missing (ERROR) callIFlowURL", func(t *testing.T) {
		//init
		iFlowServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := integrationArtifactTriggerIntegrationTestOptions{
			IFlowServiceKey:   iFlowServiceKey,
			IntegrationFlowID: "CPI_IFlow_Call_using_Cert",
			Platform:          "cf",
			MessageBodyPath:   "test.txt",
			ContentType:       "txt",
		}

		utils := newIntegrationArtifactTriggerIntegrationTestTestsUtils()
		//no file created here. error expected
		httpClient := httpMockCpis{CPIFunction: "TriggerIntegrationTest", ResponseBody: ``, TestType: "Positive"}

		//test
		err := callIFlowURL(&config, nil, utils, &httpClient, "")

		//assert
		assert.Error(t, err)
	})

	t.Run("MessageBodyPath, ContentType, and file good (SUCCESS) callIFlowURL", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "")
		defer os.RemoveAll(dir) // clean up
		assert.NoError(t, err, "Error when creating temp dir")
		//init
		iFlowServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := integrationArtifactTriggerIntegrationTestOptions{
			IFlowServiceKey:   iFlowServiceKey,
			IntegrationFlowID: "CPI_IFlow_Call_using_Cert",
			Platform:          "cf",
			MessageBodyPath:   filepath.Join(dir, "test.txt"),
			ContentType:       "txt",
		}

		utils := newIntegrationArtifactTriggerIntegrationTestTestsUtils()
		utils.AddFile(config.MessageBodyPath, []byte("dummycontent1")) //have to add a file here to see in utils
		ioutil.WriteFile(config.MessageBodyPath, []byte("dummycontent2"), 0755)
		httpClient := httpMockCpis{CPIFunction: "TriggerIntegrationTest", ResponseBody: ``, TestType: "Positive"}

		//test
		err = callIFlowURL(&config, nil, utils, &httpClient, "https://my-service.com/endpoint")

		//assert
		assert.NoError(t, err)
		assert.Equal(t,"POST", httpClient.Method)
		assert.Equal(t,"https://my-service.com/endpoint", httpClient.URL)
	})

	t.Run("No MessageBodyPath still works (SUCCESS) callIFlowURL", func(t *testing.T) {
		//init
		iFlowServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := integrationArtifactTriggerIntegrationTestOptions{
			IFlowServiceKey:   iFlowServiceKey,
			IntegrationFlowID: "CPI_IFlow_Call_using_Cert",
			Platform:          "cf",
			MessageBodyPath:   "",
			ContentType:       "txt",
		}

		utils := newIntegrationArtifactTriggerIntegrationTestTestsUtils()
		//utils.AddFile(config.MessageBodyPath, []byte("dummycontent1")) //have to add a file here to see in utils
		//ioutil.WriteFile(config.MessageBodyPath, []byte("dummycontent2"), 0755)
		httpClient := httpMockCpis{CPIFunction: "TriggerIntegrationTest", ResponseBody: ``, TestType: "Positive"}

		//test
		err := callIFlowURL(&config, nil, utils, &httpClient, "https://my-service.com/endpoint")

		//assert
		assert.NoError(t, err)
		assert.Equal(t,"GET", httpClient.Method)
		assert.Equal(t,"https://my-service.com/endpoint", httpClient.URL)

	})

	t.Run("nil fileBody (SUCCESS) callIFlowURL", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "")
		defer os.RemoveAll(dir) // clean up
		assert.NoError(t, err, "Error when creating temp dir")
		//init
		iFlowServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := integrationArtifactTriggerIntegrationTestOptions{
			IFlowServiceKey:   iFlowServiceKey,
			IntegrationFlowID: "CPI_IFlow_Call_using_Cert",
			Platform:          "cf",
			MessageBodyPath:   filepath.Join(dir, "test.txt"),
			ContentType:       "txt",
		}

		utils := newIntegrationArtifactTriggerIntegrationTestTestsUtils()
		utils.AddFile(config.MessageBodyPath, []byte(nil)) //have to add a file here to see in utils
		ioutil.WriteFile(config.MessageBodyPath, []byte(nil), 0755)
		httpClient := httpMockCpis{CPIFunction: "TriggerIntegrationTest", ResponseBody: ``, TestType: "Positive"}

		//test
		err = callIFlowURL(&config, nil, utils, &httpClient, "")

		//assert
		assert.NoError(t, err)
	})
}
