package cmd

import (
	"encoding/json"
	"fmt"
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

func runAbapEnvironmentPullGitRepo(options *abapEnvironmentPullGitRepoOptions, telemetryData *telemetry.CustomData, com abaputils.Communication, client piperhttp.Sender) (err error) {
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
	connectionDetails, err := com.GetAbapCommunicationArrangementInfo(subOptions, "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull")
	if err != nil {
		return errors.Wrap(err, "Parameters for the ABAP Connection not available")
	}

	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return errors.Wrap(err, "Could not create a Cookie Jar")
	}
	clientOptions := piperhttp.ClientOptions{
		MaxRequestDuration: 180 * time.Second,
		CookieJar:          cookieJar,
		Username:           connectionDetails.User,
		Password:           connectionDetails.Password,
	}
	client.SetOptions(clientOptions)
	pollIntervall := com.GetPollIntervall()

	repositories := []abaputils.Repository{}
	err = checkPullRepositoryConfiguration(*options)

	if err == nil {
		repositories, err = abaputils.GetRepositories(&abaputils.RepositoriesConfig{RepositoryNames: options.RepositoryNames, Repositories: options.Repositories})
	}

	if err == nil {
		err = pullRepositories(repositories, connectionDetails, client, pollIntervall)
	}

	if err != nil {
		return fmt.Errorf("Something failed during the pull of the repositories: %w", err)
	}

	log.Entry().Info("-------------------------")
	log.Entry().Info("All repositories were pulled successfully")
	return err
}

func pullRepositories(repositories []abaputils.Repository, pullConnectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender, pollIntervall time.Duration) (err error) {
	log.Entry().Infof("Start cloning %v repositories", len(repositories))
	for _, repo := range repositories {
		err = handlePull(repo, pullConnectionDetails, client, pollIntervall)
		if err != nil {
			break
		}
		finishPullLogs()
	}
	return err
}

func triggerPull(repositoryName string, pullConnectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) (abaputils.ConnectionDetailsHTTP, error) {

	uriConnectionDetails := pullConnectionDetails
	uriConnectionDetails.URL = ""
	pullConnectionDetails.XCsrfToken = "fetch"

	// Loging into the ABAP System - getting the x-csrf-token and cookies
	resp, err := abaputils.GetHTTPResponse("HEAD", pullConnectionDetails, nil, client)
	if err != nil {
		err = abaputils.HandleHTTPError(resp, err, "Authentication on the ABAP system failed", pullConnectionDetails)
		return uriConnectionDetails, err
	}
	defer resp.Body.Close()

	// workaround until golang version 1.16 is used
	time.Sleep(1 * time.Second)

	log.Entry().WithField("StatusCode", resp.Status).WithField("ABAP Endpoint", pullConnectionDetails.URL).Info("Authentication on the ABAP system successful")
	uriConnectionDetails.XCsrfToken = resp.Header.Get("X-Csrf-Token")
	pullConnectionDetails.XCsrfToken = uriConnectionDetails.XCsrfToken

	// Trigger the Pull of a Repository
	if repositoryName == "" {
		return uriConnectionDetails, errors.New("An empty string was passed for the parameter 'repositoryName'")
	}
	jsonBody := []byte(`{"sc_name":"` + repositoryName + `"}`)
	resp, err = abaputils.GetHTTPResponse("POST", pullConnectionDetails, jsonBody, client)
	if err != nil {
		err = abaputils.HandleHTTPError(resp, err, "Could not pull the Repository / Software Component "+repositoryName, uriConnectionDetails)
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

func checkPullRepositoryConfiguration(options abapEnvironmentPullGitRepoOptions) error {
	if len(options.RepositoryNames) > 0 && options.Repositories != "" {
		log.Entry().Info("It seems like you have specified repositories directly via the configuration parameter 'repositoryNames' as well as in the dedicated repositories configuration file. Please note that in this case both configurations will be handled and pulled.")
	}
	if len(options.RepositoryNames) == 0 && options.Repositories == "" {
		return fmt.Errorf("Checking configuration failed: %w", errors.New("You have not specified any repository configuration to be pulled into the ABAP Environment System. Please make sure that you specified the repositories that should be pulled either in a dedicated file or via the parameter 'repositoryNames'. For more information please read the User documentation"))
	}
	return nil
}

func handlePull(repo abaputils.Repository, pullConnectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender, pollIntervall time.Duration) (err error) {
	startPullLogs(repo.Branch)

	uriConnectionDetails, err := triggerPull(repo.Name, pullConnectionDetails, client)
	if err != nil {
		return errors.Wrapf(err, "Pull of '%s' failed on the ABAP System", repo.Name)
	}

	// Polling the status of the repository import on the ABAP Environment system
	status, errorPollEntity := abaputils.PollEntity(repo.Name, uriConnectionDetails, client, pollIntervall)
	if errorPollEntity != nil {
		return errors.Wrapf(errorPollEntity, "Pull of '%s' failed on the ABAP System", repo.Name)
	}
	if status == "E" {
		return errors.New("Pull of " + repo.Name + " failed on the ABAP System")
	}
	log.Entry().Info(repo.Name + " was pulled successfully")
	return err
}

func startPullLogs(repositoryName string) {
	log.Entry().Info("-------------------------")
	log.Entry().Info("Start pulling " + repositoryName)
	log.Entry().Info("-------------------------")
}

func finishPullLogs() {
	log.Entry().Info("-------------------------")
	log.Entry().Info("All repositories were pulled successfully")
}
