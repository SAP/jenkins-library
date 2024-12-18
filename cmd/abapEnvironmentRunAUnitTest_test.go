//go:build unit
// +build unit

package cmd

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestBuildAUnitRequestBody(t *testing.T) {
	t.Parallel()

	t.Run("Test AUnit test run body with no data", func(t *testing.T) {
		t.Parallel()

		var config abapEnvironmentRunAUnitTestOptions

		bodyString, err := buildAUnitRequestBody(config)
		expectedBodyString := ""

		assert.Equal(t, expectedBodyString, bodyString)
		assert.EqualError(t, err, "No configuration provided - please provide either an AUnit configuration file or a repository configuration file")
	})

	t.Run("Test AUnit test run body with example yaml config of not supported Object Sets", func(t *testing.T) {
		t.Parallel()

		expectedoptionsString := `<aunit:options><aunit:measurements type="none"/><aunit:scope ownTests="false" foreignTests="false"/><aunit:riskLevel harmless="false" dangerous="false" critical="false"/><aunit:duration short="false" medium="false" long="false"/></aunit:options>`
		expectedobjectSetString := ``

		config := AUnitConfig{
			Title:   "Test Title",
			Context: "Test Context",
			Options: AUnitOptions{
				Measurements: "none",
				Scope: Scope{
					OwnTests:     new(bool),
					ForeignTests: new(bool),
				},
				RiskLevel: RiskLevel{
					Harmless:  new(bool),
					Dangerous: new(bool),
					Critical:  new(bool),
				},
				Duration: Duration{
					Short:  new(bool),
					Medium: new(bool),
					Long:   new(bool),
				},
			},
			ObjectSet: abaputils.ObjectSet{
				Type: "testSet",
				Set: []abaputils.Set{
					{
						Type: "testSet",
						Set: []abaputils.Set{
							{
								Type: "testAUnitFlatObjectSet",
								FlatObjectSet: []abaputils.FlatObjectSet{
									{
										Name: "TestCLAS",
										Type: "CLAS",
									},
									{
										Name: "TestINTF",
										Type: "INTF",
									},
								},
							},
							{
								Type: "testAUnitObjectTypeSet",
								ObjectTypeSet: []abaputils.ObjectTypeSet{
									{
										Name: "TestObjectType",
									},
								},
							},
						},
					},
				},
			},
		}

		objectSetString := abaputils.BuildOSLString(config.ObjectSet)
		optionsString := buildAUnitOptionsString(config)

		assert.Equal(t, expectedoptionsString, optionsString)
		assert.Equal(t, expectedobjectSetString, objectSetString)
	})

	t.Run("Test AUnit test run body with example yaml config of only Multi Property Set", func(t *testing.T) {
		t.Parallel()

		expectedoptionsString := `<aunit:options><aunit:measurements type="none"/><aunit:scope ownTests="false" foreignTests="false"/><aunit:riskLevel harmless="false" dangerous="false" critical="false"/><aunit:duration short="false" medium="false" long="false"/></aunit:options>`
		expectedobjectSetString := `<osl:objectSet xsi:type="multiPropertySet" xmlns:osl="http://www.sap.com/api/osl" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"><osl:softwareComponent name="testComponent1"/><osl:softwareComponent name="testComponent2"/></osl:objectSet>`

		config := AUnitConfig{
			Title:   "Test Title",
			Context: "Test Context",
			Options: AUnitOptions{
				Measurements: "none",
				Scope: Scope{
					OwnTests:     new(bool),
					ForeignTests: new(bool),
				},
				RiskLevel: RiskLevel{
					Harmless:  new(bool),
					Dangerous: new(bool),
					Critical:  new(bool),
				},
				Duration: Duration{
					Short:  new(bool),
					Medium: new(bool),
					Long:   new(bool),
				},
			},
			ObjectSet: abaputils.ObjectSet{
				Type: "multiPropertySet",
				MultiPropertySet: abaputils.MultiPropertySet{
					SoftwareComponents: []abaputils.SoftwareComponents{
						{
							Name: "testComponent1",
						},
						{
							Name: "testComponent2",
						},
					},
				},
			},
		}

		objectSetString := abaputils.BuildOSLString(config.ObjectSet)
		optionsString := buildAUnitOptionsString(config)

		assert.Equal(t, expectedoptionsString, optionsString)
		assert.Equal(t, expectedobjectSetString, objectSetString)
	})

	t.Run("Test AUnit test run body with example yaml config of only Multi Property Set but empty type", func(t *testing.T) {
		t.Parallel()

		expectedoptionsString := `<aunit:options><aunit:measurements type="none"/><aunit:scope ownTests="false" foreignTests="false"/><aunit:riskLevel harmless="false" dangerous="false" critical="false"/><aunit:duration short="false" medium="false" long="false"/></aunit:options>`
		expectedobjectSetString := `<osl:objectSet xsi:type="multiPropertySet" xmlns:osl="http://www.sap.com/api/osl" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"><osl:softwareComponent name="testComponent1"/><osl:softwareComponent name="testComponent2"/></osl:objectSet>`

		config := AUnitConfig{
			Title:   "Test Title",
			Context: "Test Context",
			Options: AUnitOptions{
				Measurements: "none",
				Scope: Scope{
					OwnTests:     new(bool),
					ForeignTests: new(bool),
				},
				RiskLevel: RiskLevel{
					Harmless:  new(bool),
					Dangerous: new(bool),
					Critical:  new(bool),
				},
				Duration: Duration{
					Short:  new(bool),
					Medium: new(bool),
					Long:   new(bool),
				},
			},
			ObjectSet: abaputils.ObjectSet{
				Type: "",
				MultiPropertySet: abaputils.MultiPropertySet{
					SoftwareComponents: []abaputils.SoftwareComponents{
						{
							Name: "testComponent1",
						},
						{
							Name: "testComponent2",
						},
					},
				},
			},
		}

		objectSetString := abaputils.BuildOSLString(config.ObjectSet)
		optionsString := buildAUnitOptionsString(config)

		assert.Equal(t, expectedoptionsString, optionsString)
		assert.Equal(t, expectedobjectSetString, objectSetString)
	})

	t.Run("Test AUnit test run body with example yaml config of only Multi Property Set with scomps & packages on top level", func(t *testing.T) {
		t.Parallel()

		expectedoptionsString := `<aunit:options><aunit:measurements type="none"/><aunit:scope ownTests="false" foreignTests="false"/><aunit:riskLevel harmless="false" dangerous="false" critical="false"/><aunit:duration short="false" medium="false" long="false"/></aunit:options>`
		expectedobjectSetString := `<osl:objectSet xsi:type="multiPropertySet" xmlns:osl="http://www.sap.com/api/osl" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"><osl:package name="testPackage1"/><osl:package name="testPackage2"/><osl:softwareComponent name="testComponent1"/><osl:softwareComponent name="testComponent2"/></osl:objectSet>`

		config := AUnitConfig{
			Title:   "Test Title",
			Context: "Test Context",
			Options: AUnitOptions{
				Measurements: "none",
				Scope: Scope{
					OwnTests:     new(bool),
					ForeignTests: new(bool),
				},
				RiskLevel: RiskLevel{
					Harmless:  new(bool),
					Dangerous: new(bool),
					Critical:  new(bool),
				},
				Duration: Duration{
					Short:  new(bool),
					Medium: new(bool),
					Long:   new(bool),
				},
			},
			ObjectSet: abaputils.ObjectSet{
				PackageNames: []abaputils.Package{{
					Name: "testPackage1",
				}, {
					Name: "testPackage2",
				}},
				SoftwareComponents: []abaputils.SoftwareComponents{{
					Name: "testComponent1",
				}, {
					Name: "testComponent2",
				}},
			},
		}

		objectSetString := abaputils.BuildOSLString(config.ObjectSet)
		optionsString := buildAUnitOptionsString(config)

		assert.Equal(t, expectedoptionsString, optionsString)
		assert.Equal(t, expectedobjectSetString, objectSetString)
	})

	t.Run("Test AUnit test run body with example yaml config: no Options", func(t *testing.T) {
		t.Parallel()

		expectedoptionsString := `<aunit:options><aunit:measurements type="none"/><aunit:scope ownTests="true" foreignTests="true"/><aunit:riskLevel harmless="true" dangerous="true" critical="true"/><aunit:duration short="true" medium="true" long="true"/></aunit:options>`
		config := AUnitConfig{
			Title: "Test", Context: "Test",
			ObjectSet: abaputils.ObjectSet{
				PackageNames: []abaputils.Package{{
					Name: "testPackage1",
				}},
				SoftwareComponents: []abaputils.SoftwareComponents{{
					Name: "testComponent1",
				}},
			},
		}

		optionsString := buildAUnitOptionsString(config)
		assert.Equal(t, expectedoptionsString, optionsString)
	})

	t.Run("Config with repository-yml", func(t *testing.T) {
		config := abapEnvironmentRunAUnitTestOptions{
			AUnitResultsFileName: "aUnitResults.xml",
			Repositories:         "repositories.yml",
		}

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		repositories := `repositories:
  - name: /DMO/REPO
    branch: main
`
		expectedBodyString := "<?xml version=\"1.0\" encoding=\"UTF-8\"?><aunit:run title=\"AUnit Test Run\" context=\"ABAP Environment Pipeline\" xmlns:aunit=\"http://www.sap.com/adt/api/aunit\"><aunit:options><aunit:measurements type=\"none\"/><aunit:scope ownTests=\"true\" foreignTests=\"true\"/><aunit:riskLevel harmless=\"true\" dangerous=\"true\" critical=\"true\"/><aunit:duration short=\"true\" medium=\"true\" long=\"true\"/></aunit:options><osl:objectSet xsi:type=\"multiPropertySet\" xmlns:osl=\"http://www.sap.com/api/osl\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"><osl:softwareComponent name=\"/DMO/REPO\"/></osl:objectSet></aunit:run>"
		err := os.WriteFile(config.Repositories, []byte(repositories), 0o644)
		if assert.Equal(t, err, nil) {
			bodyString, err := buildAUnitRequestBody(config)
			assert.Equal(t, nil, err)
			assert.Equal(t, expectedBodyString, bodyString)
		}
	})

	t.Run("Config with aunitconfig-yml", func(t *testing.T) {
		config := abapEnvironmentRunAUnitTestOptions{
			AUnitResultsFileName: "aUnitResults.xml",
			AUnitConfig:          "aunit.yml",
		}

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		yamlBody := `title: My AUnit run
objectset:
  packages:
  - name: Z_TEST
  softwarecomponents:
  - name: Z_TEST
  - name: /DMO/SWC
`
		expectedBodyString := "<?xml version=\"1.0\" encoding=\"UTF-8\"?><aunit:run title=\"My AUnit run\" context=\"ABAP Environment Pipeline\" xmlns:aunit=\"http://www.sap.com/adt/api/aunit\"><aunit:options><aunit:measurements type=\"none\"/><aunit:scope ownTests=\"true\" foreignTests=\"true\"/><aunit:riskLevel harmless=\"true\" dangerous=\"true\" critical=\"true\"/><aunit:duration short=\"true\" medium=\"true\" long=\"true\"/></aunit:options><osl:objectSet xsi:type=\"multiPropertySet\" xmlns:osl=\"http://www.sap.com/api/osl\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"><osl:package name=\"Z_TEST\"/><osl:softwareComponent name=\"Z_TEST\"/><osl:softwareComponent name=\"/DMO/SWC\"/></osl:objectSet></aunit:run>"
		err := os.WriteFile(config.AUnitConfig, []byte(yamlBody), 0o644)
		if assert.Equal(t, err, nil) {
			bodyString, err := buildAUnitRequestBody(config)
			assert.Equal(t, nil, err)
			assert.Equal(t, expectedBodyString, bodyString)
		}
	})

	t.Run("Config with aunitconfig-yml mps", func(t *testing.T) {
		config := abapEnvironmentRunAUnitTestOptions{
			AUnitResultsFileName: "aUnitResults.xml",
			AUnitConfig:          "aunit.yml",
		}

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		yamlBody := `title: My AUnit run
objectset:
  type: multiPropertySet
  multipropertyset:
    packages:
      - name: Z_TEST
    softwarecomponents:
      - name: Z_TEST
      - name: /DMO/SWC
`
		expectedBodyString := "<?xml version=\"1.0\" encoding=\"UTF-8\"?><aunit:run title=\"My AUnit run\" context=\"ABAP Environment Pipeline\" xmlns:aunit=\"http://www.sap.com/adt/api/aunit\"><aunit:options><aunit:measurements type=\"none\"/><aunit:scope ownTests=\"true\" foreignTests=\"true\"/><aunit:riskLevel harmless=\"true\" dangerous=\"true\" critical=\"true\"/><aunit:duration short=\"true\" medium=\"true\" long=\"true\"/></aunit:options><osl:objectSet xsi:type=\"multiPropertySet\" xmlns:osl=\"http://www.sap.com/api/osl\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"><osl:package name=\"Z_TEST\"/><osl:softwareComponent name=\"Z_TEST\"/><osl:softwareComponent name=\"/DMO/SWC\"/></osl:objectSet></aunit:run>"
		err := os.WriteFile(config.AUnitConfig, []byte(yamlBody), 0o644)
		if assert.Equal(t, err, nil) {
			bodyString, err := buildAUnitRequestBody(config)
			assert.Equal(t, nil, err)
			assert.Equal(t, expectedBodyString, bodyString)
		}
	})

	t.Run("No AUnit config file - expect no panic", func(t *testing.T) {
		config := abapEnvironmentRunAUnitTestOptions{
			AUnitConfig: "aunit.yml",
		}

		_, err := buildAUnitRequestBody(config)
		assert.Equal(t, "Could not find aunit.yml", err.Error())
	})

	t.Run("No Repo config file - expect no panic", func(t *testing.T) {
		config := abapEnvironmentRunAUnitTestOptions{
			Repositories: "repo.yml",
		}

		_, err := buildAUnitRequestBody(config)
		assert.Equal(t, "Could not find repo.yml", err.Error())
	})
}

