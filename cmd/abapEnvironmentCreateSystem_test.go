//go:build unit

package cmd

import (
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

	t.Run("Create service with cf create-service", func(t *testing.T) {
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
		err := runAbapEnvironmentCreateSystem(&config, nil, cf, u)
		if assert.NoError(t, err) {
			assert.Equal(t, []mock.ExecCall{
				{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"login", "-a", "https://api.endpoint.com", "-o", "testOrg", "-s", "testSpace", "-u", "testUser", "-p", "testPassword"}},
				{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"create-service", config.CfService, config.CfServicePlan, config.CfServiceInstance, "-c", "{\"is_development_allowed\":false}", "--wait"}},
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

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
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
		err := os.WriteFile("customManifest.yml", manifestFileStringBody, 0644)

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

	t.Run("Create service with addon - development", func(t *testing.T) {
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

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		addonYML := `addonProduct: myProduct
addonVersion: 1.2.3
repositories:
  - name: '/DMO/REPO'
`

		addonYMLBytes := []byte(addonYML)
		err := os.WriteFile("addon.yml", addonYMLBytes, 0644)

		expectedResult := "{\"admin_email\":\"user@example.com\",\"is_development_allowed\":true,\"sapsystemname\":\"H02\",\"size_of_persistence\":4,\"size_of_runtime\":4,\"addon_product_name\":\"myProduct\",\"addon_product_version\":\"1.2.3\",\"parent_saas_appname\":\"addon_test\"}"

		result, err := generateServiceParameterString(&config)

		if assert.NoError(t, err) {
			assert.Equal(t, expectedResult, result, "Result not as expected")
		}
	})

	t.Run("Test IsDevelopmentAllowed", func(t *testing.T) {
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

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		addonYML := `addonProduct: myProduct
addonVersion: 1.2.3
repositories:
  - name: '/DMO/REPO'
`

		addonYMLBytes := []byte(addonYML)
		err := os.WriteFile("addon.yml", addonYMLBytes, 0644)

		expectedResult := "{\"admin_email\":\"user@example.com\",\"is_development_allowed\":true,\"sapsystemname\":\"H02\",\"size_of_persistence\":4,\"size_of_runtime\":4}"

		result, err := generateServiceParameterString(&config)

		if assert.NoError(t, err) {
			assert.Equal(t, expectedResult, result, "Result not as expected")
		}
	})

	t.Run("Create service with addon - no development", func(t *testing.T) {

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
			AbapSystemIsDevelopmentAllowed: false,
			AbapSystemSizeOfPersistence:    4,
			AbapSystemSizeOfRuntime:        4,
			AddonDescriptorFileName:        "addon.yml",
			IncludeAddon:                   true,
		}

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		addonYML := `addonProduct: myProduct
addonVersion: 1.2.3
repositories:
  - name: '/DMO/REPO'
`

		addonYMLBytes := []byte(addonYML)
		err := os.WriteFile("addon.yml", addonYMLBytes, 0644)

		expectedResult := "{\"admin_email\":\"user@example.com\",\"is_development_allowed\":false,\"sapsystemname\":\"H02\",\"size_of_persistence\":4,\"size_of_runtime\":4,\"addon_product_name\":\"myProduct\",\"addon_product_version\":\"1.2.3\",\"parent_saas_appname\":\"addon_test\"}"

		result, err := generateServiceParameterString(&config)

		if assert.NoError(t, err) {
			assert.Equal(t, expectedResult, result, "Result not as expected")
		}
	})
}

type uuidMock struct {
}

func (u *uuidMock) getUUID() string {
	return "my-uuid"
}
