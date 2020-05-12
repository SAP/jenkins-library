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
	"time"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

func abapEnvironmentRunATCCheck(config abapEnvironmentRunATCCheckOptions, telemetryData *telemetry.CustomData) {

	var c = command.Command{}

	var err error

	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())

	client := piperhttp.Client{}
	cookieJar, _ := cookiejar.New(nil)
	clientOptions := piperhttp.ClientOptions{
		CookieJar: cookieJar,
	}
	client.SetOptions(clientOptions)

	var details connectionDetailsHTTP
	var abapEndpoint string
	//If Host flag is empty read ABAP endpoint from Service Key instead. Otherwise take ABAP system endpoint from config instead
	if err == nil {
		details, err = checkHost(config, details)
	}
	var resp *http.Response
	//Fetch Xcrsf-Token
	if err == nil {
		abapEndpoint = details.URL
		credentialsOptions := piperhttp.ClientOptions{
			Username:  details.User,
			Password:  details.Password,
			CookieJar: cookieJar,
		}
		client.SetOptions(credentialsOptions)
		details.XCsrfToken, err = fetchXcsrfToken("GET", details, nil, &client)
	}
	if err == nil {
		resp, err = triggerATCrun(config, details, &client, abapEndpoint)
	}
	if err == nil {
		err = handleATCresults(resp, details, &client, abapEndpoint)
	}
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}

	log.Entry().Info("ATC run completed succesfully. The respective run results are listed above.")
}

func handleATCresults(resp *http.Response, details connectionDetailsHTTP, client piperhttp.Sender, abapEndpoint string) error {
	var err error
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
		err = parseATCResult(body)
	}
	if err != nil {
		return fmt.Errorf("Handling ATC result failed: %w", err)
	}
	return nil
}

func triggerATCrun(config abapEnvironmentRunATCCheckOptions, details connectionDetailsHTTP, client piperhttp.Sender, abapEndpoint string) (*http.Response, error) {
	var atcConfigyamlFile []byte
	filelocation, err := filepath.Glob(config.AtcConfig)
	//Parse YAML ATC run configuration as body for ATC run trigger
	if err == nil {
		filename, _ := filepath.Abs(filelocation[0])
		atcConfigyamlFile, err = ioutil.ReadFile(filename)
	}
	var ATCConfig ATCconfig
	if err == nil {
		var result []byte
		result, err = yaml.YAMLToJSON(atcConfigyamlFile)
		json.Unmarshal(result, &ATCConfig)
	}
	var packageString string
	var softwareComponentString string
	if err == nil {
		packageString, softwareComponentString, err = buildATCCheckBody(ATCConfig)
	}

	//Trigger ATC run
	var resp *http.Response
	var bodyString = `<?xml version="1.0" encoding="UTF-8"?><atc:runparameters xmlns:atc="http://www.sap.com/adt/atc" xmlns:obj="http://www.sap.com/adt/objectset"><obj:objectSet>` + softwareComponentString + packageString + `</obj:objectSet></atc:runparameters>`
	var body = []byte(bodyString)
	if err == nil {
		details.URL = abapEndpoint + "/sap/bc/adt/api/atc/runs?clientWait=false"
		resp, err = runATC("POST", details, body, client)
	}
	if err != nil {
		return resp, fmt.Errorf("Triggering ATC result failed: %w", err)
	}
	return resp, nil
}

func buildATCCheckBody(ATCConfig ATCconfig) (string, string, error) {
	if len(ATCConfig.Objects.Package) == 0 && len(ATCConfig.Objects.SoftwareComponent) == 0 {
		return "", "", fmt.Errorf("Error while parsing ATC run config. Please provide the packages and/or the software components to be checked! %w", errors.New("No Package or Software Component specified. Please provide either one or both of them"))
	}

	var packageString string
	var softwareComponentString string

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
	return packageString, softwareComponentString, nil
}

