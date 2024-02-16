package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type emitEventsMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newEmitEventsTestsUtils() emitEventsMockUtils {
	utils := emitEventsMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunEmitEvents(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := emitEventsOptions{}

		utils := newEmitEventsTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runEmitEvents(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := emitEventsOptions{}

		utils := newEmitEventsTestsUtils()

		// test
		err := runEmitEvents(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
