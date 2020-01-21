package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os/exec"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

func abapEnvironmentPullGitRepo(config abapEnvironmentPullGitRepoOptions) error {
	c := command.Command{}

	var abapUrl, user, password, error = getAbapCommunicationArrangementInfo(config, &c)
	if error != nil {
		return error
	}
	var uri, xCsrfToken, cookieJar, err = triggerPull(config, abapUrl, user, password)
	if err != nil {
		return err
	}
	var status, _ = pollEntity(config, uri, user, password, xCsrfToken, cookieJar)

	if status == "E" {
		log.Entry().Error("Pull failed on the ABAP system")
		return errors.New("Pull failed on the ABAP system")
	}

	return nil
}

func pollEntity(config abapEnvironmentPullGitRepoOptions, uri string, user string, password string, xCsrfToken string, cookieJar *cookiejar.Jar) (string, error) {

	log.Entry().Info("Start polling the status...")
	var status string = "R"

	for {
		var resp, err = getHttpResponse("GET", uri, nil, xCsrfToken, user, password, cookieJar)
		defer resp.Body.Close()
		if err != nil {
			log.Entry().Error("Request to ABAP System not successful")
			return "", err
		}

		if resp.StatusCode != 200 {
			log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", config.RepositoryName).Error("Could not pull the Repository / Software Component")
			var err = errors.New("Request to ABAP System not successful")
			return "", err
		}

		var body AbapResponse
		bodyText, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(bodyText, &body)
		if body.D == (AbapEntity{}) {
			log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", config.RepositoryName).Error("Could not pull the Repository / Software Component")
			var err = errors.New("Request to ABAP System not successful")
			return "", err
		}
		status = body.D.Status
		log.Entry().Info("Pull Status: " + body.D.Status_descr)
		if body.D.Status != "R" {
			break
		}
		time.Sleep(10 * time.Second)
	}

	return status, nil
}

func triggerPull(config abapEnvironmentPullGitRepoOptions, abapUrl string, user string, password string) (string, string, *cookiejar.Jar, error) {

	var entityUri string
	var xCsrfToken string
	cookieJar, _ := cookiejar.New(nil)

	// Loging into the ABAP System - getting the x-csrf-token and cookies
	log.Entry().WithField("ABAP Endpoint", abapUrl).Info("Calling the ABAP System...")
	log.Entry().Info("Trying to authenticate on the ABAP system...")

	var resp, err = getHttpResponse("GET", abapUrl, nil, "fetch", user, password, cookieJar)
	defer resp.Body.Close()
	if err != nil {
		log.Entry().Error("Request to ABAP System not successful")
		return entityUri, xCsrfToken, cookieJar, err
	}

	if resp.StatusCode != 200 {
		log.Entry().WithField("StatusCode", resp.Status).Error("Authentication failed")
		var err = errors.New("Request to ABAP System not successful")
		return entityUri, xCsrfToken, cookieJar, err
	} else {
		log.Entry().WithField("StatusCode", resp.Status).Info("Authentication successfull")
	}
	xCsrfToken = resp.Header.Get("X-Csrf-Token")

	// Trigger the Pull of a Repository
	var jsonBody = []byte(`{"sc_name":"` + config.RepositoryName + `"}`)
	log.Entry().WithField("repositoryName", config.RepositoryName).Info("Pulling Repository / Software Component")

	resp, err = getHttpResponse("POST", abapUrl, jsonBody, xCsrfToken, user, password, cookieJar)
	defer resp.Body.Close()
	if err != nil {
		log.Entry().Error("Request to ABAP System not successful")
		return entityUri, xCsrfToken, cookieJar, err
	}
	if resp.StatusCode >= 300 {
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", config.RepositoryName).Error("Could not pull the Repository / Software Component")
		var err = errors.New("Request to ABAP System not successful")
		return entityUri, xCsrfToken, cookieJar, err
	} else {
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", config.RepositoryName).Info("Triggered Pull of Repository / Software Component")
	}

	// Parse Response
	var body AbapResponse
	bodyText, err := ioutil.ReadAll(resp.Body)
	json.Unmarshal(bodyText, &body)
	if body.D == (AbapEntity{}) {
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", config.RepositoryName).Error("Could not pull the Repository / Software Component")
		var err = errors.New("Request to ABAP System not successful")
		return entityUri, xCsrfToken, cookieJar, err
	}
	entityUri = body.D.Metadata.Uri
	return entityUri, xCsrfToken, cookieJar, nil
}

