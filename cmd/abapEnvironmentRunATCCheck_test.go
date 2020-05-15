package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostConfig(t *testing.T) {
	t.Run("Check Host: ABAP Endpoint", func(t *testing.T) {
		config := abapEnvironmentRunATCCheckOptions{
			Username: "testUser",
			Password: "testPassword",
			Host:     "https://api.endpoint.com",
		}
		var con connectionDetailsHTTP
		con, error := checkHost(config, con)
		if error == nil {
			assert.Equal(t, "testUser", con.User)
			assert.Equal(t, "testPassword", con.Password)
			assert.Equal(t, "https://api.endpoint.com", con.URL)
			assert.Equal(t, "", con.XCsrfToken)
		}
	})
	t.Run("No host/ServiceKey configuration", func(t *testing.T) {
		//Testing without CfOrg parameter
		config := abapEnvironmentRunATCCheckOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
		}
		var con connectionDetailsHTTP
		con, err := checkHost(config, con)
		assert.EqualError(t, err, "Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510")
		//Testing without ABAP Host
		config = abapEnvironmentRunATCCheckOptions{
			Username: "testUser",
			Password: "testPassword",
		}
		con, err = checkHost(config, con)
		assert.EqualError(t, err, "Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510")
	})
	t.Run("Check Host: CF Service Key", func(t *testing.T) {
		config := abapEnvironmentRunATCCheckOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfSpace:           "testSpace",
			CfOrg:             "Test",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
		}
		var con connectionDetailsHTTP
		con, error := checkHost(config, con)
		if error == nil {
			assert.Equal(t, "", con.User)
			assert.Equal(t, "", con.Password)
			assert.Equal(t, "", con.URL)
			assert.Equal(t, "", con.XCsrfToken)
		}
	})

}

func TestATCTrigger(t *testing.T) {
	t.Run("Trigger ATC run test", func(t *testing.T) {
		tokenExpected := "myToken"

		client := &clientMock{
			Body:  `ATC trigger test`,
			Token: tokenExpected,
		}

		con := connectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}
		resp, error := runATC("GET", con, []byte(client.Body), client)
		if error == nil {
			assert.Equal(t, tokenExpected, resp.Header["X-Csrf-Token"][0])
			assert.Equal(t, int64(0), resp.ContentLength)
			assert.Equal(t, []string([]string(nil)), resp.Header["Location"])
		}
	})
}

func TestFetchXcsrfToken(t *testing.T) {
	t.Run("FetchXcsrfToken Test", func(t *testing.T) {
		tokenExpected := "myToken"

		client := &clientMock{
			Body:  `Xcsrf Token test`,
			Token: tokenExpected,
		}

		con := connectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}
		token, error := fetchXcsrfToken("GET", con, []byte(client.Body), client)
		if error == nil {
			assert.Equal(t, tokenExpected, token)
		}
	})
	t.Run("failure case: fetch token", func(t *testing.T) {
		tokenExpected := ""

		client := &clientMock{
			Body:  `Xcsrf Token test`,
			Token: "",
		}

		con := connectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}
		token, error := fetchXcsrfToken("GET", con, []byte(client.Body), client)
		if error == nil {
			assert.Equal(t, tokenExpected, token)
		}
	})
}

func TestPollATCRun(t *testing.T) {
	t.Run("ATC run Poll Test", func(t *testing.T) {
		tokenExpected := "myToken"

		client := &clientMock{
			Body:  `ATC Poll test`,
			Token: tokenExpected,
		}

		con := connectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}
		resp, err := pollATCRun(con, []byte(client.Body), client)
		if err != nil {
			assert.Equal(t, "", resp)
			assert.EqualError(t, err, "Could not get any response from ATC poll: Status from ATC run is empty. Either it's not an ABAP system or ATC run hasn't started")

		}
	})
}

func TestGetHTTPResponseATCRun(t *testing.T) {
	t.Run("Get HTTP Response from ATC run Test", func(t *testing.T) {
		client := &clientMock{
			Body: `HTTP response test`,
		}

		con := connectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}
		resp, err := getHTTPResponseATCRun("GET", con, []byte(client.Body), client)
		defer resp.Body.Close()
		if err == nil {
			assert.Equal(t, int64(0), resp.ContentLength)
			assert.Equal(t, []string([]string(nil)), resp.Header["X-Crsf-Token"])
		}
	})
}

func TestGetResultATCRun(t *testing.T) {
	t.Run("Get HTTP Response from ATC run Test", func(t *testing.T) {
		client := &clientMock{
			BodyList: []string{
				`ATC result body`,
			},
		}

		con := connectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}
		resp, err := getResultATCRun("GET", con, []byte(client.Body), client)
		defer resp.Body.Close()
		if err == nil {
			assert.Equal(t, int64(0), resp.ContentLength)
			assert.Equal(t, []string([]string(nil)), resp.Header["X-Crsf-Token"])
		}
	})
}

