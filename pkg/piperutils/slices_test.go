package piperutils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContainsInt(t *testing.T) {
	var intList []int
	assert.Equal(t, false, ContainsInt(intList, 4), "False expected but returned true")

	intList = append(intList, 1, 2, 3, 4, 5, 6, 20)
	assert.Equal(t, true, ContainsInt(intList, 20), "True expected but returned false")
	assert.Equal(t, true, ContainsInt(intList, 1), "True expected but returned false")
	assert.Equal(t, true, ContainsInt(intList, 4), "True expected but returned false")
	assert.Equal(t, false, ContainsInt(intList, 13), "False expected but returned true")
}
