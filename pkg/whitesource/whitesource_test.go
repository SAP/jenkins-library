package whitesource

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/package-url/packageurl-go"
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
	return &http.Response{StatusCode: c.httpStatusCode, Body: io.NopCloser(bytes.NewReader([]byte(c.responseBody)))}, nil
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
	assert.NoError(t, err)

	requestBody, err := io.ReadAll(myTestClient.requestBody)
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
		assert.EqualError(t, err, "WhiteSource request failed after 3 retries: invalid request, error code 3000, message 'WhiteSource backend has a hickup'")
		requestBody, err := io.ReadAll(myTestClient.requestBody)
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
		requestBody, err := io.ReadAll(myTestClient.requestBody)
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
		requestBody, err := io.ReadAll(myTestClient.requestBody)
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
	assert.NoError(t, err)

	requestBody, err := io.ReadAll(myTestClient.requestBody)
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

func TestTransformLibToPurlType(t *testing.T) {
	tt := []struct {
		libType  string
		expected string
	}{
		{libType: "Java", expected: packageurl.TypeMaven},
		{libType: "MAVEN_ARTIFACT", expected: packageurl.TypeMaven},
		{libType: "javascript/Node.js", expected: packageurl.TypeNPM},
		{libType: "node_packaged_module", expected: packageurl.TypeNPM},
		{libType: "javascript/bower", expected: "bower"},
		{libType: "go", expected: packageurl.TypeGolang},
		{libType: "go_package", expected: packageurl.TypeGolang},
		{libType: "python", expected: packageurl.TypePyPi},
		{libType: "python_package", expected: packageurl.TypePyPi},
		{libType: "debian", expected: packageurl.TypeDebian},
		{libType: "debian_package", expected: packageurl.TypeDebian},
		{libType: "docker", expected: packageurl.TypeDocker},
		{libType: ".net", expected: packageurl.TypeNuget},
		{libType: "dot_net_resource", expected: packageurl.TypeNuget},
	}

	for i, test := range tt {
		assert.Equalf(t, test.expected, transformLibToPurlType(test.libType), "run %v failed", i)
	}
}

