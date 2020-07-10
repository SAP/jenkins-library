package abaputils

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/mock"
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
		connectionDetails, err = GetAbapCommunicationArrangementInfo(options, &command.Command{}, "", false)

		//then
		assert.Equal(t, "", connectionDetails.URL)
		assert.Equal(t, "", connectionDetails.User)
		assert.Equal(t, "", connectionDetails.Password)
		assert.Equal(t, "", connectionDetails.XCsrfToken)

		assert.EqualError(t, err, "Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510")
		assert.Error(t, err)
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
		connectionDetails, err = GetAbapCommunicationArrangementInfo(options, &command.Command{}, "", false)

		//then
		assert.Equal(t, "", connectionDetails.URL)
		assert.Equal(t, "", connectionDetails.User)
		assert.Equal(t, "", connectionDetails.Password)
		assert.Equal(t, "", connectionDetails.XCsrfToken)

		assert.Error(t, err)
	})
	t.Run("CF GetAbapCommunicationArrangementInfo - Success", func(t *testing.T) {

		//given
		m := &mock.ExecMockRunner{}

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

		m.StdoutReturn = map[string]string{"cf service-key testInstance testServiceKeyName": serviceKey}

		//when
		var connectionDetails ConnectionDetailsHTTP
		var err error
		connectionDetails, err = GetAbapCommunicationArrangementInfo(options, m, oDataURL, false)

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
		m := &mock.ExecMockRunner{}

		const testURL = "https://testurl.com"
		const oDataURL = "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull"
		const username = "test_user"
		const password = "test_password"
		const serviceKey = `
		cf comment test \n\n
		{"sap.cloud.service":"com.sap.cloud.abap","url": "` + testURL + `" ,"systemid":"XYZ","abap":{"username":"` + username + `","password":"` + password + `","communication_scenario_id": "SAP_COM_XYZ","communication_arrangement_id": "SK_testing","communication_system_id": "SK_testing","communication_inbound_user_id": "CC0000000000","communication_inbound_user_auth_mode": "2"},"binding":{"env": "cf","version": "0.0.1.1","type": "basic","id": "i6cBiRfZppJdtestKynaTqa32W"},"preserve_host_header": true}`

		options := AbapEnvironmentOptions{
			Host:     testURL,
			Username: username,
			Password: password,
		}

		m.StdoutReturn = map[string]string{"cf service-key testInstance testServiceKeyName": serviceKey}

		//when
		var connectionDetails ConnectionDetailsHTTP
		var err error
		connectionDetails, err = GetAbapCommunicationArrangementInfo(options, m, oDataURL, false)

		//then
		assert.Equal(t, testURL+oDataURL, connectionDetails.URL)
		assert.Equal(t, username, connectionDetails.User)
		assert.Equal(t, password, connectionDetails.Password)
		assert.Equal(t, "", connectionDetails.XCsrfToken)

		assert.NoError(t, err)
	})
	t.Run("HOST GetAbapCommunicationArrangementInfo - Success - w/o https", func(t *testing.T) {

		//given
		m := &mock.ExecMockRunner{}

		const testURL = "testurl.com"
		const oDataURL = "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull"
		const username = "test_user"
		const password = "test_password"
		const serviceKey = `
		cf comment test \n\n
		{"sap.cloud.service":"com.sap.cloud.abap","url": "` + testURL + `" ,"systemid":"H01","abap":{"username":"` + username + `","password":"` + password + `","communication_scenario_id": "SAP_COM_0510","communication_arrangement_id": "SK_I6CBIRFZPPJDKYNATQA32W","communication_system_id": "SK_I6CBIRFZPPJDKYNATQA32W","communication_inbound_user_id": "CC0000000001","communication_inbound_user_auth_mode": "2"},"binding":{"env": "cf","version": "0.0.1.1","type": "basic","id": "i6cBiRfZppJdKynaTqa32W"},"preserve_host_header": true}`

		options := AbapEnvironmentOptions{
			Host:     testURL,
			Username: username,
			Password: password,
		}

		m.StdoutReturn = map[string]string{"cf service-key testInstance testServiceKeyName": serviceKey}

		//when
		var connectionDetails ConnectionDetailsHTTP
		var err error
		connectionDetails, err = GetAbapCommunicationArrangementInfo(options, m, oDataURL, false)

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
		abapKey, err = ReadServiceKeyAbapEnvironment(options, &command.Command{}, true)

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

		assert.Error(t, err)
	})
}
