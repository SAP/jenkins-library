package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryCreateSpace(t *testing.T) {
	m := &mock.ExecMockRunner{}
	s := mock.ShellMockRunner{}

	cf := cloudfoundry.CFUtils{Exec: m}
	var telemetryData telemetry.CustomData

	config := cloudFoundryCreateSpaceOptions{
		CfAPIEndpoint: "https://api.endpoint.com",
		CfOrg:         "testOrg",
		CfSpace:       "testSpace",
		Username:      "testUser",
		Password:      "testPassword",
	}

	// Business code has moved to pkg/cloudfoundry/CreateSpace.go and is tested in that package
	t.Run("happy path", func(t *testing.T) {
		// test
		err := runCloudFoundryCreateSpace(&config, &telemetryData, cf, &s)

		// assert
		assert.NoError(t, err)
	})
}
