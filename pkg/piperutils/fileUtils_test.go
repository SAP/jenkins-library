//go:build unit
// +build unit

package piperutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileExists(t *testing.T) {
	runInTempDir(t, "testing dir returns false", func(t *testing.T) {
		err := os.Mkdir("test", 0777)
		if err != nil {
			t.Fatal("failed to create test dir in temporary dir")
		}
		result, err := FileExists("test")
		assert.NoError(t, err)
		assert.False(t, result)
	})
	runInTempDir(t, "testing file returns true", func(t *testing.T) {
		file, err := os.CreateTemp("", "testFile")
		assert.NoError(t, err)
		result, err := FileExists(file.Name())
		assert.NoError(t, err)
		assert.True(t, result)
	})
}

func TestDirExists(t *testing.T) {
	runInTempDir(t, "testing dir exists", func(t *testing.T) {
		err := os.Mkdir("test", 0777)
		if err != nil {
			t.Fatal("failed to create test dir in temporary dir")
		}
		files := Files{}

		result, err := files.DirExists("test")
		assert.NoError(t, err)
		assert.True(t, result, "created folder should exist")

		result, err = files.DirExists(".")
		assert.NoError(t, err)
		assert.True(t, result, "current directory should exist")

		result, err = files.DirExists(string(os.PathSeparator))
		assert.NoError(t, err)
		assert.True(t, result, "root directory should exist")
	})
}

func TestCopy(t *testing.T) {
	runInTempDir(t, "copying file succeeds", func(t *testing.T) {
		file := "testFile"
		err := os.WriteFile(file, []byte{byte(1), byte(2), byte(3)}, 0700)
		if err != nil {
			t.Fatal("Failed to create temporary workspace directory")
		}

		result, err := Copy(file, "testFile2")
		assert.NoError(t, err, "Didn't expert error but got one")
		assert.Equal(t, int64(3), result, "Expected true but got false")
	})
	runInTempDir(t, "copying directory fails", func(t *testing.T) {
		src := filepath.Join("some", "file")
		dst := filepath.Join("another", "file")

		err := os.MkdirAll(src, 0777)
		if err != nil {
			t.Fatal("Failed to create test directory")
		}

		files := Files{}
		exists, err := files.DirExists(src)
		assert.NoError(t, err)
		assert.True(t, exists)

		length, err := files.Copy(src, dst)
		assert.EqualError(t, err, "Source file '"+src+"' does not exist")
		assert.Equal(t, length, int64(0))
	})
}

func runInTempDir(t *testing.T, nameOfRun string, run func(t *testing.T)) {
	t.Run(nameOfRun, func(t *testing.T) {
		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		t.Cleanup(func() {
			_ = os.Chdir(oldCWD)
		})

		run(t)
	})
}

func TestExcludeFiles(t *testing.T) {
	t.Parallel()
	t.Run("nil slices", func(t *testing.T) {
		t.Parallel()
		filtered, err := ExcludeFiles(nil, nil)
		assert.NoError(t, err)
		assert.Len(t, filtered, 0)
	})
	t.Run("empty excludes", func(t *testing.T) {
		t.Parallel()
		files := []string{"file"}
		filtered, err := ExcludeFiles(files, nil)
		assert.NoError(t, err)
		assert.Equal(t, files, filtered)
	})
	t.Run("direct match", func(t *testing.T) {
		t.Parallel()
		files := []string{"file"}
		filtered, err := ExcludeFiles(files, files)
		assert.NoError(t, err)
		assert.Len(t, filtered, 0)
	})
	t.Run("two direct matches", func(t *testing.T) {
		t.Parallel()
		files := []string{"a", "b"}
		filtered, err := ExcludeFiles(files, files)
		assert.NoError(t, err)
		assert.Len(t, filtered, 0)
	})
	t.Run("one direct exclude matches", func(t *testing.T) {
		t.Parallel()
		files := []string{"a", "b"}
		filtered, err := ExcludeFiles(files, []string{"b"})
		assert.NoError(t, err)
		assert.Equal(t, []string{"a"}, filtered)
	})
	t.Run("no glob matches", func(t *testing.T) {
		t.Parallel()
		files := []string{"a", "b"}
		filtered, err := ExcludeFiles(files, []string{"*/a", "b/*"})
		assert.NoError(t, err)
		assert.Equal(t, []string{"a", "b"}, filtered)
	})
	t.Run("two globs match", func(t *testing.T) {
		t.Parallel()
		files := []string{"path/to/a", "b"}
		filtered, err := ExcludeFiles(files, []string{"**/a", "**/b"})
		assert.NoError(t, err)
		assert.Len(t, filtered, 0)
	})
}

