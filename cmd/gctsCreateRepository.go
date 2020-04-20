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
		Repository: config.RepositoryName,
		Data: repoData{
			RID:             config.RepositoryName,
			Name:            config.RepositoryName,
			Role:            config.Role,
			Type:            config.Type,
			VSID:            config.VSID,
			GithubURLstring: config.GithubURL,
		},
	}
	jsonBody, marshalErr := json.Marshal(reqBody)

	if marshalErr != nil {
		return fmt.Errorf("creating the repository locally failed: %w", marshalErr)
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
		return fmt.Errorf("creating the repository locally failed: %w", httpErr)
	}

	var response createResponseBody
	parsingErr := parseHTTPResponseBodyJSON(resp, &response)

	if parsingErr != nil {
		log.Entry().Warning(parsingErr)
	}

	if httpErr != nil {
		if resp.StatusCode == 500 && response.Exception == "Repository already exists" {
			log.Entry().
				WithField("repositoryName", config.RepositoryName).
				Info("the repository already exists locally")
			return nil
		}
		return fmt.Errorf("creating the repository locally failed: %w", httpErr)
	}

	log.Entry().
		WithField("repositoryName", config.RepositoryName).
		Info("successfully created the local repository")
	return nil
}

func parseHTTPResponseBodyJSON(resp *http.Response, response interface{}) error {
	if resp == nil {
		return fmt.Errorf("cannot parse HTTP response with value <nil>")
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read HTTP response body: %w", err)
	}
	json.Unmarshal(bodyText, &response)

	return nil
}
