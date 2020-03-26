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

// SetDir sets the working directory for the execution
func (c *Command) SetDir(d string) {
	c.dir = d
}

// SetEnv sets explicit environment variables to be used for execution
func (c *Command) SetEnv(e []string) {
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

	appendEnvironment(cmd, c.env)

	in := bytes.Buffer{}
	in.Write([]byte(script))
	cmd.Stdin = &in

	if err := runCmd(cmd, _out, _err); err != nil {
		return errors.Wrapf(err, "running shell script failed with %v", shell)
	}
	return nil
}

// RunExecutable runs the specified executable with parameters
// !! While the cmd.Env is applied during command execution, it is NOT involved when the actual executable is resolved.
//    Thus the executable needs to be on the PATH of the current process and it is not sufficient to alter the PATH on cmd.Env.
func (c *Command) RunExecutable(executable string, params ...string) error {

	_out, _err := prepareOut(c.stdout, c.stderr)

	cmd := ExecCommand(executable, params...)

	if len(c.dir) > 0 {
		cmd.Dir = c.dir
	}

	appendEnvironment(cmd, c.env)

	if err := runCmd(cmd, _out, _err); err != nil {
		return errors.Wrapf(err, "running command '%v' failed", executable)
	}
	return nil
}

func appendEnvironment(cmd *exec.Cmd, env []string) {

	if len(env) > 0 {

		// When cmd.Env is nil the environment variables from the current
		// process are also used by the forked process. Our environment variables
		// should not replace the existing environment, but they should be appended.
		// Hence we populate cmd.Env first with the current environment in case we
		// find it empty. In case there is already something, we append to that environment.
		// In that case we assume the current values of `cmd.Env` has either been setup based
		// on `os.Environ()` or that was initialized in another way for a good reason.
		//
		// In case we have the same environment variable as in the current environment (`os.Environ()`)
		// and in `env`, the environment variable from `env` is effectively used since this is the
		// later one. There is no merging between both environment variables.
		//
		// cf. https://golang.org/pkg/os/exec/#Command
		//     If Env contains duplicate environment keys, only the last
		//     value in the slice for each duplicate key is used.

		if len(cmd.Env) == 0 {
			cmd.Env = os.Environ()
		}
		cmd.Env = append(cmd.Env, env...)
	}
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
