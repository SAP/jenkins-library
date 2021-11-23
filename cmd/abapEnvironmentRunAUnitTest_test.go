package cmd

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type abapEnvironmentRunAUnitTestMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newAbapEnvironmentRunAUnitTestTestsUtils() abapEnvironmentRunAUnitTestMockUtils {
	utils := abapEnvironmentRunAUnitTestMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestBuildAUnitTestBody(t *testing.T) {
	t.Parallel()

	t.Run("Test AUnit test run body with no data", func(t *testing.T) {
		t.Parallel()

		expectedmetadataString := ""
		expectedoptionsString := ""
		expectedobjectSetString := ""

		var err error
		var config AUnitConfig
		var metadataString, optionsString, objectSetString string

		metadataString, optionsString, objectSetString, err = buildAUnitTestBody(config)

		assert.Equal(t, expectedmetadataString, metadataString)
		assert.Equal(t, expectedoptionsString, optionsString)
		assert.Equal(t, expectedobjectSetString, objectSetString)
		assert.EqualError(t, err, "Error while parsing AUnit test run config. No title for the AUnit run has been provided. Please configure an appropriate title for the respective test run")
	})

	t.Run("Test AUnit test run body with example yaml config", func(t *testing.T) {
		t.Parallel()

		expectedmetadataString := `<aunit:run title="Test Title" context="Test Context" xmlns:aunit="http://www.sap.com/adt/api/aunit">`
		expectedoptionsString := `<aunit:options><aunit:measurements type="none"/><aunit:scope ownTests="false" foreignTests="false"/><aunit:riskLevel harmless="false" dangerous="false" critical="false"/><aunit:duration short="false" medium="false" long="false"/></aunit:options>`
		expectedobjectSetString := `<osl:objectSet xsi:type="testSet" xmlns:osl="http://www.sap.com/api/osl" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"><osl:set xsi:type="testSet"><osl:package name="TestPackage" includeSubpackages="false"/><osl:object name="TestObject" type="CLAS"/></osl:set></osl:objectSet>`

		var err error
		var config AUnitConfig

		config = AUnitConfig{
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
			ObjectSet: []ObjectSet{{
				Type: "testSet",
				Set: []Set{{
					Type: "testSet",
					PackageSet: []AUnitPackage{{
						Name:               "TestPackage",
						IncludeSubpackages: new(bool),
					}},
					FlatObjectSet: []AUnitObject{{
						Name: "TestObject",
						Type: "CLAS",
					}},
				}},
			}},
		}

		var metadataString, optionsString, objectSetString string

		metadataString, optionsString, objectSetString, err = buildAUnitTestBody(config)

		assert.Equal(t, expectedmetadataString, metadataString)
		assert.Equal(t, expectedoptionsString, optionsString)
		assert.Equal(t, expectedobjectSetString, objectSetString)
		assert.Equal(t, nil, err)
		//assert.EqualError(t, err, "Error while parsing AUnit test run config. No title for the AUnit run has been provided. Please configure an appropriate title for the respective test run")
	})

	t.Run("Test AUnit test run body with example yaml config fail: no Title", func(t *testing.T) {
		t.Parallel()

		expectedmetadataString := ""
		expectedoptionsString := ""
		expectedobjectSetString := ""

		var err error
		var config AUnitConfig

		config = AUnitConfig{}

		var metadataString, optionsString, objectSetString string

		metadataString, optionsString, objectSetString, err = buildAUnitTestBody(config)

		assert.Equal(t, expectedmetadataString, metadataString)
		assert.Equal(t, expectedoptionsString, optionsString)
		assert.Equal(t, expectedobjectSetString, objectSetString)
		assert.EqualError(t, err, "Error while parsing AUnit test run config. No title for the AUnit run has been provided. Please configure an appropriate title for the respective test run")
	})

	t.Run("Test AUnit test run body with example yaml config fail: no Context", func(t *testing.T) {
		t.Parallel()

		expectedmetadataString := ""
		expectedoptionsString := ""
		expectedobjectSetString := ""

		var err error
		var config AUnitConfig

		config = AUnitConfig{Title: "Test"}

		var metadataString, optionsString, objectSetString string

		metadataString, optionsString, objectSetString, err = buildAUnitTestBody(config)

		assert.Equal(t, expectedmetadataString, metadataString)
		assert.Equal(t, expectedoptionsString, optionsString)
		assert.Equal(t, expectedobjectSetString, objectSetString)
		assert.EqualError(t, err, "Error while parsing AUnit test run config. No context for the AUnit run has been provided. Please configure an appropriate context for the respective test run")
	})

	t.Run("Test AUnit test run body with example yaml config fail: no Options", func(t *testing.T) {
		t.Parallel()

		expectedmetadataString := ""
		expectedoptionsString := ""
		expectedobjectSetString := ""

		var err error
		var config AUnitConfig

		config = AUnitConfig{Title: "Test", Context: "Test"}

		var metadataString, optionsString, objectSetString string

		metadataString, optionsString, objectSetString, err = buildAUnitTestBody(config)

		assert.Equal(t, expectedmetadataString, metadataString)
		assert.Equal(t, expectedoptionsString, optionsString)
		assert.Equal(t, expectedobjectSetString, objectSetString)
		assert.EqualError(t, err, "Error while parsing AUnit test run config. No options have been provided. Please configure the options for the respective test run")
	})

	t.Run("Test AUnit test run body with example yaml config fail: no ObjectSet", func(t *testing.T) {
		t.Parallel()

		expectedmetadataString := ""
		expectedoptionsString := ""
		expectedobjectSetString := ""

		var err error
		var config AUnitConfig

		config = AUnitConfig{Title: "Test", Context: "Test", Options: AUnitOptions{Measurements: "Test"}}

		var metadataString, optionsString, objectSetString string

		metadataString, optionsString, objectSetString, err = buildAUnitTestBody(config)

		assert.Equal(t, expectedmetadataString, metadataString)
		assert.Equal(t, expectedoptionsString, optionsString)
		assert.Equal(t, expectedobjectSetString, objectSetString)
		assert.EqualError(t, err, "Error while parsing AUnit test run object set config. No object set has been provided. Please configure the set of objects you want to be checked for the respective test run")
	})
}

func TestParseAUnitResult(t *testing.T) {
	t.Parallel()

	t.Run("succes case: test parsing example XML result", func(t *testing.T) {
		t.Parallel()

		dir, err := ioutil.TempDir("", "test get result AUnit test run")
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
		bodyString := `<?xml version="1.0" encoding="utf-8"?><testsuites title="My AUnit run" system="TST" client="100" executedBy="TESTUSER" time="000.000" timestamp="2021-01-01T00:00:00Z" failures="2" errors="2" skipped="0" asserts="0" tests="2"><testsuite name="" tests="2" failures="2" errors="0" skipped="0" asserts="0" package="testpackage" timestamp="2021-01-01T00:00:00ZZ" time="0.000" hostname="test"><testcase classname="test" name="execute" time="0.000" asserts="2"><failure message="testMessage1" type="Assert Failure">Test1</failure><failure message="testMessage2" type="Assert Failure">Test2</failure></testcase></testsuite></testsuites>`
		body := []byte(bodyString)
		err = parseAUnitResult(body, "AUnitResults.xml", false)
		assert.Equal(t, nil, err)
	})

	t.Run("succes case: test parsing empty AUnit run XML result", func(t *testing.T) {
		t.Parallel()

		dir, err := ioutil.TempDir("", "test get result AUnit test run")
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
		bodyString := `<?xml version="1.0" encoding="UTF-8"?>`
		body := []byte(bodyString)
		err = parseAUnitResult(body, "AUnitResults.xml", false)
		assert.Equal(t, nil, err)
	})

	t.Run("failure case: parsing empty xml", func(t *testing.T) {
		t.Parallel()

		var bodyString string
		body := []byte(bodyString)

		err := parseAUnitResult(body, "AUnitResults.xml", false)
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
		resp, err := getResultAUnitRun("GET", con, []byte(client.Body), client)
		defer resp.Body.Close()
		if assert.Equal(t, nil, err) {
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
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
		resp, err := getResultAUnitRun("GET", con, []byte(client.Body), client)
		defer resp.Body.Close()
		if assert.EqualError(t, err, "Getting AUnit run results failed: Test fail") {
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			newStr := buf.String()
			assert.Equal(t, "AUnit test result body", newStr)
			assert.Equal(t, int64(0), resp.ContentLength)
			assert.Equal(t, 400, resp.StatusCode)
			assert.Equal(t, []string([]string(nil)), resp.Header["X-Crsf-Token"])
		}
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
		token, error := fetchAUnitXcsrfToken("GET", con, []byte(client.Body), client)
		if assert.Equal(t, nil, error) {
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
		token, error := fetchAUnitXcsrfToken("GET", con, []byte(client.Body), client)
		if assert.Equal(t, nil, error) {
			assert.Equal(t, tokenExpected, token)
		}
	})

	t.Run("AUnit test run Poll Test", func(t *testing.T) {
		t.Parallel()

		tokenExpected := "myToken"

		client := &abaputils.ClientMock{
			Body:  `<?xml version="1.0" encoding="utf-8"?><aunit:run xmlns:aunit="http://www.sap.com/adt/api/aunit"><aunit:progress status="FINISHED"/><aunit:time/><atom:link href="/sap/bc/adt/api/abapunit/results/test" rel="http://www.sap.com/adt/relations/api/abapunit/run-result" type="application/vnd.sap.adt.api.junit.run-result.v1+xml" title="JUnit Run Result" xmlns:atom="http://www.w3.org/2005/Atom"/></aunit:run>`,
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
			Body:  `<?xml version="1.0" encoding="utf-8"?><aunit:run xmlns:aunit="http://www.sap.com/adt/api/aunit"><aunit:progress status="Not Created"/><aunit:time/><atom:link href="/sap/bc/adt/api/abapunit/results/test" rel="http://www.sap.com/adt/relations/api/abapunit/run-result" type="application/vnd.sap.adt.api.junit.run-result.v1+xml" title="JUnit Run Result" xmlns:atom="http://www.w3.org/2005/Atom"/></aunit:run>`,
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
		defer resp.Body.Close()
		if assert.Equal(t, nil, err) {
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			newStr := buf.String()
			assert.Equal(t, "HTTP response test", newStr)
			assert.Equal(t, int64(0), resp.ContentLength)
			assert.Equal(t, []string([]string(nil)), resp.Header["X-Crsf-Token"])
		}
	})
}

func TestGenerateHTMLDocumentAUnit(t *testing.T) {
	t.Parallel()

	t.Run("Test empty XML Result", func(t *testing.T) {
		t.Parallel()

		expectedString := `<!DOCTYPE html><html lang="en" xmlns="http://www.w3.org/1999/xhtml"><head><title>AUnit Results</title><meta http-equiv="Content-Type" content="text/html; charset=UTF-8" /><style>table,th,td {border-collapse:collapse;}th,td{padding: 5px;text-align:left;font-size:medium;}</style></head><body><table style="border: 1px solid black"><tr><th>Run title</th><td style="padding-right: 20px"></td><th>System</th><td style="padding-right: 20px"></td><th>Client</th><td style="padding-right: 20px"></td><th>ExecutedBy</th><td style="padding-right: 20px"></td><th>Duration</th><td style="padding-right: 20px">s</td><th>Timestamp</th><td style="padding-right: 20px"></td></tr><tr><th>Failures</th><td style="padding-right: 20px"></td><th>Errors</th><td style="padding-right: 20px"></td><th>Skipped</th><td style="padding-right: 20px"></td><th>Asserts</th><td style="padding-right: 20px"></td><th>Tests</th><td style="padding-right: 20px"></td></tr></table><br><table style="width:100%; border: 1px solid black""><tr style="border: 1px solid black"><th style="border: 1px solid black">Severity</th><th style="border: 1px solid black">File</th><th style="border: 1px solid black">Message</th><th style="border: 1px solid black">Type</th><th style="border: 1px solid black">Text</th></tr><tr><td colspan="5"><b>There are no AUnit findings to be displayed</b></td></tr></table></body></html>`

		result := AUnitResult{}

		resultString := generateHTMLDocumentAUnit(&result)

		assert.Equal(t, expectedString, resultString)
	})

	t.Run("Test AUnit XML Result", func(t *testing.T) {
		t.Parallel()

		expectedString := `<!DOCTYPE html><html lang="en" xmlns="http://www.w3.org/1999/xhtml"><head><title>AUnit Results</title><meta http-equiv="Content-Type" content="text/html; charset=UTF-8" /><style>table,th,td {border-collapse:collapse;}th,td{padding: 5px;text-align:left;font-size:medium;}</style></head><body><table style="border: 1px solid black"><tr><th>Run title</th><td style="padding-right: 20px"></td><th>System</th><td style="padding-right: 20px"></td><th>Client</th><td style="padding-right: 20px"></td><th>ExecutedBy</th><td style="padding-right: 20px"></td><th>Duration</th><td style="padding-right: 20px">s</td><th>Timestamp</th><td style="padding-right: 20px"></td></tr><tr><th>Failures</th><td style="padding-right: 20px"></td><th>Errors</th><td style="padding-right: 20px"></td><th>Skipped</th><td style="padding-right: 20px"></td><th>Asserts</th><td style="padding-right: 20px"></td><th>Tests</th><td style="padding-right: 20px"></td></tr></table><br><table style="width:100%; border: 1px solid black""><tr style="border: 1px solid black"><th style="border: 1px solid black">Severity</th><th style="border: 1px solid black">File</th><th style="border: 1px solid black">Message</th><th style="border: 1px solid black">Type</th><th style="border: 1px solid black">Text</th></tr><tr><td colspan="5"><b>There are no AUnit findings to be displayed</b></td></tr></table></body></html>`

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
			Testsuite: {
				Tests:     "6",
				Asserts:   "6",
				Skipped:   "2",
				Errors:    "2",
				Failures:  "2",
				Timestamp: "2021-00-00T00:00:00Z",
				Time:      "0.1",
				Hostname:  "0xb",
				Package:   "testPackage",
				Name:      "ZCL_testPackage",
				Testcase  []struct {
					Asserts:   "2",
					Time:      "2",
					Name:      "my_test",
					Classname: "ZCL_my_test",
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
				} "xml:\"testcase\""
			}{},
		}



		/*
				Tests:     "6",
				Asserts:   "6",
				Skipped:   "2",
				Errors:    "2",
				Failures:  "2",
				Timestamp: "2021-00-00T00:00:00Z",
				Time:      "0.1",
				Hostname:  "0xb",
				Package:   "testPackage",
				Name:      "ZCL_testPackage",
					Testcase: {
						Asserts:   "2",
						Time:      "2",
						Name:      "my_test",
						Classname: "ZCL_my_test",
						Error: {
							Text:    "testText",
							Type:    "Assert Error",
							Message: "Error in ZCL_my_test",
						}, {
							Text:    "testText",
							Type:    "Assert Error",
							Message: "Error in ZCL_my_test2",
						},
						Failure: {
							Text:    "testText",
							Type:    "Assert Failure",
							Message: "Error in ZCL_my_test",
						}, {
							Text:    "testText",
							Type:    "Assert Failure",
							Message: "Error in ZCL_my_test2",
						},
						Skipped: {
							Text:    "testText",
							Type:    "Skipped",
							Message: "Skipped ZCL_my_test",
						}, {
							Text:    "testText",
							Type:    "Skipped",
							Message: "Skipped ZCL_my_test2",
						},
					},
				},
			}

		*/

		resultString := generateHTMLDocumentAUnit(&result)

		assert.Equal(t, expectedString, resultString)
	})

}
