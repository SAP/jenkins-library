package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestRunAbapEnvironmentCreateSystem(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		// init
		config := abapEnvironmentCreateSystemOptions{}

		m := &mock.ExecMockRunner{}
		cf := cloudfoundry.CFUtils{Exec: m}

		// test
		err := runAbapEnvironmentCreateSystem(&config, nil, cf)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		// init
		config := abapEnvironmentCreateSystemOptions{}

		m := &mock.ExecMockRunner{}
		cf := cloudfoundry.CFUtils{Exec: m}

		// test
		err := runAbapEnvironmentCreateSystem(&config, nil, cf)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
