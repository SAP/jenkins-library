package telesiactl

import (
	"fmt"
	"path/filepath"
	"runtime"
)

const workspaceBaseDir = ".pipeline/tools/telesiactl"

// FileChecker provides access to workspace binary paths.
type FileChecker interface {
	FileExists(filename string) (bool, error)
}

// ResolveBinary returns the first available telesiactl workspace binary.
func ResolveBinary(files FileChecker) (string, bool, error) {
	for _, binary := range BinaryCandidates() {
		exists, err := files.FileExists(binary)
		if err != nil {
			return "", false, fmt.Errorf("failed to check telesiactl binary '%s': %w", binary, err)
		}
		if exists {
			return binary, true, nil
		}
	}

	return "", false, nil
}

// BinaryCandidates returns the platform-specific binary path.
func BinaryCandidates() []string {
	platformPath := filepath.Join(workspaceBaseDir, runtime.GOOS+"-"+runtime.GOARCH, "telesiactl")

	return []string{platformPath}
}
