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
		return errors.Wrap(err, "could not create a Cookie Jar")
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
		return fmt.Errorf("pushing ATC System Configuration failed (File: "+config.AtcSystemConfigFilePath+") - %w", err)
	}
	var atcSystemConfiguartionJsonFile []byte
	filename, err := filepath.Abs(filelocation[0])
	if err != nil {
		return fmt.Errorf("pushing ATC System Configuration failed (File: "+config.AtcSystemConfigFilePath+") - %w", err)
	}
	atcSystemConfiguartionJsonFile, err = ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("pushing ATC System Configuration failed (File: "+config.AtcSystemConfigFilePath+") - %w", err)
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
		err = abaputils.HandleHTTPError(resp, err, "authentication on the ABAP system failed", connectionDetails)
		return err
	}
	defer resp.Body.Close()

	log.Entry().WithField("StatusCode", resp.Status).WithField("ABAP Endpoint", connectionDetails.URL).Debug("Authentication on the ABAP system successful")
	uriConnectionDetails.XCsrfToken = resp.Header.Get("X-Csrf-Token")
	connectionDetails.XCsrfToken = uriConnectionDetails.XCsrfToken

	//first check, if ATC configuration with given name already exists
	configDoesExist, configName, configUUID, err := checkConfigExists(config, atcSystemConfiguartionJsonFile, connectionDetails, client)
	if err != nil {
		return err
	}
	if !configDoesExist || (config.OverwriteExistingSystemConfig && configDoesExist) {
		if configDoesExist {
			err = doOverwriteATCSystemConfig(configName, configUUID, connectionDetails, client)
			if err != nil {
				return err
			}
		}
		if !configDoesExist {
			err = doPushATCSystemConfig(config, atcSystemConfiguartionJsonFile, connectionDetails, client)
			if err != nil {
				return err
			}
		}

	} else {
		log.Entry().Warningf("pushing ATC System Configuration skipped - Reason: ATC Configuration with same name '" + configName + "' (UUDI " + configUUID + ") already exists but OverwriteExistingSystemConfig is set to false.")
	}
	return nil

}

func doOverwriteATCSystemConfig(configName string, configUUID string, connectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) error {
	return nil
}

func doPushATCSystemConfig(config *abapEnvironmentPushATCSystemConfigOptions, atcSystemConfiguartionJsonFile []byte, connectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) error {
	abapEndpoint := connectionDetails.URL
	connectionDetails.URL = abapEndpoint + "/configuration"

	jsonBody := atcSystemConfiguartionJsonFile
	resp, err := abaputils.GetHTTPResponse("POST", connectionDetails, jsonBody, client)
	return parseOdataResponse(resp, err, connectionDetails, config)
}

func checkConfigExists(config *abapEnvironmentPushATCSystemConfigOptions, atcSystemConfiguartionJsonFile []byte, connectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) (bool, string, string, error) {
	var configName string
	var configUUID string

	//extract Configuration Name from atcSystemConfiguartionJsonFile
	var parsedConfigurationJson parsedConfigJson
	err := json.Unmarshal(atcSystemConfiguartionJsonFile, &parsedConfigurationJson)
	if err != nil {
		return false, configName, configUUID, err
	}

	//call a get on config with filter on given name
	configName = parsedConfigurationJson.ConfName
	abapEndpoint := connectionDetails.URL
	connectionDetails.URL = abapEndpoint + "/configuration" + "?$filter=conf_name%20eq%20" + "'" + configName + "'"
	if err != nil {
		return false, configName, configUUID, err
	}
	resp, err := abaputils.GetHTTPResponse("GET", connectionDetails, nil, client)
	if err != nil {
		return false, configName, configUUID, fmt.Errorf("oData response errors: %w", err)
	}
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, configName, configUUID, fmt.Errorf("parsing oData response failed: %w", err)
	}

	var parsedoDataResponse parsedOdataResp
	err = json.Unmarshal(body, &parsedoDataResponse)
	if err != nil {
		return false, configName, configUUID, fmt.Errorf("unmarshal oData response json failed: %w", err)
	}
	if len(parsedoDataResponse.Value) > 0 {
		configUUID = parsedoDataResponse.Value[0].ConfUUID
	}
	if configUUID == "" {
		return false, configName, configUUID, nil
	} else {
		return true, configName, configUUID, nil
	}

}

func parseOdataResponse(resp *http.Response, errorIn error, connectionDetails abaputils.ConnectionDetailsHTTP, config *abapEnvironmentPushATCSystemConfigOptions) error {

	if resp == nil {
		return errorIn
	}

	log.Entry().WithField("func", "parsedOdataResp: StatusCode").Info(resp.Status)

	switch resp.StatusCode {
	case 200: //Retrieved entities
		return nil

	case 201: //CREATED
		return nil

	case 400: //BAD REQUEST
		//Parse response
		var err error
		var body []byte
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("parsing oData response failed: %w", err)
		}
		if len(body) == 0 {
			return fmt.Errorf("parsing oData response failed: %w", errors.New("body is empty, can't parse empty body"))
		}
		var parsedOdataErrors interface{}
		err = json.Unmarshal(body, &parsedOdataErrors)
		if err != nil {
			return fmt.Errorf("unmarshal oData response json failed: %w", err)
		}
		err = extractErrAndLogMessages(parsedOdataErrors, config)
		if err != nil {
			return fmt.Errorf("bad Request Errors: %w", err)
		}

	default: //unhandled OK Code
		return fmt.Errorf("unhandled StatusCode: %w", errors.New(resp.Status))
	}

	defer resp.Body.Close()
	return nil
}

func extractErrAndLogMessages(parsedOdataMessages interface{}, config *abapEnvironmentPushATCSystemConfigOptions) error {
	var err error
	//find relevant messages to handle specially
	return err
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

type parsedOdataResp struct {
	Value []parsedConfigJson `json:"value"`
}

type parsedConfigJson struct {
	RootId   string `json:"root_id"`
	ConfName string `json:"conf_name"`
	ConfUUID string `json:"conf_id"`
}

type oDataResponseErrors []struct {
	Error oDataResponseError `json:"error"`
}

type oDataResponseError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Target     string `json:"target"`
	Details    []oDataResponseErrorDetail
	Innererror struct{}
}

type oDataResponseErrorDetail struct {
	code    string
	message string
	target  string
}
