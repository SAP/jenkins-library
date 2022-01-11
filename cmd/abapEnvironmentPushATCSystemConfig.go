package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"path/filepath"

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
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newAbapEnvironmentPushATCSystemConfigUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.

	err := runAbapEnvironmentPushATCSystemConfig(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentPushATCSystemConfig(config *abapEnvironmentPushATCSystemConfigOptions, telemetryData *telemetry.CustomData, utils abapEnvironmentPushATCSystemConfigUtils) error {

	log.Entry().WithField("func", "Enter: runAbapEnvironmentPushATCSystemConfig").Info("successful")

	exists, err := utils.FileExists(config.AtcSystemConfig)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)

		return fmt.Errorf("failed to check for important file: %w", err)
	}
	if !exists {
		log.Entry().WithField("func", "Leave: runAbapEnvironmentPushATCSystemConfig").Info("No ATC System Configuguration file (%w). Push of ATC System Configuration skipped.")
		return err
	}

	subOptions := convertATCSysOptions(config)

	//Define Client
	var details abaputils.ConnectionDetailsHTTP
	var abapCom abaputils.Communication

	abapCom = GetABAPCom()
	details, err = abapCom.GetAbapCommunicationArrangementInfo(subOptions, "")
	client := piperhttp.Client{}
	cookieJar, _ := cookiejar.New(nil)
	clientOptions := piperhttp.ClientOptions{
		CookieJar: cookieJar,
	}

	//Fetch Xcrsf-Token
	if err == nil {
		client.SetOptions(clientOptions)
		credentialsOptions := piperhttp.ClientOptions{
			Username:  details.User,
			Password:  details.Password,
			CookieJar: cookieJar,
		}
		client.SetOptions(credentialsOptions)
		details.XCsrfToken, err = fetchATCXcsrfToken("GET", details, nil, &client)
	}
	if err == nil {
		err = pushATCSystemConfig(config, details, &client)
	}

	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
	} else {
		log.Entry().WithField("func", "Leave: runAbapEnvironmentPushATCSystemConfig").Info("ATC System Configuration successfully pushed to system")
	}

	return err
}

func pushATCSystemConfig(config *abapEnvironmentPushATCSystemConfigOptions, details abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) error {

	filelocation, err := filepath.Glob(config.AtcSystemConfig)
	//check ATC system configuration json
	var resp *http.Response
	var atcSystemConfiguartionJsonFile []byte

	if err == nil {
		filename, err := filepath.Abs(filelocation[0])
		if err == nil {
			atcSystemConfiguartionJsonFile, err = ioutil.ReadFile(filename)
		}
	}
	if err == nil {
		resp, err = getOdataResponse("POST", details, atcSystemConfiguartionJsonFile, client)
	}
	if err == nil {
		err = parseOdataResponse(resp)
	}
	if err != nil {
		return fmt.Errorf("Pushing ATC System Configuration failed: %w", err)
	}
	return err
}

func parseOdataResponse(resp *http.Response) error {
	//Parse response
	var err error
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err == nil {
		defer resp.Body.Close()
		if len(body) == 0 {
			return fmt.Errorf("Parsing oData result failed: %w", errors.New("Body is empty, can't parse empty body"))
		}
	}
	if err != nil {
		return fmt.Errorf("Parsing oData result failed: %w", err)
	}
	return nil
}

func getOdataResponse(requestType string, details abaputils.ConnectionDetailsHTTP, body []byte, client piperhttp.Sender) (*http.Response, error) {

	header := make(map[string][]string)
	header["x-csrf-token"] = []string{details.XCsrfToken}
	header["Accept"] = []string{"application/json"}
	header["Content-Type"] = []string{"application/json"}
	abapEndpoint := details.URL
	details.URL = abapEndpoint + "/sap/opu/odata4/sap/satc_ci_cf_api/srvd_a2x/sap/satc_ci_cf_sv_api/0001/configuration"
	resp, err := client.SendRequest(requestType, details.URL, bytes.NewBuffer(body), header, nil)
	if err != nil {
		return resp, fmt.Errorf("Sending Request for ATC System Configuration failed: %w", err)
	}
	return resp, nil
}

func fetchATCXcsrfToken(requestType string, details abaputils.ConnectionDetailsHTTP, body []byte, client piperhttp.Sender) (string, error) {

	log.Entry().WithField("ABAP Endpoint: ", details.URL).Debug("Fetching Xcrsf-Token")

	details.URL += "/sap/opu/odata4/sap/satc_ci_cf_api/srvd_a2x/sap/satc_ci_cf_sv_api/0001"
	details.XCsrfToken = "fetch"
	header := make(map[string][]string)
	header["X-Csrf-Token"] = []string{details.XCsrfToken}
	req, err := client.SendRequest(requestType, details.URL, bytes.NewBuffer(body), header, nil)
	if err != nil {
		return "", fmt.Errorf("Fetching Xcsrf-Token failed: %w", err)
	}
	defer req.Body.Close()

	token := req.Header.Get("X-Csrf-Token")
	return token, err
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

func GetABAPCom() abaputils.Communication {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	var autils = abaputils.AbapUtils{
		Exec: &c,
	}

	return &autils
}
