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

func abapEnvironmentCheckoutBranch(options abapEnvironmentCheckoutBranchOptions, telemetryData *telemetry.CustomData) {

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
	err := runAbapEnvironmentCheckoutBranch(&options, telemetryData, &autils, &client)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentCheckoutBranch(options *abapEnvironmentCheckoutBranchOptions, telemetryData *telemetry.CustomData, com abaputils.Communication, client piperhttp.Sender) (err error) {

	// Mapping for options
	subOptions := convertCheckoutConfig(options)

	//  Determine the host, user and password, either via the input parameters or via a cloud foundry service key
	connectionDetails, errorGetInfo := com.GetAbapCommunicationArrangementInfo(subOptions, "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/")
	if errorGetInfo != nil {
		log.Entry().WithError(errorGetInfo).Fatal("Parameters for the ABAP Connection not available")
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
	pollIntervall := com.GetPollIntervall()

	repositories := []abaputils.Repository{}
	err = checkCheckoutBranchRepositoryConfiguration(*options)

	if err == nil {
		repositories, err = abaputils.GetRepositories(&abaputils.RepositoriesConfig{BranchName: options.BranchName, RepositoryName: options.RepositoryName, Repositories: options.Repositories})
	}
	if err == nil {
		err = checkoutBranches(repositories, connectionDetails, client, pollIntervall)
	}
	if err != nil {
		return fmt.Errorf("Something failed during the checkout: %w", err)
	}
	log.Entry().Info("-------------------------")
	log.Entry().Info("All branches were checked out successfully")
	return nil
}

func checkoutBranches(repositories []abaputils.Repository, checkoutConnectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender, pollIntervall time.Duration) (err error) {
	log.Entry().Infof("Start switching %v branches", len(repositories))
	for _, repo := range repositories {
		err = handleCheckout(repo, checkoutConnectionDetails, client, pollIntervall)
		if err != nil {
			break
		}
		finishCheckoutLogs(repo.Branch, repo.Name)
	}
	return err
}

func triggerCheckout(repositoryName string, branchName string, checkoutConnectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) (abaputils.ConnectionDetailsHTTP, error) {
	uriConnectionDetails := checkoutConnectionDetails
	uriConnectionDetails.URL = ""
	checkoutConnectionDetails.XCsrfToken = "fetch"

	if repositoryName == "" || branchName == "" {
		return uriConnectionDetails, fmt.Errorf("Failed to trigger checkout: %w", errors.New("Repository and/or Branch Configuration is empty. Please make sure that you have specified the correct values"))
	}

	// Loging into the ABAP System - getting the x-csrf-token and cookies
	resp, err := abaputils.GetHTTPResponse("HEAD", checkoutConnectionDetails, nil, client)
	if err != nil {
		err = abaputils.HandleHTTPError(resp, err, "Authentication on the ABAP system failed", checkoutConnectionDetails)
		return uriConnectionDetails, err
	}
	defer resp.Body.Close()

	// workaround until golang version 1.16 is used
	time.Sleep(100 * time.Millisecond)

	log.Entry().WithField("StatusCode", resp.Status).WithField("ABAP Endpoint", checkoutConnectionDetails.URL).Debug("Authentication on the ABAP system was successful")
	uriConnectionDetails.XCsrfToken = resp.Header.Get("X-Csrf-Token")
	checkoutConnectionDetails.XCsrfToken = uriConnectionDetails.XCsrfToken

	// the request looks like: POST/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/checkout_branch?branch_name='newBranch'&sc_name=/DMO/GIT_REPOSITORY'
	checkoutConnectionDetails.URL = checkoutConnectionDetails.URL + `/checkout_branch?branch_name='` + branchName + `'&sc_name='` + repositoryName + `'`
	jsonBody := []byte(``)

	// no JSON body needed
	resp, err = abaputils.GetHTTPResponse("POST", checkoutConnectionDetails, jsonBody, client)
	if err != nil {
		err = abaputils.HandleHTTPError(resp, err, "Could not trigger checkout of branch "+branchName, uriConnectionDetails)
		return uriConnectionDetails, err
	}
	defer resp.Body.Close()
	log.Entry().WithField("StatusCode", resp.StatusCode).WithField("repositoryName", repositoryName).WithField("branchName", branchName).Debug("Triggered checkout of branch")

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
		log.Entry().WithField("StatusCode", resp.Status).WithField("branchName", branchName).Error("Could not switch to specified branch")
		err := errors.New("Request to ABAP System failed")
		return uriConnectionDetails, err
	}

	expandLog := "?$expand=to_Execution_log,to_Transport_log"
	uriConnectionDetails.URL = body.Metadata.URI + expandLog
	return uriConnectionDetails, nil
}

func checkCheckoutBranchRepositoryConfiguration(options abapEnvironmentCheckoutBranchOptions) error {
	if options.Repositories == "" && options.RepositoryName == "" && options.BranchName == "" {
		return fmt.Errorf("Checking configuration failed: %w", errors.New("You have not specified any repository or branch configuration to be checked out in the ABAP Environment System. Please make sure that you specified the repositories with their branches that should be checked out either in a dedicated file or via the parameters 'repositoryName' and 'branchName'. For more information please read the User documentation"))
	}
	if options.Repositories != "" && options.RepositoryName != "" && options.BranchName != "" {
		log.Entry().Info("It seems like you have specified repositories directly via the configuration parameters 'repositoryName' and 'branchName' as well as in the dedicated repositories configuration file. Please note that in this case both configurations will be handled and checked out.")
	}
	if options.Repositories != "" && ((options.RepositoryName == "") != (options.BranchName == "")) {
		log.Entry().Info("It seems like you have specified a dedicated repository configuration file but also a wrong configuration for the parameters 'repositoryName' and 'branchName' to be checked out.")
		if options.RepositoryName != "" {
			log.Entry().Info("Please also add the value for the branchName parameter or remove the repositoryName parameter.")
		} else {
			log.Entry().Info("Please also add the value for the repositoryName parameter or remove the branchName parameter.")
		}
	}
	return nil
}

func handleCheckout(repo abaputils.Repository, checkoutConnectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender, pollIntervall time.Duration) (err error) {
	if reflect.DeepEqual(abaputils.Repository{}, repo) {
		return fmt.Errorf("Failed to read repository configuration: %w", errors.New("Error in configuration, most likely you have entered empty or wrong configuration values. Please make sure that you have correctly specified the branches in the repositories to be checked out"))
	}
	startCheckoutLogs(repo.Branch, repo.Name)

	uriConnectionDetails, err := triggerCheckout(repo.Name, repo.Branch, checkoutConnectionDetails, client)
	if err != nil {
		return fmt.Errorf("Failed to trigger Checkout: %w", errors.New("Checkout of "+repo.Branch+" for software component "+repo.Name+" failed on the ABAP System"))
	}

	// Polling the status of the repository import on the ABAP Environment system
	status, err := abaputils.PollEntity(repo.Name, uriConnectionDetails, client, pollIntervall)
	if err != nil {
		return fmt.Errorf("Failed to poll Checkout: %w", errors.New("Status of checkout action on repository"+repo.Name+" failed on the ABAP System"))
	}
	const abapStatusCheckoutFail = "E"
	if status == abapStatusCheckoutFail {
		return fmt.Errorf("Checkout failed: %w", errors.New("Checkout of branch "+repo.Branch+" failed on the ABAP System"))
	}
	finishCheckoutLogs(repo.Branch, repo.Name)

	return err
}

func startCheckoutLogs(branchName string, repositoryName string) {
	log.Entry().Infof("Starting to switch branch to branch '%v' on repository '%v'", branchName, repositoryName)
	log.Entry().Info("--------------------------------")
	log.Entry().Info("Start checkout branch: " + branchName)
	log.Entry().Info("--------------------------------")
}

func finishCheckoutLogs(branchName string, repositoryName string) {
	log.Entry().Info("--------------------------------")
	log.Entry().Infof("Checkout of branch %v on repository %v was successful", branchName, repositoryName)
	log.Entry().Info("--------------------------------")
}

func convertCheckoutConfig(config *abapEnvironmentCheckoutBranchOptions) abaputils.AbapEnvironmentOptions {
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
