//go:build unit
// +build unit

package piperutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeys(t *testing.T) {
	intStringMap := map[int]string{1: "eins", 2: "zwei", 3: "drei", 4: "vier"}

	intList := Keys(intStringMap)

	assert.Equal(t, 4, len(intList))
	assert.Equal(t, true, ContainsInt(intList, 1))
	assert.Equal(t, true, ContainsInt(intList, 2))
	assert.Equal(t, true, ContainsInt(intList, 3))
	assert.Equal(t, true, ContainsInt(intList, 4))
}

func TestValues(t *testing.T) {
	intStringMap := map[int]string{1: "eins", 2: "zwei", 3: "drei", 4: "vier"}

	intList := Values(intStringMap)

	assert.Equal(t, 4, len(intList))
	assert.Equal(t, true, ContainsString(intList, "eins"))
	assert.Equal(t, true, ContainsString(intList, "zwei"))
	assert.Equal(t, true, ContainsString(intList, "drei"))
	assert.Equal(t, true, ContainsString(intList, "vier"))
}
