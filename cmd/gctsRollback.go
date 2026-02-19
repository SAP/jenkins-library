package cmd

import (
	"fmt"
	"io"
	"net/http/cookiejar"
	"net/url"

	"errors"

	"github.com/Jeffail/gabs/v2"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/gcts"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func gctsRollback(config gctsRollbackOptions, telemetryData *telemetry.CustomData) {
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
	err := rollback(&config, telemetryData, &c, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func rollback(config *gctsRollbackOptions, telemetryData *telemetry.CustomData, command command.ExecRunner, httpClient piperhttp.Sender) error {

	clientOptions, err := gcts.NewHttpClientOptions(config.Username, config.Password, config.Proxy, config.SkipSSLVerification)
	if err != nil {
		return err
	}
	httpClient.SetOptions(clientOptions)

	repoInfo, err := getRepoInfo(config, telemetryData, httpClient)
	if err != nil {
		return fmt.Errorf("could not get local repository data: %w", err)
	}

	if repoInfo.Result.URL == "" {
		return fmt.Errorf("no remote repository URL configured")
	}

	if err != nil {
		return fmt.Errorf("could not parse remote repository URL as valid URL: %w", err)
	}

	var deployOptions gctsDeployOptions

	if config.Commit != "" {
		log.Entry().Infof("rolling back to specified commit %v", config.Commit)

		deployOptions = gctsDeployOptions{
			Username:            config.Username,
			Password:            config.Password,
			Host:                config.Host,
			Repository:          config.Repository,
			Client:              config.Client,
			Commit:              config.Commit,
			SkipSSLVerification: config.SkipSSLVerification,
		}

	} else {
		repoHistory, err := getRepoHistory(config, telemetryData, httpClient)
		if err != nil {
			return fmt.Errorf("could not retrieve repository commit history: %w", err)
		}
		if repoHistory.Result[0].FromCommit != "" {

			log.Entry().WithField("repository", config.Repository).Infof("rolling back to last active commit %v", repoHistory.Result[0].FromCommit)
			deployOptions = gctsDeployOptions{
				Username:            config.Username,
				Password:            config.Password,
				Host:                config.Host,
				Repository:          config.Repository,
				Client:              config.Client,
				Commit:              repoHistory.Result[0].FromCommit,
				SkipSSLVerification: config.SkipSSLVerification,
			}

		} else {
			return fmt.Errorf("no commit to rollback to (fromCommit) could be identified from the repository commit history")
		}
	}

	deployErr := pullByCommit(&deployOptions, telemetryData, command, httpClient)

	if deployErr != nil {
		return fmt.Errorf("rollback commit failed: %w", deployErr)
	}

	log.Entry().
		WithField("repository", config.Repository).
		Infof("rollback was successful")
	return nil
}

func getLastSuccessfullCommit(config *gctsRollbackOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender, githubURL *url.URL, commitList []string) (string, error) {

	cookieJar, cookieErr := cookiejar.New(nil)
	if cookieErr != nil {
		return "", cookieErr
	}
	clientOptions := piperhttp.ClientOptions{
		CookieJar: cookieJar,
	}

	if config.GithubPersonalAccessToken != "" {
		clientOptions.Token = "Bearer " + config.GithubPersonalAccessToken
	} else {
		log.Entry().Warning("no GitHub personal access token was provided")
	}
	// Add proxy support if configured
	if config.Proxy != "" {
		proxyURL, err := url.Parse(config.Proxy)
		if err != nil {
			return "", fmt.Errorf("failed to parse proxy URL: %w", err)
		}
		clientOptions.TransportProxy = proxyURL
		log.Entry().Infof("Using proxy: %v", config.Proxy)
	}

	httpClient.SetOptions(clientOptions)

	for _, commit := range commitList {

		url := githubURL.Scheme + "://api." + githubURL.Host + "/repos" + githubURL.Path + "/commits/" + commit + "/status"

		url, urlErr := addQueryToURL(url, config.QueryParameters)

		if urlErr != nil {

			return "", urlErr
		}

		resp, httpErr := httpClient.SendRequest("GET", url, nil, nil, nil)

		defer func() {
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
		}()

		if httpErr != nil {
			return "", httpErr
		} else if resp == nil {
			return "", errors.New("did not retrieve a HTTP response")
		}

		bodyText, readErr := io.ReadAll(resp.Body)

		if readErr != nil {
			return "", fmt.Errorf("HTTP response body could not be read: %w", readErr)
		}

		response, parsingErr := gabs.ParseJSON([]byte(bodyText))

		if parsingErr != nil {
			return "", fmt.Errorf("HTTP response body could not be parsed as JSON: %v: %w", string(bodyText), parsingErr)
		}

		if status, ok := response.Path("state").Data().(string); ok && status == "success" {
			log.Entry().
				WithField("repository", config.Repository).
				Infof("last successful commit was determined to be %v", commit)
			return commit, nil
		}
	}

	return "", fmt.Errorf("no commit with status 'success' could be found")
}

func getCommits(config *gctsRollbackOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender) ([]string, error) {

	url := config.Host +
		"/sap/bc/cts_abapvcs/repository/" + config.Repository +
		"/getCommit?sap-client=" + config.Client

	url, urlErr := addQueryToURL(url, config.QueryParameters)

	if urlErr != nil {

		return nil, urlErr
	}

	type commitsResponseBody struct {
		Commits []struct {
			ID          string `json:"id"`
			Author      string `json:"author"`
			AuthorMail  string `json:"authorMail"`
			Message     string `json:"message"`
			Description string `json:"description"`
			Date        string `json:"date"`
		} `json:"commits"`
	}

	resp, httpErr := httpClient.SendRequest("GET", url, nil, nil, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return []string{}, httpErr
	} else if resp == nil {
		return []string{}, errors.New("did not retrieve a HTTP response")
	}

	var response commitsResponseBody
	parsingErr := piperhttp.ParseHTTPResponseBodyJSON(resp, &response)
	if parsingErr != nil {
		return []string{}, parsingErr
	}

	commitList := []string{}
	for _, commit := range response.Commits {
		commitList = append(commitList, commit.ID)
	}

	return commitList, nil
}

func getRepoInfo(config *gctsRollbackOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender) (*getRepoInfoResponseBody, error) {

	var response getRepoInfoResponseBody

	url := config.Host +
		"/sap/bc/cts_abapvcs/repository/" + config.Repository +
		"?sap-client=" + config.Client

	url, urlErr := addQueryToURL(url, config.QueryParameters)

	if urlErr != nil {

		return nil, urlErr
	}

	resp, httpErr := httpClient.SendRequest("GET", url, nil, nil, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
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

type getRepoInfoResponseBody struct {
	Result struct {
		Rid           string `json:"rid"`
		Name          string `json:"name"`
		Role          string `json:"role"`
		Type          string `json:"type"`
		Vsid          string `json:"vsid"`
		Status        string `json:"status"`
		Branch        string `json:"branch"`
		URL           string `json:"url"`
		Version       string `json:"version"`
		Objects       any    `json:"objects"`
		CurrentCommit string `json:"currentCommit"`
		Connection    string `json:"connection"`
		Config        []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"config"`
	} `json:"result"`
}

func getRepoHistory(config *gctsRollbackOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender) (*getRepoHistoryResponseBody, error) {

	var response getRepoHistoryResponseBody

	url := config.Host +
		"/sap/bc/cts_abapvcs/repository/" + config.Repository +
		"/getHistory?sap-client=" + config.Client

	url, urlErr := addQueryToURL(url, config.QueryParameters)

	if urlErr != nil {

		return nil, urlErr
	}

	resp, httpErr := httpClient.SendRequest("GET", url, nil, nil, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
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

type getRepoHistoryResponseBody struct {
	Result []struct {
		Rid          string `json:"rid"`
		CheckoutTime int64  `json:"checkoutTime"`
		FromCommit   string `json:"fromCommit"`
		ToCommit     string `json:"toCommit"`
		Caller       string `json:"caller"`
		Request      string `json:"request"`
		Type         string `json:"type"`
	} `json:"result"`
}
