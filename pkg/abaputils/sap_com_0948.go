package abaputils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"reflect"
	"regexp"
	"strings"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
	"k8s.io/utils/strings/slices"
)

type SAP_COM_0948 struct {
	con                     ConnectionDetailsHTTP
	client                  piperhttp.Sender
	repository              Repository
	path                    string
	cloneAction             string
	pullAction              string
	softwareComponentEntity string
	branchEntity            string
	tagsEntity              string
	checkoutAction          string
	actionsEntity           string
	uuid                    string
	failureMessage          string
	maxRetries              int
	retryBaseSleepUnit      time.Duration
	retryMaxSleepTime       time.Duration
	retryAllowedErrorCodes  []string
}

func (api *SAP_COM_0948) init(con ConnectionDetailsHTTP, client piperhttp.Sender, repo Repository) {
	api.con = con
	api.client = client
	api.repository = repo
	api.path = "/sap/opu/odata4/sap/a4c_mswc_api/srvd_a2x/sap/manage_software_components/0001"
	api.checkoutAction = "/SAP__self.checkout_branch"
	api.softwareComponentEntity = "/SoftwareComponents"
	api.actionsEntity = "/Actions"
	api.branchEntity = "/Branches"
	api.cloneAction = "/SAP__self.clone"
	api.pullAction = "/SAP__self.pull"
	api.tagsEntity = "/Tags"
	api.failureMessage = "The action of the Repository / Software Component " + api.repository.Name + " failed"
	api.maxRetries = 3
	api.setSleepTimeConfig(1*time.Second, 120*time.Second)
	api.retryAllowedErrorCodes = append(api.retryAllowedErrorCodes, "A4C_A2G/228")
	api.retryAllowedErrorCodes = append(api.retryAllowedErrorCodes, "A4C_A2G/501")
}

func (api *SAP_COM_0948) getUUID() string {
	return api.uuid
}

// reads the execution log from the ABAP system
func (api *SAP_COM_0948) GetExecutionLog() (execLog ExecutionLog, err error) {

	connectionDetails := api.con
	connectionDetails.URL = api.con.URL + api.path + api.actionsEntity + "/" + api.getUUID() + "/_Execution_log"
	resp, err := GetHTTPResponse("GET", connectionDetails, nil, api.client)
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		_, err = handleHTTPError(resp, err, api.failureMessage, connectionDetails)
		return execLog, err
	}
	defer resp.Body.Close()

	// Parse response
	bodyText, _ := io.ReadAll(resp.Body)

	marshallError := json.Unmarshal(bodyText, &execLog)
	if marshallError != nil {
		return execLog, errors.Wrap(marshallError, "Could not parse response from the ABAP Environment system")
	}

	if reflect.DeepEqual(ExecutionLog{}, execLog) {
		log.Entry().WithField("StatusCode", resp.Status).Error(api.failureMessage)
		log.SetErrorCategory(log.ErrorInfrastructure)
		var err = errors.New("Request to ABAP System not successful")
		return execLog, err
	}
	return execLog, nil
}

func (api *SAP_COM_0948) CreateTag(tag Tag) error {

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

func (api *SAP_COM_0948) CheckoutBranch() error {

	if api.repository.Name == "" || api.repository.Branch == "" {
		return fmt.Errorf("Failed to trigger checkout: %w", errors.New("Repository and/or Branch Configuration is empty. Please make sure that you have specified the correct values"))
	}

	checkoutConnectionDetails := api.con
	checkoutConnectionDetails.URL = api.con.URL + api.path + api.branchEntity + api.getRepoNameForPath() + api.getBranchNameForPath() + api.checkoutAction
	jsonBody := []byte(`{
		"import_mode" : "",
		"execution_mode": ""
		}`)

	return api.triggerRequest(checkoutConnectionDetails, jsonBody)
}

func (api *SAP_COM_0948) parseActionResponse(resp *http.Response, err error) (ActionEntity, error) {

	var body ActionEntity
	bodyText, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return ActionEntity{}, err
	}
	if err := json.Unmarshal(bodyText, &body); err != nil {
		return ActionEntity{}, err
	}
	if reflect.DeepEqual(ActionEntity{}, body) {
		log.Entry().WithField("StatusCode", resp.Status).WithField("branchName", api.repository.Branch).Error("Could not switch to specified branch")
		err := errors.New("Request to ABAP System not successful")
		return ActionEntity{}, err
	}
	return body, nil
}

func (api *SAP_COM_0948) Pull() error {

	// Trigger the Pull of a Repository
	if api.repository.Name == "" {
		return errors.New("An empty string was passed for the parameter 'repositoryName'")
	}

	pullConnectionDetails := api.con
	pullConnectionDetails.URL = api.con.URL + api.path + api.softwareComponentEntity + api.getRepoNameForPath() + api.pullAction

	jsonBody := []byte(api.repository.GetPullActionRequestBody())
	return api.triggerRequest(pullConnectionDetails, jsonBody)
}

