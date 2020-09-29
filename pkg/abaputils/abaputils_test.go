package abaputils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/pkg/errors"
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
			Exec: &command.Command{},
		}
		connectionDetails, err = autils.GetAbapCommunicationArrangementInfo(options, "")

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
		var autils = AbapUtils{
			Exec: &command.Command{},
		}
		connectionDetails, err = autils.GetAbapCommunicationArrangementInfo(options, "")

		//then
		assert.Equal(t, "", connectionDetails.URL)
		assert.Equal(t, "", connectionDetails.User)
		assert.Equal(t, "", connectionDetails.Password)
		assert.Equal(t, "", connectionDetails.XCsrfToken)

		assert.Error(t, err)
	})
	t.Run("CF GetAbapCommunicationArrangementInfo - Success", func(t *testing.T) {

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
		{"sap.cloud.service":"com.sap.cloud.abap","url": "` + testURL + `" ,"systemid":"XYZ","abap":{"username":"` + username + `","password":"` + password + `","communication_scenario_id": "SAP_COM_XYZ","communication_arrangement_id": "SK_testing","communication_system_id": "SK_testing","communication_inbound_user_id": "CC0000000000","communication_inbound_user_auth_mode": "2"},"binding":{"env": "cf","version": "0.0.1.1","type": "basic","id": "i6cBiRfZppJdtestKynaTqa32W"},"preserve_host_header": true}`

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
		{"sap.cloud.service":"com.sap.cloud.abap","url": "` + testURL + `" ,"systemid":"H01","abap":{"username":"` + username + `","password":"` + password + `","communication_scenario_id": "SAP_COM_0510","communication_arrangement_id": "SK_I6CBIRFZPPJDKYNATQA32W","communication_system_id": "SK_I6CBIRFZPPJDKYNATQA32W","communication_inbound_user_id": "CC0000000001","communication_inbound_user_auth_mode": "2"},"binding":{"env": "cf","version": "0.0.1.1","type": "basic","id": "i6cBiRfZppJdKynaTqa32W"},"preserve_host_header": true}`

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
		abapKey, err = ReadServiceKeyAbapEnvironment(options, &command.Command{})

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

func TestTimeConverter(t *testing.T) {
	t.Run("Test example time", func(t *testing.T) {
		inputDate := "/Date(1585576809000+0000)/"
		expectedDate := "2020-03-30 14:00:09 +0000 UTC"
		result := ConvertTime(inputDate)
		assert.Equal(t, expectedDate, result.String(), "Dates do not match after conversion")
	})
	t.Run("Test Unix time", func(t *testing.T) {
		inputDate := "/Date(0000000000000+0000)/"
		expectedDate := "1970-01-01 00:00:00 +0000 UTC"
		result := ConvertTime(inputDate)
		assert.Equal(t, expectedDate, result.String(), "Dates do not match after conversion")
	})
	t.Run("Test unexpected format", func(t *testing.T) {
		inputDate := "/Date(0012300000001+0000)/"
		expectedDate := "1970-01-01 00:00:00 +0000 UTC"
		result := ConvertTime(inputDate)
		assert.Equal(t, expectedDate, result.String(), "Dates do not match after conversion")
	})
}

func TestReadAddonDescriptor(t *testing.T) {
	t.Run("Test: success case", func(t *testing.T) {

		dir, err := ioutil.TempDir("", "test read addon descriptor")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
			_ = os.RemoveAll(dir)
		}()

		body := `---
addonProduct: /DMO/myAddonProduct
addonVersion: 3.1.4
addonUniqueId: myAddonId
customerID: 1234
repositories:
- name: /DMO/REPO_A
  tag: v-1.0.1-build-0001
  branch: branchA
  version: 1.0.1
- name: /DMO/REPO_B
  tag: rel-2.1.1-build-0001
  branch: branchB
  version: 2.1.1
`
		file, _ := os.Create("filename.yaml")
		file.Write([]byte(body))

		addonDescriptor, err := ReadAddonDescriptor("filename.yaml")
		assert.NoError(t, err)
		assert.Equal(t, `/DMO/myAddonProduct`, addonDescriptor.AddonProduct)
		assert.Equal(t, `3.1.4`, addonDescriptor.AddonVersionYAML)
		assert.Equal(t, `myAddonId`, addonDescriptor.AddonUniqueID)
		assert.Equal(t, float64(1234), addonDescriptor.CustomerID)
		assert.Equal(t, ``, addonDescriptor.AddonSpsLevel)
		assert.Equal(t, `/DMO/REPO_A`, addonDescriptor.Repositories[0].Name)
		assert.Equal(t, `/DMO/REPO_B`, addonDescriptor.Repositories[1].Name)
		assert.Equal(t, `v-1.0.1-build-0001`, addonDescriptor.Repositories[0].Tag)
		assert.Equal(t, `rel-2.1.1-build-0001`, addonDescriptor.Repositories[1].Tag)
		assert.Equal(t, `branchA`, addonDescriptor.Repositories[0].Branch)
		assert.Equal(t, `branchB`, addonDescriptor.Repositories[1].Branch)
		assert.Equal(t, `1.0.1`, addonDescriptor.Repositories[0].VersionYAML)
		assert.Equal(t, `2.1.1`, addonDescriptor.Repositories[1].VersionYAML)
		assert.Equal(t, ``, addonDescriptor.Repositories[0].SpLevel)
		assert.Equal(t, ``, addonDescriptor.Repositories[1].SpLevel)

		err = CheckAddonDescriptorForRepositories(addonDescriptor)
		assert.NoError(t, err)
	})
	t.Run("Test: file does not exist", func(t *testing.T) {
		expectedErrorMessage := "AddonDescriptor doesn't contain any repositories"

		addonDescriptor, err := ReadAddonDescriptor("filename.yaml")
		assert.EqualError(t, err, fmt.Sprintf("Could not find %v", "filename.yaml"))
		assert.Equal(t, AddonDescriptor{}, addonDescriptor)

		err = CheckAddonDescriptorForRepositories(addonDescriptor)
		assert.EqualError(t, err, expectedErrorMessage)
	})
	t.Run("Test: empty config - failure case", func(t *testing.T) {
		expectedErrorMessage := "AddonDescriptor doesn't contain any repositories"

		addonDescriptor, err := ReadAddonDescriptor("")

		assert.EqualError(t, err, fmt.Sprintf("Could not find %v", ""))
		assert.Equal(t, AddonDescriptor{}, addonDescriptor)

		err = CheckAddonDescriptorForRepositories(addonDescriptor)
		assert.EqualError(t, err, expectedErrorMessage)
	})
	t.Run("Read empty addon descriptor from wrong config - failure case", func(t *testing.T) {
		expectedErrorMessage := "AddonDescriptor doesn't contain any repositories"
		expectedRepositoryList := AddonDescriptor{Repositories: []Repository{{}, {}}}

		dir, err := ioutil.TempDir("", "test abap utils")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir

		defer func() {
			_ = os.Chdir(oldCWD)
			_ = os.RemoveAll(dir)
		}()

		manifestFileString := `
repositories:
- repo: 'testRepo'
- repo: 'testRepo2'`

		err = ioutil.WriteFile("repositories.yml", []byte(manifestFileString), 0644)

		addonDescriptor, err := ReadAddonDescriptor("repositories.yml")

		assert.Equal(t, expectedRepositoryList, addonDescriptor)
		assert.NoError(t, err)

		err = CheckAddonDescriptorForRepositories(addonDescriptor)
		assert.EqualError(t, err, expectedErrorMessage)
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
			Body:       ioutil.NopCloser(bytes.NewReader(body)),
		}
		receivedErr := errors.New(errorValue)
		message := "Custom Error Message"

		err := HandleHTTPError(&resp, receivedErr, message, ConnectionDetailsHTTP{})
		assert.Error(t, err, "Error was expected")
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
			Body:       ioutil.NopCloser(bytes.NewReader(body)),
		}
		receivedErr := errors.New(errorValue)
		message := "Custom Error Message"

		err := HandleHTTPError(&resp, receivedErr, message, ConnectionDetailsHTTP{})
		assert.Error(t, err, "Error was expected")
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
			Body:       ioutil.NopCloser(bytes.NewReader(body)),
		}
		receivedErr := errors.New(errorValue)
		message := "Custom Error Message"

		err := HandleHTTPError(&resp, receivedErr, message, ConnectionDetailsHTTP{})
		assert.Error(t, err, "Error was expected")
		assert.EqualError(t, err, fmt.Sprintf("%s", receivedErr.Error()))
		log.Entry().Info(err.Error())
	})
}
