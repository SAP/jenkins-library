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

	t.Run("Create service with generated manifest - abap-oem", func(t *testing.T) {
		config := abapEnvironmentCreateSystemOptions{
			CfAPIEndpoint:                       "https://api.endpoint.com",
			CfOrg:                               "testOrg",
			CfSpace:                             "testSpace",
			Username:                            "testUser",
			Password:                            "testPassword",
			CfService:                           "abap-oem",
			CfServiceInstance:                   "testName",
			CfServicePlan:                       "testPlan",
			AbapSystemAdminEmail:                "user@example.com",
			AbapSystemID:                        "H02",
			AbapSystemIsDevelopmentAllowed:      true,
			AbapSystemSizeOfPersistence:         4,
			AbapSystemSizeOfRuntime:             4,
			AddonDescriptorFileName:             "addon.yml",
			AbapSystemParentServiceLabel:        "abap-trial",
			AbapSystemParentServiceInstanceGuid: "131bb94b-3045-4303-94bc-34df92072302",
			AbapSystemParentSaasAppname:         "abapcp-saas-itapcao1",
			AbapSystemParentServiceParameters:   `{"foo":"bar","veryspecialfeature":"true"}`,
			AbapSystemConsumerTenantLimit:       1,
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
- broker: abap-oem
  name: testName
  parameters: '{"admin_email":"user@example.com","is_development_allowed":true,"sapsystemname":"H02","size_of_persistence":4,"size_of_runtime":4,"parent_service_label":"abap-trial","parent_service_instance_guid":"131bb94b-3045-4303-94bc-34df92072302","parent_saas_appname":"abapcp-saas-itapcao1","parent_service_parameters":"{\"foo\":\"bar\",\"veryspecialfeature\":\"true\"}","consumer_tenant_limit":1}'
  plan: testPlan
`

		resultBytes, err := generateManifestYAML(&config)

		if assert.NoError(t, err) {
			result := string(resultBytes)
			assert.Equal(t, expectedResult, result, "Result not as expected")
		}
	})
}

func TestManifestCheck(t *testing.T) {

	t.Run("check manifest parameters - success", func(t *testing.T) {
		config := abapEnvironmentCreateSystemOptions{
			CfAPIEndpoint:                       "https://api.endpoint.com",
			CfOrg:                               "testOrg",
			CfSpace:                             "testSpace",
			Username:                            "testUser",
			Password:                            "testPassword",
			CfService:                           "testService",
			CfServiceInstance:                   "testName",
			CfServicePlan:                       "testPlan",
			AbapSystemAdminEmail:                "user@example.com",
			AbapSystemID:                        "H02",
			AbapSystemIsDevelopmentAllowed:      true,
			AbapSystemSizeOfPersistence:         4,
			AbapSystemSizeOfRuntime:             4,
			AddonDescriptorFileName:             "addon.yml",
			AbapSystemParentServiceLabel:        "abap-trial",
			AbapSystemParentServiceInstanceGuid: "131bb94b-3045-4303-94bc-34df92072302",
			AbapSystemParentSaasAppname:         "abapcp-saas-itapcao1",
			AbapSystemParentServiceParameters:   `{"foo":"bar","veryspecialfeature":"true"}`,
			AbapSystemConsumerTenantLimit:       1,
		}
		err := checkManifestParameters(&config)

		assert.NoError(t, err)
	})

	t.Run("check manifest parameters - wrong SID", func(t *testing.T) {
		expectedErrorMessage := "It seems like you have incorrectly specified the AbapSystemID step parameter. Please check that the parameters follows the respective syntax to specify the AbapSystemID. For more information please refer to the step documentation"

		config := abapEnvironmentCreateSystemOptions{
			CfAPIEndpoint:                       "https://api.endpoint.com",
			CfOrg:                               "testOrg",
			CfSpace:                             "testSpace",
			Username:                            "testUser",
			Password:                            "testPassword",
			CfService:                           "testService",
			CfServiceInstance:                   "testName",
			CfServicePlan:                       "testPlan",
			AbapSystemAdminEmail:                "user@example.com",
			AbapSystemID:                        "***",
			AbapSystemIsDevelopmentAllowed:      true,
			AbapSystemSizeOfPersistence:         4,
			AbapSystemSizeOfRuntime:             4,
			AddonDescriptorFileName:             "addon.yml",
			AbapSystemParentServiceLabel:        "abap-trial",
			AbapSystemParentServiceInstanceGuid: "131bb94b-3045-4303-94bc-34df92072302",
			AbapSystemParentSaasAppname:         "abapcp-saas-itapcao1",
			AbapSystemParentServiceParameters:   `{"foo":"bar","veryspecialfeature":"true"}`,
			AbapSystemConsumerTenantLimit:       1,
		}
		err := checkManifestParameters(&config)

		assert.EqualError(t, err, expectedErrorMessage)
	})

	t.Run("check manifest parameters - wrong consumer tenant limit", func(t *testing.T) {
		expectedErrorMessage := "You have specified 0 tenants o be created in the system for the step parameter AbapSystemConsumerTenantLimit. Please check that you have set the parameter value correctly. For more information please refer to the step documentation"
		config := abapEnvironmentCreateSystemOptions{
			CfAPIEndpoint:                       "https://api.endpoint.com",
			CfOrg:                               "testOrg",
			CfSpace:                             "testSpace",
			Username:                            "testUser",
			Password:                            "testPassword",
			CfService:                           "testService",
			CfServiceInstance:                   "testName",
			CfServicePlan:                       "testPlan",
			AbapSystemAdminEmail:                "user@example.com",
			AbapSystemID:                        "H02",
			AbapSystemIsDevelopmentAllowed:      true,
			AbapSystemSizeOfPersistence:         4,
			AbapSystemSizeOfRuntime:             4,
			AddonDescriptorFileName:             "addon.yml",
			AbapSystemParentServiceLabel:        "abap-trial",
			AbapSystemParentServiceInstanceGuid: "131bb94b-3045-4303-94bc-34df92072302",
			AbapSystemParentSaasAppname:         "abapcp-saas-itapcao1",
			AbapSystemParentServiceParameters:   `{"foo":"bar","veryspecialfeature":"true"}`,
			AbapSystemConsumerTenantLimit:       0,
		}
		err := checkManifestParameters(&config)

		assert.EqualError(t, err, expectedErrorMessage)
	})

	t.Run("check manifest parameters - wrong config", func(t *testing.T) {
		expectedErrorMessage := "Both parameters AbapSystemParentServiceLabel and AbapSystemParentSaasAppname seem to be empty. Please specify either AbapSystemParentServiceLabel or AbapSystemParentSaasAppname depending on who created the oem-instance in the step configuration. For more information please refer to the step documentation"
		config := abapEnvironmentCreateSystemOptions{
			CfAPIEndpoint:                       "https://api.endpoint.com",
			CfOrg:                               "testOrg",
			CfSpace:                             "testSpace",
			Username:                            "testUser",
			Password:                            "testPassword",
			CfService:                           "testService",
			CfServiceInstance:                   "testName",
			CfServicePlan:                       "testPlan",
			AbapSystemAdminEmail:                "user@example.com",
			AbapSystemID:                        "H02",
			AbapSystemIsDevelopmentAllowed:      true,
			AbapSystemSizeOfPersistence:         4,
			AbapSystemSizeOfRuntime:             4,
			AddonDescriptorFileName:             "addon.yml",
			AbapSystemParentServiceInstanceGuid: "131bb94b-3045-4303-94bc-34df92072302",
			AbapSystemParentServiceParameters:   `{"foo":"bar","veryspecialfeature":"true"}`,
			AbapSystemConsumerTenantLimit:       1,
		}
		err := checkManifestParameters(&config)

		assert.EqualError(t, err, expectedErrorMessage)

	})

	t.Run("check manifest parameters - wrong Saas Appname", func(t *testing.T) {
		expectedErrorMessage := "It seems like you have incorrectly specified the AbapSystemParentSaasAppname step parameter. Please check that the parameters follows the respective syntax to specify the AbapSystemParentSaasAppname. For more information please refer to the step documentation"
		config := abapEnvironmentCreateSystemOptions{
			CfAPIEndpoint:                       "https://api.endpoint.com",
			CfOrg:                               "testOrg",
			CfSpace:                             "testSpace",
			Username:                            "testUser",
			Password:                            "testPassword",
			CfService:                           "testService",
			CfServiceInstance:                   "testName",
			CfServicePlan:                       "testPlan",
			AbapSystemAdminEmail:                "user@example.com",
			AbapSystemID:                        "H02",
			AbapSystemIsDevelopmentAllowed:      true,
			AbapSystemSizeOfPersistence:         4,
			AbapSystemSizeOfRuntime:             4,
			AddonDescriptorFileName:             "addon.yml",
			AbapSystemParentServiceLabel:        "abap-trial",
			AbapSystemParentServiceInstanceGuid: "131bb94b-3045-4303-94bc-34df92072302",
			AbapSystemParentSaasAppname:         "***",
			AbapSystemParentServiceParameters:   `{"foo":"bar","veryspecialfeature":"true"}`,
			AbapSystemConsumerTenantLimit:       1,
		}
		err := checkManifestParameters(&config)

		assert.EqualError(t, err, expectedErrorMessage)

	})
}

type uuidMock struct {
}

func (u *uuidMock) getUUID() string {
	return "my-uuid"
}
