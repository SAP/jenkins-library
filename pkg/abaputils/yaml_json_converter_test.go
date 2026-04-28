//go:build unit
// +build unit

package abaputils

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYAMLToJSON(t *testing.T) {
	t.Run("simple key-value pairs", func(t *testing.T) {
		yaml := []byte("key: value\nnum: 42\n")
		result, err := YAMLToJSON(yaml)
		require.NoError(t, err)
		var got map[string]interface{}
		require.NoError(t, json.Unmarshal(result, &got))
		assert.Equal(t, "value", got["key"])
		assert.Equal(t, float64(42), got["num"])
	})

	t.Run("nested map", func(t *testing.T) {
		yaml := []byte("outer:\n  inner: hello\n")
		result, err := YAMLToJSON(yaml)
		require.NoError(t, err)
		var got map[string]interface{}
		require.NoError(t, json.Unmarshal(result, &got))
		outer, ok := got["outer"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "hello", outer["inner"])
	})

	t.Run("array value", func(t *testing.T) {
		yaml := []byte("items:\n  - a\n  - b\n  - c\n")
		result, err := YAMLToJSON(yaml)
		require.NoError(t, err)
		var got map[string]interface{}
		require.NoError(t, json.Unmarshal(result, &got))
		items, ok := got["items"].([]interface{})
		require.True(t, ok)
		assert.Equal(t, []interface{}{"a", "b", "c"}, items)
	})

	t.Run("boolean values", func(t *testing.T) {
		yaml := []byte("enabled: true\ndisabled: false\n")
		result, err := YAMLToJSON(yaml)
		require.NoError(t, err)
		var got map[string]interface{}
		require.NoError(t, json.Unmarshal(result, &got))
		assert.Equal(t, true, got["enabled"])
		assert.Equal(t, false, got["disabled"])
	})

	t.Run("empty YAML produces null JSON", func(t *testing.T) {
		result, err := YAMLToJSON([]byte(""))
		require.NoError(t, err)
		assert.Equal(t, []byte("null"), result)
	})

	t.Run("invalid YAML returns error", func(t *testing.T) {
		_, err := YAMLToJSON([]byte("key: [\ninvalid"))
		assert.Error(t, err)
	})

	t.Run("integer keys become string keys in JSON", func(t *testing.T) {
		yaml := []byte("1: one\n2: two\n")
		result, err := YAMLToJSON(yaml)
		require.NoError(t, err)
		var got map[string]interface{}
		require.NoError(t, json.Unmarshal(result, &got))
		assert.Equal(t, "one", got["1"])
		assert.Equal(t, "two", got["2"])
	})
}

func TestYamlKeyToString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
		wantErr  bool
	}{
		{"string key", "hello", "hello", false},
		{"int key", int(42), "42", false},
		{"int64 key", int64(100), "100", false},
		{"float64 key", float64(3.14), "3.14", false},
		{"bool true", true, "true", false},
		{"bool false", false, "false", false},
		{"unsupported type", []int{1}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := yamlKeyToString(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestYamlPrimitiveToString(t *testing.T) {
	tests := []struct {
		name   string
		input  interface{}
		want   string
		wantOK bool
	}{
		{"int", int(7), "7", true},
		{"int64", int64(64), "64", true},
		{"float64", float64(1.5), "1.5", true},
		{"uint64", uint64(99), "99", true},
		{"bool true", true, "true", true},
		{"bool false", false, "false", true},
		{"string (not converted)", "hello", "", false},
		{"nil (not converted)", nil, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := yamlPrimitiveToString(tt.input)
			assert.Equal(t, tt.wantOK, ok)
			if tt.wantOK {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTagOptionsContains(t *testing.T) {
	tests := []struct {
		opts   tagOptions
		search string
		want   bool
	}{
		{"omitempty", "omitempty", true},
		{"omitempty,string", "omitempty", true},
		{"omitempty,string", "string", true},
		{"omitempty,string", "missing", false},
		{"", "omitempty", false},
		{"other", "omitempty", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.opts)+"/"+tt.search, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.opts.Contains(tt.search))
		})
	}
}

func TestParseTag(t *testing.T) {
	tests := []struct {
		tag      string
		wantName string
		wantOpts tagOptions
	}{
		{"name,omitempty", "name", "omitempty"},
		{"name", "name", ""},
		{",omitempty", "", "omitempty"},
		{"", "", ""},
		{"field,omitempty,string", "field", "omitempty,string"},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			name, opts := parseTag(tt.tag)
			assert.Equal(t, tt.wantName, name)
			assert.Equal(t, tt.wantOpts, opts)
		})
	}
}

func TestIsValidTag(t *testing.T) {
	tests := []struct {
		tag  string
		want bool
	}{
		{"valid", true},
		{"also_valid", true},
		{"with-dash", true},
		{"with.dot", true},
		{"", false},
		{"has\x00null", false},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			assert.Equal(t, tt.want, isValidTag(tt.tag))
		})
	}
}

