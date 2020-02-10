package fortify

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetProjectByName(t *testing.T) {
	t.Run("test success", func(t *testing.T) {

		sys := NewSystemInstance("https://fortify.mo.sap.corp", "/ssc/api/v1", "N2VhMzMyMjctZmMwMi00ODJlLTk0NTQtZWZmZDI3NDAzMjMx", (60 * time.Second))

		result, err := sys.GetProjectByName("python-test-sven")
		assert.NoError(t, err, "GetProjectByName call not successful")
		assert.Equal(t, "python-test-sven", strings.ToLower(*result.Name), "Expected to receive python-test-sven")

		result2, err := sys.GetProjectVersionDetailsByNameAndProjectID(result.ID, "0")
		assert.NoError(t, err, "GetProjectVersionDetailsByNameAndProjectID call not successful")
		assert.Equal(t, "0", *result2.Name, "Expected project version with different name")
	})
}
