package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type tmsExportMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newTmsExportTestsUtils() tmsExportMockUtils {
	utils := tmsExportMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunTmsExport(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := tmsExportOptions{}

		utils := newTmsExportTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runTmsExport(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := tmsExportOptions{}

		utils := newTmsExportTestsUtils()

		// test
		err := runTmsExport(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
