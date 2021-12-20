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

func abapEnvironmentPullGitRepo(options abapEnvironmentPullGitRepoOptions, _ *telemetry.CustomData) {

	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	var autils = abaputils.AbapUtils{
		Exec: &c,
	}

	client := piperhttp.Client{}

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapEnvironmentPullGitRepo(&options, &autils, &client)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentPullGitRepo(options *abapEnvironmentPullGitRepoOptions, com abaputils.Communication, client piperhttp.Sender) (err error) {

	subOptions := convertPullConfig(options)

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
		handleIgnoreCommit(repositories, options.IgnoreCommit)
	}

	if err == nil {
		err = pullRepositories(repositories, connectionDetails, client, pollIntervall)
	}

	if err != nil {
		return fmt.Errorf("Something failed during the pull of the repositories: %w", err)
	}

	return err
}

func pullRepositories(repositories []abaputils.Repository, pullConnectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender, pollIntervall time.Duration) (err error) {
	log.Entry().Infof("Start pulling %v repositories", len(repositories))
	for _, repo := range repositories {
		err = handlePull(repo, pullConnectionDetails, client, pollIntervall)
		if err != nil {
			break
		}
	}
	if err == nil {
		finishPullLogs()
	}
	return err
}

func handlePull(repo abaputils.Repository, pullConnectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender, pollIntervall time.Duration) (err error) {

	startPullLogs(repo)

	_, logString := abaputils.CreateAdditionalBodyParameters(repo.CommitID, repo.Tag)

	uriConnectionDetails, err := triggerPull(repo, pullConnectionDetails, client)
	if err != nil {
		return errors.Wrapf(err, "Pull of '%s'%s failed on the ABAP System", repo.Name, logString)
	}

	// Polling the status of the repository import on the ABAP Environment system
	status, errorPollEntity := abaputils.PollEntity(repo.Name, uriConnectionDetails, client, pollIntervall)
	if errorPollEntity != nil {
		return errors.Wrapf(errorPollEntity, "Pull of '%s'%s failed on the ABAP System", repo.Name, logString)
	}
	if status == "E" {
		return errors.New("Pull of '" + repo.Name + "'" + logString + " failed on the ABAP System")
	}
	log.Entry().Info(repo.Name + " was pulled successfully")
	return err
}

func triggerPull(repo abaputils.Repository, pullConnectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) (abaputils.ConnectionDetailsHTTP, error) {

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

	log.Entry().WithField("StatusCode", resp.Status).WithField("ABAP Endpoint", pullConnectionDetails.URL).Debug("Authentication on the ABAP system successful")
	uriConnectionDetails.XCsrfToken = resp.Header.Get("X-Csrf-Token")
	pullConnectionDetails.XCsrfToken = uriConnectionDetails.XCsrfToken

	// Trigger the Pull of a Repository
	if repo.Name == "" {
		return uriConnectionDetails, errors.New("An empty string was passed for the parameter 'repositoryName'")
	}
	query, logString := abaputils.CreateAdditionalBodyParameters(repo.CommitID, repo.Tag)
	jsonBody := []byte(`{"sc_name":"` + repo.Name + `"` + query + `}`)
	resp, err = abaputils.GetHTTPResponse("POST", pullConnectionDetails, jsonBody, client)
	if err != nil {
		err = abaputils.HandleHTTPError(resp, err, "Could not pull the Repository / Software Component '"+repo.Name+"'"+logString, uriConnectionDetails)
		return uriConnectionDetails, err
	}
	defer resp.Body.Close()
	log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", repo.Name).WithField("commitID", repo.CommitID).Debug("Triggered Pull of Repository / Software Component")

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
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", repo.Name).WithField("commitID", repo.CommitID).Error("Could not pull the Repository / Software Component")
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

func startPullLogs(repo abaputils.Repository) {
	_, logString := abaputils.CreateAdditionalBodyParameters(repo.CommitID, repo.Tag)
	log.Entry().Info("-------------------------")
	log.Entry().Info("Start pulling '" + repo.Name + "'" + logString)
	log.Entry().Info("-------------------------")
}

func finishPullLogs() {
	log.Entry().Info("-------------------------")
	log.Entry().Info("All repositories were pulled successfully")
}

func convertPullConfig(config *abapEnvironmentPullGitRepoOptions) abaputils.AbapEnvironmentOptions {
	subOptions := abaputils.AbapEnvironmentOptions{}

	subOptions.CfAPIEndpoint = config.CfAPIEndpoint
	subOptions.CfServiceInstance = config.CfServiceInstance
	subOptions.CfServiceKeyName = config.CfServiceKeyName
	subOptions.CfOrg = config.CfOrg
	subOptions.CfSpace = config.CfSpace
	subOptions.Host = config.Host
	subOptions.Password = config.Password
	subOptions.Username = config.Username
	return subOptions
}

func handleIgnoreCommit(repositories []abaputils.Repository, ignoreCommit bool) {
	for i := range repositories {
		if ignoreCommit {
			repositories[i].CommitID = ""
		}
	}
}
