package command

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/cumulusurllog"
	"golang.org/x/sync/errgroup"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/pkg/errors"
)

// Command defines the information required for executing a call to any executable
type Command struct {
	ErrorCategoryMapping map[string][]string
	StepName             string
	dir                  string
	stdin                io.Reader
	stdout               io.Writer
	stderr               io.Writer
	env                  []string
	exitCode             int
}

type runner interface {
	SetDir(dir string)
	SetEnv(env []string)
	AppendEnv(env []string)
	Stdin(in io.Reader)
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	GetStdout() io.Writer
	GetStderr() io.Writer
}

// ExecRunner mock for intercepting calls to executables
type ExecRunner interface {
	runner
	RunExecutable(executable string, params ...string) error
	RunExecutableInBackground(executable string, params ...string) (Execution, error)
}

// ShellRunner mock for intercepting shell calls
type ShellRunner interface {
	runner
	RunShell(shell string, command string) error
}

// SetDir sets the working directory for the execution
func (c *Command) SetDir(dir string) {
	c.dir = dir
}

// SetEnv sets explicit environment variables to be used for execution
func (c *Command) SetEnv(env []string) {
	c.env = env
}

// AppendEnv appends environment variables to be used for execution
func (c *Command) AppendEnv(env []string) {
	c.env = append(c.env, env...)
}

func (c *Command) GetOsEnv() []string {
	return os.Environ()
}

// Stdin ..
func (c *Command) Stdin(stdin io.Reader) {
	c.stdin = stdin
}

// Stdout ..
func (c *Command) Stdout(stdout io.Writer) {
	c.stdout = stdout
}

// Stderr ..
func (c *Command) Stderr(stderr io.Writer) {
	c.stderr = stderr
}

// GetStdout Returns the writer for stdout
func (c *Command) GetStdout() io.Writer {
	return c.stdout
}

