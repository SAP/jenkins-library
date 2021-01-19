package command

import (
	"os/exec"
	"sync"
)

//errCopyStdout and errCopyStderr are filled after the command execution after Wait() terminates
type execution struct {
	cmd           *exec.Cmd
	wg            sync.WaitGroup
	errCopyStdout error
	errCopyStderr error
}

func (execution *execution) Kill() error {
	return execution.cmd.Process.Kill()
}

func (execution *execution) Wait() error {
	execution.wg.Wait()
	return execution.cmd.Wait()
}

// Execution references a background process which is started by RunExecutableInBackground
type Execution interface {
	Kill() error
	Wait() error
}
