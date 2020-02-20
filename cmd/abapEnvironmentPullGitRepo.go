package cmd

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func abapEnvironmentPullGitRepo(config abapEnvironmentPullGitRepoOptions, telemetryData *telemetry.CustomData) error {
	c := command.Command{}
	var connectionDetails, error = getAbapCommunicationArrangementInfo(config, &c)
	if error != nil {
		log.Entry().WithError(error).Fatal("Parameters for the ABAP Connection not available")
	}

	client := piperhttp.Client{}
	cookieJar, _ := cookiejar.New(nil)
	clientOptions := piperhttp.ClientOptions{
		CookieJar: cookieJar,
		Username:  connectionDetails.User,
		Password:  connectionDetails.Password,
	}
	client.SetOptions(clientOptions)

	var uriConnectionDetails, err = triggerPull(config, connectionDetails, &client)
	if err != nil {
		log.Entry().WithError(err).Fatal("Pull failed on the ABAP System")
	}

	var status, er = pollEntity(config, uriConnectionDetails, &client, 10*time.Second)
	if er != nil {
		log.Entry().WithError(er).Fatal("Pull failed on the ABAP System")
	}
	if status == "E" {
		log.Entry().Fatal("Pull failed on the ABAP System")
	}

	return nil
}

func triggerPull(config abapEnvironmentPullGitRepoOptions, pullConnectionDetails connectionDetailsHTTP, client piperhttp.Sender) (connectionDetailsHTTP, error) {

	uriConnectionDetails := pullConnectionDetails
	uriConnectionDetails.URL = ""
	pullConnectionDetails.XCsrfToken = "fetch"

	// Loging into the ABAP System - getting the x-csrf-token and cookies
	var resp, err = getHTTPResponse("HEAD", pullConnectionDetails, nil, client)
	defer resp.Body.Close()
	if err != nil {
		log.Entry().WithField("StatusCode", resp.Status).WithField("ABAP Endpoint", pullConnectionDetails.URL).Error("Authentication on the ABAP system failed")
		return uriConnectionDetails, err
	}
	log.Entry().WithField("StatusCode", resp.Status).WithField("ABAP Endpoint", pullConnectionDetails.URL).Info("Authentication on the ABAP system successfull")
	uriConnectionDetails.XCsrfToken = resp.Header.Get("X-Csrf-Token")
	pullConnectionDetails.XCsrfToken = uriConnectionDetails.XCsrfToken

	// Trigger the Pull of a Repository
	var jsonBody = []byte(`{"sc_name":"` + config.RepositoryName + `"}`)
	resp, err = getHTTPResponse("POST", pullConnectionDetails, jsonBody, client)
	defer resp.Body.Close()
	if err != nil {
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", config.RepositoryName).Error("Could not pull the Repository / Software Component")
		return uriConnectionDetails, err
	}
	log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", config.RepositoryName).Info("Triggered Pull of Repository / Software Component")

	// Parse Response
	var body abapEntity
	var abapResp map[string]*json.RawMessage
	bodyText, err := ioutil.ReadAll(resp.Body)
	json.Unmarshal(bodyText, &abapResp)
	json.Unmarshal(*abapResp["d"], &body)
	if body == (abapEntity{}) {
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", config.RepositoryName).Error("Could not pull the Repository / Software Component")
		var err = errors.New("Request to ABAP System not successful")
		return uriConnectionDetails, err
	}
	uriConnectionDetails.URL = body.Metadata.URI
	return uriConnectionDetails, nil
}

func pollEntity(config abapEnvironmentPullGitRepoOptions, connectionDetails connectionDetailsHTTP, client piperhttp.Sender, pollIntervall time.Duration) (string, error) {

	log.Entry().Info("Start polling the status...")
	var status string = "R"

	for {
		var resp, err = getHTTPResponse("GET", connectionDetails, nil, client)
		defer resp.Body.Close()
		if err != nil {
			log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", config.RepositoryName).Error("Could not pull the Repository / Software Component")
			return "", err
		}

		// Parse response
		var body abapEntity
		bodyText, _ := ioutil.ReadAll(resp.Body)
		var abapResp map[string]*json.RawMessage
		json.Unmarshal(bodyText, &abapResp)
		json.Unmarshal(*abapResp["d"], &body)
		if body == (abapEntity{}) {
			log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", config.RepositoryName).Error("Could not pull the Repository / Software Component")
			var err = errors.New("Request to ABAP System not successful")
			return "", err
		}
		status = body.Status
		log.Entry().WithField("StatusCode", resp.Status).Info("Pull Status: " + body.StatusDescr)
		if body.Status != "R" {
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

	c.Stderr(log.Entry().Writer())

	// Logging into the Cloud Foundry via CF CLI
	log.Entry().WithField("cfApiEndpoint", config.CfAPIEndpoint).WithField("cfSpace", config.CfSpace).WithField("cfOrg", config.CfOrg).WithField("User", config.Username).Info("Cloud Foundry parameters: ")
	cfLoginSlice := []string{"login", "-a", config.CfAPIEndpoint, "-u", config.Username, "-p", config.Password, "-o", config.CfOrg, "-s", config.CfSpace}
	error := c.RunExecutable("cf", cfLoginSlice...)
	if error != nil {
		log.Entry().Error("Login at cloud foundry failed.")
		return abapServiceKey, error
	}

	// Reading the Service Key via CF CLI
	var serviceKeyBytes bytes.Buffer
	c.Stdout(&serviceKeyBytes)
	cfReadServiceKeySlice := []string{"service-key", config.CfServiceInstance, config.CfServiceKey}
	error = c.RunExecutable("cf", cfReadServiceKeySlice...)
	var serviceKeyJSON string
	if len(serviceKeyBytes.String()) > 0 {
		var lines []string = strings.Split(serviceKeyBytes.String(), "\n")
		serviceKeyJSON = strings.Join(lines[2:], "")
	}
	if error != nil {
		return abapServiceKey, error
	}
	log.Entry().WithField("cfServiceInstance", config.CfServiceInstance).WithField("cfServiceKey", config.CfServiceKey).Info("Read service key for service instance")
	json.Unmarshal([]byte(serviceKeyJSON), &abapServiceKey)
	if abapServiceKey == (serviceKey{}) {
		return abapServiceKey, errors.New("Parsing the service key failed")
	}
	return abapServiceKey, error
}

func getHTTPResponse(requestType string, connectionDetails connectionDetailsHTTP, body []byte, client piperhttp.Sender) (*http.Response, error) {

	header := make(map[string][]string)
	header["Content-Type"] = []string{"application/json"}
	header["Accept"] = []string{"application/json"}
	header["x-csrf-token"] = []string{connectionDetails.XCsrfToken}

	req, err := client.SendRequest(requestType, connectionDetails.URL, bytes.NewBuffer(body), header, nil)
	return req, err
}

type abapEntity struct {
	Metadata       abapMetadata `json:"__metadata"`
	UUID           string       `json:"uuid"`
	ScName         string       `json:"sc_name"`
	Namespace      string       `json:"namepsace"`
	Status         string       `json:"status"`
	StatusDescr    string       `json:"status_descr"`
	ToExecutionLog deferred     `json:"to_Execution_log"`
	ToTransportLog deferred     `json:"to_Transport_log"`
}

type abapMetadata struct {
	URI string `json:"uri"`
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
