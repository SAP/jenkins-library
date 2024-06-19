//go:build unit
// +build unit

package abaputils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryGetAbapCommunicationInfo(t *testing.T) {
	t.Run("CF GetAbapCommunicationArrangementInfo - Error - parameters missing", func(t *testing.T) {

		//given
		options := AbapEnvironmentOptions{
			//CfAPIEndpoint:     "https://api.endpoint.com",
			CfSpace:           "testSpace",
			CfOrg:             "testOrg",
			CfServiceInstance: "testInstance",
			Username:          "testUser",
			Password:          "testPassword",
			CfServiceKeyName:  "testServiceKeyName",
		}

		//when
		var connectionDetails ConnectionDetailsHTTP
		var err error
		var autils = AbapUtils{
			Exec: &mock.ExecMockRunner{},
		}
		connectionDetails, err = autils.GetAbapCommunicationArrangementInfo(options, "")

		//then
		assert.Equal(t, "", connectionDetails.URL)
		assert.Equal(t, "", connectionDetails.User)
		assert.Equal(t, "", connectionDetails.Password)
		assert.Equal(t, "", connectionDetails.XCsrfToken)

		assert.EqualError(t, err, "Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry API Endpoint, Organization, Space, Service Instance and Service Key")
	})
	t.Run("CF GetAbapCommunicationArrangementInfo - Error - reading service Key", func(t *testing.T) {
		//given
		options := AbapEnvironmentOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfSpace:           "testSpace",
			CfOrg:             "testOrg",
			CfServiceInstance: "testInstance",
			Username:          "testUser",
			Password:          "testPassword",
			CfServiceKeyName:  "testServiceKeyName",
		}

		//when
		var connectionDetails ConnectionDetailsHTTP
		var err error
		var autils = AbapUtils{
			Exec: &mock.ExecMockRunner{},
		}
		connectionDetails, err = autils.GetAbapCommunicationArrangementInfo(options, "")

		//then
		assert.Equal(t, "", connectionDetails.URL)
		assert.Equal(t, "", connectionDetails.User)
		assert.Equal(t, "", connectionDetails.Password)
		assert.Equal(t, "", connectionDetails.XCsrfToken)

		assert.EqualError(t, err, "Read service key failed: Parsing the service key failed for all supported formats. Service key is empty")
	})
	t.Run("CF GetAbapCommunicationArrangementInfo - Success V8", func(t *testing.T) {

		//given

		const testURL = "https://testurl.com"
		const oDataURL = "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull"
		const username = "test_user"
		const password = "test_password"
		const serviceKey = `
		cf comment test \n\n
		{ "credentials": {"sap.cloud.service":"com.sap.cloud.abap","url": "` + testURL + `" ,"systemid":"H01","abap":{"username":"` + username + `","password":"` + password + `","communication_scenario_id": "SAP_COM_0510","communication_arrangement_id": "SK_I6CBIRFZPPJDKYNATQA32W","communication_system_id": "SK_I6CBIRFZPPJDKYNATQA32W","communication_inbound_user_id": "CC0000000001","communication_inbound_user_auth_mode": "2"},"binding":{"env": "cf","version": "0.0.1.1","type": "basic","id": "i6cBiRfZppJdKynaTqa32W"},"preserve_host_header": true} }`

		options := AbapEnvironmentOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfSpace:           "testSpace",
			CfOrg:             "testOrg",
			CfServiceInstance: "testInstance",
			Username:          "testUser",
			Password:          "testPassword",
			CfServiceKeyName:  "testServiceKeyName",
		}

		m := &mock.ExecMockRunner{}
		m.StdoutReturn = map[string]string{"cf service-key testInstance testServiceKeyName": serviceKey}
		var autils = AbapUtils{
			Exec: m,
		}
		//when
		var connectionDetails ConnectionDetailsHTTP
		var err error
		connectionDetails, err = autils.GetAbapCommunicationArrangementInfo(options, oDataURL)

		//then
		assert.Equal(t, testURL+oDataURL, connectionDetails.URL)
		assert.Equal(t, username, connectionDetails.User)
		assert.Equal(t, password, connectionDetails.Password)
		assert.Equal(t, "", connectionDetails.XCsrfToken)

		assert.NoError(t, err)
	})

	t.Run("CF GetAbapCommunicationArrangementInfo - Success V7", func(t *testing.T) {

		//given

		const testURL = "https://testurl.com"
		const oDataURL = "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull"
		const username = "test_user"
		const password = "test_password"
		const serviceKey = `
		cf comment test \n\n
		{"sap.cloud.service":"com.sap.cloud.abap","url": "` + testURL + `" ,"systemid":"H01","abap":{"username":"` + username + `","password":"` + password + `","communication_scenario_id": "SAP_COM_0510","communication_arrangement_id": "SK_I6CBIRFZPPJDKYNATQA32W","communication_system_id": "SK_I6CBIRFZPPJDKYNATQA32W","communication_inbound_user_id": "CC0000000001","communication_inbound_user_auth_mode": "2"},"binding":{"env": "cf","version": "0.0.1.1","type": "basic","id": "i6cBiRfZppJdKynaTqa32W"},"preserve_host_header": true}`

		options := AbapEnvironmentOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfSpace:           "testSpace",
			CfOrg:             "testOrg",
			CfServiceInstance: "testInstance",
			Username:          "testUser",
			Password:          "testPassword",
			CfServiceKeyName:  "testServiceKeyName",
		}

		m := &mock.ExecMockRunner{}
		m.StdoutReturn = map[string]string{"cf service-key testInstance testServiceKeyName": serviceKey}
		var autils = AbapUtils{
			Exec: m,
		}
		//when
		var connectionDetails ConnectionDetailsHTTP
		var err error
		connectionDetails, err = autils.GetAbapCommunicationArrangementInfo(options, oDataURL)

		//then
		assert.Equal(t, testURL+oDataURL, connectionDetails.URL)
		assert.Equal(t, username, connectionDetails.User)
		assert.Equal(t, password, connectionDetails.Password)
		assert.Equal(t, "", connectionDetails.XCsrfToken)

		assert.NoError(t, err)
	})
}

