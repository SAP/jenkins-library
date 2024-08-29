package abaputils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

// AbapUtils Struct
type AbapUtils struct {
	Exec      command.ExecRunner
	Intervall time.Duration
}

/*
Communication for defining function used for communication
*/
type Communication interface {
	GetAbapCommunicationArrangementInfo(options AbapEnvironmentOptions, oDataURL string) (ConnectionDetailsHTTP, error)
	GetPollIntervall() time.Duration
}

// GetAbapCommunicationArrangementInfo function fetches the communcation arrangement information in SAP CP ABAP Environment
func (abaputils *AbapUtils) GetAbapCommunicationArrangementInfo(options AbapEnvironmentOptions, oDataURL string) (ConnectionDetailsHTTP, error) {
	c := abaputils.Exec
	var connectionDetails ConnectionDetailsHTTP
	var error error

	if options.Host != "" {
		// Host, User and Password are directly provided -> check for host schema (double https)
		match, err := regexp.MatchString(`^(https|HTTPS):\/\/.*`, options.Host)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return connectionDetails, errors.Wrap(err, "Schema validation for host parameter failed. Check for https.")
		}
		var hostOdataURL = options.Host + oDataURL
		if match {
			connectionDetails.URL = hostOdataURL
			connectionDetails.Host = options.Host
		} else {
			connectionDetails.URL = "https://" + hostOdataURL
			connectionDetails.Host = "https://" + options.Host
		}
		connectionDetails.User = options.Username
		connectionDetails.Password = options.Password
	} else {
		if options.CfAPIEndpoint == "" || options.CfOrg == "" || options.CfSpace == "" || options.CfServiceInstance == "" || options.CfServiceKeyName == "" {
			var err = errors.New("Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry API Endpoint, Organization, Space, Service Instance and Service Key")
			log.SetErrorCategory(log.ErrorConfiguration)
			return connectionDetails, err
		}
		// Url, User and Password should be read from a cf service key
		var abapServiceKey, error = ReadServiceKeyAbapEnvironment(options, c)
		if error != nil {
			return connectionDetails, errors.Wrap(error, "Read service key failed")
		}
		connectionDetails.Host = abapServiceKey.URL
		connectionDetails.URL = abapServiceKey.URL + oDataURL
		connectionDetails.User = abapServiceKey.Abap.Username
		connectionDetails.Password = abapServiceKey.Abap.Password
	}
	return connectionDetails, error
}

// ReadServiceKeyAbapEnvironment from Cloud Foundry and returns it. Depending on user/developer requirements if he wants to perform further Cloud Foundry actions
func ReadServiceKeyAbapEnvironment(options AbapEnvironmentOptions, c command.ExecRunner) (AbapServiceKey, error) {

	var abapServiceKeyV8 AbapServiceKeyV8
	var abapServiceKey AbapServiceKey
	var serviceKeyJSON string
	var err error

	cfconfig := cloudfoundry.ServiceKeyOptions{
		CfAPIEndpoint:     options.CfAPIEndpoint,
		CfOrg:             options.CfOrg,
		CfSpace:           options.CfSpace,
		CfServiceInstance: options.CfServiceInstance,
		CfServiceKeyName:  options.CfServiceKeyName,
		Username:          options.Username,
		Password:          options.Password,
	}

	cf := cloudfoundry.CFUtils{Exec: c}

	serviceKeyJSON, err = cf.ReadServiceKey(cfconfig)

	if err != nil {
		// Executing cfReadServiceKeyScript failed
		return abapServiceKeyV8.Credentials, err
	}

	// Depending on the cf cli version, the service key may be returned in a different format. For compatibility reason, both formats are supported
	unmarshalErrorV8 := json.Unmarshal([]byte(serviceKeyJSON), &abapServiceKeyV8)
	if abapServiceKeyV8 == (AbapServiceKeyV8{}) {
		if unmarshalErrorV8 != nil {
			log.Entry().Debug(unmarshalErrorV8.Error())
		}
		log.Entry().Debug("Could not parse the service key in the cf cli v8 format.")
	} else {
		log.Entry().Info("Service Key read successfully")
		return abapServiceKeyV8.Credentials, nil
	}

	unmarshalError := json.Unmarshal([]byte(serviceKeyJSON), &abapServiceKey)
	if abapServiceKey == (AbapServiceKey{}) {
		if unmarshalError != nil {
			log.Entry().Debug(unmarshalError.Error())
		}
		log.Entry().Debug("Could not parse the service key in the cf cli v7 format.")
	} else {
		log.Entry().Info("Service Key read successfully")
		return abapServiceKey, nil
	}
	log.SetErrorCategory(log.ErrorInfrastructure)
	return abapServiceKeyV8.Credentials, errors.New("Parsing the service key failed for all supported formats. Service key is empty")
}

