package cmd

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func abapEnvironmentPullGitRepo(options abapEnvironmentPullGitRepoOptions, telemetryData *telemetry.CustomData) {

	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	var autils = abaputils.AbapUtils{
		Exec: &c,
	}

	client := piperhttp.Client{}

	// for http calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapEnvironmentPullGitRepo(&options, telemetryData, &autils, &client)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentPullGitRepo(options *abapEnvironmentPullGitRepoOptions, telemetryData *telemetry.CustomData, com abaputils.Communication, client piperhttp.Sender) error {

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

	// Determine the host, user and password, either via the input parameters or via a cloud foundry service key
	connectionDetails, errorGetInfo := com.GetAbapCommunicationArrangementInfo(subOptions, "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull")
	if errorGetInfo != nil {
		return errors.Wrap(errorGetInfo, "Parameters for the ABAP Connection not available")
	}

	// Configuring the HTTP Client and CookieJar

	cookieJar, errorCookieJar := cookiejar.New(nil)
	if errorCookieJar != nil {
		return errors.Wrap(errorCookieJar, "Could not create a Cookie Jar")
	}
	clientOptions := piperhttp.ClientOptions{
		MaxRequestDuration: 180 * time.Second,
		CookieJar:          cookieJar,
		Username:           connectionDetails.User,
		Password:           connectionDetails.Password,
	}
	client.SetOptions(clientOptions)
	pollIntervall := com.GetPollIntervall()

	log.Entry().Infof("Start pulling %v repositories", len(options.RepositoryNames))
	for _, repositoryName := range options.RepositoryNames {

		log.Entry().Info("-------------------------")
		log.Entry().Info("Start pulling " + repositoryName)
		log.Entry().Info("-------------------------")

		// Triggering the Pull of the repository into the ABAP Environment system
		uriConnectionDetails, errorTriggerPull := triggerPull(repositoryName, connectionDetails, client)
		if errorTriggerPull != nil {
			return errors.Wrapf(errorTriggerPull, "Pull of '%s' failed on the ABAP System", repositoryName)

		}

		// Polling the status of the repository import on the ABAP Environment system
		status, errorPollEntity := pollEntity(repositoryName, uriConnectionDetails, client, pollIntervall)
		if errorPollEntity != nil {
			return errors.Wrapf(errorPollEntity, "Pull of '%s' failed on the ABAP System", repositoryName)
		}
		if status == "E" {
			return errors.New("Pull of " + repositoryName + " failed on the ABAP System")
		}

		log.Entry().Info(repositoryName + " was pulled successfully")
	}
	log.Entry().Info("-------------------------")
	log.Entry().Info("All repositories were pulled successfully")
	return nil
}

