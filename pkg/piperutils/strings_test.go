//go:build unit

package piperutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTitle(t *testing.T) {
	assert.Equal(t, "TEST", Title("tEST"))
	assert.Equal(t, "Test", Title("test"))
	assert.Equal(t, "TEST", Title("TEST"))
	assert.Equal(t, "Test", Title("Test"))
	assert.Equal(t, "TEST1 Test2 TEsT3 Test4", Title("TEST1 test2 tEsT3 Test4"))
}
