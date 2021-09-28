package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Field1 int    `validate:"eq=1"`
	Field2 string `validate:"oneof=value1 value2 value3"`
	Field3 string `validate:"required_if=Field1 1"`
}

type testStructWithJSONTags struct {
	Field1 int    `json:"field1,omitempty" validate:"eq=1"`
	Field2 string `json:"field2,omitempty" validate:"oneof=value1 value2 value3"`
	Field3 string `json:"field3,omitempty" validate:"required_if=Field1 1"`
}

func TestValidateStruct(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		validation, err := New()
		assert.NoError(t, err)
		tStruct := testStruct{
			Field1: 1,
			Field2: "value1",
			Field3: "field3",
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
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "The Field2 must use the following values: value1 value2 value3.")
		assert.Contains(t, err.Error(), "The Field3 is required since the Field1 is 1.")
	})

	t.Run("failed case - custom error message", func(t *testing.T) {
		validation, err := New()
		assert.NoError(t, err)
		tStruct := testStructWithJSONTags{
			Field1: 1,
			Field2: "value4",
		}
		err = validation.ValidateStruct(tStruct)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "The field2 must use the following values: value1 value2 value3.")
		assert.Contains(t, err.Error(), "The field3 is required since the Field1 is 1.")
	})
}
