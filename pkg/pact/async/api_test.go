package async

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIComparisonMap(t *testing.T) {

	spec := &AsyncAPISpec{}

	res, err := spec.ComparisonMap()
	assert.NoError(t, err)
	assert.Equal(t, "", res)

}