func TestTriggerAUnitrun(t *testing.T) {
	t.Run("succes case: test parsing example yaml config", func(t *testing.T) {
		config := abapEnvironmentRunAUnitTestOptions{
			AUnitConfig:          "aUnitConfig.yml",
			AUnitResultsFileName: "aUnitResults.xml",
		}

		client := &abaputils.ClientMock{
			Body:       `AUnit test result body`,
			StatusCode: 200,
		}

		con := abaputils.ConnectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		yamlBody := `title: My AUnit run
context: AIE integration tests
options:
  measurements: none
  scope:
    owntests: true
    foreigntests: true
  riskLevel:
    harmless: true
    dangerous: true
    critical: true
  duration:
    short: true
    medium: true
    long: true
objectset:
  packages:
  - name: Z_TEST
  softwarecomponents:
  - name: Z_TEST
`

		err := os.WriteFile(config.AUnitConfig, []byte(yamlBody), 0o644)
		if assert.Equal(t, err, nil) {
			_, err := triggerAUnitrun(config, con, client)
			assert.Equal(t, nil, err)
		}
	})

	t.Run("succes case: test parsing example yaml config", func(t *testing.T) {
		config := abapEnvironmentRunAUnitTestOptions{
			AUnitConfig:          "aUnitConfig.yml",
			AUnitResultsFileName: "aUnitResults.xml",
		}

		client := &abaputils.ClientMock{
			Body: `AUnit test result body`,
		}

		con := abaputils.ConnectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		yamlBody := `title: My AUnit run
context: AIE integration tests
options:
  measurements: none
  scope:
    owntests: true
    foreigntests: true
  riskLevel:
    harmless: true
    dangerous: true
    critical: true
  duration:
    short: true
    medium: true
    long: true
objectset:
  type: unionSet
  set:
    - type: componentSet
      component:
      - name: Z_TEST_SC
`

		err := os.WriteFile(config.AUnitConfig, []byte(yamlBody), 0o644)
		if assert.Equal(t, err, nil) {
			_, err := triggerAUnitrun(config, con, client)
			assert.Equal(t, nil, err)
		}
	})
}

