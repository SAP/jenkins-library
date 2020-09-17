package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryDeleteSpace(t *testing.T) {
	m := &mock.ExecMockRunner{}
	cf := cloudfoundry.CFUtils{Exec: m}
	cfUtilsMock := cloudfoundry.CfUtilsMock{}
	var telemetryData telemetry.CustomData

	config := cloudFoundryDeleteSpaceOptions{
		CfAPIEndpoint: "https://api.endpoint.com",
		CfOrg:         "testOrg",
		CfSpace:       "testSpace",
		Username:      "testUser",
		Password:      "testPassword",
	}

	t.Run("CF Delete Space : success case", func(t *testing.T) {

		err := runCloudFoundryDeleteSpace(&config, &telemetryData, cf)
		if assert.NoError(t, err) {
			assert.Equal(t, "cf", m.Calls[0].Exec)
			assert.Equal(t, []string{"delete-space", "testSpace", "-o", "testOrg", "-f"}, m.Calls[1].Params)
		}
	})

	t.Run("CF Login Error", func(t *testing.T) {
		//Needs to be fixed
		defer cfUtilsMock.Cleanup()
		errorMessage := "cf login error"

		cfUtilsMock.LoginError = errors.New(errorMessage)

		gotError := runCloudFoundryDeleteSpace(&config, &telemetryData, cf)
		assert.EqualError(t, gotError, "Error while logging in occured: "+errorMessage, "Wrong error message")
	})
}
