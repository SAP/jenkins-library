package codeql

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppendCodeQLPaths(t *testing.T) {
	t.Run("AppendCodeQLPaths", func(t *testing.T) {
		t.Run("creates_file_if_not_exists_with_both_keys", func(t *testing.T) {
			dir := t.TempDir()
			cfgPath := filepath.Join(dir, "codeql.yml")

			err := AppendCodeQLPaths(cfgPath,
				[]string{"src", "lib/utils"},
				[]string{"vendor", "**/*.gen.go"},
			)
			require.NoError(t, err)

			out := readFile(t, cfgPath)
			assert.Contains(t, out, "paths:")
			assert.Contains(t, out, "- src")
			assert.Contains(t, out, "- lib/utils")

			assert.Contains(t, out, "paths-ignore:")
			assert.Contains(t, out, "- vendor")
			assert.Contains(t, out, "- '**/*.gen.go'") // YAML encoder uses single quotes for '*'
		})

		t.Run("creates_file_if_not_exists_with_only_paths", func(t *testing.T) {
			dir := t.TempDir()
			cfgPath := filepath.Join(dir, "codeql.yml")

			err := AppendCodeQLPaths(cfgPath, []string{"a", "b"}, nil)
			require.NoError(t, err)

			out := readFile(t, cfgPath)
			assert.Contains(t, out, "paths:")
			assert.Contains(t, out, "- a")
			assert.Contains(t, out, "- b")
			assert.NotContains(t, out, "paths-ignore:")
		})

		t.Run("creates_file_if_not_exists_with_only_paths_ignore", func(t *testing.T) {
			dir := t.TempDir()
			cfgPath := filepath.Join(dir, "codeql.yml")

			err := AppendCodeQLPaths(cfgPath, nil, []string{"vendor"})
			require.NoError(t, err)

			out := readFile(t, cfgPath)
			assert.NotContains(t, out, "paths:\n")
			assert.Contains(t, out, "paths-ignore:")
			assert.Contains(t, out, "- vendor")
		})

		t.Run("appends_to_empty_yaml_creates_paths_and_paths_ignore", func(t *testing.T) {
			dir := t.TempDir()
			cfgPath := filepath.Join(dir, "codeql.yml")
			writeFile(t, cfgPath, "")

			err := AppendCodeQLPaths(cfgPath,
				[]string{"src", "lib/utils"},
				[]string{"vendor", "**/*.gen.go"},
			)
			require.NoError(t, err)

			out := readFile(t, cfgPath)

			assert.Contains(t, out, "paths:")
			assert.Contains(t, out, "- src")
			assert.Contains(t, out, "- lib/utils")

			assert.Contains(t, out, "paths-ignore:")
			assert.Contains(t, out, "- vendor")
			assert.Contains(t, out, "- '**/*.gen.go'")
		})

		t.Run("overwrites_on_type_mismatch_and_preserves_other_keys", func(t *testing.T) {
			dir := t.TempDir()
			cfgPath := filepath.Join(dir, "codeql.yml")
			// Intentionally set scalar for paths to force overwrite; and keep an existing ignore.
			initial := "paths: foo\npaths-ignore:\n  - old-ignore\n"
			writeFile(t, cfgPath, initial)

			err := AppendCodeQLPaths(cfgPath, []string{"new-path"}, nil)
			require.NoError(t, err)

			out := readFile(t, cfgPath)

			// 'paths' should now be a sequence containing only the new value
			assert.Contains(t, out, "paths:")
			assert.Contains(t, out, "- new-path")
			assert.NotContains(t, out, "foo")

			// paths-ignore should remain present (we didn't provide a new ignore list)
			assert.Contains(t, out, "paths-ignore:")
			assert.Contains(t, out, "- old-ignore")
		})

		t.Run("updates_only_paths_when_ignore_not_provided", func(t *testing.T) {
			dir := t.TempDir()
			cfgPath := filepath.Join(dir, "codeql.yml")
			initial := "paths:\n  - old\npaths-ignore:\n  - keep-ignore\n"
			writeFile(t, cfgPath, initial)

			err := AppendCodeQLPaths(cfgPath, []string{"new-a", "new-b"}, nil)
			require.NoError(t, err)

			out := readFile(t, cfgPath)
			assert.Contains(t, out, "paths:")
			assert.Contains(t, out, "- new-a")
			assert.Contains(t, out, "- new-b")
			assert.NotContains(t, out, "- old")

			assert.Contains(t, out, "paths-ignore:")
			assert.Contains(t, out, "- keep-ignore")
		})

		t.Run("updates_only_paths_ignore_when_paths_not_provided", func(t *testing.T) {
			dir := t.TempDir()
			cfgPath := filepath.Join(dir, "codeql.yml")
			initial := "paths:\n  - keep\npaths-ignore:\n  - old-ignore\n"
			writeFile(t, cfgPath, initial)

			err := AppendCodeQLPaths(cfgPath, nil, []string{"new-ignore-1", "**/*.gen.go"})
			require.NoError(t, err)

			out := readFile(t, cfgPath)
			assert.Contains(t, out, "paths:")
			assert.Contains(t, out, "- keep")

			assert.Contains(t, out, "paths-ignore:")
			assert.Contains(t, out, "- new-ignore-1")
			assert.Contains(t, out, "- '**/*.gen.go'")
			assert.NotContains(t, out, "- old-ignore")
		})

		t.Run("no_op_when_no_values_on_existing_file", func(t *testing.T) {
			dir := t.TempDir()
			cfgPath := filepath.Join(dir, "codeql.yml")
			initial := "paths:\n  - keep\npaths-ignore:\n  - keep-ignore\n"
			writeFile(t, cfgPath, initial)
			before := readFile(t, cfgPath)

			// Touch mtime differences robustly across filesystems.
			time.Sleep(10 * time.Millisecond)

			err := AppendCodeQLPaths(cfgPath, nil, nil)
			require.NoError(t, err)

			after := readFile(t, cfgPath)
			assert.Equal(t, before, after, "file content should be unchanged for no-op with nil slices")
		})

		t.Run("no_op_when_no_values_and_file_missing_does_not_create_file", func(t *testing.T) {
			dir := t.TempDir()
			cfgPath := filepath.Join(dir, "codeql.yml")

			err := AppendCodeQLPaths(cfgPath, []string{}, []string{})
			require.NoError(t, err)

			_, statErr := os.Stat(cfgPath)
			require.True(t, os.IsNotExist(statErr), "file should not be created when both inputs are empty")
		})

		t.Run("works_with_existing_empty_file", func(t *testing.T) {
			dir := t.TempDir()
			cfgPath := filepath.Join(dir, "codeql.yml")
			writeFile(t, cfgPath, "")

			err := AppendCodeQLPaths(cfgPath, []string{"p1"}, []string{"i1"})
			require.NoError(t, err)

			out := readFile(t, cfgPath)
			assert.Contains(t, out, "paths:")
			assert.Contains(t, out, "- p1")
			assert.Contains(t, out, "paths-ignore:")
			assert.Contains(t, out, "- i1")
		})
	})
}

