package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	gabs "github.com/Jeffail/gabs/v2"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/gcts"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func gctsCreateRepository(config gctsCreateRepositoryOptions, telemetryData *telemetry.CustomData) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	// for http calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go
	httpClient := &piperhttp.Client{}

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := createRepository(&config, telemetryData, &c, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func createRepository(config *gctsCreateRepositoryOptions, telemetryData *telemetry.CustomData, command command.ExecRunner, httpClient piperhttp.Sender) error {

	clientOptions, err := gcts.NewHttpClientOptions(config.Username, config.Password, config.Proxy, config.SkipSSLVerification)
	if err != nil {
		return errors.Wrapf(err, "creating repository on the ABAP system %v failed", config.Host)
	}
	httpClient.SetOptions(clientOptions)

	type repoData struct {
		RID                 string `json:"rid"`
		Name                string `json:"name"`
		Role                string `json:"role"`
		Type                string `json:"type"`
		VSID                string `json:"vsid"`
		RemoteRepositoryURL string `json:"url"`
	}

	type createRequestBody struct {
		Repository string   `json:"repository"`
		Data       repoData `json:"data"`
	}

	reqBody := createRequestBody{
		Repository: config.Repository,
		Data: repoData{
			RID:                 config.Repository,
			Name:                config.Repository,
			Role:                config.Role,
			Type:                config.Type,
			VSID:                config.VSID,
			RemoteRepositoryURL: config.RemoteRepositoryURL,
		},
	}
	jsonBody, marshalErr := json.Marshal(reqBody)

	if marshalErr != nil {
		return errors.Wrapf(marshalErr, "creating repository on the ABAP system %v failed", config.Host)
	}

	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	header.Add("Accept", "application/json")

	url := config.Host + "/sap/bc/cts_abapvcs/repository?sap-client=" + config.Client

	url, urlErr := addQueryToURL(url, config.QueryParameters)

	if urlErr != nil {

		return urlErr
	}

	resp, httpErr := httpClient.SendRequest("POST", url, bytes.NewBuffer(jsonBody), header, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if resp == nil {
		return errors.Errorf("creating repository on the ABAP system %v failed: %v", config.Host, httpErr)
	}

	bodyText, readErr := io.ReadAll(resp.Body)

	if readErr != nil {
		return errors.Wrapf(readErr, "creating repository on the ABAP system %v failed", config.Host)
	}

	response, parsingErr := gabs.ParseJSON([]byte(bodyText))

	if parsingErr != nil {
		return errors.Wrapf(parsingErr, "creating repository on the ABAP system %v failed", config.Host)
	}

	if httpErr != nil {
		if resp.StatusCode == 500 {
			if exception, ok := response.Path("exception").Data().(string); ok && exception == "Repository already exists" {
				log.Entry().
					WithField("repository", config.Repository).
					Infof("the repository already exists on the ABAP system %v", config.Host)
				return nil
			}
		}
		log.Entry().Errorf("a HTTP error occurred! Response body: %v", response)
		return errors.Wrapf(httpErr, "creating repository on the ABAP system %v failed", config.Host)
	}

	log.Entry().
		WithField("repository", config.Repository).
		Infof("successfully created the repository on ABAP system %v", config.Host)
	return nil
}
