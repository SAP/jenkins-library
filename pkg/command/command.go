package command

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/pkg/errors"
)

// Command defines the information required for executing a call to any executable
type Command struct {
	dir    string
	stdout io.Writer
	stderr io.Writer
	env    []string
}

// Dir sets the working directory for the execution
func (c *Command) Dir(d string) {
	c.dir = d
}

// Env sets explicit environment variables to be used for execution
func (c *Command) Env(e []string) {
	c.env = e
}

// Stdout ..
func (c *Command) Stdout(stdout io.Writer) {
	c.stdout = stdout
}

// Stderr ..
func (c *Command) Stderr(stderr io.Writer) {
	c.stderr = stderr
}

// ExecCommand defines how to execute os commands
var ExecCommand = exec.Command

// RunShell runs the specified command on the shell
func (c *Command) RunShell(shell, script string) error {

	_out, _err := prepareOut(c.stdout, c.stderr)

	cmd := ExecCommand(shell)

	if len(c.dir) > 0 {
		cmd.Dir = c.dir
	}

	if len(c.env) > 0 {
		cmd.Env = c.env
	}

	in := bytes.Buffer{}
	in.Write([]byte(script))
	cmd.Stdin = &in

	if err := runCmd(cmd, _out, _err); err != nil {
		return errors.Wrapf(err, "running shell script failed with %v", shell)
	}
	return nil
}

// RunExecutable runs the specified executable with parameters
func (c *Command) RunExecutable(executable string, params ...string) error {

	_out, _err := prepareOut(c.stdout, c.stderr)

	cmd := ExecCommand(executable, params...)

	if len(c.dir) > 0 {
		cmd.Dir = c.dir
	}

	if len(c.env) > 0 {
		cmd.Env = c.env
	}

	if err := runCmd(cmd, _out, _err); err != nil {
		return errors.Wrapf(err, "running command '%v' failed", executable)
	}
	return nil
}

func runCmd(cmd *exec.Cmd, _out, _err io.Writer) error {

	stdout, stderr, err := cmdPipes(cmd)

	if err != nil {
		return errors.Wrap(err, "getting commmand pipes failed")
	}

	err = cmd.Start()
	if err != nil {
		return errors.Wrap(err, "starting command failed")
	}

	var wg sync.WaitGroup
	wg.Add(2)

	var errStdout, errStderr error

	go func() {
		_, errStdout = io.Copy(_out, stdout)
		wg.Done()
	}()

	go func() {
		_, errStderr = io.Copy(_err, stderr)
		wg.Done()
	}()

	wg.Wait()

	err = cmd.Wait()

	if err != nil {
		return errors.Wrap(err, "cmd.Run() failed")
	}

	if errStdout != nil || errStderr != nil {
		return fmt.Errorf("failed to capture stdout/stderr: '%v'/'%v'", errStdout, errStderr)
	}

	return nil
}

func prepareOut(stdout, stderr io.Writer) (io.Writer, io.Writer) {

	//ToDo: check use of multiwriter instead to always write into os.Stdout and os.Stdin?
	//stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	//stderr := io.MultiWriter(os.Stderr, &stderrBuf)

	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}

	return stdout, stderr
}

func cmdPipes(cmd *exec.Cmd) (io.ReadCloser, io.ReadCloser, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, errors.Wrap(err, "getting Stdout pipe failed")
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, errors.Wrap(err, "getting Stderr pipe failed")
	}
	return stdout, stderr, nil
}
