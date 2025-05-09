//go:build unit

package cmd

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

var executionLogStringCreateTag string
var logResultSuccess string

func init() {
	logResultSuccess = `{"d": { "sc_name": "/DMO/SWC", "status": "S", "to_Log_Overview": { "results": [ { "log_index": 1, "log_name": "Main Import", "type_of_found_issues": "Success", "timestamp": "/Date(1644332299000+0000)/", "to_Log_Protocol": { "results": [ { "log_index": 1, "index_no": "1", "log_name": "", "type": "Info", "descr": "Main import", "timestamp": null, "criticality": 0 } ] } } ] } } }`
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
		Count: "1",
	}
	executionLogResponse, _ := json.Marshal(executionLog)
	executionLogStringCreateTag = string(executionLogResponse)

}

func TestRunAbapEnvironmentCreateTag(t *testing.T) {

	t.Run("happy path", func(t *testing.T) {

		var autils = &abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		body := `---
addonVersion: "1.2.3"
addonProduct: "/DMO/PRODUCT"
repositories:
  - name: /DMO/SWC
    branch: main
    commitID: 1234abcd
    version: "4.5.6"
`
		file, _ := os.Create("repo.yml")
		_, err := file.Write([]byte(body))
		assert.NoError(t, err)
		config := &abapEnvironmentCreateTagOptions{
			Username:                            "dummy",
			Password:                            "dummy",
			Host:                                "https://test.com",
			Repositories:                        "repo.yml",
			TagName:                             "tag",
			TagDescription:                      "desc",
			GenerateTagForAddonProductVersion:   true,
			GenerateTagForAddonComponentVersion: true,
		}
		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : ` + executionLogStringCreateTag + `}`,
				logResultSuccess,
				`{"d" : { "Status" : "S" } }`,
				`{"d" : { "uuid" : "abc" } }`,
				`{"d" : ` + executionLogStringCreateTag + `}`,
				logResultSuccess,
				`{"d" : { "Status" : "S" } }`,
				`{"d" : { "uuid" : "abc" } }`,
				`{"d" : ` + executionLogStringCreateTag + `}`,
				logResultSuccess,
				`{"d" : { "Status" : "S" } }`,
				`{"d" : { "uuid" : "abc" } }`,
				`{"d" : { "empty" : "body" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		_, hook := test.NewNullLogger()
		log.RegisterHook(hook)

		apiManager := &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond, Force0510: true}
		err = runAbapEnvironmentCreateTag(config, autils, apiManager)

		assert.Error(t, err, "Did expect error")
		assert.Equal(t, 18, len(hook.Entries), "Expected a different number of entries")
		assert.Equal(t, `Created tag v4.5.6 for repository /DMO/SWC with commitID 1234abcd`, hook.AllEntries()[6].Message, "Expected a different message")
		assert.Equal(t, `NOT created: Tag -DMO-PRODUCT-1.2.3 for repository /DMO/SWC with commitID 1234abcd`, hook.AllEntries()[11].Message, "Expected a different message")
		assert.Equal(t, `NOT created: Tag tag for repository /DMO/SWC with commitID 1234abcd`, hook.AllEntries()[16].Message, "Expected a different message")
		hook.Reset()
	})

	t.Run("backend error", func(t *testing.T) {

		var autils = &abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		body := `---
addonVersion: "1.2.3"
addonProduct: "/DMO/PRODUCT"
repositories:
  - name: /DMO/SWC
    branch: main
    commitID: 1234abcd
    version: "4.5.6"
`
		file, _ := os.Create("repo.yml")
		_, err := file.Write([]byte(body))
		assert.NoError(t, err)
		config := &abapEnvironmentCreateTagOptions{
			Username:                            "dummy",
			Password:                            "dummy",
			Host:                                "https://test.com",
			Repositories:                        "repo.yml",
			TagName:                             "tag",
			TagDescription:                      "desc",
			GenerateTagForAddonProductVersion:   true,
			GenerateTagForAddonComponentVersion: true,
		}
		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : ` + executionLogStringCreateTag + `}`,
				logResultSuccess,
				`{"d" : { "Status" : "E" } }`,
				`{"d" : { "uuid" : "abc" } }`,
				`{"d" : { "empty" : "body" } }`,
				`{"d" : ` + executionLogStringCreateTag + `}`,
				logResultSuccess,
				`{"d" : { "Status" : "E" } }`,
				`{"d" : { "uuid" : "abc" } }`,
				`{"d" : { "empty" : "body" } }`,
				`{"d" : ` + executionLogStringCreateTag + `}`,
				logResultSuccess,
				`{"d" : { "Status" : "E" } }`,
				`{"d" : { "uuid" : "abc" } }`,
				`{"d" : { "empty" : "body" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		_, hook := test.NewNullLogger()
		log.RegisterHook(hook)

		apiManager := &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond, Force0510: true}
		err = runAbapEnvironmentCreateTag(config, autils, apiManager)

		assert.NoError(t, err, "Did expect error")
		assert.Equal(t, 21, len(hook.Entries), "Expected a different number of entries")
		assert.Equal(t, `Created tag v4.5.6 for repository /DMO/SWC with commitID 1234abcd`, hook.AllEntries()[6].Message, "Expected a different message")
		assert.Equal(t, `Created tag -DMO-PRODUCT-1.2.3 for repository /DMO/SWC with commitID 1234abcd`, hook.AllEntries()[13].Message, "Expected a different message")
		assert.Equal(t, `Created tag tag for repository /DMO/SWC with commitID 1234abcd`, hook.AllEntries()[20].Message, "Expected a different message")
		hook.Reset()

	})

}

