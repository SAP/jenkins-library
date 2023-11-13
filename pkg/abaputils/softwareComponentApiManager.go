package abaputils

import (
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
)

type SoftwareComponentApiManagerInterface interface {
	getAPI(con ConnectionDetailsHTTP, client piperhttp.Sender) (SoftwareComponentApiInterface, error)
}

type SoftwareComponentApiManager struct{}

func (manager *SoftwareComponentApiManager) getAPI(con ConnectionDetailsHTTP, client piperhttp.Sender) (SoftwareComponentApiInterface, error) {
	sap_com_0510 := SAP_COM_0510{}
	sap_com_0510.init(con, client)

	err := sap_com_0510.initialRequest()
	return &sap_com_0510, err
}

type SoftwareComponentApiInterface interface {
	init(con ConnectionDetailsHTTP, client piperhttp.Sender)
	initialRequest() error
}

type SAP_COM_0510 struct {
	con    ConnectionDetailsHTTP
	client piperhttp.Sender
}

// initialRequest implements SoftwareComponentApiInterface.
func (api *SAP_COM_0510) initialRequest() error {
	// Loging into the ABAP System - getting the x-csrf-token and cookies
	resp, err := GetHTTPResponse("HEAD", api.con, nil, api.client)
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
}
