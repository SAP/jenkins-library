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
	"github.com/ghodss/yaml"
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

	// for http calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapEnvironmentCheckoutBranch(&options, telemetryData, &autils, &client)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentCheckoutBranch(options *abapEnvironmentCheckoutBranchOptions, telemetryData *telemetry.CustomData, com abaputils.Communication, client piperhttp.Sender) (err error) {

	// Mapping for options
	subOptions := abaputils.AbapEnvironmentOptions{}

	subOptions.CfAPIEndpoint = options.CfAPIEndpoint
	subOptions.CfServiceInstance = options.CfServiceInstance
	subOptions.CfServiceKeyName = options.CfServiceKeyName
	subOptions.CfOrg = options.CfOrg
	subOptions.CfSpace = options.CfSpace
	subOptions.Host = options.Host
	subOptions.Password = options.Password
	subOptions.Username = options.Username

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

	checkCheckoutBranchRepositoryConfiguration(*options)

	if len(options.RepositoryNamesFiles) > 0 {
		err = checkoutBranchesFromFileConfig(options.RepositoryNamesFiles, connectionDetails, client, pollIntervall)
	}
	if err == nil {
		err = checkoutBranchFromConfig(options, connectionDetails, client, pollIntervall)
	}
	if err != nil {
		return fmt.Errorf("Checking out branches failed : %w", err)
	}

	log.Entry().Info("-------------------------")
	log.Entry().Info("All branches were checked out successfully")
	return nil
}

func triggerCheckout(repositoryName string, branchName string, checkoutConnectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) (abaputils.ConnectionDetailsHTTP, error) {
	uriConnectionDetails := checkoutConnectionDetails
	uriConnectionDetails.URL = ""
	checkoutConnectionDetails.XCsrfToken = "fetch"

	if repositoryName == "" || branchName == "" {
		return uriConnectionDetails, fmt.Errorf("Failed to trigger checkout: %w", errors.New("Repository and Branch Configuration is empty. Please make sure that you have specified the correct values"))
	}

	// Loging into the ABAP System - getting the x-csrf-token and cookies
	resp, err := abaputils.GetHTTPResponse("HEAD", checkoutConnectionDetails, nil, client)
	if err != nil {
		err = abaputils.HandleHTTPError(resp, err, "Authentication on the ABAP system failed", checkoutConnectionDetails)
		return uriConnectionDetails, err
	}
	defer resp.Body.Close()
	log.Entry().WithField("StatusCode", resp.Status).WithField("ABAP Endpoint", checkoutConnectionDetails.URL).Info("Authentication on the ABAP system was successful")
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
	log.Entry().WithField("StatusCode", resp.StatusCode).WithField("repositoryName", repositoryName).WithField("branchName", branchName).Info("Triggered checkout of branch")

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
	if len(options.RepositoryNamesFiles) == 0 && options.RepositoryName == "" && options.BranchName == "" {
		return fmt.Errorf("Checking configuration failed: %w", errors.New("You have not specified any repository configuration to be pulled into the ABAP Environment System. Please make sure that you specified the repositories with their branches that should be pulled either in a dedicated file or via in-line configuration. For more information please read the User documentation"))
	}
	if len(options.RepositoryNamesFiles) > 0 && options.RepositoryName != "" && options.BranchName != "" {
		log.Entry().Info("It seems like you have specified both the repositories with their branches to be pulled as an in-line configuration as well as in the dedicated repositories configuration file.")
		log.Entry().Info("Please note that in this case the dedicated repositories configuration file will be handled with priority.")
	}
	if len(options.RepositoryNamesFiles) > 0 && ((options.RepositoryName == "") != (options.BranchName == "")) {
		log.Entry().Info("It seems like you have specified a dedicated repository configuration file but also an in-line configuration for the repository or branch to be pulled.")
		log.Entry().Info("Please note that in this case the dedicated repositories configuration file will be handled only.")
	}
	return nil
}