func TestParseAUnitResult(t *testing.T) {
	t.Parallel()

	t.Run("succes case: test parsing example XML result", func(t *testing.T) {
		bodyString := `<?xml version="1.0" encoding="utf-8"?><testsuites title="My AUnit run" system="TST" client="100" executedBy="TESTUSER" time="000.000" timestamp="2021-01-01T00:00:00Z" failures="2" errors="2" skipped="0" asserts="0" tests="2"><testsuite name="" tests="2" failures="2" errors="0" skipped="0" asserts="0" package="testpackage" timestamp="2021-01-01T00:00:00ZZ" time="0.000" hostname="test"><testcase classname="test" name="execute" time="0.000" asserts="2"><failure message="testMessage1" type="Assert Failure">Test1</failure><failure message="testMessage2" type="Assert Failure">Test2</failure></testcase></testsuite></testsuites>`
		body := []byte(bodyString)
		err := persistAUnitResult(&mock.FilesMock{}, body, "AUnitResults.xml", false)
		assert.Equal(t, nil, err)
	})

	t.Run("succes case: test parsing empty AUnit run XML result", func(t *testing.T) {
		bodyString := `<?xml version="1.0" encoding="UTF-8"?>`
		body := []byte(bodyString)
		err := persistAUnitResult(&mock.FilesMock{}, body, "AUnitResults.xml", false)
		assert.Equal(t, nil, err)
	})

	t.Run("failure case: parsing empty xml", func(t *testing.T) {
		var bodyString string
		body := []byte(bodyString)
		err := persistAUnitResult(&mock.FilesMock{}, body, "AUnitResults.xml", false)
		assert.EqualError(t, err, "Parsing AUnit result failed: Body is empty, can't parse empty body")
	})
}

