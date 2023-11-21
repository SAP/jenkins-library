package abaputils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http/cookiejar"
	"reflect"
	"strings"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

type SAP_COM_0510 struct {
	con              ConnectionDetailsHTTP
	client           piperhttp.Sender
	repository       Repository
	path             string
	cloneEntity      string
	repositoryEntity string
	tagsEntity       string
	checkoutAction   string
	actionEntity     string
	uuid             string
	failureMessage   string
}

func (api *SAP_COM_0510) getUUID() string {
	return api.uuid
}

func (api *SAP_COM_0510) CreateTag(tag Tag) error {

	con := api.con
	con.URL = api.con.URL + api.path + api.tagsEntity

	requestBodyStruct := CreateTagBody{RepositoryName: api.repository.Name, CommitID: api.repository.CommitID, Tag: tag.TagName, Description: tag.TagDescription}
	requestBodyJson, err := json.Marshal(&requestBodyStruct)
	if err != nil {
		return err
	}

	log.Entry().Debugf("Request body: %s", requestBodyJson)
	resp, err := GetHTTPResponse("POST", con, requestBodyJson, api.client)
	if err != nil {
		errorMessage := "Could not create tag " + requestBodyStruct.Tag + " for repository " + requestBodyStruct.RepositoryName + " with commitID " + requestBodyStruct.CommitID
		err = HandleHTTPError(resp, err, errorMessage, con)
		return err
	}
	defer resp.Body.Close()

	// Parse response
	var createTagResponse CreateTagResponse
	var abapResp map[string]*json.RawMessage
	bodyText, _ := io.ReadAll(resp.Body)

	if err = json.Unmarshal(bodyText, &abapResp); err != nil {
		return err
	}
	if err = json.Unmarshal(*abapResp["d"], &createTagResponse); err != nil {
		return err
	}
	api.uuid = createTagResponse.UUID
	return nil
}

func (api *SAP_COM_0510) CheckoutBranch() error {

	if api.repository.Name == "" || api.repository.Branch == "" {
		return fmt.Errorf("Failed to trigger checkout: %w", errors.New("Repository and/or Branch Configuration is empty. Please make sure that you have specified the correct values"))
	}

	// the request looks like: POST/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/checkout_branch?branch_name='newBranch'&sc_name=/DMO/GIT_REPOSITORY'
	checkoutConnectionDetails := api.con
	checkoutConnectionDetails.URL = api.con.URL + api.path + api.checkoutAction + `?branch_name='` + api.repository.Branch + `'&sc_name='` + api.repository.Name + `'`

	jsonBody := []byte(``)

	// no JSON body needed
	resp, err := GetHTTPResponse("POST", checkoutConnectionDetails, jsonBody, api.client)
	if err != nil {
		err = HandleHTTPError(resp, err, "Could not trigger checkout of branch "+api.repository.Branch, checkoutConnectionDetails)
		return err
	}
	defer resp.Body.Close()
	log.Entry().WithField("StatusCode", resp.StatusCode).WithField("repositoryName", api.repository.Name).WithField("branchName", api.repository.Branch).Debug("Triggered checkout of branch")

	// Parse Response
	var body ActionEntity
	var abapResp map[string]*json.RawMessage
	bodyText, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return err
	}
	if err := json.Unmarshal(bodyText, &abapResp); err != nil {
		return err
	}
	if err := json.Unmarshal(*abapResp["d"], &body); err != nil {
		return err
	}

	if reflect.DeepEqual(ActionEntity{}, body) {
		log.Entry().WithField("StatusCode", resp.Status).WithField("branchName", api.repository.Branch).Error("Could not switch to specified branch")
		err := errors.New("Request to ABAP System failed")
		return err
	}

	api.uuid = body.UUID
	return nil
}

func (api *SAP_COM_0510) Pull() error {

	// Trigger the Pull of a Repository
	if api.repository.Name == "" {
		return errors.New("An empty string was passed for the parameter 'repositoryName'")
	}

	pullConnectionDetails := api.con
	pullConnectionDetails.URL = api.con.URL + api.path + api.actionEntity

	jsonBody := []byte(api.repository.GetPullRequestBody())
	resp, err := GetHTTPResponse("POST", pullConnectionDetails, jsonBody, api.client)
	if err != nil {
		err = HandleHTTPError(resp, err, "Could not pull the "+api.repository.GetPullLogString(), pullConnectionDetails)
		return err
	}
	defer resp.Body.Close()
	log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", api.repository.Name).WithField("commitID", api.repository.CommitID).WithField("Tag", api.repository.Tag).Debug("Triggered Pull of repository / software component")

	// Parse Response
	var body ActionEntity
	var abapResp map[string]*json.RawMessage
	bodyText, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return err
	}
	if err := json.Unmarshal(bodyText, &abapResp); err != nil {
		return err
	}
	if err := json.Unmarshal(*abapResp["d"], &body); err != nil {
		return err
	}
	if reflect.DeepEqual(ActionEntity{}, body) {
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", api.repository.Name).WithField("commitID", api.repository.CommitID).WithField("Tag", api.repository.Tag).Error("Could not pull the repository / software component")
		err := errors.New("Request to ABAP System not successful")
		return err
	}

	api.uuid = body.UUID
	return nil
}

