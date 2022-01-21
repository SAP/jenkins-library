package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"path/filepath"
	"reflect"
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

	//check, if given ATC System configuration File
	parsedConfigurationJson, atcSystemConfiguartionJsonFile, validFilename, err := checkATCSystemConfigurationFile(config)
	if err != nil {
		return err
	}
	//check, if ATC configuration with given name already exists in Backend
	configDoesExist, configName, configUUID, configLastChangedBE, err := checkConfigExistsInBE(config, atcSystemConfiguartionJsonFile, connectionDetails, client)
	if err != nil {
		return err
	}
	if !configDoesExist {
		//regular push of configuration
		configUUID = ""
		return handlePushConfiguration(config, validFilename, configUUID, configDoesExist, atcSystemConfiguartionJsonFile, connectionDetails, client)
	}
	if configDoesExist && configLastChangedBE.Before(parsedConfigurationJson.LastChangedAt) && !config.PatchIfExistingAndOutdated {
		//config exists, is not recent but must NOT be patched
		log.Entry().Warn("pushing ATC System Configuration skipped. Reason: ATC System Configuration with name " + configName + " exists and is outdated (Backend: " + configLastChangedBE.Local().String() + " vs. File: " + parsedConfigurationJson.LastChangedAt.Local().String() + ") but should not be overwritten (check step configuration parameter).")
		return nil
	}
	if configDoesExist && (configLastChangedBE.After(parsedConfigurationJson.LastChangedAt) || configLastChangedBE == parsedConfigurationJson.LastChangedAt) {
		//configuration exists and is most recent
		log.Entry().Info("pushing ATC System Configuration skipped. Reason: ATC System Configuration with name " + configName + " exists and is most recent (Backend: " + configLastChangedBE.Local().String() + " vs. File: " + parsedConfigurationJson.LastChangedAt.Local().String() + "). Therefore no update needed.")
		return nil
	}
	if configDoesExist && configLastChangedBE.Before(parsedConfigurationJson.LastChangedAt) && config.PatchIfExistingAndOutdated {
		//configuration exists and is older than current config and should be patched
		return handlePushConfiguration(config, validFilename, configUUID, configDoesExist, atcSystemConfiguartionJsonFile, connectionDetails, client)
	}

	return nil
}

func checkATCSystemConfigurationFile(config *abapEnvironmentPushATCSystemConfigOptions) (parsedConfigJsonWithExpand, []byte, string, error) {
	var parsedConfigurationJson parsedConfigJsonWithExpand
	var emptyConfigurationJson parsedConfigJsonWithExpand
	var atcSystemConfiguartionJsonFile []byte
	var filename string
	//check ATC system configuration json
	fileExists, err := newAbapEnvironmentPushATCSystemConfigUtils().FileExists(config.AtcSystemConfigFilePath)
	if err != nil {
		return parsedConfigurationJson, atcSystemConfiguartionJsonFile, filename, err
	}
	if !fileExists {
		return parsedConfigurationJson, atcSystemConfiguartionJsonFile, filename, fmt.Errorf("pushing ATC System Configuration failed. Reason: Configured File does not exist(File: " + config.AtcSystemConfigFilePath + ")")
	}

	filelocation, err := filepath.Glob(config.AtcSystemConfigFilePath)
	if err != nil {
		return parsedConfigurationJson, atcSystemConfiguartionJsonFile, filename, err
	}

	if len(filelocation) == 0 {
		return parsedConfigurationJson, atcSystemConfiguartionJsonFile, filename, fmt.Errorf("pushing ATC System Configuration failed. Reason: Configured Filelocation is empty (File: " + config.AtcSystemConfigFilePath + ")")
	}

	filename, err = filepath.Abs(filelocation[0])
	if err != nil {
		return parsedConfigurationJson, atcSystemConfiguartionJsonFile, filename, err
	}
	atcSystemConfiguartionJsonFile, err = ioutil.ReadFile(filename)
	if err != nil {
		return parsedConfigurationJson, atcSystemConfiguartionJsonFile, filename, err
	}
	if len(atcSystemConfiguartionJsonFile) == 0 {
		return parsedConfigurationJson, atcSystemConfiguartionJsonFile, filename, fmt.Errorf("pushing ATC System Configuration failed. Reason: Configured File is empty (File: " + config.AtcSystemConfigFilePath + ")")
	}

	err = json.Unmarshal(atcSystemConfiguartionJsonFile, &parsedConfigurationJson)
	if err != nil {
		return emptyConfigurationJson, atcSystemConfiguartionJsonFile, filename, err
	}
	//check if parsedConfigurationJson is not initial
	if reflect.DeepEqual(parsedConfigurationJson, emptyConfigurationJson) ||
		parsedConfigurationJson.ConfName == "" {
		return parsedConfigurationJson, atcSystemConfiguartionJsonFile, filename, fmt.Errorf("pushing ATC System Configuration failed. Reason: Configured File does not contain ATC System Configuration attributes (File: " + config.AtcSystemConfigFilePath + ")")
	}

	return parsedConfigurationJson, atcSystemConfiguartionJsonFile, filename, nil
}

