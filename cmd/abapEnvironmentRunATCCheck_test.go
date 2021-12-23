package cmd

import (
	"encoding/xml"
	"io/ioutil"
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestHostConfig(t *testing.T) {
	t.Run("Check Host: ABAP Endpoint", func(t *testing.T) {
		config := abaputils.AbapEnvironmentOptions{
			Username: "testUser",
			Password: "testPassword",
			Host:     "https://api.endpoint.com",
		}
		options := abaputils.AbapEnvironmentRunATCCheckOptions{
			AbapEnvOptions: config,
		}

		execRunner := &mock.ExecMockRunner{}
		var autils = abaputils.AbapUtils{
			Exec: execRunner,
		}
		var con abaputils.ConnectionDetailsHTTP
		con, error := autils.GetAbapCommunicationArrangementInfo(options.AbapEnvOptions, "")

		if error == nil {
			assert.Equal(t, "testUser", con.User)
			assert.Equal(t, "testPassword", con.Password)
			assert.Equal(t, "https://api.endpoint.com", con.URL)
			assert.Equal(t, "", con.XCsrfToken)
		}
	})
	t.Run("No host/ServiceKey configuration", func(t *testing.T) {
		//Testing without CfOrg parameter
		config := abaputils.AbapEnvironmentOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
		}
		options := abaputils.AbapEnvironmentRunATCCheckOptions{
			AbapEnvOptions: config,
		}

		execRunner := &mock.ExecMockRunner{}
		var autils = abaputils.AbapUtils{
			Exec: execRunner,
		}

		_, err := autils.GetAbapCommunicationArrangementInfo(options.AbapEnvOptions, "")
		assert.EqualError(t, err, "Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510")
		//Testing without ABAP Host
		config = abaputils.AbapEnvironmentOptions{
			Username: "testUser",
			Password: "testPassword",
		}
		_, err = autils.GetAbapCommunicationArrangementInfo(options.AbapEnvOptions, "")
		assert.EqualError(t, err, "Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510")
	})

	t.Run("Check Host: CF Service Key", func(t *testing.T) {
		config := abaputils.AbapEnvironmentOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfSpace:           "testSpace",
			CfOrg:             "Test",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
		}
		options := abaputils.AbapEnvironmentRunATCCheckOptions{
			AbapEnvOptions: config,
		}
		execRunner := &mock.ExecMockRunner{}
		var autils = abaputils.AbapUtils{
			Exec: execRunner,
		}
		var con abaputils.ConnectionDetailsHTTP
		con, error := autils.GetAbapCommunicationArrangementInfo(options.AbapEnvOptions, "")
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

		client := &abaputils.ClientMock{
			Body:  `ATC trigger test`,
			Token: tokenExpected,
		}

		con := abaputils.ConnectionDetailsHTTP{
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

		client := &abaputils.ClientMock{
			Body:  `Xcsrf Token test`,
			Token: tokenExpected,
		}

		con := abaputils.ConnectionDetailsHTTP{
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

		client := &abaputils.ClientMock{
			Body:  `Xcsrf Token test`,
			Token: "",
		}

		con := abaputils.ConnectionDetailsHTTP{
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

		client := &abaputils.ClientMock{
			Body:  `ATC Poll test`,
			Token: tokenExpected,
		}

		con := abaputils.ConnectionDetailsHTTP{
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
		client := &abaputils.ClientMock{
			Body: `HTTP response test`,
		}

		con := abaputils.ConnectionDetailsHTTP{
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
		client := &abaputils.ClientMock{
			BodyList: []string{
				`ATC result body`,
			},
		}

		con := abaputils.ConnectionDetailsHTTP{
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
		dir, err := ioutil.TempDir("", "test get result ATC run")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
			_ = os.RemoveAll(dir)
		}()
		bodyString := `<?xml version="1.0" encoding="UTF-8"?>
		<checkstyle>
			<file name="testFile">
				<error message="testMessage1" source="sourceTester" line="1" severity="error">
				</error>
				<error message="testMessage2" source="sourceTester" line="2" severity="info">
				</error>
			</file>
			<file name="testFile2">
			<error message="testMessage" source="sourceTester" line="1" severity="error">
				</error>
			</file>
		</checkstyle>`
		body := []byte(bodyString)
		err = parseATCResult(body, "ATCResults.xml", false)
		assert.Equal(t, nil, err)
	})
	t.Run("succes case: test parsing empty XML result", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "test get result ATC run")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
			_ = os.RemoveAll(dir)
		}()
		bodyString := `<?xml version="1.0" encoding="UTF-8"?>
		<checkstyle>
		</checkstyle>`
		body := []byte(bodyString)
		err = parseATCResult(body, "ATCResults.xml", false)
		assert.Equal(t, nil, err)
	})
	t.Run("failure case: parsing empty xml", func(t *testing.T) {
		var bodyString string
		body := []byte(bodyString)

		err := parseATCResult(body, "ATCResults.xml", false)
		assert.EqualError(t, err, "Parsing ATC result failed: Body is empty, can't parse empty body")
	})
	t.Run("failure case: html response", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "test get result ATC run")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
			_ = os.RemoveAll(dir)
		}()
		bodyString := `<html><head><title>HTMLTestResponse</title</head></html>`
		body := []byte(bodyString)
		err = parseATCResult(body, "ATCResults.xml", false)
		assert.EqualError(t, err, "The Software Component could not be checked. Please make sure the respective Software Component has been cloned successfully on the system")
	})
}

