package shell

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWrapInQuotes(t *testing.T) {
	tt := []struct {
		in       string
		expected string
	}{
		{in: "ZdTq8@5gj$^9yYMy", expected: "'ZdTq8@5gj$^9yYMy'"},
		{in: "ZdTq8@5gj'^9yYMy", expected: "'ZdTq8@5gj'\"'\"'^9yYMy'"},
	}

	for _, test := range tt {
		assert.Equal(t, test.expected, WrapInQuotes(test.in))
	}
}
