package abaputils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"reflect"
	"strconv"
	"strings"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
	"k8s.io/utils/strings/slices"
)

type SAP_COM_0510 struct {
	con                    ConnectionDetailsHTTP
	client                 piperhttp.Sender
	repository             Repository
	path                   string
	cloneEntity            string
	repositoryEntity       string
	tagsEntity             string
	checkoutAction         string
	actionEntity           string
	uuid                   string
	failureMessage         string
	maxRetries             int
	retryBaseSleepUnit     time.Duration
	retryMaxSleepTime      time.Duration
	retryAllowedErrorCodes []string
}

func (api *SAP_COM_0510) init(con ConnectionDetailsHTTP, client piperhttp.Sender, repo Repository) {
	api.con = con
	api.client = client
	api.repository = repo
	api.path = "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY"
	api.cloneEntity = "/Clones"
	api.repositoryEntity = "/Repositories"
	api.tagsEntity = "/Tags"
	api.actionEntity = "/Pull"
	api.checkoutAction = "/checkout_branch"
	api.failureMessage = "The action of the Repository / Software Component " + api.repository.Name + " failed"
	api.maxRetries = 3
	api.setSleepTimeConfig(1*time.Second, 120*time.Second)
	api.retryAllowedErrorCodes = append(api.retryAllowedErrorCodes, "A4C_A2G/228")
	api.retryAllowedErrorCodes = append(api.retryAllowedErrorCodes, "A4C_A2G/501")
}

func (api *SAP_COM_0510) GetExecutionLog() (execLog ExecutionLog, err error) {
	return execLog, errors.New("Not implemented")
}

func (api *SAP_COM_0510) getUUID() string {
	return api.uuid
}

func (api *SAP_COM_0510) CreateTag(tag Tag) error {

	if reflect.DeepEqual(Tag{}, tag) {
		return errors.New("No Tag provided")
	}

	con := api.con
	con.URL = api.con.URL + api.path + api.tagsEntity

	requestBodyStruct := CreateTagBody{RepositoryName: api.repository.Name, CommitID: api.repository.CommitID, Tag: tag.TagName, Description: tag.TagDescription}
	jsonBody, err := json.Marshal(&requestBodyStruct)
	if err != nil {
		return err
	}
	return api.triggerRequest(con, jsonBody)
}

func (api *SAP_COM_0510) CheckoutBranch() error {

	if api.repository.Name == "" || api.repository.Branch == "" {
		return fmt.Errorf("Failed to trigger checkout: %w", errors.New("Repository and/or Branch Configuration is empty. Please make sure that you have specified the correct values"))
	}

	// the request looks like: POST/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/checkout_branch?branch_name='newBranch'&sc_name=/DMO/GIT_REPOSITORY'
	checkoutConnectionDetails := api.con
	checkoutConnectionDetails.URL = api.con.URL + api.path + api.checkoutAction + `?branch_name='` + api.repository.Branch + `'&sc_name='` + api.repository.Name + `'`
	jsonBody := []byte(``)

	return api.triggerRequest(checkoutConnectionDetails, jsonBody)
}

func (api *SAP_COM_0510) parseActionResponse(resp *http.Response, err error) (ActionEntity, error) {
	var body ActionEntity
	var abapResp map[string]*json.RawMessage
	bodyText, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return ActionEntity{}, err
	}
	if err := json.Unmarshal(bodyText, &abapResp); err != nil {
		return ActionEntity{}, err
	}
	if err := json.Unmarshal(*abapResp["d"], &body); err != nil {
		return ActionEntity{}, err
	}

	if reflect.DeepEqual(ActionEntity{}, body) {
		log.Entry().WithField("StatusCode", resp.Status).WithField("branchName", api.repository.Branch).Error("Could not switch to specified branch")
		err := errors.New("Request to ABAP System not successful")
		return ActionEntity{}, err
	}
	return body, nil
}

func (api *SAP_COM_0510) Pull() error {

	// Trigger the Pull of a Repository
	if api.repository.Name == "" {
		return errors.New("An empty string was passed for the parameter 'repositoryName'")
	}

	pullConnectionDetails := api.con
	pullConnectionDetails.URL = api.con.URL + api.path + api.actionEntity

	jsonBody := []byte(api.repository.GetPullRequestBody())
	return api.triggerRequest(pullConnectionDetails, jsonBody)
}