/*
GetPollIntervall returns the specified intervall from AbapUtils or a default value of 10 seconds
*/
func (abaputils *AbapUtils) GetPollIntervall() time.Duration {
	if abaputils.Intervall != 0 {
		return abaputils.Intervall
	}
	return 10 * time.Second
}

/*
ReadCOnfigFile reads a file from a specific path and returns the json string as []byte
*/
func ReadConfigFile(path string) (file []byte, err error) {
	filelocation, err := filepath.Glob(path)
	if err != nil {
		return nil, err
	}
	if len(filelocation) == 0 {
		return nil, errors.New("Could not find " + path)
	}
	filename, err := filepath.Abs(filelocation[0])
	if err != nil {
		return nil, err
	}
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var jsonFile []byte
	jsonFile, err = yaml.YAMLToJSON(yamlFile)
	return jsonFile, err
}

// GetHTTPResponse wraps the SendRequest function of piperhttp
func GetHTTPResponse(requestType string, connectionDetails ConnectionDetailsHTTP, body []byte, client piperhttp.Sender) (*http.Response, error) {

	log.Entry().Debugf("Request body: %s", string(body))
	log.Entry().Debugf("Request user: %s", connectionDetails.User)

	header := make(map[string][]string)
	header["Content-Type"] = []string{"application/json"}
	header["Accept"] = []string{"application/json"}
	header["x-csrf-token"] = []string{connectionDetails.XCsrfToken}

	httpResponse, err := client.SendRequest(requestType, connectionDetails.URL, bytes.NewBuffer(body), header, nil)
	return httpResponse, err
}

// HandleHTTPError handles ABAP error messages which can occur when using OData V2 services
//
// The point of this function is to enrich the error received from a HTTP Request (which is passed as a parameter to this function).
// Further error details may be present in the response body of the HTTP response.
// If the response body is parseable, the included details are wrapped around the original error from the HTTP repsponse.
// If this is not possible, the original error is returned.
func HandleHTTPError(resp *http.Response, err error, message string, connectionDetails ConnectionDetailsHTTP) (string, error) {

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

		errorText, errorCode, parsingError = GetErrorDetailsFromResponse(resp)
		if parsingError != nil {
			return "", err
		}
		abapError := errors.New(fmt.Sprintf("%s - %s", errorCode, errorText))
		err = errors.Wrap(abapError, err.Error())

	}
	return errorCode, err
}

// GetErrorDetailsFromResponse parses OData V2 Responses containing ABAP Error messages
func GetErrorDetailsFromResponse(resp *http.Response) (errorString string, errorCode string, err error) {

	// Include the error message of the ABAP Environment system, if available
	var abapErrorResponse AbapErrorODataV2
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
		if (AbapErrorODataV2{}) != abapErrorResponse {
			log.Entry().WithField("ErrorCode", abapErrorResponse.Code).Debug(abapErrorResponse.Message.Value)
			return abapErrorResponse.Message.Value, abapErrorResponse.Code, nil
		}
	}

	return "", "", errors.New("Could not parse the JSON error response")

}

// AddDefaultDashedLine adds 25 dashes
func AddDefaultDashedLine(j int) {
	for i := 1; i <= j; i++ {
		log.Entry().Infof(strings.Repeat("-", 25))
	}
}

// AddDefaultDebugLine adds 25 dashes in debug
func AddDebugDashedLine() {
	log.Entry().Debugf(strings.Repeat("-", 25))
}

/*******************************
 *	Structs for specific steps *
 *******************************/

// AbapEnvironmentPullGitRepoOptions struct for the PullGitRepo piper step
type AbapEnvironmentPullGitRepoOptions struct {
	AbapEnvOptions  AbapEnvironmentOptions
	RepositoryNames []string `json:"repositoryNames,omitempty"`
}

// AbapEnvironmentCheckoutBranchOptions struct for the CheckoutBranch piper step
type AbapEnvironmentCheckoutBranchOptions struct {
	AbapEnvOptions AbapEnvironmentOptions
	RepositoryName string `json:"repositoryName,omitempty"`
}

// AbapEnvironmentRunATCCheckOptions struct for the RunATCCheck piper step
type AbapEnvironmentRunATCCheckOptions struct {
	AbapEnvOptions AbapEnvironmentOptions
	AtcConfig      string `json:"atcConfig,omitempty"`
}

/********************************
 *	Structs for ABAP in general *
 ********************************/

// AbapEnvironmentOptions contains cloud foundry fields and the host parameter for connections to ABAP Environment instances
type AbapEnvironmentOptions struct {
	Username          string `json:"username,omitempty"`
	Password          string `json:"password,omitempty"`
	ByogUsername      string `json:"byogUsername,omitempty"`
	ByogPassword      string `json:"byogPassword,omitempty"`
	ByogAuthMethod    string `json:"byogAuthMethod,omitempty"`
	Host              string `json:"host,omitempty"`
	CfAPIEndpoint     string `json:"cfApiEndpoint,omitempty"`
	CfOrg             string `json:"cfOrg,omitempty"`
	CfSpace           string `json:"cfSpace,omitempty"`
	CfServiceInstance string `json:"cfServiceInstance,omitempty"`
	CfServiceKeyName  string `json:"cfServiceKeyName,omitempty"`
}

