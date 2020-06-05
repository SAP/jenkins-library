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

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func abapEnvironmentPullGitRepo(config abapEnvironmentPullGitRepoOptions, telemetryData *telemetry.CustomData) error {

	// Determine the host, user and password, either via the input parameters or via a cloud foundry service key
	c := command.Command{}
	connectionDetails, errorGetInfo := getAbapCommunicationArrangementInfo(config, &c)
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

	log.Entry().Infof("Start pulling %v repositories", len(config.RepositoryNames))
	for _, repositoryName := range config.RepositoryNames {

		log.Entry().Info("-------------------------")
		log.Entry().Info("Start pulling " + repositoryName)
		log.Entry().Info("-------------------------")

		// Triggering the Pull of the repository into the ABAP Environment system
		uriConnectionDetails, errorTriggerPull := triggerPull(repositoryName, connectionDetails, &client)
		if errorTriggerPull != nil {
			log.Entry().WithError(errorTriggerPull).Fatal("Pull of " + repositoryName + " failed on the ABAP System")
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
	}
	log.Entry().Info("-------------------------")
	log.Entry().Info("All repositories were pulled successfully")
	return nil
}

func triggerPull(repositoryName string, pullConnectionDetails connectionDetailsHTTP, client piperhttp.Sender) (connectionDetailsHTTP, error) {

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
	var body abapEntity
	var abapResp map[string]*json.RawMessage
	bodyText, errRead := ioutil.ReadAll(resp.Body)
	if errRead != nil {
		return uriConnectionDetails, err
	}
	json.Unmarshal(bodyText, &abapResp)
	json.Unmarshal(*abapResp["d"], &body)
	if reflect.DeepEqual(abapEntity{}, body) {
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", repositoryName).Error("Could not pull the Repository / Software Component")
		err := errors.New("Request to ABAP System not successful")
		return uriConnectionDetails, err
	}

	expandLog := "?$expand=to_Execution_log,to_Transport_log"
	uriConnectionDetails.URL = body.Metadata.URI + expandLog
	return uriConnectionDetails, nil
}

func pollEntity(repositoryName string, connectionDetails connectionDetailsHTTP, client piperhttp.Sender, pollIntervall time.Duration) (string, error) {

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
		var body abapEntity
		bodyText, _ := ioutil.ReadAll(resp.Body)
		var abapResp map[string]*json.RawMessage
		json.Unmarshal(bodyText, &abapResp)
		json.Unmarshal(*abapResp["d"], &body)
		if reflect.DeepEqual(abapEntity{}, body) {
			log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", repositoryName).Error("Could not pull the Repository / Software Component")
			var err = errors.New("Request to ABAP System not successful")
			return "", err
		}
		status = body.Status
		log.Entry().WithField("StatusCode", resp.Status).Info("Pull Status: " + body.StatusDescr)
		if body.Status != "R" {
			printLogs(body)
			break
		}
		time.Sleep(pollIntervall)
	}

	return status, nil
}

func getAbapCommunicationArrangementInfo(config abapEnvironmentPullGitRepoOptions, c execRunner) (connectionDetailsHTTP, error) {

	var connectionDetails connectionDetailsHTTP
	var error error

	if config.Host != "" {
		// Host, User and Password are directly provided
		connectionDetails.URL = "https://" + config.Host + "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull"
		connectionDetails.User = config.Username
		connectionDetails.Password = config.Password
	} else {
		if config.CfAPIEndpoint == "" || config.CfOrg == "" || config.CfSpace == "" || config.CfServiceInstance == "" || config.CfServiceKey == "" {
			var err = errors.New("Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510")
			return connectionDetails, err
		}
		// Url, User and Password should be read from a cf service key
		var abapServiceKey, error = readCfServiceKey(config, c)
		if error != nil {
			return connectionDetails, errors.Wrap(error, "Read service key failed")
		}
		connectionDetails.URL = abapServiceKey.URL + "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull"
		connectionDetails.User = abapServiceKey.Abap.Username
		connectionDetails.Password = abapServiceKey.Abap.Password
	}
	return connectionDetails, error
}

