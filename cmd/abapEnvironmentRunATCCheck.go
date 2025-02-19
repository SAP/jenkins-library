package cmd

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func abapEnvironmentRunATCCheck(options abapEnvironmentRunATCCheckOptions, _ *telemetry.CustomData) {
	// Mapping for options
	subOptions := convertATCOptions(&options)

	c := &command.Command{}
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())

	autils := abaputils.AbapUtils{
		Exec: c,
	}

	client := piperhttp.Client{}
	fileUtils := piperutils.Files{}
	cookieJar, _ := cookiejar.New(nil)
	clientOptions := piperhttp.ClientOptions{
		CookieJar: cookieJar,
	}
	client.SetOptions(clientOptions)

	err := runAbapEnvironmentRunATCCheck(autils, subOptions, cookieJar, client, options, fileUtils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentRunATCCheck(autils abaputils.AbapUtils, subOptions abaputils.AbapEnvironmentOptions, cookieJar *cookiejar.Jar, client piperhttp.Client, options abapEnvironmentRunATCCheckOptions, fileUtils piperutils.Files) error {
	var details abaputils.ConnectionDetailsHTTP
	// If Host flag is empty read ABAP endpoint from Service Key instead. Otherwise take ABAP system endpoint from config instead

	details, err := autils.GetAbapCommunicationArrangementInfo(subOptions, "")
	if err != nil {
		return err
	}

	credentialsOptions := piperhttp.ClientOptions{
		Username:  details.User,
		Password:  details.Password,
		CookieJar: cookieJar,
	}
	client.SetOptions(credentialsOptions)
	details.XCsrfToken, err = fetchXcsrfToken("GET", details, nil, &client)
	if err != nil {
		return err
	}
	resp, err := triggerATCRun(options, details, &client)
	if err != nil {
		return err
	}
	if err = fetchAndPersistATCResults(resp, details, &client, &fileUtils, options.AtcResultsFileName, options.GenerateHTML, options.FailOnSeverity); err != nil {
		return err
	}

	log.Entry().Info("ATC run completed successfully. If there are any results from the respective run they will be listed in the logs above as well as being saved in the output .xml file")
	return nil
}

func fetchAndPersistATCResults(resp *http.Response, details abaputils.ConnectionDetailsHTTP, client piperhttp.Sender, utils piperutils.FileUtils, atcResultFileName string, generateHTML bool, failOnSeverityLevel string) error {
	var err error
	var failStep bool
	abapEndpoint := details.URL
	location := resp.Header.Get("Location")
	details.URL = abapEndpoint + location
	location, err = pollATCRun(details, nil, client)
	if err == nil {
		details.URL = abapEndpoint + location
		resp, err = getResultATCRun("GET", details, nil, client)
	}
	// Parse response
	var body []byte
	if err == nil {
		body, err = io.ReadAll(resp.Body)
	}
	if err == nil {
		defer resp.Body.Close()
		err, failStep = logAndPersistAndEvaluateATCResults(utils, body, atcResultFileName, generateHTML, failOnSeverityLevel)
	}
	if err != nil {
		return errors.Errorf("Handling ATC result failed: %v", err)
	}
	if failStep {
		return errors.Errorf("Step execution failed due to at least one ATC finding with severity equal to or higher than the failOnSeverity parameter of this step (see config.yml)")
	}
	return nil
}

func triggerATCRun(config abapEnvironmentRunATCCheckOptions, details abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) (*http.Response, error) {
	bodyString, err := buildATCRequestBody(config)
	if err != nil {
		return nil, err
	}
	var resp *http.Response
	abapEndpoint := details.URL

	log.Entry().Infof("Request Body: %s", bodyString)
	body := []byte(bodyString)
	details.URL = abapEndpoint + "/sap/bc/adt/api/atc/runs?clientWait=false"
	resp, err = runATC("POST", details, body, client)
	return resp, err
}

func buildATCRequestBody(config abapEnvironmentRunATCCheckOptions) (bodyString string, err error) {
	atcConfig, err := resolveATCConfiguration(config)
	if err != nil {
		return "", err
	}

	// Create string for the run parameters
	variant := "ABAP_CLOUD_DEVELOPMENT_DEFAULT"
	if atcConfig.CheckVariant != "" {
		variant = atcConfig.CheckVariant
	}
	log.Entry().Infof("ATC Check Variant: %s", variant)
	runParameters := ` checkVariant="` + variant + `"`
	if atcConfig.Configuration != "" {
		runParameters += ` configuration="` + atcConfig.Configuration + `"`
	}

	var objectSetString string
	// check if OSL Objectset is present
	if !reflect.DeepEqual(abaputils.ObjectSet{}, atcConfig.ObjectSet) {
		objectSetString = abaputils.BuildOSLString(atcConfig.ObjectSet)
	}
	// if initial - check if ATC Object set is present
	if objectSetString == "" && (len(atcConfig.Objects.Package) != 0 || len(atcConfig.Objects.SoftwareComponent) != 0) {
		objectSetString, err = getATCObjectSet(atcConfig)
	}

	if objectSetString == "" {
		return objectSetString, errors.Errorf("Error while parsing ATC test run object set config. No object set has been provided. Please configure the objects you want to be checked for the respective test run")
	}

	bodyString = `<?xml version="1.0" encoding="UTF-8"?><atc:runparameters xmlns:atc="http://www.sap.com/adt/atc" xmlns:obj="http://www.sap.com/adt/objectset"` + runParameters + `>` + objectSetString + `</atc:runparameters>`
	return bodyString, err
}

func resolveATCConfiguration(config abapEnvironmentRunATCCheckOptions) (atcConfig ATCConfiguration, err error) {
	if config.AtcConfig != "" {
		// Configuration defaults to ATC Config
		log.Entry().Infof("ATC Configuration: %s", config.AtcConfig)
		atcConfigFile, err := abaputils.ReadConfigFile(config.AtcConfig)
		if err != nil {
			return atcConfig, err
		}
		if err := json.Unmarshal(atcConfigFile, &atcConfig); err != nil {
			log.Entry().WithError(err).Warning("failed to unmarschal json")
		}
		return atcConfig, nil

	} else if config.Repositories != "" {
		// Fallback / EasyMode is the Repositories configuration
		log.Entry().Infof("ATC Configuration derived from: %s", config.Repositories)
		repositories, err := abaputils.GetRepositories((&abaputils.RepositoriesConfig{Repositories: config.Repositories}), false)
		if err != nil {
			return atcConfig, err
		}
		for _, repository := range repositories {
			atcConfig.Objects.SoftwareComponent = append(atcConfig.Objects.SoftwareComponent, SoftwareComponent{Name: repository.Name})
		}
		return atcConfig, nil
	} else {
		// Fail if no configuration is provided
		return atcConfig, errors.New("No configuration provided - please provide either an ATC configuration file or a repository configuration file")
	}
}

func getATCObjectSet(ATCConfig ATCConfiguration) (objectSet string, err error) {
	objectSet += `<obj:objectSet>`

	// Build SC XML body
	if len(ATCConfig.Objects.SoftwareComponent) != 0 {
		objectSet += "<obj:softwarecomponents>"
		for _, s := range ATCConfig.Objects.SoftwareComponent {
			objectSet += `<obj:softwarecomponent value="` + s.Name + `"/>`
		}
		objectSet += "</obj:softwarecomponents>"
	}

	// Build Package XML body
	if len(ATCConfig.Objects.Package) != 0 {
		objectSet += "<obj:packages>"
		for _, s := range ATCConfig.Objects.Package {
			objectSet += `<obj:package value="` + s.Name + `" includeSubpackages="` + strconv.FormatBool(s.IncludeSubpackages) + `"/>`
		}
		objectSet += "</obj:packages>"
	}

	objectSet += `</obj:objectSet>`

	return objectSet, nil
}

func logAndPersistAndEvaluateATCResults(utils piperutils.FileUtils, body []byte, atcResultFileName string, generateHTML bool, failOnSeverityLevel string) (error, bool) {
	var failStep bool
	if len(body) == 0 {
		return errors.Errorf("Parsing ATC result failed: %v", errors.New("Body is empty, can't parse empty body")), failStep
	}

	responseBody := string(body)
	log.Entry().Debugf("Response body: %s", responseBody)
	if strings.HasPrefix(responseBody, "<html>") {
		return errors.New("The Software Component could not be checked. Please make sure the respective Software Component has been cloned successfully on the system"), failStep
	}

	parsedXML := new(Result)
	if err := xml.Unmarshal([]byte(body), &parsedXML); err != nil {
		log.Entry().WithError(err).Warning("failed to unmarschal xml response")
	}
	if len(parsedXML.Files) == 0 {
		log.Entry().Info("There were no results from this run, most likely the checked Software Components are empty or contain no ATC findings")
	}

	err := os.WriteFile(atcResultFileName, body, 0o644)
	if err == nil {
		log.Entry().Infof("Writing %s file was successful", atcResultFileName)
		var reports []piperutils.Path
		reports = append(reports, piperutils.Path{Target: atcResultFileName, Name: "ATC Results", Mandatory: true})
		for _, s := range parsedXML.Files {
			for _, t := range s.ATCErrors {
				log.Entry().Infof("%s in file '%s': %s in line %s found by %s", t.Severity, s.Key, t.Message, t.Line, t.Source)
				if !failStep {
					failStep = checkStepFailing(t.Severity, failOnSeverityLevel)
				}
			}
		}
		if generateHTML {
			htmlString := generateHTMLDocument(parsedXML)
			htmlStringByte := []byte(htmlString)
			atcResultHTMLFileName := strings.Trim(atcResultFileName, ".xml") + ".html"
			err = os.WriteFile(atcResultHTMLFileName, htmlStringByte, 0o644)
			if err == nil {
				log.Entry().Info("Writing " + atcResultHTMLFileName + " file was successful")
				reports = append(reports, piperutils.Path{Target: atcResultFileName, Name: "ATC Results HTML file", Mandatory: true})
			}
		}
		piperutils.PersistReportsAndLinks("abapEnvironmentRunATCCheck", "", utils, reports, nil)
	}
	if err != nil {
		return errors.Errorf("Writing results failed: %v", err), failStep
	}
	return nil, failStep
}
func checkStepFailing(severity string, failOnSeverityLevel string) bool {
	switch failOnSeverityLevel {
	case "error":
		switch severity {
		case "error":
			return true
		case "warning":
			return false
		case "info":
			return false
		default:
			return false
		}
	case "warning":
		switch severity {
		case "error":
			return true
		case "warning":
			return true
		case "info":
			return false
		default:
			return false
		}
	case "info":
		switch severity {
		case "error":
			return true
		case "warning":
			return true
		case "info":
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func runATC(requestType string, details abaputils.ConnectionDetailsHTTP, body []byte, client piperhttp.Sender) (*http.Response, error) {
	log.Entry().WithField("ABAP endpoint: ", details.URL).Info("triggering ATC run")

	header := make(map[string][]string)
	header["X-Csrf-Token"] = []string{details.XCsrfToken}
	header["Content-Type"] = []string{"application/vnd.sap.atc.run.parameters.v1+xml; charset=utf-8;"}

	resp, err := client.SendRequest(requestType, details.URL, bytes.NewBuffer(body), header, nil)
	_ = logResponseBody(resp)
	if err != nil || (resp != nil && resp.StatusCode == 400) { // send request does not seem to produce error with StatusCode 400!!!
		_, err = abaputils.HandleHTTPError(resp, err, "triggering ATC run failed with Status: "+resp.Status, details)
		log.SetErrorCategory(log.ErrorService)
		return resp, errors.Errorf("triggering ATC run failed: %v", err)
	}
	defer resp.Body.Close()
	return resp, err
}

func logResponseBody(resp *http.Response) error {
	var bodyText []byte
	var readError error
	if resp != nil {
		bodyText, readError = io.ReadAll(resp.Body)
		if readError != nil {
			return readError
		}
		log.Entry().Infof("Response body: %s", bodyText)
	}
	return nil
}

func fetchXcsrfToken(requestType string, details abaputils.ConnectionDetailsHTTP, body []byte, client piperhttp.Sender) (string, error) {
	log.Entry().WithField("ABAP Endpoint: ", details.URL).Debug("Fetching Xcrsf-Token")

	details.URL += "/sap/bc/adt/api/atc/runs/00000000000000000000000000000000"
	details.XCsrfToken = "fetch"
	header := make(map[string][]string)
	header["X-Csrf-Token"] = []string{details.XCsrfToken}
	header["Accept"] = []string{"application/vnd.sap.atc.run.v1+xml"}
	req, err := client.SendRequest(requestType, details.URL, bytes.NewBuffer(body), header, nil)
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		return "", errors.Errorf("Fetching Xcsrf-Token failed: %v", err)
	}
	defer req.Body.Close()

	token := req.Header.Get("X-Csrf-Token")
	return token, err
}

func pollATCRun(details abaputils.ConnectionDetailsHTTP, body []byte, client piperhttp.Sender) (string, error) {
	log.Entry().WithField("ABAP endpoint", details.URL).Info("Polling ATC run status")

	for {
		resp, err := getHTTPResponseATCRun("GET", details, nil, client)
		if err != nil {
			return "", errors.Errorf("Getting HTTP response failed: %v", err)
		}
		bodyText, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", errors.Errorf("Reading response body failed: %v", err)
		}

		x := new(Run)
		if err := xml.Unmarshal(bodyText, &x); err != nil {
			log.Entry().WithError(err).Warning("failed to unmarschal xml response")
		}
		log.Entry().WithField("StatusCode", resp.StatusCode).Info("Status: " + x.Status)

		if x.Status == "Not Created" {
			return "", err
		}
		if x.Status == "Completed" {
			return x.Link[0].Key, err
		}
		if x.Status == "" {
			return "", errors.Errorf("Could not get any response from ATC poll: %v", errors.New("Status from ATC run is empty. Either it's not an ABAP system or ATC run hasn't started"))
		}
		time.Sleep(5 * time.Second)
	}
}

func getHTTPResponseATCRun(requestType string, details abaputils.ConnectionDetailsHTTP, body []byte, client piperhttp.Sender) (*http.Response, error) {
	header := make(map[string][]string)
	header["Accept"] = []string{"application/vnd.sap.atc.run.v1+xml"}

	resp, err := client.SendRequest(requestType, details.URL, bytes.NewBuffer(body), header, nil)
	if err != nil {
		return resp, errors.Errorf("Getting ATC run status failed: %v", err)
	}
	return resp, err
}

func getResultATCRun(requestType string, details abaputils.ConnectionDetailsHTTP, body []byte, client piperhttp.Sender) (*http.Response, error) {
	log.Entry().WithField("ABAP Endpoint: ", details.URL).Info("Getting ATC results")

	header := make(map[string][]string)
	header["x-csrf-token"] = []string{details.XCsrfToken}
	header["Accept"] = []string{"application/vnd.sap.atc.checkstyle.v1+xml"}

	resp, err := client.SendRequest(requestType, details.URL, bytes.NewBuffer(body), header, nil)
	if err != nil {
		return resp, errors.Errorf("Getting ATC run results failed: %v", err)
	}
	return resp, err
}

func convertATCOptions(options *abapEnvironmentRunATCCheckOptions) abaputils.AbapEnvironmentOptions {
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

func generateHTMLDocument(parsedXML *Result) (htmlDocumentString string) {
	htmlDocumentString = `<!DOCTYPE html><html lang="en" xmlns="http://www.w3.org/1999/xhtml"><head><title>ATC Results</title><meta http-equiv="Content-Type" content="text/html; charset=UTF-8" /><style>table,th,td {border: 1px solid black;border-collapse:collapse;}th,td{padding: 5px;text-align:left;font-size:medium;}</style></head><body><h1 style="text-align:left;font-size:large">ATC Results</h1><table style="width:100%"><tr><th>Severity</th><th>File</th><th>Message</th><th>Line</th><th>Checked by</th></tr>`
	var htmlDocumentStringError, htmlDocumentStringWarning, htmlDocumentStringInfo, htmlDocumentStringDefault string
	for _, s := range parsedXML.Files {
		for _, t := range s.ATCErrors {
			var trBackgroundColor string
			if t.Severity == "error" {
				trBackgroundColor = "rgba(227,85,0)"
				htmlDocumentStringError += `<tr style="background-color: ` + trBackgroundColor + `">` + `<td>` + t.Severity + `</td>` + `<td>` + s.Key + `</td>` + `<td>` + t.Message + `</td>` + `<td style="text-align:center">` + t.Line + `</td>` + `<td>` + t.Source + `</td>` + `</tr>`
			}
			if t.Severity == "warning" {
				trBackgroundColor = "rgba(255,175,0, 0.75)"
				htmlDocumentStringWarning += `<tr style="background-color: ` + trBackgroundColor + `">` + `<td>` + t.Severity + `</td>` + `<td>` + s.Key + `</td>` + `<td>` + t.Message + `</td>` + `<td style="text-align:center">` + t.Line + `</td>` + `<td>` + t.Source + `</td>` + `</tr>`
			}
			if t.Severity == "info" {
				trBackgroundColor = "rgba(255,175,0, 0.2)"
				htmlDocumentStringInfo += `<tr style="background-color: ` + trBackgroundColor + `">` + `<td>` + t.Severity + `</td>` + `<td>` + s.Key + `</td>` + `<td>` + t.Message + `</td>` + `<td style="text-align:center">` + t.Line + `</td>` + `<td>` + t.Source + `</td>` + `</tr>`
			}
			if t.Severity != "info" && t.Severity != "warning" && t.Severity != "error" {
				trBackgroundColor = "rgba(255,175,0, 0)"
				htmlDocumentStringDefault += `<tr style="background-color: ` + trBackgroundColor + `">` + `<td>` + t.Severity + `</td>` + `<td>` + s.Key + `</td>` + `<td>` + t.Message + `</td>` + `<td style="text-align:center">` + t.Line + `</td>` + `<td>` + t.Source + `</td>` + `</tr>`
			}
		}
	}
	htmlDocumentString += htmlDocumentStringError + htmlDocumentStringWarning + htmlDocumentStringInfo + htmlDocumentStringDefault + `</table></body></html>`

	return htmlDocumentString
}

// ATCConfiguration object for parsing yaml config of software components and packages
type ATCConfiguration struct {
	CheckVariant  string              `json:"checkvariant,omitempty"`
	Configuration string              `json:"configuration,omitempty"`
	Objects       ATCObjects          `json:"atcobjects"`
	ObjectSet     abaputils.ObjectSet `json:"objectset,omitempty"`
}

// ATCObjects in form of packages and software components to be checked
type ATCObjects struct {
	Package           []Package           `json:"package"`
	SoftwareComponent []SoftwareComponent `json:"softwarecomponent"`
}

// Package for ATC run  to be checked
type Package struct {
	Name               string `json:"name"`
	IncludeSubpackages bool   `json:"includesubpackage"`
}

// SoftwareComponent for ATC run to be checked
type SoftwareComponent struct {
	Name string `json:"name"`
}

// Run Object for parsing XML
type Run struct {
	XMLName xml.Name `xml:"run"`
	Status  string   `xml:"status,attr"`
	Link    []Link   `xml:"link"`
}

// Link of XML object
type Link struct {
	Key   string `xml:"href,attr"`
	Value string `xml:",chardata"`
}

// Result from ATC check for all files that were checked
type Result struct {
	XMLName xml.Name `xml:"checkstyle"`
	Files   []File   `xml:"file"`
}

// File that contains ATC check with error for checked file
type File struct {
	Key       string     `xml:"name,attr"`
	Value     string     `xml:",chardata"`
	ATCErrors []ATCError `xml:"error"`
}

// ATCError with message
type ATCError struct {
	Text     string `xml:",chardata"`
	Message  string `xml:"message,attr"`
	Source   string `xml:"source,attr"`
	Line     string `xml:"line,attr"`
	Severity string `xml:"severity,attr"`
}