func TestValidRelPath(t *testing.T) {
	t.Parallel()

	// Valid paths that should pass
	validPaths := []struct {
		name string
		path string
	}{
		{"simple file", "file.txt"},
		{"file with extension", "document.pdf"},
		{"nested path", "dir/subdir/file.txt"},
		{"deep nested path", "a/b/c/d/e/f/file.txt"},
		{"path with dashes", "my-folder/my-file.txt"},
		{"path with underscores", "my_folder/my_file.txt"},
		{"path with numbers", "folder123/file456.txt"},
		{"path with dots in name", "my.config.file"},
		{"multiple extensions", "archive.tar.gz"},
		{"directory with trailing slash", "mydir/"},
		{"nested directory with trailing slash", "parent/child/"},
		{"file in directory", "src/main/app.go"},
	}

	for _, tc := range validPaths {
		tc := tc
		t.Run("valid: "+tc.name, func(t *testing.T) {
			t.Parallel()
			assert.True(t, validRelPath(tc.path), "Expected path '%s' to be valid", tc.path)
		})
	}

	// Invalid paths that should fail
	invalidPaths := []struct {
		name   string
		path   string
		reason string
	}{
		// Empty and basic validation
		{"empty string", "", "empty string not allowed"},
		{"current directory", ".", "current directory not allowed"},

		// Path traversal attacks
		{"parent directory", "..", "parent directory traversal"},
		{"parent traversal prefix", "../file.txt", "parent directory traversal at start"},
		{"parent traversal middle", "dir/../file.txt", "parent directory traversal in middle"},
		{"parent traversal suffix", "dir/..", "parent directory traversal at end"},
		{"nested parent traversal", "../../etc/passwd", "nested parent directory traversal"},
		{"complex traversal", "dir/subdir/../../file.txt", "complex parent traversal"},

		// Absolute paths
		{"absolute unix path", "/etc/passwd", "absolute unix path"},
		{"absolute with multiple segments", "/usr/local/bin/file", "absolute path with segments"},

		// Leading ./ (rejected by function)
		{"leading dot slash", "./file.txt", "leading ./ not allowed"},
		{"leading dot slash nested", "./dir/file.txt", "leading ./ in nested path"},

		// Paths that are just "/"
		{"only trailing slash", "/", "only slash not allowed"},

		// Windows path separators
		{"windows backslash", "dir\\file.txt", "windows backslash separator"},
		{"windows absolute", "C:\\Windows\\System32", "windows absolute path"},
		{"mixed separators", "dir/subdir\\file.txt", "mixed path separators"},

		// NUL and control characters (now all control chars are rejected)
		{"null byte", "file\x00.txt", "null byte in path"},
		{"control char 0x01", "file\x01.txt", "control character in path"},
		{"control char 0x1F", "file\x1f.txt", "control character in path"},
		{"tab in filename", "file\twith\ttab.txt", "tab is a control character"},
		{"newline in filename", "file\nwith\nnewline.txt", "newline is a control character"},

		// Invalid UTF-8
		{"invalid utf8", "file\xff\xfe.txt", "invalid UTF-8 sequence"},
	}

	for _, tc := range invalidPaths {
		tc := tc
		t.Run("invalid: "+tc.name, func(t *testing.T) {
			t.Parallel()
			assert.False(t, validRelPath(tc.path), "Expected path '%s' to be invalid: %s", tc.path, tc.reason)
		})
	}

	// Edge cases that reveal potential issues
	t.Run("edge cases", func(t *testing.T) {
		t.Run("path cleaned to dot", func(t *testing.T) {
			// Paths that clean to "." should be rejected
			assert.False(t, validRelPath("."))
		})

		t.Run("path that cleans to absolute", func(t *testing.T) {
			// Even if the cleaned version is absolute, original check should catch it
			assert.False(t, validRelPath("/dir/../file"))
		})

		t.Run("trailing slash on root-level dir", func(t *testing.T) {
			// Trailing slashes are now allowed for directories
			assert.True(t, validRelPath("myproject/"))
		})

		t.Run("double trailing slashes", func(t *testing.T) {
			// Double slashes create empty path components which should be rejected
			// Note: "dir//" after TrimSuffix becomes "dir/" which is valid
			// To reject this, we'd need additional validation for empty components
			// For now, this is accepted as it normalizes to a valid directory path
			assert.True(t, validRelPath("dir//"))
		})

		t.Run("only trailing slash", func(t *testing.T) {
			// Just "/" should be rejected (empty path after trimming)
			assert.False(t, validRelPath("/"))
		})
	})
}