// AbapMetadata contains the URI of metadata files
type AbapMetadata struct {
	URI string `json:"uri"`
}

// ConnectionDetailsHTTP contains fields for HTTP connections including the XCSRF token
type ConnectionDetailsHTTP struct {
	Host             string
	User             string   `json:"user"`
	Password         string   `json:"password"`
	URL              string   `json:"url"`
	XCsrfToken       string   `json:"xcsrftoken"`
	CertificateNames []string `json:"-"`
}

// AbapErrorODataV2 contains the error code and the error message for ABAP errors
type AbapErrorODataV2 struct {
	Code    string           `json:"code"`
	Message AbapErrorMessage `json:"message"`
}

// AbapErrorODataV4 contains the error code and the error message for ABAP errors
type AbapErrorODataV4 struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// AbapErrorMessage contains the lanuage and value fields for ABAP errors
type AbapErrorMessage struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

// AbapServiceKeyV8 contains the new format of an ABAP service key

type AbapServiceKeyV8 struct {
	Credentials AbapServiceKey `json:"credentials"`
}

// AbapServiceKey contains information about an ABAP service key
type AbapServiceKey struct {
	SapCloudService    string         `json:"sap.cloud.service"`
	URL                string         `json:"url"`
	SystemID           string         `json:"systemid"`
	Abap               AbapConnection `json:"abap"`
	Binding            AbapBinding    `json:"binding"`
	PreserveHostHeader bool           `json:"preserve_host_header"`
}

// AbapConnection contains information about the ABAP connection for the ABAP endpoint
type AbapConnection struct {
	Username                         string `json:"username"`
	Password                         string `json:"password"`
	CommunicationScenarioID          string `json:"communication_scenario_id"`
	CommunicationArrangementID       string `json:"communication_arrangement_id"`
	CommunicationSystemID            string `json:"communication_system_id"`
	CommunicationInboundUserID       string `json:"communication_inbound_user_id"`
	CommunicationInboundUserAuthMode string `json:"communication_inbound_user_auth_mode"`
}

// AbapBinding contains information about service binding in Cloud Foundry
type AbapBinding struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Env     string `json:"env"`
}

/********************************
 *	Testing with a client mock  *
 ********************************/

// ClientMock contains information about the client mock
type ClientMock struct {
	Token              string
	Body               string
	BodyList           []string
	StatusCode         int
	Error              error
	NilResponse        bool
	ErrorInsteadOfDump bool
	ErrorList          []error
}

// SetOptions sets clientOptions for a client mock
func (c *ClientMock) SetOptions(opts piperhttp.ClientOptions) {}

// SendRequest sets a HTTP response for a client mock
func (c *ClientMock) SendRequest(method, url string, bdy io.Reader, hdr http.Header, cookies []*http.Cookie) (*http.Response, error) {

	if c.NilResponse {
		return nil, c.Error
	}

	var body []byte
	var responseError error
	if c.Body != "" {
		body = []byte(c.Body)
		responseError = c.Error
	} else {
		if c.ErrorInsteadOfDump && len(c.BodyList) == 0 {
			return nil, errors.New("No more bodies in the list")
		}
		bodyString := c.BodyList[len(c.BodyList)-1]
		c.BodyList = c.BodyList[:len(c.BodyList)-1]
		body = []byte(bodyString)
		if len(c.ErrorList) == 0 {
			responseError = c.Error
		} else {
			responseError = c.ErrorList[len(c.ErrorList)-1]
			c.ErrorList = c.ErrorList[:len(c.ErrorList)-1]
		}
	}
	header := http.Header{}
	header.Set("X-Csrf-Token", c.Token)
	return &http.Response{
		StatusCode: c.StatusCode,
		Header:     header,
		Body:       io.NopCloser(bytes.NewReader(body)),
	}, responseError
}

// DownloadFile : Empty file download
func (c *ClientMock) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	return nil
}

// AUtilsMock mock
type AUtilsMock struct {
	ReturnedConnectionDetailsHTTP ConnectionDetailsHTTP
	ReturnedError                 error
}

// GetAbapCommunicationArrangementInfo mock
func (autils *AUtilsMock) GetAbapCommunicationArrangementInfo(options AbapEnvironmentOptions, oDataURL string) (ConnectionDetailsHTTP, error) {
	return autils.ReturnedConnectionDetailsHTTP, autils.ReturnedError
}

// GetPollIntervall mock
func (autils *AUtilsMock) GetPollIntervall() time.Duration {
	return 1 * time.Microsecond
}

// Cleanup to reset AUtilsMock
func (autils *AUtilsMock) Cleanup() {
	autils.ReturnedConnectionDetailsHTTP = ConnectionDetailsHTTP{}
	autils.ReturnedError = nil
}
