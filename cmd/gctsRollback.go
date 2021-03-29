package cmd

import (
	"io/ioutil"
	"net/http/cookiejar"
	"net/url"

	"github.com/Jeffail/gabs/v2"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
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

	cookieJar, cookieErr := cookiejar.New(nil)
	if cookieErr != nil {
		return cookieErr
	}
	clientOptions := piperhttp.ClientOptions{
		CookieJar: cookieJar,
		Username:  config.Username,
		Password:  config.Password,
	}
	httpClient.SetOptions(clientOptions)

	repoInfo, err := getRepoInfo(config, telemetryData, httpClient)
	if err != nil {
		return errors.Wrap(err, "could not get local repository data")
	}

	if repoInfo.Result.URL == "" {
		return errors.Errorf("no remote repository URL configured")
	}

	parsedURL, err := url.Parse(repoInfo.Result.URL)
	if err != nil {
		return errors.Wrap(err, "could not parse remote repository URL as valid URL")
	}

	var deployOptions gctsDeployOptions

	if config.Commit != "" {
		log.Entry().Infof("rolling back to specified commit %v", config.Commit)

		deployOptions = gctsDeployOptions{
			Username:   config.Username,
			Password:   config.Password,
			Host:       config.Host,
			Repository: config.Repository,
			Client:     config.Client,
			Commit:     config.Commit,
		}

	} else if parsedURL.Host == "github.com" {
		log.Entry().Info("Remote repository domain is 'github.com'. Trying to rollback to last commit with status 'success'.")

		commitList, err := getCommits(config, telemetryData, httpClient)
		if err != nil {
			return errors.Wrap(err, "could not get repository commits")
		}

		successCommit, err := getLastSuccessfullCommit(config, telemetryData, httpClient, parsedURL, commitList)
		if err != nil {
			return errors.Wrap(err, "could not determine successful commit")
		}

		deployOptions = gctsDeployOptions{
			Username:   config.Username,
			Password:   config.Password,
			Host:       config.Host,
			Repository: config.Repository,
			Client:     config.Client,
			Commit:     successCommit,
		}

	} else {
		repoHistory, err := getRepoHistory(config, telemetryData, httpClient)
		if err != nil {
			return errors.Wrap(err, "could not retrieve repository commit history")
		}
		if repoHistory.Result[0].FromCommit != "" {

			log.Entry().WithField("repository", config.Repository).Infof("rolling back to last active commit %v", repoHistory.Result[0].FromCommit)
			deployOptions = gctsDeployOptions{
				Username:   config.Username,
				Password:   config.Password,
				Host:       config.Host,
				Repository: config.Repository,
				Client:     config.Client,
				Commit:     repoHistory.Result[0].FromCommit,
			}

		} else {
			return errors.Errorf("no commit to rollback to (fromCommit) could be identified from the repository commit history")
		}
	}

	deployErr := pullByCommit(&deployOptions, telemetryData, command, httpClient)

	if deployErr != nil {
		return errors.Wrap(deployErr, "rollback commit failed")
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

	httpClient.SetOptions(clientOptions)

	for _, commit := range commitList {

		url := githubURL.Scheme + "://api." + githubURL.Host + "/repos" + githubURL.Path + "/commits/" + commit + "/status"

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

		bodyText, readErr := ioutil.ReadAll(resp.Body)

		if readErr != nil {
			return "", errors.Wrapf(readErr, "HTTP response body could not be read")
		}

		response, parsingErr := gabs.ParseJSON([]byte(bodyText))

		if parsingErr != nil {
			return "", errors.Wrapf(parsingErr, "HTTP response body could not be parsed as JSON: %v", string(bodyText))
		}

		if status, ok := response.Path("state").Data().(string); ok && status == "success" {
			log.Entry().
				WithField("repository", config.Repository).
				Infof("last successful commit was determined to be %v", commit)
			return commit, nil
		}
	}

	return "", errors.Errorf("no commit with status 'success' could be found")
}

func getCommits(config *gctsRollbackOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender) ([]string, error) {

	url := config.Host +
		"/sap/bc/cts_abapvcs/repository/" + config.Repository +
		"/getCommit?sap-client=" + config.Client

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
		Objects       int    `json:"objects"`
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
