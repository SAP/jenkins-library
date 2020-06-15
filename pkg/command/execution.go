package command

import "os/exec"

type Execution struct {
	cmd *exec.Cmd
}

func (execution *Execution) Kill() error {
	return execution.cmd.Process.Kill()
}

func (execution *Execution) Wait() error {
	return execution.cmd.Wait()
}
