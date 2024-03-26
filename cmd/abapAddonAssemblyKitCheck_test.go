package cmd

import (
	"testing"
	// "github.com/stretchr/testify/assert"
)

// type abapAddonAssemblyKitCheckMockUtils struct {
// 	*mock.ExecMockRunner
// 	*mock.HttpClientMock
// }

// func newAbapAddonAssemblyKitCheckTestsUtils() abapAddonAssemblyKitCheckUtils {
// 	utils := abapAddonAssemblyKitCheckMockUtils{
// 		ExecMockRunner: &mock.ExecMockRunner{},
// 	}
// 	return utils
// }

func TestRunAbapAddonAssemblyKitCheck(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		// t.Parallel()
		// // init
		// config := abapAddonAssemblyKitCheckOptions{}

		// utils := newAbapAddonAssemblyKitCheckTestsUtils()

		// // test
		// err := runAbapAddonAssemblyKitCheck(&config, nil, utils, nil)

		// // assert
		// assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		// t.Parallel()
		// // init
		// config := abapAddonAssemblyKitCheckOptions{}

		// utils := newAbapAddonAssemblyKitCheckTestsUtils()

		// // test
		// err := runAbapAddonAssemblyKitCheck(&config, nil, utils, nil)

		// // assert
		// assert.EqualError(t, err, "cannot run without important file")
	})
}