func TestFoldFunc(t *testing.T) {
	t.Run("simpleLetterEqualFold for plain letters", func(t *testing.T) {
		fn := foldFunc([]byte("abc"))
		assert.True(t, fn([]byte("abc"), []byte("ABC")))
		assert.True(t, fn([]byte("ABC"), []byte("abc")))
		assert.False(t, fn([]byte("abc"), []byte("xyz")))
	})

	t.Run("asciiEqualFold for non-letter ascii", func(t *testing.T) {
		fn := foldFunc([]byte("a1b"))
		assert.True(t, fn([]byte("a1b"), []byte("A1B")))
		assert.False(t, fn([]byte("a1b"), []byte("a2b")))
	})

	t.Run("equalFoldRight for K/S special cases", func(t *testing.T) {
		fn := foldFunc([]byte("sky"))
		assert.NotNil(t, fn)
		assert.True(t, fn([]byte("sky"), []byte("SKY")))
	})

	t.Run("bytes.EqualFold for multibyte", func(t *testing.T) {
		fn := foldFunc([]byte("caf\xc3\xa9")) // "café"
		assert.NotNil(t, fn)
		assert.True(t, fn([]byte("caf\xc3\xa9"), []byte("caf\xc3\xa9")))
	})
}

func TestSimpleLetterEqualFold(t *testing.T) {
	assert.True(t, simpleLetterEqualFold([]byte("abc"), []byte("ABC")))
	assert.True(t, simpleLetterEqualFold([]byte("ABC"), []byte("abc")))
	assert.False(t, simpleLetterEqualFold([]byte("abc"), []byte("ab")))
	assert.False(t, simpleLetterEqualFold([]byte("abc"), []byte("xyz")))
}

func TestAsciiEqualFold(t *testing.T) {
	assert.True(t, asciiEqualFold([]byte("Hello1"), []byte("hello1")))
	assert.False(t, asciiEqualFold([]byte("abc"), []byte("ab")))
	assert.False(t, asciiEqualFold([]byte("a1b"), []byte("a2b")))
}

func TestEqualFoldRight(t *testing.T) {
	assert.True(t, equalFoldRight([]byte("abc"), []byte("ABC")))
	assert.False(t, equalFoldRight([]byte("abc"), []byte("ab")))
	assert.False(t, equalFoldRight([]byte("abc"), []byte("xyz")))
}

func TestDominantField(t *testing.T) {
	t.Run("single field is dominant", func(t *testing.T) {
		fields := []field{{name: "X", index: []int{0}}}
		got, ok := dominantField(fields)
		assert.True(t, ok)
		assert.Equal(t, "X", got.name)
	})

	t.Run("tagged field wins over untagged", func(t *testing.T) {
		fields := []field{
			{name: "X", index: []int{0}, tag: false},
			{name: "X", index: []int{1}, tag: true},
		}
		got, ok := dominantField(fields)
		assert.True(t, ok)
		assert.True(t, got.tag)
	})

	t.Run("two tagged fields at same depth: ambiguous", func(t *testing.T) {
		fields := []field{
			{name: "X", index: []int{0}, tag: true},
			{name: "X", index: []int{1}, tag: true},
		}
		_, ok := dominantField(fields)
		assert.False(t, ok)
	})

	t.Run("deeper field is excluded", func(t *testing.T) {
		fields := []field{
			{name: "X", index: []int{0}},
			{name: "X", index: []int{1, 2}},
		}
		got, ok := dominantField(fields)
		assert.True(t, ok)
		assert.Equal(t, []int{0}, got.index)
	})
}