func handlePushConfiguration(config *abapEnvironmentPushATCSystemConfigOptions, validFilename string, configUUID string, configDoesExist bool, atcSystemConfiguartionJsonFile []byte, connectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) error {

	var err error
	connectionDetails.XCsrfToken, err = fetchXcsrfTokenFromHead(connectionDetails, client)
	if err != nil {
		return err
	}
	if configDoesExist {
		err = doPatchATCSystemConfig(config, validFilename, configUUID, atcSystemConfiguartionJsonFile, connectionDetails, client)
		if err != nil {
			return err
		}
	}
	if !configDoesExist {
		err = doPushATCSystemConfig(config, validFilename, atcSystemConfiguartionJsonFile, connectionDetails, client)
		if err != nil {
			return err
		}
	}

	return nil

}

func fetchXcsrfTokenFromHead(connectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) (string, error) {

	log.Entry().WithField("ABAP Endpoint: ", connectionDetails.URL).Debug("Fetching Xcrsf-Token")
	uriConnectionDetails := connectionDetails
	uriConnectionDetails.URL = ""
	connectionDetails.XCsrfToken = "fetch"

	// Loging into the ABAP System - getting the x-csrf-token and cookies
	resp, err := abaputils.GetHTTPResponse("HEAD", connectionDetails, nil, client)
	if err != nil {
		err = abaputils.HandleHTTPError(resp, err, "authentication on the ABAP system failed", connectionDetails)
		return connectionDetails.XCsrfToken, err
	}
	defer resp.Body.Close()

	log.Entry().WithField("StatusCode", resp.Status).WithField("ABAP Endpoint", connectionDetails.URL).Debug("Authentication on the ABAP system successful")
	uriConnectionDetails.XCsrfToken = resp.Header.Get("X-Csrf-Token")
	connectionDetails.XCsrfToken = uriConnectionDetails.XCsrfToken

	return connectionDetails.XCsrfToken, err
}

