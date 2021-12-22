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
	"strings"
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
		err = fetchAUnitResults(resp, details, client, config.AUnitResultsFileName, config.GenerateHTML)
	}
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
	log.Entry().Info("AUnit test run completed successfully. If there are any results from the respective run they will be listed in the logs above as well as being saved in the output .xml file")
	return nil
}

func triggerAUnitrun(config abapEnvironmentRunAUnitTestOptions, details abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) (*http.Response, error) {
	abapEndpoint := details.URL

	AUnitConfig, err := resolveAUnitConfiguration(config)
	if err != nil {
		return nil, err
	}

	var metadataString, optionsString, objectSetString string
	metadataString, optionsString, objectSetString, err = buildAUnitRequestBody(AUnitConfig)
	if err != nil {
		return nil, err
	}

	//Trigger AUnit run
	var resp *http.Response
	var bodyString = `<?xml version="1.0" encoding="UTF-8"?>` + metadataString + optionsString + objectSetString
	var body = []byte(bodyString)

	log.Entry().Debugf("Request Body: %s", bodyString)
	details.URL = abapEndpoint + "/sap/bc/adt/api/abapunit/runs"
	resp, err = runAUnit("POST", details, body, client)
	return resp, err
}

func resolveAUnitConfiguration(config abapEnvironmentRunAUnitTestOptions) (aUnitConfig AUnitConfig, err error) {

	if config.AUnitConfig != "" {
		// Configuration defaults to AUnitConfig
		result, err := readAUnitConfigFile(config.AUnitConfig)
		if err != nil {
			return aUnitConfig, err
		}
		err = json.Unmarshal(result, &aUnitConfig)
		return aUnitConfig, err

	} else if config.Repositories != "" {
		// Fallback / EasyMode is the Repositories configuration
		result, err := readAUnitConfigFile(config.Repositories)
		if err != nil {
			return aUnitConfig, err
		}
		repos := []abaputils.Repository{}
		err = json.Unmarshal(result, &repos)
		if err != nil {
			return aUnitConfig, err
		}
		aUnitConfig.ObjectSet = append(aUnitConfig.ObjectSet, ObjectSet{})
		for _, repo := range repos {
			aUnitConfig.ObjectSet[0].SoftwareComponents = append(aUnitConfig.ObjectSet[0].SoftwareComponents, SoftwareComponents{Name: repo.Name})
		}
		return aUnitConfig, nil
	} else {
		// Fail if no configuration is provided
		return aUnitConfig, errors.New("No configuration provided")
	}
}

func readAUnitConfigFile(path string) (file []byte, err error) {
	filelocation, err := filepath.Glob(path)
	//Parse YAML AUnit run configuration as body for AUnit run trigger
	if err != nil {
		return nil, err
	}
	filename, err := filepath.Abs(filelocation[0])
	if err != nil {
		return nil, err
	}
	aUnitConfigYamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var aUnitConfigJsonFile []byte
	aUnitConfigJsonFile, err = yaml.YAMLToJSON(aUnitConfigYamlFile)
	return aUnitConfigJsonFile, err

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

func fetchAUnitResults(resp *http.Response, details abaputils.ConnectionDetailsHTTP, client piperhttp.Sender, aunitResultFileName string, generateHTML bool) error {
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
		err = parseAUnitResult(body, aunitResultFileName, generateHTML)
	}
	if err != nil {
		return fmt.Errorf("Handling AUnit result failed: %w", err)
	}
	return nil
}