func TestHostGetAbapCommunicationInfo(t *testing.T) {
	t.Run("HOST GetAbapCommunicationArrangementInfo - Success", func(t *testing.T) {

		//given

		const testURL = "https://testurl.com"
		const oDataURL = "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull"
		const username = "test_user"
		const password = "test_password"
		const serviceKey = `
		cf comment test \n\n
		{ "credentials": {"sap.cloud.service":"com.sap.cloud.abap","url": "` + testURL + `" ,"systemid":"XYZ","abap":{"username":"` + username + `","password":"` + password + `","communication_scenario_id": "SAP_COM_XYZ","communication_arrangement_id": "SK_testing","communication_system_id": "SK_testing","communication_inbound_user_id": "CC0000000000","communication_inbound_user_auth_mode": "2"},"binding":{"env": "cf","version": "0.0.1.1","type": "basic","id": "i6cBiRfZppJdtestKynaTqa32W"},"preserve_host_header": true} }`

		options := AbapEnvironmentOptions{
			Host:     testURL,
			Username: username,
			Password: password,
		}

		m := &mock.ExecMockRunner{}
		m.StdoutReturn = map[string]string{"cf service-key testInstance testServiceKeyName": serviceKey}
		var autils = AbapUtils{
			Exec: m,
		}

		//when
		var connectionDetails ConnectionDetailsHTTP
		var err error
		connectionDetails, err = autils.GetAbapCommunicationArrangementInfo(options, oDataURL)

		//then
		assert.Equal(t, testURL+oDataURL, connectionDetails.URL)
		assert.Equal(t, username, connectionDetails.User)
		assert.Equal(t, password, connectionDetails.Password)
		assert.Equal(t, "", connectionDetails.XCsrfToken)

		assert.NoError(t, err)
	})
	t.Run("HOST GetAbapCommunicationArrangementInfo - Success - w/o https", func(t *testing.T) {

		//given

		const testURL = "testurl.com"
		const oDataURL = "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull"
		const username = "test_user"
		const password = "test_password"
		const serviceKey = `
		cf comment test \n\n
		{ "credentials": {"sap.cloud.service":"com.sap.cloud.abap","url": "` + testURL + `" ,"systemid":"H01","abap":{"username":"` + username + `","password":"` + password + `","communication_scenario_id": "SAP_COM_0510","communication_arrangement_id": "SK_I6CBIRFZPPJDKYNATQA32W","communication_system_id": "SK_I6CBIRFZPPJDKYNATQA32W","communication_inbound_user_id": "CC0000000001","communication_inbound_user_auth_mode": "2"},"binding":{"env": "cf","version": "0.0.1.1","type": "basic","id": "i6cBiRfZppJdKynaTqa32W"},"preserve_host_header": true} }`

		options := AbapEnvironmentOptions{
			Host:     testURL,
			Username: username,
			Password: password,
		}

		m := &mock.ExecMockRunner{}
		m.StdoutReturn = map[string]string{"cf service-key testInstance testServiceKeyName": serviceKey}
		var autils = AbapUtils{
			Exec: m,
		}

		//when
		var connectionDetails ConnectionDetailsHTTP
		var err error
		connectionDetails, err = autils.GetAbapCommunicationArrangementInfo(options, oDataURL)

		//then
		assert.Equal(t, "https://"+testURL+oDataURL, connectionDetails.URL)
		assert.Equal(t, username, connectionDetails.User)
		assert.Equal(t, password, connectionDetails.Password)
		assert.Equal(t, "", connectionDetails.XCsrfToken)

		assert.NoError(t, err)
	})
}

