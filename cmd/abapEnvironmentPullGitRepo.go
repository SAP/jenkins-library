package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os/exec"

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
	var uri, err = triggerPull(config, abapUrl, user, password)
	if err != nil {
		return err
	}
	var _, _ = pollEntity(config, uri, user, password)

	return nil
}

func pollEntity(config abapEnvironmentPullGitRepoOptions, uri string, user string, password string) (string, error) {
	return "", nil
}

func triggerPull(config abapEnvironmentPullGitRepoOptions, abapUrl string, user string, password string) (string, error) {

	var entityUri string

	log.Entry().WithField("ABAP Endpoint", abapUrl).Info("Calling the ABAP System...")
	log.Entry().Info("Trying to authenticate on the ABAP system...")
	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: cookieJar,
	}
	req, _ := http.NewRequest("GET", abapUrl, nil)
	req.Header.Add("x-csrf-token", "fetch")
	req.SetBasicAuth(user, password)
	resp, err := client.Do(req)
	if err != nil {
		return entityUri, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Entry().WithField("StatusCode", resp.Status).Error("Authentication failed")
		var err = errors.New("Request to ABAP System not successful")
		return entityUri, err
	} else {
		log.Entry().WithField("StatusCode", resp.Status).Info("Authentication successfull")
	}
	xCsrfToken := resp.Header.Get("X-Csrf-Token")

	var jsonBody = []byte(`{"sc_name":"` + config.RepositoryName + `"}`)

	log.Entry().WithField("repositoryName", config.RepositoryName).Info("Pulling Repository / Software Component")
	client = &http.Client{
		Jar: cookieJar,
	}
	req, _ = http.NewRequest("POST", abapUrl, bytes.NewBuffer(jsonBody))
	req.Header.Add("x-csrf-token", xCsrfToken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(user, password)
	resp, err = client.Do(req)
	if err != nil {
		return entityUri, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", config.RepositoryName).Error("Could not pull the Repository / Software Component")
		var err = errors.New("Request to ABAP System not successful")
		return entityUri, err
	} else {
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", config.RepositoryName).Info("Triggered Pull of Repository / Software Component")
	}

	var body AbapResponse
	bodyText, err := ioutil.ReadAll(resp.Body)
	json.Unmarshal(bodyText, &body)
	if body.D == (AbapEntity{}) {
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", config.RepositoryName).Error("Could not pull the Repository / Software Component")
		var err = errors.New("Request to ABAP System not successful")
		return entityUri, err
	}
	entityUri = body.D.Metadata.Uri
	// TODO In error case json looks different
	return entityUri, nil
}

func getAbapCommunicationArrangementInfo(config abapEnvironmentPullGitRepoOptions, s shellRunner) (string, string, string, error) {

	var abapUrl string
	var user string
	var password string
	var error error

	if config.Host != "" {
		abapUrl = "https://" + config.Host + "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull"
		user = config.User
		password = config.Password
	} else {
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
	log.Entry().WithField("cfApiEndpoint", config.CfAPIEndpoint).WithField("cfSpace", config.CfSpace).WithField("cfOrg", config.CfOrg).WithField("User", config.User).Info("Cloud Foundry parameters: ")
	var cfLoginScript = "cf login -a " + config.CfAPIEndpoint + " -u " + config.User + " -p " + config.Password + " -o " + config.CfOrg + " -s " + config.CfSpace

	cflogin, error := exec.Command("sh", "-c", cfLoginScript).Output()
	fmt.Printf("%s\n\n", cflogin)
	if error != nil {
		return serviceKey, error
	}

	log.Entry().WithField("cfServiceInstance", config.CfServiceInstance).WithField("cfServiceKey", config.CfServiceKey).Info("Reading service key for service instance:")
	var cfReadServiceKeyScript = "cf service-key " + config.CfServiceInstance + " " + config.CfServiceKey + " | awk '{if(NR>1)print}'"
	cfServiceKey, error := exec.Command("sh", "-c", cfReadServiceKeyScript).Output()
	if error != nil {
		return serviceKey, error
	}

	json.Unmarshal([]byte(cfServiceKey), &serviceKey)
	return serviceKey, error
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
