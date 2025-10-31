package python

import (
	"fmt"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
)

func getBinary(virtualEnv string, binary string) string {
	if len(virtualEnv) > 0 {
		return filepath.Join(virtualEnv, "bin", binary)
	}
	return binary
}

func CreateVirtualEnvironment(
	executeFn func(executable string, params ...string) error,
	removeFn func(executable string) error,
	virtualEnv string,
) (func(), error) {
	exitHandler := func() {
		if err := removeFn(virtualEnv); err != nil {
			log.Entry().Debugf("failed to remove virtual environment %s: %v", virtualEnv, err)
		}
	}

	// Implementation for creating a virtual environment
	if err := executeFn("python3", "-m", "venv", virtualEnv); err != nil {
		return exitHandler, fmt.Errorf("failed to create virtual environment %s: %w", virtualEnv, err)
	}

	if err := executeFn("bash", "-c", fmt.Sprintf("source %s", filepath.Join(virtualEnv, "bin", "activate"))); err != nil {
		return exitHandler, fmt.Errorf("failed to activate virtual environment %s: %w", virtualEnv, err)
	}

	pipPath := filepath.Join(virtualEnv, "bin", "pip")
	if err := executeFn(pipPath, "install", "--upgrade", "pip", "build", "wheel", "setuptools"); err != nil {
		return exitHandler, fmt.Errorf("failed to activate virtual environment %s: %w", virtualEnv, err)
	}

	return exitHandler, nil
}
