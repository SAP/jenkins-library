//go:build !release

package cnbutils

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

type MockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
	*mock.DownloadMock
}

func (c *MockUtils) GetFileUtils() piperutils.FileUtils {
	return c.FilesMock
}
