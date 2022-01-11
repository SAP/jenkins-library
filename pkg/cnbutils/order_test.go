package cnbutils_test

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestOrderSave(t *testing.T) {
	t.Run("successfully Encode struct to toml format (multiple buildpacks)", func(t *testing.T) {
		mockUtils := &cnbutils.MockUtils{
			ExecMockRunner: &mock.ExecMockRunner{},
			FilesMock:      &mock.FilesMock{},
		}

		testBuildpacks := []cnbutils.BuildPackMetadata{
			{
				ID:      "paketo-buildpacks/sap-machine",
				Version: "1.1.1",
			},
			{
				ID:      "paketo-buildpacks/java",
				Version: "2.2.2",
			},
		}
		testOrder := cnbutils.Order{
			Utils: mockUtils,
		}

		var testEntry cnbutils.OrderEntry
		testEntry.Group = append(testEntry.Group, testBuildpacks...)
		testOrder.Order = []cnbutils.OrderEntry{testEntry}

		err := testOrder.Save("/tmp/order.toml")

		assert.NoError(t, err)
		assert.True(t, mockUtils.HasWrittenFile("/tmp/order.toml"))
		result, err := mockUtils.FileRead("/tmp/order.toml")
		assert.NoError(t, err)
		assert.Equal(t, "\n[[order]]\n\n  [[order.group]]\n    id = \"paketo-buildpacks/sap-machine\"\n    version = \"1.1.1\"\n\n  [[order.group]]\n    id = \"paketo-buildpacks/java\"\n    version = \"2.2.2\"\n", string(result))
	})

	t.Run("raises an error if unable to write the file", func(t *testing.T) {
		mockUtils := &cnbutils.MockUtils{
			ExecMockRunner: &mock.ExecMockRunner{},
			FilesMock:      &mock.FilesMock{},
		}
		mockUtils.FileWriteErrors = map[string]error{
			"/tmp/order.toml": fmt.Errorf("unable to write to file"),
		}
		testOrder := cnbutils.Order{
			Utils: mockUtils,
		}

		err := testOrder.Save("/tmp/order.toml")

		assert.Error(t, err, "unable to write to file")
		assert.False(t, mockUtils.HasWrittenFile("/tmp/order.toml"))
	})
}
