package piperutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsInt(t *testing.T) {
	var intList []int
	assert.Equal(t, false, ContainsInt(intList, 4))

	intList = append(intList, 1, 2, 3, 4, 5, 6, 20)
	assert.Equal(t, true, ContainsInt(intList, 20))
	assert.Equal(t, true, ContainsInt(intList, 1))
	assert.Equal(t, true, ContainsInt(intList, 4))
	assert.Equal(t, false, ContainsInt(intList, 13))
}

func TestContainsString(t *testing.T) {
	var stringList []string
	assert.False(t, ContainsString(stringList, "test"))
	assert.False(t, ContainsString(stringList, ""))

	stringList = append(stringList, "", "foo", "bar", "foo")
	assert.True(t, ContainsString(stringList, ""))
	assert.True(t, ContainsString(stringList, "bar"))
	assert.True(t, ContainsString(stringList, "foo"))
	assert.False(t, ContainsString(stringList, "baz"))
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
