//go:build unit

package piperutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveAll(t *testing.T) {
	t.Parallel()
	t.Run("empty array", func(t *testing.T) {
		result, removed := RemoveAll([]string{}, "A")
		assert.Len(t, result, 0)
		assert.False(t, removed)
	})
	t.Run("two As", func(t *testing.T) {
		result, removed := RemoveAll([]string{"A", "B", "C", "A", "C", "", "D"}, "A")
		assert.Equal(t, []string{"B", "C", "C", "", "D"}, result)
		assert.True(t, removed)
	})
	t.Run("one B", func(t *testing.T) {
		result, removed := RemoveAll([]string{"A", "B", "C", "A", "C", "", "D"}, "B")
		assert.Equal(t, []string{"A", "C", "A", "C", "", "D"}, result)
		assert.True(t, removed)
	})
	t.Run("empty e", func(t *testing.T) {
		result, removed := RemoveAll([]string{"A", "B", "C", "A", "C", "", "D"}, "")
		assert.Equal(t, []string{"A", "B", "C", "A", "C", "D"}, result)
		assert.True(t, removed)
	})
	t.Run("one D", func(t *testing.T) {
		result, removed := RemoveAll([]string{"A", "B", "C", "A", "C", "", "D"}, "D")
		assert.Equal(t, []string{"A", "B", "C", "A", "C", ""}, result)
		assert.True(t, removed)
	})
	t.Run("not found", func(t *testing.T) {
		result, removed := RemoveAll([]string{"A", "B", "C", "A", "C", "", "D"}, "X")
		assert.Equal(t, []string{"A", "B", "C", "A", "C", "", "D"}, result)
		assert.False(t, removed)
	})
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
		s = SplitAndTrim(s, "-")
		// assert
		assert.Equal(t, []string{"a", "a", "b", "c", "d", "e", "f", "f"}, s)
	})
	t.Run("Separator is space", func(t *testing.T) {
		// init
		s := []string{" a", " a b  c ", "d e", "f", "f ", ""}
		// test
		s = SplitAndTrim(s, " ")
		// assert
		assert.Equal(t, []string{"a", "a", "b", "c", "d", "e", "f", "f"}, s)
	})
	t.Run("Separator is multi-char", func(t *testing.T) {
		// init
		s := []string{" a", " a** b**c ", "**d **e", "f**", "f ", ""}
		// test
		s = SplitAndTrim(s, "**")
		// assert
		assert.Equal(t, []string{"a", "a", "b", "c", "d", "e", "f", "f"}, s)
	})
	t.Run("Separator is empty string", func(t *testing.T) {
		// init
		s := []string{" a", " a bc ", "d e", "f", "f ", ""}
		// test
		s = SplitAndTrim(s, "")
		// assert
		// If "sep" is empty, underlying strings.Split() splits after each UTF-8 char sequence.
		assert.Equal(t, []string{"a", "a", "b", "c", "d", "e", "f", "f"}, s)
	})
}

func TestUniqueStrings(t *testing.T) {

	unique := UniqueStrings([]string{"abc", "xyz", "123", "abc"})
	if assert.Len(t, unique, 3) {
		assert.Subset(t, []string{"123", "abc", "xyz"}, unique)
	}
}

func TestCopyAtoB(t *testing.T) {
	src := []string{"abc", "xyz", "123", "abc"}
	target := make([]string, 4)
	CopyAtoB(src, target)
	if assert.Len(t, target, 4) {
		assert.EqualValues(t, src, target)
	}
}