func TestParsePaths(t *testing.T) {
	t.Run("ParsePaths", func(t *testing.T) {
		cases := []struct {
			name string
			in   string
			out  []string
		}{
			{
				name: "trims_and_skips_blanks",
				in:   " src \n\nlib\n  \nvendor  ",
				out:  []string{"src", "lib", "vendor"},
			},
			{
				name: "handles_single_line",
				in:   "only/one",
				out:  []string{"only/one"},
			},
			{
				name: "keeps_globs_and_spaces_inside_line",
				in:   "**/*.go\npath with space/\n./rel",
				out:  []string{"**/*.go", "path with space/", "./rel"},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				got := ParsePaths(tc.in)
				assert.Equal(t, tc.out, got)
			})
		}
	})
}

func TestWhich(t *testing.T) {
	t.Run("Which", func(t *testing.T) {
		// Save and restore PATH for isolation.
		origPath := os.Getenv("PATH")
		t.Cleanup(func() { _ = os.Setenv("PATH", origPath) })

		t.Run("nonexistent_returns_ErrNotExist", func(t *testing.T) {
			_, err := Which("definitely-not-a-real-binary-name")
			assert.ErrorIs(t, err, os.ErrNotExist)
		})

		t.Run("finds_executable_in_PATH", func(t *testing.T) {
			if runtime.GOOS == "windows" {
				t.Skip("executable bit semantics differ on Windows; test skipped")
			}
			dir := t.TempDir()
			bin := filepath.Join(dir, "mytool")
			writeExec(t, bin, "#!/bin/sh\necho ok\n")

			// Put our temp dir first on PATH
			require.NoError(t, os.Setenv("PATH", dir+string(os.PathListSeparator)+origPath))

			found, err := Which("mytool")
			require.NoError(t, err)

			// Compare against the resolved path (handles /var -> /private/var on macOS)
			expected, err := resolve(bin)
			require.NoError(t, err)
			assert.Equal(t, expected, found)
		})

		t.Run("resolves_symlinks", func(t *testing.T) {
			// Symlink behavior varies on Windows (needs admin/developer mode).
			if runtime.GOOS == "windows" {
				t.Skip("symlink creation is restricted on Windows; test skipped")
			}
			dir := t.TempDir()
			target := filepath.Join(dir, "realtool")
			link := filepath.Join(dir, "alias")

			writeExec(t, target, "#!/bin/sh\necho real\n")
			err := os.Symlink(target, link)
			if err != nil {
				t.Skipf("symlinks not supported in this environment: %v", err)
			}

			require.NoError(t, os.Setenv("PATH", dir))

			found, err := Which("alias")
			require.NoError(t, err)

			expected, err := resolve(target)
			require.NoError(t, err)
			assert.Equal(t, expected, found, "Which should resolve through the symlink to the real path")
		})

		t.Run("direct_path_with_separator", func(t *testing.T) {
			if runtime.GOOS == "windows" {
				t.Skip("executable bit semantics differ on Windows; test skipped")
			}
			dir := t.TempDir()
			bin := filepath.Join(dir, "runme")
			writeExec(t, bin, "#!/bin/sh\necho hi\n")

			found, err := Which(bin)
			require.NoError(t, err)

			expected, err := resolve(bin)
			require.NoError(t, err)
			assert.Equal(t, expected, found)
		})
	})
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)
}

func writeExec(t *testing.T, path, content string) {
	t.Helper()
	err := os.WriteFile(path, []byte(content), 0o755)
	require.NoError(t, err)

	// On some systems, writing then chmod is safer/more explicit.
	err = os.Chmod(path, 0o755)
	require.NoError(t, err)

	// Sanity: ensure it's not a directory and has any execute bit set (for non-Windows cases we test)
	info, err := os.Stat(path)
	require.NoError(t, err)
	require.False(t, info.IsDir())
	require.True(t, info.Mode()&0o111 != 0)
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	// Normalize Windows newlines just in case, so substring checks are reliable.
	return strings.ReplaceAll(string(b), "\r\n", "\n")
}
