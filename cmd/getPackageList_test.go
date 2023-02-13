package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type getPackageListMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newGetPackageListTestsUtils() getPackageListMockUtils {
	utils := getPackageListMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunGetPackageList(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := getPackageListOptions{}

		utils := newGetPackageListTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runGetPackageList(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := getPackageListOptions{}

		utils := newGetPackageListTestsUtils()

		// test
		err := runGetPackageList(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
