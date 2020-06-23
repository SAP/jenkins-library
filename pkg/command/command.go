package command

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

// Command defines the information required for executing a call to any executable
type Command struct {
	ErrorCategoryMapping map[string][]string
	dir                  string
	stdout               io.Writer
	stderr               io.Writer
	env                  []string
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

	c.prepareOut()

	cmd := ExecCommand(shell)

	if len(c.dir) > 0 {
		cmd.Dir = c.dir
	}

	appendEnvironment(cmd, c.env)

	in := bytes.Buffer{}
	in.Write([]byte(script))
	cmd.Stdin = &in

	log.Entry().Infof("running shell script: %v %v", shell, script)

	if err := c.runCmd(cmd); err != nil {
		return errors.Wrapf(err, "running shell script failed with %v", shell)
	}
	return nil
}

// RunExecutable runs the specified executable with parameters
// !! While the cmd.Env is applied during command execution, it is NOT involved when the actual executable is resolved.
//    Thus the executable needs to be on the PATH of the current process and it is not sufficient to alter the PATH on cmd.Env.
func (c *Command) RunExecutable(executable string, params ...string) error {

	c.prepareOut()

	cmd := ExecCommand(executable, params...)

	if len(c.dir) > 0 {
		cmd.Dir = c.dir
	}

	log.Entry().Infof("running command: %v %v", executable, strings.Join(params, (" ")))

	appendEnvironment(cmd, c.env)

	if err := c.runCmd(cmd); err != nil {
		return errors.Wrapf(err, "running command '%v' failed", executable)
	}
	return nil
}

// RunExecutableInBackground runs the specified executable with parameters in the background non blocking
// !! While the cmd.Env is applied during command execution, it is NOT involved when the actual executable is resolved.
//    Thus the executable needs to be on the PATH of the current process and it is not sufficient to alter the PATH on cmd.Env.
func (c *Command) RunExecutableInBackground(executable string, params ...string) (Execution, error) {

	c.prepareOut()

	cmd := ExecCommand(executable, params...)

	if len(c.dir) > 0 {
		cmd.Dir = c.dir
	}

	log.Entry().Infof("running command: %v %v", executable, strings.Join(params, (" ")))

	appendEnvironment(cmd, c.env)

	execution, err := c.startCmd(cmd)

	if err != nil {
		return nil, errors.Wrapf(err, "starting command '%v' failed", executable)
	}

	return execution, nil
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

func (c *Command) startCmd(cmd *exec.Cmd) (*execution, error) {

	stdout, stderr, err := cmdPipes(cmd)

	if err != nil {
		return nil, errors.Wrap(err, "getting command pipes failed")
	}

	err = cmd.Start()
	if err != nil {
		return nil, errors.Wrap(err, "starting command failed")
	}

	execution := execution{cmd: cmd}
	execution.wg.Add(2)

	srcOut := stdout
	srcErr := stderr

	if c.ErrorCategoryMapping != nil {
		prOut, pwOut := io.Pipe()
		trOut := io.TeeReader(stdout, pwOut)
		srcOut = prOut

		prErr, pwErr := io.Pipe()
		trErr := io.TeeReader(stderr, pwErr)
		srcErr = prErr

		execution.wg.Add(2)

		go func() {

			defer execution.wg.Done()
			defer pwOut.Close()

			scanner := bufio.NewScanner(trOut)
			for scanner.Scan() {
				line := scanner.Text()
				c.parseConsoleErrors(line)
			}
			if err := scanner.Err(); err != nil {
				log.Entry().WithError(err).Info("failed to scan log file")
			}

		}()

		go func() {

			defer execution.wg.Done()
			defer pwErr.Close()

			scanner := bufio.NewScanner(trErr)
			for scanner.Scan() {
				line := scanner.Text()
				c.parseConsoleErrors(line)
			}
			if err := scanner.Err(); err != nil {
				log.Entry().WithError(err).Info("failed to scan log file")
			}

		}()
	}

	go func() {
		_, execution.errCopyStdout = io.Copy(c.stdout, srcOut)
		execution.wg.Done()
	}()

	go func() {
		_, execution.errCopyStderr = io.Copy(c.stderr, srcErr)
		execution.wg.Done()
	}()

	return &execution, nil
}

func (c *Command) parseConsoleErrors(logLine string) {
	for category, categoryErrors := range c.ErrorCategoryMapping {
		for _, errorPart := range categoryErrors {
			if strings.Contains(logLine, errorPart) {
				log.SetErrorCategory(log.ErrorCategoryByString(category))
				return
			}
		}
	}
}

func (c *Command) runCmd(cmd *exec.Cmd) error {

	execution, err := c.startCmd(cmd)
	if err != nil {
		return err
	}

	err = execution.Wait()

	if execution.errCopyStdout != nil || execution.errCopyStderr != nil {
		return fmt.Errorf("failed to capture stdout/stderr: '%v'/'%v'", execution.errCopyStdout, execution.errCopyStderr)
	}

	if err != nil {
		return errors.Wrap(err, "cmd.Run() failed")
	}

	return nil
}

func (c *Command) prepareOut() {

	//ToDo: check use of multiwriter instead to always write into os.Stdout and os.Stdin?
	//stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	//stderr := io.MultiWriter(os.Stderr, &stderrBuf)

	if c.stdout == nil {
		c.stdout = os.Stdout
	}
	if c.stderr == nil {
		c.stderr = os.Stderr
	}
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
