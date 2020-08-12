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

func abapEnvironmentCloneGitRepo(config abapEnvironmentCloneGitRepoOptions, telemetryData *telemetry.CustomData) {

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
	err := runAbapEnvironmentCloneGitRepo(&config, telemetryData, &autils, &client)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentCloneGitRepo(config *abapEnvironmentCloneGitRepoOptions, telemetryData *telemetry.CustomData, com abaputils.Communication, client piperhttp.Sender) error {
	// Mapping for options
	subOptions := abaputils.AbapEnvironmentOptions{}

	subOptions.CfAPIEndpoint = config.CfAPIEndpoint
	subOptions.CfServiceInstance = config.CfServiceInstance
	subOptions.CfServiceKeyName = config.CfServiceKeyName
	subOptions.CfOrg = config.CfOrg
	subOptions.CfSpace = config.CfSpace
	subOptions.Host = config.Host
	subOptions.Password = config.Password
	subOptions.Username = config.Username

	// Determine the host, user and password, either via the input parameters or via a cloud foundry service key
	connectionDetails, errorGetInfo := com.GetAbapCommunicationArrangementInfo(subOptions, "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Clones")
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
	pollIntervall := 10 * time.Second

	var repositories = make([]Repository, 0)

	if config.RepositoryName != "" && config.BranchName != "" {
		repositories = append(repositories, Repository{Name: config.RepositoryName, Branch: config.BranchName})
	}
	if config.Repositories != "" {
		// ToDo Parse json and append to repositories
	}

	log.Entry().Infof("Start cloning %v repositories", len(repositories))
	for _, repo := range repositories {
		log.Entry().Info("-------------------------")
		log.Entry().Info("Start cloning " + repo.Name + " with branch " + repo.Branch)
		log.Entry().Info("-------------------------")

		// Triggering the Clone of the repository into the ABAP Environment system
		uriConnectionDetails, errorTriggerClone := triggerClone(repo.Name, repo.Branch, connectionDetails, client)
		if errorTriggerClone != nil {
			return errors.Wrapf(errorTriggerClone, "Clone of '%s' with branch '%s' failed on the ABAP System", repo.Name, repo.Branch)
		}

		// Polling the status of the repository import on the ABAP Environment system
		status, errorPollEntity := pollEntity(repo.Name, uriConnectionDetails, client, pollIntervall)
		if errorPollEntity != nil {
			return errors.Wrapf(errorPollEntity, "Clone of '%s' with branch '%s' failed on the ABAP System", repo.Name, repo.Branch)
		}
		if status == "E" {
			return errors.New("Clone of " + repo.Name + " with branch " + repo.Branch + " failed on the ABAP System")
		}

		log.Entry().Info(repo.Name + " with branch  " + repo.Branch + " was cloned successfully")
	}
	log.Entry().Info("-------------------------")
	log.Entry().Info("All repositories were cloned successfully")
	return nil
}

/*
Repository type
*/
type Repository struct {
	Name   string
	Branch string
}

func triggerClone(repositoryName string, branchName string, cloneConnectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) (abaputils.ConnectionDetailsHTTP, error) {
	uriConnectionDetails := cloneConnectionDetails
	uriConnectionDetails.URL = ""
	cloneConnectionDetails.XCsrfToken = "fetch"

	// Loging into the ABAP System - getting the x-csrf-token and cookies
	resp, err := getHTTPResponse("HEAD", cloneConnectionDetails, nil, client)
	if err != nil {
		err = handleHTTPError(resp, err, "Authentication on the ABAP system failed", cloneConnectionDetails)
		return uriConnectionDetails, err
	}
	defer resp.Body.Close()
	log.Entry().WithField("StatusCode", resp.Status).WithField("ABAP Endpoint", cloneConnectionDetails.URL).Info("Authentication on the ABAP system successfull")
	uriConnectionDetails.XCsrfToken = resp.Header.Get("X-Csrf-Token")
	cloneConnectionDetails.XCsrfToken = uriConnectionDetails.XCsrfToken

	// Trigger the Clone of a Repository
	if repositoryName == "" {
		return uriConnectionDetails, errors.New("An empty string was passed for the parameter 'repositoryName'")
	}
	jsonBody := []byte(`{"sc_name":"` + repositoryName + `"}`)
	resp, err = getHTTPResponse("POST", cloneConnectionDetails, jsonBody, client)
	if err != nil {
		err = handleHTTPError(resp, err, "Could not clone the Repository / Software Component "+repositoryName+" with branch "+branchName, uriConnectionDetails)
		return uriConnectionDetails, err
	}
	defer resp.Body.Close()
	log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", repositoryName).WithField("branchName", branchName).Info("Triggered Clone of Repository / Software Component")

	// Parse Response
	var body abaputils.CloneEntity
	var abapResp map[string]*json.RawMessage
	bodyText, errRead := ioutil.ReadAll(resp.Body)
	if errRead != nil {
		return uriConnectionDetails, err
	}
	json.Unmarshal(bodyText, &abapResp)
	json.Unmarshal(*abapResp["d"], &body)
	if reflect.DeepEqual(abaputils.CloneEntity{}, body) {
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", repositoryName).WithField("branchName", branchName).Error("Could not Clone the Repository / Software Component")
		err := errors.New("Request to ABAP System not successful")
		return uriConnectionDetails, err
	}

	expandLog := "?$expand=to_Execution_log,to_Transport_log"
	uriConnectionDetails.URL = body.Metadata.URI + expandLog
	return uriConnectionDetails, nil
}
