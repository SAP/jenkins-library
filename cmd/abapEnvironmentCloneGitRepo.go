package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func abapEnvironmentCloneGitRepo(config abapEnvironmentCloneGitRepoOptions, _ *telemetry.CustomData) {

	c := command.Command{}

	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	var autils = abaputils.AbapUtils{
		Exec: &c,
	}

	client := piperhttp.Client{}
	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapEnvironmentCloneGitRepo(&config, &autils, &client)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentCloneGitRepo(config *abapEnvironmentCloneGitRepoOptions, com abaputils.Communication, client piperhttp.Sender) error {
	// Mapping for options
	subOptions := convertCloneConfig(config)

	// Determine the host, user and password, either via the input parameters or via a cloud foundry service key
	connectionDetails, errorGetInfo := com.GetAbapCommunicationArrangementInfo(subOptions, "")
	if errorGetInfo != nil {
		return errors.Wrap(errorGetInfo, "Parameters for the ABAP Connection not available")
	}

	// Configuring the HTTP Client and CookieJar
	cookieJar, errorCookieJar := cookiejar.New(nil)
	if errorCookieJar != nil {
		return errors.Wrap(errorCookieJar, "Could not create a Cookie Jar")
	}

	client.SetOptions(piperhttp.ClientOptions{
		MaxRequestDuration: 180 * time.Second,
		CookieJar:          cookieJar,
		Username:           connectionDetails.User,
		Password:           connectionDetails.Password,
	})

	errConfig := checkConfiguration(config)
	if errConfig != nil {
		return errors.Wrap(errConfig, "The provided configuration is not allowed")
	}

	repositories, errGetRepos := abaputils.GetRepositories(&abaputils.RepositoriesConfig{BranchName: config.BranchName, RepositoryName: config.RepositoryName, Repositories: config.Repositories}, true)
	if errGetRepos != nil {
		return fmt.Errorf("Something failed during the clone: %w", errGetRepos)
	}

	log.Entry().Infof("Start cloning %v repositories", len(repositories))
	for _, repo := range repositories {

		logString := repo.GetCloneLogString()
		errorString := "Clone of " + logString + " failed on the ABAP system"

		abaputils.AddDefaultDashedLine()
		log.Entry().Info("Start cloning " + logString)
		abaputils.AddDefaultDashedLine()

		// Triggering the Clone of the repository into the ABAP Environment system
		uriConnectionDetails, errorTriggerClone, didCheckoutPullInstead := triggerClone(repo, connectionDetails, client)
		if errorTriggerClone != nil {
			return errors.Wrapf(errorTriggerClone, errorString)
		}

		if !didCheckoutPullInstead {
			// Polling the status of the repository import on the ABAP Environment system
			// If the repository had been cloned already, as checkout/pull has been done - polling the status is not necessary anymore
			status, errorPollEntity := abaputils.PollEntity(repo.Name, uriConnectionDetails, client, com.GetPollIntervall())
			if errorPollEntity != nil {
				return errors.Wrapf(errorPollEntity, errorString)
			}
			if status == "E" {
				return errors.New("Clone of " + logString + " failed on the ABAP System")
			}
			log.Entry().Info("The " + logString + " was cloned successfully")
		}
	}
	abaputils.AddDefaultDashedLine()
	log.Entry().Info("All repositories were cloned successfully")
	return nil
}

func checkConfiguration(config *abapEnvironmentCloneGitRepoOptions) error {
	if config.Repositories != "" && config.RepositoryName != "" {
		return errors.New("It is not allowed to configure the parameters `repositories`and `repositoryName` at the same time")
	}
	if config.Repositories == "" && config.RepositoryName == "" {
		return errors.New("Please provide one of the following parameters: `repositories` or `repositoryName`")
	}
	return nil
}

func triggerClone(repo abaputils.Repository, cloneConnectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) (abaputils.ConnectionDetailsHTTP, error, bool) {

	uriConnectionDetails := cloneConnectionDetails
	cloneConnectionDetails.XCsrfToken = "fetch"

	cloneConnectionDetails.URL = cloneConnectionDetails.URL + "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Clones"

	// Loging into the ABAP System - getting the x-csrf-token and cookies
	resp, err := abaputils.GetHTTPResponse("HEAD", cloneConnectionDetails, nil, client)
	if err != nil {
		err = abaputils.HandleHTTPError(resp, err, "Authentication on the ABAP system failed", cloneConnectionDetails)
		return uriConnectionDetails, err, false
	}
	defer resp.Body.Close()

	log.Entry().WithField("StatusCode", resp.Status).WithField("ABAP Endpoint", cloneConnectionDetails.URL).Debug("Authentication on the ABAP system successful")
	uriConnectionDetails.XCsrfToken = resp.Header.Get("X-Csrf-Token")
	cloneConnectionDetails.XCsrfToken = uriConnectionDetails.XCsrfToken

	// Trigger the Clone of a Repository
	if repo.Name == "" {
		return uriConnectionDetails, errors.New("An empty string was passed for the parameter 'repositoryName'"), false
	}

	jsonBody := []byte(repo.GetCloneRequestBody())
	resp, err = abaputils.GetHTTPResponse("POST", cloneConnectionDetails, jsonBody, client)
	if err != nil {
		err, alreadyCloned := handleCloneError(resp, err, cloneConnectionDetails, client, repo)
		return uriConnectionDetails, err, alreadyCloned
	}
	defer resp.Body.Close()
	log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", repo.Name).WithField("branchName", repo.Branch).WithField("commitID", repo.CommitID).WithField("Tag", repo.Tag).Info("Triggered Clone of Repository / Software Component")

	// Parse Response
	var body abaputils.CloneEntity
	var abapResp map[string]*json.RawMessage
	bodyText, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return uriConnectionDetails, err, false
	}
	if err := json.Unmarshal(bodyText, &abapResp); err != nil {
		return uriConnectionDetails, err, false
	}
	if err := json.Unmarshal(*abapResp["d"], &body); err != nil {
		return uriConnectionDetails, err, false
	}
	if reflect.DeepEqual(abaputils.CloneEntity{}, body) {
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", repo.Name).WithField("branchName", repo.Branch).WithField("commitID", repo.CommitID).WithField("Tag", repo.Tag).Error("Could not Clone the Repository / Software Component")
		err := errors.New("Request to ABAP System not successful")
		return uriConnectionDetails, err, false
	}

	// The entity "Clones" does not allow for polling. To poll the progress, the related entity "Pull" has to be called
	// While "Clones" has the key fields UUID, SC_NAME and BRANCH_NAME, "Pull" only has the key field UUID
	uriConnectionDetails.URL = uriConnectionDetails.URL + "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull(uuid=guid'" + body.UUID + "')"
	return uriConnectionDetails, nil, false
}

