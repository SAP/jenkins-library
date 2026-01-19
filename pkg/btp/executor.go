package btp

import (
	"bytes"
	"io"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
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
		return errors.Wrap(err, "Failed to execute BTP CLI")
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

	if err != nil && !opts.IgnoreErrorOnFirstCall {
		return errors.Wrap(err, "Failed to execute BTP CLI (Sync)")
	}

	// Save if we are now waiting things to be Ready
	waitingForReady := false

	nbRetry := 0
	maxRetries := 3

	// Poll to check completion
	timeoutDuration := time.Duration(opts.TimeoutSeconds) * time.Second
	pollIntervall := time.Duration(opts.PollInterval) * time.Second
	startTime := time.Now()

	log.Entry().Info("Checking command completion...")

	for time.Since(startTime) < timeoutDuration {
		if nbRetry >= maxRetries {
			return errors.New("Maximum number of retries reached while polling for command completion")
		}

		// Wait before the next check
		time.Sleep(pollIntervall)

		check := opts.CheckFunc()

		if check.successful {
			if check.done {
				log.Entry().Info("Command execution completed successfully!")
				return nil
			} else {
				waitingForReady = true
				log.Entry().Info("Command not yet completed, waiting for Readiness...")
			}
		} else {
			if waitingForReady {
				log.Entry().Info("Command was previously in progress, but now reports failure.")
				nbRetry++
				return errors.New("Command execution failed during polling")
			} else {
				log.Entry().Info("Command not yet completed, checking again...")
			}
		}

	}

	return errors.New("Command did not completed within the timeout period")
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
