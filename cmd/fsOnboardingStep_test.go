package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type fsOnboardingStepMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newFsOnboardingStepTestsUtils() fsOnboardingStepMockUtils {
	utils := fsOnboardingStepMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunFsOnboardingStep(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := fsOnboardingStepOptions{}

		utils := newFsOnboardingStepTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runFsOnboardingStep(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := fsOnboardingStepOptions{}

		utils := newFsOnboardingStepTestsUtils()

		// test
		err := runFsOnboardingStep(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
