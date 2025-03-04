//go:build unit
// +build unit

package cmd

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryCreateSpace(t *testing.T) {
	m := &mock.ExecMockRunner{}
	s := mock.ShellMockRunner{}

	cf := cloudfoundry.CFUtils{Exec: m}
	cfUtilsMock := cloudfoundry.CfUtilsMock{}
	var telemetryData telemetry.CustomData

	config := cloudFoundryCreateSpaceOptions{
		CfAPIEndpoint: "https://api.endpoint.com",
		CfOrg:         "testOrg",
		CfSpace:       "testSpace",
		Username:      "testUser",
		Password:      "testPassword",
	}

	t.Run("CF login: Success", func(t *testing.T) {

		err := runCloudFoundryCreateSpace(&config, &telemetryData, cf, &s)
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

		e := runCloudFoundryCreateSpace(&config, &telemetryData, cf, &s)
		assert.EqualError(t, e, "Error while logging in occured: "+errorMessage)
	})

	t.Run("CF space creation: Success", func(t *testing.T) {

		err := runCloudFoundryCreateSpace(&config, &telemetryData, cf, &s)
		if assert.NoError(t, err) {
			assert.Equal(t, "cf", m.Calls[0].Exec)
			assert.Equal(t, []string{"create-space", "testSpace", "-o", "testOrg"}, m.Calls[0].Params)
		}
	})

	t.Run("CF space creation: FAILURE", func(t *testing.T) {

		defer cfUtilsMock.Cleanup()
		errorMessage := "cf space creation error"

		m.ShouldFailOnCommand = map[string]error{"cf create-space testSpace -o testOrg": fmt.Errorf("%s", errorMessage)}

		cfUtilsMock.LoginError = errors.New(errorMessage)

		gotError := runCloudFoundryCreateSpace(&config, &telemetryData, cf, &s)
		assert.EqualError(t, gotError, "Creating a cf space has failed: "+errorMessage, "Wrong error message")
	})
}
