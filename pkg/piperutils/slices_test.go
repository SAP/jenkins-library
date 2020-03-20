package piperutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestPrefix(t *testing.T) {
	// init
	s := []string{"tree", "pie", "applejuice"}
	// test
	s = Prefix(s, "apple")
	// assert
	assert.Contains(t, s, "appletree")
	assert.Contains(t, s, "applepie")
	assert.Contains(t, s, "appleapplejuice")
}

func TestTrim(t *testing.T) {
	// init
	s := []string{" orange", "banana ", "	apple", "mango	", " ", ""}
	// test
	s = Trim(s)
	// assert
	assert.Equal(t, 4, len(s))
	assert.Contains(t, s, "orange")
	assert.Contains(t, s, "banana")
	assert.Contains(t, s, "apple")
	assert.Contains(t, s, "mango")
}
