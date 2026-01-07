package btp

import (
	"bytes"
	"fmt"
	"io"
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

func (e *Executor) Run(cmdScript []string) (err error) {
	if err := e.Cmd.RunExecutable(cmdScript[0], cmdScript[1:]...); err != nil {
		return fmt.Errorf("Failed to execute BTP CLI: %w", err)
	}
	return nil
}

/*
@param timeout : in seconds
@param pollInterval : in seconds
@param negativeCheck : set to false if you want to check the negation of the response of `cmdCheck`
*/
func (e *Executor) RunSync(opts RunSyncOptions) (err error) {
	err = e.Run(opts.CmdScript)

	if err != nil {
		return fmt.Errorf("Failed to execute BTP CLI: %w", err)
	}

	// Poll to check completion
	timeoutDuration := time.Duration(opts.TimeoutSeconds) * time.Second
	pollIntervall := time.Duration(opts.PollInterval) * time.Second
	startTime := time.Now()

	fmt.Println("Checking command completion...")

	for time.Since(startTime) < timeoutDuration {
		// Wait before the next check
		time.Sleep(pollIntervall)

		check := opts.CheckFunc()

		if check {
			fmt.Println("Command execution completed successfully!")
			return nil
		} else {
			fmt.Println("Command not yet completed, checking again...")
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
	Run(cmdScript []string) error
	RunSync(opts RunSyncOptions) error
}
