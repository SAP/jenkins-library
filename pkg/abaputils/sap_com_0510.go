package abaputils

import (
	"encoding/json"
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
	con               ConnectionDetailsHTTP
	client            piperhttp.Sender
	softwareComponent Repository
	path              string
	cloneEntity       string
	repositoryEntity  string
	pullEntity        string
	UUID              string
	failureMessage    string
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
	connectionDetails.URL = api.con.URL + api.path + api.pullEntity + "(uuid=guid'" + api.UUID + "')" + "?$expand=to_Log_Overview"
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
	connectionDetails.URL = api.con.URL + api.path + api.pullEntity + "(uuid=guid'" + api.UUID + "')"
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

	if api.softwareComponent.Name == "" {
		return false, "", errors.New("An empty string was passed for the parameter 'repositoryName'")
	}

	swcConnectionDetails := api.con
	swcConnectionDetails.URL = api.con.URL + api.path + api.repositoryEntity + "('" + strings.Replace(api.softwareComponent.Name, "/", "%2F", -1) + "')"
	resp, err := GetHTTPResponse("GET", swcConnectionDetails, nil, api.client)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()
	log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", api.softwareComponent.Name).WithField("branchName", api.softwareComponent.Branch).WithField("commitID", api.softwareComponent.CommitID).WithField("Tag", api.softwareComponent.Tag).Info("Triggered Clone of Repository / Software Component")

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
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", api.softwareComponent.Name).WithField("branchName", api.softwareComponent.Branch).WithField("commitID", api.softwareComponent.CommitID).WithField("Tag", api.softwareComponent.Tag).Error("Could not Clone the Repository / Software Component")
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
	if api.softwareComponent.Name == "" {
		return errors.New("An empty string was passed for the parameter 'repositoryName'")
	}

	cloneConnectionDetails := api.con
	cloneConnectionDetails.URL = api.con.URL + api.path + api.cloneEntity

	jsonBody := []byte(api.softwareComponent.GetCloneRequestBody())
	resp, err := GetHTTPResponse("POST", cloneConnectionDetails, jsonBody, api.client)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", api.softwareComponent.Name).WithField("branchName", api.softwareComponent.Branch).WithField("commitID", api.softwareComponent.CommitID).WithField("Tag", api.softwareComponent.Tag).Info("Triggered Clone of Repository / Software Component")

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
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", api.softwareComponent.Name).WithField("branchName", api.softwareComponent.Branch).WithField("commitID", api.softwareComponent.CommitID).WithField("Tag", api.softwareComponent.Tag).Error("Could not Clone the Repository / Software Component")
		err := errors.New("Request to ABAP System not successful")
		return err
	}

	api.UUID = body.UUID
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
	api.softwareComponent = repo
	api.path = "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY"
	api.cloneEntity = "/Clones"
	api.repositoryEntity = "/Repositories"
	api.pullEntity = "/Pull"
	api.failureMessage = "Could not pull the Repository / Software Component " + api.softwareComponent.Name

}
