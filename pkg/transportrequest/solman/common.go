package solman

import (
	"github.com/SAP/jenkins-library/pkg/command"
)

// Exec interface collecting everything which is execution related
// and needed in the context of a SOLMAN upload.
type Exec interface {
	command.ExecRunner
	GetExitCode() int
}

// Connection Everything we need for connecting to Solution Manager
type Connection struct {
	Endpoint string
	User     string
	Password string
}
