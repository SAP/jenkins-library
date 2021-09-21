package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Field1 int    `validate:"eq=1"`
	Field2 string `validate:"oneof=value1 value2 value3"`
}

type testStructWithJSONTags struct {
	Field1 int    `json:"field1,omitempty" validate:"eq=1"`
	Field2 string `json:"field2,omitempty" validate:"oneof=value1 value2 value3"`
}

func TestValidateStruct(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		validation, err := New()
		assert.NoError(t, err)
		tStruct := testStruct{
			Field1: 1,
			Field2: "value1",
		}
		err = validation.ValidateStruct(tStruct)
		assert.NoError(t, err)
	})

	t.Run("failed case - custom error message", func(t *testing.T) {
		validation, err := New()
		assert.NoError(t, err)
		tStruct := testStruct{
			Field1: 1,
			Field2: "value4",
		}
		err = validation.ValidateStruct(tStruct)
		assert.Contains(t, err.Error(), "Field2 must use the folowing values: value1 value2 value3.")
	})

	t.Run("failed case - custom error message with json tag name", func(t *testing.T) {
		validation, err := New()
		assert.NoError(t, err)
		tStruct := testStructWithJSONTags{
			Field1: 1,
			Field2: "value4",
		}
		err = validation.ValidateStruct(tStruct)
		assert.Contains(t, err.Error(), "field2 must use the folowing values: value1 value2 value3.")
	})
}
