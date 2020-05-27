package whitesource

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
)

type whitesourceMockClient struct {
	httpMethod     string
	httpStatusCode int
	urlsCalled     string
	requestBody    io.Reader
	responseBody   string
}

func (c *whitesourceMockClient) SetOptions(opts piperhttp.ClientOptions) {
	//noop
}

func (c *whitesourceMockClient) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	c.httpMethod = method
	c.urlsCalled = url
	c.requestBody = body
	return &http.Response{StatusCode: c.httpStatusCode, Body: ioutil.NopCloser(bytes.NewReader([]byte(c.responseBody)))}, nil
}

func TestGetProductsMetaInfo(t *testing.T) {
	myTestClient := whitesourceMockClient{
		responseBody: `{
	"productVitals":[
		{
			"name": "Test Product",
			"token": "test_product_token",
			"creationDate": "2020-01-01 00:00:00",
			"lastUpdatedDate": "2020-01-01 01:00:00"
		}
	]
}`,
	}

	expectedRequestBody := `{"requestType":"getOrganizationProductVitals","userKey":"test_user_token","orgToken":"test_org_token"}`

	sys := System{ServerURL: "https://my.test.server", HTTPClient: &myTestClient, OrgToken: "test_org_token", UserToken: "test_user_token"}
	products, err := sys.GetProductsMetaInfo()

	requestBody, err := ioutil.ReadAll(myTestClient.requestBody)
	assert.NoError(t, err)
	assert.Equal(t, expectedRequestBody, string(requestBody))

	assert.NoError(t, err)
	assert.Equal(t, []Product{{Name: "Test Product", Token: "test_product_token", CreationDate: "2020-01-01 00:00:00", LastUpdateDate: "2020-01-01 01:00:00"}}, products)
}

func TestGetMetaInfoForProduct(t *testing.T) {
	myTestClient := whitesourceMockClient{
		responseBody: `{
	"productVitals":[
		{
			"name": "Test Product 1",
			"token": "test_product_token1",
			"creationDate": "2020-01-01 00:00:00",
			"lastUpdatedDate": "2020-01-01 01:00:00"
		},
		{
			"name": "Test Product 2",
			"token": "test_product_token2",
			"creationDate": "2020-02-01 00:00:00",
			"lastUpdatedDate": "2020-02-01 01:00:00"
		}
	]
}`,
	}

	sys := System{ServerURL: "https://my.test.server", HTTPClient: &myTestClient, OrgToken: "test_org_token", UserToken: "test_user_token"}
	product, err := sys.GetMetaInfoForProduct("Test Product 2")

	assert.NoError(t, err)
	assert.Equal(t, product.Name, "Test Product 2")
	assert.Equal(t, product.Token, "test_product_token2")

}

func TestGetProjectsMetaInfo(t *testing.T) {
	myTestClient := whitesourceMockClient{
		responseBody: `{
	"projectVitals":[
		{
			"pluginName":"test-plugin",
			"name": "Test Project",
			"token": "test_project_token",
			"uploadedBy": "test_upload_user",
			"creationDate": "2020-01-01 00:00:00",
			"lastUpdatedDate": "2020-01-01 01:00:00"
		}
	]
}`,
	}

	expectedRequestBody := `{"requestType":"getProductProjectVitals","userKey":"test_user_token","productToken":"test_product_token","orgToken":"test_org_token"}`

	sys := System{ServerURL: "https://my.test.server", HTTPClient: &myTestClient, OrgToken: "test_org_token", UserToken: "test_user_token"}
	projects, err := sys.GetProjectsMetaInfo("test_product_token")

	requestBody, err := ioutil.ReadAll(myTestClient.requestBody)
	assert.NoError(t, err)
	assert.Equal(t, expectedRequestBody, string(requestBody))

	assert.NoError(t, err)
	assert.Equal(t, "Test Project", projects[0].Name)
	assert.Equal(t, "test_project_token", projects[0].Token)
	assert.Equal(t, "test-plugin", projects[0].PluginName)
	assert.Equal(t, "test_upload_user", projects[0].UploadedBy)
	assert.Equal(t, "2020-01-01 00:00:00", projects[0].CreationDate)
	assert.Equal(t, "2020-01-01 01:00:00", projects[0].LastUpdateDate)

}

func TestGetProjectToken(t *testing.T) {
	myTestClient := whitesourceMockClient{
		responseBody: `{
	"projectVitals":[
		{
			"pluginName":"test-plugin",
			"name": "Test Project1",
			"token": "test_project_token1",
			"uploadedBy": "test_upload_user",
			"creationDate": "2020-01-01 00:00:00",
			"lastUpdatedDate": "2020-01-01 01:00:00"
		},
		{
			"pluginName":"test-plugin",
			"name": "Test Project2",
			"token": "test_project_token2",
			"uploadedBy": "test_upload_user",
			"creationDate": "2020-01-01 00:00:00",
			"lastUpdatedDate": "2020-01-01 01:00:00"
		}
	]
}`,
	}

	sys := System{ServerURL: "https://my.test.server", HTTPClient: &myTestClient, OrgToken: "test_org_token", UserToken: "test_user_token"}

	projectToken, err := sys.GetProjectToken("test_product_token", "Test Project1")
	assert.NoError(t, err)
	assert.Equal(t, "test_project_token1", projectToken)

	projectToken, err = sys.GetProjectToken("test_product_token", "Test Project2")
	assert.NoError(t, err)
	assert.Equal(t, "test_project_token2", projectToken)

	projectToken, err = sys.GetProjectToken("test_product_token", "Test Project3")
	assert.NoError(t, err)
	assert.Equal(t, "", projectToken)
}

func TestGetProjectTokens(t *testing.T) {
	myTestClient := whitesourceMockClient{
		responseBody: `{
	"projectVitals":[
		{
			"pluginName":"test-plugin",
			"name": "Test Project1",
			"token": "test_project_token1",
			"uploadedBy": "test_upload_user",
			"creationDate": "2020-01-01 00:00:00",
			"lastUpdatedDate": "2020-01-01 01:00:00"
		},
		{
			"pluginName":"test-plugin",
			"name": "Test Project2",
			"token": "test_project_token2",
			"uploadedBy": "test_upload_user",
			"creationDate": "2020-01-01 00:00:00",
			"lastUpdatedDate": "2020-01-01 01:00:00"
		}
	]
}`,
	}

	sys := System{ServerURL: "https://my.test.server", HTTPClient: &myTestClient, OrgToken: "test_org_token", UserToken: "test_user_token"}

	projectTokens, err := sys.GetProjectTokens("test_product_token", []string{"Test Project1", "Test Project2"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"test_project_token1", "test_project_token2"}, projectTokens)

	projectTokens, err = sys.GetProjectTokens("test_product_token", []string{"Test Project3"})
	assert.NoError(t, err)
	assert.Equal(t, []string{}, projectTokens)
}

func TestGetProductName(t *testing.T) {
	myTestClient := whitesourceMockClient{
		responseBody: `{
	"productTags":[
		{
			"name": "Test Product",
			"token": "test_product_token"
		}
	]
}`,
	}

	sys := System{ServerURL: "https://my.test.server", HTTPClient: &myTestClient, OrgToken: "test_org_token", UserToken: "test_user_token"}

	productName, err := sys.GetProductName("test_product_token")
	assert.NoError(t, err)
	assert.Equal(t, "Test Product", productName)
}
