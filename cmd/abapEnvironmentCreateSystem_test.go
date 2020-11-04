package cmd

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestRunAbapEnvironmentCreateSystem(t *testing.T) {
	m := &mock.ExecMockRunner{}
	cf := cloudfoundry.CFUtils{Exec: m}
	u := &uuidMock{}

	t.Run("Create service with generated manifest", func(t *testing.T) {
		defer cfMockCleanup(m)
		config := abapEnvironmentCreateSystemOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			Username:          "testUser",
			Password:          "testPassword",
			CfService:         "testService",
			CfServiceInstance: "testName",
			CfServicePlan:     "testPlan",
		}
		wd, _ := os.Getwd()
		err := runAbapEnvironmentCreateSystem(&config, nil, cf, u)
		if assert.NoError(t, err) {
			assert.Equal(t, []mock.ExecCall{
				{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"login", "-a", "https://api.endpoint.com", "-o", "testOrg", "-s", "testSpace", "-u", "testUser", "-p", "testPassword"}},
				{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"create-service-push", "--no-push", "--service-manifest", wd + "/generated_service_manifest-my-uuid.yml"}},
				{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"logout"}}},
				m.Calls)
		}
	})

	t.Run("Create service with mainfest", func(t *testing.T) {
		defer cfMockCleanup(m)
		config := abapEnvironmentCreateSystemOptions{
			CfAPIEndpoint:   "https://api.endpoint.com",
			CfOrg:           "testOrg",
			CfSpace:         "testSpace",
			Username:        "testUser",
			Password:        "testPassword",
			ServiceManifest: "customManifest.yml",
		}

		dir, err := ioutil.TempDir("", "test variable substitution")
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

		manifestFileString := `
		---
		create-services:
		- name:   ((name))
		  broker: "testBroker"
		  plan:   "testPlan"

		- name:   ((name2))
		  broker: "testBroker"
		  plan:   "testPlan"

		- name:   "test3"
		  broker: "testBroker"
		  plan:   "testPlan"`

		manifestFileStringBody := []byte(manifestFileString)
		err = ioutil.WriteFile("customManifest.yml", manifestFileStringBody, 0644)

		err = runAbapEnvironmentCreateSystem(&config, nil, cf, u)
		if assert.NoError(t, err) {
			assert.Equal(t, []mock.ExecCall{
				{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"login", "-a", "https://api.endpoint.com", "-o", "testOrg", "-s", "testSpace", "-u", "testUser", "-p", "testPassword"}},
				{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"create-service-push", "--no-push", "--service-manifest", "customManifest.yml"}},
				{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"logout"}}},
				m.Calls)
		}
	})
}

func TestManifestGeneration(t *testing.T) {

	t.Run("Create service with generated manifest", func(t *testing.T) {
		config := abapEnvironmentCreateSystemOptions{
			CfAPIEndpoint:                  "https://api.endpoint.com",
			CfOrg:                          "testOrg",
			CfSpace:                        "testSpace",
			Username:                       "testUser",
			Password:                       "testPassword",
			CfService:                      "testService",
			CfServiceInstance:              "testName",
			CfServicePlan:                  "testPlan",
			AbapSystemAdminEmail:           "user@example.com",
			AbapSystemID:                   "H02",
			AbapSystemIsDevelopmentAllowed: true,
			AbapSystemSizeOfPersistence:    4,
			AbapSystemSizeOfRuntime:        4,
			AddonDescriptorFileName:        "addon.yml",
		}

		dir, err := ioutil.TempDir("", "test variable substitution")
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

		addonYML := `addonProduct: myProduct
addonVersion: 1.2.3
repositories:
  - name: '/DMO/REPO'
`

		addonYMLBytes := []byte(addonYML)
		err = ioutil.WriteFile("addon.yml", addonYMLBytes, 0644)

		expectedResult := `create-services:
- broker: testService
  name: testName
  parameters: '{"admin_email":"user@example.com","is_development_allowed":true,"sapsystemname":"H02","size_of_persistence":4,"size_of_runtime":4}'
  plan: testPlan
`

		resultBytes, err := generateManifestYAML(&config)

		if assert.NoError(t, err) {
			result := string(resultBytes)
			assert.Equal(t, expectedResult, result, "Result not as expected")
		}
	})

	t.Run("Create service with generated manifest - with addon", func(t *testing.T) {
		config := abapEnvironmentCreateSystemOptions{
			CfAPIEndpoint:                  "https://api.endpoint.com",
			CfOrg:                          "testOrg",
			CfSpace:                        "testSpace",
			Username:                       "testUser",
			Password:                       "testPassword",
			CfService:                      "testService",
			CfServiceInstance:              "testName",
			CfServicePlan:                  "testPlan",
			AbapSystemAdminEmail:           "user@example.com",
			AbapSystemID:                   "H02",
			AbapSystemIsDevelopmentAllowed: true,
			AbapSystemSizeOfPersistence:    4,
			AbapSystemSizeOfRuntime:        4,
			AddonDescriptorFileName:        "addon.yml",
			IncludeAddon:                   true,
		}

		dir, err := ioutil.TempDir("", "test variable substitution")
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

		addonYML := `addonProduct: myProduct
addonVersion: 1.2.3
repositories:
  - name: '/DMO/REPO'
`

		addonYMLBytes := []byte(addonYML)
		err = ioutil.WriteFile("addon.yml", addonYMLBytes, 0644)

		expectedResult := `create-services:
- broker: testService
  name: testName
  parameters: '{"admin_email":"user@example.com","is_development_allowed":true,"sapsystemname":"H02","size_of_persistence":4,"size_of_runtime":4,"addon_product_name":"myProduct","addon_product_version":"1.2.3"}'
  plan: testPlan
`

		resultBytes, err := generateManifestYAML(&config)

		if assert.NoError(t, err) {
			result := string(resultBytes)
			assert.Equal(t, expectedResult, result, "Result not as expected")
		}
	})
}

type uuidMock struct {
}

func (u *uuidMock) getUUID() string {
	return "my-uuid"
}