func (api *SAP_COM_0510) GetLogProtocol(logOverviewEntry LogResultsV2, page int) (body LogProtocolResults, err error) {

	connectionDetails := api.con
	connectionDetails.URL = logOverviewEntry.ToLogProtocol.Deferred.URI + getLogProtocolQuery(page)
	resp, err := GetHTTPResponse("GET", connectionDetails, nil, api.client)
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		err = HandleHTTPError(resp, err, api.failureMessage, connectionDetails)
		return body, err
	}
	defer resp.Body.Close()

	// Parse response
	var abapResp map[string]*json.RawMessage
	bodyText, _ := io.ReadAll(resp.Body)

	marshallError := json.Unmarshal(bodyText, &abapResp)
	if marshallError != nil {
		return body, errors.Wrap(marshallError, "Could not parse response from the ABAP Environment system")
	}
	marshallError = json.Unmarshal(*abapResp["d"], &body)
	if marshallError != nil {
		return body, errors.Wrap(marshallError, "Could not parse response from the ABAP Environment system")
	}

	return body, nil
}

func (api *SAP_COM_0510) GetLogOverview() (body ActionEntity, err error) {

	connectionDetails := api.con
	connectionDetails.URL = api.con.URL + api.path + api.actionEntity + "(uuid=guid'" + api.getUUID() + "')" + "?$expand=to_Log_Overview"
	resp, err := GetHTTPResponse("GET", connectionDetails, nil, api.client)
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		err = HandleHTTPError(resp, err, api.failureMessage, connectionDetails)
		return body, err
	}
	defer resp.Body.Close()

	// Parse response
	var abapResp map[string]*json.RawMessage
	bodyText, _ := io.ReadAll(resp.Body)

	marshallError := json.Unmarshal(bodyText, &abapResp)
	if marshallError != nil {
		return body, errors.Wrap(marshallError, "Could not parse response from the ABAP Environment system")
	}
	marshallError = json.Unmarshal(*abapResp["d"], &body)
	if marshallError != nil {
		return body, errors.Wrap(marshallError, "Could not parse response from the ABAP Environment system")
	}

	if reflect.DeepEqual(ActionEntity{}, body) {
		log.Entry().WithField("StatusCode", resp.Status).Error(api.failureMessage)
		log.SetErrorCategory(log.ErrorInfrastructure)
		var err = errors.New("Request to ABAP System not successful")
		return body, err
	}

	abapStatusCode := body.Status
	log.Entry().Info("Status: " + abapStatusCode + " - " + body.StatusDescription)
	return body, nil

}

func (api *SAP_COM_0510) GetAction() (string, error) {

	connectionDetails := api.con
	connectionDetails.URL = api.con.URL + api.path + api.actionEntity + "(uuid=guid'" + api.getUUID() + "')"
	resp, err := GetHTTPResponse("GET", connectionDetails, nil, api.client)
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		err = HandleHTTPError(resp, err, api.failureMessage, connectionDetails)
		return "E", err
	}
	defer resp.Body.Close()

	// Parse response
	var abapResp map[string]*json.RawMessage
	bodyText, _ := io.ReadAll(resp.Body)
	var body ActionEntity

	marshallError := json.Unmarshal(bodyText, &abapResp)
	if marshallError != nil {
		return "E", errors.Wrap(marshallError, "Could not parse response from the ABAP Environment system")
	}
	marshallError = json.Unmarshal(*abapResp["d"], &body)
	if marshallError != nil {
		return "E", errors.Wrap(marshallError, "Could not parse response from the ABAP Environment system")
	}

	if reflect.DeepEqual(ActionEntity{}, body) {
		log.Entry().WithField("StatusCode", resp.Status).Error(api.failureMessage)
		log.SetErrorCategory(log.ErrorInfrastructure)
		var err = errors.New("Request to ABAP System not successful")
		return "E", err
	}

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
		errRepo := HandleHTTPError(resp, err, "Reading the Repository / Software Component failed", api.con)
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

	if body.AvailableOnInstance {
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

	jsonBody := []byte(api.repository.GetCloneRequestBody())
	resp, err := GetHTTPResponse("POST", cloneConnectionDetails, jsonBody, api.client)
	if err != nil {
		errClone := HandleHTTPError(resp, err, "Triggering the clone action failed", api.con)
		return errClone
	}
	defer resp.Body.Close()
	log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", api.repository.Name).WithField("branchName", api.repository.Branch).WithField("commitID", api.repository.CommitID).WithField("Tag", api.repository.Tag).Info("Triggered Clone of Repository / Software Component")

	// Parse Response
	var body CloneEntity
	var abapResp map[string]*json.RawMessage
	bodyText, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return err
	}
	if err := json.Unmarshal(bodyText, &abapResp); err != nil {
		return err
	}
	if err := json.Unmarshal(*abapResp["d"], &body); err != nil {
		return err
	}
	if reflect.DeepEqual(CloneEntity{}, body) {
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", api.repository.Name).WithField("branchName", api.repository.Branch).WithField("commitID", api.repository.CommitID).WithField("Tag", api.repository.Tag).Error("Could not Clone the Repository / Software Component")
		err := errors.New("Request to ABAP System not successful")
		return err
	}

	api.uuid = body.UUID
	return nil

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
		err = HandleHTTPError(resp, err, "Authentication on the ABAP system failed", api.con)
		return err
	}
	defer resp.Body.Close()

	log.Entry().WithField("StatusCode", resp.Status).WithField("ABAP Endpoint", api.con).Debug("Authentication on the ABAP system successful")
	api.con.XCsrfToken = resp.Header.Get("X-Csrf-Token")
	return nil
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

}