func TestBuildATCCheckBody(t *testing.T) {
	t.Run("Test build body with no software component and package", func(t *testing.T) {
		expectedObjectSet := ""

		var config ATCConfiguration

		objectSet, err := getATCObjectSet(config)

		assert.Equal(t, expectedObjectSet, objectSet)
		assert.EqualError(t, err, "Error while parsing ATC run config. Please provide the packages and/or the software components to be checked! No Package or Software Component specified. Please provide either one or both of them")
	})
	t.Run("success case: Test build body with example yaml config", func(t *testing.T) {

		expectedObjectSet := "<obj:objectSet><obj:softwarecomponents><obj:softwarecomponent value=\"testSoftwareComponent\"/><obj:softwarecomponent value=\"testSoftwareComponent2\"/></obj:softwarecomponents><obj:packages><obj:package value=\"testPackage\" includeSubpackages=\"true\"/><obj:package value=\"testPackage2\" includeSubpackages=\"false\"/></obj:packages></obj:objectSet>"

		config := ATCConfiguration{
			"",
			"",
			ATCObjects{
				Package: []Package{
					{Name: "testPackage", IncludeSubpackages: true},
					{Name: "testPackage2", IncludeSubpackages: false},
				},
				SoftwareComponent: []SoftwareComponent{
					{Name: "testSoftwareComponent"},
					{Name: "testSoftwareComponent2"},
				},
			},
		}

		objectSet, err := getATCObjectSet(config)

		assert.Equal(t, expectedObjectSet, objectSet)
		assert.Equal(t, nil, err)
	})
	t.Run("failure case: Test build body with example yaml config with only packages and no software components", func(t *testing.T) {

		expectedObjectSet := `<obj:objectSet><obj:packages><obj:package value="testPackage" includeSubpackages="true"/><obj:package value="testPackage2" includeSubpackages="false"/></obj:packages></obj:objectSet>`

		var err error
		var config ATCConfiguration

		config = ATCConfiguration{
			"",
			"",
			ATCObjects{
				Package: []Package{
					{Name: "testPackage", IncludeSubpackages: true},
					{Name: "testPackage2", IncludeSubpackages: false},
				},
			},
		}

		objectSet, err := getATCObjectSet(config)

		assert.Equal(t, expectedObjectSet, objectSet)
		assert.Equal(t, nil, err)

	})
	t.Run("success case: Test build body with example yaml config with no packages and only software components", func(t *testing.T) {

		expectedObjectSet := `<obj:objectSet><obj:softwarecomponents><obj:softwarecomponent value="testSoftwareComponent"/><obj:softwarecomponent value="testSoftwareComponent2"/></obj:softwarecomponents></obj:objectSet>`

		config := ATCConfiguration{
			"",
			"",
			ATCObjects{
				SoftwareComponent: []SoftwareComponent{
					{Name: "testSoftwareComponent"},
					{Name: "testSoftwareComponent2"},
				},
			},
		}

		objectSet, err := getATCObjectSet(config)

		assert.Equal(t, expectedObjectSet, objectSet)
		assert.Equal(t, nil, err)
	})
}

