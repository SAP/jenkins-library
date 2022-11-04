package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateContract(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T){
		comparisonMap := &ComparisonMap{
			Topics: map[string]map[string]*Field{
				"topic1": {
					"field_timestamp": &Field{
						MatchingRule: "timestamp",
						PayloadValue: "2022-01-01",
					},
					"field_date": &Field{
						MatchingRule: "date",
					},
					"field_decimal": &Field{
						MatchingRule: "decimal",
					},
				},
				"topic2": {
					"field_integer": &Field{
						MatchingRule: "integer",
					},
					"field_number1": &Field{
						MatchingRule: "number",
					},
					"field_number2": &Field{
						MatchingRule: "number",
					},
					"field_type": &Field{
						MatchingRule: "type",
						PayloadValue: "someValue",
					},
				},
			},
		}
		pactMap := &ComparisonMap{
			Topics: map[string]map[string]*Field{
				"topic1": {
					"field_timestamp": &Field{
						MatchingRule: "timestamp",
						PayloadValue: "2022-01-01",
					},
					"field_date": &Field{
						MatchingRule: "string",
						Format: "date-time",
					},
					"field_decimal": &Field{
						MatchingRule: "number",
					},
				},
				"topic2": {
					"field_integer": &Field{
						MatchingRule: "integer",
					},
					"field_number1": &Field{
						MatchingRule: "number",
					},
					"field_number2": &Field{
						MatchingRule: "integer",
					},
					"field_type": &Field{
						PayloadValue: "someOtherValue",
					},
				},
			},
		}
	
		expectedResult := &FailureData{}
	
	
		res, ok := comparisonMap.ValidateContract(pactMap)
		assert.True(t, ok)
		assert.Equal(t, expectedResult, res)
	})

	/*

	t.Run("failure", func(t *testing.T){
		comparisonMap := &ComparisonMap{}
		pactMap := &ComparisonMap{}
	
		expectedResult := &FailureData{}
	
	
		res, ok := comparisonMap.ValidateContract(pactMap)
		assert.False(t, ok)
		assert.Equal(t, expectedResult, res)
	})

	*/
}

func TestValidateContracts(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T){
		comparisonMap := &ComparisonMap{}
		pactMap := map[string]*ComparisonMap{}

		ok, err := comparisonMap.ValidateContracts(pactMap)
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("failed contracts", func(t *testing.T){
		comparisonMap := &ComparisonMap{}
		pactMap := map[string]*ComparisonMap{}

		ok, err := comparisonMap.ValidateContracts(pactMap)
		assert.NoError(t, err)
		assert.True(t, ok)
	})
}