package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetErrorCategory(t *testing.T) {
	SetErrorCategory(ErrorCustom)
	assert.Equal(t, errorCategory, ErrorCustom)
}

func TestGetErrorCategory(t *testing.T) {
	errorCategory = ErrorCompliance
	assert.Equal(t, GetErrorCategory(), errorCategory)
}
