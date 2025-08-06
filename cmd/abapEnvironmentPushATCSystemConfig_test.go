package cmd

import (
	"bytes"
	"os"
	"path/filepath"
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

func TestCheckATCSystemConfigurationFile(t *testing.T) {
	t.Parallel()
	t.Run("Check ATC Configuration File - empty", func(t *testing.T) {
		errExpected := "pushing ATC System Configuration failed. Reason: Configured Filelocation is empty (File: atcSystemConfig.json)"
		var parsedConfigurationJsonExpected parsedConfigJsonWithExpand
		var atcSystemConfiguartionJsonFileExpected []byte

		config := abapEnvironmentPushATCSystemConfigOptions{AtcSystemConfigFilePath: "atcSystemConfig.json"}

		parsedConfigurationJson, atcSystemConfiguartionJsonFile, err := checkATCSystemConfigurationFile(&config)
		assert.Equal(t, errExpected, err.Error())
		assert.Equal(t, parsedConfigurationJson, parsedConfigurationJsonExpected)
		assert.Equal(t, atcSystemConfiguartionJsonFile, atcSystemConfiguartionJsonFileExpected)
	})
}

func TestHandleHttpResponse(t *testing.T) {
	t.Parallel()

	t.Run("failiure case: HandleHttpResponse", func(t *testing.T) {
		bodyText := `
--B772E21DAA42B9571C778276B829D6C20
Content-Type: multipart/mixed; boundary=B772E21DAA42B9571C778276B829D6C21
Content-Length:         1973

--B772E21DAA42B9571C778276B829D6C21
Content-Type: application/http
Content-Length: 646
content-transfer-encoding: binary
content-id: 1

HTTP/1.1 200 OK
Content-Type: application/json;odata.metadata=minimal;charset=utf-8
Content-Length: 465
odata-version: 4.0
cache-control: no-cache, no-store, must-revalidate

{"@odata.context":"$metadata#configuration/$entity","@odata.metadataEtag":"W/\"20220211135922\"","root_id":"1","conf_id":"aef8f52b-fe16-1edc-a3fe-27a1e0226c7b","conf_name":"Z_CONFIG_VIA_PIPELINE_STEP","checkvariant":"ABAP_CLOUD_DEVELOPMENT_DEFAULT","pseudo_comment_policy":"SP","last_changed_by":"CC0000000017","last_changed_at":"2022-03-02T11:16:51.336172Z","block_findings":"0","inform_findings":"1","transport_check_policy":"C","check_tasks":true,"check_requests":false,"check_tocs":true,"is_default":false,"is_proxy_variant":false,"SAP__Messages":[]}
--B772E21DAA42B9571C778276B829D6C21
Content-Type: application/http
Content-Length: 428
content-transfer-encoding: binary
content-id: 2

HTTP/1.1 200 OK
Content-Type: application/json;odata.metadata=minimal;charset=utf-8
Content-Length: 247
odata-version: 4.0
cache-control: no-cache, no-store, must-revalidate

{"@odata.context":"$metadata#priority/$entity","@odata.metadataEtag":"W/\"20220211135922\"","root_id":"1","conf_id":"aef8f52b-fe16-1edc-a3fe-27a1e0226c7b","test":"CL_CI_ARS_COMPATIBILITY_CHECK","message_id":"010","default_priority":1,"priority":2}
--B772E21DAA42B9571C778276B829D6C21
Content-Type: application/http
Content-Length: 428
content-transfer-encoding: binary
content-id: 3

HTTP/1.1 4** OK
Content-Type: application/json;odata.metadata=minimal;charset=utf-8
Content-Length: 247
odata-version: 4.0
cache-control: no-cache, no-store, must-revalidate

{"Some Error Messages possible in here!"}
--B772E21DAA42B9571C778276B829D6C21--

--B772E21DAA42B9571C778276B829D6C20--`

		client := &abaputils.ClientMock{
			Body:       bodyText,
			Token:      "myToken",
			StatusCode: 200,
		}
		con := abaputils.ConnectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}

		body := []byte(client.Body)
		resp, err := client.SendRequest("POST", con.URL, bytes.NewBuffer(body), nil, nil)
		if err != nil ||
			resp == nil {
			t.Fatal("Mock should not fail")
		}
		resp.Header.Set("Content-type", "multipart/mixed")
		err = HandleHttpResponse(resp, err, "Unit Test", con)
		// inner error expected
		errExpected := "Outer Response Code: 200 - but at least one Inner response returned StatusCode 4* or 5*. Please check Log for details."
		assert.Equal(t, errExpected, err.Error())
	})

	t.Run("success case: HandleHttpResponse", func(t *testing.T) {
		bodyText := `
--B772E21DAA42B9571C778276B829D6C20
Content-Type: multipart/mixed; boundary=B772E21DAA42B9571C778276B829D6C21
Content-Length:         1973

--B772E21DAA42B9571C778276B829D6C21
Content-Type: application/http
Content-Length: 646
content-transfer-encoding: binary
content-id: 1

HTTP/1.1 200 OK
Content-Type: application/json;odata.metadata=minimal;charset=utf-8
Content-Length: 465
odata-version: 4.0
cache-control: no-cache, no-store, must-revalidate

{"@odata.context":"$metadata#configuration/$entity","@odata.metadataEtag":"W/\"20220211135922\"","root_id":"1","conf_id":"aef8f52b-fe16-1edc-a3fe-27a1e0226c7b","conf_name":"Z_CONFIG_VIA_PIPELINE_STEP","checkvariant":"ABAP_CLOUD_DEVELOPMENT_DEFAULT","pseudo_comment_policy":"SP","last_changed_by":"CC0000000017","last_changed_at":"2022-03-02T11:16:51.336172Z","block_findings":"0","inform_findings":"1","transport_check_policy":"C","check_tasks":true,"check_requests":false,"check_tocs":true,"is_default":false,"is_proxy_variant":false,"SAP__Messages":[]}
--B772E21DAA42B9571C778276B829D6C21
Content-Type: application/http
Content-Length: 428
content-transfer-encoding: binary
content-id: 2

HTTP/1.1 200 OK
Content-Type: application/json;odata.metadata=minimal;charset=utf-8
Content-Length: 247
odata-version: 4.0
cache-control: no-cache, no-store, must-revalidate

{"@odata.context":"$metadata#priority/$entity","@odata.metadataEtag":"W/\"20220211135922\"","root_id":"1","conf_id":"aef8f52b-fe16-1edc-a3fe-27a1e0226c7b","test":"CL_CI_ARS_COMPATIBILITY_CHECK","message_id":"010","default_priority":1,"priority":2}
--B772E21DAA42B9571C778276B829D6C21
Content-Type: application/http
Content-Length: 428
content-transfer-encoding: binary
content-id: 3

HTTP/1.1 200 OK
Content-Type: application/json;odata.metadata=minimal;charset=utf-8
Content-Length: 247
odata-version: 4.0
cache-control: no-cache, no-store, must-revalidate

{"@odata.context":"$metadata#priority/$entity","@odata.metadataEtag":"W/\"20220211135922\"","root_id":"1","conf_id":"aef8f52b-fe16-1edc-a3fe-27a1e0226c7b","test":"CL_CI_ARS_COMPATIBILITY_CHECK","message_id":"011","default_priority":2,"priority":1}
--B772E21DAA42B9571C778276B829D6C21--

--B772E21DAA42B9571C778276B829D6C20--`

		client := &abaputils.ClientMock{
			Body:       bodyText,
			Token:      "myToken",
			StatusCode: 200,
		}
		con := abaputils.ConnectionDetailsHTTP{
			User:     "Test",
			Password: "Test",
			URL:      "https://api.endpoint.com/Entity/",
		}

		body := []byte(client.Body)
		resp, err := client.SendRequest("POST", con.URL, bytes.NewBuffer(body), nil, nil)
		if err != nil ||
			resp == nil {
			t.Fatal("Mock should not fail")
		}
		resp.Header.Set("Content-type", "multipart/mixed")
		err = HandleHttpResponse(resp, err, "Unit Test", con)
		assert.NoError(t, err, "No error expected")
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

{"conf_id":"4711","block_findings":"0","check_requests":false,"check_tasks":true,"check_tocs":true,"checkvariant":"SAP_CLOUD_PLATFORM_ATC_DEFAULT","conf_name":"UNITTEST_PIPERSTEP","inform_findings":"1","pseudo_comment_policy":"MK","transport_check_policy":"C"}

--changeset
Content-Type: application/http
Content-Transfer-Encoding: binary
Content-ID: 2

PATCH priority(root_id='1',conf_id=4711,test='CL_CI_TEST_AMDP_HDB_MIGRATION',message_id='FAIL_ABAP') HTTP/1.1
Content-Type: application/json

{"priority":1}

--changeset--

--request-separator--`

		// no Configuration name supplied
		atcSystemConfigFileString := `{
			"conf_name": "UNITTEST_PIPERSTEP",
			"checkvariant": "SAP_CLOUD_PLATFORM_ATC_DEFAULT",
			"pseudo_comment_policy": "MK",
			"block_findings": "0",
			"inform_findings": "1",
			"transport_check_policy": "C",
			"check_tasks": true,
			"check_requests": false,
			"check_tocs": true,
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

{"conf_id":"4711","block_findings":"0","check_requests":false,"check_tasks":true,"check_tocs":true,"checkvariant":"SAP_CLOUD_PLATFORM_ATC_DEFAULT","conf_name":"UNITTEST_PIPERSTEP","inform_findings":"1","pseudo_comment_policy":"MK","transport_check_policy":"C"}

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

		// no Configuration name supplied
		atcSystemConfigFileString := `{
			"conf_name": "UNITTEST_PIPERSTEP",
			"checkvariant": "SAP_CLOUD_PLATFORM_ATC_DEFAULT",
			"pseudo_comment_policy": "MK",
			"block_findings": "0",
			"inform_findings": "1",
			"transport_check_policy": "C",
			"check_tasks": true,
			"check_requests": false,
			"check_tocs": true,
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
			t.Fatal("Failed to Build ATC System Config Batch Request")
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

{"conf_id":"4711","block_findings":"0","check_requests":false,"check_tasks":true,"check_tocs":true,"checkvariant":"SAP_CLOUD_PLATFORM_ATC_DEFAULT","conf_name":"UNITTEST_PIPERSTEP","inform_findings":"1","pseudo_comment_policy":"MK","transport_check_policy":"C"}

--changeset--

--request-separator--`

		// no Configuration name supplied
		atcSystemConfigFileString := `{
			"conf_name": "UNITTEST_PIPERSTEP",
			"checkvariant": "SAP_CLOUD_PLATFORM_ATC_DEFAULT",
			"pseudo_comment_policy": "MK",
			"block_findings": "0",
			"inform_findings": "1",
			"transport_check_policy": "C",
			"check_tasks": true,
			"check_requests": false,
			"check_tocs": true
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

{"conf_id":"4711","block_findings":"0","check_requests":false,"check_tasks":true,"check_tocs":true,"checkvariant":"SAP_CLOUD_PLATFORM_ATC_DEFAULT","conf_name":"UNITTEST_PIPERSTEP","inform_findings":"1","pseudo_comment_policy":"MK","transport_check_policy":"C"}

--changeset--

--request-separator--`

		// no Configuration name supplied
		atcSystemConfigFileString := `{
			"conf_name": "UNITTEST_PIPERSTEP",
			"checkvariant": "SAP_CLOUD_PLATFORM_ATC_DEFAULT",
			"pseudo_comment_policy": "MK",
			"block_findings": "0",
			"inform_findings": "1",
			"transport_check_policy": "C",
			"check_tasks": true,
			"check_requests": false,
			"check_tocs": true,
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

	t.Run("success case: BuildATCSystemConfigBatch - Config Base only (empty expand _priorities) - Settings Global", func(t *testing.T) {
		batchATCSystemConfigFileExpected := `
--request-separator
Content-Type: multipart/mixed;boundary=changeset

--changeset
Content-Type: application/http
Content-Transfer-Encoding: binary
Content-ID: 1

PATCH configuration(root_id='1',conf_id=4711) HTTP/1.1
Content-Type: application/json

{"conf_id":"4711","block_findings":"0","checkvariant":"SAP_CLOUD_PLATFORM_ATC_DEFAULT","conf_name":"UNITTEST_PIPERSTEP","inform_findings":"1","pseudo_comment_policy":"MK","transport_check_policy":"G"}

--changeset--

--request-separator--`

		// no Configuration name supplied
		atcSystemConfigFileString := `{
			"conf_name": "UNITTEST_PIPERSTEP",
			"checkvariant": "SAP_CLOUD_PLATFORM_ATC_DEFAULT",
			"pseudo_comment_policy": "MK",
			"block_findings": "0",
			"inform_findings": "1",
			"transport_check_policy": "G",
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

		// no Configuration name supplied
		atcSystemConfigFileString := `{
			"conf_name": "UNITTEST_PIPERSTEP",
			"checkvariant": "SAP_CLOUD_PLATFORM_ATC_DEFAULT",
			"pseudo_comment_policy": "MK",
			"block_findings": "0",
			"inform_findings": "1",
			"transport_check_policy": "C",
			"check_tasks": true,
			"check_requests": false,
			"check_tocs": true,
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
		autils := abaputils.AUtilsMock{}
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

		dir := t.TempDir()

		config := abapEnvironmentPushATCSystemConfigOptions{AtcSystemConfigFilePath: filepath.Join(dir, "atcSystemConfig.json")}

		atcSystemConfigFileString := ``

		err := os.WriteFile(config.AtcSystemConfigFilePath, []byte(atcSystemConfigFileString), 0o644)
		if err != nil {
			t.Fatal("Failed to write File: " + config.AtcSystemConfigFilePath)
		}

		expectedErrorMessage := "pushing ATC System Configuration failed. Reason: Configured File is empty (File: " + config.AtcSystemConfigFilePath + ")"

		err = runAbapEnvironmentPushATCSystemConfig(&config, nil, &autils, client)
		assert.Equal(t, expectedErrorMessage, err.Error(), "Different error message expected")
	})

	t.Run("run Step Failure - ATC System Configuration invalid", func(t *testing.T) {
		autils := abaputils.AUtilsMock{}
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

		dir := t.TempDir()

		config := abapEnvironmentPushATCSystemConfigOptions{AtcSystemConfigFilePath: filepath.Join(dir, "atcSystemConfig.json")}

		// no Configuration name supplied
		atcSystemConfigFileString := `{
			"conf_name": "",
			"checkvariant": "SAP_CLOUD_PLATFORM_ATC_DEFAULT",
			"pseudo_comment_policy": "MK",
			"block_findings": "0",
			"inform_findings": "1",
			"is_default": false,
			"is_proxy_variant": false,
			"transport_check_policy": "C",
			"check_tasks": true,
			"check_requests": false,
			"check_tocs": true,
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
		err := os.WriteFile(config.AtcSystemConfigFilePath, []byte(atcSystemConfigFileString), 0o644)
		if err != nil {
			t.Fatal("Failed to write File: " + config.AtcSystemConfigFilePath)
		}

		expectedErrorMessage := "pushing ATC System Configuration failed. Reason: Configured File does not contain required ATC System Configuration attributes (File: " + config.AtcSystemConfigFilePath + ")"

		err = runAbapEnvironmentPushATCSystemConfig(&config, nil, &autils, client)
		assert.Equal(t, expectedErrorMessage, err.Error(), "Different error message expected")
	})

	t.Run("run Step Successful - Push ATC System Configuration", func(t *testing.T) {
		autils := abaputils.AUtilsMock{}
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

		dir := t.TempDir()

		config := abapEnvironmentPushATCSystemConfigOptions{AtcSystemConfigFilePath: filepath.Join(dir, "atcSystemConfig.json")}

		// valid ATC System Configuration File
		atcSystemConfigFileString := `{
			"conf_name": "UNITTEST_PIPERSTEP",
			"checkvariant": "SAP_CLOUD_PLATFORM_ATC_DEFAULT",
			"pseudo_comment_policy": "MK",
			"block_findings": "0",
			"inform_findings": "1",
			"is_default": false,
			"is_proxy_variant": false,
			"transport_check_policy": "C",
			"check_tasks": true,
			"check_requests": false,
			"check_tocs": true,
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
		err := os.WriteFile(config.AtcSystemConfigFilePath, []byte(atcSystemConfigFileString), 0o644)
		if err != nil {
			t.Fatal("Failed to write File: " + config.AtcSystemConfigFilePath)
		}

		err = runAbapEnvironmentPushATCSystemConfig(&config, nil, &autils, client)
		assert.NoError(t, err, "No error expected")
	})

	t.Run("run Step Failure - ATC System Configuration File does not exist", func(t *testing.T) {
		autils := abaputils.AUtilsMock{}
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
