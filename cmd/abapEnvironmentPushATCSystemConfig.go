package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"path/filepath"
	"reflect"
	"strconv"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

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

	// Determine the host, user and password, either via the input parameters or via a cloud foundry service key.
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
	parsedConfigurationJson, atcSystemConfiguartionJsonFile, err := checkATCSystemConfigurationFile(config)
	if err != nil {
		return err
	}
	//check, if ATC configuration with given name already exists in Backend
	configDoesExist, configName, configUUID, configLastChangedBackend, err := checkConfigExistsInBackend(config, atcSystemConfiguartionJsonFile, connectionDetails, client)
	if err != nil {
		return err
	}

	if !configDoesExist {
		//regular push of configuration
		configUUID = ""
		return handlePushConfiguration(config, configUUID, configDoesExist, atcSystemConfiguartionJsonFile, connectionDetails, client)
	}
	if !parsedConfigurationJson.LastChangedAt.IsZero() {
		if configLastChangedBackend.Before(parsedConfigurationJson.LastChangedAt) && !config.PatchIfExisting {
			//config exists, is not recent but must NOT be patched
			log.Entry().Info("pushing ATC System Configuration skipped. Reason: ATC System Configuration with name " + configName + " exists and is outdated (Backend: " + configLastChangedBackend.Local().String() + " vs. File: " + parsedConfigurationJson.LastChangedAt.Local().String() + ") but should not be overwritten (check step configuration parameter).")
			return nil
		}
		if configLastChangedBackend.After(parsedConfigurationJson.LastChangedAt) || configLastChangedBackend == parsedConfigurationJson.LastChangedAt {
			//configuration exists and is most recent
			log.Entry().Info("pushing ATC System Configuration skipped. Reason: ATC System Configuration with name " + configName + " exists and is most recent (Backend: " + configLastChangedBackend.Local().String() + " vs. File: " + parsedConfigurationJson.LastChangedAt.Local().String() + "). Therefore no update needed.")
			return nil
		}
	}
	if configLastChangedBackend.Before(parsedConfigurationJson.LastChangedAt) || parsedConfigurationJson.LastChangedAt.IsZero() {
		if config.PatchIfExisting {
			//configuration exists and is older than current config (or does not provide information about lastChanged) and should be patched
			return handlePushConfiguration(config, configUUID, configDoesExist, atcSystemConfiguartionJsonFile, connectionDetails, client)
		} else {
			//config exists, is not recent but must NOT be patched
			log.Entry().Info("pushing ATC System Configuration skipped. Reason: ATC System Configuration with name " + configName + " exists but should not be overwritten (check step configuration parameter).")
			return nil
		}
	}

	return nil
}

func checkATCSystemConfigurationFile(config *abapEnvironmentPushATCSystemConfigOptions) (parsedConfigJsonWithExpand, []byte, error) {
	var parsedConfigurationJson parsedConfigJsonWithExpand
	var emptyConfigurationJson parsedConfigJsonWithExpand
	var atcSystemConfiguartionJsonFile []byte

	parsedConfigurationJson, atcSystemConfiguartionJsonFile, err := readATCSystemConfigurationFile(config)
	if err != nil {
		return parsedConfigurationJson, atcSystemConfiguartionJsonFile, err
	}

	//check if parsedConfigurationJson is not initial or Configuration Name not supplied
	if reflect.DeepEqual(parsedConfigurationJson, emptyConfigurationJson) ||
		parsedConfigurationJson.ConfName == "" {
		return parsedConfigurationJson, atcSystemConfiguartionJsonFile, fmt.Errorf("pushing ATC System Configuration failed. Reason: Configured File does not contain required ATC System Configuration attributes (File: " + config.AtcSystemConfigFilePath + ")")
	}

	return parsedConfigurationJson, atcSystemConfiguartionJsonFile, nil
}

func readATCSystemConfigurationFile(config *abapEnvironmentPushATCSystemConfigOptions) (parsedConfigJsonWithExpand, []byte, error) {
	var parsedConfigurationJson parsedConfigJsonWithExpand
	var emptyConfigurationJson parsedConfigJsonWithExpand
	var atcSystemConfiguartionJsonFile []byte
	var filename string

	filelocation, err := filepath.Glob(config.AtcSystemConfigFilePath)
	if err != nil {
		return parsedConfigurationJson, atcSystemConfiguartionJsonFile, err
	}

	if len(filelocation) == 0 {
		return parsedConfigurationJson, atcSystemConfiguartionJsonFile, fmt.Errorf("pushing ATC System Configuration failed. Reason: Configured Filelocation is empty (File: " + config.AtcSystemConfigFilePath + ")")
	}

	filename, err = filepath.Abs(filelocation[0])
	if err != nil {
		return parsedConfigurationJson, atcSystemConfiguartionJsonFile, err
	}
	atcSystemConfiguartionJsonFile, err = ioutil.ReadFile(filename)
	if err != nil {
		return parsedConfigurationJson, atcSystemConfiguartionJsonFile, err
	}
	if len(atcSystemConfiguartionJsonFile) == 0 {
		return parsedConfigurationJson, atcSystemConfiguartionJsonFile, fmt.Errorf("pushing ATC System Configuration failed. Reason: Configured File is empty (File: " + config.AtcSystemConfigFilePath + ")")
	}

	err = json.Unmarshal(atcSystemConfiguartionJsonFile, &parsedConfigurationJson)
	if err != nil {
		return emptyConfigurationJson, atcSystemConfiguartionJsonFile, err
	}

	return parsedConfigurationJson, atcSystemConfiguartionJsonFile, err
}

