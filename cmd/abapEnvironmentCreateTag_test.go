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
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		var autils = &abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		// autils.ReturnedConnectionDetailsHTTP.Host = "https://example.com"
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
			Username:                        "dummy",
			Password:                        "dummy",
			Host:                            "https://test.com",
			Repositories:                    "repo.yml",
			TagName:                         "tag",
			TagDescription:                  "desc",
			CreateTagForAddonProductVersion: true,
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
		assert.Equal(t, 3, len(hook.Entries))
		assert.Equal(t, `Created tag v4.5.6 for repository /DMO/SWC with commitID 1234abcd`, hook.AllEntries()[0].Message)
		assert.Equal(t, `Created tag -DMO-PRODUCT-1.2.3 for repository /DMO/SWC with commitID 1234abcd`, hook.AllEntries()[1].Message)
		assert.Equal(t, `Created tag tag for repository /DMO/SWC with commitID 1234abcd`, hook.AllEntries()[2].Message)

	})

}
