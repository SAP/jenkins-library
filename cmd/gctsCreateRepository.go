package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func gctsCreateRepository(config gctsCreateRepositoryOptions, telemetryData *telemetry.CustomData) {
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
	err := createRepository(&config, telemetryData, &c, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func createRepository(config *gctsCreateRepositoryOptions, telemetryData *telemetry.CustomData, command execRunner, httpClient piperhttp.Sender) error {

	cookieJar, _ := cookiejar.New(nil)
	clientOptions := piperhttp.ClientOptions{
		CookieJar: cookieJar,
		Username:  config.Username,
		Password:  config.Password,
	}
	httpClient.SetOptions(clientOptions)

	type repoData struct {
		RID             string `json:"rid"`
		Name            string `json:"name"`
		Role            string `json:"role"`
		Type            string `json:"type"`
		VSID            string `json:"vsid"`
		GithubURLstring string `json:"url"`
	}

	type createRequestBody struct {
		Repository string   `json:"repository"`
		Data       repoData `json:"data"`
	}

	type repoConfig struct {
		Key      string `json:"key"`
		Value    string `json:"value"`
		Category string `json:"category"`
	}

	type createResultBody struct {
		RID         string       `json:"rid"`
		Name        string       `json:"name"`
		Role        string       `json:"role"`
		Type        string       `json:"type"`
		VSID        string       `json:"vsid"`
		Status      string       `json:"status"`
		Branch      string       `json:"branch"`
		URL         string       `json:"url"`
		CreatedBy   string       `json:"createdBy"`
		CreatedDate string       `json:"createdDate"`
		Connection  string       `json:"connection"`
		Config      []repoConfig `json:"config"`
	}

	type createResponseBody struct {
		Repository createResultBody `json:"repository"`
		Exception  string           `json:"exception"`
	}

	reqBody := createRequestBody{
		Repository: config.Repository,
		Data: repoData{
			RID:             config.Repository,
			Name:            config.Repository,
			Role:            config.Role,
			Type:            config.Type,
			VSID:            config.VSID,
			GithubURLstring: config.RemoteRepositoryURL,
		},
	}
	jsonBody, marshalErr := json.Marshal(reqBody)

	if marshalErr != nil {
		return fmt.Errorf("creating repository on the ABAP system %v failed: %w", config.Host, marshalErr)
	}

	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	header.Add("Accept", "application/json")

	url := "http://" + config.Host +
		"/sap/bc/cts_abapvcs/repository?sap-client=" + config.Client

	resp, httpErr := httpClient.SendRequest("POST", url, bytes.NewBuffer(jsonBody), header, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if resp == nil {
		return fmt.Errorf("creating repository on the ABAP system %v failed: %w", config.Host, httpErr)
	}

	var response createResponseBody
	parsingErr := parseHTTPResponseBodyJSON(resp, &response)

	if parsingErr != nil {
		fmt.Errorf("creating repository on the ABAP system %v failed: %w", config.Host, parsingErr)
	}

	if httpErr != nil {
		if resp.StatusCode == 500 && response.Exception == "Repository already exists" {
			log.Entry().
				WithField("repository", config.Repository).
				Infof("the repository already exists on the ABAP system %v", config.Host)
			return nil
		}
		return fmt.Errorf("creating repository on the ABAP system %v failed: %w", config.Host, httpErr)
	}

	log.Entry().
		WithField("repository", config.Repository).
		Infof("successfully created the repository on the ABAP system %v", config.Host)
	return nil
}

func parseHTTPResponseBodyJSON(resp *http.Response, response interface{}) error {
	if resp == nil {
		return fmt.Errorf("cannot parse HTTP response with value <nil>")
	}

	bodyText, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return fmt.Errorf("cannot read HTTP response body: %w", readErr)
	}

	marshalErr := json.Unmarshal(bodyText, &response)
	if marshalErr != nil {
		return fmt.Errorf("cannot parse HTTP response as JSON: %w", marshalErr)
	}

	return nil
}