func doPatchATCSystemConfig(config *abapEnvironmentPushATCSystemConfigOptions, validFilename string, confUUID string, atcSystemConfiguartionJsonFile []byte, connectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) error {
	abapEndpoint := connectionDetails.URL

	//TBD: Check if BATCH request can be formed instead of single calls!

	//splitting json into configuration base and configuration properties & build a batch request for oData - patch config & patch priorities
	var configBaseJson parsedConfigJsonBase
	err := json.Unmarshal(atcSystemConfiguartionJsonFile, &configBaseJson)
	if err != nil {
		return err
	}
	var parsedConfigPriorities parsedConfigPriorities
	err = json.Unmarshal(atcSystemConfiguartionJsonFile, &parsedConfigPriorities)
	if err != nil {
		return err
	}

	//Patch configuration
	//marshall into base config (no expanded priorities)
	configBaseJson.RootId = "1"
	configBaseJson.ConfUUID = confUUID
	configBaseJsonBody, err := json.Marshal(&configBaseJson)
	if err != nil {
		return err
	}
	//root ID is always 1!
	connectionDetails.URL = abapEndpoint + "/configuration(root_id='1',conf_id=" + confUUID + ")"
	resp, err := abaputils.GetHTTPResponse("PATCH", connectionDetails, configBaseJsonBody, client)
	err = parseOdataResponse(resp, err, connectionDetails, config, validFilename)
	if err != nil {
		return err
	}
	log.Entry().Info("ATC System configuration (Base) successfully patched from file" + validFilename)
	defer resp.Body.Close()

	if len(parsedConfigPriorities.Priorities) > 0 {
		//Patch message priorities
		// by now, PATCH needs to be done for each given priority
		var priority priorityJson
		for i, priorityLine := range parsedConfigPriorities.Priorities {
			connectionDetails.URL = abapEndpoint + "/priority(root_id='1',conf_id=" + confUUID + ",test='" + priorityLine.Test + "',message_id='" + priorityLine.MessageId + "')"
			priority.Priority = priorityLine.Priority
			priorityJsonBody, err := json.Marshal(&priority)
			if err != nil {
				log.Entry().Errorf("problem with marshall of single priority in line "+string(rune(i)), err)
				continue
			}
			resp, err = abaputils.GetHTTPResponse("PATCH", connectionDetails, priorityJsonBody, client)
			err = parseOdataResponse(resp, err, connectionDetails, config, validFilename)
			defer resp.Body.Close()
			if err != nil {
				log.Entry().Errorf("problem with response of patch of single priority in line "+string(rune(i)), err)
				continue
			}
		}
		log.Entry().Info("Message Priorities patched from file " + validFilename)
	}

	return nil
}

func doPushATCSystemConfig(config *abapEnvironmentPushATCSystemConfigOptions, validFilename string, atcSystemConfiguartionJsonFile []byte, connectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) error {
	abapEndpoint := connectionDetails.URL
	connectionDetails.URL = abapEndpoint + "/configuration"

	jsonBody := atcSystemConfiguartionJsonFile
	resp, err := abaputils.GetHTTPResponse("POST", connectionDetails, jsonBody, client)
	return parseOdataResponse(resp, err, connectionDetails, config, validFilename)
}

func checkConfigExistsInBE(config *abapEnvironmentPushATCSystemConfigOptions, atcSystemConfiguartionJsonFile []byte, connectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) (bool, string, string, time.Time, error) {
	var configName string
	var configUUID string
	var configLastChangedAt time.Time

	//extract Configuration Name from atcSystemConfiguartionJsonFile
	var parsedConfigurationJson parsedConfigJsonWithExpand
	err := json.Unmarshal(atcSystemConfiguartionJsonFile, &parsedConfigurationJson)
	if err != nil {
		return false, configName, configUUID, configLastChangedAt, err
	}

	//call a get on config with filter on given name
	configName = parsedConfigurationJson.ConfName
	abapEndpoint := connectionDetails.URL
	connectionDetails.URL = abapEndpoint + "/configuration" + "?$filter=conf_name%20eq%20" + "'" + configName + "'"
	if err != nil {
		return false, configName, configUUID, configLastChangedAt, err
	}
	resp, err := abaputils.GetHTTPResponse("GET", connectionDetails, nil, client)
	if err != nil {
		return false, configName, configUUID, configLastChangedAt, err
	}
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, configName, configUUID, configLastChangedAt, err
	}

	var parsedoDataResponse parsedOdataResp
	err = json.Unmarshal(body, &parsedoDataResponse)
	if err != nil {
		return false, configName, configUUID, configLastChangedAt, err
	}
	if len(parsedoDataResponse.Value) > 0 {
		configUUID = parsedoDataResponse.Value[0].ConfUUID
		configLastChangedAt = parsedoDataResponse.Value[0].LastChangedAt
		log.Entry().Info("ATC System Configuration " + configName + " does exist and last changed at " + configLastChangedAt.Local().String())
		return true, configName, configUUID, configLastChangedAt, nil
	} else {
		//response value is empty, so NOT found entity with this name!
		log.Entry().Info("ATC System Configuration " + configName + " does not exist!")
		return false, configName, "", configLastChangedAt, nil
	}
}

