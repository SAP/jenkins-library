package cmd

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/stretchr/testify/assert"
)

var executionLogStringCheckout string

func init() {
	executionLog := abaputils.LogProtocolResults{
		Results: []abaputils.LogProtocol{
			{
				ProtocolLine:  1,
				OverviewIndex: 1,
				Type:          "LogEntry",
				Description:   "S",
				Timestamp:     "/Date(1644332299000+0000)/",
			},
		},
	}
	executionLogResponse, _ := json.Marshal(executionLog)
	executionLogStringCheckout = string(executionLogResponse)
}

func TestCheckoutBranchStep(t *testing.T) {
	t.Run("Run Step Successful - repositoryName and branchName config", func(t *testing.T) {

		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		config := abapEnvironmentCheckoutBranchOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryName:    "testRepo1",
			BranchName:        "testBranch",
			LogOutput:         "STANDARD",
		}

		logResultSuccess := `{"d": { "sc_name": "/DMO/SWC", "status": "S", "to_Log_Overview": { "results": [ { "log_index": 1, "log_name": "Main Import", "type_of_found_issues": "Success", "timestamp": "/Date(1644332299000+0000)/", "to_Log_Protocol": { "results": [ { "log_index": 1, "index_no": "1", "log_name": "", "type": "Info", "descr": "Main import", "timestamp": null, "criticality": 0 } ] } } ] } } }`
		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : [] }`,
				`{"d" : ` + executionLogStringCheckout + `}`,
				logResultSuccess,
				`{"d" : { "status" : "S" } }`,
				`{"d" : { "status" : "S" } }`,
				`{"d" : { "status" : "S" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		var reports []piperutils.Path
		logOutputManager := abaputils.LogOutputManager{
			LogOutput:    config.LogOutput,
			PiperStep:    "checkoutBranch",
			FileNameStep: "checkoutBranch",
			StepReports:  reports,
		}

		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond, Force0510: true}
		err := runAbapEnvironmentCheckoutBranch(&config, &autils, apiManager, &logOutputManager)
		assert.NoError(t, err, "Did not expect error")
	})
	t.Run("Run Step Failure - empty config", func(t *testing.T) {
		expectedErrorMessage := "Configuration is not consistent: You have not specified any repository or branch configuration to be checked out in the ABAP Environment System. Please make sure that you specified the repositories with their branches that should be checked out either in a dedicated file or via the parameters 'repositoryName' and 'branchName'. For more information please read the user documentation"

		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		config := abapEnvironmentCheckoutBranchOptions{}

		logResultError := `{"d": { "sc_name": "/DMO/SWC", "status": "S", "to_Log_Overview": { "results": [ { "log_index": 1, "log_name": "Main Import", "type_of_found_issues": "Error", "timestamp": "/Date(1644332299000+0000)/", "to_Log_Protocol": { "results": [ { "log_index": 1, "index_no": "1", "log_name": "", "type": "Info", "descr": "Main import", "timestamp": null, "criticality": 0 } ] } } ] } } }`
		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : ` + executionLogStringCheckout + `}`,
				logResultError,
				`{"d" : { "status" : "E" } }`,
				`{"d" : { "status" : "E" } }`,
				`{"d" : { "status" : "E" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		var reports []piperutils.Path
		logOutputManager := abaputils.LogOutputManager{
			LogOutput:    "STANDARD",
			PiperStep:    "checkoutBranch",
			FileNameStep: "checkoutBranch",
			StepReports:  reports,
		}

		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond, Force0510: true}
		err := runAbapEnvironmentCheckoutBranch(&config, &autils, apiManager, &logOutputManager)
		assert.EqualError(t, err, expectedErrorMessage)
	})
	t.Run("Run Step Failure - wrong status", func(t *testing.T) {
		expectedErrorMessage := "Something failed during the checkout: Checkout failed: Checkout of branch testBranch failed on the ABAP System"

		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		config := abapEnvironmentCheckoutBranchOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryName:    "testRepo1",
			BranchName:        "testBranch",
			LogOutput:         "STANDARD",
		}

		logResultError := `{"d": { "sc_name": "/DMO/SWC", "status": "S", "to_Log_Overview": { "results": [ { "log_index": 1, "log_name": "Main Import", "type_of_found_issues": "Error", "timestamp": "/Date(1644332299000+0000)/", "to_Log_Protocol": { "results": [ { "log_index": 1, "index_no": "1", "log_name": "", "type": "Info", "descr": "Main import", "timestamp": null, "criticality": 0 } ] } } ] } } }`
		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : ` + executionLogStringCheckout + `}`,
				logResultError,
				`{"d" : { "status" : "E" } }`,
				`{"d" : { "status" : "E" } }`,
				`{"d" : { "status" : "E" } }`,
				`{"d" : { "status" : "E" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		var reports []piperutils.Path
		logOutputManager := abaputils.LogOutputManager{
			LogOutput:    config.LogOutput,
			PiperStep:    "checkoutBranch",
			FileNameStep: "checkoutBranch",
			StepReports:  reports,
		}

		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond, Force0510: true}
		err := runAbapEnvironmentCheckoutBranch(&config, &autils, apiManager, &logOutputManager)
		assert.EqualError(t, err, expectedErrorMessage)
	})
	t.Run("Success case: checkout Branches from file config", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		receivedURI := "example.com/Branches"
		client := &abaputils.ClientMock{
			Body:       `{"d" : { "__metadata" : { "uri" : "` + receivedURI + `" } } }`,
			Token:      "myToken",
			StatusCode: 200,
		}

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir

		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		manifestFileString := `
repositories:
- name: 'testRepo'
  branch: 'testBranch'
- name: 'testRepo2'
  branch: 'testBranch2'
- name: 'testRepo3'
  branch: 'testBranch3'`

		err := os.WriteFile("repositoriesTest.yml", []byte(manifestFileString), 0644)

		config := abapEnvironmentCheckoutBranchOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			Repositories:      "repositoriesTest.yml",
			LogOutput:         "STANDARD",
		}

		var reports []piperutils.Path
		logOutputManager := abaputils.LogOutputManager{
			LogOutput:    config.LogOutput,
			PiperStep:    "checkoutBranch",
			FileNameStep: "checkoutBranch",
			StepReports:  reports,
		}

		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond, Force0510: true}
		err = runAbapEnvironmentCheckoutBranch(&config, &autils, apiManager, &logOutputManager)
		assert.NoError(t, err)
	})
	t.Run("Failure case: checkout Branches from empty file config", func(t *testing.T) {
		expectedErrorMessage := "Could not read repositories: Error in config file repositoriesTest.yml, AddonDescriptor doesn't contain any repositories"

		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		receivedURI := "example.com/Branches"
		client := &abaputils.ClientMock{
			Body:       `{"d" : { "__metadata" : { "uri" : "` + receivedURI + `" } } }`,
			Token:      "myToken",
			StatusCode: 200,
		}

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		manifestFileString := ``

		manifestFileStringBody := []byte(manifestFileString)
		err := os.WriteFile("repositoriesTest.yml", manifestFileStringBody, 0644)

		config := abapEnvironmentCheckoutBranchOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			Repositories:      "repositoriesTest.yml",
			LogOutput:         "STANDARD",
		}

		var reports []piperutils.Path
		logOutputManager := abaputils.LogOutputManager{
			LogOutput:    config.LogOutput,
			PiperStep:    "checkoutBranch",
			FileNameStep: "checkoutBranch",
			StepReports:  reports,
		}

		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond, Force0510: true}
		err = runAbapEnvironmentCheckoutBranch(&config, &autils, apiManager, &logOutputManager)
		assert.EqualError(t, err, expectedErrorMessage)
	})
	t.Run("Failure case: checkout Branches from wrong file config", func(t *testing.T) {
		expectedErrorMessage := "Could not read repositories: Could not unmarshal repositoriesTest.yml"

		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		pollIntervall := abaputils.AUtilsMock{}
		defer pollIntervall.Cleanup()

		receivedURI := "example.com/Branches"
		client := &abaputils.ClientMock{
			Body:       `{"d" : { "__metadata" : { "uri" : "` + receivedURI + `" } } }`,
			Token:      "myToken",
			StatusCode: 200,
		}

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		manifestFileString := `
- repo: 'testRepo'
- repo: 'testRepo2'`

		manifestFileStringBody := []byte(manifestFileString)
		err := os.WriteFile("repositoriesTest.yml", manifestFileStringBody, 0644)

		config := abapEnvironmentCheckoutBranchOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			Repositories:      "repositoriesTest.yml",
			LogOutput:         "STANDARD",
		}

		var reports []piperutils.Path
		logOutputManager := abaputils.LogOutputManager{
			LogOutput:    config.LogOutput,
			PiperStep:    "checkoutBranch",
			FileNameStep: "checkoutBranch",
			StepReports:  reports,
		}

		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond, Force0510: true}
		err = runAbapEnvironmentCheckoutBranch(&config, &autils, apiManager, &logOutputManager)
		assert.EqualError(t, err, expectedErrorMessage)
	})
}

func TestCheckoutConfigChecker(t *testing.T) {
	t.Run("Success case: check config", func(t *testing.T) {
		config := abapEnvironmentCheckoutBranchOptions{
			RepositoryName: "testRepo1",
			BranchName:     "feature-unit-test",
		}
		err := checkCheckoutBranchRepositoryConfiguration(config)
		assert.NoError(t, err)
	})
	t.Run("Success case: check file config", func(t *testing.T) {
		config := abapEnvironmentCheckoutBranchOptions{
			Repositories: "test.file",
			BranchName:   "feature-unit-test",
		}
		err := checkCheckoutBranchRepositoryConfiguration(config)
		assert.NoError(t, err)
	})
	t.Run("Failure case: check empty config", func(t *testing.T) {
		expectedErrorMessage := "You have not specified any repository or branch configuration to be checked out in the ABAP Environment System. Please make sure that you specified the repositories with their branches that should be checked out either in a dedicated file or via the parameters 'repositoryName' and 'branchName'. For more information please read the user documentation"

		config := abapEnvironmentCheckoutBranchOptions{}
		err := checkCheckoutBranchRepositoryConfiguration(config)
		assert.Equal(t, expectedErrorMessage, err.Error(), "Different error message expected")
	})
}
