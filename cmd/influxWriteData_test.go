package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type influxWriteDataMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newInfluxWriteDataTestsUtils() influxWriteDataMockUtils {
	utils := influxWriteDataMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunInfluxWriteData(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := influxWriteDataOptions{}

		utils := newInfluxWriteDataTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runInfluxWriteData(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := influxWriteDataOptions{}

		utils := newInfluxWriteDataTestsUtils()

		// test
		err := runInfluxWriteData(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