func TestReadServiceKeyAbapEnvironment(t *testing.T) {
	t.Run("CF ReadServiceKeyAbapEnvironment - Failed to login to Cloud Foundry", func(t *testing.T) {
		//given .
		options := AbapEnvironmentOptions{
			Username:          "testUser",
			Password:          "testPassword",
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfSpace:           "testSpace",
			CfOrg:             "testOrg",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testKey",
		}

		//when
		var abapKey AbapServiceKey
		var err error
		abapKey, err = ReadServiceKeyAbapEnvironment(options, &mock.ExecMockRunner{})

		//then
		assert.Equal(t, "", abapKey.Abap.Password)
		assert.Equal(t, "", abapKey.Abap.Username)
		assert.Equal(t, "", abapKey.Abap.CommunicationArrangementID)
		assert.Equal(t, "", abapKey.Abap.CommunicationScenarioID)
		assert.Equal(t, "", abapKey.Abap.CommunicationSystemID)

		assert.Equal(t, "", abapKey.Binding.Env)
		assert.Equal(t, "", abapKey.Binding.Type)
		assert.Equal(t, "", abapKey.Binding.ID)
		assert.Equal(t, "", abapKey.Binding.Version)
		assert.Equal(t, "", abapKey.SystemID)
		assert.Equal(t, "", abapKey.URL)

		assert.EqualError(t, err, "Parsing the service key failed for all supported formats. Service key is empty")
	})
}

func TestHandleHTTPError(t *testing.T) {
	t.Run("Test", func(t *testing.T) {

		errorValue := "Received Error"
		abapErrorCode := "abapErrorCode"
		abapErrorMessage := "abapErrorMessage"
		bodyString := `{"error" : { "code" : "` + abapErrorCode + `", "message" : { "lang" : "en", "value" : "` + abapErrorMessage + `" } } }`
		body := []byte(bodyString)

		resp := http.Response{
			Status:     "400 Bad Request",
			StatusCode: 400,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}
		receivedErr := errors.New(errorValue)
		message := "Custom Error Message"

		_, err := HandleHTTPError(&resp, receivedErr, message, ConnectionDetailsHTTP{})
		assert.EqualError(t, err, fmt.Sprintf("%s: %s - %s", receivedErr.Error(), abapErrorCode, abapErrorMessage))
		log.Entry().Info(err.Error())
	})

	t.Run("Non JSON Error", func(t *testing.T) {

		errorValue := "Received Error"
		bodyString := `Error message`
		body := []byte(bodyString)

		resp := http.Response{
			Status:     "400 Bad Request",
			StatusCode: 400,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}
		receivedErr := errors.New(errorValue)
		message := "Custom Error Message"

		_, err := HandleHTTPError(&resp, receivedErr, message, ConnectionDetailsHTTP{})
		assert.EqualError(t, err, fmt.Sprintf("%s", receivedErr.Error()))
		log.Entry().Info(err.Error())
	})

	t.Run("Different JSON Error", func(t *testing.T) {

		errorValue := "Received Error"
		bodyString := `{"abap" : { "key" : "value" } }`
		body := []byte(bodyString)

		resp := http.Response{
			Status:     "400 Bad Request",
			StatusCode: 400,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}
		receivedErr := errors.New(errorValue)
		message := "Custom Error Message"

		_, err := HandleHTTPError(&resp, receivedErr, message, ConnectionDetailsHTTP{})
		assert.EqualError(t, err, fmt.Sprintf("%s", receivedErr.Error()))
		log.Entry().Info(err.Error())
	})

	t.Run("EOF Error", func(t *testing.T) {

		message := "Custom Error Message"
		errorValue := "Received Error EOF"
		receivedErr := errors.New(errorValue)

		_, hook := test.NewNullLogger()
		log.RegisterHook(hook)

		_, err := HandleHTTPError(nil, receivedErr, message, ConnectionDetailsHTTP{})

		assert.EqualError(t, err, fmt.Sprintf("%s", receivedErr.Error()))
		assert.Equal(t, 5, len(hook.Entries), "Expected a different number of entries")
		assert.Equal(t, `A connection could not be established to the ABAP system. The typical root cause is the network configuration (firewall, IP allowlist, etc.)`, hook.AllEntries()[2].Message, "Expected a different message")
		hook.Reset()
	})
}
