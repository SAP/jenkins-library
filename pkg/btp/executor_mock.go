package btp

import (
	"bytes"
	"io"
	"regexp"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

func (b *BtpExecutorMock) Stdin(in io.Reader) {
	b.stdin = in
}

func (b *BtpExecutorMock) Stdout(out io.Writer) {
	b.stdout = out
}

func (b *BtpExecutorMock) GetStdoutValue() string {
	return b.stdout.(*bytes.Buffer).String()
}

func (b *BtpExecutorMock) Run(cmdScript []string) (err error) {
	execCall := BtpExecCall{Exec: cmdScript[0], Params: cmdScript[1:]}
	b.Calls = append(b.Calls, execCall)

	return b.handleCall(cmdScript, b.StdoutReturn, b.ShouldFailOnCommand, b.stdout)
}

func (b *BtpExecutorMock) RunSync(opts RunSyncOptions) error {
	err := b.Run(opts.CmdScript)
	if err != nil {
		return errors.Wrap(err, "Initial command execution failed")
	}

	log.Entry().Infof("Started polling. Timeout: %d minutes\n", opts.TimeoutSeconds/60)
	log.Entry().Info("Checking command completion...")

	// Simulate polling
	check := opts.CheckFunc()

	if check.successful && check.done {
		log.Entry().Info("Command execution completed successfully!")
		return nil
	} else {
		log.Entry().Info("Command not yet completed, checking again...")
	}

	return errors.New("Command did not complete within the timeout period")
}

// Processes command results based on predefined mock data.
func (e *BtpExecutorMock) handleCall(call []string, stdoutReturn map[string]string, shouldFailOnCommand map[string]error, stdout io.Writer) error {
	// Check if the command should return a specific output
	if stdoutReturn != nil {
		for pattern, output := range stdoutReturn {
			if matchCommand(pattern, call) {
				stdout.Write([]byte(output))
				return nil
			}
		}
	}

	// Check if the command should fail
	if shouldFailOnCommand != nil {
		for pattern, err := range shouldFailOnCommand {
			if matchCommand(pattern, call) {
				return err
			}
		}
	}

	return nil
}

// matchCommand checks if a command matches a pattern (direct string match or regex).
func matchCommand(pattern string, command []string) bool {
	if pattern == strings.Join(command, " ") {
		return true
	}
	r, err := regexp.Compile(pattern)
	return err == nil && r.MatchString(strings.Join(command, " "))
}

type BtpExecutorMock struct {
	Cmd                 command.Command
	Calls               []BtpExecCall
	StdoutReturn        map[string]string
	ShouldFailOnCommand map[string]error
	ExitCode            int
	stdin               io.Reader
	stdout              io.Writer
	stderr              io.Writer
}

// Stores information about executed commands.
type BtpExecCall struct {
	Exec   string
	Params []string
}
