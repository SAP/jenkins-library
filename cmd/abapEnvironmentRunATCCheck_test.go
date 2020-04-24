package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostConfig(t *testing.T) {

	t.Run("Check Host: ABAP Endpoint", func(t *testing.T) {
		config := abapEnvironmentRunATCCheckOptions{
			Username: "testUser",
			Password: "testPassword",
			Host:     "https://api.endpoint.com",
		}
		var con connectionDetailsHTTP
		con, error := checkHost(config, con)
		if error == nil {
			assert.Equal(t, "testUser", con.User)
			assert.Equal(t, "testPassword", con.Password)
			assert.Equal(t, "https://api.endpoint.com", con.URL)
			assert.Equal(t, "", con.XCsrfToken)
		}
	})

	t.Run("No host/ServiceKey configuration", func(t *testing.T) {
		//Testing without CfOrg parameter
		config := abapEnvironmentRunATCCheckOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
		}
		var con connectionDetailsHTTP
		con, err := checkHost(config, con)
		assert.EqualError(t, err, "Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510")
		//Testing without ABAP Host
		config = abapEnvironmentRunATCCheckOptions{
			Username: "testUser",
			Password: "testPassword",
		}
		con, err = checkHost(config, con)
		assert.EqualError(t, err, "Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0510")
	})

	t.Run("Check Host: CF Service Key", func(t *testing.T) {
		config := abapEnvironmentRunATCCheckOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfSpace:           "testSpace",
			CfOrg:             "Test",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
		}
		var con connectionDetailsHTTP
		con, error := checkHost(config, con)
		if error == nil {
			assert.Equal(t, "", con.User)
			assert.Equal(t, "", con.Password)
			assert.Equal(t, "", con.URL)
			assert.Equal(t, "", con.XCsrfToken)
		}
	})

}

func TestATCTrigger(t *testing.T) {
	t.Run("Trigger ATC run test", func(t *testing.T) {
		tokenExpected := "myToken"

		client := &clientMock{
			Body:  `ATC trigger test`,
			Token: tokenExpected,
		}

		con := connectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}
		resp, error := runATC("GET", con, []byte(client.Body), client)
		if error == nil {
			assert.Equal(t, tokenExpected, resp.Header["X-Csrf-Token"][0])
			assert.Equal(t, int64(0), resp.ContentLength)
			assert.Equal(t, []string([]string(nil)), resp.Header["Location"])
		}
	})
}

func TestFetchXcsrfToken(t *testing.T) {
	t.Run("FetchXcsrfToken Test", func(t *testing.T) {
		tokenExpected := "myToken"

		client := &clientMock{
			Body:  `Xcsrf Token test`,
			Token: tokenExpected,
		}

		con := connectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}
		resp, error := fetchXcsrfToken("GET", con, []byte(client.Body), client)
		if error == nil {
			assert.Equal(t, tokenExpected, resp)
		}
	})
}

func TestPollATCRun(t *testing.T) {
	t.Run("ATC run Poll Test", func(t *testing.T) {
		tokenExpected := "myToken"

		client := &clientMock{
			Body:  `ATC Poll test`,
			Token: tokenExpected,
		}

		con := connectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}
		resp, err := pollATCRun(con, []byte(client.Body), client)
		if err != nil {
			assert.Equal(t, "", resp)
			assert.EqualError(t, err, "Could not get any response from ATC poll: Status from ATC run is empty. Either it's not an ABAP system or ATC run hasn't started")

		}
	})
}

func TestGetHTTPResponseATCRun(t *testing.T) {
	t.Run("Get HTTP Response from ATC run Test", func(t *testing.T) {
		client := &clientMock{
			Body: `HTTP response test`,
		}

		con := connectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}
		resp, err := getHTTPResponseATCRun("GET", con, []byte(client.Body), client)
		defer resp.Body.Close()
		if err == nil {
			assert.Equal(t, 200, resp.StatusCode)
			assert.Equal(t, int64(0), resp.ContentLength)
			assert.Equal(t, []string([]string(nil)), resp.Header["X-Crsf-Token"])
		}
	})
}

func TestGetResultATCRun(t *testing.T) {
	t.Run("Get HTTP Response from ATC run Test", func(t *testing.T) {
		client := &clientMock{
			BodyList: []string{
				`ATC result body`,
			},
		}

		con := connectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}
		resp, err := getResultATCRun("GET", con, []byte(client.Body), client)
		defer resp.Body.Close()
		if err == nil {
			assert.Equal(t, 200, resp.StatusCode)
			assert.Equal(t, int64(0), resp.ContentLength)
			assert.Equal(t, []string([]string(nil)), resp.Header["X-Crsf-Token"])
		}
	})
}
