package mock

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
)

type BtpExecuterMock struct {
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

func (b *BtpExecuterMock) Stdin(in io.Reader) {
	b.stdin = in
}

func (b *BtpExecuterMock) Stdout(out io.Writer) {
	b.stdout = out
}

func (b *BtpExecuterMock) GetStdoutValue() string {
	return b.stdout.(*bytes.Buffer).String()
}

func (b *BtpExecuterMock) Run(cmdScript string) (err error) {
	parts := strings.Fields(cmdScript)
	execCall := BtpExecCall{Exec: parts[0], Params: parts[1:]}
	b.Calls = append(b.Calls, execCall)

	return b.handleCall(cmdScript, b.StdoutReturn, b.ShouldFailOnCommand, b.stdout)
}

func (b *BtpExecuterMock) RunSync(cmdScript string, cmdCheck string, timeoutMin int, pollIntervalSec int, negativeCheck bool) error {
	err := b.Run(cmdScript)
	if err != nil {
		return fmt.Errorf("Initial command execution failed: %w", err)
	}

	fmt.Printf("Started polling. Timeout: %d minutes\n", timeoutMin)

	fmt.Println("Checking command completion...")

	// Simulate polling
	err2 := b.Run(cmdCheck)

	outputStr := strings.TrimSpace(string(b.GetStdoutValue()))

	if err2 == nil && isCommandCompleted(outputStr, negativeCheck) {
		fmt.Println("Command execution completed successfully!")
		return nil
	}

	return fmt.Errorf("Command did not complete within the timeout period")
}

func isCommandCompleted(output string, negativeCheck bool) bool {
	var lines []string = strings.Split(output, "\n")

	check := strings.Contains(lines[len(lines)-1], "OK") || strings.Contains(output, "COMPLETED") || strings.Contains(output, "SUCCEEDED")
	if negativeCheck {
		return !check
	}
	return check
}

// Processes command results based on predefined mock data.
func (e *BtpExecuterMock) handleCall(call string, stdoutReturn map[string]string, shouldFailOnCommand map[string]error, stdout io.Writer) error {
	// Check if the command should return a specific output
	if stdoutReturn != nil {
		for pattern, output := range stdoutReturn {
			if matchCommand(pattern, call) {
				stdout.Write([]byte(output))
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
func matchCommand(pattern, command string) bool {
	if pattern == command {
		return true
	}
	r, err := regexp.Compile(pattern)
	return err == nil && r.MatchString(command)
}
