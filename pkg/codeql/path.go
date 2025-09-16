package codeql

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// AppendCodeQLPaths updates the CodeQL config YAML with new paths/paths-ignore.
func AppendCodeQLPaths(cfgPath string, scanPaths, ignorePaths []string) error {
	if len(scanPaths) == 0 && len(ignorePaths) == 0 {
		// if both paths are empty - do not touch anything
		return nil
	}
	var cfg map[string]any

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		// file exists; unmarshal if non-empty
		if len(data) > 0 {
			if err = yaml.Unmarshal(data, &cfg); err != nil {
				return err
			}
		}
	}
	if cfg == nil {
		// start from empty config if file doesn't exist or empty
		cfg = make(map[string]any)
	}

	if len(scanPaths) != 0 {
		cfg["paths"] = scanPaths
	}
	if len(ignorePaths) != 0 {
		cfg["paths-ignore"] = ignorePaths
	}

	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(cfgPath, out, 0o644)
}

func ParsePaths(pathsStr string) []string {
	var paths []string
	patterns := strings.Split(pathsStr, "\n")
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		paths = append(paths, p)
	}
	return paths
}

// Which finds the first executable in PATH and resolves symlinks.
func Which(name string) (string, error) {
	// If the name already has a path separator, check it directly.
	if strings.ContainsRune(name, filepath.Separator) {
		if isExecutable(name) {
			return resolve(name)
		}
		return "", os.ErrNotExist
	}

	// Search in PATH
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" {
			dir = "."
		}
		candidate := filepath.Join(dir, name)
		if isExecutable(candidate) {
			return resolve(candidate)
		}
	}
	return "", os.ErrNotExist
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}
	// POSIX: check execute bit
	return info.Mode()&0o111 != 0
}

func resolve(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}
