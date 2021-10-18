package whitesource

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type whitesourceMockClient struct {
	httpMethod     string
	httpStatusCode int
	urlsCalled     string
	requestBody    io.Reader
	responseBody   string
	requestError   error
}

func (c *whitesourceMockClient) SetOptions(opts piperhttp.ClientOptions) {
	//noop
}

func (c *whitesourceMockClient) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	c.httpMethod = method
	c.urlsCalled = url
	c.requestBody = body
	if c.requestError != nil {
		return &http.Response{}, c.requestError
	}
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

	sys := System{serverURL: "https://my.test.server", httpClient: &myTestClient, orgToken: "test_org_token", userToken: "test_user_token"}
	products, err := sys.GetProductsMetaInfo()

	requestBody, err := ioutil.ReadAll(myTestClient.requestBody)
	assert.NoError(t, err)
	assert.Equal(t, expectedRequestBody, string(requestBody))

	assert.NoError(t, err)
	assert.Equal(t, []Product{{Name: "Test Product", Token: "test_product_token", CreationDate: "2020-01-01 00:00:00", LastUpdateDate: "2020-01-01 01:00:00"}}, products)
}

func TestCreateProduct(t *testing.T) {
	t.Parallel()
	t.Run("retryable error", func(t *testing.T) {
		// init
		myTestClient := whitesourceMockClient{
			responseBody: `{"errorCode":3000,"errorMessage":"WhiteSource backend has a hickup"}`,
		}
		expectedRequestBody := `{"requestType":"createProduct","userKey":"test_user_token","productName":"test_product_name","orgToken":"test_org_token"}`
		sys := System{serverURL: "https://my.test.server", httpClient: &myTestClient, orgToken: "test_org_token", userToken: "test_user_token"}
		sys.maxRetries = 3
		sys.retryInterval = 1 * time.Microsecond
		// test
		productToken, err := sys.CreateProduct("test_product_name")
		// assert
		assert.EqualError(t, err, "WhiteSource request failed: 3 retries failed: invalid request, error code 3000, message 'WhiteSource backend has a hickup'")
		requestBody, err := ioutil.ReadAll(myTestClient.requestBody)
		require.NoError(t, err)
		assert.Equal(t, "", productToken)
		assert.Equal(t, expectedRequestBody, string(requestBody))
	})
	t.Run("not allowed error", func(t *testing.T) {
		// init
		myTestClient := whitesourceMockClient{
			responseBody: `{"errorCode":5001,"errorMessage":"User is not allowed to perform this action"}`,
		}
		expectedRequestBody := `{"requestType":"createProduct","userKey":"test_user_token","productName":"test_product_name","orgToken":"test_org_token"}`
		sys := System{serverURL: "https://my.test.server", httpClient: &myTestClient, orgToken: "test_org_token", userToken: "test_user_token"}
		// test
		productToken, err := sys.CreateProduct("test_product_name")
		// assert
		assert.EqualError(t, err, "invalid request, error code 5001, message 'User is not allowed to perform this action'")
		requestBody, err := ioutil.ReadAll(myTestClient.requestBody)
		require.NoError(t, err)
		assert.Equal(t, "", productToken)
		assert.Equal(t, expectedRequestBody, string(requestBody))
	})
	t.Run("happy path", func(t *testing.T) {
		// init
		myTestClient := whitesourceMockClient{
			responseBody: `{"productToken":"test_product_token"}`,
		}
		expectedRequestBody := `{"requestType":"createProduct","userKey":"test_user_token","productName":"test_product_name","orgToken":"test_org_token"}`
		sys := System{serverURL: "https://my.test.server", httpClient: &myTestClient, orgToken: "test_org_token", userToken: "test_user_token"}
		// test
		productToken, err := sys.CreateProduct("test_product_name")
		// assert
		assert.NoError(t, err)
		requestBody, err := ioutil.ReadAll(myTestClient.requestBody)
		require.NoError(t, err)
		assert.Equal(t, "test_product_token", productToken)
		assert.Equal(t, expectedRequestBody, string(requestBody))
	})
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

	sys := System{serverURL: "https://my.test.server", httpClient: &myTestClient, orgToken: "test_org_token", userToken: "test_user_token"}
	product, err := sys.GetProductByName("Test Product 2")

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

	sys := System{serverURL: "https://my.test.server", httpClient: &myTestClient, orgToken: "test_org_token", userToken: "test_user_token"}
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

	sys := System{serverURL: "https://my.test.server", httpClient: &myTestClient, orgToken: "test_org_token", userToken: "test_user_token"}

	t.Parallel()

	t.Run("find project 1", func(t *testing.T) {
		projectToken, err := sys.GetProjectToken("test_product_token", "Test Project1")
		assert.NoError(t, err)
		assert.Equal(t, "test_project_token1", projectToken)
	})

	t.Run("find project 2", func(t *testing.T) {
		projectToken, err := sys.GetProjectToken("test_product_token", "Test Project2")
		assert.NoError(t, err)
		assert.Equal(t, "test_project_token2", projectToken)
	})

	t.Run("not finding project 3 is an error", func(t *testing.T) {
		projectToken, err := sys.GetProjectToken("test_product_token", "Test Project3")
		assert.NoError(t, err)
		assert.Equal(t, "", projectToken)
	})
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

	sys := System{serverURL: "https://my.test.server", httpClient: &myTestClient, orgToken: "test_org_token", userToken: "test_user_token"}

	t.Run("success case", func(t *testing.T) {
		projectTokens, err := sys.GetProjectTokens("test_product_token", []string{"Test Project1", "Test Project2"})
		assert.NoError(t, err)
		assert.Equal(t, []string{"test_project_token1", "test_project_token2"}, projectTokens)
	})

	t.Run("no tokens found", func(t *testing.T) {
		projectTokens, err := sys.GetProjectTokens("test_product_token", []string{"Test Project3"})
		assert.Contains(t, fmt.Sprint(err), "no project token(s) found for provided projects")
		assert.Equal(t, []string{}, projectTokens)
	})

	t.Run("not all tokens found", func(t *testing.T) {
		projectTokens, err := sys.GetProjectTokens("test_product_token", []string{"Test Project1", "Test Project3"})
		assert.Contains(t, fmt.Sprint(err), "not all project token(s) found for provided projects")
		assert.Equal(t, []string{"test_project_token1"}, projectTokens)
	})
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

	sys := System{serverURL: "https://my.test.server", httpClient: &myTestClient, orgToken: "test_org_token", userToken: "test_user_token"}

	productName, err := sys.GetProductName("test_product_token")
	assert.NoError(t, err)
	assert.Equal(t, "Test Product", productName)
}

