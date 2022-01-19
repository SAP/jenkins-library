package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
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
	fileExists, err := newAbapEnvironmentPushATCSystemConfigUtils().FileExists(config.AtcSystemConfigFilePath)
	if err != nil {
		return err
	}
	if fileExists {
		filelocation, err := filepath.Glob(config.AtcSystemConfigFilePath)
		if err != nil {
			return err
		}
		if len(filelocation) > 0 {
			var atcSystemConfiguartionJsonFile []byte
			filename, err := filepath.Abs(filelocation[0])
			if err != nil {
				return err
			}
			atcSystemConfiguartionJsonFile, err = ioutil.ReadFile(filename)
			if err != nil {
				return err
			}
			if len(atcSystemConfiguartionJsonFile) == 0 {
				return fmt.Errorf("pushing ATC System Configuration failed. Reason: Configured File is empty (File: " + config.AtcSystemConfigFilePath + ")")
			}
			//check, if ATC configuration with given name already exists
			configDoesExist, configName, configUUID, configLastChangedBE, err := checkConfigExists(config, atcSystemConfiguartionJsonFile, connectionDetails, client)
			if err != nil {
				return err
			}
			var parsedConfigurationJson parsedConfigJsonWithExpand
			err = json.Unmarshal(atcSystemConfiguartionJsonFile, &parsedConfigurationJson)
			if err != nil {
				return err
			}
			if !configDoesExist {
				configUUID = ""
				return handlePushConfiguration(config, filename, configUUID, atcSystemConfiguartionJsonFile, connectionDetails, client)
			}
			if configDoesExist && configLastChangedBE.Before(parsedConfigurationJson.LastChangedAt) && !config.PatchExistingSystemConfig {
				//config exists, is not recent but must NOT be patched
				log.Entry().Warn("pushing ATC System Configuration skipped. Reason: ATC System Configuration with name " + configName + " exists and is outdated (Backend: " + configLastChangedBE.Local().String() + " vs. File: " + parsedConfigurationJson.LastChangedAt.Local().String() + ") but should not be overwritten (check step configuration).")
				return nil
			}
			if configDoesExist && configLastChangedBE.After(parsedConfigurationJson.LastChangedAt) {
				//configuration exists and is most recent
				log.Entry().Info("pushing ATC System Configuration skipped. Reason: ATC System Configuration with name " + configName + " exists and is most recent (Backend: " + configLastChangedBE.Local().String() + " vs. File: " + parsedConfigurationJson.LastChangedAt.Local().String() + "). Therefore no update needed.")
				return nil
			}
			if configDoesExist && configLastChangedBE.Before(parsedConfigurationJson.LastChangedAt) && config.PatchExistingSystemConfig {
				//configuration exists and is older than current config and should be patched
				return handlePushConfiguration(config, filename, configUUID, atcSystemConfiguartionJsonFile, connectionDetails, client)
			}
		} else {
			return fmt.Errorf("pushing ATC System Configuration failed. Reason: Configured Filelocation is empty (File: "+config.AtcSystemConfigFilePath+") - %w", err)
		}
	} else {
		return fmt.Errorf("pushing ATC System Configuration failed. Reason: Configured File does not exist(File: "+config.AtcSystemConfigFilePath+") - %w", err)
	}
	return nil
}

func handlePushConfiguration(config *abapEnvironmentPushATCSystemConfigOptions, validFilename string, configUUID string, atcSystemConfiguartionJsonFile []byte, connectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) error {
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

	if configUUID != "" {
		err = doPatchATCSystemConfig(config, validFilename, configUUID, atcSystemConfiguartionJsonFile, connectionDetails, client)
		if err != nil {
			return err
		}
	}
	if configUUID == "" {
		err = doPushATCSystemConfig(config, validFilename, atcSystemConfiguartionJsonFile, connectionDetails, client)
		if err != nil {
			return err
		}
	}

	return nil

}

func doPatchATCSystemConfig(config *abapEnvironmentPushATCSystemConfigOptions, validFilename string, confUUID string, atcSystemConfiguartionJsonFile []byte, connectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) error {
	abapEndpoint := connectionDetails.URL

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

func checkConfigExists(config *abapEnvironmentPushATCSystemConfigOptions, atcSystemConfiguartionJsonFile []byte, connectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) (bool, string, string, time.Time, error) {
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
	switch resp.StatusCode {
	case 200: //Retrieved entities & OK in Patch
		var patchedATCSystemConfig []byte
		var permWrite fs.FileMode
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("parsing oData response failed: %w", err)
		}
		patchedATCSystemConfig = body
		var parsedConfigurationJson parsedConfigJsonWithExpand
		err := json.Unmarshal(patchedATCSystemConfig, &parsedConfigurationJson)
		if err != nil {
			return err
		}
		//in case it was an pact on configuration base, this value may not be initial!
		if parsedConfigurationJson.ConfName != "" {
			//write Patched Config Base Info back to File
			patchedATCSystemConfig, err = json.MarshalIndent(&parsedConfigurationJson, "", "\t")
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(validFilename, patchedATCSystemConfig, permWrite)
			if err != nil {
				return err
			}
			log.Entry().Info("ATC System configuration file successfully updated " + validFilename + " with patched ATC System configuration information.")
			defer resp.Body.Close()
			return nil
		}

		var parsedPriorityJson parsedConfigPriority
		err = json.Unmarshal(patchedATCSystemConfig, &parsedPriorityJson)
		if err != nil {
			return err
		}
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
			patchedATCSystemConfig, err = json.MarshalIndent(&parsedConfigurationJson, "", "\t")
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(validFilename, patchedATCSystemConfig, permWrite)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
		}

		return nil

	case 201: //CREATED
		log.Entry().Info("ATC System configuration successfully pushed from file " + validFilename + " to system")
		//save response entity as
		var createdATCSystemConfig []byte
		var permWrite fs.FileMode
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("parsing oData response failed: %w", err)
		}
		createdATCSystemConfig = body
		var parsedConfigurationJson parsedConfigJsonWithExpand
		err := json.Unmarshal(createdATCSystemConfig, &parsedConfigurationJson)
		if err != nil {
			return err
		}
		createdATCSystemConfig, err = json.MarshalIndent(&parsedConfigurationJson, "", "\t")
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(validFilename, createdATCSystemConfig, permWrite)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		log.Entry().Info("ATC System configuration file successfully updated " + validFilename + " with created ATC System configuration information.")
		return nil

	case 400: //BAD REQUEST
		//Parse response
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
		defer resp.Body.Close()
		return fmt.Errorf("bad Request Errors: %v", parsedOdataErrors)
		/* 		err = extractErrAndLogMessages(parsedOdataErrors, config)
		   		if err != nil {
		   			return fmt.Errorf("bad Request Errors: %w", err)
		   		} */

	default: //unhandled OK Code
		defer resp.Body.Close()
		return fmt.Errorf("unhandled StatusCode: %w", errors.New(resp.Status))
	}

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
