package blackduck

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
)

type httpMockClient struct {
	responseBodyForURL map[string]string
	errorMessageForURL map[string]string
	header             map[string]http.Header
}

func (c *httpMockClient) SetOptions(opts piperhttp.ClientOptions) {}
func (c *httpMockClient) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	c.header[url] = header
	response := http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte(""))),
	}

	if c.errorMessageForURL[url] != "" {
		response.StatusCode = 400
		return &response, fmt.Errorf("%s", c.errorMessageForURL[url])
	}

	if c.responseBodyForURL[url] != "" {
		response.Body = io.NopCloser(bytes.NewReader([]byte(c.responseBodyForURL[url])))
		return &response, nil
	}

	return &response, nil
}

const (
	authContent = `{
		"bearerToken":"bearerTestToken",
		"expiresInMilliseconds":7199997
	}`
	projectContent = `{
		"totalCount": 1,
		"items": [
			{
				"name": "SHC-PiperTest",
				"_meta": {
					"href": "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf",
					"links": [
						{
							"rel": "versions",
							"href": "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions"
						}
					]
				}
			}
		]
	}`
	projectVersionContent = `{
		"totalCount": 1,
		"items": [
			{
				"versionName": "1.0",
				"_meta": {
					"href": "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36",
					"links": [
						{
							"rel": "components",
							"href": "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36/components"
						},
						{
							"rel": "vulnerable-components",
							"href": "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36/vunlerable-bom-components"
						},
						{
							"rel": "policy-status",
							"href": "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36/policy-status"
						}
					]
				}
			}
		]
	}`
)

func TestGetProject(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":                                        authContent,
				"https://my.blackduck.system/api/projects?limit=50&offset=0&q=name%3ASHC-PiperTest&sort=asc": projectContent,
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		project, err := bdClient.GetProject("SHC-PiperTest")
		assert.NoError(t, err)
		assert.Equal(t, "SHC-PiperTest", project.Name)
		assert.Equal(t, "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions", project.Metadata.Links[0].Href)
		headerExpected := http.Header{"Authorization": []string{"Bearer bearerTestToken"}, "Accept": {"application/vnd.blackducksoftware.project-detail-4+json"}}
		assert.Equal(t, headerExpected, myTestClient.header["https://my.blackduck.system/api/projects?limit=50&offset=0&q=name%3ASHC-PiperTest&sort=asc"])
	})

	t.Run("failure - not found", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate": authContent,
			},
			errorMessageForURL: map[string]string{
				"https://my.blackduck.system/api/projects?limit=50&offset=0&q=name%3ASHC-PiperTest&sort=asc": "not found",
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		_, err := bdClient.GetProject("SHC-PiperTest")
		assert.Contains(t, fmt.Sprint(err), "failed to get project 'SHC-PiperTest'")
	})

	t.Run("failure - 0 results", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate": authContent,
				"https://my.blackduck.system/api/projects?limit=50&offset=0&q=name%3ASHC-PiperTest&sort=asc": `{
					"totalCount": 0,
					"items": []
				}`,
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		_, err := bdClient.GetProject("SHC-PiperTest")
		assert.Contains(t, fmt.Sprint(err), "project 'SHC-PiperTest' not found")
	})

	t.Run("failure - unmarshalling", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":                                        authContent,
				"https://my.blackduck.system/api/projects?limit=50&offset=0&q=name%3ASHC-PiperTest&sort=asc": "",
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		_, err := bdClient.GetProject("SHC-PiperTest")
		assert.Contains(t, fmt.Sprint(err), "failed to retrieve details for project 'SHC-PiperTest'")
	})
}

