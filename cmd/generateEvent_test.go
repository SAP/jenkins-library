package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type generateEventMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newGenerateEventTestsUtils() generateEventMockUtils {
	utils := generateEventMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunGenerateEvent(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		utils := newGenerateEventTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runGenerateEvent(nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		utils := newGenerateEventTestsUtils()

		// test
		err := runGenerateEvent(nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
