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
	"strconv"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/gcs"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

func abapEnvironmentRunATCCheck(options abapEnvironmentRunATCCheckOptions, telemetryData *telemetry.CustomData) {

	// Mapping for options
	subOptions := convertATCOptions(&options)

	c := &command.Command{}
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())

	var autils = abaputils.AbapUtils{
		Exec: c,
	}
	var err error

	client := piperhttp.Client{}
	cookieJar, _ := cookiejar.New(nil)
	clientOptions := piperhttp.ClientOptions{
		CookieJar: cookieJar,
	}
	client.SetOptions(clientOptions)

	var details abaputils.ConnectionDetailsHTTP
	//If Host flag is empty read ABAP endpoint from Service Key instead. Otherwise take ABAP system endpoint from config instead
	if err == nil {
		details, err = autils.GetAbapCommunicationArrangementInfo(subOptions, "")
	}
	var resp *http.Response
	//Fetch Xcrsf-Token
	if err == nil {
		credentialsOptions := piperhttp.ClientOptions{
			Username:  details.User,
			Password:  details.Password,
			CookieJar: cookieJar,
		}
		client.SetOptions(credentialsOptions)
		details.XCsrfToken, err = fetchXcsrfToken("GET", details, nil, &client)
	}
	if err == nil {
		resp, err = triggerATCrun(options, details, &client)
	}
	var gcsClient gcs.ClientInterface
	if GeneralConfig.UploadReportsToGCS {
		if GeneralConfig.CorrelationID == "" {
			log.Entry().WithError(errors.New("CorrelationID is used as GCS bucketID and mustn't be empty")).Fatal("Execution failed")
		}
		var err error
		if gcsClient, err = gcs.NewClient(GeneralConfig.GCPJsonKeyFilePath); err != nil {
			log.Entry().WithError(err).Fatal("Execution failed")
		}
	}
	if err == nil {
		err = handleATCresults(resp, details, &client, options.AtcResultsFileName, options.GenerateHTML, gcsClient)
	}
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}

	log.Entry().Info("ATC run completed successfully. If there are any results from the respective run they will be listed in the logs above as well as being saved in the output .xml file")
}

func handleATCresults(resp *http.Response, details abaputils.ConnectionDetailsHTTP, client piperhttp.Sender, atcResultFileName string, generateHTML bool, gcsClient gcs.ClientInterface) error {
	var err error
	var abapEndpoint string
	abapEndpoint = details.URL
	location := resp.Header.Get("Location")
	details.URL = abapEndpoint + location
	location, err = pollATCRun(details, nil, client)
	if err == nil {
		details.URL = abapEndpoint + location
		resp, err = getResultATCRun("GET", details, nil, client)
	}
	//Parse response
	var body []byte
	if err == nil {
		body, err = ioutil.ReadAll(resp.Body)
	}
	if err == nil {
		defer resp.Body.Close()
		err = parseATCResult(body, atcResultFileName, generateHTML, gcsClient)
	}
	if err != nil {
		return fmt.Errorf("Handling ATC result failed: %w", err)
	}
	return nil
}

func triggerATCrun(config abapEnvironmentRunATCCheckOptions, details abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) (*http.Response, error) {
	var atcConfigyamlFile []byte
	abapEndpoint := details.URL
	filelocation, err := filepath.Glob(config.AtcConfig)
	//Parse YAML ATC run configuration as body for ATC run trigger
	if err == nil {
		filename, err := filepath.Abs(filelocation[0])
		if err == nil {
			atcConfigyamlFile, err = ioutil.ReadFile(filename)
		}
	}
	var ATCConfig ATCconfig
	if err == nil {
		var result []byte
		result, err = yaml.YAMLToJSON(atcConfigyamlFile)
		json.Unmarshal(result, &ATCConfig)
	}
	var checkVariantString, packageString, softwareComponentString string
	if err == nil {
		checkVariantString, packageString, softwareComponentString, err = buildATCCheckBody(ATCConfig)
	}

	//Trigger ATC run
	var resp *http.Response
	var bodyString = `<?xml version="1.0" encoding="UTF-8"?><atc:runparameters xmlns:atc="http://www.sap.com/adt/atc" xmlns:obj="http://www.sap.com/adt/objectset"` + checkVariantString + `><obj:objectSet>` + softwareComponentString + packageString + `</obj:objectSet></atc:runparameters>`
	log.Entry().Debugf("Request Body: %s", bodyString)
	var body = []byte(bodyString)
	if err == nil {
		details.URL = abapEndpoint + "/sap/bc/adt/api/atc/runs?clientWait=false"
		resp, err = runATC("POST", details, body, client)
	}
	if err != nil {
		return resp, fmt.Errorf("Triggering ATC run failed: %w", err)
	}
	return resp, nil
}