func (api *SAP_COM_0510) GetLogProtocol(logOverviewEntry LogResultsV2, page int) (result []LogProtocol, count int, err error) {

	connectionDetails := api.con
	connectionDetails.URL = logOverviewEntry.ToLogProtocol.Deferred.URI + api.getLogProtocolQuery(page)
	resp, err := GetHTTPResponse("GET", connectionDetails, nil, api.client)
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		_, err = HandleHTTPError(resp, err, api.failureMessage, connectionDetails)
		return nil, 0, err
	}
	defer resp.Body.Close()

	// Parse response
	var body LogProtocolResults
	var abapResp map[string]*json.RawMessage
	bodyText, _ := io.ReadAll(resp.Body)

	marshallError := json.Unmarshal(bodyText, &abapResp)
	if marshallError != nil {
		return nil, 0, errors.Wrap(marshallError, "Could not parse response from the ABAP Environment system")
	}
	marshallError = json.Unmarshal(*abapResp["d"], &body)
	if marshallError != nil {
		return nil, 0, errors.Wrap(marshallError, "Could not parse response from the ABAP Environment system")
	}

	count, errConv := strconv.Atoi(body.Count)
	if errConv != nil {
		return nil, 0, errors.Wrap(errConv, "Could not parse response from the ABAP Environment system")
	}

	return body.Results, count, nil
}

func (api *SAP_COM_0510) GetLogOverview() (result []LogResultsV2, err error) {

	connectionDetails := api.con
	connectionDetails.URL = api.con.URL + api.path + api.actionEntity + "(uuid=guid'" + api.getUUID() + "')" + "?$expand=to_Log_Overview"
	resp, err := GetHTTPResponse("GET", connectionDetails, nil, api.client)
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		_, err = HandleHTTPError(resp, err, api.failureMessage, connectionDetails)
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response
	var body ActionEntity
	var abapResp map[string]*json.RawMessage
	bodyText, _ := io.ReadAll(resp.Body)

	marshallError := json.Unmarshal(bodyText, &abapResp)
	if marshallError != nil {
		return nil, errors.Wrap(marshallError, "Could not parse response from the ABAP Environment system")
	}
	marshallError = json.Unmarshal(*abapResp["d"], &body)
	if marshallError != nil {
		return nil, errors.Wrap(marshallError, "Could not parse response from the ABAP Environment system")
	}

	if reflect.DeepEqual(ActionEntity{}, body) {
		log.Entry().WithField("StatusCode", resp.Status).Error(api.failureMessage)
		log.SetErrorCategory(log.ErrorInfrastructure)
		var err = errors.New("Request to ABAP System not successful")
		return nil, err
	}

	abapStatusCode := body.Status
	log.Entry().Info("Status: " + abapStatusCode + " - " + body.StatusDescription)
	return body.ToLogOverview.Results, nil

}

func (api *SAP_COM_0510) GetAction() (string, error) {

	connectionDetails := api.con
	connectionDetails.URL = api.con.URL + api.path + api.actionEntity + "(uuid=guid'" + api.getUUID() + "')"
	resp, err := GetHTTPResponse("GET", connectionDetails, nil, api.client)
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		_, err = HandleHTTPError(resp, err, api.failureMessage, connectionDetails)
		return "E", err
	}
	defer resp.Body.Close()

	// Parse Response
	body, parseError := api.parseActionResponse(resp, err)
	if parseError != nil {
		return "E", parseError
	}

	api.uuid = body.UUID

	abapStatusCode := body.Status
	log.Entry().Info("Status: " + abapStatusCode + " - " + body.StatusDescription)
	return abapStatusCode, nil
}

func (api *SAP_COM_0510) GetRepository() (bool, string, error) {

	if api.repository.Name == "" {
		return false, "", errors.New("An empty string was passed for the parameter 'repositoryName'")
	}

	swcConnectionDetails := api.con
	swcConnectionDetails.URL = api.con.URL + api.path + api.repositoryEntity + "('" + strings.Replace(api.repository.Name, "/", "%2F", -1) + "')"
	resp, err := GetHTTPResponse("GET", swcConnectionDetails, nil, api.client)
	if err != nil {
		_, errRepo := HandleHTTPError(resp, err, "Reading the Repository / Software Component failed", api.con)
		return false, "", errRepo
	}
	defer resp.Body.Close()

	var body RepositoryEntity
	var abapResp map[string]*json.RawMessage
	bodyText, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return false, "", err
	}

	if err := json.Unmarshal(bodyText, &abapResp); err != nil {
		return false, "", err
	}
	if err := json.Unmarshal(*abapResp["d"], &body); err != nil {
		return false, "", err
	}
	if reflect.DeepEqual(RepositoryEntity{}, body) {
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", api.repository.Name).WithField("branchName", api.repository.Branch).WithField("commitID", api.repository.CommitID).WithField("Tag", api.repository.Tag).Error("Could not Clone the Repository / Software Component")
		err := errors.New("Request to ABAP System not successful")
		return false, "", err
	}

	if body.AvailOnInst {
		return true, body.ActiveBranch, nil
	}
	return false, "", err

}