func checkoutBranchesFromFileConfig(repositoriesFilesConfig []string, checkoutConnectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender, pollIntervall time.Duration) (err error) {
	for _, v := range repositoriesFilesConfig {
		if fileExists(v) {
			fileContent, err := ioutil.ReadFile(v)
			if err != nil {
				return fmt.Errorf("Failed to read repository configuration file: %w", err)
			}
			var repositoriesFileConfig []abaputils.Repository
			var result []byte
			result, err = yaml.YAMLToJSON(fileContent)
			if err == nil {
				err = json.Unmarshal(result, &repositoriesFileConfig)
			}
			if err != nil {
				return fmt.Errorf("Failed to parse repository configuration file: %w", err)
			}
			if len(repositoriesFileConfig) == 0 {
				return fmt.Errorf("Failed to parse repository configuration file: %w", errors.New("Empty or wrong configuration file. Please make sure that you have correctly specified the branches in the repositories to be checked out"))
			}

			log.Entry().Infof("Start switch of branches in %v repositories", len(repositoriesFileConfig))
			for _, repositoryFileConfig := range repositoriesFileConfig {
				if reflect.DeepEqual(abaputils.Repository{}, repositoryFileConfig) {
					return fmt.Errorf("Failed to read repository configuration file: %w", errors.New("Eror in configuration file, most likely you have entered empty or wrong configuration values. Please make sure that you have correctly specified the branches in the repositories to be checked out"))
				}
				startCheckoutLogs(repositoryFileConfig.Branch, repositoryFileConfig.Name)

				// Triggering the Checkout of the repository into the ABAP Environment system
				uriConnectionDetails, err := triggerCheckout(repositoryFileConfig.Name, repositoryFileConfig.Branch, checkoutConnectionDetails, client)
				if err != nil {
					log.Entry().WithError(err).Fatal("Checkout of " + repositoryFileConfig.Branch + " for software component " + repositoryFileConfig.Name + " failed on the ABAP System")
				}

				// Polling the status of the repository import on the ABAP Environment system
				status, err := abaputils.PollEntity(repositoryFileConfig.Name, uriConnectionDetails, client, pollIntervall)
				if err != nil {
					log.Entry().WithError(err).Fatal("Status of checkout action on repository" + repositoryFileConfig.Name + " failed on the ABAP System")
				}
				if status == "E" {
					log.Entry().Fatal("Checkout of branch " + repositoryFileConfig.Branch + " failed on the ABAP System")
				}
				finishCheckoutLogs(repositoryFileConfig.Branch, repositoryFileConfig.Name)
			}
		} else {
			return fmt.Errorf("Failed to read repository configuration file: %w", errors.New(v+" is not a file or doesn't exist"))
		}
	}
	return err
}

func checkoutBranchFromConfig(options *abapEnvironmentCheckoutBranchOptions, checkoutConnectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender, pollIntervall time.Duration) (err error) {
	startCheckoutLogs(options.BranchName, options.RepositoryName)

	// Triggering the Checkout of the repository into the ABAP Environment system
	uriConnectionDetails, err := triggerCheckout(options.RepositoryName, options.BranchName, checkoutConnectionDetails, client)
	if err != nil {
		return fmt.Errorf("Checkout of "+options.BranchName+" for software component "+options.RepositoryName+" failed on the ABAP System: %w", err)
	}

	// Polling the status of the repository import on the ABAP Environment system
	status, err := abaputils.PollEntity(options.RepositoryName, uriConnectionDetails, client, pollIntervall)
	if err != nil {
		return fmt.Errorf("Status of checkout action on repository"+options.RepositoryName+" failed on the ABAP System: %w", err)
	}
	if status == "E" {
		return fmt.Errorf("Checkout of branch "+options.BranchName+" failed on the ABAP System: %w", err)
	}
	finishCheckoutLogs(options.BranchName, options.RepositoryName)

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
