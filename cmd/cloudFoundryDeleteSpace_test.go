package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryDeleteSpace(t *testing.T) {
	m := &mock.ExecMockRunner{}
	s := mock.ShellMockRunner{}
	cf := cloudfoundry.CFUtils{Exec: m}
	var telemetryData telemetry.CustomData

	config := cloudFoundryDeleteSpaceOptions{
		CfAPIEndpoint: "https://api.endpoint.com",
		CfOrg:         "testOrg",
		CfSpace:       "testSpace",
		Username:      "testUser",
		Password:      "testPassword",
	}

	t.Run("CF Delete Space : success case", func(t *testing.T) {

		err := runCloudFoundryDeleteSpace(&config, &telemetryData, cf, &s)
		if assert.NoError(t, err) {
			assert.Equal(t, "cf", m.Calls[0].Exec)
			assert.Equal(t, []string{"delete-space", "testSpace", "-o", "testOrg", "-f"}, m.Calls[0].Params)
		}
	})
}
