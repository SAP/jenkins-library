package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type abapEnvironmentUpdateAddOnProductMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newAbapEnvironmentUpdateAddOnProductTestsUtils() abapEnvironmentUpdateAddOnProductMockUtils {
	utils := abapEnvironmentUpdateAddOnProductMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunAbapEnvironmentUpdateAddOnProduct(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := abapEnvironmentUpdateAddOnProductOptions{}

		utils := newAbapEnvironmentUpdateAddOnProductTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runAbapEnvironmentUpdateAddOnProduct(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := abapEnvironmentUpdateAddOnProductOptions{}

		utils := newAbapEnvironmentUpdateAddOnProductTestsUtils()

		// test
		err := runAbapEnvironmentUpdateAddOnProduct(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
