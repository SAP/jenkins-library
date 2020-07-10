package cmd

import (
	"encoding/json"
	"io/ioutil"
	"net/http/cookiejar"
	"reflect"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func abapEnvironmentCheckoutBranch(options abapEnvironmentCheckoutBranchOptions, telemetryData *telemetry.CustomData) error {

	// Mapping for options
	subOptions := abaputils.AbapEnvironmentOptions{}

	subOptions.CfAPIEndpoint = options.CfAPIEndpoint
	subOptions.CfServiceInstance = options.CfServiceInstance
	subOptions.CfServiceKeyName = options.CfServiceKeyName
	subOptions.CfOrg = options.CfOrg
	subOptions.CfSpace = options.CfSpace
	subOptions.Host = options.Host
	subOptions.Password = options.Password
	subOptions.Username = options.Username

	var c command.ExecRunner = &command.Command{}

	// Determine the host, user and password, either via the input parameters or via a cloud foundry service key
	connectionDetails, errorGetInfo := abaputils.GetAbapCommunicationArrangementInfo(subOptions, c, "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull", false)
	if errorGetInfo != nil {
		log.Entry().WithError(errorGetInfo).Fatal("Parameters for the ABAP Connection not available")
	}

	// Configuring the HTTP Client and CookieJar
	client := piperhttp.Client{}
	cookieJar, errorCookieJar := cookiejar.New(nil)
	if errorCookieJar != nil {
		log.Entry().WithError(errorCookieJar).Fatal("Could not create a Cookie Jar")
	}
	clientOptions := piperhttp.ClientOptions{
		MaxRequestDuration: 180 * time.Second,
		CookieJar:          cookieJar,
		Username:           connectionDetails.User,
		Password:           connectionDetails.Password,
	}
	client.SetOptions(clientOptions)
	pollIntervall := 10 * time.Second

	log.Entry().Infof("Start to switch branch %v on repository %v ", options.BranchName, options.RepositoryName)
	log.Entry().Info("--------------------------------")
	log.Entry().Info("Start checkout branch: " + options.BranchName)
	log.Entry().Info("--------------------------------")

	// Triggering the Checkout of the repository into the ABAP Environment system
	uriConnectionDetails, errorTriggerCheckout := triggerCheckout(options.RepositoryName, options.BranchName, connectionDetails, &client)
	if errorTriggerCheckout != nil {
		log.Entry().WithError(errorTriggerCheckout).Fatal("Checkout of " + options.BranchName + " for software component " + options.RepositoryName + " failed on the ABAP System")
	}

	// Polling the status of the repository import on the ABAP Environment system
	status, errorPollEntity := pollEntity(options.RepositoryName, uriConnectionDetails, &client, pollIntervall)
	if errorPollEntity != nil {
		log.Entry().WithError(errorPollEntity).Fatal("Status of checkout action on repository" + options.RepositoryName + " failed on the ABAP System")
	}
	if status == "E" {
		log.Entry().Fatal("Checkout of branch " + options.BranchName + " failed on the ABAP System")
	}

	log.Entry().Info("--------------------------------")
	log.Entry().Infof("Checkout of branch %v on repository %v was successful", options.BranchName, options.RepositoryName)
	log.Entry().Info("--------------------------------")
	return nil
}

func triggerCheckout(repositoryName string, branchName string, checkoutConnectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) (abaputils.ConnectionDetailsHTTP, error) {

	uriConnectionDetails := checkoutConnectionDetails
	uriConnectionDetails.URL = ""
	checkoutConnectionDetails.XCsrfToken = "fetch"

	// Loging into the ABAP System - getting the x-csrf-token and cookies
	resp, err := getHTTPResponse("HEAD", checkoutConnectionDetails, nil, client)
	if err != nil {
		err = handleHTTPError(resp, err, "Authentication on the ABAP system failed", checkoutConnectionDetails)
		return uriConnectionDetails, err
	}
	defer resp.Body.Close()
	log.Entry().WithField("StatusCode", resp.Status).WithField("ABAP Endpoint", checkoutConnectionDetails.URL).Info("Authentication on the ABAP system successful")
	uriConnectionDetails.XCsrfToken = resp.Header.Get("X-Csrf-Token")
	checkoutConnectionDetails.XCsrfToken = uriConnectionDetails.XCsrfToken

	// Initial checks
	if repositoryName == "" {
		return uriConnectionDetails, errors.New("An empty string was passed for the parameter 'repositoryName'")
	}
	if branchName == "" {
		return uriConnectionDetails, errors.New("An empty string was passed for the parameter 'branchName'")
	}

	// the request looks like: POST/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/checkout_branch?branch_name='newBranch'&sc_name=/DMO/GIT_REPOSITORY'
	checkoutConnectionDetails.URL = checkoutConnectionDetails.URL + `/checkout_branch?branch_name='` + branchName + `'&sc_name='` + repositoryName + `'`
	jsonBody := []byte(``)

	// no JSON body needed
	resp, err = getHTTPResponse("POST", checkoutConnectionDetails, jsonBody, client)
	if err != nil {
		err = handleHTTPError(resp, err, "Could not trigger checkout of branch "+branchName, uriConnectionDetails)
		return uriConnectionDetails, err
	}
	defer resp.Body.Close()
	log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", repositoryName).WithField("branchName", branchName).Info("Triggered checkout of branch")

	// Parse Response
	var body abaputils.PullEntity
	var abapResp map[string]*json.RawMessage
	bodyText, errRead := ioutil.ReadAll(resp.Body)
	if errRead != nil {
		return uriConnectionDetails, err
	}
	json.Unmarshal(bodyText, &abapResp)
	json.Unmarshal(*abapResp["d"], &body)

	if reflect.DeepEqual(abaputils.PullEntity{}, body) {
		log.Entry().WithField("StatusCode", resp.Status).WithField("branchName", branchName).Error("Could not switch to specified branch")
		err := errors.New("Request to ABAP System failed")
		return uriConnectionDetails, err
	}

	expandLog := "?$expand=to_Execution_log,to_Transport_log"
	uriConnectionDetails.URL = body.Metadata.URI + expandLog
	return uriConnectionDetails, nil
}