func readCfServiceKey(config abapEnvironmentPullGitRepoOptions, c execRunner) (serviceKey, error) {

	var abapServiceKey serviceKey

	c.Stderr(log.Writer())

	// Logging into the Cloud Foundry via CF CLI
	log.Entry().WithField("cfApiEndpoint", config.CfAPIEndpoint).WithField("cfSpace", config.CfSpace).WithField("cfOrg", config.CfOrg).WithField("User", config.Username).Info("Cloud Foundry parameters: ")
	cfLoginSlice := []string{"login", "-a", config.CfAPIEndpoint, "-u", config.Username, "-p", config.Password, "-o", config.CfOrg, "-s", config.CfSpace}
	errorRunExecutable := c.RunExecutable("cf", cfLoginSlice...)
	if errorRunExecutable != nil {
		log.Entry().Error("Login at cloud foundry failed.")
		return abapServiceKey, errorRunExecutable
	}

	// Reading the Service Key via CF CLI
	var serviceKeyBytes bytes.Buffer
	c.Stdout(&serviceKeyBytes)
	cfReadServiceKeySlice := []string{"service-key", config.CfServiceInstance, config.CfServiceKey}
	errorRunExecutable = c.RunExecutable("cf", cfReadServiceKeySlice...)
	var serviceKeyJSON string
	if len(serviceKeyBytes.String()) > 0 {
		var lines []string = strings.Split(serviceKeyBytes.String(), "\n")
		serviceKeyJSON = strings.Join(lines[2:], "")
	}
	if errorRunExecutable != nil {
		return abapServiceKey, errorRunExecutable
	}
	log.Entry().WithField("cfServiceInstance", config.CfServiceInstance).WithField("cfServiceKey", config.CfServiceKey).Info("Read service key for service instance")
	json.Unmarshal([]byte(serviceKeyJSON), &abapServiceKey)
	if abapServiceKey == (serviceKey{}) {
		return abapServiceKey, errors.New("Parsing the service key failed")
	}
	return abapServiceKey, errorRunExecutable
}

func getHTTPResponse(requestType string, connectionDetails connectionDetailsHTTP, body []byte, client piperhttp.Sender) (*http.Response, error) {

	header := make(map[string][]string)
	header["Content-Type"] = []string{"application/json"}
	header["Accept"] = []string{"application/json"}
	header["x-csrf-token"] = []string{connectionDetails.XCsrfToken}

	req, err := client.SendRequest(requestType, connectionDetails.URL, bytes.NewBuffer(body), header, nil)
	return req, err
}

func handleHTTPError(resp *http.Response, err error, message string, connectionDetails connectionDetailsHTTP) error {
	if resp == nil {
		// Response is nil in case of a timeout
		log.Entry().WithError(err).WithField("ABAP Endpoint", connectionDetails.URL).Error("Request failed")
	} else {
		log.Entry().WithField("StatusCode", resp.Status).Error(message)

		// Include the error message of the ABAP Environment system, if available
		var abapErrorResponse abapError
		bodyText, readError := ioutil.ReadAll(resp.Body)
		if readError != nil {
			return readError
		}
		var abapResp map[string]*json.RawMessage
		json.Unmarshal(bodyText, &abapResp)
		json.Unmarshal(*abapResp["error"], &abapErrorResponse)
		if (abapError{}) != abapErrorResponse {
			log.Entry().WithField("ErrorCode", abapErrorResponse.Code).Error(abapErrorResponse.Message.Value)
			abapError := errors.New(abapErrorResponse.Code + " - " + abapErrorResponse.Message.Value)
			err = errors.Wrap(abapError, err.Error())
		}
		resp.Body.Close()
	}
	return err
}

func printLogs(entity abapEntity) {

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

type abapEntity struct {
	Metadata       abapMetadata `json:"__metadata"`
	UUID           string       `json:"uuid"`
	ScName         string       `json:"sc_name"`
	Namespace      string       `json:"namepsace"`
	Status         string       `json:"status"`
	StatusDescr    string       `json:"status_descr"`
	ToExecutionLog abapLogs     `json:"to_Execution_log"`
	ToTransportLog abapLogs     `json:"to_Transport_log"`
}

type abapMetadata struct {
	URI string `json:"uri"`
}

type abapLogs struct {
	Results []logResults `json:"results"`
}

type logResults struct {
	Index       string `json:"index_no"`
	Type        string `json:"type"`
	Description string `json:"descr"`
	Timestamp   string `json:"timestamp"`
}

type serviceKey struct {
	Abap     abapConenction `json:"abap"`
	Binding  abapBinding    `json:"binding"`
	Systemid string         `json:"systemid"`
	URL      string         `json:"url"`
}

type deferred struct {
	URI string `json:"uri"`
}

type abapConenction struct {
	CommunicationArrangementID string `json:"communication_arrangement_id"`
	CommunicationScenarioID    string `json:"communication_scenario_id"`
	CommunicationSystemID      string `json:"communication_system_id"`
	Password                   string `json:"password"`
	Username                   string `json:"username"`
}

type abapBinding struct {
	Env     string `json:"env"`
	ID      string `json:"id"`
	Type    string `json:"type"`
	Version string `json:"version"`
}

type connectionDetailsHTTP struct {
	User       string `json:"user"`
	Password   string `json:"password"`
	URL        string `json:"url"`
	XCsrfToken string `json:"xcsrftoken"`
}

type abapError struct {
	Code    string           `json:"code"`
	Message abapErrorMessage `json:"message"`
}

type abapErrorMessage struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}
