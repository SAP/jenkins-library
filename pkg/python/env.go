package python

import (
	"fmt"
	"path/filepath"
)

func CreateVirtualEnvironment(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
) error {
	// Implementation for creating a virtual environment
	if err := executeFn("python3", "-m", "venv", virtualEnv); err != nil {
		return fmt.Errorf("failed to create virtual environment %s: %w", virtualEnv, err)
	}
	if err := executeFn("bash", "-c", fmt.Sprintf("source %s", filepath.Join(virtualEnv, "bin", "activate"))); err != nil {
		return fmt.Errorf("failed to activate virtual environment %s: %w", virtualEnv, err)
	}
	return nil
}

func RemoveVirtualEnvironment(
	removeFn func(executable string) error,
	virtualEnv string,
) error {
	if err := removeFn(virtualEnv); err != nil {
		return fmt.Errorf("failed to remove virtual environment %s: %w", virtualEnv, err)
	}
	return nil
}
