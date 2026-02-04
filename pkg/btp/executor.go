package btp

import (
	"bytes"
	"io"
	"strings"
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
	err := e.Run(opts.CmdScript)

	if err != nil {
		_, errorMessageCode, err := GetErrorInfos(e.GetStderrValue())
		if err != nil {
			return errors.Wrap(err, "Failed to extract error code from JSON response")
		}

		switch errorMessageCode {
		case "SERVICE_INSTANCE_NOT_FOUND":
			if strings.Contains(strings.Join(opts.CmdScript, " "), "create services/binding") {
				return errors.New("Service instance not found.")
			}
		case "INSTANCE_ALREADY_EXISTS":
			if strings.Contains(strings.Join(opts.CmdScript, " "), "create services/instance") {
				return errors.New("Service instance with the same name already exists.")
			}
		case "BINDING_ALREADY_EXISTS":
			if strings.Contains(strings.Join(opts.CmdScript, " "), "create services/binding") {
				return errors.New("Service binding with the same name already exists for the service instance.")
			}
		}

		if !opts.IgnoreErrorOnFirstCall {
			return errors.Wrap(err, "Failed to execute BTP CLI (Sync)")
		}
	}

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
			badRequestCount = 0

			if check.done {
				log.Entry().Info("Command execution completed successfully!")
				return nil
			} else {
				waitingForReady = true
				retryCount = 0
				log.Entry().Info("Command not yet completed, waiting for status ready...")
			}
		} else {
			if check.errorData.Error == "Conflict" {
				return errors.New("Command check returned a conflict error.")
			} else {
				switch check.errorMessageCode {
				case "MULTIPLE_BINDINGS_FOUND":
					return errors.New("Multiple service bindings found with the given name.")
				default:
					err := opts.LoginFunc()
					if err != nil {
						return errors.Wrap(err, "Failed to re-login during polling")
					}
				}
			}

			if waitingForReady {
				log.Entry().Infof("Command was previously in progress, but now reports failure : Retry count %d.", retryCount)
				retryCount++
			} else {
				if check.errorData.Error != "BadRequest" {
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
