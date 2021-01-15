package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type transportRequestUploadSOLMANMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newTransportRequestUploadSOLMANTestsUtils() transportRequestUploadSOLMANMockUtils {
	utils := transportRequestUploadSOLMANMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunTransportRequestUploadSOLMAN(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := transportRequestUploadSOLMANOptions{}

		utils := newTransportRequestUploadSOLMANTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runTransportRequestUploadSOLMAN(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := transportRequestUploadSOLMANOptions{}

		utils := newTransportRequestUploadSOLMANTestsUtils()

		// test
		err := runTransportRequestUploadSOLMAN(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