func TestGetResultAUnitRun(t *testing.T) {
	t.Parallel()

	t.Run("Get HTTP Response from AUnit test run Test", func(t *testing.T) {
		t.Parallel()

		client := &abaputils.ClientMock{
			Body: `AUnit test result body`,
		}

		con := abaputils.ConnectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}
		resp, err := getAUnitResults("GET", con, []byte(client.Body), client)
		assert.NoError(t, err)
		defer resp.Body.Close()
		if assert.Equal(t, nil, err) {
			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(resp.Body)
			assert.NoError(t, err)
			newStr := buf.String()
			assert.Equal(t, "AUnit test result body", newStr)
			assert.Equal(t, int64(0), resp.ContentLength)
			assert.Equal(t, []string([]string(nil)), resp.Header["X-Crsf-Token"])
		}
	})

	t.Run("Get HTTP Response from AUnit test run Test Failure", func(t *testing.T) {
		t.Parallel()

		client := &abaputils.ClientMock{
			Body:       `AUnit test result body`,
			BodyList:   []string{},
			StatusCode: 400,
			Error:      fmt.Errorf("%w", errors.New("Test fail")),
		}

		con := abaputils.ConnectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}
		resp, err := getAUnitResults("GET", con, []byte(client.Body), client)
		assert.EqualError(t, err, "Getting AUnit run results failed: Test fail")
		defer resp.Body.Close()

		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(resp.Body)
		assert.NoError(t, err)
		newStr := buf.String()
		assert.Equal(t, "AUnit test result body", newStr)
		assert.Equal(t, int64(0), resp.ContentLength)
		assert.Equal(t, 400, resp.StatusCode)
		assert.Equal(t, []string([]string(nil)), resp.Header["X-Crsf-Token"])
	})
}

