package log

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetErrorCategory(t *testing.T) {
	SetErrorCategory(ErrorCustom)
	assert.Equal(t, errorCategory, ErrorCustom)
	assert.Equal(t, "custom", fmt.Sprint(errorCategory))
}

func TestGetErrorCategory(t *testing.T) {
	errorCategory = ErrorCompliance
	assert.Equal(t, GetErrorCategory(), errorCategory)
}
