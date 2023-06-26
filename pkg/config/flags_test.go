//go:build unit
// +build unit

package config

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestAvailableFlagValues(t *testing.T) {
	var f StepFilters

	var test0 string
	var test1 string
	var test2 []string
	var test3 bool

	var c = &cobra.Command{
		Use:   "test",
		Short: "..",
	}

	c.Flags().StringVar(&test0, "test0", "val0", "Test 0")
	c.Flags().StringVar(&test1, "test1", "", "Test 1")
	c.Flags().StringSliceVar(&test2, "test2", []string{}, "Test 2")
	c.Flags().BoolVar(&test3, "test3", false, "Test 3")

	c.Flags().Set("test1", "val1")
	c.Flags().Set("test2", "val3_1")
	c.Flags().Set("test3", "true")

	v := AvailableFlagValues(c, &f)

	if v["test0"] != nil {
		t.Errorf("expected: 'test0' to be empty but was %v", v["test0"])
	}

	assert.Equal(t, "val1", v["test1"])
	assert.Equal(t, []string{"val3_1"}, v["test2"])
	assert.Equal(t, true, v["test3"])

}

func TestMarkFlagsWithValue(t *testing.T) {
	var test0 string
	var test1 string
	var test2 string
	var c = &cobra.Command{
		Use:   "test",
		Short: "..",
	}
	c.Flags().StringVar(&test0, "test0", "val0", "Test 0")
	c.Flags().StringVar(&test1, "test1", "", "Test 1")
	c.Flags().StringVar(&test2, "test2", "", "Test 2")

	s := StepConfig{
		Config: map[string]interface{}{
			"test2": "val2",
		},
	}

	MarkFlagsWithValue(c, s)

	assert.Equal(t, true, c.Flags().Changed("test0"), "default not considered")
	assert.Equal(t, false, c.Flags().Changed("test1"), "no value: considered as set")
	assert.Equal(t, true, c.Flags().Changed("test2"), "config not considered")
}