func TestGetProjectHierarchy(t *testing.T) {
	myTestClient := whitesourceMockClient{
		responseBody: `{
	"libraries": [
		{
			"keyUuid": "1f9ee6ec-eded-45d3-8fdb-2d0d735e5b14",
			"keyId": 43,
			"filename": "log4j-1.2.17.jar",
			"name": "log4j",
			"groupId": "log4j",
			"artifactId": "log4j",
			"version": "1.2.17",
			"sha1": "5af35056b4d257e4b64b9e8069c0746e8b08629f",
			"type": "UNKNOWN_ARTIFACT",
			"coordinates": "log4j:log4j:1.2.17"
		},
		{
			"keyUuid": "f362c53f-ce25-4d0c-b53b-ee2768b32d1a",
			"keyId": 45,
			"filename": "akka-actor_2.11-2.5.2.jar",
			"name": "akka-actor",
			"groupId": "com.typesafe.akka",
			"artifactId": "akka-actor_2.11",
			"version": "2.5.2",
			"sha1": "183ccaed9002bfa10628a5df48e7bac6f1c03f7b",
			"type": "MAVEN_ARTIFACT",
			"coordinates": "com.typesafe.akka:akka-actor_2.11:2.5.2",
			"dependencies": [
				{
					"keyUuid": "49c6840d-bf96-470f-8892-6c2a536c91eb",
					"keyId": 44,
					"filename": "scala-library-2.11.11.jar",
					"name": "Scala Library",
					"groupId": "org.scala-lang",
					"artifactId": "scala-library",
					"version": "2.11.11",
					"sha1": "e283d2b7fde6504f6a86458b1f6af465353907cc",
					"type": "MAVEN_ARTIFACT",
					"coordinates": "org.scala-lang:scala-library:2.11.11"
				},
				{
					"keyUuid": "e5e730d1-8b41-4d2d-a8c5-610a374b6501",
					"keyId": 46,
					"filename": "scala-java8-compat_2.11-0.7.0.jar",
					"name": "scala-java8-compat_2.11",
					"groupId": "org.scala-lang.modules",
					"artifactId": "scala-java8-compat_2.11",
					"version": "0.7.0",
					"sha1": "a31b1b36bcf0d53657733b5d40c78d5f090a5dea",
					"type": "UNKNOWN_ARTIFACT",
					"coordinates": "org.scala-lang.modules:scala-java8-compat_2.11:0.7.0"
				},
				{
					"keyUuid": "426c0056-f180-4cac-a9dd-c266a76b32c9",
					"keyId": 47,
					"filename": "config-1.3.1.jar",
					"name": "config",
					"groupId": "com.typesafe",
					"artifactId": "config",
					"version": "1.3.1",
					"sha1": "2cf7a6cc79732e3bdf1647d7404279900ca63eb0",
					"type": "UNKNOWN_ARTIFACT",
					"coordinates": "com.typesafe:config:1.3.1"
				}
			]
		},
		{
			"keyUuid": "25a8ceaa-4548-4fe4-9819-8658b8cbe9aa",
			"keyId": 48,
			"filename": "kafka-clients-0.10.2.1.jar",
			"name": "Apache Kafka",
			"groupId": "org.apache.kafka",
			"artifactId": "kafka-clients",
			"version": "0.10.2.1",
			"sha1": "3dd2aa4c9f87ac54175d017bcb63b4bb5dca63dd",
			"type": "MAVEN_ARTIFACT",
			"coordinates": "org.apache.kafka:kafka-clients:0.10.2.1",
			"dependencies": [
				{
					"keyUuid": "71065ffb-e509-4e2d-88bc-9184bc50888d",
					"keyId": 49,
					"filename": "lz4-1.3.0.jar",
					"name": "LZ4 and xxHash",
					"groupId": "net.jpountz.lz4",
					"artifactId": "lz4",
					"version": "1.3.0",
					"sha1": "c708bb2590c0652a642236ef45d9f99ff842a2ce",
					"type": "MAVEN_ARTIFACT",
					"coordinates": "net.jpountz.lz4:lz4:1.3.0"
				},
				{
					"keyUuid": "e44ab569-de95-4562-8efa-a2ebfe808471",
					"keyId": 50,
					"filename": "slf4j-api-1.7.21.jar",
					"name": "SLF4J API Module",
					"groupId": "org.slf4j",
					"artifactId": "slf4j-api",
					"version": "1.7.21",
					"sha1": "139535a69a4239db087de9bab0bee568bf8e0b70",
					"type": "MAVEN_ARTIFACT",
					"coordinates": "org.slf4j:slf4j-api:1.7.21"
				},
				{
					"keyUuid": "72ecad5e-9f35-466c-9ed8-0974e7ce4e29",
					"keyId": 51,
					"filename": "snappy-java-1.1.2.6.jar",
					"name": "snappy-java",
					"groupId": "org.xerial.snappy",
					"artifactId": "snappy-java",
					"version": "1.1.2.6",
					"sha1": "48d92871ca286a47f230feb375f0bbffa83b85f6",
					"type": "UNKNOWN_ARTIFACT",
					"coordinates": "org.xerial.snappy:snappy-java:1.1.2.6"
				}
			]
		}
	],
	"warningMessages":[
      "Invalid input: orgToken"
    ]
}`,
	}

	sys := System{serverURL: "https://my.test.server", httpClient: &myTestClient, orgToken: "test_org_token", userToken: "test_user_token"}

	libraries, err := sys.GetProjectHierarchy("test_project_token", true)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(libraries))
	assert.Nil(t, libraries[0].Dependencies)
	assert.NotNil(t, libraries[1].Dependencies)
	assert.Equal(t, 3, len(libraries[1].Dependencies))
	assert.NotNil(t, libraries[2].Dependencies)
	assert.Equal(t, 3, len(libraries[2].Dependencies))
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
		requestBody, err := io.ReadAll(myTestClient.requestBody)
		assert.NoError(t, err)
		assert.Contains(t, string(requestBody), `"requestType":"getProjectAlertsByType"`)
		assert.Equal(t, []Alert{{Vulnerability: Vulnerability{Name: "testVulnerability1"}, Type: "SECURITY_VULNERABILITY"}}, alerts)
	})

	t.Run("error case", func(t *testing.T) {
		myTestClient := whitesourceMockClient{requestError: fmt.Errorf("request failed")}
		sys := System{serverURL: "https://my.test.server", httpClient: &myTestClient, orgToken: "test_org_token", userToken: "test_user_token"}

		_, err := sys.GetProjectAlertsByType("test_project_token", "SECURITY_VULNERABILITY")
		assert.EqualError(t, err, "sending whiteSource request failed: failed to send request to WhiteSource: request failed")

	})
}
