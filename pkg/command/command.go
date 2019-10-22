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

// Shell defines the information required for executing a shell call
type Shell struct {
	Dir    string
	Shell  string
	Script string
	Stdout io.Writer
	Stderr io.Writer
}

// Executable defines the information required for executing a call to any executable
type Executable struct {
	Dir        string
	Executable string
	Parameters []string
	Stdout     io.Writer
	Stderr     io.Writer
}

// ExecCommand defines how to execute os commands
var ExecCommand = exec.Command

// Run the specified command on the shell
func (s *Shell) Run() error {

	_out, _err := prepareOut(s.Stdout, s.Stderr)

	cmd := ExecCommand(s.Shell)

	cmd.Dir = s.Dir
	in := bytes.Buffer{}
	in.Write([]byte(s.Script))
	cmd.Stdin = &in

	if err := runCmd(cmd, _out, _err); err != nil {
		return errors.Wrapf(err, "running shell script failed with %v", s.Shell)
	}
	return nil
}

// Run the specified executable with parameters
func (e *Executable) Run() error {

	_out, _err := prepareOut(e.Stdout, e.Stderr)

	cmd := ExecCommand(e.Executable, e.Parameters...)

	cmd.Dir = e.Dir

	if err := runCmd(cmd, _out, _err); err != nil {
		return errors.Wrapf(err, "running command '%v' failed", e.Executable)
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