func handleCloneError(resp *http.Response, err error, cloneConnectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender, repo abaputils.Repository) (returnedError error, alreadyCloned bool) {
	alreadyCloned = false
	returnedError = nil
	if resp == nil {
		log.Entry().WithError(err).WithField("ABAP Endpoint", cloneConnectionDetails.URL).Error("Request failed")
		returnedError = errors.New("Response is nil")
		return
	}
	defer resp.Body.Close()
	errorText, errorCode, parsingError := abaputils.GetErrorDetailsFromResponse(resp)
	if parsingError != nil {
		returnedError = err
		return
	}
	if errorCode == "A4C_A2G/257" {
		// With the latest release, a repeated "clone" was prohibited
		// As an intermediate workaround, we react to the error message A4C_A2G/257 that gets thrown, if the repository had already been cloned
		// In this case, a checkout branch and a pull will be performed
		alreadyCloned = true
		abaputils.AddDefaultDashedLine()
		abaputils.AddDefaultDashedLine()
		log.Entry().Infof("%s", "The repository / software component has already been cloned on the ABAP Environment system ")
		log.Entry().Infof("%s", "A `checkout branch` and a `pull` will be performed instead")
		abaputils.AddDefaultDashedLine()
		abaputils.AddDefaultDashedLine()
		checkoutOptions := abapEnvironmentCheckoutBranchOptions{
			Username:       cloneConnectionDetails.User,
			Password:       cloneConnectionDetails.Password,
			Host:           cloneConnectionDetails.Host,
			RepositoryName: repo.Name,
			BranchName:     repo.Branch,
		}
		c := command.Command{}
		c.Stdout(log.Writer())
		c.Stderr(log.Writer())
		com := abaputils.AbapUtils{
			Exec: &c,
		}
		returnedError = runAbapEnvironmentCheckoutBranch(&checkoutOptions, &com, client)
		if returnedError != nil {
			return
		}
		abaputils.AddDefaultDashedLine()
		abaputils.AddDefaultDashedLine()
		pullOptions := abapEnvironmentPullGitRepoOptions{
			Username:       cloneConnectionDetails.User,
			Password:       cloneConnectionDetails.Password,
			Host:           cloneConnectionDetails.Host,
			RepositoryName: repo.Name,
			CommitID:       repo.CommitID,
		}
		returnedError = runAbapEnvironmentPullGitRepo(&pullOptions, &com, client)
		if returnedError != nil {
			return
		}
	} else {
		log.Entry().WithField("StatusCode", resp.Status).Error("Could not clone the " + repo.GetCloneLogString())
		abapError := errors.New(fmt.Sprintf("%s - %s", errorCode, errorText))
		returnedError = errors.Wrap(abapError, err.Error())
	}
	return
}

func convertCloneConfig(config *abapEnvironmentCloneGitRepoOptions) abaputils.AbapEnvironmentOptions {
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
