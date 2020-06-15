// +build !release

package mock

import (
	"fmt"
	"github.com/bmatcuk/doublestar"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var dirContent []byte

// FilesMock implements the functions from piperutils.Files with an in-memory file system.
type FilesMock struct {
	files        map[string]*[]byte
	removedFiles map[string]*[]byte
	currentDir   string
}

func (f *FilesMock) init() {
	if f.files == nil {
		f.files = map[string]*[]byte{}
	}
	if f.removedFiles == nil {
		f.removedFiles = map[string]*[]byte{}
	}
}

func (f *FilesMock) toAbsPath(path string) string {
	if !strings.HasPrefix(path, string(os.PathSeparator)) {
		path = string(os.PathSeparator) + filepath.Join(f.currentDir, path)
	}
	return path
}

// AddFile establishes the existence of a virtual file.
func (f *FilesMock) AddFile(path string, contents []byte) {
	f.init()
	f.files[f.toAbsPath(path)] = &contents
}

// AddDir establishes the existence of a virtual directory.
func (f *FilesMock) AddDir(path string) {
	f.init()
	f.files[f.toAbsPath(path)] = &dirContent
}

// HasFile returns true if the virtual file system contains an entry for the given path.
func (f *FilesMock) HasFile(path string) bool {
	_, exists := f.files[f.toAbsPath(path)]
	return exists
}

// HasRemovedFile returns true if the virtual file system at one point contained an entry for the given path,
// and it was removed via FileRemove().
func (f *FilesMock) HasRemovedFile(path string) bool {
	_, exists := f.removedFiles[f.toAbsPath(path)]
	return exists
}

// FileExists returns true if file content has been associated with the given path, false otherwise.
// Only relative paths are supported.
func (f *FilesMock) FileExists(path string) (bool, error) {
	if f.files == nil {
		return false, nil
	}
	content, exists := f.files[f.toAbsPath(path)]
	if !exists {
		return false, fmt.Errorf("'%s': %w", path, os.ErrNotExist)
	}
	return content != &dirContent, nil
}

// DirExists returns true, if the given path is a previously added directory, or a parent directory for any of the
// previously added files.
func (f *FilesMock) DirExists(path string) (bool, error) {
	path = f.toAbsPath(path)
	for entry, content := range f.files {
		var dirComponents []string
		if content == &dirContent {
			dirComponents = strings.Split(entry, string(os.PathSeparator))
		} else {
			dirComponents = strings.Split(filepath.Dir(entry), string(os.PathSeparator))
		}
		if len(dirComponents) > 0 {
			dir := ""
			for i, component := range dirComponents {
				if i == 0 {
					dir = component
				} else {
					dir = dir + string(os.PathSeparator) + component
				}
				if dir == path {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

// Copy checks if content has been associated with the given src path, and if so copies it under the given path dst.
func (f *FilesMock) Copy(src, dst string) (int64, error) {
	f.init()
	content, exists := f.files[f.toAbsPath(src)]
	if !exists || content == &dirContent {
		return 0, fmt.Errorf("cannot copy '%s': %w", src, os.ErrNotExist)
	}
	f.AddFile(dst, *content)
	return int64(len(*content)), nil
}

// FileRead returns the content previously associated with the given path via AddFile(), or an error if no
// content has been associated.
func (f *FilesMock) FileRead(path string) ([]byte, error) {
	f.init()
	content, exists := f.files[f.toAbsPath(path)]
	if !exists {
		return nil, fmt.Errorf("could not read '%s'", path)
	}
	// check if trying to open a directory for reading
	if content == &dirContent {
		return nil, fmt.Errorf("could not read '%s': %w", path, os.ErrInvalid)
	}
	return *content, nil
}

// FileWrite just forwards to AddFile(), i.e. the content is associated with the given path.
func (f *FilesMock) FileWrite(path string, content []byte, _ os.FileMode) error {
	// NOTE: FilesMock could be extended to have a set of paths for which FileWrite should fail.
	// This is why AddFile() exists separately, to differentiate the notion of setting up the mocking
	// versus implementing the methods from Files.
	f.AddFile(path, content)
	return nil
}

// FileRemove deletes the association of the given path with any content and records the removal of the file.
// If the path has not been registered before, it returns an error.
func (f *FilesMock) FileRemove(path string) error {
	if f.files == nil {
		return fmt.Errorf("the file '%s' does not exist: %w", path, os.ErrNotExist)
	}
	absPath := f.toAbsPath(path)
	content, exists := f.files[absPath]
	if !exists {
		return fmt.Errorf("the file '%s' does not exist: %w", path, os.ErrNotExist)
	}
	delete(f.files, absPath)
	f.removedFiles[absPath] = content
	return nil
}

// MkdirAll creates a directory in the in-memory file system, so that this path is established to exist.
func (f *FilesMock) MkdirAll(path string, _ os.FileMode) error {
	// NOTE: FilesMock could be extended to have a set of paths for which MkdirAll should fail.
	// This is why AddDir() exists separately, to differentiate the notion of setting up the mocking
	// versus implementing the methods from Files.
	f.AddDir(path)
	return nil
}

// Glob returns an array of path strings which match the given glob-pattern. Double star matching is supported.
func (f *FilesMock) Glob(pattern string) ([]string, error) {
	var matches []string
	if f.files == nil {
		return matches, nil
	}
	for path := range f.files {
		path = strings.TrimLeft(path, string(os.PathSeparator))
		matched, _ := doublestar.Match(pattern, path)
		if matched {
			matches = append(matches, path)
		}
	}
	// The order in f.files is not deterministic, this would result in flaky tests.
	sort.Strings(matches)
	return matches, nil
}

// Getwd returns the rooted current virtual working directory
func (f *FilesMock) Getwd() (string, error) {
	return f.toAbsPath(""), nil
}

// Chdir changes virtually in to the given directory.
// The directory needs to exist according to the files and directories via AddFile() and AddDirectory().
// The implementation is limited
func (f *FilesMock) Chdir(path string) error {
	if path == "." || path == "."+string(os.PathSeparator) {
		return nil
	}

	path = f.toAbsPath(path)

	exists, _ := f.DirExists(path)
	if !exists {
		return fmt.Errorf("failed to change current directory into '%s': %w", path, os.ErrNotExist)
	}

	f.currentDir = strings.TrimLeft(path, string(os.PathSeparator))
	return nil
}
