package btp

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
)

// Stdin ..
func (e *Executor) Stdin(stdin io.Reader) {
	e.Cmd.Stdin(stdin)
}

// Stdout ..
func (e *Executor) Stdout(stdout io.Writer) {
	e.Cmd.Stdout(stdout)
}

func (e *Executor) GetStdoutValue() string {
	return e.Cmd.GetStdout().(*bytes.Buffer).String()
}

func (e *Executor) Run(cmdScript string) (err error) {
	parts := strings.Fields(cmdScript)
	if err := e.Cmd.RunExecutable(parts[0], parts[1:]...); err != nil {
		return fmt.Errorf("Failed to execute BTP CLI: %w", err)
	}
	return nil
}

/*
@param timeout : in seconds
@param pollInterval : in seconds
@param negativeCheck : set to false if you whant to check the negation of the response of `cmdCheck`
*/
func (e *Executor) RunSync(cmdScript string, cmdCheck string, timeout int, pollInterval int, negativeCheck bool) (err error) {
	err = e.Run(cmdScript)

	// Poll to check completion
	timeoutDuration := time.Duration(timeout) * time.Second
	pollIntervall := time.Duration(pollInterval) * time.Second
	startTime := time.Now()

	fmt.Println("Checking command completion...")

	for time.Since(startTime) < timeoutDuration {
		// Wait before the next check
		time.Sleep(pollIntervall)

		parts := strings.Fields(cmdCheck)
		err := e.Cmd.RunExecutable(parts[0], parts[1:]...)

		if (negativeCheck && (err != nil)) || (!negativeCheck && (err == nil)) {
			fmt.Println("Command execution completed successfully!")
			return nil
		}
	}

	return fmt.Errorf("Command did not completed within the timeout period")
}

type Executor struct {
	Cmd command.Command
}

type btpRunner interface {
	Stdin(in io.Reader)
	Stdout(out io.Writer)
	GetStdoutValue() string
}

type ExecRunner interface {
	btpRunner
	Run(cmdScript string) error
	RunSync(cmdScript string, cmdCheck string, timeoutMin int, pollIntervalSec int, negativeCheck bool) error
}
