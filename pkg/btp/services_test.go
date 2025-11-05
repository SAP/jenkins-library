package btp

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBTPCreateServiceBinding(t *testing.T) {
	t.Run("BTP CreateServiceBinding", func(t *testing.T) {
		//given
		btpConfig := CreateServiceBindingOptions{
			Url:             "https://api.endpoint.com",
			Subdomain:       "xxxxxxx",
			Subaccount:      "yyyyyyy",
			BindingName:     "testServiceBindingName",
			User:            "test_user",
			Password:        "test_password",
			ServiceInstance: "test_instance",
			Parameters:      "test.json",
			Timeout:         3600,
			PollInterval:    600,
		}

		m := &BtpExecutorMock{
			StdoutReturn: map[string]string{
				"btp login .*": "Authentication successful",
				"btp get services/binding": fmt.Sprintf(`
				{
				"id": "xxxx",
				"name": "%s",
				"ready": true
				}`, btpConfig.BindingName),
			},
		}

		m.Stdout(new(bytes.Buffer))

		defer loginMockCleanup(m)

		//when
		var err error
		var btpServiceBinding string
		btp := NewBTPUtils(m)

		btpServiceBinding, err = btp.CreateServiceBinding(btpConfig)

		//then
		if assert.NoError(t, err) {
			assert.NotEmpty(t, btpServiceBinding)
		}
	})
}

func TestBTPGetServiceBinding(t *testing.T) {
	t.Run("BTP GetServiceBinding", func(t *testing.T) {
		//given
		btpConfig := GetServiceBindingOptions{
			Url:         "https://api.endpoint.com",
			Subdomain:   "xxxxxxx",
			Subaccount:  "yyyyyyy",
			BindingName: "testServiceBindingName",
			User:        "test_user",
			Password:    "test_password",
		}

		m := &BtpExecutorMock{
			StdoutReturn: map[string]string{
				"btp login .*": "Authentication successful",
				"btp get services/binding": fmt.Sprintf(`
				{
				"id": "xxxx",
				"name": "%s",
				"ready": true
				}`, btpConfig.BindingName),
			},
		}

		m.Stdout(new(bytes.Buffer))

		defer loginMockCleanup(m)

		//when
		var err error
		var btpServiceBinding string
		btp := NewBTPUtils(m)

		btpServiceBinding, err = btp.GetServiceBinding(btpConfig)

		//then
		if assert.NoError(t, err) {
			assert.NotEmpty(t, btpServiceBinding)
		}
	})
}

func TestBTPDeleteServiceBinding(t *testing.T) {
	t.Run("BTP DeleteServiceBinding not working", func(t *testing.T) {
		//given
		btpConfig := DeleteServiceBindingOptions{
			Url:          "https://api.endpoint.com",
			Subdomain:    "xxxxxxx",
			Subaccount:   "yyyyyyy",
			BindingName:  "testServiceBindingName",
			User:         "test_user",
			Password:     "test_password",
			Timeout:      3600,
			PollInterval: 600,
		}

		m := &BtpExecutorMock{
			StdoutReturn: map[string]string{
				"btp login .*": "Authentication successful",
			},
			ShouldFailOnCommand: map[string]error{
				"btp delete services/binding": fmt.Errorf(`
				{
				"error": "BadRequest",
				"description": "Could not find such binding"
				}`),
			},
		}

		m.Stdout(new(bytes.Buffer))

		defer loginMockCleanup(m)

		//when
		var err error
		btp := NewBTPUtils(m)

		err = btp.DeleteServiceBinding(btpConfig)

		//then
		assert.Error(t, err)
	})

	t.Run("BTP DeleteServiceBinding working", func(t *testing.T) {
		//given
		btpConfig := DeleteServiceBindingOptions{
			Url:          "https://api.endpoint.com",
			Subdomain:    "xxxxxxx",
			Subaccount:   "yyyyyyy",
			BindingName:  "testServiceBindingName",
			User:         "test_user",
			Password:     "test_password",
			Timeout:      3600,
			PollInterval: 600,
		}

		m := &BtpExecutorMock{
			StdoutReturn: map[string]string{
				"btp login .*": "Authentication successful",
			},
			ShouldFailOnCommand: map[string]error{
				"btp get services/binding": fmt.Errorf(`
				{
				"error": "BadRequest",
				"description": "Could not find such binding"
				}`),
			},
		}

		m.Stdout(new(bytes.Buffer))

		defer loginMockCleanup(m)

		//when
		var err error
		btp := NewBTPUtils(m)

		err = btp.DeleteServiceBinding(btpConfig)

		//then
		assert.NoError(t, err)
	})
}