func TestGetProjectsByIDs(t *testing.T) {
	responseBody :=
		`{
	"projectVitals":[
		{
			"id":1,
			"name":"prj-1"
		},
		{
			"id":2,
			"name":"prj-2"
		},
		{
			"id":3,
			"name":"prj-3"
		},
		{
			"id":4,
			"name":"prj-4"
		}
	]
}`

	t.Parallel()

	t.Run("find projects by ids", func(t *testing.T) {
		myTestClient := whitesourceMockClient{responseBody: responseBody}
		sys := System{serverURL: "https://my.test.server", httpClient: &myTestClient, orgToken: "test_org_token", userToken: "test_user_token"}

		projects, err := sys.GetProjectsByIDs("test_product_token", []int64{4, 2})

		assert.NoError(t, err)
		assert.Equal(t, []Project{{ID: 2, Name: "prj-2"}, {ID: 4, Name: "prj-4"}}, projects)
	})

	t.Run("find no projects by ids", func(t *testing.T) {
		myTestClient := whitesourceMockClient{responseBody: responseBody}
		sys := System{serverURL: "https://my.test.server", httpClient: &myTestClient, orgToken: "test_org_token", userToken: "test_user_token"}

		projects, err := sys.GetProjectsByIDs("test_product_token", []int64{5})

		assert.NoError(t, err)
		assert.Equal(t, []Project(nil), projects)
	})
}

func TestGetProjectAlertsByType(t *testing.T) {
	t.Parallel()

	t.Run("success case", func(t *testing.T) {
		responseBody := `{"alerts":[{"type":"SECURITY_VULNERABILITY", "vulnerability":{"name":"testVulnerability1"}}]}`
		myTestClient := whitesourceMockClient{responseBody: responseBody}
		sys := System{serverURL: "https://my.test.server", httpClient: &myTestClient, orgToken: "test_org_token", userToken: "test_user_token"}

		alerts, err := sys.GetProjectAlertsByType("test_project_token", "SECURITY_VULNERABILITY")

		assert.NoError(t, err)
		requestBody, err := ioutil.ReadAll(myTestClient.requestBody)
		assert.NoError(t, err)
		assert.Contains(t, string(requestBody), `"requestType":"getProjectAlertsByType"`)
		assert.Equal(t, []Alert{{Vulnerability: Vulnerability{Name: "testVulnerability1"}}}, alerts)
	})

	t.Run("error case", func(t *testing.T) {
		myTestClient := whitesourceMockClient{requestError: fmt.Errorf("request failed")}
		sys := System{serverURL: "https://my.test.server", httpClient: &myTestClient, orgToken: "test_org_token", userToken: "test_user_token"}

		_, err := sys.GetProjectAlertsByType("test_project_token", "SECURITY_VULNERABILITY")
		assert.EqualError(t, err, "sending whiteSource request failed: failed to send request to WhiteSource: request failed")

	})
}
