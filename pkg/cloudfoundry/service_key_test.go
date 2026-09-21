//go:build unit
// +build unit

package cloudfoundry

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestReadServiceKey(t *testing.T) {
	const serviceKey = "header\nheader\n{\"url\":\"https://example.test\"}"

	runner := &mock.ExecMockRunner{
		StdoutReturn: map[string]string{"cf service-key test-instance test-key": serviceKey},
	}
	key, err := ReadServiceKey(runner, ServiceKeyOptions{
		CfAPIEndpoint:     "https://api.example.test",
		CfOrg:             "test-org",
		CfSpace:           "test-space",
		CfServiceInstance: "test-instance",
		CfServiceKeyName:  "test-key",
		Username:          "test-user",
		Password:          "test-password",
	})

	if assert.NoError(t, err) {
		assert.Equal(t, `{"url":"https://example.test"}`, key)
		assert.Equal(t, []mock.ExecCall{
			{Exec: "cf", Params: []string{"login", "-a", "https://api.example.test", "-o", "test-org", "-s", "test-space", "-u", "test-user", "-p", "test-password"}},
			{Exec: "cf", Params: []string{"service-key", "test-instance", "test-key"}},
			{Exec: "cf", Params: []string{"logout"}},
		}, runner.Calls)
	}
}
