// +build !release

package mock

import (
	"io"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
)

type ExecMockRunner struct {
	Dir                 []string
	Env                 []string
	ExitCode            int
	Calls               []ExecCall
	stdin               io.Reader
	stdout              io.Writer
	stderr              io.Writer
	StdoutReturn        map[string]string
	ShouldFailOnCommand map[string]error
}

type ExecCall struct {
	Execution *Execution
	Async     bool
	Exec      string
	Params    []string
}

type Execution struct {
	Killed bool
}

type ShellMockRunner struct {
	Dir                 string
	Env                 []string
	ExitCode            int
	Calls               []string
	Shell               []string
	stdin               io.Reader
	stdout              io.Writer
	stderr              io.Writer
	StdoutReturn        map[string]string
	ShouldFailOnCommand map[string]error
}

func (m *ExecMockRunner) SetDir(d string) {
	m.Dir = append(m.Dir, d)
}

func (m *ExecMockRunner) SetEnv(e []string) {
	m.Env = e
}

func (m *ExecMockRunner) AppendEnv(e []string) {
	m.Env = append(m.Env, e...)
}

func (m *ExecMockRunner) RunExecutable(e string, p ...string) error {

	exec := ExecCall{Exec: e, Params: p}
	m.Calls = append(m.Calls, exec)

	c := strings.Join(append([]string{e}, p...), " ")

	return handleCall(c, m.StdoutReturn, m.ShouldFailOnCommand, m.stdout)
}

func (m *ExecMockRunner) GetExitCode() int {
	return m.ExitCode
}

func (m *ExecMockRunner) RunExecutableInBackground(e string, p ...string) (command.Execution, error) {

	execution := Execution{}
	exec := ExecCall{Exec: e, Params: p, Async: true, Execution: &execution}
	m.Calls = append(m.Calls, exec)

	c := strings.Join(append([]string{e}, p...), " ")

	err := handleCall(c, m.StdoutReturn, m.ShouldFailOnCommand, m.stdout)
	if err != nil {
		return nil, err
	}
	return &execution, nil
}

func (m *ExecMockRunner) Stdin(in io.Reader) {
	m.stdin = in
}

func (m *ExecMockRunner) Stdout(out io.Writer) {
	m.stdout = out
}

func (m *ExecMockRunner) GetStdout() io.Writer {
	return m.stdout
}

func (m *ExecMockRunner) Stderr(err io.Writer) {
	m.stderr = err
}

func (m *ExecMockRunner) GetStderr() io.Writer {
	return m.stderr
}

func (m *ShellMockRunner) SetDir(d string) {
	m.Dir = d
}

func (m *ShellMockRunner) SetEnv(e []string) {
	m.Env = append(m.Env, e...)
}

func (m *ShellMockRunner) AppendEnv(e []string) {
	m.Env = append(m.Env, e...)
}

func (m *ShellMockRunner) RunShell(s string, c string) error {

	m.Shell = append(m.Shell, s)
	m.Calls = append(m.Calls, c)

	return handleCall(c, m.StdoutReturn, m.ShouldFailOnCommand, m.stdout)
}

func (m *ShellMockRunner) GetExitCode() int {
	return m.ExitCode
}

func (e *Execution) Kill() error {
	e.Killed = true
	return nil
}

func (e *Execution) Wait() error {
	return nil
}

func handleCall(call string, stdoutReturn map[string]string, shouldFailOnCommand map[string]error, stdout io.Writer) error {

	if stdoutReturn != nil {
		for k, v := range stdoutReturn {

			found := k == call

			if !found {

				r, e := regexp.Compile(k)
				if e != nil {
					return e
					// we don't distinguish here between an error returned
					// since it was configured or returning this error here
					// indicating an invalid regex. Anyway: when running the
					// test we will see it ...
				}
				if r.MatchString(call) {
					found = true

				}
			}

			if found {
				stdout.Write([]byte(v))
			}
		}
	}

	if shouldFailOnCommand != nil {
		for k, v := range shouldFailOnCommand {

			found := k == call

			if !found {
				r, e := regexp.Compile(k)
				if e != nil {
					return e
					// we don't distinguish here between an error returned
					// since it was configured or returning this error here
					// indicating an invalid regex. Anyway: when running the
					// test we will see it ...
				}
				if r.MatchString(call) {
					found = true

				}
			}

			if found {
				return v
			}
		}
	}

	return nil
}

func (m *ShellMockRunner) Stdin(in io.Reader) {
	m.stdin = in
}

func (m *ShellMockRunner) Stdout(out io.Writer) {
	m.stdout = out
}

func (m *ShellMockRunner) GetStdout() io.Writer {
	return m.stdout
}

func (m *ShellMockRunner) GetStderr() io.Writer {
	return m.stderr
}

func (m *ShellMockRunner) Stderr(err io.Writer) {
	m.stderr = err
}

type StepOptions struct {
	TestParam string `json:"testParam,omitempty"`
}

func OpenFileMock(name string, tokens map[string]string) (io.ReadCloser, error) {
	var r string
	switch name {
	case "testDefaults.yml":
		r = "general:\n  testParam: testValue"
	case "testDefaultsInvalid.yml":
		r = "invalid yaml"
	default:
		r = ""
	}
	return ioutil.NopCloser(strings.NewReader(r)), nil
}