func buildATCCheckBody(ATCConfig ATCconfig) (checkVariantString string, packageString string, softwareComponentString string, err error) {
	if len(ATCConfig.Objects.Package) == 0 && len(ATCConfig.Objects.SoftwareComponent) == 0 {
		return "", "", "", fmt.Errorf("Error while parsing ATC run config. Please provide the packages and/or the software components to be checked! %w", errors.New("No Package or Software Component specified. Please provide either one or both of them"))
	}

	//Build Check Variant and Configuration XML Body
	if len(ATCConfig.CheckVariant) != 0 {
		checkVariantString += ` checkVariant="` + ATCConfig.CheckVariant + `"`
		log.Entry().Infof("ATC Check Variant: %s", ATCConfig.CheckVariant)
		if len(ATCConfig.Configuration) != 0 {
			checkVariantString += ` configuration="` + ATCConfig.Configuration + `"`
		}
	} else {
		const defaultCheckVariant = "SAP_CLOUD_PLATFORM_ATC_DEFAULT"
		checkVariantString += ` checkVariant="` + defaultCheckVariant + `"`
		log.Entry().Infof("ATC Check Variant: %s", checkVariantString)
	}

	//Build Package XML body
	if len(ATCConfig.Objects.Package) != 0 {
		packageString += "<obj:packages>"
		for _, s := range ATCConfig.Objects.Package {
			packageString += `<obj:package value="` + s.Name + `" includeSubpackages="` + strconv.FormatBool(s.IncludeSubpackages) + `"/>`
		}
		packageString += "</obj:packages>"
	}

	//Build SC XML body
	if len(ATCConfig.Objects.SoftwareComponent) != 0 {
		softwareComponentString += "<obj:softwarecomponents>"
		for _, s := range ATCConfig.Objects.SoftwareComponent {
			softwareComponentString += `<obj:softwarecomponent value="` + s.Name + `"/>`
		}
		softwareComponentString += "</obj:softwarecomponents>"
	}
	return checkVariantString, packageString, softwareComponentString, nil
}

func parseATCResult(body []byte, atcResultFileName string, generateHTML bool, gcsClient gcs.ClientInterface) (err error) {
	if len(body) == 0 {
		return fmt.Errorf("Parsing ATC result failed: %w", errors.New("Body is empty, can't parse empty body"))
	}

	responseBody := string(body)
	log.Entry().Debugf("Response body: %s", responseBody)
	if strings.HasPrefix(responseBody, "<html>") {
		return errors.New("The Software Component could not be checked. Please make sure the respective Software Component has been cloned successfully on the system")
	}

	parsedXML := new(Result)
	xml.Unmarshal([]byte(body), &parsedXML)
	if len(parsedXML.Files) == 0 {
		log.Entry().Info("There were no results from this run, most likely the checked Software Components are empty or contain no ATC findings")
	}

	err = ioutil.WriteFile(atcResultFileName, body, 0644)
	if err == nil {
		log.Entry().Infof("Writing %s file was successful", atcResultFileName)
		var reports []piperutils.Path
		reports = append(reports, piperutils.Path{Target: atcResultFileName, Name: "ATC Results", Mandatory: true})
		for _, s := range parsedXML.Files {
			for _, t := range s.ATCErrors {
				log.Entry().Infof("%s in file '%s': %s in line %s found by %s", t.Severity, s.Key, t.Message, t.Line, t.Source)
			}
		}
		if generateHTML == true {
			htmlString := generateHTMLDocument(parsedXML)
			htmlStringByte := []byte(htmlString)
			atcResultHTMLFileName := strings.Trim(atcResultFileName, ".xml") + ".html"
			err = ioutil.WriteFile(atcResultHTMLFileName, htmlStringByte, 0644)
			if err == nil {
				log.Entry().Info("Writing " + atcResultHTMLFileName + " file was successful")
				reports = append(reports, piperutils.Path{Target: atcResultFileName, Name: "ATC Results HTML file", Mandatory: true})
			}
		}
		piperutils.PersistReportsAndLinks("abapEnvironmentRunATCCheck", "", reports, nil, gcsClient, GeneralConfig.CorrelationID)
	}
	if err != nil {
		return fmt.Errorf("Writing results failed: %w", err)
	}
	return nil
}