func TestByNameSort(t *testing.T) {
	fields := []field{
		{name: "Z", index: []int{2}},
		{name: "A", index: []int{0}},
		{name: "M", index: []int{1}},
	}
	sorted := byName(fields)

	assert.Equal(t, 3, sorted.Len())

	// "A" (index 1) < "Z" (index 0) by name
	assert.True(t, sorted.Less(1, 0))
	// "Z" (index 0) is not less than "A" (index 1)
	assert.False(t, sorted.Less(0, 1))

	// Swap positions 0 and 1: fields becomes [A, Z, M]
	sorted.Swap(0, 1)
	assert.Equal(t, "A", fields[0].name)
	assert.Equal(t, "Z", fields[1].name)
	assert.Equal(t, "M", fields[2].name)
}

func TestByIndexSort(t *testing.T) {
	fields := []field{
		{name: "B", index: []int{1}},
		{name: "A", index: []int{0}},
	}
	bi := byIndex(fields)
	assert.Equal(t, 2, bi.Len())
	assert.True(t, bi.Less(1, 0)) // index {0} < {1}
	bi.Swap(0, 1)
	assert.Equal(t, "A", fields[0].name)
}

func TestCachedTypeFields(t *testing.T) {
	type Sample struct {
		Name  string `json:"name"`
		Value int    `json:"value,omitempty"`
	}

	typ := reflect.TypeOf(Sample{})
	fields := cachedTypeFields(typ)
	assert.Len(t, fields, 2)

	names := []string{fields[0].name, fields[1].name}
	assert.Contains(t, names, "name")
	assert.Contains(t, names, "value")

	// Second call should hit cache and return the same result.
	fields2 := cachedTypeFields(typ)
	assert.Equal(t, len(fields), len(fields2))
}

func TestTypeFields_OmitEmptyAndQuoted(t *testing.T) {
	type T struct {
		A string `json:"a,omitempty"`
		B int    `json:"b,string"`
	}
	fields := typeFields(reflect.TypeOf(T{}))
	require.Len(t, fields, 2)
	for _, f := range fields {
		switch f.name {
		case "a":
			assert.True(t, f.omitEmpty)
		case "b":
			assert.True(t, f.quoted)
		}
	}
}

func TestTypeFields_IgnoresUnexportedAndDashTag(t *testing.T) {
	type T struct {
		exported   string //nolint
		unexported string //nolint
		Skipped    string `json:"-"`
		Kept       string `json:"kept"`
	}
	fields := typeFields(reflect.TypeOf(T{}))
	for _, f := range fields {
		assert.NotEqual(t, "-", f.name)
		assert.NotEqual(t, "Skipped", f.name)
	}
	names := make([]string, len(fields))
	for i, f := range fields {
		names[i] = f.name
	}
	assert.Contains(t, names, "kept")
}

func TestConvertToJSONableObject_StringTarget(t *testing.T) {
	// When the JSON target is a string, numeric primitives should be converted.
	target := reflect.ValueOf("")
	result, err := convertToJSONableObject(int(42), &target)
	require.NoError(t, err)
	assert.Equal(t, "42", result)
}

func TestConvertToJSONableObject_Slice(t *testing.T) {
	input := []interface{}{1, 2, 3}
	result, err := convertToJSONableObject(input, nil)
	require.NoError(t, err)
	arr, ok := result.([]interface{})
	require.True(t, ok)
	assert.Equal(t, 3, len(arr))
}
