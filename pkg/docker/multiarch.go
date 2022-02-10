package docker

import (
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

// IsBinfmtMiscSupportedByHost checks if the hosts kernel does support binfmt_misc
func IsBinfmtMiscSupportedByHost(utils piperutils.FileUtils) (bool, error) {
	return utils.DirExists("/proc/sys/fs/binfmt_misc")
}
