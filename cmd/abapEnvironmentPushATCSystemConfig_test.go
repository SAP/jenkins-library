package cmd

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/stretchr/testify/assert"
)

func TestFetchXcsrfTokenFromHead(t *testing.T) {
	t.Parallel()
	t.Run("FetchXcsrfToken Test", func(t *testing.T) {
		tokenExpected := "myToken"

		client := &abaputils.ClientMock{
			Body:  `Xcsrf Token test`,
			Token: tokenExpected,
		}

		con := abaputils.ConnectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}
		token, error := fetchXcsrfTokenFromHead(con, client)
		if error == nil {
			assert.Equal(t, tokenExpected, token)
		}
	})
	t.Run("failure case: fetch token", func(t *testing.T) {
		tokenExpected := ""

		client := &abaputils.ClientMock{
			Body:  `Xcsrf Token test`,
			Token: "",
		}

		con := abaputils.ConnectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}
		token, error := fetchXcsrfTokenFromHead(con, client)
		if error == nil {
			assert.Equal(t, tokenExpected, token)
		}
	})
}

func TestBuildATCSystemConfigBatchRequest(t *testing.T) {
	t.Parallel()

	t.Run("success case: BuildATCSystemConfigBatch - Config Base & 1 Priority", func(t *testing.T) {

		batchATCSystemConfigFileExpected := `
--request-separator
Content-Type: multipart/mixed;boundary=changeset

--changeset
Content-Type: application/http
Content-Transfer-Encoding: binary
Content-ID: 1

PATCH configuration(root_id='1',conf_id=4711) HTTP/1.1
Content-Type: application/json

{"conf_name":"UNITTEST_PIPERSTEP","conf_id":"4711","checkvariant":"SAP_CLOUD_PLATFORM_ATC_DEFAULT","pseudo_comment_policy":"MK","block_findings":"0","inform_findings":"1","is_default":false,"is_proxy_variant":false}

--changeset
Content-Type: application/http
Content-Transfer-Encoding: binary
Content-ID: 2

PATCH priority(root_id='1',conf_id=4711,test='CL_CI_TEST_AMDP_HDB_MIGRATION',message_id='FAIL_ABAP') HTTP/1.1
Content-Type: application/json

{"priority":1}

--changeset--

--request-separator--`

		//no Configuration name supplied
		atcSystemConfigFileString := `{
			"conf_name": "UNITTEST_PIPERSTEP",
			"checkvariant": "SAP_CLOUD_PLATFORM_ATC_DEFAULT",
			"pseudo_comment_policy": "MK",
			"block_findings": "0",
			"inform_findings": "1",
			"is_default": false,
			"is_proxy_variant": false,
			"_priorities": [
				{
					"test": "CL_CI_TEST_AMDP_HDB_MIGRATION",
					"message_id": "FAIL_ABAP",
					"default_priority": 3,
					"priority": 1
				}
			]
		}
		`

		confUUID := "4711"
		batchATCSystemConfigFile, err := buildATCSystemConfigBatchRequest(confUUID, []byte(atcSystemConfigFileString))
		if err != nil {
			t.Fatal("Failed to Build ATC System Config Batch")
		}
		assert.Equal(t, batchATCSystemConfigFileExpected, batchATCSystemConfigFile)

	})

	t.Run("success case: BuildATCSystemConfigBatch - Config Base & 2 Priorities", func(t *testing.T) {

		batchATCSystemConfigFileExpected := `
--request-separator
Content-Type: multipart/mixed;boundary=changeset

--changeset
Content-Type: application/http
Content-Transfer-Encoding: binary
Content-ID: 1

PATCH configuration(root_id='1',conf_id=4711) HTTP/1.1
Content-Type: application/json

{"conf_name":"UNITTEST_PIPERSTEP","conf_id":"4711","checkvariant":"SAP_CLOUD_PLATFORM_ATC_DEFAULT","pseudo_comment_policy":"MK","block_findings":"0","inform_findings":"1","is_default":false,"is_proxy_variant":false}

--changeset
Content-Type: application/http
Content-Transfer-Encoding: binary
Content-ID: 2

PATCH priority(root_id='1',conf_id=4711,test='CL_CI_TEST_AMDP_HDB_MIGRATION',message_id='FAIL_ABAP') HTTP/1.1
Content-Type: application/json

{"priority":1}

--changeset
Content-Type: application/http
Content-Transfer-Encoding: binary
Content-ID: 3

PATCH priority(root_id='1',conf_id=4711,test='CL_CI_TEST_AMDP_HDB_MIGRATION',message_id='FAIL_AMDP') HTTP/1.1
Content-Type: application/json

{"priority":2}

--changeset--

--request-separator--`

		//no Configuration name supplied
		atcSystemConfigFileString := `{
			"conf_name": "UNITTEST_PIPERSTEP",
			"checkvariant": "SAP_CLOUD_PLATFORM_ATC_DEFAULT",
			"pseudo_comment_policy": "MK",
			"block_findings": "0",
			"inform_findings": "1",
			"is_default": false,
			"is_proxy_variant": false,
			"_priorities": [
				{
					"test": "CL_CI_TEST_AMDP_HDB_MIGRATION",
					"message_id": "FAIL_ABAP",
					"default_priority": 3,
					"priority": 1
				},
				{
					"test": "CL_CI_TEST_AMDP_HDB_MIGRATION",
					"message_id": "FAIL_AMDP",
					"priority": 2
				}
			]
		}
		`

		confUUID := "4711"
		batchATCSystemConfigFile, err := buildATCSystemConfigBatchRequest(confUUID, []byte(atcSystemConfigFileString))
		if err != nil {
			t.Fatal("Failed to Build ATC System Config Batch  Request")
		}
		assert.Equal(t, batchATCSystemConfigFileExpected, batchATCSystemConfigFile)

	})

	t.Run("success case: BuildATCSystemConfigBatch - Config Base only (no existing _priorities)", func(t *testing.T) {

		batchATCSystemConfigFileExpected := `
--request-separator
Content-Type: multipart/mixed;boundary=changeset

--changeset
Content-Type: application/http
Content-Transfer-Encoding: binary
Content-ID: 1

PATCH configuration(root_id='1',conf_id=4711) HTTP/1.1
Content-Type: application/json

{"conf_name":"UNITTEST_PIPERSTEP","conf_id":"4711","checkvariant":"SAP_CLOUD_PLATFORM_ATC_DEFAULT","pseudo_comment_policy":"MK","block_findings":"0","inform_findings":"1","is_default":false,"is_proxy_variant":false}

--changeset--

--request-separator--`

		//no Configuration name supplied
		atcSystemConfigFileString := `{
			"conf_name": "UNITTEST_PIPERSTEP",
			"checkvariant": "SAP_CLOUD_PLATFORM_ATC_DEFAULT",
			"pseudo_comment_policy": "MK",
			"block_findings": "0",
			"inform_findings": "1",
			"is_default": false,
			"is_proxy_variant": false
		}
		`

		confUUID := "4711"
		batchATCSystemConfigFile, err := buildATCSystemConfigBatchRequest(confUUID, []byte(atcSystemConfigFileString))
		if err != nil {
			t.Fatal("Failed to Build ATC System Config Batch")
		}
		assert.Equal(t, batchATCSystemConfigFileExpected, batchATCSystemConfigFile)

	})

	t.Run("success case: BuildATCSystemConfigBatch - Config Base only (empty expand _priorities)", func(t *testing.T) {

		batchATCSystemConfigFileExpected := `
--request-separator
Content-Type: multipart/mixed;boundary=changeset

--changeset
Content-Type: application/http
Content-Transfer-Encoding: binary
Content-ID: 1

PATCH configuration(root_id='1',conf_id=4711) HTTP/1.1
Content-Type: application/json

{"conf_name":"UNITTEST_PIPERSTEP","conf_id":"4711","checkvariant":"SAP_CLOUD_PLATFORM_ATC_DEFAULT","pseudo_comment_policy":"MK","block_findings":"0","inform_findings":"1","is_default":false,"is_proxy_variant":false}

--changeset--

--request-separator--`

		//no Configuration name supplied
		atcSystemConfigFileString := `{
			"conf_name": "UNITTEST_PIPERSTEP",
			"checkvariant": "SAP_CLOUD_PLATFORM_ATC_DEFAULT",
			"pseudo_comment_policy": "MK",
			"block_findings": "0",
			"inform_findings": "1",
			"is_default": false,
			"is_proxy_variant": false,
			"_priorities": [
			]
		}
		`

		confUUID := "4711"
		batchATCSystemConfigFile, err := buildATCSystemConfigBatchRequest(confUUID, []byte(atcSystemConfigFileString))
		if err != nil {
			t.Fatal("Failed to Build ATC System Config Batch")
		}
		assert.Equal(t, batchATCSystemConfigFileExpected, batchATCSystemConfigFile)

	})

	t.Run("failure case: BuildATCSystemConfigBatch", func(t *testing.T) {

		batchATCSystemConfigFileExpected := ``

		//no Configuration name supplied
		atcSystemConfigFileString := `{
			"conf_name": "UNITTEST_PIPERSTEP",
			"checkvariant": "SAP_CLOUD_PLATFORM_ATC_DEFAULT",
			"pseudo_comment_policy": "MK",
			"block_findings": "0",
			"inform_findings": "1",
			"is_default": false,
			"is_proxy_variant": false,
			"_priorities": [
				{
					"test": "CL_CI_TEST_AMDP_HDB_MIGRATION",
					"message_id": "FAIL_ABAP",
					"default_priority": 3,
					"priority": 1
				}
			]
		}
		`

		confUUID := "4711"
		batchATCSystemConfigFile, err := buildATCSystemConfigBatchRequest(confUUID, []byte(atcSystemConfigFileString))
		if err != nil {
			t.Fatal("Failed to Build ATC System Config Batch")
		}
		assert.NotEqual(t, batchATCSystemConfigFileExpected, batchATCSystemConfigFile)

	})
}
func TestRunAbapEnvironmentPushATCSystemConfig(t *testing.T) {
	t.Parallel()

	t.Run("run Step Failure - ATC System Configuration File empty", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		receivedURI := "example.com/Entity"
		tokenExpected := "myToken"

		client := &abaputils.ClientMock{
			Body:       `{"d" : { "__metadata" : { "uri" : "` + receivedURI + `" } } }`,
			Token:      tokenExpected,
			StatusCode: 200,
		}

		dir, err := ioutil.TempDir("", "test dir for test file with ATC System Configuration")
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

		config := abapEnvironmentPushATCSystemConfigOptions{AtcSystemConfigFilePath: "atcSystemConfig.json"}

		atcSystemConfigFileString := ``

		err = ioutil.WriteFile(config.AtcSystemConfigFilePath, []byte(atcSystemConfigFileString), 0644)
		if err != nil {
			t.Fatal("Failed to write File: " + config.AtcSystemConfigFilePath)
		}

		expectedErrorMessage := "pushing ATC System Configuration failed. Reason: Configured File is empty (File: " + config.AtcSystemConfigFilePath + ")"

		err = runAbapEnvironmentPushATCSystemConfig(&config, nil, &autils, client)
		assert.Equal(t, expectedErrorMessage, err.Error(), "Different error message expected")
	})

	t.Run("run Step Failure - ATC System Configuration invalid", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		receivedURI := "example.com/Entity"
		tokenExpected := "myToken"

		client := &abaputils.ClientMock{
			Body:       `{"d" : { "__metadata" : { "uri" : "` + receivedURI + `" } } }`,
			Token:      tokenExpected,
			StatusCode: 200,
		}

		dir, err := ioutil.TempDir("", "test dir for test file with ATC System Configuration")
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

		config := abapEnvironmentPushATCSystemConfigOptions{AtcSystemConfigFilePath: "atcSystemConfig.json"}

		//no Configuration name supplied
		atcSystemConfigFileString := `{
			"conf_name": "",
			"checkvariant": "SAP_CLOUD_PLATFORM_ATC_DEFAULT",
			"pseudo_comment_policy": "MK",
			"block_findings": "0",
			"inform_findings": "1",
			"is_default": false,
			"is_proxy_variant": false,
			"_priorities": [
				{
					"test": "CL_CI_TEST_AMDP_HDB_MIGRATION",
					"message_id": "FAIL_ABAP",
					"default_priority": 3,
					"priority": 1
				}
			]
		}
		`
		err = ioutil.WriteFile(config.AtcSystemConfigFilePath, []byte(atcSystemConfigFileString), 0644)
		if err != nil {
			t.Fatal("Failed to write File: " + config.AtcSystemConfigFilePath)
		}

		expectedErrorMessage := "pushing ATC System Configuration failed. Reason: Configured File does not contain required ATC System Configuration attributes (File: " + config.AtcSystemConfigFilePath + ")"

		err = runAbapEnvironmentPushATCSystemConfig(&config, nil, &autils, client)
		assert.Equal(t, expectedErrorMessage, err.Error(), "Different error message expected")
	})

	t.Run("run Step Successful - Push ATC System Configuration", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		receivedURI := "example.com/Entity"
		tokenExpected := "myToken"

		client := &abaputils.ClientMock{
			Body:       `{"d" : { "__metadata" : { "uri" : "` + receivedURI + `" } } }`,
			Token:      tokenExpected,
			StatusCode: 200,
		}

		dir, err := ioutil.TempDir("", "test dir for test file with ATC System Configuration")
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

		config := abapEnvironmentPushATCSystemConfigOptions{AtcSystemConfigFilePath: "atcSystemConfig.json"}

		//valid ATC System Configuration File
		atcSystemConfigFileString := `{
			"conf_name": "UNITTEST_PIPERSTEP",
			"checkvariant": "SAP_CLOUD_PLATFORM_ATC_DEFAULT",
			"pseudo_comment_policy": "MK",
			"block_findings": "0",
			"inform_findings": "1",
			"is_default": false,
			"is_proxy_variant": false,
			"_priorities": [
				{
					"test": "CL_SOMECLASS",
					"message_id": "SOME_MESSAGE_ID",
					"default_priority": 3,
					"priority": 1
				}
			]
		}
		`
		err = ioutil.WriteFile(config.AtcSystemConfigFilePath, []byte(atcSystemConfigFileString), 0644)
		if err != nil {
			t.Fatal("Failed to write File: " + config.AtcSystemConfigFilePath)
		}

		err = runAbapEnvironmentPushATCSystemConfig(&config, nil, &autils, client)
		assert.NoError(t, err, "No error expected")
	})

	t.Run("run Step Failure - ATC System Configuration File does not exist", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		receivedURI := "example.com/Entity"
		tokenExpected := "myToken"

		client := &abaputils.ClientMock{
			Body:       `{"d" : { "__metadata" : { "uri" : "` + receivedURI + `" } } }`,
			Token:      tokenExpected,
			StatusCode: 200,
		}

		config := abapEnvironmentPushATCSystemConfigOptions{AtcSystemConfigFilePath: "test.json"}

		expectedErrorMessage := "pushing ATC System Configuration failed. Reason: Configured Filelocation is empty (File: " + config.AtcSystemConfigFilePath + ")"

		err := runAbapEnvironmentPushATCSystemConfig(&config, nil, &autils, client)
		assert.Equal(t, expectedErrorMessage, err.Error(), "Different error message expected")
	})
}