func parseOdataResponse(resp *http.Response, errorIn error, connectionDetails abaputils.ConnectionDetailsHTTP, config *abapEnvironmentPushATCSystemConfigOptions, validFilename string) error {

	if resp == nil {
		return errorIn
	}

	log.Entry().Info("parsedOdataResp: StatusCode: " + resp.Status)

	var err error
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("parsing oData response failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200: //Retrieved entities & OK in Patch
		return logAndPersistResponseBody(body, validFilename, errorIn)

	case 201: //CREATED
		log.Entry().Info("ATC System configuration successfully pushed from file " + validFilename + " to system")
		return logAndPersistResponseBody(body, validFilename, errorIn)

	case 400: //BAD REQUEST
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
		defer resp.Body.Close()
		return fmt.Errorf("bad Request Errors: %v", parsedOdataErrors)

	default: //unhandled OK Code
		defer resp.Body.Close()
		return fmt.Errorf("unhandled StatusCode: %w", errors.New(resp.Status))
	}

}

func logAndPersistResponseBody(body []byte, validFilename string, errorIn error) error {
	var err error
	var parsedConfigurationJson parsedConfigJsonWithExpand
	err = json.Unmarshal(body, &parsedConfigurationJson)
	if err != nil {
		return err
	}
	//in case it was an configuration, this value may not be initial!
	if parsedConfigurationJson.ConfName != "" {
		//write Patched Config Base Info back to File
		returnedATCSystemConfig, err := json.MarshalIndent(&parsedConfigurationJson, "", "\t")
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(validFilename, returnedATCSystemConfig, 0644)
		if err != nil {
			return err
		}
		log.Entry().Info("ATC System configuration file successfully updated " + validFilename + " with patched ATC System configuration information.")

		return nil
	}

	var parsedPriorityJson parsedConfigPriority
	err = json.Unmarshal(body, &parsedPriorityJson)
	if err != nil {
		return err
	}
	//in case it was an priority of the configuration, this value may not be initial!
	if parsedPriorityJson.MessageId != "" {
		atcSystemConfiguartionJsonFile, err := ioutil.ReadFile(validFilename)
		if err != nil {
			return err
		}
		err = json.Unmarshal(atcSystemConfiguartionJsonFile, &parsedConfigurationJson)
		if err != nil {
			return err
		}
		parsedConfigurationJson.Priorities = append(parsedConfigurationJson.Priorities, parsedPriorityJson)
		updatedATCSystemConfig, err := json.MarshalIndent(&parsedConfigurationJson, "", "\t")
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(validFilename, updatedATCSystemConfig, 0644)
		if err != nil {
			return err
		}

	}

	log.Entry().Info("ATC System configuration file successfully updated " + validFilename + " with created ATC System configuration information.")
	return nil

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

type parsedOdataResp struct {
	Value []parsedConfigJsonWithExpand `json:"value"`
}

type parsedConfigJsonWithExpand struct {
	RootId         string                 `json:"root_id"`
	ConfName       string                 `json:"conf_name"`
	ConfUUID       string                 `json:"conf_id"`
	Checkvariant   string                 `json:"checkvariant"`
	LastChangedAt  time.Time              `json:"last_changed_at"`
	BlockFindings  string                 `json:"block_findings"`
	InformFindings string                 `json:"inform_findings"`
	IsDefault      bool                   `json:"is_default"`
	IsProxyVariant bool                   `json:"is_proxy_variant"`
	Priorities     []parsedConfigPriority `json:"_priorities"`
}

type parsedConfigJsonBase struct {
	RootId         string `json:"root_id"`
	ConfName       string `json:"conf_name"`
	ConfUUID       string `json:"conf_id"`
	Checkvariant   string `json:"checkvariant"`
	BlockFindings  string `json:"block_findings"`
	InformFindings string `json:"inform_findings"`
	IsDefault      bool   `json:"is_default"`
	IsProxyVariant bool   `json:"is_proxy_variant"`
}

type parsedConfigPriorities struct {
	Priorities []parsedConfigPriority `json:"_priorities"`
}

type parsedConfigPriority struct {
	RootId          string      `json:"root_id"`
	ConfUUID        string      `json:"conf_id"`
	Test            string      `json:"test"`
	MessageId       string      `json:"message_id"`
	DefaultPriority json.Number `json:"default_priority"`
	Priority        json.Number `json:"priority"`
}

type priorityJson struct {
	Priority json.Number `json:"priority"`
}