func runATC(requestType string, details abaputils.ConnectionDetailsHTTP, body []byte, client piperhttp.Sender) (*http.Response, error) {

	log.Entry().WithField("ABAP endpoint: ", details.URL).Info("Triggering ATC run")

	header := make(map[string][]string)
	header["X-Csrf-Token"] = []string{details.XCsrfToken}
	header["Content-Type"] = []string{"application/vnd.sap.atc.run.parameters.v1+xml; charset=utf-8;"}

	req, err := client.SendRequest(requestType, details.URL, bytes.NewBuffer(body), header, nil)
	if err != nil {
		return req, fmt.Errorf("Triggering ATC run failed: %w", err)
	}
	defer req.Body.Close()
	return req, err
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
		return "", fmt.Errorf("Fetching Xcsrf-Token failed: %w", err)
	}
	defer req.Body.Close()

	// workaround until golang version 1.16 is used
	time.Sleep(1 * time.Second)

	token := req.Header.Get("X-Csrf-Token")
	return token, err
}

func pollATCRun(details abaputils.ConnectionDetailsHTTP, body []byte, client piperhttp.Sender) (string, error) {

	log.Entry().WithField("ABAP endpoint", details.URL).Info("Polling ATC run status")

	for {
		resp, err := getHTTPResponseATCRun("GET", details, nil, client)
		if err != nil {
			return "", fmt.Errorf("Getting HTTP response failed: %w", err)
		}
		bodyText, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("Reading response body failed: %w", err)
		}

		x := new(Run)
		xml.Unmarshal(bodyText, &x)
		log.Entry().WithField("StatusCode", resp.StatusCode).Info("Status: " + x.Status)

		if x.Status == "Not Created" {
			return "", err
		}
		if x.Status == "Completed" {
			return x.Link[0].Key, err
		}
		if x.Status == "" {
			return "", fmt.Errorf("Could not get any response from ATC poll: %w", errors.New("Status from ATC run is empty. Either it's not an ABAP system or ATC run hasn't started"))
		}
		time.Sleep(5 * time.Second)
	}
}

func getHTTPResponseATCRun(requestType string, details abaputils.ConnectionDetailsHTTP, body []byte, client piperhttp.Sender) (*http.Response, error) {

	log.Entry().WithField("ABAP Endpoint: ", details.URL).Info("Polling ATC run status")

	header := make(map[string][]string)
	header["Accept"] = []string{"application/vnd.sap.atc.run.v1+xml"}

	req, err := client.SendRequest(requestType, details.URL, bytes.NewBuffer(body), header, nil)
	if err != nil {
		return req, fmt.Errorf("Getting ATC run status failed: %w", err)
	}
	return req, err
}

func getResultATCRun(requestType string, details abaputils.ConnectionDetailsHTTP, body []byte, client piperhttp.Sender) (*http.Response, error) {

	log.Entry().WithField("ABAP Endpoint: ", details.URL).Info("Getting ATC results")

	header := make(map[string][]string)
	header["x-csrf-token"] = []string{details.XCsrfToken}
	header["Accept"] = []string{"application/vnd.sap.atc.checkstyle.v1+xml"}

	req, err := client.SendRequest(requestType, details.URL, bytes.NewBuffer(body), header, nil)
	if err != nil {
		return req, fmt.Errorf("Getting ATC run results failed: %w", err)
	}
	return req, err
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

//ATCconfig object for parsing yaml config of software components and packages
type ATCconfig struct {
	CheckVariant  string     `json:"checkvariant,omitempty"`
	Configuration string     `json:"configuration,omitempty"`
	Objects       ATCObjects `json:"atcobjects"`
}

//ATCObjects in form of packages and software components to be checked
type ATCObjects struct {
	Package           []Package           `json:"package"`
	SoftwareComponent []SoftwareComponent `json:"softwarecomponent"`
}

//Package for ATC run  to be checked
type Package struct {
	Name               string `json:"name"`
	IncludeSubpackages bool   `json:"includesubpackage"`
}

//SoftwareComponent for ATC run to be checked
type SoftwareComponent struct {
	Name string `json:"name"`
}

//Run Object for parsing XML
type Run struct {
	XMLName xml.Name `xml:"run"`
	Status  string   `xml:"status,attr"`
	Link    []Link   `xml:"link"`
}

//Link of XML object
type Link struct {
	Key   string `xml:"href,attr"`
	Value string `xml:",chardata"`
}

//Result from ATC check for all files that were checked
type Result struct {
	XMLName xml.Name `xml:"checkstyle"`
	Files   []File   `xml:"file"`
}

//File that contains ATC check with error for checked file
type File struct {
	Key       string     `xml:"name,attr"`
	Value     string     `xml:",chardata"`
	ATCErrors []ATCError `xml:"error"`
}

//ATCError with message
type ATCError struct {
	Text     string `xml:",chardata"`
	Message  string `xml:"message,attr"`
	Source   string `xml:"source,attr"`
	Line     string `xml:"line,attr"`
	Severity string `xml:"severity,attr"`
}
