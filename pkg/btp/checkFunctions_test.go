package btp

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsServiceInstanceCreated(t *testing.T) {
	btpConfig := GetServiceInstanceOptions{
		Url:          "https://api.endpoint.com",
		Subdomain:    "xxxxxxx",
		Subaccount:   "yyyyyyy",
		InstanceName: "testServiceInstanceName",
		User:         "test_user",
		Password:     "test_password",
	}

	t.Run("success ready true", func(t *testing.T) {
		//given
		data := map[string]interface{}{"ready": true}
		jsonData, _ := json.Marshal(data)
		m := &BtpExecutorMock{
			StdoutReturn: map[string]string{
				"btp .* get services/instance": string(jsonData),
			},
		}
		m.Stdout(new(bytes.Buffer))
		btp := NewBTPUtils(m)
		result := IsServiceInstanceCreated(btp, btpConfig)
		assert.True(t, result.done && result.successful)
	})

	t.Run("success ready false", func(t *testing.T) {
		//given
		data := map[string]interface{}{"ready": false}
		jsonData, _ := json.Marshal(data)
		m := &BtpExecutorMock{
			StdoutReturn: map[string]string{
				"btp .* get services/instance": string(jsonData),
			},
		}
		m.Stdout(new(bytes.Buffer))
		btp := NewBTPUtils(m)
		result := IsServiceInstanceCreated(btp, btpConfig)
		assert.False(t, result.done && result.successful)
	})

	t.Run("GetServiceInstance error", func(t *testing.T) {
		//given
		m := &BtpExecutorMock{
			ShouldFailOnCommand: map[string]error{
				"btp .* get services/instance": errors.New("not found"),
			},
		}
		m.Stdout(new(bytes.Buffer))
		btp := NewBTPUtils(m)
		result := IsServiceInstanceCreated(btp, btpConfig)
		assert.False(t, result.done && result.successful)
	})

	t.Run("unmarshal error", func(t *testing.T) {
		//given
		m := &BtpExecutorMock{
			StdoutReturn: map[string]string{
				"btp get services/instance": "not-json",
			},
		}
		m.Stdout(new(bytes.Buffer))
		btp := NewBTPUtils(m)
		result := IsServiceInstanceCreated(btp, btpConfig)
		assert.False(t, result.done && result.successful)
	})
}

func TestIsServiceInstanceDeleted(t *testing.T) {
	//given
	btpConfig := GetServiceInstanceOptions{
		Url:          "https://api.endpoint.com",
		Subdomain:    "xxxxxxx",
		Subaccount:   "yyyyyyy",
		InstanceName: "testServiceInstanceName",
		User:         "test_user",
		Password:     "test_password",
	}

	t.Run("instance still exists", func(t *testing.T) {
		m := &BtpExecutorMock{
			StdoutReturn: map[string]string{
				"btp .* get services/instance": "{}",
			},
		}
		m.Stdout(new(bytes.Buffer))
		btp := NewBTPUtils(m)
		result := IsServiceInstanceDeleted(btp, btpConfig)
		assert.False(t, result.done && result.successful)
	})

	t.Run("instance not found", func(t *testing.T) {
		//given
		m := &BtpExecutorMock{
			ShouldFailOnCommand: map[string]error{
				"btp .* get services/instance .+": errors.New(`
				{
				"error": "BadRequest",
				"description": "Could not find such instance"
				}`),
			},
		}
		m.Stdout(new(bytes.Buffer))
		btp := NewBTPUtils(m)
		result := IsServiceInstanceDeleted(btp, btpConfig)
		assert.True(t, result.done && result.successful)
	})
}

func TestIsServiceBindingCreated(t *testing.T) {
	//given
	btpConfig := GetServiceBindingOptions{
		Url:               "https://api.endpoint.com",
		Subdomain:         "xxxxxxx",
		Subaccount:        "yyyyyyy",
		BindingName:       "testServiceBindingName",
		User:              "test_user",
		Password:          "test_password",
		ServiceInstance:   "testServiceInstanceName",
		ServiceInstanceId: "xxx",
	}

	t.Run("success ready true", func(t *testing.T) {
		//given
		data := []map[string]interface{}{{"ready": true}}
		jsonData, _ := json.Marshal(data)

		m := &BtpExecutorMock{
			StdoutReturn: map[string]string{
				"btp .* list services/binding": string(jsonData),
			},
		}
		m.Stdout(new(bytes.Buffer))
		btp := NewBTPUtils(m)
		result := IsServiceBindingCreated(btp, btpConfig)
		assert.True(t, result.done && result.successful)
	})

	t.Run("success ready false", func(t *testing.T) {
		//given
		data := []map[string]interface{}{{"ready": false}}
		jsonData, _ := json.Marshal(data)

		m := &BtpExecutorMock{
			StdoutReturn: map[string]string{
				"btp .* list services/binding": string(jsonData),
			},
		}
		m.Stdout(new(bytes.Buffer))
		btp := NewBTPUtils(m)
		result := IsServiceBindingCreated(btp, btpConfig)
		assert.False(t, result.done && result.successful)
	})

	t.Run("GetServiceBinding error", func(t *testing.T) {
		//given
		m := &BtpExecutorMock{
			ShouldFailOnCommand: map[string]error{
				"btp .* list services/binding": errors.New("not found"),
			},
		}
		m.Stdout(new(bytes.Buffer))
		btp := NewBTPUtils(m)
		result := IsServiceBindingCreated(btp, btpConfig)
		assert.False(t, result.done && result.successful)
	})

	t.Run("unmarshal error", func(t *testing.T) {
		//given
		m := &BtpExecutorMock{
			StdoutReturn: map[string]string{
				"btp .* get services/binding": "not-json",
			},
		}
		m.Stdout(new(bytes.Buffer))
		btp := NewBTPUtils(m)
		result := IsServiceBindingCreated(btp, btpConfig)
		assert.False(t, result.done && result.successful)
	})
}

func TestIsServiceBindingDeleted(t *testing.T) {
	//given
	btpConfig := GetServiceBindingOptions{
		Url:         "https://api.endpoint.com",
		Subdomain:   "xxxxxxx",
		Subaccount:  "yyyyyyy",
		BindingName: "testServiceBindingName",
		User:        "test_user",
		Password:    "test_password",
	}

	t.Run("binding still exists", func(t *testing.T) {
		m := &BtpExecutorMock{
			StdoutReturn: map[string]string{
				"btp .* get services/binding": "{}",
			},
		}
		m.Stdout(new(bytes.Buffer))
		btp := NewBTPUtils(m)
		result := IsServiceBindingDeleted(btp, btpConfig)
		assert.False(t, result.done && result.successful)
	})

	t.Run("binding not found", func(t *testing.T) {
		m := &BtpExecutorMock{
			ShouldFailOnCommand: map[string]error{
				"btp .* get services/binding": errors.New(`
				{
				"error": "BadRequest",
				"description": "Could not find such binding"
				}`),
			},
		}
		m.Stdout(new(bytes.Buffer))
		btp := NewBTPUtils(m)
		result := IsServiceBindingDeleted(btp, btpConfig)
		assert.True(t, result.done && result.successful)
	})
}