func getAbapCommunicationArrangementInfo(config abapEnvironmentPullGitRepoOptions, s shellRunner) (string, string, string, error) {

	var abapUrl string
	var user string
	var password string
	var error error

	if config.Host != "" {
		// Host, User and Password are directly provided
		abapUrl = "https://" + config.Host + "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull"
		user = config.User
		password = config.Password
	} else {
		// Url, User and Password should be read from a cf service key
		var serviceKey, error = readCfServiceKey(config, s)
		if error != nil {
			log.Entry().Error(error)
		}
		abapUrl = serviceKey.Url + "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull"
		user = serviceKey.Abap.Username
		password = serviceKey.Abap.Password
	}
	return abapUrl, user, password, error
}

func readCfServiceKey(config abapEnvironmentPullGitRepoOptions, s shellRunner) (ServiceKey, error) {

	var serviceKey ServiceKey

	if config.CfAPIEndpoint == "" || config.CfOrg == "" || config.CfSpace == "" || config.CfServiceInstance == "" || config.CfServiceKey == "" {
		var err = errors.New("Cloud Foundry parameters are not provided. Please provide the ApiEndpoint, the Organization, the Space, the Service Instance, a corresponding Service Key, a User and the corresponding Password")
		return serviceKey, err
	}

	// Logging into the Cloud Foundry via CF CLI
	log.Entry().WithField("cfApiEndpoint", config.CfAPIEndpoint).WithField("cfSpace", config.CfSpace).WithField("cfOrg", config.CfOrg).WithField("User", config.User).Info("Cloud Foundry parameters: ")
	var cfLoginScript = "cf login -a " + config.CfAPIEndpoint + " -u " + config.User + " -p " + config.Password + " -o " + config.CfOrg + " -s " + config.CfSpace
	cflogin, error := exec.Command("sh", "-c", cfLoginScript).Output()
	fmt.Printf("%s\n\n", cflogin)
	if error != nil {
		return serviceKey, error
	}

	// Reading the Service Key via CF CLI
	log.Entry().WithField("cfServiceInstance", config.CfServiceInstance).WithField("cfServiceKey", config.CfServiceKey).Info("Reading service key for service instance:")
	var cfReadServiceKeyScript = "cf service-key " + config.CfServiceInstance + " " + config.CfServiceKey + " | awk '{if(NR>1)print}'"
	cfServiceKey, error := exec.Command("sh", "-c", cfReadServiceKeyScript).Output()
	if error != nil {
		return serviceKey, error
	}

	json.Unmarshal([]byte(cfServiceKey), &serviceKey)
	return serviceKey, error
}

func getHttpResponse(requestType string, url string, body []byte, xCsrfToken string, user string, password string, cookieJar *cookiejar.Jar) (*http.Response, error) {
	client := &http.Client{
		Jar: cookieJar,
	}
	req, _ := http.NewRequest(requestType, url, bytes.NewBuffer(body))
	req.Header.Add("x-csrf-token", xCsrfToken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(user, password)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type AbapResponse struct {
	D AbapEntity
}

type AbapEntity struct {
	Metadata         AbapMetadata `json:"__metadata"`
	Uuid             string
	Sc_name          string
	Namespace        string
	Status           string
	Status_descr     string
	To_Execution_log Deferred `json:"__deferred"`
	To_Transport_log Deferred
}

type AbapMetadata struct {
	Uri string
}

type ServiceKey struct {
	Abap     AbapConenction
	Binding  AbapBinding
	Systemid string
	Url      string
}

type Deferred struct {
	Uri string
}

type AbapConenction struct {
	Communication_arrangement_id string
	Communication_scenario_id    string
	Communication_system_id      string
	Password                     string
	Username                     string
}

type AbapBinding struct {
	Env     string
	Id      string
	Type    string
	Tersion string
}
