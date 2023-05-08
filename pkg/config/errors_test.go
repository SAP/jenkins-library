//go:build unit
// +build unit

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseError(t *testing.T) {
	err := NewParseError("Parsing failed")

	assert.Equal(t, "Parsing failed", err.Error())
}
