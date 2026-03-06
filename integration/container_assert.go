//go:build integration
// +build integration

package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

// ContainerAssertions extends testify/assert with better output for container output assertions.
type ContainerAssertions struct {
	*assert.Assertions
	t *testing.T
}

// NewContainerAssert creates a ContainerAssertions instance for the given test.
// Use this in integration tests that work with container output.
func NewContainerAssert(t *testing.T) *ContainerAssertions {
	return &ContainerAssertions{
		Assertions: assert.New(t),
		t:          t,
	}
}

// Contains asserts that the container output contains the expected substring.
func (ca *ContainerAssertions) Contains(output, expected string, msgAndArgs ...any) bool {
	ca.t.Helper()

	if !strings.Contains(output, expected) {
		// Custom error message with properly formatted output (preserves newlines)
		msg := fmt.Sprintf("Expected '%s' to contain %s", output, expected)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + "\n\n" + msg
		}
		ca.t.Error(msg)
		return false
	}
	return true
}

// NotContains asserts that the container output does NOT contain the given substring.
func (ca *ContainerAssertions) NotContains(output, notExpected string, msgAndArgs ...any) bool {
	ca.t.Helper()

	if strings.Contains(output, notExpected) {
		// Custom error message with properly formatted output (preserves newlines)
		msg := fmt.Sprintf("Expected '%s' to not contain %s", output, notExpected)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + "\n\n" + msg
		}
		ca.t.Error(msg)
		return false
	}
	return true
}

// FileExists asserts that one or more files exist in the container.
// It uses the stat command to check file existence.
func (ca *ContainerAssertions) FileExists(container testcontainers.Container, paths ...string) bool {
	ca.t.Helper()

	if len(paths) == 0 {
		ca.t.Error("FileExists requires at least one path")
		return false
	}

	ctx := context.Background()
	cmd := append([]string{"stat"}, paths...)

	code, reader, err := container.Exec(ctx, cmd)
	output, _ := io.ReadAll(reader)

	if !ca.Assertions.NoError(err, "Failed to execute stat command: %v", err) {
		return false
	}

	if !ca.Assertions.Equal(0, code, "One or more files do not exist: %v\nstat output:\n%s", paths, string(output)) {
		return false
	}

	return true
}