func TestParseATCResult(t *testing.T) {
	t.Run("succes case: test parsing example XML result", func(t *testing.T) {
		bodyString := `<?xml version="1.0" encoding="UTF-8"?>
		<checkstyle>
			<file name="testFile">
				<error message="testMessage">
				</error>
				<error message="testMessage2">
				</error>
			</file>
			<file name="testFile2">
				<error message="testMessage3">
				</error>
			</file>
		</checkstyle>`
		body := []byte(bodyString)

		err := parseATCResult(body)
		assert.Equal(t, nil, err)
	})
	t.Run("failure case: parsing empty xml", func(t *testing.T) {
		var bodyString string
		body := []byte(bodyString)

		err := parseATCResult(body)
		assert.EqualError(t, err, "Parsing ATC result failed: Body is empty, can't parse empty body")
	})
}

func TestBuildATCCheckBody(t *testing.T) {
	t.Run("Test build body with no software component and package", func(t *testing.T) {
		expectedpackagestring := ""
		expectedsoftwarecomponentstring := ""

		var err error
		var config ATCconfig
		var packageString, softwarecomponentString string

		packageString, softwarecomponentString, err = buildATCCheckBody(config)

		assert.Equal(t, expectedpackagestring, packageString)
		assert.Equal(t, expectedsoftwarecomponentstring, softwarecomponentString)
		assert.EqualError(t, err, "Error while parsing ATC run config. Please provide the packages and/or the software components to be checked! No Package or Software Component specified. Please provide either one or both of them")
	})
	t.Run("success case: Test build body with example yaml config", func(t *testing.T) {
		expectedpackagestring := "<obj:packages><obj:package value=\"testPackage\" includeSubpackages=\"true\"/><obj:package value=\"testPackage2\" includeSubpackages=\"false\"/></obj:packages>"
		expectedsoftwarecomponentstring := "<obj:softwarecomponents><obj:softwarecomponent value=\"testSoftwareComponent\"/><obj:softwarecomponent value=\"testSoftwareComponent2\"/></obj:softwarecomponents>"

		var err error
		var config ATCconfig

		config = ATCconfig{
			ATCObjects{
				Package: []Package{
					Package{Name: "testPackage", IncludeSubpackages: true},
					Package{Name: "testPackage2", IncludeSubpackages: false},
				},
				SoftwareComponent: []SoftwareComponent{
					SoftwareComponent{Name: "testSoftwareComponent"},
					SoftwareComponent{Name: "testSoftwareComponent2"},
				},
			},
		}

		var packageString, softwarecomponentString string

		packageString, softwarecomponentString, err = buildATCCheckBody(config)

		assert.Equal(t, expectedpackagestring, packageString)
		assert.Equal(t, expectedsoftwarecomponentstring, softwarecomponentString)
		assert.Equal(t, nil, err)
	})
	t.Run("failure case: Test build body with example yaml config with only packages and no software components", func(t *testing.T) {
		expectedpackagestring := `<obj:packages><obj:package value="testPackage" includeSubpackages="true"/><obj:package value="testPackage2" includeSubpackages="false"/></obj:packages>`
		expectedsoftwarecomponentstring := ""

		var err error
		var config ATCconfig

		config = ATCconfig{
			ATCObjects{
				Package: []Package{
					Package{Name: "testPackage", IncludeSubpackages: true},
					Package{Name: "testPackage2", IncludeSubpackages: false},
				},
			},
		}

		var packageString, softwarecomponentString string

		packageString, softwarecomponentString, err = buildATCCheckBody(config)

		assert.Equal(t, expectedpackagestring, packageString)
		assert.Equal(t, expectedsoftwarecomponentstring, softwarecomponentString)
		assert.Equal(t, nil, err)

	})
	t.Run("success case: Test build body with example yaml config with no packages and only software components", func(t *testing.T) {
		expectedpackagestring := ""
		expectedsoftwarecomponentstring := `<obj:softwarecomponents><obj:softwarecomponent value="testSoftwareComponent"/><obj:softwarecomponent value="testSoftwareComponent2"/></obj:softwarecomponents>`

		var err error
		var config ATCconfig

		config = ATCconfig{
			ATCObjects{
				SoftwareComponent: []SoftwareComponent{
					SoftwareComponent{Name: "testSoftwareComponent"},
					SoftwareComponent{Name: "testSoftwareComponent2"},
				},
			},
		}

		var packageString, softwarecomponentString string

		packageString, softwarecomponentString, err = buildATCCheckBody(config)

		assert.Equal(t, expectedpackagestring, packageString)
		assert.Equal(t, expectedsoftwarecomponentstring, softwarecomponentString)
		assert.Equal(t, nil, err)
	})
}