func triggerPull(repositoryName string, pullConnectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) (abaputils.ConnectionDetailsHTTP, error) {

	uriConnectionDetails := pullConnectionDetails
	uriConnectionDetails.URL = ""
	pullConnectionDetails.XCsrfToken = "fetch"

	// Loging into the ABAP System - getting the x-csrf-token and cookies
	resp, err := getHTTPResponse("HEAD", pullConnectionDetails, nil, client)
	if err != nil {
		err = handleHTTPError(resp, err, "Authentication on the ABAP system failed", pullConnectionDetails)
		return uriConnectionDetails, err
	}
	defer resp.Body.Close()
	log.Entry().WithField("StatusCode", resp.Status).WithField("ABAP Endpoint", pullConnectionDetails.URL).Info("Authentication on the ABAP system successfull")
	uriConnectionDetails.XCsrfToken = resp.Header.Get("X-Csrf-Token")
	pullConnectionDetails.XCsrfToken = uriConnectionDetails.XCsrfToken

	// Trigger the Pull of a Repository
	if repositoryName == "" {
		return uriConnectionDetails, errors.New("An empty string was passed for the parameter 'repositoryName'")
	}
	jsonBody := []byte(`{"sc_name":"` + repositoryName + `"}`)
	resp, err = getHTTPResponse("POST", pullConnectionDetails, jsonBody, client)
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

func pollEntity(repositoryName string, connectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender, pollIntervall time.Duration) (string, error) {

	log.Entry().Info("Start polling the status...")
	var status string = "R"

	for {
		var resp, err = getHTTPResponse("GET", connectionDetails, nil, client)
		if err != nil {
			err = handleHTTPError(resp, err, "Could not pull the Repository / Software Component "+repositoryName, connectionDetails)
			return "", err
		}
		defer resp.Body.Close()

		// Parse response
		var body abaputils.PullEntity
		bodyText, _ := ioutil.ReadAll(resp.Body)
		var abapResp map[string]*json.RawMessage
		json.Unmarshal(bodyText, &abapResp)
		json.Unmarshal(*abapResp["d"], &body)
		if reflect.DeepEqual(abaputils.PullEntity{}, body) {
			log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", repositoryName).Error("Could not pull the Repository / Software Component")
			var err = errors.New("Request to ABAP System not successful")
			return "", err
		}
		status = body.Status
		log.Entry().WithField("StatusCode", resp.Status).Info("Pull Status: " + body.StatusDescription)
		if body.Status != "R" {
			printLogs(body)
			break
		}
		time.Sleep(pollIntervall)
	}

	return status, nil
}

func getHTTPResponse(requestType string, connectionDetails abaputils.ConnectionDetailsHTTP, body []byte, client piperhttp.Sender) (*http.Response, error) {

	header := make(map[string][]string)
	header["Content-Type"] = []string{"application/json"}
	header["Accept"] = []string{"application/json"}
	header["x-csrf-token"] = []string{connectionDetails.XCsrfToken}

	req, err := client.SendRequest(requestType, connectionDetails.URL, bytes.NewBuffer(body), header, nil)
	return req, err
}

func handleHTTPError(resp *http.Response, err error, message string, connectionDetails abaputils.ConnectionDetailsHTTP) error {
	if resp == nil {
		// Response is nil in case of a timeout
		log.Entry().WithError(err).WithField("ABAP Endpoint", connectionDetails.URL).Error("Request failed")
	} else {
		log.Entry().WithField("StatusCode", resp.Status).Error(message)

		// Include the error message of the ABAP Environment system, if available
		var abapErrorResponse abaputils.AbapError
		bodyText, readError := ioutil.ReadAll(resp.Body)
		if readError != nil {
			return readError
		}
		var abapResp map[string]*json.RawMessage
		json.Unmarshal(bodyText, &abapResp)
		json.Unmarshal(*abapResp["error"], &abapErrorResponse)
		if (abaputils.AbapError{}) != abapErrorResponse {
			log.Entry().WithField("ErrorCode", abapErrorResponse.Code).Error(abapErrorResponse.Message.Value)
			abapError := errors.New(abapErrorResponse.Code + " - " + abapErrorResponse.Message.Value)
			err = errors.Wrap(abapError, err.Error())
		}
		resp.Body.Close()
	}
	return err
}

func printLogs(entity abaputils.PullEntity) {

	// Sort logs
	sort.SliceStable(entity.ToExecutionLog.Results, func(i, j int) bool {
		return entity.ToExecutionLog.Results[i].Index < entity.ToExecutionLog.Results[j].Index
	})

	sort.SliceStable(entity.ToTransportLog.Results, func(i, j int) bool {
		return entity.ToTransportLog.Results[i].Index < entity.ToTransportLog.Results[j].Index
	})

	log.Entry().Info("-------------------------")
	log.Entry().Info("Transport Log")
	log.Entry().Info("-------------------------")
	for _, logEntry := range entity.ToTransportLog.Results {

		log.Entry().WithField("Timestamp", convertTime(logEntry.Timestamp)).Info(logEntry.Description)
	}

	log.Entry().Info("-------------------------")
	log.Entry().Info("Execution Log")
	log.Entry().Info("-------------------------")
	for _, logEntry := range entity.ToExecutionLog.Results {
		log.Entry().WithField("Timestamp", convertTime(logEntry.Timestamp)).Info(logEntry.Description)
	}
	log.Entry().Info("-------------------------")

}

func convertTime(logTimeStamp string) time.Time {
	// The ABAP Environment system returns the date in the following format: /Date(1585576807000+0000)/
	seconds := strings.TrimPrefix(strings.TrimSuffix(logTimeStamp, "000+0000)/"), "/Date(")
	n, error := strconv.ParseInt(seconds, 10, 64)
	if error != nil {
		return time.Unix(0, 0).UTC()
	}
	t := time.Unix(n, 0).UTC()
	return t
}
