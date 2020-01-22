package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAbapCommunicationArrangementInfo(t *testing.T) {

	t.Run("Test cf cli command: success case", func(t *testing.T) {

		r := &MockExecRunner{}
		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKey:      "testServiceKey",
			User:              "testUser",
			Password:          "testPassword",
		}

		getAbapCommunicationArrangementInfo(config, r)
		assert.Equal(t, "cf login -a https://api.endpoint.com -u testUser -p testPassword -o testOrg -s testSpace", r.logs[0], "Login command not as expected.")
		assert.Equal(t, "cf service-key testInstance testServiceKey | awk '{if(NR>1)print}'", r.logs[1], "Read Service Key command not as expected.")

	})

	t.Run("Test cf cli command: params missing", func(t *testing.T) {

		r := &MockExecRunner{}
		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			User:              "testUser",
			Password:          "testPassword",
		}

		var _, _, _, err = getAbapCommunicationArrangementInfo(config, r)
		assert.Equal(t, "Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510", err.Error(), "Expected to fail")
	})

	t.Run("Test cf cli command: params missing", func(t *testing.T) {

		r := &MockExecRunner{}
		config := abapEnvironmentPullGitRepoOptions{
			User:     "testUser",
			Password: "testPassword",
		}

		var _, _, _, err = getAbapCommunicationArrangementInfo(config, r)
		assert.Equal(t, "Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510", err.Error(), "Expected to fail")
	})

}

type MockExecRunner struct {
	logs []string
}

func (runner *MockExecRunner) run(script string) ([]byte, error) {

	myString := script
	runner.logs = append(runner.logs, myString)
	return []byte(myString), nil

}
