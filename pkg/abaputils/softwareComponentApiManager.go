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

type SoftwareComponentApiManagerInterface interface {
	GetAPI(con ConnectionDetailsHTTP, client piperhttp.Sender) (SoftwareComponentApiInterface, error)
}

type SoftwareComponentApiManager struct{}

func (manager *SoftwareComponentApiManager) GetAPI(con ConnectionDetailsHTTP, client piperhttp.Sender) (SoftwareComponentApiInterface, error) {
	sap_com_0510 := SAP_COM_0510{}
	sap_com_0510.init(con, client)

	// Initialize all APIs, use the one that returns a response
	// Currently SAP_COM_0510, later SAP_COM_0948
	err := sap_com_0510.initialRequest()
	return &sap_com_0510, err
}

type SoftwareComponentApiInterface interface {
	init(con ConnectionDetailsHTTP, client piperhttp.Sender)
	initialRequest() error
	Clone(repo Repository) error
	CheckIfAlreadyCloned(repo Repository) (bool, string, error)
}

type SAP_COM_0510 struct {
	con              ConnectionDetailsHTTP
	client           piperhttp.Sender
	path             string
	cloneEntity      string
	repositoryEntity string
	UUID             string
}

func (api *SAP_COM_0510) CheckIfAlreadyCloned(repo Repository) (bool, string, error) {
	if repo.Name == "" {
		return false, "", errors.New("An empty string was passed for the parameter 'repositoryName'")
	}

	swcConnectionDetails := api.con
	swcConnectionDetails.URL = api.con.URL + api.path + api.repositoryEntity + "('" + strings.Replace(repo.Name, "/", "%2F", -1) + "')"
	resp, err := GetHTTPResponse("GET", swcConnectionDetails, nil, api.client)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()
	log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", repo.Name).WithField("branchName", repo.Branch).WithField("commitID", repo.CommitID).WithField("Tag", repo.Tag).Info("Triggered Clone of Repository / Software Component")

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
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", repo.Name).WithField("branchName", repo.Branch).WithField("commitID", repo.CommitID).WithField("Tag", repo.Tag).Error("Could not Clone the Repository / Software Component")
		err := errors.New("Request to ABAP System not successful")
		return false, "", err
	}

	if body.AvailableOnInstance {
		return true, body.ActiveBranch, nil
	}
	return false, "", err

}

func (api *SAP_COM_0510) Clone(repo Repository) error {

	// Trigger the Clone of a Repository
	if repo.Name == "" {
		return errors.New("An empty string was passed for the parameter 'repositoryName'")
	}

	cloneConnectionDetails := api.con
	cloneConnectionDetails.URL = api.con.URL + api.path + api.cloneEntity

	jsonBody := []byte(repo.GetCloneRequestBody())
	resp, err := GetHTTPResponse("POST", cloneConnectionDetails, jsonBody, api.client)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", repo.Name).WithField("branchName", repo.Branch).WithField("commitID", repo.CommitID).WithField("Tag", repo.Tag).Info("Triggered Clone of Repository / Software Component")

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
		log.Entry().WithField("StatusCode", resp.Status).WithField("repositoryName", repo.Name).WithField("branchName", repo.Branch).WithField("commitID", repo.CommitID).WithField("Tag", repo.Tag).Error("Could not Clone the Repository / Software Component")
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

func (api *SAP_COM_0510) init(con ConnectionDetailsHTTP, client piperhttp.Sender) {
	api.con = con
	api.client = client
	api.path = "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY"
	api.cloneEntity = "/Clones"
	api.repositoryEntity = "/Repositories"

}
