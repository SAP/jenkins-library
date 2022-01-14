package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"path/filepath"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

type abapEnvironmentPushATCSystemConfigUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The abapEnvironmentPushATCSystemConfigUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type abapEnvironmentPushATCSystemConfigUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to abapEnvironmentPushATCSystemConfigUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// abapEnvironmentPushATCSystemConfigUtilsBundle and forward to the implementation of the dependency.
}

func newAbapEnvironmentPushATCSystemConfigUtils() abapEnvironmentPushATCSystemConfigUtils {
	utils := abapEnvironmentPushATCSystemConfigUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}

	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())

	return &utils
}

func abapEnvironmentPushATCSystemConfig(config abapEnvironmentPushATCSystemConfigOptions, telemetryData *telemetry.CustomData) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	var autils = abaputils.AbapUtils{
		Exec: &c,
	}

	client := piperhttp.Client{}

	err := runAbapEnvironmentPushATCSystemConfig(&config, telemetryData, &autils, &client)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentPushATCSystemConfig(config *abapEnvironmentPushATCSystemConfigOptions, telemetryData *telemetry.CustomData, autils abaputils.Communication, client piperhttp.Sender) error {

	subOptions := convertATCSysOptions(config)

	// Determine the host, user and password, either via the input parameters or via a cloud foundry service key
	connectionDetails, err := autils.GetAbapCommunicationArrangementInfo(subOptions, "/sap/opu/odata4/sap/satc_ci_cf_api/srvd_a2x/sap/satc_ci_cf_sv_api/0001")
	if err != nil {
		return errors.Wrap(err, "Parameters for the ABAP Connection not available")
	}

	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return errors.Wrap(err, "Could not create a Cookie Jar")
	}
	clientOptions := piperhttp.ClientOptions{
		MaxRequestDuration: 180 * time.Second,
		CookieJar:          cookieJar,
		Username:           connectionDetails.User,
		Password:           connectionDetails.Password,
	}
	client.SetOptions(clientOptions)

	return pushATCSystemConfig(config, connectionDetails, client)

}

func pushATCSystemConfig(config *abapEnvironmentPushATCSystemConfigOptions, connectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) error {

	//check ATC system configuration json
	filelocation, err := filepath.Glob(config.AtcSystemConfigFilePath)
	if err != nil {
		return fmt.Errorf("Pushing ATC System Configuration failed (File: "+config.AtcSystemConfigFilePath+") - %w", err)
	}
	var atcSystemConfiguartionJsonFile []byte
	filename, err := filepath.Abs(filelocation[0])
	if err != nil {
		return fmt.Errorf("Pushing ATC System Configuration failed (File: "+config.AtcSystemConfigFilePath+") - %w", err)
	}
	atcSystemConfiguartionJsonFile, err = ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Pushing ATC System Configuration failed (File: "+config.AtcSystemConfigFilePath+") - %w", err)
	}

	return handlePushConfiguration(config, atcSystemConfiguartionJsonFile, connectionDetails, client)

}

func handlePushConfiguration(config *abapEnvironmentPushATCSystemConfigOptions, atcSystemConfiguartionJsonFile []byte, connectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) error {
	uriConnectionDetails := connectionDetails
	uriConnectionDetails.URL = ""
	connectionDetails.XCsrfToken = "fetch"

	// Loging into the ABAP System - getting the x-csrf-token and cookies
	resp, err := abaputils.GetHTTPResponse("HEAD", connectionDetails, nil, client)
	if err != nil {
		err = abaputils.HandleHTTPError(resp, err, "Authentication on the ABAP system failed", connectionDetails)
		return err
	}
	defer resp.Body.Close()

	log.Entry().WithField("StatusCode", resp.Status).WithField("ABAP Endpoint", connectionDetails.URL).Debug("Authentication on the ABAP system successful")
	uriConnectionDetails.XCsrfToken = resp.Header.Get("X-Csrf-Token")
	connectionDetails.XCsrfToken = uriConnectionDetails.XCsrfToken

	abapEndpoint := connectionDetails.URL
	connectionDetails.URL = abapEndpoint + "/configuration"

	jsonBody := atcSystemConfiguartionJsonFile
	resp, err = abaputils.GetHTTPResponse("POST", connectionDetails, jsonBody, client)
	return parseOdataResponse(resp, err, uriConnectionDetails)

}

func parseOdataResponse(resp *http.Response, errorIn error, connectionDetails abaputils.ConnectionDetailsHTTP) error {

	if resp == nil {
		return errorIn
	}

	log.Entry().WithField("func", "parsedOdataResp: StatusCode").Info(resp.Status)

	switch resp.StatusCode {
	case 201: //CREATED
		return nil

	case 400: //BAD REQUEST
		//Parse response
		var err error
		var body []byte
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Parsing oData response failed: %w", err)
		}
		if len(body) == 0 {
			return fmt.Errorf("Parsing oData response failed: %w", errors.New("Body is empty, can't parse empty body"))
		}
		var parsedOdataErrors oDataResponseErrors
		err = json.Unmarshal(body, &parsedOdataErrors)
		if err != nil {
			return fmt.Errorf("Unmarshal oData response json failed: %w", err)
		}
		err = extractErrorMessages(parsedOdataErrors)
		return fmt.Errorf("Bad Request Errors: %w", err)

		if err != nil {
			return fmt.Errorf("Parsing oData result failed: %w", err)
		}

	default: //unhandled OK Code
		return fmt.Errorf("Unhandled StatusCode: %w", resp.Status)
	}

	defer resp.Body.Close()
	return nil
}

func extractErrorMessages(parsedOdataErrors oDataResponseErrors) error {
	var errorMessages []string
	/* 	switch parsedOdataErrors.(type) {
	   	case map[string]interface{}:
	   		parsedOdataErrorsTab := parsedOdataErrors.(map[string]interface{})
	   		fmt.Printf("message", parsedOdataErrorsTab["error"])
	   	case []interface{}:
	   		parsedOdataErrorsList := parsedOdataErrors.([]interface{})
	   		fmt.Printf("error", parsedOdataErrorsList[1])
	   	default:
	   		panic(fmt.Errorf("type %T unexpected", parsedOdataErrors))
	   	} */
	/* 	if errorMessage != "" {
		errorMessages = append(errorMessages, errorMessage)
	} */
	errorMessages = append(errorMessages, "Messages:")

	return &responseError{}
}

func convertATCSysOptions(options *abapEnvironmentPushATCSystemConfigOptions) abaputils.AbapEnvironmentOptions {
	subOptions := abaputils.AbapEnvironmentOptions{}

	subOptions.CfAPIEndpoint = options.CfAPIEndpoint
	subOptions.CfServiceInstance = options.CfServiceInstance
	subOptions.CfServiceKeyName = options.CfServiceKeyName
	subOptions.CfOrg = options.CfOrg
	subOptions.CfSpace = options.CfSpace
	subOptions.Host = options.Host
	subOptions.Password = options.Password
	subOptions.Username = options.Username

	return subOptions
}

func (responseError *responseError) Error() string {
	return "Messages: "
}

type responseError struct {
}

type oDataResponseErrors []struct {
	error oDataResponseError `json:"error"`
}

type oDataResponseError struct {
	code       string `json:"code"`
	message    string `json:"message"`
	target     string `json:"target"`
	details    []oDataResponseErrorDetail
	innererror struct{}
}

type oDataResponseErrorDetail struct {
	code    string
	message string
	target  string
}
