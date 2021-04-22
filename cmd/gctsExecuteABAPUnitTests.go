package cmd

import (
	"bytes"
	"encoding/xml"
	"net/http"
	"net/http/cookiejar"
	"fmt"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func gctsExecuteABAPUnitTests(config gctsExecuteABAPUnitTestsOptions, telemetryData *telemetry.CustomData) {

	// for http calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go
	httpClient := &piperhttp.Client{}

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runUnitTestsForAllRepoPackages(&config, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runUnitTestsForAllRepoPackages(config *gctsExecuteABAPUnitTestsOptions, httpClient piperhttp.Sender) error {

	cookieJar, cookieErr := cookiejar.New(nil)
	if cookieErr != nil {
		return errors.Wrap(cookieErr, "execution of unit tests failed")
	}
	clientOptions := piperhttp.ClientOptions{
		CookieJar: cookieJar,
		Username:  config.Username,
		Password:  config.Password,
	}
	httpClient.SetOptions(clientOptions)

	repoObjects, getPackageErr := getPackageList(config, httpClient)

	if getPackageErr != nil {
		return errors.Wrap(getPackageErr, "execution of unit tests failed")
	}

	discHeader, discError := discoverServer(config, httpClient)

	if discError != nil {
		return errors.Wrap(discError, "execution of unit tests failed")
	}

	if discHeader.Get("X-Csrf-Token") == "" {
		return errors.Errorf("could not retrieve x-csrf-token from server")
	}

	header := make(http.Header)
	header.Add("x-csrf-token", discHeader.Get("X-Csrf-Token"))
	header.Add("Accept", "application/xml")
	header.Add("Content-Type", "application/vnd.sap.adt.abapunit.testruns.result.v1+xml")

	for _, object := range repoObjects {
		executeTestsErr := executeTestsForPackage(config, httpClient, header, object)

		if executeTestsErr != nil {
			return errors.Wrap(executeTestsErr, "execution of unit tests failed")
		}
	}

	log.Entry().
		WithField("repository", config.Repository).
		Info("all unit tests were successful")
	fmt.Printf("%v", config.CommitID)
	return nil
}

func discoverServer(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender) (*http.Header, error) {

	url := config.Host +
		"/sap/bc/adt/core/discovery?sap-client=" + config.Client

	header := make(http.Header)
	header.Add("Accept", "application/atomsvc+xml")
	header.Add("x-csrf-token", "fetch")
	header.Add("saml2", "disabled")

	disc, httpErr := client.SendRequest("GET", url, nil, header, nil)

	defer func() {
		if disc != nil && disc.Body != nil {
			disc.Body.Close()
		}
	}()

	if httpErr != nil {
		return nil, errors.Wrap(httpErr, "discovery of the ABAP server failed")
	} else if disc == nil || disc.Header == nil {
		return nil, errors.New("discovery of the ABAP server failed: did not retrieve a HTTP response")
	}

	return &disc.Header, nil
}

func executeTestsForPackage(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender, header http.Header, packageName string) error {

	var xmlBody = []byte(`<?xml version="1.0" encoding="UTF-8"?>
	<aunit:runConfiguration
			xmlns:aunit="http://www.sap.com/adt/aunit">
			<external>
					<coverage active="false"/>
			</external>
			<options>
					<uriType value="semantic"/>
					<testDeterminationStrategy sameProgram="true" assignedTests="false" appendAssignedTestsPreview="true"/>
					<testRiskLevels harmless="true" dangerous="true" critical="true"/>
					<testDurations short="true" medium="true" long="true"/>
			</options>
			<adtcore:objectSets
					xmlns:adtcore="http://www.sap.com/adt/core">
					<objectSet kind="inclusive">
							<adtcore:objectReferences>
									<adtcore:objectReference adtcore:uri="/sap/bc/adt/packages/` + packageName + `"/>
							</adtcore:objectReferences>
					</objectSet>
			</adtcore:objectSets>
	</aunit:runConfiguration>`)

	url := config.Host +
		"/sap/bc/adt/abapunit/testruns?sap-client=" + config.Client

	resp, httpErr := client.SendRequest("POST", url, bytes.NewBuffer(xmlBody), header, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return errors.Wrap(httpErr, "execution of unit tests failed")
	} else if resp == nil {
		return errors.New("execution of unit tests failed: did not retrieve a HTTP response")
	}

	var response runResult
	parsingErr := piperhttp.ParseHTTPResponseBodyXML(resp, &response)
	if parsingErr != nil {
		log.Entry().Warning(parsingErr)
	}

	aunitError := parseAUnitResponse(&response)
	if aunitError != nil {
		return aunitError
	}

	return nil
}

func parseAUnitResponse(response *runResult) error {
	var node string
	aunitError := false

	for _, program := range response.Program {
		log.Entry().Infof("testing class %v", program.Name)
		for _, testClass := range program.TestClasses.TestClass {
			log.Entry().Infof("using test class %v", testClass.Name)
			for _, testMethod := range testClass.TestMethods.TestMethod {
				node = testMethod.Name
				if len(testMethod.Alerts.Alert) > 0 {
					log.Entry().Errorf("%v - error", node)
					aunitError = true
				} else {
					log.Entry().Infof("%v - ok", node)
				}
			}
		}
	}
	if aunitError {
		return errors.Errorf("some unit tests failed")
	}
	return nil
}

func getPackageList(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender) ([]string, error) {

	type object struct {
		Pgmid       string `json:"pgmid"`
		Object      string `json:"object"`
		Type        string `json:"type"`
		Description string `json:"description"`
	}

	type objectsResponseBody struct {
		Objects   []object      `json:"objects"`
		Log       []gctsLogs    `json:"log"`
		Exception gctsException `json:"exception"`
		ErrorLogs []gctsLogs    `json:"errorLog"`
	}

	url := config.Host +
		"/sap/bc/cts_abapvcs/repository/" + config.Repository +
		"/getObjects?sap-client=" + config.Client

	resp, httpErr := client.SendRequest("GET", url, nil, nil, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return []string{}, errors.Wrap(httpErr, "getting repository object/package list failed")
	} else if resp == nil {
		return []string{}, errors.New("getting repository object/package list failed: did not retrieve a HTTP response")
	}

	var response objectsResponseBody
	parsingErr := piperhttp.ParseHTTPResponseBodyJSON(resp, &response)
	if parsingErr != nil {
		return []string{}, errors.Errorf("%v", parsingErr)
	}

	repoObjects := []string{}
	for _, object := range response.Objects {
		if object.Type == "DEVC" {
			repoObjects = append(repoObjects, object.Object)
		}
	}

	return repoObjects, nil
}

type gctsException struct {
	Message     string `json:"message"`
	Description string `json:"description"`
	Code        int    `json:"code"`
}

type gctsLogs struct {
	Time     int    `json:"time"`
	User     string `json:"user"`
	Section  string `json:"section"`
	Action   string `json:"action"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Code     string `json:"code"`
}

type runResult struct {
	XMLName xml.Name `xml:"runResult"`
	Text    string   `xml:",chardata"`
	Aunit   string   `xml:"aunit,attr"`
	Program []struct {
		Text        string `xml:",chardata"`
		URI         string `xml:"uri,attr"`
		Type        string `xml:"type,attr"`
		Name        string `xml:"name,attr"`
		URIType     string `xml:"uriType,attr"`
		Adtcore     string `xml:"adtcore,attr"`
		TestClasses struct {
			Text      string `xml:",chardata"`
			TestClass []struct {
				Text             string `xml:",chardata"`
				URI              string `xml:"uri,attr"`
				Type             string `xml:"type,attr"`
				Name             string `xml:"name,attr"`
				URIType          string `xml:"uriType,attr"`
				NavigationURI    string `xml:"navigationUri,attr"`
				DurationCategory string `xml:"durationCategory,attr"`
				RiskLevel        string `xml:"riskLevel,attr"`
				TestMethods      struct {
					Text       string `xml:",chardata"`
					TestMethod []struct {
						Text          string `xml:",chardata"`
						URI           string `xml:"uri,attr"`
						Type          string `xml:"type,attr"`
						Name          string `xml:"name,attr"`
						ExecutionTime string `xml:"executionTime,attr"`
						URIType       string `xml:"uriType,attr"`
						NavigationURI string `xml:"navigationUri,attr"`
						Unit          string `xml:"unit,attr"`
						Alerts        struct {
							Text  string `xml:",chardata"`
							Alert []struct {
								Text     string `xml:",chardata"`
								Kind     string `xml:"kind,attr"`
								Severity string `xml:"severity,attr"`
								Title    string `xml:"title"`
								Details  struct {
									Text   string `xml:",chardata"`
									Detail []struct {
										Text     string `xml:",chardata"`
										AttrText string `xml:"text,attr"`
										Details  struct {
											Text   string `xml:",chardata"`
											Detail []struct {
												Text     string `xml:",chardata"`
												AttrText string `xml:"text,attr"`
											} `xml:"detail"`
										} `xml:"details"`
									} `xml:"detail"`
								} `xml:"details"`
								Stack struct {
									Text       string `xml:",chardata"`
									StackEntry struct {
										Text        string `xml:",chardata"`
										URI         string `xml:"uri,attr"`
										Type        string `xml:"type,attr"`
										Name        string `xml:"name,attr"`
										Description string `xml:"description,attr"`
									} `xml:"stackEntry"`
								} `xml:"stack"`
							} `xml:"alert"`
						} `xml:"alerts"`
					} `xml:"testMethod"`
				} `xml:"testMethods"`
			} `xml:"testClass"`
		} `xml:"testClasses"`
	} `xml:"program"`
}