func TestRunAbapEnvironmentRunAUnitTest(t *testing.T) {
	t.Parallel()

	t.Run("FetchXcsrfToken Test", func(t *testing.T) {
		t.Parallel()

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
		token, err := fetchAUnitXcsrfToken("GET", con, []byte(client.Body), client)
		if assert.Equal(t, nil, err) {
			assert.Equal(t, tokenExpected, token)
		}
	})

	t.Run("failure case: fetch token", func(t *testing.T) {
		t.Parallel()

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
		token, err := fetchAUnitXcsrfToken("GET", con, []byte(client.Body), client)
		if assert.Equal(t, nil, err) {
			assert.Equal(t, tokenExpected, token)
		}
	})

	t.Run("AUnit test run Poll Test", func(t *testing.T) {
		t.Parallel()

		tokenExpected := "myToken"

		client := &abaputils.ClientMock{
			Body:  `<?xml version="1.0" encoding="utf-8"?><aunit:run xmlns:aunit="http://www.sap.com/adt/api/aunit"><aunit:progress status="FINISHED"/><aunit:time/><atom:link href="/sap/bc/adt/api/abapunit/results/test" rel="http://www.sap.com/adt/relations/api/abapunit/run-result" type="application/vnd.sap.adt.api.junit.run-result.v1xml" title="JUnit Run Result" xmlns:atom="http://www.w3.org/2005/Atom"/></aunit:run>`,
			Token: tokenExpected,
		}

		con := abaputils.ConnectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}
		resp, err := pollAUnitRun(con, []byte(client.Body), client)
		if assert.Equal(t, nil, err) {
			assert.Equal(t, "/sap/bc/adt/api/abapunit/results/test", resp)
		}
	})

	t.Run("AUnit test run Poll Test Fail", func(t *testing.T) {
		t.Parallel()

		tokenExpected := "myToken"

		client := &abaputils.ClientMock{
			Body:  `<?xml version="1.0" encoding="utf-8"?><aunit:run xmlns:aunit="http://www.sap.com/adt/api/aunit"><aunit:progress status="Not Created"/><aunit:time/><atom:link href="/sap/bc/adt/api/abapunit/results/test" rel="http://www.sap.com/adt/relations/api/abapunit/run-result" type="application/vnd.sap.adt.api.junit.run-result.v1xml" title="JUnit Run Result" xmlns:atom="http://www.w3.org/2005/Atom"/></aunit:run>`,
			Token: tokenExpected,
		}

		con := abaputils.ConnectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}
		resp, err := pollAUnitRun(con, []byte(client.Body), client)
		if assert.Equal(t, nil, err) {
			assert.Equal(t, "", resp)
		}
	})

	t.Run("Get HTTP Response from AUnit test run Test", func(t *testing.T) {
		t.Parallel()

		client := &abaputils.ClientMock{
			Body: `HTTP response test`,
		}

		con := abaputils.ConnectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}
		fmt.Println("Body:" + string([]byte(client.Body)))
		resp, err := getHTTPResponseAUnitRun("GET", con, []byte(client.Body), client)
		assert.NoError(t, err)
		defer resp.Body.Close()
		if assert.Equal(t, nil, err) {
			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(resp.Body)
			assert.NoError(t, err)
			newStr := buf.String()
			assert.Equal(t, "HTTP response test", newStr)
			assert.Equal(t, int64(0), resp.ContentLength)
			assert.Equal(t, []string([]string(nil)), resp.Header["X-Crsf-Token"])
		}
	})
}

