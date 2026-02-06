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

// Stderr ..
func (e *Executor) Stderr(stderr io.Writer) {
	e.Cmd.Stderr(stderr)
}

func (e *Executor) GetStdoutValue() string {
	return e.Cmd.GetStdout().(*bytes.Buffer).String()
}

func (e *Executor) GetStderrValue() string {
	return e.Cmd.GetStderr().(*bytes.Buffer).String()
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
func (e *Executor) RunSync(opts RunSyncOptions) error {
	var errorResBytes bytes.Buffer
	e.Stderr(&errorResBytes)

	if e.Run(opts.CmdScript) != nil {
		err := handleInitialCheck(e, opts)
		if err != nil {
			return err
		}
	}

	return handlePolling(opts)
}

func handleInitialCheck(e *Executor, opts RunSyncOptions) error {
	errorData, err := GetErrorInfos(e.GetStderrValue())
	if err != nil {
		return errors.Wrap(err, "Failed to extract error code from JSON response")
	}

	if errorData.Error == "Conflict" {
		return errors.Wrap(errors.New(errorData.Description), "Command returned a conflict error.")
	} else {
		if !opts.IgnoreErrorOnFirstCall {
			return errors.Wrap(err, "Failed to execute BTP CLI (Sync)")
		}
	}
	return nil
}

func handlePolling(opts RunSyncOptions) error {
	// Save if we are now waiting things to be Ready
	waitingForReady := false

	retryCount := 0
	badRequestCount := 0
	maxRetries := 6
	maxBadRequests := 10

	// Poll to check completion
	timeoutDuration := time.Duration(opts.TimeoutSeconds) * time.Second
	pollIntervall := time.Duration(opts.PollInterval) * time.Second
	startTime := time.Now()

	log.Entry().Info("Checking command completion...")

	for time.Since(startTime) < timeoutDuration {
		if retryCount >= maxRetries {
			return errors.New("Maximum number of retries reached while polling for command completion")
		}
		if badRequestCount >= maxBadRequests {
			return errors.New("Too many bad request errors received while polling for command completion")
		}

		// Wait before the next check
		additionalTime := time.Duration(retryCount/3) * time.Minute
		additionalTime += time.Duration(badRequestCount/3) * time.Minute
		time.Sleep(pollIntervall + additionalTime)

		check := opts.CheckFunc()

		if check.successful {
			if check.done {
				log.Entry().Info("Command execution completed successfully!")
				return nil
			} else {
				waitingForReady = true
				retryCount = 0
				badRequestCount = 0
				log.Entry().Info("Command not yet completed, waiting for status ready...")
			}
		} else {
			err := handlePollingCheck(opts, check)
			if err != nil {
				return err
			}

			// Re-login before next check
			_ = opts.LoginFunc()

			if waitingForReady {
				log.Entry().Infof("Command was previously in progress, but now reports failure : Retry count %d.", retryCount)
				retryCount++
			} else {
				if check.errorData.Error == "BadRequest" {
					badRequestCount++
					log.Entry().Infof("Command not yet completed, checking again... : BadRequest count %d", badRequestCount)
				} else {
					log.Entry().Info("Command not yet completed, checking again...")
				}
			}
		}
	}

	return errors.New("Command did not completed within the timeout period")
}

func handlePollingCheck(opts RunSyncOptions, check CheckResponse) error {
	if check.errorData.Error == "Conflict" {
		return errors.Wrap(errors.New(check.errorData.Description), "Command check returned a conflict error.")
	}
	return nil
}

type Executor struct {
	Cmd command.Command
}

type btpRunner interface {
	Stdin(in io.Reader)
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	GetStdoutValue() string
	GetStderrValue() string
}

type ExecRunner interface {
	btpRunner
	Run(cmdScript []string) error
	RunSync(opts RunSyncOptions) error
}
