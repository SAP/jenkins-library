package interpolation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveMap(t *testing.T) {
	t.Parallel()

	t.Run("Lookup lookup works", func(t *testing.T) {
		testMap := map[string]interface{}{
			"prop1": "val1",
			"prop2": "val2",
			"prop3": "$(prop1)/$(prop2)",
		}

		err := ResolveMap(testMap)
		assert.NoError(t, err)

		assert.Equal(t, "val1/val2", testMap["prop3"])
	})

	t.Run("That resolve loops are aborted", func(t *testing.T) {
		testMap := map[string]interface{}{
			"prop1": "$(prop2)",
			"prop2": "$(prop1)",
		}
		err := ResolveMap(testMap)
		assert.Error(t, err)
	})

}