func TestGenerateHTMLDocumentAUnit(t *testing.T) {
	t.Run("Test empty XML Result", func(t *testing.T) {
		expectedString := `<!DOCTYPE html><html lang="en" xmlns="http://www.w3.org/1999/xhtml"><head><title>AUnit Results</title><meta http-equiv="Content-Type" content="text/html; charset=UTF-8" /><style>table,th,td {border-collapse:collapse;}th,td{padding: 5px;text-align:left;font-size:medium;}</style></head><body><h1 style="text-align:left;font-size:large">AUnit Results</h1><table><tr><th>Run title</th><td style="padding-right: 20px"></td><th>System</th><td style="padding-right: 20px"></td><th>Client</th><td style="padding-right: 20px"></td><th>ExecutedBy</th><td style="padding-right: 20px"></td><th>Duration</th><td style="padding-right: 20px">s</td><th>Timestamp</th><td style="padding-right: 20px"></td></tr><tr><th>Failures</th><td style="padding-right: 20px"></td><th>Errors</th><td style="padding-right: 20px"></td><th>Skipped</th><td style="padding-right: 20px"></td><th>Asserts</th><td style="padding-right: 20px"></td><th>Tests</th><td style="padding-right: 20px"></td></tr></table><br><table style="width:100%; border: 1px solid black""><tr style="border: 1px solid black"><th style="border: 1px solid black">Severity</th><th style="border: 1px solid black">File</th><th style="border: 1px solid black">Message</th><th style="border: 1px solid black">Type</th><th style="border: 1px solid black">Text</th></tr><tr><td colspan="5"><b>There are no AUnit findings to be displayed</b></td></tr></table></body></html>`

		result := AUnitResult{}

		resultString := generateHTMLDocumentAUnit(&result)

		assert.Equal(t, expectedString, resultString)
	})

	t.Run("Test AUnit XML Result", func(t *testing.T) {
		expectedString := `<!DOCTYPE html><html lang="en" xmlns="http://www.w3.org/1999/xhtml"><head><title>AUnit Results</title><meta http-equiv="Content-Type" content="text/html; charset=UTF-8" /><style>table,th,td {border-collapse:collapse;}th,td{padding: 5px;text-align:left;font-size:medium;}</style></head><body><h1 style="text-align:left;font-size:large">AUnit Results</h1><table><tr><th>Run title</th><td style="padding-right: 20px">Test title</td><th>System</th><td style="padding-right: 20px">Test system</td><th>Client</th><td style="padding-right: 20px">000</td><th>ExecutedBy</th><td style="padding-right: 20px">CC00000</td><th>Duration</th><td style="padding-right: 20px">0.15s</td><th>Timestamp</th><td style="padding-right: 20px">2021-00-00T00:00:00Z</td></tr><tr><th>Failures</th><td style="padding-right: 20px">4</td><th>Errors</th><td style="padding-right: 20px">4</td><th>Skipped</th><td style="padding-right: 20px">4</td><th>Asserts</th><td style="padding-right: 20px">12</td><th>Tests</th><td style="padding-right: 20px">12</td></tr></table><br><table style="width:100%; border: 1px solid black""><tr style="border: 1px solid black"><th style="border: 1px solid black">Severity</th><th style="border: 1px solid black">File</th><th style="border: 1px solid black">Message</th><th style="border: 1px solid black">Type</th><th style="border: 1px solid black">Text</th></tr><tr style="background-color: grey"><td colspan="5"><b>Testcase: my_test for class ZCL_my_test</b></td></tr><tr style="background-color: rgba(227,85,0)"><td style="border: 1px solid black">Failure</td><td style="border: 1px solid black">ZCL_my_test</td><td style="border: 1px solid black">testMessage</td><td style="border: 1px solid black">Assert Error</td><td style="border: 1px solid black">testError</td></tr><tr style="background-color: rgba(227,85,0)"><td style="border: 1px solid black">Failure</td><td style="border: 1px solid black">ZCL_my_test</td><td style="border: 1px solid black">testMessage2</td><td style="border: 1px solid black">Assert Error2</td><td style="border: 1px solid black">testError2</td></tr><tr style="background-color: rgba(227,85,0)"><td style="border: 1px solid black">Failure</td><td style="border: 1px solid black">ZCL_my_test</td><td style="border: 1px solid black">testMessage</td><td style="border: 1px solid black">Assert Failure</td><td style="border: 1px solid black">testFailure</td></tr><tr style="background-color: rgba(227,85,0)"><td style="border: 1px solid black">Failure</td><td style="border: 1px solid black">ZCL_my_test</td><td style="border: 1px solid black">testMessage2</td><td style="border: 1px solid black">Assert Failure2</td><td style="border: 1px solid black">testFailure2</td></tr><tr style="background-color: rgba(255,175,0, 0.2)"><td style="border: 1px solid black">Failure</td><td style="border: 1px solid black">ZCL_my_test</td><td style="border: 1px solid black">testSkipped</td><td style="border: 1px solid black">-</td><td style="border: 1px solid black">testSkipped</td></tr><tr style="background-color: rgba(255,175,0, 0.2)"><td style="border: 1px solid black">Failure</td><td style="border: 1px solid black">ZCL_my_test</td><td style="border: 1px solid black">testSkipped2</td><td style="border: 1px solid black">-</td><td style="border: 1px solid black">testSkipped2</td></tr><tr style="background-color: grey"><td colspan="5"><b>Testcase: my_test2 for class ZCL_my_test2</b></td></tr><tr style="background-color: rgba(227,85,0)"><td style="border: 1px solid black">Failure</td><td style="border: 1px solid black">ZCL_my_test2</td><td style="border: 1px solid black">testMessage3</td><td style="border: 1px solid black">Assert Error3</td><td style="border: 1px solid black">testError3</td></tr><tr style="background-color: rgba(227,85,0)"><td style="border: 1px solid black">Failure</td><td style="border: 1px solid black">ZCL_my_test2</td><td style="border: 1px solid black">testMessage4</td><td style="border: 1px solid black">Assert Error4</td><td style="border: 1px solid black">testError4</td></tr><tr style="background-color: rgba(227,85,0)"><td style="border: 1px solid black">Failure</td><td style="border: 1px solid black">ZCL_my_test2</td><td style="border: 1px solid black">testMessage5</td><td style="border: 1px solid black">Assert Failure5</td><td style="border: 1px solid black">testFailure5</td></tr><tr style="background-color: rgba(227,85,0)"><td style="border: 1px solid black">Failure</td><td style="border: 1px solid black">ZCL_my_test2</td><td style="border: 1px solid black">testMessage6</td><td style="border: 1px solid black">Assert Failure6</td><td style="border: 1px solid black">testFailure6</td></tr><tr style="background-color: rgba(255,175,0, 0.2)"><td style="border: 1px solid black">Failure</td><td style="border: 1px solid black">ZCL_my_test2</td><td style="border: 1px solid black">testSkipped7</td><td style="border: 1px solid black">-</td><td style="border: 1px solid black">testSkipped7</td></tr><tr style="background-color: rgba(255,175,0, 0.2)"><td style="border: 1px solid black">Failure</td><td style="border: 1px solid black">ZCL_my_test2</td><td style="border: 1px solid black">testSkipped8</td><td style="border: 1px solid black">-</td><td style="border: 1px solid black">testSkipped8</td></tr></table></body></html>`

		result := AUnitResult{
			XMLName:    xml.Name{Space: "testSpace", Local: "testLocal"},
			Title:      "Test title",
			System:     "Test system",
			Client:     "000",
			ExecutedBy: "CC00000",
			Time:       "0.15",
			Timestamp:  "2021-00-00T00:00:00Z",
			Failures:   "4",
			Errors:     "4",
			Skipped:    "4",
			Asserts:    "12",
			Tests:      "12",
			Testsuite: struct {
				Tests     string `xml:"tests,attr"`
				Asserts   string `xml:"asserts,attr"`
				Skipped   string `xml:"skipped,attr"`
				Errors    string `xml:"errors,attr"`
				Failures  string `xml:"failures,attr"`
				Timestamp string `xml:"timestamp,attr"`
				Time      string `xml:"time,attr"`
				Hostname  string `xml:"hostname,attr"`
				Package   string `xml:"package,attr"`
				Name      string `xml:"name,attr"`
				Testcase  []struct {
					Asserts   string `xml:"asserts,attr"`
					Time      string `xml:"time,attr"`
					Name      string `xml:"name,attr"`
					Classname string `xml:"classname,attr"`
					Error     []struct {
						Text    string `xml:",chardata"`
						Type    string `xml:"type,attr"`
						Message string `xml:"message,attr"`
					} `xml:"error"`
					Failure []struct {
						Text    string `xml:",chardata"`
						Type    string `xml:"type,attr"`
						Message string `xml:"message,attr"`
					} `xml:"failure"`
					Skipped []struct {
						Text    string `xml:",chardata"`
						Message string `xml:"message,attr"`
					} `xml:"skipped"`
				} `xml:"testcase"`
			}{
				Tests:     "6",
				Asserts:   "4",
				Skipped:   "2",
				Errors:    "2",
				Failures:  "2",
				Timestamp: "2021-00-00T00:00:00Z",
				Time:      "0.15",
				Hostname:  "0xb",
				Package:   "testPackage",
				Name:      "ZCL_testPackage",
				Testcase: []struct {
					Asserts   string "xml:\"asserts,attr\""
					Time      string "xml:\"time,attr\""
					Name      string "xml:\"name,attr\""
					Classname string "xml:\"classname,attr\""
					Error     []struct {
						Text    string "xml:\",chardata\""
						Type    string "xml:\"type,attr\""
						Message string "xml:\"message,attr\""
					} "xml:\"error\""
					Failure []struct {
						Text    string "xml:\",chardata\""
						Type    string "xml:\"type,attr\""
						Message string "xml:\"message,attr\""
					} "xml:\"failure\""
					Skipped []struct {
						Text    string "xml:\",chardata\""
						Message string "xml:\"message,attr\""
					} "xml:\"skipped\""
				}{{
					Asserts:   "4",
					Time:      "0.15",
					Name:      "my_test",
					Classname: "ZCL_my_test",
					Error: []struct {
						Text    string "xml:\",chardata\""
						Type    string "xml:\"type,attr\""
						Message string "xml:\"message,attr\""
					}{{
						Text:    "testError",
						Type:    "Assert Error",
						Message: "testMessage",
					}, {
						Text:    "testError2",
						Type:    "Assert Error2",
						Message: "testMessage2",
					}},
					Failure: []struct {
						Text    string "xml:\",chardata\""
						Type    string "xml:\"type,attr\""
						Message string "xml:\"message,attr\""
					}{{
						Text:    "testFailure",
						Type:    "Assert Failure",
						Message: "testMessage",
					}, {
						Text:    "testFailure2",
						Type:    "Assert Failure2",
						Message: "testMessage2",
					}},
					Skipped: []struct {
						Text    string "xml:\",chardata\""
						Message string "xml:\"message,attr\""
					}{{
						Text:    "testSkipped",
						Message: "testSkipped",
					}, {
						Text:    "testSkipped2",
						Message: "testSkipped2",
					}},
				}, {
					Asserts:   "4",
					Time:      "0.15",
					Name:      "my_test2",
					Classname: "ZCL_my_test2",
					Error: []struct {
						Text    string "xml:\",chardata\""
						Type    string "xml:\"type,attr\""
						Message string "xml:\"message,attr\""
					}{{
						Text:    "testError3",
						Type:    "Assert Error3",
						Message: "testMessage3",
					}, {
						Text:    "testError4",
						Type:    "Assert Error4",
						Message: "testMessage4",
					}},
					Failure: []struct {
						Text    string "xml:\",chardata\""
						Type    string "xml:\"type,attr\""
						Message string "xml:\"message,attr\""
					}{{
						Text:    "testFailure5",
						Type:    "Assert Failure5",
						Message: "testMessage5",
					}, {
						Text:    "testFailure6",
						Type:    "Assert Failure6",
						Message: "testMessage6",
					}},
					Skipped: []struct {
						Text    string "xml:\",chardata\""
						Message string "xml:\"message,attr\""
					}{{
						Text:    "testSkipped7",
						Message: "testSkipped7",
					}, {
						Text:    "testSkipped8",
						Message: "testSkipped8",
					}},
				}},
			},
		}

		resultString := generateHTMLDocumentAUnit(&result)
		fmt.Println(resultString)
		assert.Equal(t, expectedString, resultString)
	})
}
