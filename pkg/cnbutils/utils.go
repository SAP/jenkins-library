package cnbutils

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/docker"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

type BuildUtils interface {
	command.ExecRunner
	piperutils.FileUtils
	docker.Download
}