func handlePushConfiguration(config *abapEnvironmentPushATCSystemConfigOptions, confUUID string, configDoesExist bool, atcSystemConfiguartionJsonFile []byte, connectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) error {

	var err error
	connectionDetails.XCsrfToken, err = fetchXcsrfTokenFromHead(connectionDetails, client)
	if err != nil {
		return err
	}
	if configDoesExist {
		err = doPatchATCSystemConfig(config, confUUID, atcSystemConfiguartionJsonFile, connectionDetails, client)
		if err != nil {
			return err
		}
		log.Entry().Info("ATC System Configuration successfully pushed from file " + config.AtcSystemConfigFilePath + " and patched in system")
	}
	if !configDoesExist {
		err = doPushATCSystemConfig(config, atcSystemConfiguartionJsonFile, connectionDetails, client)
		if err != nil {
			return err
		}
		log.Entry().Info("ATC System Configuration successfully pushed from file " + config.AtcSystemConfigFilePath + " and created in system")
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

func doPatchATCSystemConfig(config *abapEnvironmentPushATCSystemConfigOptions, confUUID string, atcSystemConfiguartionJsonFile []byte, connectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) error {

	batchATCSystemConfigFile, err := buildATCSystemConfigBatchRequest(confUUID, atcSystemConfiguartionJsonFile)
	if err != nil {
		return err
	}
	return doBatchATCSystemConfig(config, batchATCSystemConfigFile, connectionDetails, client)

}

func buildATCSystemConfigBatchRequest(confUUID string, atcSystemConfiguartionJsonFile []byte) (string, error) {

	var batchRequestString string

	//splitting json into configuration base and configuration properties & build a batch request for oData - patch config & patch priorities
	var configBaseJson parsedConfigJsonBase
	err := json.Unmarshal(atcSystemConfiguartionJsonFile, &configBaseJson)
	if err != nil {
		return batchRequestString, err
	}
	var parsedConfigPriorities parsedConfigPriorities
	err = json.Unmarshal(atcSystemConfiguartionJsonFile, &parsedConfigPriorities)
	if err != nil {
		return batchRequestString, err
	}

	configBaseJson.ConfUUID = confUUID

	//build the Batch request string
	contentID := 1

	batchRequestString = addBeginOfBatch(batchRequestString)
	//now adding opening Changeset as at least config base is to be patched
	batchRequestString = addChangesetBegin(batchRequestString, contentID)

	configBaseJsonBody, err := json.Marshal(&configBaseJson)
	if err != nil {
		return batchRequestString, err
	}
	batchRequestString = addPatchConfigBaseChangeset(batchRequestString, confUUID, configBaseJsonBody)

	if len(parsedConfigPriorities.Priorities) > 0 {
		// in case Priorities need patches too
		var priority priorityJson
		for i, priorityLine := range parsedConfigPriorities.Priorities {

			//for each line, add content id
			contentID += 1
			priority.Priority = priorityLine.Priority
			priorityJsonBody, err := json.Marshal(&priority)
			if err != nil {
				log.Entry().Errorf("problem with marshall of single priority in line "+strconv.Itoa(i), err)
				continue
			}
			batchRequestString = addChangesetBegin(batchRequestString, contentID)

			//now PATCH command for priority
			batchRequestString = addPatchSinglePriorityChangeset(batchRequestString, confUUID, priorityLine.Test, priorityLine.MessageId, priorityJsonBody)

		}
	}

	//at the end, add closing inner and outer boundary tags
	batchRequestString = addEndOfBatch(batchRequestString)

	log.Entry().Info("Batch Request String: " + batchRequestString)

	return batchRequestString, nil

}

func addPatchConfigBaseChangeset(inputString string, confUUID string, configBaseJsonBody []byte) string {

	entityIdString := `(root_id='1',conf_id=` + confUUID + `)`
	newString := addCommandEntityChangeset("PATCH", "configuration", entityIdString, inputString, configBaseJsonBody)

	return newString
}

func addPatchSinglePriorityChangeset(inputString string, confUUID string, test string, messageId string, priorityJsonBody []byte) string {

	entityIdString := `(root_id='1',conf_id=` + confUUID + `,test='` + test + `',message_id='` + messageId + `')`
	newString := addCommandEntityChangeset("PATCH", "priority", entityIdString, inputString, priorityJsonBody)

	return newString
}

func addChangesetBegin(inputString string, contentID int) string {

	newString := inputString + `
--changeset
Content-Type: application/http
Content-Transfer-Encoding: binary
Content-ID: ` + strconv.Itoa(contentID) + `
`
	return newString
}

func addBeginOfBatch(inputString string) string {
	//Starting always with outer boundary - followed by mandatory Contenttype and boundary for changeset
	newString := inputString + `
--request-separator
Content-Type: multipart/mixed;boundary=changeset
`
	return newString
}

func addEndOfBatch(inputString string) string {
	//Starting always with outer boundary - followed by mandatory Contenttype and boundary for changeset
	newString := inputString + `
--changeset--

--request-separator--`

	return newString
}

func addCommandEntityChangeset(command string, entity string, entityIdString string, inputString string, jsonBody []byte) string {

	newString := inputString + `
` + command + ` ` + entity + entityIdString + ` HTTP/1.1
Content-Type: application/json

`
	if len(jsonBody) > 0 {
		newString += string(jsonBody) + `
`
	}

	return newString

}

func doPushATCSystemConfig(config *abapEnvironmentPushATCSystemConfigOptions, atcSystemConfiguartionJsonFile []byte, connectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) error {
	abapEndpoint := connectionDetails.URL
	connectionDetails.URL = abapEndpoint + "/configuration"

	resp, err := abaputils.GetHTTPResponse("POST", connectionDetails, atcSystemConfiguartionJsonFile, client)
	return parseOdataResponse(resp, err, connectionDetails, config)
}

func doBatchATCSystemConfig(config *abapEnvironmentPushATCSystemConfigOptions, batchRequestBodyFile string, connectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) error {
	abapEndpoint := connectionDetails.URL
	connectionDetails.URL = abapEndpoint + "/$batch"

	header := make(map[string][]string)
	header["X-Csrf-Token"] = []string{connectionDetails.XCsrfToken}
	header["Content-Type"] = []string{"multipart/mixed;boundary=request-separator"}

	batchRequestBodyFileByte := []byte(batchRequestBodyFile)
	resp, err := client.SendRequest("POST", connectionDetails.URL, bytes.NewBuffer(batchRequestBodyFileByte), header, nil)

	return parseOdataResponse(resp, err, connectionDetails, config)
}

func checkConfigExistsInBackend(config *abapEnvironmentPushATCSystemConfigOptions, atcSystemConfiguartionJsonFile []byte, connectionDetails abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) (bool, string, string, time.Time, error) {
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

func parseOdataResponse(resp *http.Response, errorIn error, connectionDetails abaputils.ConnectionDetailsHTTP, config *abapEnvironmentPushATCSystemConfigOptions) error {

	if resp == nil {
		return errorIn
	}

	log.Entry().Info("parsedResp: StatusCode: " + resp.Status)

	var err error
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		defer resp.Body.Close()
		return fmt.Errorf("parsing response failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200: //Retrieved entities & OK in Patch & OK in Batch
		log.Entry().Infof("parsedRespBody: " + string(body))

	case 201: //CREATED
		log.Entry().Infof("parsedRespBody: " + string(body))

	case 400: //BAD REQUEST
		//no errorIn, Error in Body
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
		return fmt.Errorf("bad Request Errors: %v", parsedOdataErrors)

	default: //unhandled OK Code
		return fmt.Errorf("unhandled StatusCode: "+resp.Status, errorIn)
	}

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
	ConfName      string                 `json:"conf_name"`
	ConfUUID      string                 `json:"conf_id"`
	LastChangedAt time.Time              `json:"last_changed_at"`
	Priorities    []parsedConfigPriority `json:"_priorities"`
}

type parsedConfigJsonBase struct {
	ConfName            string `json:"conf_name"`
	ConfUUID            string `json:"conf_id"`
	Checkvariant        string `json:"checkvariant"`
	PseudoCommentPolicy string `json:"pseudo_comment_policy"`
	BlockFindings       string `json:"block_findings"`
	InformFindings      string `json:"inform_findings"`
	IsDefault           bool   `json:"is_default"`
	IsProxyVariant      bool   `json:"is_proxy_variant"`
}

type parsedConfigPriorities struct {
	Priorities []parsedConfigPriority `json:"_priorities"`
}

type parsedConfigPriority struct {
	Test      string      `json:"test"`
	MessageId string      `json:"message_id"`
	Priority  json.Number `json:"priority"`
}

type priorityJson struct {
	Priority json.Number `json:"priority"`
}
