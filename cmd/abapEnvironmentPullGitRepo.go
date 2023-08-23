package cmd

import (
	"encoding/json"
	"fmt"
	"io"
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

	var repositories []abaputils.Repository
	err = checkPullRepositoryConfiguration(*options)
	if err != nil {
		return err
	}
	repositories, err = abaputils.GetRepositories(&abaputils.RepositoriesConfig{RepositoryNames: options.RepositoryNames, Repositories: options.Repositories, RepositoryName: options.RepositoryName, CommitID: options.CommitID}, false)
	handleIgnoreCommit(repositories, options.IgnoreCommit)
	if err != nil {
		return err
	}

	err = pullRepositories(repositories, connectionDetails, client, pollIntervall)
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

	logString := repo.GetPullLogString()
	errorString := "Pull of the " + logString + " failed on the ABAP system"

	abaputils.AddDefaultDashedLine()
	log.Entry().Info("Start pulling the " + logString)
	abaputils.AddDefaultDashedLine()

	uriConnectionDetails, err := triggerPull(repo, pullConnectionDetails, client)
	if err != nil {
		return errors.Wrapf(err, errorString)
	}

	// Polling the status of the repository import on the ABAP Environment system
	status, errorPollEntity := abaputils.PollEntity(repo.Name, uriConnectionDetails, client, pollIntervall)
	if errorPollEntity != nil {
		return errors.Wrapf(errorPollEntity, errorString)
	}
	if status == "E" {
		return errors.New(errorString)
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

	jsonBody := []byte(repo.GetPullRequestBody())
	resp, err = abaputils.GetHTTPResponse("POST", pullConnectionDetails, jsonBody, client)
	if err != nil {
		err = abaputils.HandleHTTPError(resp, err, "Could not pull the "+repo.GetPullLogString(), uriConnectionDetails)
		return uriConnectionDetails, err
	}
	defer resp.Body.Close()
	log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", repo.Name).WithField("commitID", repo.CommitID).WithField("Tag", repo.Tag).Debug("Triggered Pull of repository / software component")

	// Parse Response
	var body abaputils.PullEntity
	var abapResp map[string]*json.RawMessage
	bodyText, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return uriConnectionDetails, err
	}
	if err := json.Unmarshal(bodyText, &abapResp); err != nil {
		return uriConnectionDetails, err
	}
	if err := json.Unmarshal(*abapResp["d"], &body); err != nil {
		return uriConnectionDetails, err
	}
	if reflect.DeepEqual(abaputils.PullEntity{}, body) {
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", repo.Name).WithField("commitID", repo.CommitID).WithField("Tag", repo.Tag).Error("Could not pull the repository / software component")
		err := errors.New("Request to ABAP System not successful")
		return uriConnectionDetails, err
	}

	uriConnectionDetails.URL = body.Metadata.URI
	return uriConnectionDetails, nil
}

func checkPullRepositoryConfiguration(options abapEnvironmentPullGitRepoOptions) error {

	if (len(options.RepositoryNames) > 0 && options.Repositories != "") || (len(options.RepositoryNames) > 0 && options.RepositoryName != "") || (options.RepositoryName != "" && options.Repositories != "") {
		return fmt.Errorf("Checking configuration failed: %w", errors.New("Only one of the paramters `RepositoryName`,`RepositoryNames` or `Repositories` may be configured at the same time"))
	}
	if len(options.RepositoryNames) == 0 && options.Repositories == "" && options.RepositoryName == "" {
		return fmt.Errorf("Checking configuration failed: %w", errors.New("You have not specified any repository configuration to be pulled into the ABAP Environment System. Please make sure that you specified the repositories that should be pulled either in a dedicated file or via the parameter 'repositoryNames'. For more information please read the User documentation"))
	}
	return nil
}

func finishPullLogs() {
	abaputils.AddDefaultDashedLine()
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
