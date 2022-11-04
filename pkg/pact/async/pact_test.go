package async

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPactComparisonMap(t *testing.T) {
	spec := &AsyncPactSpec{}

	res, err := spec.ComparisonMap()
	assert.NoError(t, err)
	assert.Equal(t, "", res)
}