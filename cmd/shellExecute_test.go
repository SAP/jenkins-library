package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
)

type shellExecuteMockUtils struct {
	t      *testing.T
	config *shellExecuteOptions
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newShellExecuteTestsUtils() shellExecuteMockUtils {
	utils := shellExecuteMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func newShellExecuteMockUtils() *shellExecuteOptions {
	return &shellExecuteOptions{
		VaultServerURL: "",
		VaultNamespace: "",
		IsOutputNeed:   true,
		Sources:        nil,
	}
}

func (v *shellExecuteMockUtils) GetConfig() *shellExecuteOptions {
	return v.config
}

func TestRunShellExecute(t *testing.T) {
	// todo
	t.Parallel()
}