func parseATCResult(body []byte) error {
	if len(body) == 0 {
		return fmt.Errorf("Parsing ATC result failed: %w", errors.New("Body is empty, can't parse empty body"))
	}
	parsedXML := new(Result)
	xml.Unmarshal([]byte(body), &parsedXML)
	err := ioutil.WriteFile("ATCResults.xml", body, 0644)
	if err == nil {
		var reports []piperutils.Path
		reports = append(reports, piperutils.Path{Target: "ATCResults.xml", Name: "ATC Results", Mandatory: true})
		piperutils.PersistReportsAndLinks("abapEnvironmentRunATCCheck", "", reports, nil)
		for _, s := range parsedXML.Files {
			for _, t := range s.ATCErrors {
				log.Entry().Error("Error in file " + s.Key + ": " + t.Key)
			}
		}
	}
	if err != nil {
		return fmt.Errorf("Writing results to XML file failed: %w", err)
	}
	return nil
}

func runATC(requestType string, details connectionDetailsHTTP, body []byte, client piperhttp.Sender) (*http.Response, error) {

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

func fetchXcsrfToken(requestType string, details connectionDetailsHTTP, body []byte, client piperhttp.Sender) (string, error) {

	log.Entry().WithField("ABAP Endpoint: ", details.URL).Info("Fetching Xcrsf-Token")

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
	token := req.Header.Get("X-Csrf-Token")
	return token, err
}

func checkHost(config abapEnvironmentRunATCCheckOptions, details connectionDetailsHTTP) (connectionDetailsHTTP, error) {

	var err error

	if config.Host == "" {
		cfconfig := cloudfoundry.ServiceKeyOptions{
			CfAPIEndpoint:     config.CfAPIEndpoint,
			CfOrg:             config.CfOrg,
			CfSpace:           config.CfSpace,
			Username:          config.Username,
			Password:          config.Password,
			CfServiceInstance: config.CfServiceInstance,
			CfServiceKey:      config.CfServiceKeyName,
		}
		if cfconfig.CfServiceInstance == "" || cfconfig.CfOrg == "" || cfconfig.CfAPIEndpoint == "" || cfconfig.CfSpace == "" || cfconfig.CfServiceKey == "" {
			return details, errors.New("Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510")
		}
		var abapServiceKey cloudfoundry.ServiceKey
		abapServiceKey, err = cloudfoundry.ReadServiceKeyAbapEnvironment(cfconfig, true)
		if err != nil {
			return details, fmt.Errorf("Reading Service Key failed: %w", err)
		}
		details.User = abapServiceKey.Abap.Username
		details.Password = abapServiceKey.Abap.Password
		details.URL = abapServiceKey.URL
		return details, err
	}
	details.User = config.Username
	details.Password = config.Password
	details.URL = config.Host
	return details, err
}

func pollATCRun(details connectionDetailsHTTP, body []byte, client piperhttp.Sender) (string, error) {

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

func getHTTPResponseATCRun(requestType string, details connectionDetailsHTTP, body []byte, client piperhttp.Sender) (*http.Response, error) {

	log.Entry().WithField("ABAP Endpoint: ", details.URL).Info("Polling ATC run status")

	header := make(map[string][]string)
	header["Accept"] = []string{"application/vnd.sap.atc.run.v1+xml"}

	req, err := client.SendRequest(requestType, details.URL, bytes.NewBuffer(body), header, nil)
	if err != nil {
		return req, fmt.Errorf("Getting ATC run status failed: %w", err)
	}
	return req, err
}

func getResultATCRun(requestType string, details connectionDetailsHTTP, body []byte, client piperhttp.Sender) (*http.Response, error) {

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

//ATCconfig object for parsing yaml config of software components and packages
type ATCconfig struct {
	Objects ATCObjects `json:"atcobjects"`
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
	Key   string `xml:"message,attr"`
	Value string `xml:",chardata"`
}