func (api *SAP_COM_0510) Clone() error {

	// Trigger the Clone of a Repository
	if api.repository.Name == "" {
		return errors.New("An empty string was passed for the parameter 'repositoryName'")
	}

	cloneConnectionDetails := api.con
	cloneConnectionDetails.URL = api.con.URL + api.path + api.cloneEntity
	body := []byte(api.repository.GetCloneRequestBodyWithSWC())

	return api.triggerRequest(cloneConnectionDetails, body)

}

func (api *SAP_COM_0510) triggerRequest(cloneConnectionDetails ConnectionDetailsHTTP, jsonBody []byte) error {
	var err error
	var body ActionEntity
	var resp *http.Response
	var errorCode string

	for i := 0; i <= api.maxRetries; i++ {
		if i > 0 {
			sleepTime, err := api.getSleepTime(i + 5)
			if err != nil {
				// reached max retry duration
				break
			}
			log.Entry().Infof("Retrying in %s", sleepTime.String())
			time.Sleep(sleepTime)
		}
		resp, err = GetHTTPResponse("POST", cloneConnectionDetails, jsonBody, api.client)
		if err != nil {
			errorCode, err = HandleHTTPError(resp, err, "Triggering the action failed", api.con)
			if slices.Contains(api.retryAllowedErrorCodes, errorCode) {
				// Error Code allows for retry
				continue
			} else {
				break
			}
		}
		defer resp.Body.Close()
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", api.repository.Name).WithField("branchName", api.repository.Branch).WithField("commitID", api.repository.CommitID).WithField("Tag", api.repository.Tag).Info("Triggered action of Repository / Software Component")

		body, err = api.parseActionResponse(resp, err)
		break
	}
	api.uuid = body.UUID
	return err
}

// initialRequest implements SoftwareComponentApiInterface.
func (api *SAP_COM_0510) initialRequest() error {
	// Configuring the HTTP Client and CookieJar
	cookieJar, errorCookieJar := cookiejar.New(nil)
	if errorCookieJar != nil {
		return errors.Wrap(errorCookieJar, "Could not create a Cookie Jar")
	}

	api.client.SetOptions(piperhttp.ClientOptions{
		MaxRequestDuration: 180 * time.Second,
		CookieJar:          cookieJar,
		Username:           api.con.User,
		Password:           api.con.Password,
	})

	headConnection := api.con
	headConnection.XCsrfToken = "fetch"
	headConnection.URL = api.con.URL + api.path

	// Loging into the ABAP System - getting the x-csrf-token and cookies
	resp, err := GetHTTPResponse("HEAD", headConnection, nil, api.client)
	if err != nil {
		_, err = HandleHTTPError(resp, err, "Authentication on the ABAP system failed", api.con)
		return err
	}
	defer resp.Body.Close()

	log.Entry().WithField("StatusCode", resp.Status).WithField("ABAP Endpoint", api.con).Debug("Authentication on the ABAP system successful")
	api.con.XCsrfToken = resp.Header.Get("X-Csrf-Token")
	return nil
}

// getSleepTime Should return the Fibonacci numbers in the define time unit up to the defined maximum duration
func (api *SAP_COM_0510) getSleepTime(n int) (time.Duration, error) {

	if n == 0 {
		return 0, nil
	} else if n == 1 {
		return 1 * api.retryBaseSleepUnit, nil
	} else if n < 0 {
		return 0, errors.New("Negative numbers are not allowed")
	}
	var result, i int
	prev := 0
	next := 1
	for i = 2; i <= n; i++ {
		result = prev + next
		prev = next
		next = result
	}
	sleepTime := time.Duration(result) * api.retryBaseSleepUnit

	if sleepTime > api.retryMaxSleepTime {
		return 0, errors.New("Exceeded max sleep time")
	}
	return sleepTime, nil
}

// setSleepTimeConfig sets the time unit (seconds, nanoseconds) and the maximum sleep duration
func (api *SAP_COM_0510) setSleepTimeConfig(timeUnit time.Duration, maxSleepTime time.Duration) {
	api.retryBaseSleepUnit = timeUnit
	api.retryMaxSleepTime = maxSleepTime
}

func (api *SAP_COM_0510) getLogProtocolQuery(page int) string {
	skip := page * numberOfEntriesPerPage
	top := numberOfEntriesPerPage

	return fmt.Sprintf("?$skip=%s&$top=%s&$inlinecount=allpages", fmt.Sprint(skip), fmt.Sprint(top))
}
