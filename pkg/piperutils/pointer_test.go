//go:build unit
// +build unit

package piperutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeDereferenceString(t *testing.T) {
	type testCase[T any] struct {
		name string
		p    *T
		want T
	}
	str := "test"
	tests := []testCase[string]{
		{
			name: "nil",
			p:    nil,
			want: "",
		},
		{
			name: "non-nil",
			p:    &str,
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, SafeDereference(tt.p), "SafeDereference(%v)", tt.p)
		})
	}
}

func TestSafeDereferenceInt64(t *testing.T) {
	type testCase[T any] struct {
		name string
		p    *T
		want T
	}
	i64 := int64(111)
	tests := []testCase[int64]{
		{
			name: "nil",
			p:    nil,
			want: 0,
		},
		{
			name: "non-nil",
			p:    &i64,
			want: 111,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, SafeDereference(tt.p), "SafeDereference(%v)", tt.p)
		})
	}
}