func TestBTPCreateServiceInstance(t *testing.T) {
	t.Run("BTP CreateServiceInstance", func(t *testing.T) {
		//given
		btpConfig := CreateServiceInstanceOptions{
			Url:          "https://api.endpoint.com",
			Subdomain:    "xxxxxxx",
			Subaccount:   "yyyyyyy",
			User:         "test_user",
			Password:     "test_password",
			Tenant:       "test_tenant",
			PlanName:     "test_plan",
			OfferingName: "test_offering",
			InstanceName: "test_instance",
			Parameters:   "test_parameter",
			Timeout:      3600,
			PollInterval: 600,
		}

		m := &BtpExecutorMock{
			StdoutReturn: map[string]string{
				"btp login .*": "Authentication successful",
				"btp get services/instance": fmt.Sprintf(`
				{
					"id": "xxx",
					"name": "%s",
					"ready": true
				}`, btpConfig.InstanceName),
			},
		}

		m.Stdout(new(bytes.Buffer))

		defer loginMockCleanup(m)

		//when
		var err error
		var btpServiceBinding string
		btp := NewBTPUtils(m)

		btpServiceBinding, err = btp.CreateServiceInstance(btpConfig)

		//then
		if assert.NoError(t, err) {
			assert.NotEmpty(t, btpServiceBinding)
		}
	})
}

func TestBTPGetServiceInstance(t *testing.T) {
	t.Run("BTP GetServiceInstance", func(t *testing.T) {
		//given
		btpConfig := GetServiceInstanceOptions{
			Url:          "https://api.endpoint.com",
			Subdomain:    "xxxxxxx",
			Subaccount:   "yyyyyyy",
			InstanceName: "testServiceInstanceName",
			User:         "test_user",
			Password:     "test_password",
		}

		m := &BtpExecutorMock{
			StdoutReturn: map[string]string{
				"btp login .*": "Authentication successful",
				"btp get services/instance": fmt.Sprintf(`
				{
				"id": "xxx",
				"name": "%s",
				"ready": true
				}`, btpConfig.InstanceName),
			},
		}

		m.Stdout(new(bytes.Buffer))

		defer loginMockCleanup(m)

		//when
		var err error
		var btpServiceInstance string
		btp := NewBTPUtils(m)

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
		btpConfig := DeleteServiceInstanceOptions{
			Url:          "https://api.endpoint.com",
			Subdomain:    "xxxxxxx",
			Subaccount:   "yyyyyyy",
			InstanceName: "testServiceInstanceName",
			User:         "test_user",
			Password:     "test_password",
			Timeout:      3600,
			PollInterval: 600,
		}

		m := &BtpExecutorMock{
			StdoutReturn: map[string]string{
				"btp login .*": "Authentication successful",
			},
			ShouldFailOnCommand: map[string]error{
				"btp delete services/instance": fmt.Errorf(`
				{
				"error": "BadRequest",
				"description": "Could not find such instance"
				}`),
			},
		}

		m.Stdout(new(bytes.Buffer))

		defer loginMockCleanup(m)

		//when
		var err error
		btp := NewBTPUtils(m)

		err = btp.DeleteServiceInstance(btpConfig)

		//then
		assert.Error(t, err)
	})

	t.Run("BTP DeleteServiceInstance working", func(t *testing.T) {
		//given
		btpConfig := DeleteServiceInstanceOptions{
			Url:          "https://api.endpoint.com",
			Subdomain:    "xxxxxxx",
			Subaccount:   "yyyyyyy",
			InstanceName: "testServiceInstanceName",
			User:         "test_user",
			Password:     "test_password",
			Timeout:      3600,
			PollInterval: 600,
		}

		m := &BtpExecutorMock{
			StdoutReturn: map[string]string{
				"btp login .*": "Authentication successful",
			},
			ShouldFailOnCommand: map[string]error{
				"btp get services/instance": fmt.Errorf(`
				{
				"error": "BadRequest",
				"description": "Could not find such instance"
				}`),
			},
		}

		m.Stdout(new(bytes.Buffer))

		defer loginMockCleanup(m)

		//when
		var err error
		btp := NewBTPUtils(m)

		err = btp.DeleteServiceInstance(btpConfig)

		//then
		assert.NoError(t, err)
	})
}
