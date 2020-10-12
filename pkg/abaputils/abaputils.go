package abaputils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
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
			return connectionDetails, errors.Wrap(err, "Schema validation for host parameter failed. Check for https.")
		}
		var hostOdataURL = options.Host + oDataURL
		if match {
			connectionDetails.URL = hostOdataURL
		} else {
			connectionDetails.URL = "https://" + hostOdataURL
		}
		connectionDetails.User = options.Username
		connectionDetails.Password = options.Password
	} else {
		if options.CfAPIEndpoint == "" || options.CfOrg == "" || options.CfSpace == "" || options.CfServiceInstance == "" || options.CfServiceKeyName == "" {
			var err = errors.New("Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510")
			return connectionDetails, err
		}
		// Url, User and Password should be read from a cf service key
		var abapServiceKey, error = ReadServiceKeyAbapEnvironment(options, c)
		if error != nil {
			return connectionDetails, errors.Wrap(error, "Read service key failed")
		}
		connectionDetails.URL = abapServiceKey.URL + oDataURL
		connectionDetails.User = abapServiceKey.Abap.Username
		connectionDetails.Password = abapServiceKey.Abap.Password
	}
	return connectionDetails, error
}

// ReadServiceKeyAbapEnvironment from Cloud Foundry and returns it. Depending on user/developer requirements if he wants to perform further Cloud Foundry actions
func ReadServiceKeyAbapEnvironment(options AbapEnvironmentOptions, c command.ExecRunner) (AbapServiceKey, error) {

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
		return abapServiceKey, err
	}

	// parse
	json.Unmarshal([]byte(serviceKeyJSON), &abapServiceKey)
	if abapServiceKey == (AbapServiceKey{}) {
		return abapServiceKey, errors.New("Parsing the service key failed")
	}

	log.Entry().Info("Service Key read successfully")
	return abapServiceKey, nil
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

// GetHTTPResponse wraps the SendRequest function of piperhttp
func GetHTTPResponse(requestType string, connectionDetails ConnectionDetailsHTTP, body []byte, client piperhttp.Sender) (*http.Response, error) {

	header := make(map[string][]string)
	header["Content-Type"] = []string{"application/json"}
	header["Accept"] = []string{"application/json"}
	header["x-csrf-token"] = []string{connectionDetails.XCsrfToken}

	httpResponse, err := client.SendRequest(requestType, connectionDetails.URL, bytes.NewBuffer(body), header, nil)
	return httpResponse, err
}

// HandleHTTPError handles ABAP error messages which can occur when using OData services
//
// The point of this function is to enrich the error received from a HTTP Request (which is passed as a parameter to this function).
// Further error details may be present in the response body of the HTTP response.
// If the response body is parseable, the included details are wrapped around the original error from the HTTP repsponse.
// If this is not possible, the original error is returned.
func HandleHTTPError(resp *http.Response, err error, message string, connectionDetails ConnectionDetailsHTTP) error {
	if resp == nil {
		// Response is nil in case of a timeout
		log.Entry().WithError(err).WithField("ABAP Endpoint", connectionDetails.URL).Error("Request failed")
	} else {

		defer resp.Body.Close()

		log.Entry().WithField("StatusCode", resp.Status).Error(message)

		errorDetails, parsingError := getErrorDetailsFromResponse(resp)
		if parsingError != nil {
			return err
		}
		abapError := errors.New(errorDetails)
		err = errors.Wrap(abapError, err.Error())

	}
	return err
}

func getErrorDetailsFromResponse(resp *http.Response) (errorString string, err error) {

	// Include the error message of the ABAP Environment system, if available
	var abapErrorResponse AbapError
	bodyText, readError := ioutil.ReadAll(resp.Body)
	if readError != nil {
		return errorString, readError
	}
	var abapResp map[string]*json.RawMessage
	errUnmarshal := json.Unmarshal(bodyText, &abapResp)
	if errUnmarshal != nil {
		return errorString, errUnmarshal
	}
	if _, ok := abapResp["error"]; ok {
		json.Unmarshal(*abapResp["error"], &abapErrorResponse)
		if (AbapError{}) != abapErrorResponse {
			log.Entry().WithField("ErrorCode", abapErrorResponse.Code).Error(abapErrorResponse.Message.Value)
			errorString = fmt.Sprintf("%s - %s", abapErrorResponse.Code, abapErrorResponse.Message.Value)
			return errorString, nil
		}
	}

	return errorString, errors.New("Could not parse the JSON error response")

}

// ConvertTime formats an ABAP timestamp string from format /Date(1585576807000+0000)/ into a UNIX timestamp and returns it
func ConvertTime(logTimeStamp string) time.Time {
	seconds := strings.TrimPrefix(strings.TrimSuffix(logTimeStamp, "000+0000)/"), "/Date(")
	n, error := strconv.ParseInt(seconds, 10, 64)
	if error != nil {
		return time.Unix(0, 0).UTC()
	}
	t := time.Unix(n, 0).UTC()
	return t
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

//AbapEnvironmentOptions contains cloud foundry fields and the host parameter for connections to ABAP Environment instances
type AbapEnvironmentOptions struct {
	Username          string `json:"username,omitempty"`
	Password          string `json:"password,omitempty"`
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
	User       string `json:"user"`
	Password   string `json:"password"`
	URL        string `json:"url"`
	XCsrfToken string `json:"xcsrftoken"`
}

// AbapError contains the error code and the error message for ABAP errors
type AbapError struct {
	Code    string           `json:"code"`
	Message AbapErrorMessage `json:"message"`
}

// AbapErrorMessage contains the lanuage and value fields for ABAP errors
type AbapErrorMessage struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
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
	Token      string
	Body       string
	BodyList   []string
	StatusCode int
	Error      error
}

// SetOptions sets clientOptions for a client mock
func (c *ClientMock) SetOptions(opts piperhttp.ClientOptions) {}

// SendRequest sets a HTTP response for a client mock
func (c *ClientMock) SendRequest(method, url string, bdy io.Reader, hdr http.Header, cookies []*http.Cookie) (*http.Response, error) {

	var body []byte
	if c.Body != "" {
		body = []byte(c.Body)
	} else {
		bodyString := c.BodyList[len(c.BodyList)-1]
		c.BodyList = c.BodyList[:len(c.BodyList)-1]
		body = []byte(bodyString)
	}
	header := http.Header{}
	header.Set("X-Csrf-Token", c.Token)
	return &http.Response{
		StatusCode: c.StatusCode,
		Header:     header,
		Body:       ioutil.NopCloser(bytes.NewReader(body)),
	}, c.Error
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
