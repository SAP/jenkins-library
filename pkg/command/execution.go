package command

import (
	"os/exec"
)

// Execution references a background process which is started by RunExecutableInBackground
type Execution interface {
	Kill() error
	Wait() error
}

type execution struct {
	cmd *exec.Cmd
}

func (execution *execution) Kill() error {
	return execution.cmd.Process.Kill()
}

func (execution *execution) Wait() error {
	return execution.cmd.Wait()
}
