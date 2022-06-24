package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	repositories, errGetRepos := abaputils.GetRepositories(&abaputils.RepositoriesConfig{BranchName: config.BranchName, RepositoryName: config.RepositoryName, Repositories: config.Repositories})
	if errGetRepos != nil {
		return fmt.Errorf("Something failed during the clone: %w", errGetRepos)
	}

	log.Entry().Infof("Start cloning %v repositories", len(repositories))
	for _, repo := range repositories {

		logString := repo.GetCloneLogString()
		errorString := "Clone of " + logString + " failed on the ABAP system"

		log.Entry().Info("-------------------------")
		log.Entry().Info("Start cloning " + logString)
		log.Entry().Info("-------------------------")

		// Triggering the Clone of the repository into the ABAP Environment system
		uriConnectionDetails, errorTriggerClone, checkoutPullInstead := triggerClone(repo, connectionDetails, client)
		if errorTriggerClone != nil || checkoutPullInstead {
			return errors.Wrapf(errorTriggerClone, errorString)
		}

		// Polling the status of the repository import on the ABAP Environment system
		status, errorPollEntity := abaputils.PollEntity(repo.Name, uriConnectionDetails, client, com.GetPollIntervall())
		if errorPollEntity != nil {
			return errors.Wrapf(errorPollEntity, errorString)
		}
		if status == "E" {
			return errors.New("Clone of " + logString + " failed on the ABAP System")
		}

		log.Entry().Info("The " + logString + " was cloned successfully")
	}
	log.Entry().Info("-------------------------")
	log.Entry().Info("All repositories were cloned successfully")
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
		err, alreadyCloned := handleAlreadyCloned(resp, err, cloneConnectionDetails, client, repo)
		if !alreadyCloned {
			err = abaputils.HandleHTTPError(resp, err, "Could not clone the "+repo.GetCloneLogString(), uriConnectionDetails)
		}
		return uriConnectionDetails, err, true
	}
	defer resp.Body.Close()
	log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", repo.Name).WithField("branchName", repo.Branch).WithField("commitID", repo.CommitID).WithField("Tag", repo.Tag).Info("Triggered Clone of Repository / Software Component")

	// Parse Response
	var body abaputils.CloneEntity
	var abapResp map[string]*json.RawMessage
	bodyText, errRead := ioutil.ReadAll(resp.Body)
	if errRead != nil {
		return uriConnectionDetails, err, false
	}
	json.Unmarshal(bodyText, &abapResp)
	json.Unmarshal(*abapResp["d"], &body)
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

func handleAlreadyCloned(resp *http.Response, err error, cloneConnectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender, repo abaputils.Repository) (returnedError error, alreadyCloned bool) {
	alreadyCloned = false
	returnedError = nil
	if resp == nil {
		returnedError = errors.New("Response is nil")
		return
	}
	_, errorCode, parsingError := abaputils.GetErrorDetailsFromResponse(resp)
	if parsingError != nil {
		returnedError = errors.New("Could not parse error")
		return
	}
	if errorCode == "A4C_A2G/257" {
		alreadyCloned = true
		log.Entry().Infof("-------------------------")
		log.Entry().Infof("-------------------------")
		log.Entry().Infof("%s", "The repository / software component has already been cloned on the ABAP Environment system ")
		log.Entry().Infof("%s", "A `checkout branch` and a `pull` will be performed instead")
		log.Entry().Infof("-------------------------")
		log.Entry().Infof("-------------------------")
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
		log.Entry().Infof("-------------------------")
		log.Entry().Infof("-------------------------")
		pullOptions := *&abapEnvironmentPullGitRepoOptions{
			Username:        cloneConnectionDetails.User,
			Password:        cloneConnectionDetails.Password,
			Host:            cloneConnectionDetails.Host,
			RepositoryNames: []string{repo.Name},
			// CommitdID
		}
		returnedError = runAbapEnvironmentPullGitRepo(&pullOptions, &com, client)
		if returnedError != nil {
			return
		}
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
