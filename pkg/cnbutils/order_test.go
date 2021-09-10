package cnbutils

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestOrderSave(t *testing.T) {
	t.Run("successfully Encode struct to toml format", func(t *testing.T) {
		mockUtils := MockUtils{
			ExecMockRunner: &mock.ExecMockRunner{},
			FilesMock:      &mock.FilesMock{},
			DockerMock:     &DockerMock{},
		}
		testOrder := Order{
			Order: []OrderEntry{{
				Group: []BuildpackRef{{
					ID:       "test",
					Version:  "0.0.1",
					Optional: true,
				}},
			}},
			Utils: mockUtils,
		}

		err := testOrder.Save("/tmp/order.toml")

		assert.NoError(t, err)
		assert.True(t, mockUtils.HasWrittenFile("/tmp/order.toml"))
		result, err := mockUtils.FileRead("/tmp/order.toml")
		assert.NoError(t, err)
		assert.Equal(t, "\n[[order]]\n\n  [[order.group]]\n    id = \"test\"\n    optional = true\n    version = \"0.0.1\"\n", string(result))
	})

	t.Run("raises an error if unable to write the file", func(t *testing.T) {
		mockUtils := MockUtils{
			ExecMockRunner: &mock.ExecMockRunner{},
			FilesMock:      &mock.FilesMock{},
			DockerMock:     &DockerMock{},
		}
		mockUtils.FileWriteErrors = map[string]error{
			"/tmp/order.toml": fmt.Errorf("unable to write to file"),
		}
		testOrder := Order{
			Order: []OrderEntry{{
				Group: []BuildpackRef{{
					ID:       "test",
					Version:  "0.0.1",
					Optional: true,
				}},
			}},
			Utils: mockUtils,
		}

		err := testOrder.Save("/tmp/order.toml")

		assert.Error(t, err, "unable to write to file")
		assert.False(t, mockUtils.HasWrittenFile("/tmp/order.toml"))
	})
}
