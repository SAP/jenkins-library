package cmd

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/pkg/errors"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
)

func TestTriggerPull(t *testing.T) {

	t.Run("Test trigger pull: success case", func(t *testing.T) {

		receivedURI := "example.com/Entity"
		uriExpected := receivedURI + "?$expand=to_Execution_log,to_Transport_log"
		tokenExpected := "myToken"

		client := &clientMock{
			Body:       `{"d" : { "__metadata" : { "uri" : "` + receivedURI + `" } } }`,
			Token:      tokenExpected,
			StatusCode: 200,
		}
		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKey:      "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryNames:   []string{"testRepo1", "testRepo2"},
		}

		con := connectionDetailsHTTP{
			User:     "MY_USER",
			Password: "MY_PW",
			URL:      "https://api.endpoint.com/Entity/",
		}
		entityConnection, err := triggerPull(config.RepositoryNames[0], con, client)
		assert.Nil(t, err)
		assert.Equal(t, uriExpected, entityConnection.URL)
		assert.Equal(t, tokenExpected, entityConnection.XCsrfToken)
	})

	t.Run("Test trigger pull: ABAP Error", func(t *testing.T) {

		errorMessage := "ABAP Error Message"
		errorCode := "ERROR/001"
		HTTPErrorMessage := "HTTP Error Message"
		combinedErrorMessage := "HTTP Error Message: ERROR/001 - ABAP Error Message"

		client := &clientMock{
			Body:       `{"error" : { "code" : "` + errorCode + `", "message" : { "lang" : "en", "value" : "` + errorMessage + `" } } }`,
			Token:      "myToken",
			StatusCode: 400,
			Error:      errors.New(HTTPErrorMessage),
		}
		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKey:      "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryNames:   []string{"testRepo1", "testRepo2"},
		}

		con := connectionDetailsHTTP{
			User:     "MY_USER",
			Password: "MY_PW",
			URL:      "https://api.endpoint.com/Entity/",
		}
		_, err := triggerPull(config.RepositoryNames[0], con, client)
		assert.Equal(t, combinedErrorMessage, err.Error(), "Different error message expected")
	})

}

func TestPollEntity(t *testing.T) {

	t.Run("Test poll entity: success case", func(t *testing.T) {

		client := &clientMock{
			BodyList: []string{
				`{"d" : { "status" : "S" } }`,
				`{"d" : { "status" : "R" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}
		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKey:      "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryNames:   []string{"testRepo1", "testRepo2"},
		}

		con := connectionDetailsHTTP{
			User:       "MY_USER",
			Password:   "MY_PW",
			URL:        "https://api.endpoint.com/Entity/",
			XCsrfToken: "MY_TOKEN",
		}
		status, _ := pollEntity(config.RepositoryNames[0], con, client, 0)
		assert.Equal(t, "S", status)
	})

	t.Run("Test poll entity: error case", func(t *testing.T) {

		client := &clientMock{
			BodyList: []string{
				`{"d" : { "status" : "E" } }`,
				`{"d" : { "status" : "R" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}
		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKey:      "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryNames:   []string{"testRepo1", "testRepo2"},
		}

		con := connectionDetailsHTTP{
			User:       "MY_USER",
			Password:   "MY_PW",
			URL:        "https://api.endpoint.com/Entity/",
			XCsrfToken: "MY_TOKEN",
		}
		status, _ := pollEntity(config.RepositoryNames[0], con, client, 0)
		assert.Equal(t, "E", status)
	})

}

func TestGetAbapCommunicationArrangementInfo(t *testing.T) {

	t.Run("Test cf cli command: success case", func(t *testing.T) {

		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKey:      "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
		}

		execRunner := mock.ExecMockRunner{}

		getAbapCommunicationArrangementInfo(config, &execRunner)
		assert.Equal(t, "cf", execRunner.Calls[0].Exec, "Wrong command")
		assert.Equal(t, []string{"login", "-a", "https://api.endpoint.com", "-u", "testUser", "-p", "testPassword", "-o", "testOrg", "-s", "testSpace"}, execRunner.Calls[0].Params, "Wrong parameters")
	})

	t.Run("Test cf cli command: params missing", func(t *testing.T) {

		config := abapEnvironmentPullGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			Username:          "testUser",
			Password:          "testPassword",
		}

		execRunner := mock.ExecMockRunner{}

		var _, err = getAbapCommunicationArrangementInfo(config, &execRunner)
		assert.Equal(t, "Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510", err.Error(), "Different error message expected")
	})

	t.Run("Test cf cli command: params missing", func(t *testing.T) {

		config := abapEnvironmentPullGitRepoOptions{
			Username: "testUser",
			Password: "testPassword",
		}

		execRunner := mock.ExecMockRunner{}

		var _, err = getAbapCommunicationArrangementInfo(config, &execRunner)
		assert.Equal(t, "Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510", err.Error(), "Different error message expected")
	})

}

func TestTimeConverter(t *testing.T) {
	t.Run("Test example time", func(t *testing.T) {
		inputDate := "/Date(1585576809000+0000)/"
		expectedDate := "2020-03-30 14:00:09 +0000 UTC"
		result := convertTime(inputDate)
		assert.Equal(t, expectedDate, result.String(), "Dates do not match after conversion")
	})
	t.Run("Test Unix time", func(t *testing.T) {
		inputDate := "/Date(0000000000000+0000)/"
		expectedDate := "1970-01-01 00:00:00 +0000 UTC"
		result := convertTime(inputDate)
		assert.Equal(t, expectedDate, result.String(), "Dates do not match after conversion")
	})
	t.Run("Test unexpected format", func(t *testing.T) {
		inputDate := "/Date(0012300000001+0000)/"
		expectedDate := "1970-01-01 00:00:00 +0000 UTC"
		result := convertTime(inputDate)
		assert.Equal(t, expectedDate, result.String(), "Dates do not match after conversion")
	})
}

type clientMock struct {
	Token      string
	Body       string
	BodyList   []string
	StatusCode int
	Error      error
}

func (c *clientMock) SetOptions(opts piperhttp.ClientOptions) {}

func (c *clientMock) SendRequest(method, url string, bdy io.Reader, hdr http.Header, cookies []*http.Cookie) (*http.Response, error) {

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
