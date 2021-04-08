package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/Jeffail/gabs/v2"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func gctsDeploy(config gctsDeployOptions, telemetryData *telemetry.CustomData) {
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
	err := gctsDeployRepository(&config, telemetryData, &c, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func gctsDeployRepository(config *gctsDeployOptions, telemetryData *telemetry.CustomData, command command.ExecRunner, httpClient piperhttp.Sender) error {
	cookieJar, cookieErr := cookiejar.New(nil)
	const repoStateExists = "RepoExists"
	const repoStateNew = "RepoNew"
	repoState := repoStateExists
	branchRollbackRequired := false
	if cookieErr != nil {
		return errors.Wrap(cookieErr, "creating a cookie jar failed")
	}
	clientOptions := piperhttp.ClientOptions{
		CookieJar: cookieJar,
		Username:  config.Username,
		Password:  config.Password,
	}
	httpClient.SetOptions(clientOptions)

	configurationMetadata, getConfigMetadataErr := getConfigurationMetadata(config, httpClient)

	if getConfigMetadataErr != nil {
		log.Entry().WithError(getConfigMetadataErr).Error("step execution failed at configuration metadata retrieval. Please Check if system is up!.")
		return getConfigMetadataErr
	}

	createRepoOptions := gctsCreateRepositoryOptions{
		Username:            config.Username,
		Password:            config.Password,
		Repository:          config.Repository,
		Host:                config.Host,
		Client:              config.Client,
		RemoteRepositoryURL: config.RemoteRepositoryURL,
		Role:                config.Role,
		VSID:                config.VSID,
		Type:                config.Type,
	}
	log.Entry().Infof("gCTS Deploy : Checking if repository %v already exists", config.Repository)
	repoMetadataInitState, getRepositoryErr := getRepository(config, httpClient)
	currentBranch := repoMetadataInitState.Result.Branch
	if getRepositoryErr != nil {
		if config.Scope != "" {
			log.Entry().Error("Error during deploy : deploy scope cannot be provided while deploying a new repo")
			return errors.New("Error in config file")
		}
		repoState = repoStateNew
		log.Entry().Infof("gCTS Deploy : Creating Repository Step for repository : %v", config.Repository)
		configurations, _ := splitConfigurationToMap(config.Configuration, *configurationMetadata)
		createErr := createRepositoryForDeploy(&createRepoOptions, telemetryData, command, httpClient, configurations)
		if createErr != nil {
			//Dump error log (Log it)
			log.Entry().WithError(createErr).Error("step execution failed at Create Repository")
			return createErr
		}

		cloneRepoOptions := gctsCloneRepositoryOptions{
			Username:   config.Username,
			Password:   config.Password,
			Repository: config.Repository,
			Host:       config.Host,
			Client:     config.Client,
		}

		if config.Branch != "" || config.Commit != "" {
			log.Entry().Infof("Setting VCS_NO_IMPORT to true")
			noImportConfig := setConfigKeyBody{
				Key:   "VCS_NO_IMPORT",
				Value: "X",
			}
			setConfigKeyErr := setConfigKey(config, httpClient, &noImportConfig)
			if setConfigKeyErr != nil {
				log.Entry().WithError(setConfigKeyErr).Error("step execution failed at Set Config key for VCS_NO_IMPORT")
				return setConfigKeyErr
			}
			cloneErr := cloneRepository(&cloneRepoOptions, telemetryData, httpClient)

			if cloneErr != nil {
				log.Entry().WithError(cloneErr).Error("step execution failed at Clone Repository")
				return cloneErr
			}

		} else {
			cloneErr := cloneRepository(&cloneRepoOptions, telemetryData, httpClient)
			if cloneErr != nil {
				// Dump Error Log
				log.Entry().WithError(cloneErr).Error("step execution failed at Clone Repository")
				return cloneErr
			}
			log.Entry().Infof("gCTS Deploy : Step has completed for the repository %v : ", config.Repository)
			// End of the step.
			return nil
		}
		log.Entry().Infof("gCTS Deploy : Reading repo information after cloning repository %v : ", config.Repository)
		repoMetadataInitState, getRepositoryErr = getRepository(config, httpClient)
		if getRepositoryErr != nil {
			// Dump Error Log
			log.Entry().WithError(getRepositoryErr).Error("step execution failed at get repository after clone")
			return getRepositoryErr
		}
		currentBranch = repoMetadataInitState.Result.Branch
	} else {
		log.Entry().Infof("Repository %v already exists in the system, Checking for deploy scope", config.Repository)
		if config.Scope != "" {
			log.Entry().Infof("Deploy scope exists for the repository in the configuration file")
			log.Entry().Infof("gCTS Deploy: Deploying Commit to ABAP System for Repository %v with scope %v", config.Repository, config.Scope)
			deployErr := deployCommitToAbapSystem(config, httpClient)
			if deployErr != nil {
				log.Entry().WithError(deployErr).Error("step execution failed at Deploying Commit to ABAP system.")
				return deployErr
			}
			return nil
		}
		log.Entry().Infof("Deploy scope not set in the configuration file for repository : %v", config.Repository)
	}
	targetBranch := config.Branch
	if config.Branch != "" {
		response, switchBranchErr := switchBranch(config, httpClient, currentBranch, targetBranch)
		if switchBranchErr != nil {
			log.Entry().WithError(switchBranchErr).Error("step execution failed at Switch Branch")
			log.Entry().Errorf("Error Dump: ", response)
			if repoState == repoStateNew && config.Rollback {
				// Rollback branch. Resetting branches
				targetBranch = repoMetadataInitState.Result.Branch
				currentBranch = config.Branch
				log.Entry().WithError(switchBranchErr).Error("Rolling Back from %v to %v", currentBranch, targetBranch)
				switchBranch(config, httpClient, currentBranch, targetBranch)
			}
			return switchBranchErr
		}
		currentBranch = config.Branch
		branchRollbackRequired = true
	}

	if config.Commit != "" {
		log.Entry().Infof("gCTS Deploy: Pull by Commit step execution to commit %v", config.Commit)
		pullByCommitErr := pullByCommit(config, telemetryData, command, httpClient)
		if pullByCommitErr != nil {
			log.Entry().WithError(pullByCommitErr).Error("step execution failed at Pull By Commit. Trying to rollback to last commit")
			if config.Rollback {
				//Rollback to last commit.
				rollbackOptions := gctsRollbackOptions{
					Username:   config.Username,
					Password:   config.Password,
					Repository: config.Repository,
					Host:       config.Host,
					Client:     config.Client,
				}
				rollbackErr := rollback(&rollbackOptions, telemetryData, command, httpClient)
				if rollbackErr != nil {
					log.Entry().WithError(rollbackErr).Error("step execution failed while rolling back commit")
					return rollbackErr
				}
				if repoState == repoStateNew && branchRollbackRequired {
					// Rollback branch
					// Rollback branch. Resetting branches
					targetBranch = repoMetadataInitState.Result.Branch
					currentBranch = config.Branch
					log.Entry().Error("Rolling Back from %v to %v", currentBranch, targetBranch)
					switchBranch(config, httpClient, currentBranch, targetBranch)
				}
			}
			return pullByCommitErr
		}
	} else {
		if repoState == repoStateNew && (config.Commit != "" || config.Branch != "") {
			log.Entry().Infof("Setting deploy scope as current commit")
			config.Scope = "CRNTCOMMIT"
		}

		if config.Scope != "" {
			log.Entry().Infof("Removing VCS_NO_IMPORT configuration")
			configToDelete := "VCS_NO_IMPORT"
			deleteConfigKeyErr := deleteConfigKey(config, httpClient, configToDelete)
			if deleteConfigKeyErr != nil {
				log.Entry().WithError(deleteConfigKeyErr).Error("step execution failed at Set Config key for VCS_NO_IMPORT")
				return deleteConfigKeyErr
			}
			// Get deploy scope and gctsDeploy
			log.Entry().Infof("gCTS Deploy: Deploying Commit to ABAP System for Repository %v with scope %v", config.Repository, config.Scope)
			deployErr := deployCommitToAbapSystem(config, httpClient)
			if deployErr != nil {
				log.Entry().WithError(deployErr).Error("step execution failed at Deploying Commit to ABAP system.")
				return deployErr
			}
			// Execution Ends here for the step
			return nil
		}

		log.Entry().Infof("gCTS Deploy: Pull by Commit step execution")
		pullByCommitErr := pullByCommit(config, telemetryData, command, httpClient)
		if pullByCommitErr != nil {
			log.Entry().WithError(pullByCommitErr).Error("step execution failed at Pull By Commit. Trying to rollback to last commit")
			if config.Rollback {
				//Rollback to last commit.
				if repoState == repoStateNew {
					// Rollback branch
					targetBranch := repoMetadataInitState.Result.Branch
					log.Entry().Error("Rolling Back from %v to %v", currentBranch, targetBranch)
					switchBranch(config, httpClient, currentBranch, targetBranch)
				}
			}
			return pullByCommitErr
		}
	}

	if repoState == repoStateNew {
		log.Entry().Infof("Removing VCS_NO_IMPORT configuration")
		configToDelete := "VCS_NO_IMPORT"
		deleteConfigKeyErr := deleteConfigKey(config, httpClient, configToDelete)
		if deleteConfigKeyErr != nil {
			log.Entry().WithError(deleteConfigKeyErr).Error("step execution failed at Set Config key for VCS_NO_IMPORT")
			return deleteConfigKeyErr
		}
		log.Entry().Infof("Setting deploy scope as current commit")
		config.Scope = "CRNTCOMMIT"
		log.Entry().Infof("gCTS Deploy: Deploying Commit to ABAP System for Repository %v with scope %v", config.Repository, config.Scope)
		deployErr := deployCommitToAbapSystem(config, httpClient)
		if deployErr != nil {
			log.Entry().WithError(deployErr).Error("step execution failed at Deploying Commit to ABAP system.")
			return deployErr
		}
	}

	return nil
}

func switchBranch(config *gctsDeployOptions, httpClient piperhttp.Sender, currentBranch string, targetBranch string) (*switchBranchResponseBody, error) {
	var response switchBranchResponseBody
	log.Entry().Infof("gCTS Deploy : Switching branch for repository : %v, from branch: %v to %v", config.Repository, currentBranch, targetBranch)
	requestURL := config.Host +
		"/sap/bc/cts_abapvcs/repository/" + config.Repository + "/branches/" + currentBranch +
		"/switch?branch=" + targetBranch + "&sap-client=" + config.Client
	resp, httpErr := httpClient.SendRequest("GET", requestURL, nil, nil, nil)
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()
	if httpErr != nil {
		errorDump, errorDumpParseErr := parseErrorDumpFromResponseBody(resp)
		if errorDumpParseErr != nil {
			return nil, errorDumpParseErr
		}
		log.Entry().Errorf("Switch Branch Error Log: ", errorDump)
		return &response, httpErr
	} else if resp == nil {
		return &response, errors.New("did not retrieve a HTTP response")
	}
	parsingErr := piperhttp.ParseHTTPResponseBodyJSON(resp, &response)
	if parsingErr != nil {
		return &response, parsingErr
	}
	log.Entry().Infof("Switched branches from %v to %v. The commits where switched from %v to %v", currentBranch, config.Branch, response.Result.FromCommit, response.Result.ToCommit)
	return &response, nil
}

func deployCommitToAbapSystem(config *gctsDeployOptions, httpClient piperhttp.Sender) error {
	deployRequestBody := deployCommitToAbapSystemBody{
		Scope: config.Scope,
	}
	log.Entry().Info("gCTS Deploy : Start of deploying commit to ABAP System.")
	requestURL := config.Host +
		"/sap/bc/cts_abapvcs/repository/" + config.Repository +
		"/deploy?sap-client=" + config.Client
	reqBody := deployRequestBody
	jsonBody, marshalErr := json.Marshal(reqBody)
	if marshalErr != nil {
		return errors.Wrapf(marshalErr, "Deploying repository to abap system failed json body marshalling")
	}
	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	header.Add("Accept", "application/json")
	resp, httpErr := httpClient.SendRequest("POST", requestURL, bytes.NewBuffer(jsonBody), header, nil)
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()
	if httpErr != nil {
		log.Entry().Error("Failed During Deploy to Abap system")
		return httpErr
	}
	bodyText, readErr := ioutil.ReadAll(resp.Body)

	if readErr != nil {
		return errors.Wrapf(readErr, "HTTP response body could not be read")
	}

	response, parsingErr := gabs.ParseJSON([]byte(bodyText))

	if parsingErr != nil {
		return errors.Wrapf(parsingErr, "HTTP response body could not be parsed as JSON: %v", string(bodyText))
	}

	if response != nil {
		log.Entry().Infof("Response for deploy command : ", response.Path("result").Data().(map[string]interface{}))
	}
	return nil
}

func getRepository(config *gctsDeployOptions, httpClient piperhttp.Sender) (*getRepositoryResponseBody, error) {
	var response getRepositoryResponseBody
	requestURL := config.Host +
		"/sap/bc/cts_abapvcs/repository/" + config.Repository +
		"?sap-client=" + config.Client

	resp, httpErr := httpClient.SendRequest("GET", requestURL, nil, nil, nil)
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()
	if httpErr != nil {
		log.Entry().Infof("Error while repository Check : %v", httpErr)
		return &response, httpErr
	} else if resp == nil {
		return &response, errors.New("did not retrieve a HTTP response")
	}

	parsingErr := piperhttp.ParseHTTPResponseBodyJSON(resp, &response)
	if parsingErr != nil {
		return &response, parsingErr
	}
	return &response, nil
}

func deleteConfigKey(deployConfig *gctsDeployOptions, httpClient piperhttp.Sender, configToDelete string) error {
	log.Entry().Infof("gCTS Deploy : Delete configuration key %v", configToDelete)
	requestURL := deployConfig.Host +
		"/sap/bc/cts_abapvcs/repository/" + deployConfig.Repository +
		"/config/" + configToDelete + "?sap-client=" + deployConfig.Client
	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	header.Add("Accept", "application/json")
	resp, httpErr := httpClient.SendRequest("DELETE", requestURL, nil, header, nil)
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()
	if httpErr != nil {
		log.Entry().Error("Failure during deletion of configuration value")
		return httpErr
	}
	log.Entry().Infof("gCTS Deploy : Delete configuration key %v successful", configToDelete)
	return nil
}

func setConfigKey(deployConfig *gctsDeployOptions, httpClient piperhttp.Sender, configToSet *setConfigKeyBody) error {
	log.Entry().Infof("gCTS Deploy : Start of set configuration key %v and value %v", configToSet.Key, configToSet.Value)
	requestURL := deployConfig.Host +
		"/sap/bc/cts_abapvcs/repository/" + deployConfig.Repository +
		"/config?sap-client=" + deployConfig.Client

	reqBody := configToSet
	jsonBody, marshalErr := json.Marshal(reqBody)
	if marshalErr != nil {
		return errors.Wrapf(marshalErr, "Setting config key: %v and value: %v on the ABAP system %v failed", configToSet.Key, configToSet.Value, deployConfig.Host)
	}
	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	header.Add("Accept", "application/json")
	resp, httpErr := httpClient.SendRequest("POST", requestURL, bytes.NewBuffer(jsonBody), header, nil)
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()
	if httpErr != nil {
		log.Entry().Error("Failure during setting configuration value")
		return httpErr
	}
	log.Entry().
		WithField("repository", deployConfig.Repository).
		Infof("successfully set configuration value key %v and value %v", configToSet.Key, configToSet.Value)
	return nil
}

func pullByCommit(config *gctsDeployOptions, telemetryData *telemetry.CustomData, command command.ExecRunner, httpClient piperhttp.Sender) error {

	cookieJar, cookieErr := cookiejar.New(nil)
	if cookieErr != nil {
		return errors.Wrap(cookieErr, "creating a cookie jar failed")
	}
	clientOptions := piperhttp.ClientOptions{
		CookieJar: cookieJar,
		Username:  config.Username,
		Password:  config.Password,
	}
	httpClient.SetOptions(clientOptions)

	requestURL := config.Host +
		"/sap/bc/cts_abapvcs/repository/" + config.Repository +
		"/pullByCommit?sap-client=" + config.Client + "&request=" + config.Commit

	if config.Commit != "" {
		log.Entry().Infof("preparing to deploy specified commit %v", config.Commit)
		params := url.Values{}
		params.Add("request", config.Commit)
		requestURL = requestURL + "&" + params.Encode()
	}

	resp, httpErr := httpClient.SendRequest("GET", requestURL, nil, nil, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return httpErr
	} else if resp == nil {
		return errors.New("did not retrieve a HTTP response")
	}

	bodyText, readErr := ioutil.ReadAll(resp.Body)

	if readErr != nil {
		return errors.Wrapf(readErr, "HTTP response body could not be read")
	}

	response, parsingErr := gabs.ParseJSON([]byte(bodyText))

	if parsingErr != nil {
		return errors.Wrapf(parsingErr, "HTTP response body could not be parsed as JSON: %v", string(bodyText))
	}

	log.Entry().
		WithField("repository", config.Repository).
		Infof("successfully deployed commit %v (previous commit was %v)", response.Path("toCommit").Data().(string), response.Path("fromCommit").Data().(string))
	return nil
}

func createRepositoryForDeploy(config *gctsCreateRepositoryOptions, telemetryData *telemetry.CustomData, command command.ExecRunner, httpClient piperhttp.Sender, repositoryConfig []repositoryConfiguration) error {

	cookieJar, cookieErr := cookiejar.New(nil)
	if cookieErr != nil {
		return errors.Wrapf(cookieErr, "creating repository on the ABAP system %v failed", config.Host)
	}
	clientOptions := piperhttp.ClientOptions{
		CookieJar: cookieJar,
		Username:  config.Username,
		Password:  config.Password,
	}
	httpClient.SetOptions(clientOptions)

	type repoData struct {
		RID                 string                    `json:"rid"`
		Name                string                    `json:"name"`
		Role                string                    `json:"role"`
		Type                string                    `json:"type"`
		VSID                string                    `json:"vsid"`
		RemoteRepositoryURL string                    `json:"url"`
		Config              []repositoryConfiguration `json:"config"`
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
			Config:              repositoryConfig,
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

	resp, httpErr := httpClient.SendRequest("POST", url, bytes.NewBuffer(jsonBody), header, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if resp == nil {
		return errors.Errorf("creating repository on the ABAP system %v failed: %v", config.Host, httpErr)
	}

	bodyText, readErr := ioutil.ReadAll(resp.Body)

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

func getConfigurationMetadata(config *gctsDeployOptions, httpClient piperhttp.Sender) (*configurationMetadataBody, error) {
	var response configurationMetadataBody
	log.Entry().Infof("Starting to retrieve configuration metadata from the system")
	requestURL := config.Host +
		"/sap/bc/cts_abapvcs/config"

	resp, httpErr := httpClient.SendRequest("GET", requestURL, nil, nil, nil)
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()
	if httpErr != nil {
		log.Entry().Infof("Error while repository Check : %v", httpErr)
		return &response, httpErr
	} else if resp == nil {
		return &response, errors.New("did not retrieve a HTTP response")
	}

	parsingErr := piperhttp.ParseHTTPResponseBodyJSON(resp, &response)
	if parsingErr != nil {
		return &response, parsingErr
	}
	log.Entry().Infof("System Available for further step processing. The configuration metadata was successfully retrieved.")
	return &response, nil
}

func splitConfigurationToMap(inputConfigMap map[string]interface{}, configMetadataInSystem configurationMetadataBody) ([]repositoryConfiguration, error) {
	log.Entry().Infof("Parsing the configurations from the yml file")
	var configurations []repositoryConfiguration
	for key, value := range inputConfigMap {
		foundConfigMetadata, _ := findConfigurationMetadata(key, configMetadataInSystem)
		configValue := fmt.Sprint(value)
		if (configMetadata{}) != foundConfigMetadata {
			if foundConfigMetadata.Datatype == "BOOLEAN" && foundConfigMetadata.Example == "X" {
				if configValue == "false" || configValue == "" {
					configValue = ""
				} else if configValue == "true" || configValue == "X" {
					configValue = "X"
				}
			}
		}
		configuration := repositoryConfiguration{
			Key:   key,
			Value: configValue,
		}
		configurations = append(configurations, configuration)

	}
	log.Entry().Infof("The Configurations for the repoistory creation are : %v", configurations)
	return configurations, nil
}

func findConfigurationMetadata(configToFind string, configurationsAvailable configurationMetadataBody) (configMetadata, error) {
	var configStruct configMetadata
	for _, config := range configurationsAvailable.Config {
		if config.Ckey == configToFind {
			return config, nil
		}
	}
	return configStruct, nil
}

func parseErrorDumpFromResponseBody(responseBody *http.Response) (*errorLogBody, error) {
	var errorDump errorLogBody
	parsingErr := piperhttp.ParseHTTPResponseBodyJSON(responseBody, &errorDump)
	if parsingErr != nil {
		return &errorDump, parsingErr
	}
	return &errorDump, nil
}

type repositoryConfiguration struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type getRepositoryResponseBody struct {
	Result struct {
		Rid    string `json:"rid"`
		Name   string `json:"name"`
		Role   string `json:"role"`
		Vsid   string `json:"vsid"`
		Status string `json:"status"`
		Branch string `json:"branch"`
		Url    string `json:"url"`
		Config []struct {
			Key      string `json:"key"`
			Value    string `json:"value"`
			Category string `json:"category"`
		} `json:"config"`
		Objects       int64  `json:"objects"`
		CurrentCommit string `json:"currentCommit"`
		Connection    string `json:"connection"`
	} `json:"result"`
}

type setConfigKeyBody struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type switchBranchResponseBody struct {
	Result struct {
		FromCommit string `json:"fromCommit"`
		ToCommit   string `json:"ToCommit"`
	} `json:"result"`
	Log []struct {
		Time     string `json:"time"`
		User     string `json:"user"`
		Section  string `json:"section"`
		Action   string `json:"Action"`
		Severity string `json:"Severity"`
		Message  string `json:"Message"`
	} `json:"log"`
}

type deployCommitToAbapSystemBody struct {
	Repository string `json:"repository"`
	Scope      string `json:"scope"`
	Commit     string `json:"commit"`
	Objects    []struct {
		Object string `json:"object"`
		Type   string `json:"type"`
		User   string `json:"user"`
		Pgmid  string `json:"pgmid"`
		Keys   []struct {
			Tabname string `json:"tabname"`
			Columns []struct {
				Key     string `json:"key"`
				Field   string `json:"field"`
				Value   string `json:"value"`
				Type    string `json:"type"`
				Inttype string `json:"inttype"`
				Length  string `json:"length"`
			}
		}
	} `json:"objects"`
}

type configMetadata struct {
	Ckey         string `json:"ckey"`
	Ctype        string `json:"ctype"`
	Cvisible     string `json:"cvisible"`
	Datatype     string `json:"datatype"`
	DefaultValue string `json:"defaultValue"`
	Description  string `json:"description"`
	Category     string `json:"category"`
	UiElement    string `json:"uiElement"`
	Example      string `json:"example"`
}

type configurationMetadataBody struct {
	Config []configMetadata `json:"config"`
}

type errorProtocolbody struct {
	Type     string   `json:"type"`
	Protocol []string `json:"protocol"`
}

type errorLog struct {
	Time     int                 `json:"time"`
	User     string              `json:"user"`
	Section  string              `json:"section"`
	Action   string              `json:"action"`
	Severity string              `json:"severity"`
	Message  string              `json:"message"`
	Protocol []errorProtocolbody `json:"protocol"`
}

type errorLogBody struct {
	ErrorLog []errorLog `json:"errorLog"`
}
