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
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/ghodss/yaml"
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

	//Cloud Foundry Authentication
	cfloginconfig := cloudfoundry.CloudFoundryLoginOptions{
		CfAPIEndpoint: config.CfAPIEndpoint,
		CfOrg:         config.CfOrg,
		CfSpace:       config.CfSpace,
		Username:      config.Username,
		Password:      config.Password,
	}
	err = cloudfoundry.Login(cfloginconfig)

	var details connectionDetailsHTTP
	var abapEndpoint string

	//If Host flag is empty read Service Key instead
	if err == nil {
		details, err = checkHost(config, details)
	}

	//Fetch Xcrsf-Token
	if err == nil {
		//HTTP config for fetching Xcsrf-Token
		abapEndpoint = details.URL
		details.URL += "/sap/bc/adt/api/atc/runs/00000000000000000000000000000000"
		details.XCsrfToken = "fetch"

		credentialsOptions := piperhttp.ClientOptions{
			Username:  details.User,
			Password:  details.Password,
			CookieJar: cookieJar,
		}

		client.SetOptions(credentialsOptions)

		details.XCsrfToken, err = fetchXcsrfToken("GET", details, nil, &client)
	}

	//Parse YAML ATC run configuration as body for ATC run trigger
	filelocation, _ := filepath.Glob(config.ATCRunConfig)
	fmt.Println("File location: ", filelocation)
	filename, _ := filepath.Abs(filelocation[0])
	yamlFile, err := ioutil.ReadFile(filename)
	var ATCRunConfig ATCconfig
	if err == nil {
		var result []byte
		result, err = yaml.YAMLToJSON(yamlFile)
		json.Unmarshal(result, &ATCRunConfig)
	}

	var packageString = ""
	var softwareComponentString string
	if err == nil {
		if len(ATCRunConfig.Objects.Package) == 0 || len(ATCRunConfig.Objects.SoftwareComponent) == 0 {
			err = fmt.Errorf("Error while parsing ATC run config. Please provide both the packages and the software components to be checked!")
		}

		//Build Package XML body
		packageString += "<obj:packages>"
		for _, s := range ATCRunConfig.Objects.Package {
			packageString += `<obj:package value="` + s.Name + `" includeSubpackages="` + strconv.FormatBool(s.IncludeSubpackages) + `"/>`
		}
		packageString += "</obj:packages>"

		//Build SC XML body
		softwareComponentString += "<obj:softwarecomponents>"
		for _, s := range ATCRunConfig.Objects.SoftwareComponent {
			softwareComponentString += `<obj:softwarecomponent value="` + s.Name + `"/>`
		}
		softwareComponentString += "</obj:softwarecomponents>"
	}

	//Trigger ATC run
	var resp *http.Response
	var bodyString = `<?xml version="1.0" encoding="UTF-8"?><atc:runparameters xmlns:atc="http://www.sap.com/adt/atc" xmlns:obj="http://www.sap.com/adt/objectset"><obj:objectSet>` + softwareComponentString + packageString + `</obj:objectSet></atc:runparameters>`
	var body = []byte(bodyString)

	if err == nil {
		details.URL = abapEndpoint + "/sap/bc/adt/api/atc/runs?clientWait=false"
		resp, err = runATC("POST", details, body, &client)
	}

	var location string
	//Poll ATC run
	if err == nil {
		location = resp.Header.Get("Location")
		details.URL = abapEndpoint + location
		location, err = pollATCRun(details, body, &client, config)
	}

	if err == nil {
		details.URL = abapEndpoint + location
		resp, err = getResultATCRun("GET", details, nil, &client)
	}

	//Parse response
	if err == nil {
		body, err = ioutil.ReadAll(resp.Body)
	}
	if err == nil {
		defer resp.Body.Close()
		parsedXML := new(Result)
		xml.Unmarshal([]byte(body), &parsedXML)
		err = ioutil.WriteFile("result.xml", body, 0644)
		if err == nil {
			for _, s := range parsedXML.Files {
				for _, t := range s.ATCErrors {
					log.Entry().Error("Error in file " + s.Key + ": " + t.Key)
				}
			}
		}
	}

	if err != nil {
		cloudfoundry.Logout()
		log.Entry().WithError(err).Fatal("step execution failed")
	}

	log.Entry().Info("ATC run completed succesfully. The respective run results are listes below.")
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
		cfconfig := cloudfoundry.CloudFoundryReadServiceKeyOptions{
			CfAPIEndpoint:     config.CfAPIEndpoint,
			CfOrg:             config.CfOrg,
			CfSpace:           config.CfSpace,
			Username:          config.Username,
			Password:          config.Password,
			CfServiceInstance: config.CfServiceInstance,
			CfServiceKey:      config.CfServiceKeyName,
		}
		var abapServiceKey cloudfoundry.ServiceKey
		abapServiceKey, err = cloudfoundry.ReadServiceKey(cfconfig, false)
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

func pollATCRun(details connectionDetailsHTTP, body []byte, client piperhttp.Sender, config abapEnvironmentRunATCCheckOptions) (string, error) {

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
		if x.Status != "Running" && x.Status != "Not Yet Started" {
			return x.Link[0].Key, err
		}
		time.Sleep(5 * time.Second)
	}
}

func getHTTPResponseATCRun(requestType string, details connectionDetailsHTTP, body []byte, client piperhttp.Sender) (*http.Response, error) {

	log.Entry().WithField("ABAP Endpoint: ", details.URL).Info("Getting ATC results")

	header := make(map[string][]string)
	header["Accept"] = []string{"application/vnd.sap.atc.run.v1+xml"}

	req, err := client.SendRequest(requestType, details.URL, bytes.NewBuffer(body), header, nil)
	if err != nil {
		return req, fmt.Errorf("Getting HTTP response failed: %w", err)
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
		return req, fmt.Errorf("Getting HTTP response failed: %w", err)
	}
	return req, err
}

type ATCconfig struct {
	Objects ATCObjects `json:"atcobjects"`
}

type ATCObjects struct {
	Package           []Package           `json:"package"`
	SoftwareComponent []SoftwareComponent `json:"softwarecomponent"`
}

type Package struct {
	Name               string `json:"name"`
	IncludeSubpackages bool   `json:"includesubpackage"`
}

type SoftwareComponent struct {
	Name string `json:"name"`
}

type Run struct {
	XMLName xml.Name `xml:"run"`
	Status  string   `xml:"status,attr"`
	Link    []Link   `xml:"link"`
}

type Link struct {
	Key   string `xml:"href,attr"`
	Value string `xml:",chardata"`
}

type Result struct {
	XMLName xml.Name `xml:"checkstyle"`
	Files   []File   `xml:"file"`
}

type File struct {
	Key       string     `xml:"name,attr"`
	Value     string     `xml:",chardata"`
	ATCErrors []ATCError `xml:"error"`
}

type ATCError struct {
	Key   string `xml:"message,attr"`
	Value string `xml:",chardata"`
}
