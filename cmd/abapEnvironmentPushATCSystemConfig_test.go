package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/stretchr/testify/assert"
)

func TestRunAbapEnvironmentPushATCSystemConfig(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := abapEnvironmentPushATCSystemConfigOptions{}

		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()

		// test
		err := runAbapEnvironmentPushATCSystemConfig(&config, nil, &autils, nil)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := abapEnvironmentPushATCSystemConfigOptions{}

		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()

		// test
		err := runAbapEnvironmentPushATCSystemConfig(&config, nil, &autils, nil)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
