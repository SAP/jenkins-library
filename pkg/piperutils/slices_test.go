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

func TestContainsString(t *testing.T) {
	var stringList []string
	assert.False(t, ContainsString(stringList, "test"), "False expected but returned true")
	assert.False(t, ContainsString(stringList, ""), "False expected but returned true")

	stringList = append(stringList, "", "foo", "bar", "foo")
	assert.True(t, ContainsString(stringList, ""), "True expected but returned false")
	assert.True(t, ContainsString(stringList, "bar"), "True expected but returned false")
	assert.True(t, ContainsString(stringList, "foo"), "True expected but returned false")
	assert.False(t, ContainsString(stringList, "baz"), "False expected but returned true")
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

func TestPrefixIfNeeded(t *testing.T) {
	// init
	s := []string{"tree", "pie", "applejuice"}
	// test
	s = PrefixIfNeeded(s, "apple")
	// assert
	assert.Contains(t, s, "appletree")
	assert.Contains(t, s, "applepie")
	assert.Contains(t, s, "applejuice")
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

func TestSplitTrimAndDeDup(t *testing.T) {
	t.Run("Separator is not space", func(t *testing.T) {
		// init
		s := []string{" a", "", "-a-b --c ", "d-e", "f", " f", ""}
		// test
		s = SplitTrimAndDeDup(s, "-")
		// assert
		assert.Equal(t, []string{"a", "b", "c", "d", "e", "f"}, s)
	})
	t.Run("Separator is space", func(t *testing.T) {
		// init
		s := []string{" a", " a b  c ", "d e", "f", "f ", ""}
		// test
		s = SplitTrimAndDeDup(s, " ")
		// assert
		assert.Equal(t, []string{"a", "b", "c", "d", "e", "f"}, s)
	})
	t.Run("Separator is multi-char", func(t *testing.T) {
		// init
		s := []string{" a", " a** b**c ", "**d **e", "f**", "f ", ""}
		// test
		s = SplitTrimAndDeDup(s, "**")
		// assert
		assert.Equal(t, []string{"a", "b", "c", "d", "e", "f"}, s)
	})
	t.Run("Separator is empty string", func(t *testing.T) {
		// init
		s := []string{" a", " a bc ", "d e", "f", "f ", ""}
		// test
		s = SplitTrimAndDeDup(s, "")
		// assert
		// If "sep" is empty, underlying strings.Split() splits after each UTF-8 char sequence.
		assert.Equal(t, []string{"a", "b", "c", "d", "e", "f"}, s)
	})
}