func (api *SAP_COM_0948) GetLogProtocol(logOverviewEntry LogResultsV2, page int) (result []LogProtocol, count int, err error) {

	connectionDetails := api.con
	connectionDetails.URL = api.con.URL + api.path + api.actionsEntity + "/" + api.getUUID() + "/_Log_Overview" + "/" + fmt.Sprint(logOverviewEntry.Index) + "/_Log_Protocol" + api.getLogProtocolQuery(page)
	resp, err := GetHTTPResponse("GET", connectionDetails, nil, api.client)
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		_, err = handleHTTPError(resp, err, api.failureMessage, connectionDetails)
		return nil, 0, err
	}
	defer resp.Body.Close()

	// Parse response
	var body LogProtocolResultsV4
	bodyText, _ := io.ReadAll(resp.Body)

	marshallError := json.Unmarshal(bodyText, &body)
	if marshallError != nil {
		return nil, 0, errors.Wrap(marshallError, "Could not parse response from the ABAP Environment system")
	}

	return body.Results, body.Count, nil
}

func (api *SAP_COM_0948) GetLogOverview() (result []LogResultsV2, err error) {

	connectionDetails := api.con
	connectionDetails.URL = api.con.URL + api.path + api.actionsEntity + "/" + api.getUUID() + "/_Log_Overview"
	resp, err := GetHTTPResponse("GET", connectionDetails, nil, api.client)
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		_, err = handleHTTPError(resp, err, api.failureMessage, connectionDetails)
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response
	var abapResp map[string]*json.RawMessage
	bodyText, _ := io.ReadAll(resp.Body)

	marshallError := json.Unmarshal(bodyText, &abapResp)
	if marshallError != nil {
		return nil, errors.Wrap(marshallError, "Could not parse response from the ABAP Environment system")
	}
	marshallError = json.Unmarshal(*abapResp["value"], &result)
	if marshallError != nil {
		return nil, errors.Wrap(marshallError, "Could not parse response from the ABAP Environment system")
	}

	if reflect.DeepEqual(LogResultsV2{}, result) {
		log.Entry().WithField("StatusCode", resp.Status).Error(api.failureMessage)
		log.SetErrorCategory(log.ErrorInfrastructure)
		var err = errors.New("Request to ABAP System not successful")
		return nil, err
	}
	return result, nil

}

func (api *SAP_COM_0948) GetAction() (string, error) {

	connectionDetails := api.con
	connectionDetails.URL = api.con.URL + api.path + api.actionsEntity + "/" + api.getUUID()
	resp, err := GetHTTPResponse("GET", connectionDetails, nil, api.client)
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		_, err = handleHTTPError(resp, err, api.failureMessage, connectionDetails)
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

func (api *SAP_COM_0948) GetRepository() (bool, string, error, bool) {

	if api.repository.Name == "" {
		return false, "", errors.New("An empty string was passed for the parameter 'repositoryName'"), false
	}

	swcConnectionDetails := api.con
	swcConnectionDetails.URL = api.con.URL + api.path + api.softwareComponentEntity + api.getRepoNameForPath()
	resp, err := GetHTTPResponse("GET", swcConnectionDetails, nil, api.client)
	if err != nil {
		_, errRepo := handleHTTPError(resp, err, "Reading the Repository / Software Component failed", api.con)
		return false, "", errRepo, false
	}
	defer resp.Body.Close()

	var body RepositoryEntity
	bodyText, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return false, "", err, false
	}

	if err := json.Unmarshal(bodyText, &body); err != nil {
		return false, "", err, false
	}
	if reflect.DeepEqual(RepositoryEntity{}, body) {
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", api.repository.Name).WithField("branchName", api.repository.Branch).WithField("commitID", api.repository.CommitID).WithField("Tag", api.repository.Tag).Error("Could not Clone the Repository / Software Component")
		err := errors.New("Request to ABAP System not successful")
		return false, "", err, false
	}

	if body.AvailOnInst {
		return true, body.ActiveBranch, nil, false
	}

	if body.ByogUrl != "" {
		return false, "", err, true
	}

	return false, "", err, false

}

func (api *SAP_COM_0948) UpdateRepoWithBYOGCredentials(byogAuthMethod string, byogUsername string, byogPassword string) {
	api.repository.ByogAuthMethod = byogAuthMethod
	api.repository.ByogUsername = byogUsername
	api.repository.ByogPassword = byogPassword
	api.repository.IsByog = true
}

func (api *SAP_COM_0948) Clone() error {

	// Trigger the Clone of a Repository
	if api.repository.Name == "" {
		return errors.New("An empty string was passed for the parameter 'repositoryName'")
	}

	cloneConnectionDetails := api.con
	cloneConnectionDetails.URL = api.con.URL + api.path + api.softwareComponentEntity + api.getRepoNameForPath() + api.cloneAction
	body, err := api.repository.GetCloneRequestBody()
	if err != nil {
		return errors.Wrap(err, "Failed to clone repository")
	}

	return api.triggerRequest(cloneConnectionDetails, []byte(body))

}

