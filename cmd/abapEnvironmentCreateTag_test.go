package cmd

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

type abapEnvironmentCreateTagMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newAbapEnvironmentCreateTagTestsUtils() abapEnvironmentCreateTagMockUtils {
	utils := abapEnvironmentCreateTagMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunAbapEnvironmentCreateTag(t *testing.T) {

	t.Run("happy path", func(t *testing.T) {

		var autils = &abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		dir, errDir := ioutil.TempDir("", "test read addon descriptor")
		if errDir != nil {
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
addonVersion: "1.2.3"
addonProduct: "/DMO/PRODUCT"
repositories:
  - name: /DMO/SWC
    branch: main
    commitID: 1234abcd
    version: "4.5.6"
`
		file, _ := os.Create("repo.yml")
		file.Write([]byte(body))
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

		_, hook := test.NewNullLogger()
		log.RegisterHook(hook)

		err := runAbapEnvironmentCreateTag(config, nil, autils, client)

		assert.NoError(t, err, "Did not expect error")
		assert.Equal(t, 3, len(hook.Entries), "Expected a different number of entries")
		assert.Equal(t, `Created tag v4.5.6 for repository /DMO/SWC with commitID 1234abcd`, hook.AllEntries()[0].Message, "Expected a different message")
		assert.Equal(t, `Created tag -DMO-PRODUCT-1.2.3 for repository /DMO/SWC with commitID 1234abcd`, hook.AllEntries()[1].Message, "Expected a different message")
		assert.Equal(t, `Created tag tag for repository /DMO/SWC with commitID 1234abcd`, hook.AllEntries()[2].Message, "Expected a different message")
		hook.Reset()
	})

	t.Run("backend error", func(t *testing.T) {

		var autils = &abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		dir, errDir := ioutil.TempDir("", "test read addon descriptor")
		if errDir != nil {
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
addonVersion: "1.2.3"
addonProduct: "/DMO/PRODUCT"
repositories:
  - name: /DMO/SWC
    branch: main
    commitID: 1234abcd
    version: "4.5.6"
`
		file, _ := os.Create("repo.yml")
		file.Write([]byte(body))
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
				`{"d" : { "Status" : "E" } }`,
				`{"d" : { "uuid" : "abc" } }`,
				`{"d" : { "Status" : "E" } }`,
				`{"d" : { "uuid" : "abc" } }`,
				`{"d" : { "Status" : "E" } }`,
				`{"d" : { "uuid" : "abc" } }`,
				`{"d" : { "empty" : "body" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		_, hook := test.NewNullLogger()
		log.RegisterHook(hook)

		err := runAbapEnvironmentCreateTag(config, nil, autils, client)

		assert.Error(t, err, "Did expect error")
		assert.Equal(t, 4, len(hook.Entries), "Expected a different number of entries")
		assert.Equal(t, `NOT created: Tag v4.5.6 for repository /DMO/SWC with commitID 1234abcd`, hook.AllEntries()[0].Message, "Expected a different message")
		assert.Equal(t, `NOT created: Tag -DMO-PRODUCT-1.2.3 for repository /DMO/SWC with commitID 1234abcd`, hook.AllEntries()[1].Message, "Expected a different message")
		assert.Equal(t, `NOT created: Tag tag for repository /DMO/SWC with commitID 1234abcd`, hook.AllEntries()[2].Message, "Expected a different message")
		assert.Equal(t, `At least one tag has not been created`, hook.AllEntries()[3].Message, "Expected a different message")
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
				`{"d" : { "Status" : "S" } }`,
				`{"d" : { "uuid" : "abc" } }`,
				`{"d" : { "empty" : "body" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		_, hook := test.NewNullLogger()
		log.RegisterHook(hook)

		err := runAbapEnvironmentCreateTag(config, nil, autils, client)

		assert.NoError(t, err, "Did not expect error")
		assert.Equal(t, 1, len(hook.Entries), "Expected a different number of entries")
		assert.Equal(t, `Created tag tag for repository /DMO/SWC with commitID 1234abcd`, hook.AllEntries()[0].Message, "Expected a different message")
		hook.Reset()
	})

	t.Run("backend error", func(t *testing.T) {

		var autils = &abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		dir, errDir := ioutil.TempDir("", "test read addon descriptor")
		if errDir != nil {
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
addonVersion: "1.2.3"
addonProduct: "/DMO/PRODUCT"
repositories:
  - name: /DMO/SWC
    branch: main
    commitID: 1234abcd
    version: "4.5.6"
`
		file, _ := os.Create("repo.yml")
		file.Write([]byte(body))
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

		_, hook := test.NewNullLogger()
		log.RegisterHook(hook)

		err := runAbapEnvironmentCreateTag(config, nil, autils, client)

		assert.NoError(t, err, "Did not expect error")
		assert.Equal(t, 4, len(hook.Entries), "Expected a different number of entries")
		assert.Equal(t, `Created tag v4.5.6 for repository /DMO/SWC with commitID 1234abcd`, hook.AllEntries()[0].Message, "Expected a different message")
		assert.Equal(t, `Created tag -DMO-PRODUCT-1.2.3 for repository /DMO/SWC with commitID 1234abcd`, hook.AllEntries()[1].Message, "Expected a different message")
		assert.Equal(t, `Created tag tag for repository /DMO/SWC with commitID 1234abcd`, hook.AllEntries()[2].Message, "Expected a different message")
		assert.Equal(t, `Created tag tag for repository /DMO/SWC2 with commitID 1234abcde`, hook.AllEntries()[3].Message, "Expected a different message")
		hook.Reset()

	})
}