func TestGetProjectVersion(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":                                                       authContent,
				"https://my.blackduck.system/api/projects?limit=50&offset=0&q=name%3ASHC-PiperTest&sort=asc":                projectContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions?limit=100&offset=0": projectVersionContent,
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		projectVersion, err := bdClient.GetProjectVersion("SHC-PiperTest", "1.0")
		assert.NoError(t, err)
		assert.Equal(t, "1.0", projectVersion.Name)
		assert.Equal(t, "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36", projectVersion.Metadata.Href)
		headerExpected := http.Header{"Authorization": []string{"Bearer bearerTestToken"}, "Accept": {"application/vnd.blackducksoftware.project-detail-4+json"}}
		assert.Equal(t, headerExpected, myTestClient.header["https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions?limit=100&offset=0"])
	})

	t.Run("failure - project not found", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate": authContent,
			},
			errorMessageForURL: map[string]string{
				"https://my.blackduck.system/api/projects?limit=50&offset=0&q=name%3ASHC-PiperTest&sort=asc": "not found",
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		_, err := bdClient.GetProjectVersion("SHC-PiperTest", "1.0")
		assert.Contains(t, fmt.Sprint(err), "failed to get project 'SHC-PiperTest'")
	})

	t.Run("failure - version not found", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":                                        authContent,
				"https://my.blackduck.system/api/projects?limit=50&offset=0&q=name%3ASHC-PiperTest&sort=asc": projectContent,
			},
			errorMessageForURL: map[string]string{
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions?limit=100&offset=0": "not found",
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		_, err := bdClient.GetProjectVersion("SHC-PiperTest", "1.0")
		assert.Contains(t, fmt.Sprint(err), "failed to get project version 'SHC-PiperTest:1.0'")
	})

	t.Run("failure - 0 results", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":                                        authContent,
				"https://my.blackduck.system/api/projects?limit=50&offset=0&q=name%3ASHC-PiperTest&sort=asc": projectContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions?limit=100&offset=0": `{
					"totalCount": 0,
					"items": []
				}`,
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		_, err := bdClient.GetProjectVersion("SHC-PiperTest", "1.0")
		assert.Contains(t, fmt.Sprint(err), "project version 'SHC-PiperTest:1.0' not found")
	})

	t.Run("failure - unmarshalling", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":                                                       authContent,
				"https://my.blackduck.system/api/projects?limit=50&offset=0&q=name%3ASHC-PiperTest&sort=asc":                projectContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions?limit=100&offset=0": "",
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		_, err := bdClient.GetProjectVersion("SHC-PiperTest", "1.0")
		assert.Contains(t, fmt.Sprint(err), "failed to retrieve details for project version 'SHC-PiperTest:1.0'")
	})
}

func TestGetVulnerabilities(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":                                                       authContent,
				"https://my.blackduck.system/api/projects?limit=50&offset=0&q=name%3ASHC-PiperTest&sort=asc":                projectContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions?limit=100&offset=0": projectVersionContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36/vunlerable-bom-components?limit=999&offset=0": `{
					"totalCount": 1,
					"items": [
						{
							"componentName": "Spring Framework",
							"componentVersionName": "5.3.2",
							"vulnerabilityWithRemediation" : {
								"vulnerabilityName" : "BDSA-2019-2021",
								"baseScore" : 1.0,
      							"overallScore" : 1.0,
								"severity" : "HIGH",
								"remediationStatus" : "IGNORED",
								"description" : "description"
							}
						}
					]
				}`,
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("token", "https://my.blackduck.system", &myTestClient)
		vulns, err := bdClient.GetVulnerabilities("SHC-PiperTest", "1.0")
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, vulns.TotalCount, 1)
	})

	t.Run("Success - 0 vulns", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":                                                                                                                      authContent,
				"https://my.blackduck.system/api/projects?limit=50&offset=0&q=name%3ASHC-PiperTest&sort=asc":                                                                               projectContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions?limit=100&offset=0":                                                                projectVersionContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36/vunlerable-bom-components?limit=999&offset=0": `{"totalCount":0,"items":[]}`,
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("token", "https://my.blackduck.system", &myTestClient)
		vulns, err := bdClient.GetVulnerabilities("SHC-PiperTest", "1.0")
		assert.NoError(t, err)
		assert.Equal(t, vulns.TotalCount, 0)
	})

	t.Run("Failure - unmarshalling", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":                                                                                                                      authContent,
				"https://my.blackduck.system/api/projects?limit=50&offset=0&q=name%3ASHC-PiperTest&sort=asc":                                                                               projectContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions?limit=100&offset=0":                                                                projectVersionContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36/vunlerable-bom-components?limit=999&offset=0": "",
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("token", "https://my.blackduck.system", &myTestClient)
		_, err := bdClient.GetVulnerabilities("SHC-PiperTest", "1.0")
		assert.Contains(t, fmt.Sprint(err), "failed to retrieve Vulnerability details for project version 'SHC-PiperTest:1.0'")
	})
}

func TestGetComponents(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":                                                       authContent,
				"https://my.blackduck.system/api/projects?limit=50&offset=0&q=name%3ASHC-PiperTest&sort=asc":                projectContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions?limit=100&offset=0": projectVersionContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36/components?limit=999&offset=0": `{
					"totalCount": 2,
					"items" : [
						{
							"componentName": "Spring Framework",
							"componentVersionName": "5.3.9"
						}, {
							"componentName": "Apache Tomcat",
							"componentVersionName": "9.0.52"
						}
					]
				}`,
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("token", "https://my.blackduck.system", &myTestClient)
		components, err := bdClient.GetComponents("SHC-PiperTest", "1.0")
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, components.TotalCount, 2)
	})

	t.Run("Failure - 0 components", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":                                                       authContent,
				"https://my.blackduck.system/api/projects?limit=50&offset=0&q=name%3ASHC-PiperTest&sort=asc":                projectContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions?limit=100&offset=0": projectVersionContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36/components?limit=999&offset=0": `{
					"totalCount": 0,
					"items" : []}`,
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("token", "https://my.blackduck.system", &myTestClient)
		components, err := bdClient.GetComponents("SHC-PiperTest", "1.0")
		assert.Contains(t, fmt.Sprint(err), "No Components found for project version 'SHC-PiperTest:1.0'")
		assert.Nilf(t, components, "Expected Components to be nil")
	})

	t.Run("Failure - unmarshalling", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":                                                                                                       authContent,
				"https://my.blackduck.system/api/projects?limit=50&offset=0&q=name%3ASHC-PiperTest&sort=asc":                                                                projectContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions?limit=100&offset=0":                                                 projectVersionContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36/components?limit=999&offset=0": "",
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("token", "https://my.blackduck.system", &myTestClient)
		components, err := bdClient.GetComponents("SHC-PiperTest", "1.0")
		assert.Contains(t, fmt.Sprint(err), "failed to retrieve component details for project version 'SHC-PiperTest:1.0'")
		assert.Nilf(t, components, "Expected Components to be nil")
	})
}

func TestGetComponentsWithLicensePolicyRule(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":                                                       authContent,
				"https://my.blackduck.system/api/projects?limit=50&offset=0&q=name%3ASHC-PiperTest&sort=asc":                projectContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions?limit=100&offset=0": projectVersionContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36/components?filter=policyCategory%3Alicense&limit=999&offset=0": `{
					"totalCount": 2,
					"items" : [
						{
							"componentName": "Spring Framework",
							"componentVersionName": "5.3.9",
							"policyStatus": "IN_VIOLATION"
						}, {
							"componentName": "Apache Tomcat",
							"componentVersionName": "9.0.52",
							"policyStatus": "NOT_IN_VIOLATION"
						}
					]
				}`,
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("token", "https://my.blackduck.system", &myTestClient)
		components, err := bdClient.GetComponentsWithLicensePolicyRule("SHC-PiperTest", "1.0")
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, components.TotalCount, 2)
		assert.Equal(t, components.Items[0].PolicyStatus, "IN_VIOLATION")
	})

	t.Run("Failure - 0 components", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":                                                       authContent,
				"https://my.blackduck.system/api/projects?limit=50&offset=0&q=name%3ASHC-PiperTest&sort=asc":                projectContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions?limit=100&offset=0": projectVersionContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36/components?filter=policyCategory%3Alicense&limit=999&offset=0": `{
					"totalCount": 0,
					"items" : []}`,
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("token", "https://my.blackduck.system", &myTestClient)
		components, err := bdClient.GetComponentsWithLicensePolicyRule("SHC-PiperTest", "1.0")
		assert.NoError(t, err)
		assert.NotNil(t, components)
		assert.Equal(t, components.TotalCount, 0)
	})

	t.Run("Failure - unmarshalling", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":                                                                                                                                       authContent,
				"https://my.blackduck.system/api/projects?limit=50&offset=0&q=name%3ASHC-PiperTest&sort=asc":                                                                                                projectContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions?limit=100&offset=0":                                                                                 projectVersionContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36/components?filter=policyCategory%3Alicense&limit=999&offset=0": "",
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("token", "https://my.blackduck.system", &myTestClient)
		components, err := bdClient.GetComponentsWithLicensePolicyRule("SHC-PiperTest", "1.0")
		assert.Contains(t, fmt.Sprint(err), "failed to retrieve component details for project version 'SHC-PiperTest:1.0'")
		assert.Nilf(t, components, "Expected Components to be nil")
	})
}

func TestGetPolicyStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":                                                       authContent,
				"https://my.blackduck.system/api/projects?limit=50&offset=0&q=name%3ASHC-PiperTest&sort=asc":                projectContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions?limit=100&offset=0": projectVersionContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36/policy-status": `{
					"overallStatus": "IN_VIOLATION",
					"componentVersionPolicyViolationDetails": {
						"name": "IN_VIOLATION",
						"severityLevels": [
						  {	"name": "BLOCKER", "value": 16 },
						  { "name": "CRITICAL", "value": 1 },
						  { "name": "MAJOR", "value": 0 },
						  { "name": "MINOR", "value": 0 },
						  { "name": "TRIVIAL", "value": 0 },
						  { "name": "UNSPECIFIED", "value": 0 },
						  { "name": "OK", "value": 0}
						]
					  }
				}`,
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("token", "https://my.blackduck.system", &myTestClient)
		policyStatus, err := bdClient.GetPolicyStatus("SHC-PiperTest", "1.0")
		assert.NoError(t, err)
		assert.Equal(t, policyStatus.OverallStatus, "IN_VIOLATION")
		assert.Equal(t, len(policyStatus.PolicyVersionDetails.SeverityLevels), 7)
		assert.Equal(t, policyStatus.PolicyVersionDetails.Name, "IN_VIOLATION")
	})

	t.Run("Failure - unmarshalling", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":                                                                                       authContent,
				"https://my.blackduck.system/api/projects?limit=50&offset=0&q=name%3ASHC-PiperTest&sort=asc":                                                projectContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions?limit=100&offset=0":                                 projectVersionContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36/policy-status": "",
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("token", "https://my.blackduck.system", &myTestClient)
		policyStatus, err := bdClient.GetPolicyStatus("SHC-PiperTest", "1.0")
		assert.Contains(t, fmt.Sprint(err), "failed to retrieve Policy violation details for project version 'SHC-PiperTest:1.0'")
		assert.Nilf(t, policyStatus, "Expected Components to be nil")
	})
}

func TestGetProjectVersionLink(t *testing.T) {
	t.Run("Success Case", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate":                                                       authContent,
				"https://my.blackduck.system/api/projects?limit=50&offset=0&q=name%3ASHC-PiperTest&sort=asc":                projectContent,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions?limit=100&offset=0": projectVersionContent,
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("token", "https://my.blackduck.system", &myTestClient)
		link, err := bdClient.GetProjectVersionLink("SHC-PiperTest", "1.0")
		assert.NoError(t, err)
		assert.Equal(t, link, "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36")
	})
}

func TestAuthenticate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate": authContent,
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		err := bdClient.authenticate()
		assert.NoError(t, err)
		headerExpected := http.Header{"Authorization": {"token myTestToken"}, "Accept": {"application/vnd.blackducksoftware.user-4+json"}}
		assert.Equal(t, headerExpected, myTestClient.header["https://my.blackduck.system/api/tokens/authenticate"])
	})

	t.Run("authentication failure", func(t *testing.T) {
		myTestClient := httpMockClient{
			errorMessageForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate": "not authorized",
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		err := bdClient.authenticate()
		assert.EqualError(t, err, "authentication to BlackDuck API failed: request to BlackDuck API failed: not authorized")
	})

	t.Run("parse failure", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.blackduck.system/api/tokens/authenticate": "",
			},
			header: map[string]http.Header{},
		}
		bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)
		err := bdClient.authenticate()
		assert.Contains(t, fmt.Sprint(err), "failed to parse BlackDuck response")
	})
}

func TestSendRequest(t *testing.T) {
	myTestClient := httpMockClient{
		responseBodyForURL: map[string]string{
			"https://my.blackduck.system/api/endpoint":        "testContent",
			"https://my.blackduck.system/api/endpoint?q=test": "testContentQuery",
		},
		header: map[string]http.Header{},
	}
	bdClient := NewClient("myTestToken", "https://my.blackduck.system", &myTestClient)

	t.Run("simple", func(t *testing.T) {
		responseBody, err := bdClient.sendRequest(http.MethodGet, "api/endpoint", map[string]string{}, nil, http.Header{})
		assert.NoError(t, err)
		assert.Equal(t, "testContent", string(responseBody))
	})

	t.Run("with query params and bearer", func(t *testing.T) {
		bdClient.BearerToken = "testBearer"
		responseBody, err := bdClient.sendRequest(http.MethodGet, "api/endpoint", map[string]string{"q": "test"}, nil, http.Header{})
		assert.NoError(t, err)
		assert.Equal(t, "testContentQuery", string(responseBody))
		assert.Equal(t, http.Header{"Authorization": {"Bearer testBearer"}}, myTestClient.header["https://my.blackduck.system/api/endpoint?q=test"])
	})
}

func TestApiURL(t *testing.T) {
	tt := []struct {
		description string
		apiEndpoint string
		client      Client
		expected    string
	}{
		{
			description: "trailing / in path",
			apiEndpoint: "/my/path/",
			client:      Client{serverURL: "https://my.test.server"},
			expected:    "https://my.test.server/my/path",
		},
		{
			description: "trailing / in server",
			apiEndpoint: "/my/path",
			client:      Client{serverURL: "https://my.test.server/"},
			expected:    "https://my.test.server/my/path",
		},
	}

	for _, test := range tt {
		res, err := test.client.apiURL(test.apiEndpoint)
		assert.NoError(t, err)
		assert.Equalf(t, test.expected, res.String(), test.description)
	}
}

func TestAuthenticationValid(t *testing.T) {
	tt := []struct {
		description string
		now         time.Time
		client      Client
		expected    bool
	}{
		{
			description: "login still valid",
			now:         time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC),
			client:      Client{BearerExpiresInMilliseconds: 120000, lastAuthentication: time.Date(2020, time.January, 01, 11, 59, 0, 0, time.UTC)},
			expected:    true,
		},
		{
			description: "login still valid - edge",
			now:         time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC),
			client:      Client{BearerExpiresInMilliseconds: 120000, lastAuthentication: time.Date(2020, time.January, 01, 11, 58, 1, 0, time.UTC)},
			expected:    true,
		},
		{
			description: "login expired",
			now:         time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC),
			client:      Client{BearerExpiresInMilliseconds: 120000, lastAuthentication: time.Date(2020, time.January, 01, 11, 57, 0, 0, time.UTC)},
			expected:    false,
		},
		{
			description: "login expired - edge",
			now:         time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC),
			client:      Client{BearerExpiresInMilliseconds: 120000, lastAuthentication: time.Date(2020, time.January, 01, 11, 58, 0, 0, time.UTC)},
			expected:    false,
		},
	}

	for _, test := range tt {
		assert.Equalf(t, test.expected, test.client.authenticationValid(test.now), test.description)
	}
}

func TestUrlPath(t *testing.T) {
	assert.Equal(t, "/this/is/the/path", urlPath("https://the.server.domain:8080/this/is/the/path"))
}

func TestTransformComponentOriginToPurlParts(t *testing.T) {
	tt := []struct {
		description string
		component   *Component
		expected    []string
	}{
		{
			// pkg:maven/org.apache.cxf/cxf-rt-rs-client@3.1.2
			description: "Origin with Url type, namespace, name, version",
			component: &Component{
				Name:    "Apache CXF",
				Version: "3.1.2",
				Origins: []ComponentOrigin{{
					ExternalNamespace: "maven",
					ExternalID:        "org.apache.cxf:cxf-rt-rs-client:3.1.2",
				}},
			},
			expected: []string{"maven", "org.apache.cxf", "cxf-rt-rs-client", "3.1.2"},
		},
		{
			// pkg:npm/minimist@0.0.8
			description: "Origin with Url type, name, version",
			component: &Component{
				Name:    "Minimist",
				Version: "0.0.8",
				Origins: []ComponentOrigin{{
					ExternalNamespace: "npmjs",
					ExternalID:        "minimist/0.0.8",
				}},
			},
			expected: []string{"npm", "minimist", "0.0.8"},
		},
		{
			// pkg:maven/org.springframework/spring-expression@4.1.6.RELEASE
			description: "Empty origin",
			component: &Component{
				Name:    "spring-expression",
				Version: "4.1.6.RELEASE",
				Origins: []ComponentOrigin{},
			},
			expected: []string{"generic", "", "spring-expression", "4.1.6.RELEASE"},
		},
		{
			// pkg:debian/libpython3.9-stdlib@3.9.2-1
			description: "Component with specified architecture",
			component: &Component{
				Name:    "Python programming language",
				Version: "3.9.2",
				Origins: []ComponentOrigin{
					{
						ExternalNamespace: "debian",
						ExternalID:        "libpython3.9-stdlib/3.9.2-1/amd64",
					},
				},
			},
			expected: []string{"debian", "libpython3.9-stdlib", "3.9.2-1"},
		},
	}

	for _, test := range tt {
		t.Run(test.description, func(t *testing.T) {
			got := transformComponentOriginToPurlParts(test.component)

			assert.Equal(t, test.expected, got)
		})
	}
}

func TestComponentToPackageUrl(t *testing.T) {
	tt := []struct {
		description string
		component   *Component
		expected    string
	}{
		{
			description: "Origin with Url type, namespace, name, version",
			component: &Component{
				Name:    "Apache CXF",
				Version: "3.1.2",
				Origins: []ComponentOrigin{{
					ExternalNamespace: "maven",
					ExternalID:        "org.apache.cxf:cxf-rt-rs-client:3.1.2",
				}},
			},
			expected: "pkg:maven/org.apache.cxf/cxf-rt-rs-client@3.1.2",
		},
		{
			description: "Origin with Url type, name, version",
			component: &Component{
				Name:    "Minimist",
				Version: "0.0.8",
				Origins: []ComponentOrigin{{
					ExternalNamespace: "npmjs",
					ExternalID:        "minimist/0.0.8",
				}},
			},
			expected: "pkg:npm/minimist@0.0.8",
		},
		{
			description: "Empty origin",
			component: &Component{
				Name:    "spring-expression",
				Version: "4.1.6.RELEASE",
				Origins: []ComponentOrigin{},
			},
			expected: "pkg:generic/spring-expression@4.1.6.RELEASE",
		},
	}

	for _, test := range tt {
		t.Run(test.description, func(t *testing.T) {
			got := test.component.ToPackageUrl().ToString()

			assert.Equal(t, test.expected, got)
		})
	}
}