func buildAUnitRequestBody(AUnitConfig AUnitConfig) (metadataString string, optionsString string, objectSetString string, err error) {

	//Checks before building the XML body
	if AUnitConfig.Title == "" {
		return "", "", "", fmt.Errorf("Error while parsing AUnit test run config. No title for the AUnit run has been provided. Please configure an appropriate title for the respective test run")
	}
	if AUnitConfig.Context == "" {
		AUnitConfig.Context = "ABAP Environment Pipeline"
	}
	if reflect.DeepEqual(ObjectSet{}, AUnitConfig.ObjectSet) {
		return "", "", "", fmt.Errorf("Error while parsing AUnit test run object set config. No object set has been provided. Please configure the objects you want to be checked for the respective test run")
	}

	//Build Options
	optionsString += buildAUnitOptionsString(AUnitConfig)
	//Build metadata string
	metadataString += `<aunit:run title="` + AUnitConfig.Title + `" context="` + AUnitConfig.Context + `" xmlns:aunit="http://www.sap.com/adt/api/aunit">`

	//Build Object Set
	objectSetString += buildAUnitObjectSetString(AUnitConfig)
	objectSetString += `</aunit:run>`

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
	} else {
		optionsString += `<aunit:measurements type="none"/>`
	}
	//We assume there must be one scope configured
	optionsString += `<aunit:scope`
	if AUnitConfig.Options.Scope.OwnTests != nil {
		optionsString += ` ownTests="` + fmt.Sprintf("%v", *AUnitConfig.Options.Scope.OwnTests) + `"`
	} else {
		optionsString += ` ownTests="true"`
	}
	if AUnitConfig.Options.Scope.ForeignTests != nil {
		optionsString += ` foreignTests="` + fmt.Sprintf("%v", *AUnitConfig.Options.Scope.ForeignTests) + `"`
	} else {
		optionsString += ` foreignTests="true"`
	}
	//We assume there must be one riskLevel configured
	optionsString += `/><aunit:riskLevel`
	if AUnitConfig.Options.RiskLevel.Harmless != nil {
		optionsString += ` harmless="` + fmt.Sprintf("%v", *AUnitConfig.Options.RiskLevel.Harmless) + `"`
	} else {
		optionsString += ` harmless="true"`
	}
	if AUnitConfig.Options.RiskLevel.Dangerous != nil {
		optionsString += ` dangerous="` + fmt.Sprintf("%v", *AUnitConfig.Options.RiskLevel.Dangerous) + `"`
	} else {
		optionsString += ` dangerous="true"`
	}
	if AUnitConfig.Options.RiskLevel.Critical != nil {
		optionsString += ` critical="` + fmt.Sprintf("%v", *AUnitConfig.Options.RiskLevel.Critical) + `"`
	} else {
		optionsString += ` critical="true"`
	}
	//We assume there must be one duration time configured
	optionsString += `/><aunit:duration`
	if AUnitConfig.Options.Duration.Short != nil {
		optionsString += ` short="` + fmt.Sprintf("%v", *AUnitConfig.Options.Duration.Short) + `"`
	} else {
		optionsString += ` short="true"`
	}
	if AUnitConfig.Options.Duration.Medium != nil {
		optionsString += ` medium="` + fmt.Sprintf("%v", *AUnitConfig.Options.Duration.Medium) + `"`
	} else {
		optionsString += ` medium="true"`
	}
	if AUnitConfig.Options.Duration.Long != nil {
		optionsString += ` long="` + fmt.Sprintf("%v", *AUnitConfig.Options.Duration.Long) + `"`
	} else {
		optionsString += ` long="true"`
	}
	optionsString += `/></aunit:options>`
	return optionsString
}

func buildOSLObjectSets(multipropertyset MultiPropertySet) (objectSetString string) {
	objectSetString += writeObjectSetProperties(multipropertyset)
	return objectSetString
}

func writeObjectSetProperties(set MultiPropertySet) (objectSetString string) {
	for _, packages := range set.PackageNames {
		objectSetString += `<osl:package name="` + packages.Name + `"/>`
	}
	for _, objectTypeGroup := range set.ObjectTypeGroups {
		objectSetString += `<osl:objectTypeGroup name="` + objectTypeGroup.Name + `"/>`
	}
	for _, objectType := range set.ObjectTypes {
		objectSetString += `<osl:objectType name="` + objectType.Name + `"/>`
	}
	for _, owner := range set.Owners {
		objectSetString += `<osl:owner name="` + owner.Name + `"/>`
	}
	for _, releaseState := range set.ReleaseStates {
		objectSetString += `<osl:releaseState value="` + releaseState.Value + `"/>`
	}
	for _, version := range set.Versions {
		objectSetString += `<osl:version value="` + version.Value + `"/>`
	}
	for _, applicationComponent := range set.ApplicationComponents {
		objectSetString += `<osl:applicationComponent name="` + applicationComponent.Name + `"/>`
	}
	for _, component := range set.SoftwareComponents {
		objectSetString += `<osl:softwareComponent name="` + component.Name + `"/>`
	}
	for _, transportLayer := range set.TransportLayers {
		objectSetString += `<osl:transportLayer name="` + transportLayer.Name + `"/>`
	}
	for _, language := range set.Languages {
		objectSetString += `<osl:language value="` + language.Value + `"/>`
	}
	for _, sourceSystem := range set.SourceSystems {
		objectSetString += `<osl:sourceSystem name="` + sourceSystem.Name + `"/>`
	}
	return objectSetString
}

