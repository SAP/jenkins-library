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

	// Determine the host, user and password, either via the input parameters or via a cloud foundry service key
	connectionDetails, err := com.GetAbapCommunicationArrangementInfo(subOptions, "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull")
	if err != nil {
		return errors.Wrap(err, "Parameters for the ABAP Connection not available")
	}

	// Configuring the HTTP Client and CookieJar

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

	checkPullRepositoryConfiguration(*options)

	if len(options.RepositoryNamesFiles) > 0 {
		err = pullReposFromFileConfig(options.RepositoryNamesFiles, connectionDetails, client, pollIntervall)
	}

	if len(options.RepositoryNames) > 0 && err == nil {
		err = pullReposFromConfig(options.RepositoryNames, connectionDetails, client, pollIntervall)
	}

	if err != nil {
		return fmt.Errorf("Something failed during the pull of the repositories: %w", err)
	}

	log.Entry().Info("-------------------------")
	log.Entry().Info("All repositories were pulled successfully")
	return nil
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
	log.Entry().WithField("StatusCode", resp.Status).WithField("ABAP Endpoint", pullConnectionDetails.URL).Info("Authentication on the ABAP system successfull")
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
	if len(options.RepositoryNames) > 0 && len(options.RepositoryNamesFiles) > 0 {
		log.Entry().Info("It seems like you have specified both the repositories to be pulled as an in-line configuration as well as in the dedicated repositories configuration file.")
		log.Entry().Info("Please note that in this case the dedicated repositories configuration file will be handled with priority.")
	}
	if len(options.RepositoryNames) == 0 && len(options.RepositoryNamesFiles) == 0 {
		return fmt.Errorf("Checking configuration failed: %w", errors.New("You have not specified any repository configuration to be pulled into the ABAP Environment System. Please make sure that you specified the repositories that should be pulled either in a dedicated file or via in-line configuration. For more information please read the User documentation"))
	}
	return nil
}

func pullReposFromFileConfig(repositoriesFilesConfig []string, pullConnectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender, pollIntervall time.Duration) (err error) {
	for _, v := range repositoriesFilesConfig {
		if fileExists(v) {
			fileContent, err := ioutil.ReadFile(v)
			if err != nil {
				return fmt.Errorf("Failed to read repository configuration file : %w", err)
			}
			var repositoriesFileConfig []abaputils.Repository
			var result []byte
			result, err = yaml.YAMLToJSON(fileContent)
			if err == nil {
				err = json.Unmarshal(result, &repositoriesFileConfig)
			}
			if err != nil {
				return fmt.Errorf("Failed to parse repository configuration file : %w", err)
			}
			if len(repositoriesFileConfig) == 0 {
				return fmt.Errorf("Failed to parse repository configuration file: %w", errors.New("Empty or wrong configuration file. Please make sure that you have correctly specified the branches in the repositories to be pulled"))
			}

			//Iterating through each Repository to be pulled into the ABAP Environment system
			log.Entry().Infof("Start pulling %v repositories", len(repositoriesFileConfig))
			for _, repositoryFileConfig := range repositoriesFileConfig {
				if reflect.DeepEqual(abaputils.Repository{}, repositoryFileConfig) {
					return fmt.Errorf("Failed to read repository configuration file: %w", errors.New("Eror in configuration file, most likely you have entered empty or wrong configuration values. Please make sure that you have correctly specified the branches in the repositories to be pulled"))
				}
				startPullLogs(repositoryFileConfig.Name)
				// Triggering the Pull of the repository into the ABAP Environment system
				uriConnectionDetails, errorTriggerPull := triggerPull(repositoryFileConfig.Name, pullConnectionDetails, client)
				if errorTriggerPull != nil {
					return errors.Wrapf(errorTriggerPull, "Pull of '%s' failed on the ABAP System", repositoryFileConfig.Name)
				}

				// Polling the status of the repository import on the ABAP Environment system
				status, errorPollEntity := abaputils.PollEntity(repositoryFileConfig.Name, uriConnectionDetails, client, pollIntervall)
				if errorPollEntity != nil {
					return errors.Wrapf(errorPollEntity, "Pull of '%s' failed on the ABAP System", repositoryFileConfig.Name)
				}
				if status == "E" {
					return errors.New("Pull of " + repositoryFileConfig.Name + " failed on the ABAP System")
				}
				log.Entry().Info(repositoryFileConfig.Name + " was pulled successfully")
			}
			finishPullLogs()
		} else {
			return fmt.Errorf("Failed to read repository configuration file : %w", errors.New(v+" is not a file or doesn't exist"))
		}
	}
	return err
}

func pullReposFromConfig(repositoriesConfig []string, pullConnectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender, pollIntervall time.Duration) (err error) {
	if len(repositoriesConfig) == 0 {
		return fmt.Errorf("Failed to read repository configuration: %w", errors.New("Empty or wrong configuration values. Please make sure that you have correctly specified the branches in the repositories to be pulled"))
	}
	log.Entry().Infof("Start pulling %v repositories", len(repositoriesConfig))
	for _, repositoryName := range repositoriesConfig {
		startPullLogs(repositoryName)

		// Triggering the Pull of the repository into the ABAP Environment system
		uriConnectionDetails, errorTriggerPull := triggerPull(repositoryName, pullConnectionDetails, client)
		if errorTriggerPull != nil {
			return errors.Wrapf(errorTriggerPull, "Pull of '%s' failed on the ABAP System", repositoryName)

		}

		// Polling the status of the repository import on the ABAP Environment system
		status, errorPollEntity := abaputils.PollEntity(repositoryName, uriConnectionDetails, client, pollIntervall)
		if errorPollEntity != nil {
			return errors.Wrapf(errorPollEntity, "Pull of '%s' failed on the ABAP System", repositoryName)
		}
		if status == "E" {
			return errors.New("Pull of " + repositoryName + " failed on the ABAP System")
		}
		log.Entry().Info(repositoryName + " was pulled successfully")
	}
	finishPullLogs()
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

type repositoresConfiguration struct {
	RepositoryName string `json:"checkvariant,omitempty"`
	Configuration  string `json:"configuration,omitempty"`
}
