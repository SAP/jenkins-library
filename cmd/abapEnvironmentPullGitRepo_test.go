package cmd

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTriggerPull(t *testing.T) {

	t.Run("Test trigger pull: success case", func(t *testing.T) {

		uriExpected := "example.com"
		tokenExpected := "myToken"

		client := &ClientMock{
			Body:  `{"D" : { "__metadata" : { "uri" : "` + uriExpected + `" } } }`,
			Token: tokenExpected,
		}
		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKey:      "testServiceKey",
			User:              "testUser",
			Password:          "testPassword",
		}

		uri, token, _ := triggerPull(config, "https://api.endpoint.com/Entity/", "MY_USER", "MY_PASSWORD", client)
		assert.Equal(t, uriExpected, uri)
		assert.Equal(t, tokenExpected, token)
	})

}

func TestPollEntity(t *testing.T) {

	t.Run("Test poll entity: success case", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{"D" : { "status" : "S" } }`,
				`{"D" : { "status" : "R" } }`,
			},
			Token: "myToken",
		}
		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKey:      "testServiceKey",
			User:              "testUser",
			Password:          "testPassword",
		}

		status, _ := pollEntity(config, "https://api.endpoint.com/Entity/123", "MY_USER", "MY_PASSWORD", "myToken", client, 0)
		assert.Equal(t, "S", status)
	})

	t.Run("Test poll entity: error case", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{"D" : { "status" : "E" } }`,
				`{"D" : { "status" : "R" } }`,
			},
			Token: "myToken",
		}
		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKey:      "testServiceKey",
			User:              "testUser",
			Password:          "testPassword",
		}

		status, _ := pollEntity(config, "https://api.endpoint.com/Entity/123", "MY_USER", "MY_PASSWORD", "myToken", client, 0)
		assert.Equal(t, "E", status)
	})

}

type ClientMock struct {
	Token    string
	Body     string
	BodyList []string
}

func (c *ClientMock) Do(req *http.Request) (*http.Response, error) {

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
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(bytes.NewReader(body)),
	}, nil
}

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
		assert.Equal(t, "Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510", err.Error(), "Expected error message")
	})

	t.Run("Test cf cli command: params missing", func(t *testing.T) {

		r := &MockExecRunner{}
		config := abapEnvironmentPullGitRepoOptions{
			User:     "testUser",
			Password: "testPassword",
		}

		var _, _, _, err = getAbapCommunicationArrangementInfo(config, r)
		assert.Equal(t, "Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510", err.Error(), "Expected error message")
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
