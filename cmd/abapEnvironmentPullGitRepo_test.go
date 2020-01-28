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

		con := connectionDetailsHTTP{
			User:     "MY_USER",
			Password: "MY_PW",
			URL:      "https://api.endpoint.com/Entity/",
		}
		entityConnection, _ := triggerPull(config, con, client)
		assert.Equal(t, uriExpected, entityConnection.URL)
		assert.Equal(t, tokenExpected, entityConnection.XCsrfToken)
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

		con := connectionDetailsHTTP{
			User:       "MY_USER",
			Password:   "MY_PW",
			URL:        "https://api.endpoint.com/Entity/",
			XCsrfToken: "MY_TOKEN",
		}
		status, _ := pollEntity(config, con, client, 0)
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

		con := connectionDetailsHTTP{
			User:       "MY_USER",
			Password:   "MY_PW",
			URL:        "https://api.endpoint.com/Entity/",
			XCsrfToken: "MY_TOKEN",
		}
		status, _ := pollEntity(config, con, client, 0)
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

		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKey:      "testServiceKey",
			User:              "testUser",
			Password:          "testPassword",
		}

		s := shellMockRunner{}

		getAbapCommunicationArrangementInfo(config, &s)
		assert.Equal(t, "/bin/bash", s.shell[0], "Bash shell expected")
		assert.Equal(t, "cf login -a https://api.endpoint.com -u testUser -p testPassword -o testOrg -s testSpace", s.calls[0])
		assert.Equal(t, "cf service-key testInstance testServiceKey | awk '{if(NR>1)print}'", s.calls[1])

	})

	t.Run("Test cf cli command: params missing", func(t *testing.T) {

		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			User:              "testUser",
			Password:          "testPassword",
		}

		s := shellMockRunner{}

		var _, err = getAbapCommunicationArrangementInfo(config, &s)
		assert.Equal(t, "Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510", err.Error(), "Expected error message")
	})

	t.Run("Test cf cli command: params missing", func(t *testing.T) {

		config := abapEnvironmentPullGitRepoOptions{
			User:     "testUser",
			Password: "testPassword",
		}

		s := shellMockRunner{}

		var _, err = getAbapCommunicationArrangementInfo(config, &s)
		assert.Equal(t, "Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510", err.Error(), "Expected error message")
	})

}