func TestGenerateHTMLDocument(t *testing.T) {
	//Failure case is not needed --> all failing cases would be depended on parsedXML *Result which is covered in TestParseATCResult
	t.Run("success case: html response", func(t *testing.T) {
		expectedResult := "<!DOCTYPE html><html lang=\"en\" xmlns=\"http://www.w3.org/1999/xhtml\"><head><title>ATC Results</title><meta http-equiv=\"Content-Type\" content=\"text/html; charset=UTF-8\" /><style>table,th,td {border: 1px solid black;border-collapse:collapse;}th,td{padding: 5px;text-align:left;font-size:medium;}</style></head><body><h1 style=\"text-align:left;font-size:large\">ATC Results</h1><table style=\"width:100%\"><tr><th>Severity</th><th>File</th><th>Message</th><th>Line</th><th>Checked by</th></tr><tr style=\"background-color: rgba(227,85,0)\"><td>error</td><td>testFile2</td><td>testMessage</td><td style=\"text-align:center\">1</td><td>sourceTester</td></tr><tr style=\"background-color: rgba(255,175,0, 0.75)\"><td>warning</td><td>testFile</td><td>testMessage2</td><td style=\"text-align:center\">2</td><td>sourceTester</td></tr><tr style=\"background-color: rgba(255,175,0, 0.2)\"><td>info</td><td>testFile</td><td>testMessage1</td><td style=\"text-align:center\">1</td><td>sourceTester</td></tr></table></body></html>"

		bodyString := `<?xml version="1.0" encoding="UTF-8"?>
		<checkstyle>
			<file name="testFile">
				<error message="testMessage1" source="sourceTester" line="1" severity="info">
				</error>
				<error message="testMessage2" source="sourceTester" line="2" severity="warning">
				</error>
			</file>
			<file name="testFile2">
			<error message="testMessage" source="sourceTester" line="1" severity="error">
				</error>
			</file>
		</checkstyle>`

		parsedXML := new(Result)
		err := xml.Unmarshal([]byte(bodyString), &parsedXML)
		if assert.NoError(t, err) {
			htmlDocumentResult := generateHTMLDocument(parsedXML)
			assert.Equal(t, expectedResult, htmlDocumentResult)
		}
	})
}

func TestResolveConfiguration(t *testing.T) {

	t.Run("resolve atcConfig-yml", func(t *testing.T) {

		expectedBodyString := "<?xml version=\"1.0\" encoding=\"UTF-8\"?><atc:runparameters xmlns:atc=\"http://www.sap.com/adt/atc\" xmlns:obj=\"http://www.sap.com/adt/objectset\" checkVariant=\"MY_TEST\" configuration=\"MY_CONFIG\"><obj:objectSet><obj:softwarecomponents><obj:softwarecomponent value=\"Z_TEST\"/><obj:softwarecomponent value=\"/DMO/SWC\"/></obj:softwarecomponents><obj:packages><obj:package value=\"Z_TEST\" includeSubpackages=\"false\"/></obj:packages></obj:objectSet></atc:runparameters>"
		config := abapEnvironmentRunATCCheckOptions{
			AtcConfig: "atc.yml",
		}

		dir, err := ioutil.TempDir("", "test parse AUnit yaml config2")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
			_ = os.RemoveAll(dir)
		}()

		yamlBody := `checkvariant: MY_TEST
configuration: MY_CONFIG
atcobjects:
  package:
    - name: Z_TEST
  softwarecomponent:
    - name: Z_TEST
    - name: /DMO/SWC
`

		err = ioutil.WriteFile(config.AtcConfig, []byte(yamlBody), 0644)
		if assert.Equal(t, err, nil) {
			bodyString, err := getATCRequestBody(config)
			assert.Equal(t, nil, err)
			assert.Equal(t, expectedBodyString, bodyString)
		}

	})

	t.Run("resolve repo-yml", func(t *testing.T) {

		expectedBodyString := "<?xml version=\"1.0\" encoding=\"UTF-8\"?><atc:runparameters xmlns:atc=\"http://www.sap.com/adt/atc\" xmlns:obj=\"http://www.sap.com/adt/objectset\" checkVariant=\"ABAP_CLOUD_DEVELOPMENT_DEFAULT\"><obj:objectSet><obj:softwarecomponents><obj:softwarecomponent value=\"Z_TEST\"/><obj:softwarecomponent value=\"/DMO/SWC\"/></obj:softwarecomponents></obj:objectSet></atc:runparameters>"
		config := abapEnvironmentRunATCCheckOptions{
			Repositories: "repo.yml",
		}

		dir, err := ioutil.TempDir("", "test parse AUnit yaml config2")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
			_ = os.RemoveAll(dir)
		}()

		yamlBody := `repositories:
  - name: Z_TEST
  - name: /DMO/SWC
`

		err = ioutil.WriteFile(config.Repositories, []byte(yamlBody), 0644)
		if assert.Equal(t, err, nil) {
			bodyString, err := getATCRequestBody(config)
			assert.Equal(t, nil, err)
			assert.Equal(t, expectedBodyString, bodyString)
		}

	})

	t.Run("Missing config files", func(t *testing.T) {

		config := abapEnvironmentRunATCCheckOptions{
			AtcConfig: "atc.yml",
		}

		bodyString, err := getATCRequestBody(config)
		assert.Equal(t, "Could not find atc.yml", err.Error())
		assert.Equal(t, "", bodyString)

	})
}
