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
	subOptions.Host = options.Host
	subOptions.Password = options.Password
	subOptions.Username = options.Username

	var c command.ExecRunner = &command.Command{}

	// Determine the host, user and password, either via the input parameters or via a cloud foundry service key
	connectionDetails, errorGetInfo := abaputils.GetAbapCommunicationArrangementInfo(subOptions, c, "", false)
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
	//pollIntervall := 10 * time.Second

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
	status, errorPollEntity := pollEntity(repositoryName, uriConnectionDetails, &client, pollIntervall)
	if errorPollEntity != nil {
		log.Entry().WithError(errorPollEntity).Fatal("Pull of " + repositoryName + " failed on the ABAP System")
	}
	if status == "E" {
		log.Entry().Fatal("Pull of " + repositoryName + " failed on the ABAP System")
	}

	log.Entry().Info(repositoryName + " was pulled successfully")

	log.Entry().Info("-------------------------")
	log.Entry().Info("All repositories were pulled successfully")
	return nil
}

func triggerCheckout(repositoryName string, targetBranchName string, checkoutConnectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) (abaputils.ConnectionDetailsHTTP, error) {

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
	log.Entry().WithField("StatusCode", resp.Status).WithField("ABAP Endpoint", checkoutConnectionDetails.URL).Info("Authentication on the ABAP system successfull")
	uriConnectionDetails.XCsrfToken = resp.Header.Get("X-Csrf-Token")
	checkoutConnectionDetails.XCsrfToken = uriConnectionDetails.XCsrfToken

	// Trigger the branch checkout
	if repositoryName == "" {
		return uriConnectionDetails, errors.New("An empty string was passed for the parameter 'repositoryName'")
	}
	if targetBranchName == "" {
		return uriConnectionDetails, errors.New("An empty string was passed for the parameter 'branchName'")
	}
	jsonBody := []byte(`{"sc_name":"` + repositoryName + `"}`)
	resp, err = getHTTPResponse("POST", checkoutConnectionDetails, jsonBody, client)
	if err != nil {
		err = handleHTTPError(resp, err, "Could not pull the Repository / Software Component "+repositoryName, uriConnectionDetails)
		return uriConnectionDetails, err
	}
	defer resp.Body.Close()
	log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", repositoryName).Info("Triggered Pull of Repository / Software Component")

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
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", repositoryName).Error("Could not pull the Repository / Software Component")
		err := errors.New("Request to ABAP System not successful")
		return uriConnectionDetails, err
	}

	expandLog := "?$expand=to_Execution_log,to_Transport_log"
	uriConnectionDetails.URL = body.Metadata.URI + expandLog
	return uriConnectionDetails, nil
}
