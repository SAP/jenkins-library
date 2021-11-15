package cmd

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"path/filepath"
	"reflect"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

type abapEnvironmentRunAUnitTestUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The abapEnvironmentRunAUnitTestUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type abapEnvironmentRunAUnitTestUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to abapEnvironmentRunAUnitTestUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// abapEnvironmentRunAUnitTestUtilsBundle and forward to the implementation of the dependency.
}

func newAbapEnvironmentRunAUnitTestUtils() abapEnvironmentRunAUnitTestUtils {
	utils := abapEnvironmentRunAUnitTestUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func abapEnvironmentRunAUnitTest(config abapEnvironmentRunAUnitTestOptions, telemetryData *telemetry.CustomData) {

	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	var autils = abaputils.AbapUtils{
		Exec: &c,
	}

	client := piperhttp.Client{}

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapEnvironmentRunAUnitTest(&config, telemetryData, &autils, &client)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentRunAUnitTest(config *abapEnvironmentRunAUnitTestOptions, telemetryData *telemetry.CustomData, com abaputils.Communication, client piperhttp.Sender) error {
	var details abaputils.ConnectionDetailsHTTP
	subOptions := convertAUnitOptions(config)
	details, err := com.GetAbapCommunicationArrangementInfo(subOptions, "")
	var resp *http.Response
	cookieJar, _ := cookiejar.New(nil)
	//Fetch Xcrsf-Token
	if err == nil {
		credentialsOptions := piperhttp.ClientOptions{
			Username:  details.User,
			Password:  details.Password,
			CookieJar: cookieJar,
		}
		client.SetOptions(credentialsOptions)
		details.XCsrfToken, err = fetchAUnitXcsrfToken("GET", details, nil, client)
	}
	if err == nil {
		resp, err = triggerAUnitrun(*config, details, client)
	}
	if err == nil {
		err = handleAUnitResults(resp, details, client, config.AUnitResultsFileName)
	}
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
	log.Entry().Info("AUnit test run completed successfully. If there are any results from the respective run they will be listed in the logs above as well as being saved in the output .xml file")
	return nil
}

func triggerAUnitrun(config abapEnvironmentRunAUnitTestOptions, details abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) (*http.Response, error) {
	var aUnitConfigYamlFile []byte
	abapEndpoint := details.URL
	filelocation, err := filepath.Glob(config.AUnitConfig)
	//Parse YAML AUnit run configuration as body for AUnit run trigger
	if err == nil {
		filename, err := filepath.Abs(filelocation[0])
		if err == nil {
			aUnitConfigYamlFile, err = ioutil.ReadFile(filename)
		}
	}
	var AUnitConfig AUnitConfig
	if err == nil {
		var result []byte
		result, err = yaml.YAMLToJSON(aUnitConfigYamlFile)
		json.Unmarshal(result, &AUnitConfig)
	}
	var metadataString, optionsString, objectSetString string
	if err == nil {
		metadataString, optionsString, objectSetString, err = buildAUnitTestBody(AUnitConfig)
	}

	//Trigger AUnit run
	var resp *http.Response
	var bodyString = `<?xml version="1.0" encoding="UTF-8"?>` + metadataString + optionsString + objectSetString + `</aunit:run>`
	var body = []byte(bodyString)
	if err == nil {
		log.Entry().Debugf("Request Body: %s", bodyString)
		details.URL = abapEndpoint + "/sap/bc/adt/api/abapunit/runs"
		resp, err = runAUnit("POST", details, body, client)
	}
	if err != nil {
		return resp, fmt.Errorf("Triggering AUnit test run failed: %w", err)
	}
	return resp, nil
}

func convertAUnitOptions(options *abapEnvironmentRunAUnitTestOptions) abaputils.AbapEnvironmentOptions {
	subOptions := abaputils.AbapEnvironmentOptions{}

	subOptions.CfAPIEndpoint = options.CfAPIEndpoint
	subOptions.CfServiceInstance = options.CfServiceInstance
	subOptions.CfServiceKeyName = options.CfServiceKeyName
	subOptions.CfOrg = options.CfOrg
	subOptions.CfSpace = options.CfSpace
	subOptions.Host = options.Host
	subOptions.Password = options.Password
	subOptions.Username = options.Username

	return subOptions
}

func handleAUnitResults(resp *http.Response, details abaputils.ConnectionDetailsHTTP, client piperhttp.Sender, aunitResultFileName string) error {
	var err error
	var abapEndpoint string
	abapEndpoint = details.URL
	location := resp.Header.Get("Location")
	details.URL = abapEndpoint + location
	location, err = pollAUnitRun(details, nil, client)
	if err == nil {
		details.URL = abapEndpoint + location
		resp, err = getResultAUnitRun("GET", details, nil, client)
	}
	//Parse response
	var body []byte
	if err == nil {
		body, err = ioutil.ReadAll(resp.Body)
	}
	if err == nil {
		defer resp.Body.Close()
		err = parseAUnitResult(body, aunitResultFileName)
	}
	if err != nil {
		return fmt.Errorf("Handling AUnit result failed: %w", err)
	}
	return nil
}

func buildAUnitTestBody(AUnitConfig AUnitConfig) (metadataString string, optionsString string, objectSetString string, err error) {

	//Checks before building the XML body
	if AUnitConfig.Title == "" {
		return "", "", "", fmt.Errorf("Error while parsing AUnit test run config. No title for the AUnit run has been provided. Please configure an appropriate title for the respective test run")
	} else if AUnitConfig.Context == "" {
		return "", "", "", fmt.Errorf("Error while parsing AUnit test run config. No context for the AUnit run has been provided. Please configure an appropriate context for the respective test run")
	} else if reflect.DeepEqual(AUnitOptions{}, AUnitConfig.Options) {
		return "", "", "", fmt.Errorf("Error while parsing AUnit test run config. No options have been provided. Please configure the options for the respective test run")
	} else if reflect.DeepEqual(ObjectSet{}, AUnitConfig.ObjectSet) {
		return "", "", "", fmt.Errorf("Error while parsing AUnit test run object set config. No object set has been provided. Please configure the objects you want to be checked for the respective test run")
	} else if len(AUnitConfig.ObjectSet) == 0 {
		return "", "", "", fmt.Errorf("Error while parsing AUnit test run object set config. No object set has been provided. Please configure the set of objects you want to be checked for the respective test run")
	}

	//Build metadata string
	metadataString += `<aunit:run title="` + AUnitConfig.Title + `" context="` + AUnitConfig.Context + `" xmlns:aunit="http://www.sap.com/adt/api/aunit">`

	//Build Options
	optionsString += buildAUnitOptionsString(AUnitConfig)

	//Build Object Set
	objectSetString += buildAUnitObjectSetString(AUnitConfig)

	return metadataString, optionsString, objectSetString, nil
}

func runAUnit(requestType string, details abaputils.ConnectionDetailsHTTP, body []byte, client piperhttp.Sender) (*http.Response, error) {

	log.Entry().WithField("ABAP endpoint: ", details.URL).Info("Triggering AUnit run")

	header := make(map[string][]string)
	header["X-Csrf-Token"] = []string{details.XCsrfToken}
	header["Content-Type"] = []string{"application/vnd.sap.adt.api.abapunit.run.v1+xml; charset=utf-8;"}

	req, err := client.SendRequest(requestType, details.URL, bytes.NewBuffer(body), header, nil)
	if err != nil {
		return req, fmt.Errorf("Triggering AUnit run failed: %w", err)
	}
	defer req.Body.Close()
	return req, err
}

func buildAUnitOptionsString(AUnitConfig AUnitConfig) (optionsString string) {

	optionsString += `<aunit:options>`
	if AUnitConfig.Options.Measurements != "" {
		optionsString += `<aunit:measurements type="` + AUnitConfig.Options.Measurements + `"/>`
	}
	optionsString += `<aunit:scope`
	if AUnitConfig.Options.Scope.OwnTests != nil {
		optionsString += ` ownTests="` + fmt.Sprintf("%v", *AUnitConfig.Options.Scope.OwnTests) + `"`
	}
	if AUnitConfig.Options.Scope.ForeignTests != nil {
		optionsString += ` foreignTests="` + fmt.Sprintf("%v", *AUnitConfig.Options.Scope.ForeignTests) + `"`
	}
	optionsString += `/><aunit:riskLevel`
	if AUnitConfig.Options.RiskLevel.Harmless != nil {
		optionsString += ` harmless="` + fmt.Sprintf("%v", *AUnitConfig.Options.RiskLevel.Harmless) + `"`
	}
	if AUnitConfig.Options.RiskLevel.Dangerous != nil {
		optionsString += ` dangerous="` + fmt.Sprintf("%v", *AUnitConfig.Options.RiskLevel.Dangerous) + `"`
	}
	if AUnitConfig.Options.RiskLevel.Critical != nil {
		optionsString += ` critical="` + fmt.Sprintf("%v", *AUnitConfig.Options.RiskLevel.Critical) + `"`
	}
	optionsString += `/><aunit:duration`
	if AUnitConfig.Options.Duration.Short != nil {
		optionsString += ` short="` + fmt.Sprintf("%v", *AUnitConfig.Options.Duration.Short) + `"`
	}
	if AUnitConfig.Options.Duration.Medium != nil {
		optionsString += ` medium="` + fmt.Sprintf("%v", *AUnitConfig.Options.Duration.Medium) + `"`
	}
	if AUnitConfig.Options.Duration.Long != nil {
		optionsString += ` long="` + fmt.Sprintf("%v", *AUnitConfig.Options.Duration.Long) + `"`
	}
	optionsString += `/></aunit:options>`
	return optionsString
}

func buildAUnitObjectSetString(AUnitConfig AUnitConfig) (objectSetString string) {

	//Build ObjectSets {
	for _, s := range AUnitConfig.ObjectSet {
		//Write object set
		objectSetString += `<osl:objectSet xsi:type="` + s.Type + `" xmlns:osl="http://www.sap.com/api/osl" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">`
		for _, t := range s.Set {
			objectSetString += `<osl:set xsi:type="` + t.Type + `">`
			for _, packageSet := range t.PackageSet {
				objectSetString += `<osl:package name="` + packageSet.Name + `" includeSubpackages="` + fmt.Sprintf("%v", *packageSet.IncludeSubpackages) + `"/>`
			}
			for _, flatObjectSet := range t.FlatObjectSet {
				objectSetString += `<osl:object name="` + flatObjectSet.Name + `" type="` + fmt.Sprintf("%v", *&flatObjectSet.Type) + `"/>`
			}
			objectSetString += `</osl:set>`
		}

		objectSetString += `</osl:objectSet>`
	}
	return objectSetString
}

func fetchAUnitXcsrfToken(requestType string, details abaputils.ConnectionDetailsHTTP, body []byte, client piperhttp.Sender) (string, error) {

	log.Entry().WithField("ABAP Endpoint: ", details.URL).Debug("Fetching Xcrsf-Token")

	details.URL += "/sap/bc/adt/api/abapunit/runs/00000000000000000000000000000000"
	details.XCsrfToken = "fetch"
	header := make(map[string][]string)
	header["X-Csrf-Token"] = []string{details.XCsrfToken}
	header["Accept"] = []string{"application/vnd.sap.adt.api.abapunit.run-status.v1+xml"}
	req, err := client.SendRequest(requestType, details.URL, bytes.NewBuffer(body), header, nil)
	if err != nil {
		return "", fmt.Errorf("Fetching Xcsrf-Token failed: %w", err)
	}
	defer req.Body.Close()

	token := req.Header.Get("X-Csrf-Token")
	return token, err
}

func pollAUnitRun(details abaputils.ConnectionDetailsHTTP, body []byte, client piperhttp.Sender) (string, error) {

	log.Entry().WithField("ABAP endpoint", details.URL).Info("Polling AUnit run status")

	for {
		resp, err := getHTTPResponseAUnitRun("GET", details, nil, client)
		if err != nil {
			return "", fmt.Errorf("Getting HTTP response failed: %w", err)
		}
		bodyText, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("Reading response body failed: %w", err)
		}
		x := new(AUnitRun)
		xml.Unmarshal(bodyText, &x)

		log.Entry().Infof("Current polling status: %s", x.Progress.Status)
		if x.Progress.Status == "Not Created" {
			return "", err
		}
		if x.Progress.Status == "Completed" || x.Progress.Status == "FINISHED" {
			return x.Link.Href, err
		}
		if x.Progress.Status == "" {
			return "", fmt.Errorf("Could not get any response from AUnit poll: %w", errors.New("Status from AUnit run is empty. Either it's not an ABAP system or AUnit run hasn't started"))
		}
		time.Sleep(10 * time.Second)
	}
}

func getHTTPResponseAUnitRun(requestType string, details abaputils.ConnectionDetailsHTTP, body []byte, client piperhttp.Sender) (*http.Response, error) {

	log.Entry().WithField("ABAP Endpoint: ", details.URL).Info("Polling AUnit run status")

	header := make(map[string][]string)
	header["Accept"] = []string{"application/vnd.sap.adt.api.abapunit.run-status.v1+xml"}

	req, err := client.SendRequest(requestType, details.URL, bytes.NewBuffer(body), header, nil)
	if err != nil {
		return req, fmt.Errorf("Getting AUnit run status failed: %w", err)
	}
	return req, err
}

func getResultAUnitRun(requestType string, details abaputils.ConnectionDetailsHTTP, body []byte, client piperhttp.Sender) (*http.Response, error) {

	log.Entry().WithField("ABAP Endpoint: ", details.URL).Info("Getting AUnit results")

	header := make(map[string][]string)
	header["x-csrf-token"] = []string{details.XCsrfToken}
	header["Accept"] = []string{"application/vnd.sap.adt.api.junit.run-result.v1+xml"}

	req, err := client.SendRequest(requestType, details.URL, bytes.NewBuffer(body), header, nil)
	if err != nil {
		return req, fmt.Errorf("Getting AUnit run results failed: %w", err)
	}
	return req, err
}

func parseAUnitResult(body []byte, aunitResultFileName string) (err error) {
	if len(body) == 0 {
		return fmt.Errorf("Parsing AUnit result failed: %w", errors.New("Body is empty, can't parse empty body"))
	}

	responseBody := string(body)
	log.Entry().Debugf("Response body: %s", responseBody)

	//Optional checks before writing the Results
	parsedXML := new(AUnitResult)
	xml.Unmarshal([]byte(body), &parsedXML)

	//Write Results
	err = ioutil.WriteFile(aunitResultFileName, body, 0644)
	if err != nil {
		return fmt.Errorf("Writing results failed: %w", err)
	}
	log.Entry().Infof("Writing %s file was successful.", aunitResultFileName)
	//Return before processing empty AUnit results --> XML can still be written with response body
	if len(parsedXML.Testsuite.Testcase) == 0 {
		log.Entry().Infof("There were no AUnit findings from this run. The response has been saved in the %s file", aunitResultFileName)
	} else {
		log.Entry().Infof("Please find the results from the respective AUnit run in the %s file or in below logs", aunitResultFileName)
		//Logging of AUnit findings
		log.Entry().Infof(`Here are the results for the AUnit test run '%s' executed by User %s on System %s in Client %s at %s. The AUnit run took %s seconds and contains %s tests with %s failures, %s errors, %s skipped and %s assert findings`, parsedXML.Title, parsedXML.System, parsedXML.ExecutedBy, parsedXML.Client, parsedXML.Timestamp, parsedXML.Time, parsedXML.Tests, parsedXML.Failures, parsedXML.Errors, parsedXML.Skipped, parsedXML.Asserts)
		for _, s := range parsedXML.Testsuite.Testcase {
			//Log Infos for testcase
			for _, failure := range s.Failure {
				log.Entry().Debugf("%s, %s: %s found by %s", failure.Type, failure.Message, failure.Message, s.Classname)
			}
			for _, skipped := range s.Skipped {
				log.Entry().Debugf("The following test has been skipped: %s: %s", skipped.Message, skipped.Text)
			}
		}
	}
	//Persist findings afterwards
	var reports []piperutils.Path
	reports = append(reports, piperutils.Path{Target: aunitResultFileName, Name: "AUnit Results", Mandatory: true})
	piperutils.PersistReportsAndLinks("abapEnvironmentRunAUnitTest", "", reports, nil)
	return nil
}

//
//	Object Set Structure
//

//AUnitConfig object for parsing yaml config of software components and packages
type AUnitConfig struct {
	Title     string       `json:"title,omitempty"`
	Context   string       `json:"context,omitempty"`
	Options   AUnitOptions `json:"options,omitempty"`
	ObjectSet []ObjectSet  `json:"objectset,omitempty"`
}

//AUnitOptions in form of packages and software components to be checked
type AUnitOptions struct {
	Measurements string    `json:"measurements,omitempty"`
	Scope        Scope     `json:"scope,omitempty"`
	RiskLevel    RiskLevel `json:"risklevel,omitempty"`
	Duration     Duration  `json:"duration,omitempty"`
}

//Scope in form of packages and software components to be checked
type Scope struct {
	OwnTests     *bool `json:"owntests,omitempty"`
	ForeignTests *bool `json:"foreigntests,omitempty"`
}

//RiskLevel in form of packages and software components to be checked
type RiskLevel struct {
	Harmless  *bool `json:"harmless,omitempty"`
	Dangerous *bool `json:"dangerous,omitempty"`
	Critical  *bool `json:"critical,omitempty"`
}

//Duration in form of packages and software components to be checked
type Duration struct {
	Short  *bool `json:"short,omitempty"`
	Medium *bool `json:"medium,omitempty"`
	Long   *bool `json:"long,omitempty"`
}

//ObjectSet in form of packages and software components to be checked
type ObjectSet struct {
	Type string `json:"type,omitempty"`
	Set  []Set  `json:"set,omitempty"`
}

//Set in form of packages and software components to be checked
type Set struct {
	//Set  []Set  `json:"set,omitempty"`
	Type          string         `json:"type,omitempty"`
	PackageSet    []AUnitPackage `json:"package,omitempty"`
	FlatObjectSet []AUnitObject  `json:"object,omitempty"`
	/*FlatSet       []FlatObjectSet `json:"flatobjectset,omitempty"`
	ObjectTypeSet []ObjectTypeSet `json:"objecttypeset,omitempty"`
	ComponentSet  []ComponentSet  `json:"componentset,omitempty"`
	TransportSet  []TransportSet  `json:"transportset,omitempty"`*/
}

//AUnitPackage in form of packages and software components to be checked
type AUnitPackage struct {
	Name               string `json:"name,omitempty"`
	IncludeSubpackages *bool  `json:"includesubpackages,omitempty"`
}

//AUnitObject in form of packages and software components to be checked
type AUnitObject struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

//
//	AUnit Run Structure
//

//AUnitRun Object for parsing XML
type AUnitRun struct {
	XMLName    xml.Name   `xml:"run"`
	Title      string     `xml:"title,attr"`
	Context    string     `xml:"context,attr"`
	Progress   Progress   `xml:"progress"`
	ExecutedBy ExecutedBy `xml:"executedBy"`
	Time       Time       `xml:"time"`
	Link       AUnitLink  `xml:"link"`
}

//Progress of AUnit run
type Progress struct {
	Status     string `xml:"status,attr"`
	Percentage string `xml:"percentage,attr"`
}

//ExecutedBy User
type ExecutedBy struct {
	User string `xml:"user,attr"`
}

//Time run was started and finished
type Time struct {
	Started string `xml:"started,attr"`
	Ended   string `xml:"ended,attr"`
}

//AUnitLink containing result locations
type AUnitLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

//
//	AUnit Result Structure
//

type AUnitResult struct {
	XMLName    xml.Name `xml:"testsuites"`
	Title      string   `xml:"title,attr"`
	System     string   `xml:"system,attr"`
	Client     string   `xml:"client,attr"`
	ExecutedBy string   `xml:"executedBy,attr"`
	Time       string   `xml:"time,attr"`
	Timestamp  string   `xml:"timestamp,attr"`
	Failures   string   `xml:"failures,attr"`
	Errors     string   `xml:"errors,attr"`
	Skipped    string   `xml:"skipped,attr"`
	Asserts    string   `xml:"asserts,attr"`
	Tests      string   `xml:"tests,attr"`
	Testsuite  struct {
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
	} `xml:"testsuite"`
}
