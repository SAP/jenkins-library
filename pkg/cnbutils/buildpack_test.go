package cnbutils_test

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestBuildpackDownload(t *testing.T) {
	var mockUtils = &cnbutils.MockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}

	t.Run("it creates an order object", func(t *testing.T) {
		order, err := cnbutils.DownloadBuildpacks("/destination", []string{"buildpack"}, "/tmp/config.json", mockUtils)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(order.Order))
	})
}