func buildAUnitObjectSetString(AUnitConfig AUnitConfig) (objectSetString string) {

	//Build ObjectSets
	s := AUnitConfig.ObjectSet
	if s.Type == "" {
		s.Type = "multiPropertySet"
	}
	if s.Type != "multiPropertySet" {
		log.Entry().Infof("Wrong configuration has been detected: %s has been used. This is currently not supported and this set will not be included in this run. Please check the step documentation for more information", s.Type)
	} else {
		objectSetString += `<osl:objectSet xsi:type="` + s.Type + `" xmlns:osl="http://www.sap.com/api/osl" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">`

		if !(reflect.DeepEqual(s.PackageNames, AUnitPackage{})) || !(reflect.DeepEqual(s.SoftwareComponents, SoftwareComponents{})) {
			//To ensure Scomps and packages can be assigned on this level
			mps := MultiPropertySet{
				PackageNames:       s.PackageNames,
				SoftwareComponents: s.SoftwareComponents,
			}
			objectSetString += buildOSLObjectSets(mps)
		}

		objectSetString += buildOSLObjectSets(s.MultiPropertySet)

		if !(reflect.DeepEqual(s.MultiPropertySet, MultiPropertySet{})) {
			log.Entry().Info("Wrong configuration has been detected: MultiPropertySet has been used. Please note that there is no official documentation for this usage. Please check the step documentation for more information")
		}

		for _, t := range s.Set {
			log.Entry().Infof("Wrong configuration has been detected: %s has been used. This is currently not supported and this set will not be included in this run. Please check the step documentation for more information", t.Type)
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

func parseAUnitResult(body []byte, aunitResultFileName string, generateHTML bool) (err error) {
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
	var reports []piperutils.Path
	//Return before processing empty AUnit results --> XML can still be written with response body
	if len(parsedXML.Testsuite.Testcase) == 0 {
		log.Entry().Infof("There were no AUnit findings from this run. The response has been saved in the %s file", aunitResultFileName)
	} else {
		log.Entry().Infof("Please find the results from the respective AUnit run in the %s file or in below logs", aunitResultFileName)
		//Logging of AUnit findings
		log.Entry().Infof(`Here are the results for the AUnit test run '%s' executed by User %s on System %s in Client %s at %s. The AUnit run took %s seconds and contains %s tests with %s failures, %s errors, %s skipped and %s assert findings`, parsedXML.Title, parsedXML.System, parsedXML.ExecutedBy, parsedXML.Client, parsedXML.Timestamp, parsedXML.Time, parsedXML.Tests, parsedXML.Failures, parsedXML.Errors, parsedXML.Skipped, parsedXML.Asserts)
		for _, s := range parsedXML.Testsuite.Testcase {
			//Log Infos for testcase
			//HTML Procesing can be done here
			for _, failure := range s.Failure {
				log.Entry().Debugf("%s, %s: %s found by %s", failure.Type, failure.Message, failure.Message, s.Classname)
			}
			for _, skipped := range s.Skipped {
				log.Entry().Debugf("The following test has been skipped: %s: %s", skipped.Message, skipped.Text)
			}
		}
		if generateHTML == true {
			htmlString := generateHTMLDocumentAUnit(parsedXML)
			htmlStringByte := []byte(htmlString)
			aUnitResultHTMLFileName := strings.Trim(aunitResultFileName, ".xml") + ".html"
			err = ioutil.WriteFile(aUnitResultHTMLFileName, htmlStringByte, 0644)
			if err != nil {
				return fmt.Errorf("Writing HTML document failed: %w", err)
			}
			log.Entry().Info("Writing " + aUnitResultHTMLFileName + " file was successful")
			reports = append(reports, piperutils.Path{Target: aUnitResultHTMLFileName, Name: "ATC Results HTML file", Mandatory: true})
		}
	}
	//Persist findings afterwards
	reports = append(reports, piperutils.Path{Target: aunitResultFileName, Name: "AUnit Results", Mandatory: true})
	piperutils.PersistReportsAndLinks("abapEnvironmentRunAUnitTest", "", reports, nil)
	return nil
}

func generateHTMLDocumentAUnit(parsedXML *AUnitResult) (htmlDocumentString string) {
	htmlDocumentString = `<!DOCTYPE html><html lang="en" xmlns="http://www.w3.org/1999/xhtml"><head><title>AUnit Results</title><meta http-equiv="Content-Type" content="text/html; charset=UTF-8" /><style>table,th,td {border-collapse:collapse;}th,td{padding: 5px;text-align:left;font-size:medium;}</style></head><body><h1 style="text-align:left;font-size:large">AUnit Results</h1><table><tr><th>Run title</th><td style="padding-right: 20px">` + parsedXML.Title + `</td><th>System</th><td style="padding-right: 20px">` + parsedXML.System + `</td><th>Client</th><td style="padding-right: 20px">` + parsedXML.Client + `</td><th>ExecutedBy</th><td style="padding-right: 20px">` + parsedXML.ExecutedBy + `</td><th>Duration</th><td style="padding-right: 20px">` + parsedXML.Time + `s</td><th>Timestamp</th><td style="padding-right: 20px">` + parsedXML.Timestamp + `</td></tr><tr><th>Failures</th><td style="padding-right: 20px">` + parsedXML.Failures + `</td><th>Errors</th><td style="padding-right: 20px">` + parsedXML.Errors + `</td><th>Skipped</th><td style="padding-right: 20px">` + parsedXML.Skipped + `</td><th>Asserts</th><td style="padding-right: 20px">` + parsedXML.Asserts + `</td><th>Tests</th><td style="padding-right: 20px">` + parsedXML.Tests + `</td></tr></table><br><table style="width:100%; border: 1px solid black""><tr style="border: 1px solid black"><th style="border: 1px solid black">Severity</th><th style="border: 1px solid black">File</th><th style="border: 1px solid black">Message</th><th style="border: 1px solid black">Type</th><th style="border: 1px solid black">Text</th></tr>`

	var htmlDocumentStringError, htmlDocumentStringWarning, htmlDocumentStringInfo, htmlDocumentStringDefault string
	for _, s := range parsedXML.Testsuite.Testcase {
		//Add coloring of lines inside of the respective severities, e.g. failures in red
		trBackgroundColorTestcase := "grey"
		trBackgroundColorError := "rgba(227,85,0)"
		trBackgroundColorFailure := "rgba(227,85,0)"
		trBackgroundColorSkipped := "rgba(255,175,0, 0.2)"
		if (len(s.Error) != 0) || (len(s.Failure) != 0) || (len(s.Skipped) != 0) {
			htmlDocumentString += `<tr style="background-color: ` + trBackgroundColorTestcase + `"><td colspan="5"><b>Testcase: ` + s.Name + ` for class ` + s.Classname + `</b></td></tr>`
		}
		for _, t := range s.Error {
			htmlDocumentString += `<tr style="background-color: ` + trBackgroundColorError + `"><td style="border: 1px solid black">Failure</td><td style="border: 1px solid black">` + s.Classname + `</td><td style="border: 1px solid black">` + t.Message + `</td><td style="border: 1px solid black">` + t.Type + `</td><td style="border: 1px solid black">` + t.Text + `</td></tr>`
		}
		for _, t := range s.Failure {
			htmlDocumentString += `<tr style="background-color: ` + trBackgroundColorFailure + `"><td style="border: 1px solid black">Failure</td><td style="border: 1px solid black">` + s.Classname + `</td><td style="border: 1px solid black">` + t.Message + `</td><td style="border: 1px solid black">` + t.Type + `</td><td style="border: 1px solid black">` + t.Text + `</td></tr>`
		}
		for _, t := range s.Skipped {
			htmlDocumentString += `<tr style="background-color: ` + trBackgroundColorSkipped + `"><td style="border: 1px solid black">Failure</td><td style="border: 1px solid black">` + s.Classname + `</td><td style="border: 1px solid black">` + t.Message + `</td><td style="border: 1px solid black">-</td><td style="border: 1px solid black">` + t.Text + `</td></tr>`
		}
	}
	if len(parsedXML.Testsuite.Testcase) == 0 {
		htmlDocumentString += `<tr><td colspan="5"><b>There are no AUnit findings to be displayed</b></td></tr>`
	}
	htmlDocumentString += htmlDocumentStringError + htmlDocumentStringWarning + htmlDocumentStringInfo + htmlDocumentStringDefault + `</table></body></html>`

	return htmlDocumentString
}

//
//	Object Set Structure
//

//AUnitConfig object for parsing yaml config of software components and packages
type AUnitConfig struct {
	Title     string       `json:"title,omitempty"`
	Context   string       `json:"context,omitempty"`
	Options   AUnitOptions `json:"options,omitempty"`
	ObjectSet ObjectSet    `json:"objectset,omitempty"`
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
	PackageNames       []AUnitPackage       `json:"packages,omitempty"`
	SoftwareComponents []SoftwareComponents `json:"softwarecomponents,omitempty"`
	Type               string               `json:"type,omitempty"`
	MultiPropertySet   MultiPropertySet     `json:"multipropertyset,omitempty"`
	Set                []Set                `json:"set,omitempty"`
}

//MultiPropertySet that can possibly contain any subsets/object of the OSL
type MultiPropertySet struct {
	Type                  string                 `json:"type,omitempty"`
	PackageNames          []AUnitPackage         `json:"packages,omitempty"`
	ObjectTypeGroups      []ObjectTypeGroup      `json:"objecttypegroup,omitempty"`
	ObjectTypes           []ObjectType           `json:"objecttypes,omitempty"`
	Owners                []Owner                `json:"owner,omitempty"`
	ReleaseStates         []ReleaseState         `json:"releasestate,omitempty"`
	Versions              []Version              `json:"version,omitempty"`
	ApplicationComponents []ApplicationComponent `json:"applicationcomponent,omitempty"`
	SoftwareComponents    []SoftwareComponents   `json:"softwarecomponents,omitempty"`
	TransportLayers       []TransportLayer       `json:"transportlayer,omitempty"`
	Languages             []Language             `json:"language,omitempty"`
	SourceSystems         []SourceSystem         `json:"sourcesystem,omitempty"`
}

//Set
type Set struct {
	Type          string               `json:"type,omitempty"`
	Set           []Set                `json:"set,omitempty"`
	PackageSet    []AUnitPackageSet    `json:"package,omitempty"`
	FlatObjectSet []AUnitFlatObjectSet `json:"object,omitempty"`
	ComponentSet  []AUnitComponentSet  `json:"component,omitempty"`
	TransportSet  []AUnitTransportSet  `json:"transport,omitempty"`
	ObjectTypeSet []AUnitObjectTypeSet `json:"objecttype,omitempty"`
}

//AUnitPackageSet in form of packages to be checked
type AUnitPackageSet struct {
	Name               string `json:"name,omitempty"`
	IncludeSubpackages *bool  `json:"includesubpackages,omitempty"`
}

//AUnitFlatObjectSet
type AUnitFlatObjectSet struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

//AUnitComponentSet in form of software components to be checked
type AUnitComponentSet struct {
	Name string `json:"name,omitempty"`
}

//AUnitTransportSet in form of transports to be checked
type AUnitTransportSet struct {
	Number string `json:"number,omitempty"`
}

//AUnitObjectTypeSet
type AUnitObjectTypeSet struct {
	Name string `json:"name,omitempty"`
}

//AUnitPackage for MPS
type AUnitPackage struct {
	Name string `json:"name,omitempty"`
}

//ObjectTypeGroup
type ObjectTypeGroup struct {
	Name string `json:"name,omitempty"`
}

//ObjectType
type ObjectType struct {
	Name string `json:"name,omitempty"`
}

//Owner
type Owner struct {
	Name string `json:"name,omitempty"`
}

//ReleaseState
type ReleaseState struct {
	Value string `json:"value,omitempty"`
}

//Version
type Version struct {
	Value string `json:"value,omitempty"`
}

//ApplicationComponent
type ApplicationComponent struct {
	Name string `json:"name,omitempty"`
}

//SoftwareComponents
type SoftwareComponents struct {
	Name string `json:"name,omitempty"`
}

//TransportLayer
type TransportLayer struct {
	Name string `json:"name,omitempty"`
}

//Language
type Language struct {
	Value string `json:"value,omitempty"`
}

//SourceSystem
type SourceSystem struct {
	Name string `json:"name,omitempty"`
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
