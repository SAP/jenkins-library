package btp

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestBTPReadServiceBinding(t *testing.T) {
	t.Run("BTP GetServiceBinding", func(t *testing.T) {
		//given
		const testURL = "testurl.com"
		btpConfig := GetServiceBindingOptions{
			Url:         "https://api.endpoint.com",
			Subdomain:   "xxxxxxx",
			Subaccount:  "yyyyyyy",
			BindingName: "testServiceBindingName",
			User:        "test_user",
			Password:    "test_password",
		}

		m := &mock.BtpExecuterMock{
			StdoutReturn: map[string]string{
				"btp login .*": "Authentication successful",
				"btp get services/binding": fmt.Sprintf(`
credentials:
  abap:
    communication_arrangement_id: SK-072EA3BC-778A-43CE-86D2-046BC208EFE3
    communication_inbound_user_auth_mode: 2
    communication_inbound_user_id: CC0000000001
    communication_scenario_id: SAP_COM_0948
    communication_system_id: SK-072EA3BC-778A-43CE-86D2-046BC208EFE3
    communication_type: inbound
    password: %s
    username: %s
  sap.cloud.service: com.sap.cloud.abap
  systemid: H01
  url: %s
name: %s

OK`, btpConfig.Password, btpConfig.User, testURL, btpConfig.BindingName),
			},
		}

		m.Stdout(new(bytes.Buffer))

		defer loginMockCleanup(m)

		//when
		var err error
		var btpServiceBinding string
		btp := BTPUtils{Exec: m}

		btpServiceBinding, err = btp.GetServiceBinding(btpConfig)

		//then
		if assert.NoError(t, err) {
			assert.NotEmpty(t, btpServiceBinding)
		}
	})
}

func TestBTPGetServiceInstance(t *testing.T) {
	t.Run("BTP GetServiceInstance", func(t *testing.T) {
		//given
		const testURL = "testurl.com"
		btpConfig := ServiceInstanceOptions{
			Url:          "https://api.endpoint.com",
			Subdomain:    "xxxxxxx",
			Subaccount:   "yyyyyyy",
			InstanceName: "testServiceInstanceName",
			User:         "test_user",
			Password:     "test_password",
		}

		m := &mock.BtpExecuterMock{
			StdoutReturn: map[string]string{
				"btp login .*": "Authentication successful",
				"btp get services/instance": fmt.Sprintf(`
context:
  crm_customer_id:
  env_type: sapcp
id: 005cc00a-6c85-4e01-a5b8-7dd944ff26fb
labels: subaccount_id = f57f211e-2733-4cc6-b645-74f02d034a58
name: %s

OK`, btpConfig.InstanceName),
			},
		}

		m.Stdout(new(bytes.Buffer))

		defer loginMockCleanup(m)

		//when
		var err error
		var btpServiceInstance string
		btp := BTPUtils{Exec: m}

		btpServiceInstance, err = btp.GetServiceInstance(btpConfig)

		//then
		if assert.NoError(t, err) {
			assert.NotEmpty(t, btpServiceInstance)
		}
	})
}

func TestBTPDeleteServiceInstance(t *testing.T) {
	t.Run("BTP DeleteServiceInstance not working", func(t *testing.T) {
		//given
		const testURL = "testurl.com"
		btpConfig := ServiceInstanceOptions{
			Url:          "https://api.endpoint.com",
			Subdomain:    "xxxxxxx",
			Subaccount:   "yyyyyyy",
			InstanceName: "testServiceInstanceName",
			User:         "test_user",
			Password:     "test_password",
		}

		m := &mock.BtpExecuterMock{
			StdoutReturn: map[string]string{
				"btp login .*": "Authentication successful",
				"btp get services/instance": fmt.Sprintf(`
context:
  crm_customer_id:
  env_type: sapcp
id: 005cc00a-6c85-4e01-a5b8-7dd944ff26fb
labels: subaccount_id = f57f211e-2733-4cc6-b645-74f02d034a58
name: %s

OK`, btpConfig.InstanceName),
			},
		}

		m.Stdout(new(bytes.Buffer))

		defer loginMockCleanup(m)

		//when
		var err error
		btp := BTPUtils{Exec: m}

		err = btp.DeleteServiceInstance(btpConfig)

		//then
		assert.Error(t, err)
	})

	t.Run("BTP DeleteServiceInstance working", func(t *testing.T) {
		//given
		const testURL = "testurl.com"
		btpConfig := ServiceInstanceOptions{
			Url:          "https://api.endpoint.com",
			Subdomain:    "xxxxxxx",
			Subaccount:   "yyyyyyy",
			InstanceName: "testServiceInstanceName",
			User:         "test_user",
			Password:     "test_password",
		}

		m := &mock.BtpExecuterMock{
			StdoutReturn: map[string]string{
				"btp login .*": "Authentication successful",
				"btp get services/instance": `
error: BadRequest
description: Could not find such instance

FAILED`,
			},
		}

		m.Stdout(new(bytes.Buffer))

		defer loginMockCleanup(m)

		//when
		var err error
		btp := BTPUtils{Exec: m}

		err = btp.DeleteServiceInstance(btpConfig)

		//then
		assert.NoError(t, err)
	})
}

func TestBTPDeleteServiceBinding(t *testing.T) {
	t.Run("BTP DeleteServiceBinding not working", func(t *testing.T) {
		//given
		const testURL = "testurl.com"
		btpConfig := GetServiceBindingOptions{
			Url:         "https://api.endpoint.com",
			Subdomain:   "xxxxxxx",
			Subaccount:  "yyyyyyy",
			BindingName: "testServiceBindingName",
			User:        "test_user",
			Password:    "test_password",
		}

		m := &mock.BtpExecuterMock{
			StdoutReturn: map[string]string{
				"btp login .*": "Authentication successful",
				"btp get services/binding": fmt.Sprintf(`
credentials:
  abap:
    communication_arrangement_id: SK-072EA3BC-778A-43CE-86D2-046BC208EFE3
    communication_inbound_user_auth_mode: 2
    communication_inbound_user_id: CC0000000001
    communication_scenario_id: SAP_COM_0948
    communication_system_id: SK-072EA3BC-778A-43CE-86D2-046BC208EFE3
    communication_type: inbound
    password: %s
    username: %s
  sap.cloud.service: com.sap.cloud.abap
  systemid: H01
  url: %s
name: %s

OK`, btpConfig.Password, btpConfig.User, testURL, btpConfig.BindingName),
			},
		}

		m.Stdout(new(bytes.Buffer))

		defer loginMockCleanup(m)

		//when
		var err error
		btp := BTPUtils{Exec: m}

		err = btp.DeleteServiceBinding(btpConfig)

		//then
		assert.Error(t, err)
	})

	t.Run("BTP DeleteServiceBinding working", func(t *testing.T) {
		//given
		const testURL = "testurl.com"
		btpConfig := GetServiceBindingOptions{
			Url:         "https://api.endpoint.com",
			Subdomain:   "xxxxxxx",
			Subaccount:  "yyyyyyy",
			BindingName: "testServiceBindingName",
			User:        "test_user",
			Password:    "test_password",
		}

		m := &mock.BtpExecuterMock{
			StdoutReturn: map[string]string{
				"btp login .*": "Authentication successful",
				"btp get services/binding": `
error: BadRequest
description: Could not find such binding

FAILED`,
			},
		}

		m.Stdout(new(bytes.Buffer))

		defer loginMockCleanup(m)

		//when
		var err error
		btp := BTPUtils{Exec: m}

		err = btp.DeleteServiceBinding(btpConfig)

		//then
		assert.NoError(t, err)
	})
}
