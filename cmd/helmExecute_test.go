package cmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunHelmExecute(t *testing.T) {
	t.Parallel()

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init

		// test
		err := fmt.Errorf("error")

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