func TestRunAbapEnvironmentCreateTagConfigurations(t *testing.T) {

	t.Run("no repo.yml", func(t *testing.T) {

		var autils = &abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		config := &abapEnvironmentCreateTagOptions{
			Username:                            "dummy",
			Password:                            "dummy",
			Host:                                "https://test.com",
			RepositoryName:                      "/DMO/SWC",
			CommitID:                            "1234abcd",
			TagName:                             "tag",
			TagDescription:                      "desc",
			GenerateTagForAddonProductVersion:   true,
			GenerateTagForAddonComponentVersion: true,
		}
		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : ` + executionLogStringCreateTag + `}`,
				logResultSuccess,
				`{"d" : { "Status" : "S" } }`,
				`{"d" : { "uuid" : "abc" } }`,
				`{"d" : { "empty" : "body" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		_, hook := test.NewNullLogger()
		log.RegisterHook(hook)

		apiManager := &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond, Force0510: true}
		err := runAbapEnvironmentCreateTag(config, autils, apiManager)

		assert.NoError(t, err, "Did not expect error")
		assert.Equal(t, 7, len(hook.Entries), "Expected a different number of entries")
		assert.Equal(t, `Created tag tag for repository /DMO/SWC with commitID 1234abcd`, hook.AllEntries()[6].Message, "Expected a different message")
		hook.Reset()
	})

	t.Run("backend error", func(t *testing.T) {

		var autils = &abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		body := `---
addonVersion: "1.2.3"
addonProduct: "/DMO/PRODUCT"
repositories:
  - name: /DMO/SWC
    branch: main
    commitID: 1234abcd
    version: "4.5.6"
`
		file, _ := os.Create("repo.yml")
		_, err := file.Write([]byte(body))
		assert.NoError(t, err)
		config := &abapEnvironmentCreateTagOptions{
			Username:                            "dummy",
			Password:                            "dummy",
			Host:                                "https://test.com",
			Repositories:                        "repo.yml",
			RepositoryName:                      "/DMO/SWC2",
			CommitID:                            "1234abcde",
			TagName:                             "tag",
			TagDescription:                      "desc",
			GenerateTagForAddonProductVersion:   true,
			GenerateTagForAddonComponentVersion: true,
		}
		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : { "Status" : "S" } }`,
				`{"d" : { "uuid" : "abc" } }`,
				`{"d" : { "Status" : "S" } }`,
				`{"d" : { "uuid" : "abc" } }`,
				`{"d" : { "Status" : "S" } }`,
				`{"d" : { "uuid" : "abc" } }`,
				`{"d" : { "Status" : "S" } }`,
				`{"d" : { "uuid" : "abc" } }`,
				`{"d" : { "Status" : "S" } }`,
				`{"d" : { "uuid" : "abc" } }`,
				`{"d" : { "empty" : "body" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond, Force0510: true}
		err = runAbapEnvironmentCreateTag(config, autils, apiManager)

		assert.Error(t, err, "Did expect error")
		assert.Equal(t, "Something failed during the tag creation: Configuring the parameter repositories and the parameter repositoryName at the same time is not allowed", err.Error(), "Expected different error message")

	})

	t.Run("flags false", func(t *testing.T) {

		var autils = &abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
			_ = os.RemoveAll(dir)
		}()

		body := `---
addonVersion: "1.2.3"
addonProduct: "/DMO/PRODUCT"
repositories:
  - name: /DMO/SWC
    branch: main
    commitID: 1234abcd
    version: "4.5.6"
`
		file, _ := os.Create("repo.yml")
		_, err := file.Write([]byte(body))
		assert.NoError(t, err)
		config := &abapEnvironmentCreateTagOptions{
			Username:                            "dummy",
			Password:                            "dummy",
			Host:                                "https://test.com",
			Repositories:                        "repo.yml",
			TagName:                             "tag",
			TagDescription:                      "desc",
			GenerateTagForAddonProductVersion:   false,
			GenerateTagForAddonComponentVersion: false,
		}
		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : { "Status" : "S" } }`,
				`{"d" : { "uuid" : "abc" } }`,
				`{"d" : { "Status" : "S" } }`,
				`{"d" : { "uuid" : "abc" } }`,
				`{"d" : { "empty" : "body" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		_, hook := test.NewNullLogger()
		log.RegisterHook(hook)

		apiManager := &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond, Force0510: true}
		err = runAbapEnvironmentCreateTag(config, autils, apiManager)

		assert.Error(t, err, "Did expect error")
		assert.Equal(t, 8, len(hook.Entries), "Expected a different number of entries")
		assert.Equal(t, `NOT created: Tag tag for repository /DMO/SWC with commitID 1234abcd`, hook.AllEntries()[6].Message, "Expected a different message")
		hook.Reset()

	})
}
