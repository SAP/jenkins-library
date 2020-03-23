package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/cookiejar"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
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
		Username:  config.Username,
		Password:  config.Password,
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

	var abapEndpoint string
	details := connectionDetailsHTTP{}

	//If Host is empty read Service Key
	if err == nil {
		err = checkHost(config, details)
	}

	//Fetch Xcrsf-Token
	if err == nil {
		log.Entry().WithField("ABAP Endpoint: ", config.Host).Info("Fetching Xcrsf-Token")

		//HTTP config for fetching Xcsrf-Token
		details.URL = abapEndpoint + "/sap/bc/adt/api/atc/runs/FA163EE47BDD1EDA94E463492808E837"
		details.XCsrfToken = "fetch"

		details.XCsrfToken, err = fetchXcsrfToken("GET", details, nil, &client)
	}

	//Trigger ATC run
	var resp *http.Response
	if err == nil {
		details.URL = config.Host + "/sap/bc/adt/api/atc/runs?clientWait=false"
		var bodyString = `<?xml version="1.0" encoding="UTF-8"?><atc:runparameters xmlns:atc="http://www.sap.com/adt/atc" xmlns:obj="http://www.sap.com/adt/objectset"><obj:objectSet><obj:softwarecomponents><obj:softwarecomponent value="` + config.SoftwareComponent + `"/></obj:softwarecomponents><obj:packages><obj:package value="` + config.Package + `" includeSubpackages="false"/></obj:packages></obj:objectSet></atc:runparameters>`
		var body = []byte(bodyString)
		log.Entry().WithField("ABAP endpoint: ", config.Host).Info("Trigger ATC run")
		resp, err = runATCCheck("POST", details, body, &client)
		log.Entry().WithField("response: ", resp).Info("Trigger ATC run")
	}

	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runATCCheck(requestType string, connectionDetails connectionDetailsHTTP, body []byte, client piperhttp.Sender) (*http.Response, error) {

	header := make(map[string][]string)
	header["x-csrf-token"] = []string{connectionDetails.XCsrfToken}
	//header["Accept"] = []string{"application/vnd.sap.atc.run.v1+xml"}
	header["Content-Type"] = []string{"application/vnd.sap.atc.run.parameters.v1+xml; charset=utf-8;"}

	req, err := client.SendRequest(requestType, connectionDetails.URL, bytes.NewBuffer(body), header, nil)
	if err != nil {
		return req, fmt.Errorf("Triggering ATC run failed: %w", err)
	}
	return req, err
}

func fetchXcsrfToken(requestType string, connectionDetails connectionDetailsHTTP, body []byte, client piperhttp.Sender) (string, error) {

	header := make(map[string][]string)
	header["x-csrf-token"] = []string{connectionDetails.XCsrfToken}
	header["Accept"] = []string{"application/vnd.sap.atc.run.v1+xml"}

	req, err := client.SendRequest(requestType, connectionDetails.URL, bytes.NewBuffer(body), header, nil)
	if err != nil {
		return "", fmt.Errorf("Fetching Xcsrf-Token failed: %w", err)
	}
	defer req.Body.Close()
	token := req.Header.Get("X-Csrf-Token")
	return token, err
}

func checkHost(config abapEnvironmentRunATCCheckOptions, details connectionDetailsHTTP) error {

	var err error

	if config.Host == "" {
		cfconfig := cloudfoundry.CloudFoundryReadServiceKeyOptions{
			CfAPIEndpoint:     "https://api.cf.sap.hana.ondemand.com",
			CfOrg:             "Steampunk-2-jenkins-test",
			CfSpace:           "Test",
			Username:          "P2001217173",
			Password:          "ABAPsaas1!",
			CfServiceInstance: "ATCTest",
			CfServiceKey:      "TestKey",
		}
		var abapServiceKey cloudfoundry.ServiceKey
		abapServiceKey, err = cloudfoundry.ReadServiceKey(cfconfig, true)
		details.User = abapServiceKey.URL
		return err
	}

	return err
}