func (api *SAP_COM_0948) triggerRequest(cloneConnectionDetails ConnectionDetailsHTTP, jsonBody []byte) error {
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
			errorCode, err = handleHTTPError(resp, err, "Triggering the action failed", api.con)
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
func (api *SAP_COM_0948) initialRequest() error {
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
		TrustedCerts:       api.con.CertificateNames,
	})

	// HEAD request to the root is not sufficient, as an unauthorized called is allowed to do so
	// Therefore, the request goes to the "Actions" entity without actually fetching data
	headConnection := api.con
	headConnection.XCsrfToken = "fetch"
	headConnection.URL = api.con.URL + api.path + api.actionsEntity + "?$top=0"

	// Loging into the ABAP System - getting the x-csrf-token and cookies
	resp, err := GetHTTPResponse("GET", headConnection, nil, api.client)
	if err != nil {
		_, err = handleHTTPError(resp, err, "Authentication on the ABAP system failed", api.con)
		return err
	}
	defer resp.Body.Close()

	log.Entry().WithField("StatusCode", resp.Status).WithField("ABAP Endpoint", api.con).Debug("Authentication on the ABAP system successful")
	api.con.XCsrfToken = resp.Header.Get("X-Csrf-Token")
	return nil
}

// getSleepTime Should return the Fibonacci numbers in the define time unit up to the defined maximum duration
func (api *SAP_COM_0948) getSleepTime(n int) (time.Duration, error) {

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
func (api *SAP_COM_0948) setSleepTimeConfig(timeUnit time.Duration, maxSleepTime time.Duration) {
	api.retryBaseSleepUnit = timeUnit
	api.retryMaxSleepTime = maxSleepTime
}

func (api *SAP_COM_0948) getRepoNameForPath() string {
	return "/" + strings.ReplaceAll(api.repository.Name, "/", "%2F")
}

func (api *SAP_COM_0948) getBranchNameForPath() string {
	return "/" + api.repository.Branch
}

func (api *SAP_COM_0948) getLogProtocolQuery(page int) string {
	skip := page * numberOfEntriesPerPage
	top := numberOfEntriesPerPage

	return fmt.Sprintf("?$skip=%s&$top=%s&$count=true", fmt.Sprint(skip), fmt.Sprint(top))
}

// ConvertTime formats an ISO 8601 timestamp string from format 2024-05-02T09:25:40Z into a UNIX timestamp and returns it
func (api *SAP_COM_0948) ConvertTime(logTimeStamp string) time.Time {
	t, error := time.Parse(time.RFC3339, logTimeStamp)
	if error != nil {
		return time.Unix(0, 0).UTC()
	}
	return t
}

func handleHTTPError(resp *http.Response, err error, message string, connectionDetails ConnectionDetailsHTTP) (string, error) {

	var errorText string
	var errorCode string
	var parsingError error
	if resp == nil {
		// Response is nil in case of a timeout
		log.Entry().WithError(err).WithField("ABAP Endpoint", connectionDetails.URL).Error("Request failed")

		match, _ := regexp.MatchString(".*EOF$", err.Error())
		if match {
			AddDefaultDashedLine(1)
			log.Entry().Infof("%s", "A connection could not be established to the ABAP system. The typical root cause is the network configuration (firewall, IP allowlist, etc.)")
			AddDefaultDashedLine(1)
		}

		log.Entry().Infof("Error message: %s,", err.Error())
	} else {

		defer resp.Body.Close()

		log.Entry().WithField("StatusCode", resp.Status).WithField("User", connectionDetails.User).WithField("URL", connectionDetails.URL).Error(message)

		errorText, errorCode, parsingError = getErrorDetailsFromResponse(resp)
		if parsingError != nil {
			return "", err
		}
		abapError := errors.New(fmt.Sprintf("%s - %s", errorCode, errorText))
		err = errors.Wrap(abapError, err.Error())

	}
	return errorCode, err
}

func getErrorDetailsFromResponse(resp *http.Response) (errorString string, errorCode string, err error) {

	// Include the error message of the ABAP Environment system, if available
	var abapErrorResponse AbapErrorODataV4
	bodyText, readError := io.ReadAll(resp.Body)
	if readError != nil {
		return "", "", readError
	}
	var abapResp map[string]*json.RawMessage
	errUnmarshal := json.Unmarshal(bodyText, &abapResp)
	if errUnmarshal != nil {
		return "", "", errUnmarshal
	}
	if _, ok := abapResp["error"]; ok {
		json.Unmarshal(*abapResp["error"], &abapErrorResponse)
		if (AbapErrorODataV4{}) != abapErrorResponse {
			log.Entry().WithField("ErrorCode", abapErrorResponse.Code).Debug(abapErrorResponse.Message)
			return abapErrorResponse.Message, abapErrorResponse.Code, nil
		}
	}

	return "", "", errors.New("Could not parse the JSON error response")

}
