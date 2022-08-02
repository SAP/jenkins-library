package piperutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValues(t *testing.T) {
	intStringMap := map[int]string{1: "eins", 2: "zwei", 3: "drei", 4: "vier"}

	intList := Values(intStringMap)

	assert.Equal(t, 4, len(intList))
	assert.Equal(t, true, ContainsString(intList, "eins"))
	assert.Equal(t, true, ContainsString(intList, "zwei"))
	assert.Equal(t, true, ContainsString(intList, "drei"))
	assert.Equal(t, true, ContainsString(intList, "vier"))
}
