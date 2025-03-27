package btp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	"gopkg.in/yaml.v2"
)

type Executer struct {
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

// Stdin ..
func (e *Executer) Stdin(stdin io.Reader) {
	e.Cmd.Stdin(stdin)
}

// Stdout ..
func (e *Executer) Stdout(stdout io.Writer) {
	e.Cmd.Stdout(stdout)
}

func (e *Executer) GetStdoutValue() string {
	return e.Cmd.GetStdout().(*bytes.Buffer).String()
}

func (e *Executer) Run(cmdScript string) (err error) {
	parts := strings.Fields(cmdScript)
	if err := e.Cmd.RunExecutable(parts[0], parts[1:]...); err != nil {
		return fmt.Errorf("Failed to execute BTP CLI: %w", err)
	}
	return nil
}

func (e *Executer) RunSync(cmdScript string, cmdCheck string, timeoutMin int, pollIntervalSec int, negativeCheck bool) (err error) {
	err = e.Run(cmdScript)
	/* if err != nil {
		return fmt.Errorf("Initial command execution failed: %w", err)
	} */

	// Poll to check completion
	timeoutDuration := time.Duration(timeoutMin) * time.Minute
	pollIntervall := time.Duration(pollIntervalSec) * time.Second
	startTime := time.Now()

	for time.Since(startTime) < timeoutDuration {
		fmt.Println("Checking command completion...")

		parts := strings.Fields(cmdCheck)
		err := e.Cmd.RunExecutable(parts[0], parts[1:]...)

		/* if err != nil {
			fmt.Println("Error checking completion: %w", err)
		} */

		outputStr := strings.TrimSpace(string(e.GetStdoutValue()))

		if err == nil && isCommandCompleted(outputStr, negativeCheck) {
			fmt.Println("Command execution completed successfully!")
			return nil
		}

		// Wait before the next check
		time.Sleep(pollIntervall)
	}

	return fmt.Errorf("Command did not completed within the timeout period")
}

func isCommandCompleted(output string, negativeCheck bool) bool {
	var lines []string = strings.Split(output, "\n")

	check := strings.Contains(lines[len(lines)-1], "OK") || strings.Contains(output, "COMPLETED") || strings.Contains(output, "SUCCEEDED")
	if negativeCheck {
		return !check
	}
	return check
}

func ConvertYAMLToJSON(yamlInput string, target interface{}) (string, error) {
	// Parse YAML into a generic map
	err := yaml.Unmarshal([]byte(yamlInput), target)
	if err != nil {
		return "", fmt.Errorf("Error unmarshaling YAML: %v", err)
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(target, "", "  ")
	if err != nil {
		return "", fmt.Errorf("Error marshaling JSON: %v", err)
	}

	return string(jsonData), nil
}