//GetStderr Return the writer for stderr
func (c *Command) GetStderr() io.Writer {
	return c.stderr
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
//    Thus, the executable needs to be on the PATH of the current process, and it is not sufficient to alter the PATH on cmd.Env.
func (c *Command) RunExecutable(executable string, params ...string) error {

	c.prepareOut()

	cmd := ExecCommand(executable, params...)

	if len(c.dir) > 0 {
		cmd.Dir = c.dir
	}

	log.Entry().Infof("running command: %v %v", executable, strings.Join(params, " "))

	appendEnvironment(cmd, c.env)

	if c.stdin != nil {
		cmd.Stdin = c.stdin
	}

	if err := c.runCmd(cmd); err != nil {
		return errors.Wrapf(err, "running command '%v' failed", executable)
	}
	return nil
}

// RunExecutableInBackground runs the specified executable with parameters in the background non-blocking
// !! While the cmd.Env is applied during command execution, it is NOT involved when the actual executable is resolved.
//    Thus, the executable needs to be on the PATH of the current process, and it is not sufficient to alter the PATH on cmd.Env.
func (c *Command) RunExecutableInBackground(executable string, params ...string) (Execution, error) {

	c.prepareOut()

	cmd := ExecCommand(executable, params...)

	if len(c.dir) > 0 {
		cmd.Dir = c.dir
	}

	log.Entry().Infof("running command: %v %v", executable, strings.Join(params, " "))

	appendEnvironment(cmd, c.env)

	if c.stdin != nil {
		cmd.Stdin = c.stdin
	}

	execution, err := c.startCmd(cmd)

	if err != nil {
		return nil, errors.Wrapf(err, "starting command '%v' failed", executable)
	}

	return execution, nil
}

// GetExitCode allows to retrieve the exit code of a command execution
func (c *Command) GetExitCode() int {
	return c.exitCode
}

func appendEnvironment(cmd *exec.Cmd, env []string) {

	if len(env) > 0 {

		// When cmd.Env is nil the environment variables from the current
		// process are also used by the forked process. Our environment variables
		// should not replace the existing environment, but they should be appended.
		// Hence, we populate cmd.Env first with the current environment in case we
		// find it empty. In case there is already something, we append to that environment.
		// In that case we assume the current values of `cmd.Env` has either been set up based
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
	wg := errgroup.Group{}

	srcOut := stdout
	srcErr := stderr

	if c.ErrorCategoryMapping != nil {
		prOut, pwOut := io.Pipe()
		trOut := io.TeeReader(stdout, pwOut)
		srcOut = prOut
		prErr, pwErr := io.Pipe()
		trErr := io.TeeReader(stderr, pwErr)
		srcErr = prErr
		type rw struct {
			tr io.Reader
			pw *io.PipeWriter
		}
		for _, source := range []rw{
			{trOut, pwOut},
			{trErr, pwErr},
		} {
			s := source
			wg.Go(func() (err error) {
				defer func() {
					dErr := s.pw.Close()
					if dErr != nil {
						err = fmt.Errorf("can't close piper writer: %w", dErr)
					}
				}()
				err = c.scanLog(s.tr)
				if err != nil {
					err = fmt.Errorf("can't scan log: %w", err)
				}
				return
			})
		}
	}
	if c.StepName != "" {
		cl := cumulusurllog.NewCumulusLogger(c.StepName)
		for _, source := range []io.ReadCloser{srcOut, srcErr} {
			s := source
			wg.Go(func() error {
				var buf bytes.Buffer
				w := bufio.NewWriter(&buf)
				_, err := piperutils.CopyData(w, s)
				if err != nil {
					return fmt.Errorf("can't copy data: %w", err)
				}
				err = w.Flush()
				if err != nil {
					return fmt.Errorf("can't flush: %w", err)
				}
				cl.Parse(buf)
				return nil
			})
		}
		err = wg.Wait()
		if err != nil {
			return nil, fmt.Errorf("waitgroup error: %w", err)
		}
		err = cl.WriteURLsLogToJSON()
		if err != nil {
			return nil, fmt.Errorf("can't write log: %w", err)
		}
	}
	type rw struct {
		std io.Writer
		src io.ReadCloser
	}
	for _, source := range []rw{
		{c.stdout, srcOut},
		{c.stderr, srcErr},
	} {
		s := source
		wg.Go(func() error {
			_, err := piperutils.CopyData(s.std, s.src)
			if err != nil {
				return fmt.Errorf("can't copy data: %w", err)
			}
			return nil
		})
	}
	err = wg.Wait()
	if err != nil {
		return nil, fmt.Errorf("waitgroup error: %w", err)
	}
	return &execution, nil
}

func (c *Command) scanLog(in io.Reader) error {
	scanner := bufio.NewScanner(in)
	scanner.Split(scanShortLines)
	for scanner.Scan() {
		line := scanner.Text()
		c.parseConsoleErrors(line)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan log: %w", err)
	}
	return nil
}

func scanShortLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	lenData := len(data)
	if atEOF && lenData == 0 {
		return 0, nil, nil
	}
	if lenData > 32767 && !bytes.Contains(data[0:lenData], []byte("\n")) {
		// we will neglect long output
		// no use cases known where this would be relevant
		// current accepted implication: error pattern would not be found
		// -> resulting in wrong error categorization
		return lenData, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 && i < 32767 {
		// We have a full newline-terminated line with a size limit
		//  is required since otherwise scanner would stall
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func (c *Command) parseConsoleErrors(logLine string) {
	for category, categoryErrors := range c.ErrorCategoryMapping {
		for _, errorPart := range categoryErrors {
			if matchPattern(logLine, errorPart) {
				log.SetErrorCategory(log.ErrorCategoryByString(category))
				return
			}
		}
	}
}

func matchPattern(text, pattern string) bool {
	if len(pattern) == 0 && len(text) != 0 {
		return false
	}
	parts := strings.Split(pattern, "*")
	for _, part := range parts {
		if !strings.Contains(text, part) {
			return false
		}
	}
	return true
}

func (c *Command) runCmd(cmd *exec.Cmd) error {

	execution, err := c.startCmd(cmd)
	if err != nil {
		return err
	}

	err = execution.Wait()

	if err != nil {
		// provide fallback to ensure a non 0 exit code in case of an error
		c.exitCode = 1
		// try to identify the detailed error code
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				c.exitCode = status.ExitStatus()
			}
		}
		return errors.Wrap(err, "cmd.Run() failed")
	}
	c.exitCode = 0
	return nil
}

func (c *Command) prepareOut() {

	//ToDo: check use of MultiWriter instead to always write into os.Stdout and os.Stdin?
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
