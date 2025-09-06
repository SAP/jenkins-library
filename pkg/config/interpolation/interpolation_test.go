package interpolation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveMap(t *testing.T) {
	t.Parallel()

	t.Run("That lookup works", func(t *testing.T) {
		testMap := map[string]interface{}{
			"prop1": "val1",
			"prop2": "val2",
			"prop3": "$(prop1)/$(prop2)",
		}

		ok := ResolveMap(testMap)
		assert.True(t, ok)

		assert.Equal(t, "val1/val2", testMap["prop3"])
	})

	t.Run("That lookups fails when property is not found", func(t *testing.T) {
		testMap := map[string]interface{}{
			"prop1": "val1",
			"prop2": "val2",
			"prop3": "$(prop1)/$(prop2)/$(prop5)",
		}

		ok := ResolveMap(testMap)
		assert.False(t, ok)
	})

	t.Run("That resolve loops are aborted", func(t *testing.T) {
		testMap := map[string]interface{}{
			"prop1": "$(prop2)",
			"prop2": "$(prop1)",
		}
		ok := ResolveMap(testMap)
		assert.False(t, ok)
	})

}
