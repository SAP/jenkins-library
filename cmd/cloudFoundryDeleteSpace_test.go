package cmd

import (
	"fmt"
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

	t.Run("CF login: Success", func(t *testing.T) {

		err := runCloudFoundryDeleteSpace(&config, &telemetryData, cf, &s)
		if assert.NoError(t, err) {
			assert.Contains(t, s.Calls[0], "yes '' | cf login -a https://api.endpoint.com -u testUser -p testPassword")
		}
	})

	t.Run("CF login: failure case", func(t *testing.T) {

		errorMessage := "cf login failed"

		defer func() {
			s.Calls = nil
			s.ShouldFailOnCommand = nil
		}()

		s.ShouldFailOnCommand = map[string]error{"yes '' | cf login -a https://api.endpoint.com -u testUser -p testPassword ": fmt.Errorf("%s", errorMessage)}

		e := runCloudFoundryDeleteSpace(&config, &telemetryData, cf, &s)
		assert.EqualError(t, e, "Error while logging in occured: "+errorMessage)
	})

	t.Run("CF Delete Space : success case", func(t *testing.T) {

		err := runCloudFoundryDeleteSpace(&config, &telemetryData, cf, &s)
		if assert.NoError(t, err) {
			assert.Equal(t, "cf", m.Calls[0].Exec)
			assert.Equal(t, []string{"delete-space", "testSpace", "-o", "testOrg", "-f"}, m.Calls[0].Params)
		}
	})

	t.Run("CF Delete space: failure case", func(t *testing.T) {

		errorMessage := "cf space creation error"

		m.ShouldFailOnCommand = map[string]error{"cf delete-space testSpace -o testOrg -f": fmt.Errorf("%s", errorMessage)}

		gotError := runCloudFoundryDeleteSpace(&config, &telemetryData, cf, &s)
		assert.EqualError(t, gotError, "Deletion of cf space has failed: "+errorMessage, "Wrong error message")
	})
}
