package cmd

import (
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"

	gabs "github.com/Jeffail/gabs/v2"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func gctsCloneRepository(config gctsCloneRepositoryOptions, telemetryData *telemetry.CustomData) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())

	// for http calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go
	httpClient := &piperhttp.Client{}

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := cloneRepository(&config, telemetryData, &c, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func cloneRepository(config *gctsCloneRepositoryOptions, telemetryData *telemetry.CustomData, command execRunner, httpClient piperhttp.Sender) error {

	cookieJar, cookieErr := cookiejar.New(nil)
	if cookieErr != nil {
		return errors.Wrap(cookieErr, "cloning the repository failed")
	}
	clientOptions := piperhttp.ClientOptions{
		CookieJar: cookieJar,
		Username:  config.Username,
		Password:  config.Password,
	}
	httpClient.SetOptions(clientOptions)

	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	header.Add("Accept", "application/json")

	url := config.Host +
		"/sap/bc/cts_abapvcs/repository/" + config.Repository +
		"/clone?sap-client=" + config.Client

	resp, httpErr := httpClient.SendRequest("POST", url, nil, header, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if resp == nil {
		return errors.Errorf("cloning the repository failed: %v", httpErr)
	}

	bodyText, readErr := ioutil.ReadAll(resp.Body)

	if readErr != nil {
		return errors.Wrap(readErr, "cloning the repository failed")
	}

	response, parsingErr := gabs.ParseJSON([]byte(bodyText))

	if parsingErr != nil {
		return errors.Wrap(parsingErr, "cloning the repository failed")
	}

	if httpErr != nil {
		if resp.StatusCode == 500 {
			if exception, ok := response.Path("errorLog.1.code").Data().(string); ok && exception == "GCTS.CLIENT.1420" {
				log.Entry().
					WithField("repository", config.Repository).
					Info("the repository has already been cloned")
				return nil
			} else if exception, ok := response.Path("errorLog.1.code").Data().(string); ok && exception == "GCTS.CLIENT.3302" {
				log.Entry().Errorf("%v", response.Path("errorLog.1.message").Data().(string))
				log.Entry().Error("possible reason: the remote repository is set to 'private'. you need to provide the local ABAP server repository with authentication credentials to the remote Git repository in order to clone it.")
				return errors.Wrap(httpErr, "cloning the repository failed")
			}
		}
		log.Entry().Errorf("a HTTP error occured! Response body: %v", response)
		return errors.Wrap(httpErr, "cloning the repository failed")
	}

	log.Entry().
		WithField("repository", config.Repository).
		Info("successfully cloned the Git repository to the local repository")
	return nil
}
