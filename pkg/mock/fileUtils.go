//go:build !release
// +build !release

package mock

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
)

var dirContent []byte

const (
	defaultFileMode os.FileMode = 0o644
	defaultDirMode  os.FileMode = 0o755
)

type fileInfoMock struct {
	name  string
	mode  os.FileMode
	size  int64
	isDir bool
}

func (fInfo fileInfoMock) Name() string       { return fInfo.name }
func (fInfo fileInfoMock) Size() int64        { return fInfo.size }
func (fInfo fileInfoMock) Mode() os.FileMode  { return fInfo.mode }
func (fInfo fileInfoMock) ModTime() time.Time { return time.Time{} }
func (fInfo fileInfoMock) IsDir() bool        { return fInfo.isDir }
func (fInfo fileInfoMock) Sys() interface{}   { return nil }

type fileProperties struct {
	content *[]byte
	mode    os.FileMode
	isLink  bool
	target  string
}

// isDir returns true when the properties describe a directory entry.
func (p *fileProperties) isDir() bool {
	return p.content == &dirContent
}

// FilesMock implements the functions from piperutils.Files with an in-memory file system.
type FilesMock struct {
	files            map[string]*fileProperties
	writtenFiles     []string
	copiedFiles      map[string]string
	removedFiles     []string
	CurrentDir       string
	Separator        string
	FileExistsErrors map[string]error
	FileReadErrors   map[string]error
	FileWriteError   error
	FileWriteErrors  map[string]error
}

func (f *FilesMock) init() {
	if f.files == nil {
		f.files = map[string]*fileProperties{}
	}
	if f.Separator == "" {
		f.Separator = string(os.PathSeparator)
	}
	if f.copiedFiles == nil {
		f.copiedFiles = map[string]string{}
	}
}

// toAbsPath checks if the given path is relative, and if so converts it to an absolute path considering the
// current directory of the FilesMock.
// Relative segments such as "../" are currently NOT supported.
func (f *FilesMock) toAbsPath(path string) string {
	path = filepath.FromSlash(path)
	if path == "." {
		return f.Separator + f.CurrentDir
	}
	if !strings.HasPrefix(path, f.Separator) {
		if !strings.HasPrefix(f.CurrentDir, "/") {
			path = f.Separator + filepath.Join(f.CurrentDir, path)
		} else {
			path = filepath.Join(f.CurrentDir, path)
		}
	}
	return path
}

// AddFile establishes the existence of a virtual file.
// The file is added with mode 644.
func (f *FilesMock) AddFile(path string, contents []byte) {
	f.AddFileWithMode(path, contents, defaultFileMode)
}

// AddFileWithMode establishes the existence of a virtual file.
func (f *FilesMock) AddFileWithMode(path string, contents []byte, mode os.FileMode) {
	f.associateContent(path, &contents, mode)
}

// AddDir establishes the existence of a virtual directory.
// The directory is add with default mode 755.
func (f *FilesMock) AddDir(path string) {
	f.AddDirWithMode(path, defaultDirMode)
}

// AddDirWithMode establishes the existence of a virtual directory.
func (f *FilesMock) AddDirWithMode(path string, mode os.FileMode) {
	f.associateContent(path, &dirContent, mode)
}

// SHA256 returns a random SHA256
func (f *FilesMock) SHA256(path string) (string, error) {
	hash := sha256.New()
	return fmt.Sprintf("%x", string(hash.Sum(nil))), nil
}

// CurrentTime returns the current time as a fixed value
func (f *FilesMock) CurrentTime(format string) string {
	return "20220102-150405"
}

func (f *FilesMock) associateContent(path string, content *[]byte, mode os.FileMode) {
	f.init()
	path = f.toAbsPath(path)
	f.associateContentAbs(path, content, mode)
}

func (f *FilesMock) associateContentAbs(path string, content *[]byte, mode os.FileMode) {
	f.init()
	path = strings.ReplaceAll(path, "/", f.Separator)
	path = strings.ReplaceAll(path, "\\", f.Separator)
	if _, ok := f.files[path]; !ok {
		f.files[path] = &fileProperties{}
	}
	props := f.files[path]
	props.content = content
	props.mode = mode
}

// HasFile returns true if the virtual file system contains an entry for the given path.
func (f *FilesMock) HasFile(path string) bool {
	_, exists := f.files[f.toAbsPath(path)]
	return exists
}

// HasRemovedFile returns true if the virtual file system at one point contained an entry for the given path,
// and it was removed via FileRemove().
func (f *FilesMock) HasRemovedFile(path string) bool {
	return slices.Contains(f.removedFiles, f.toAbsPath(path))
}

// HasWrittenFile returns true if the virtual file system at one point contained an entry for the given path,
// and it was written via FileWrite().
func (f *FilesMock) HasWrittenFile(path string) bool {
	return slices.Contains(f.writtenFiles, f.toAbsPath(path))
}

// HasCopiedFile returns true if the virtual file system at one point contained an entry for the given source and destination,
// and it was written via CopyFile().
func (f *FilesMock) HasCopiedFile(src string, dest string) bool {
	return f.copiedFiles[f.toAbsPath(src)] == f.toAbsPath(dest)
}

// HasCreatedSymlink returns true if the virtual file system has a symlink with a specific target.
func (f *FilesMock) HasCreatedSymlink(oldname, newname string) bool {
	if f.files == nil {
		return false
	}
	props, exists := f.files[f.toAbsPath(newname)]
	if !exists {
		return false
	}
	return props.isLink && props.target == oldname
}

// FileExists returns true if file content has been associated with the given path, false otherwise.
// Only relative paths are supported.
func (f *FilesMock) FileExists(path string) (bool, error) {
	if f.FileExistsErrors[path] != nil {
		return false, f.FileExistsErrors[path]
	}
	if f.files == nil {
		return false, nil
	}
	props, exists := f.files[f.toAbsPath(path)]
	if !exists {
		return false, nil
	}
	return props.content != &dirContent, nil
}

// DirExists returns true, if the given path is a previously added directory, or a parent directory for any of the
// previously added files.
func (f *FilesMock) DirExists(path string) (bool, error) {
	path = f.toAbsPath(path)
	if path == "." || path == "."+f.Separator || path == f.Separator {
		// The current folder, or the root folder always exist
		return true, nil
	}
	for entry, props := range f.files {
		var dirComponents []string
		if props.isDir() {
			dirComponents = strings.Split(entry, f.Separator)
		} else {
			dirComponents = strings.Split(filepath.Dir(entry), f.Separator)
		}
		if len(dirComponents) > 0 {
			dir := ""
			for i, component := range dirComponents {
				if i == 0 {
					dir = component
				} else {
					dir = dir + f.Separator + component
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
	props, exists := f.files[f.toAbsPath(src)]
	if !exists || props.isDir() {
		return 0, fmt.Errorf("cannot copy '%s': %w", src, os.ErrNotExist)
	}
	f.AddFileWithMode(dst, *props.content, props.mode)
	f.copiedFiles[f.toAbsPath(src)] = f.toAbsPath(dst)
	return int64(len(*props.content)), nil
}

// Move moves a file to the given destination
func (f *FilesMock) Move(src, dst string) error {
	if exists, err := f.FileExists(src); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("file doesn't exist: %s", src)
	}

	if _, err := f.Copy(src, dst); err != nil {
		return err
	}

	return f.FileRemove(src)
}

// FileRead returns the content previously associated with the given path via AddFile(), or an error if no
// content has been associated.
func (f *FilesMock) FileRead(path string) ([]byte, error) {
	f.init()
	if err := f.FileReadErrors[path]; err != nil {
		return nil, err
	}
	props, exists := f.files[f.toAbsPath(path)]
	if !exists {
		return nil, fmt.Errorf("could not read '%s'", path)
	}
	// check if trying to open a directory for reading
	if props.isDir() {
		return nil, fmt.Errorf("could not read '%s': %w", path, os.ErrInvalid)
	}
	return *props.content, nil
}

// ReadFile can be used as replacement for os.ReadFile in a compatible manner
func (f *FilesMock) ReadFile(name string) ([]byte, error) {
	return f.FileRead(name)
}

// FileWrite just forwards to AddFile(), i.e. the content is associated with the given path.
func (f *FilesMock) FileWrite(path string, content []byte, mode os.FileMode) error {
	if f.FileWriteError != nil {
		return f.FileWriteError
	}
	if f.FileWriteErrors[path] != nil {
		return f.FileWriteErrors[path]
	}
	f.init()
	f.writtenFiles = append(f.writtenFiles, f.toAbsPath(path))
	f.AddFileWithMode(path, content, mode)
	return nil
}

// WriteFile can be used as replacement for os.WriteFile in a compatible manner
func (f *FilesMock) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return f.FileWrite(filename, data, perm)
}

// RemoveAll is a proxy for FileRemove
func (f *FilesMock) RemoveAll(path string) error {
	return f.FileRemove(path)
}

// FileRemove deletes the association of the given path with any content and records the removal of the file.
// If the path has not been registered before, it returns an error.
func (f *FilesMock) FileRemove(path string) error {
	if f.files == nil {
		return fmt.Errorf("the file '%s' does not exist: %w", path, os.ErrNotExist)
	}
	absPath := f.toAbsPath(path)
	props, exists := f.files[absPath]

	// If there is no leaf-entry in the map, path may be a directory, but implicitly it cannot be empty
	if !exists {
		dirExists, _ := f.DirExists(path)
		if dirExists {
			return fmt.Errorf("the directory '%s' is not empty", path)
		}
		return fmt.Errorf("the file '%s' does not exist: %w", path, os.ErrNotExist)
	} else if props.isDir() {
		// Check if the directory is not empty re-using the Glob() implementation
		entries, _ := f.Glob(path + f.Separator + "*")
		if len(entries) > 0 {
			return fmt.Errorf("the directory '%s' is not empty", path)
		}
	}

	delete(f.files, absPath)
	f.removedFiles = append(f.removedFiles, absPath)

	// Make sure the parent directory still exists, if it only existed via this one entry
	leaf := filepath.Base(absPath)
	absPath = strings.TrimSuffix(absPath, f.Separator+leaf)
	if absPath != f.Separator {
		relPath := strings.TrimPrefix(absPath, f.Separator+f.CurrentDir+f.Separator)
		dirExists, _ := f.DirExists(relPath)
		if !dirExists {
			f.AddDir(relPath)
		}
	}

	return nil
}

// FileRename changes the path under which content is associated in the virtual file system.
// Only leaf-entries are supported as of yet.
func (f *FilesMock) FileRename(oldPath, newPath string) error {
	if f.files == nil {
		return fmt.Errorf("the file '%s' does not exist: %w", oldPath, os.ErrNotExist)
	}

	oldAbsPath := f.toAbsPath(oldPath)
	props, exists := f.files[oldAbsPath]
	// If there is no leaf-entry in the map, path may be a directory.
	// We only support renaming leaf-entries for now.
	if !exists {
		return fmt.Errorf("renaming file '%s' is not supported, since it does not exist, "+
			"or is not a leaf-entry", oldPath)
	}

	if oldPath == newPath {
		return nil
	}

	newAbsPath := f.toAbsPath(newPath)
	_, exists = f.files[newAbsPath]
	// Fail if the target path already exists
	if exists {
		return fmt.Errorf("cannot rename '%s', target path '%s' already exists", oldPath, newPath)
	}

	delete(f.files, oldAbsPath)
	f.files[newAbsPath] = props
	return nil
}

// TempDir create a temp-styled directory in the in-memory, so that this path is established to exist.
func (f *FilesMock) TempDir(baseDir string, pattern string) (string, error) {
	if len(baseDir) == 0 {
		baseDir = "/tmp"
	}

	tmpDir := baseDir

	if pattern != "" {
		tmpDir = fmt.Sprintf("%s/%stest", baseDir, pattern)
	}

	err := f.MkdirAll(tmpDir, 0o755)
	if err != nil {
		return "", err
	}
	return tmpDir, nil
}

// MkdirAll creates a directory in the in-memory file system, so that this path is established to exist.
func (f *FilesMock) MkdirAll(path string, mode os.FileMode) error {
	// NOTE: FilesMock could be extended to have a set of paths for which MkdirAll should fail.
	// This is why AddDirWithMode() exists separately, to differentiate the notion of setting up
	// the mocking versus implementing the methods from Files.
	f.AddDirWithMode(path, mode)
	return nil
}

// Glob returns an array of path strings which match the given glob-pattern. Double star matching is supported.
func (f *FilesMock) Glob(pattern string) ([]string, error) {
	var matches []string
	if f.files == nil {
		return matches, nil
	}
	for path := range f.files {
		path = strings.TrimLeft(path, f.Separator)
		matched, _ := doublestar.PathMatch(pattern, path)
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
	f.init()
	return f.toAbsPath(""), nil
}

// Chdir changes virtually into the given directory.
// The directory needs to exist according to the files and directories via AddFile() and AddDirectory().
// The implementation does not support relative path components such as "..".
func (f *FilesMock) Chdir(path string) error {
	path = filepath.FromSlash(path)
	if path == "." || path == "."+f.Separator {
		return nil
	}

	path = f.toAbsPath(path)

	exists, _ := f.DirExists(path)
	if !exists {
		return fmt.Errorf("failed to change current directory into '%s': %w", path, os.ErrNotExist)
	}

	f.CurrentDir = strings.TrimLeft(path, f.Separator)
	return nil
}

// Stat returns an approximated os.FileInfo. For files, it returns properties that have been associated
// via the setup methods. For directories it depends. If a directory exists only implicitly, because
// it is the parent of an added file, default values will be reflected in the file info.
func (f *FilesMock) Stat(path string) (os.FileInfo, error) {
	props, exists := f.files[f.toAbsPath(path)]
	if !exists {
		// Check if this folder exists implicitly
		isDir, err := f.DirExists(path)
		if err != nil {
			return nil, fmt.Errorf("internal error inside mock: %w", err)
		}
		if !isDir {
			return nil, &os.PathError{
				Op:   "stat",
				Path: path,
				Err:  fmt.Errorf("no such file or directory"),
			}
		}

		// we claim default umask, as no properties are stored for implicit folders
		props = &fileProperties{
			mode:    defaultDirMode,
			content: &dirContent,
		}
	}

	return fileInfoMock{
		name:  filepath.Base(path),
		mode:  props.mode,
		size:  int64(len(*props.content)),
		isDir: props.isDir(),
	}, nil
}

func (f *FilesMock) Lstat(path string) (os.FileInfo, error) {
	return f.Stat(path)
}

// Chmod changes the file mode for the entry at the given path
func (f *FilesMock) Chmod(path string, mode os.FileMode) error {
	props, exists := f.files[f.toAbsPath(path)]
	if exists {
		props.mode = mode
		return nil
	}

	// Check if the dir exists implicitly
	isDir, err := f.DirExists(path)
	if err != nil {
		return fmt.Errorf("internal error inside mock: %w", err)
	}
	if !isDir {
		return fmt.Errorf("chmod: %s: No such file or directory", path)
	}

	if mode != defaultDirMode {
		// we need to create properties to store the mode
		f.AddDirWithMode(path, mode)
	}

	return nil
}

func (f *FilesMock) Chown(path string, uid, gid int) error {
	return nil
}

func (f *FilesMock) Abs(path string) (string, error) {
	f.init()
	return f.toAbsPath(path), nil
}

func (f *FilesMock) Symlink(oldname, newname string) error {
	if f.FileWriteError != nil {
		return f.FileWriteError
	}

	if f.FileWriteErrors[newname] != nil {
		return f.FileWriteErrors[newname]
	}

	parentExists, err := f.DirExists(filepath.Dir(newname))
	if err != nil {
		return err
	}
	if !parentExists {
		return fmt.Errorf("failed to create symlink: parent directory %s doesn't exist", filepath.Dir(newname))
	}

	f.init()

	f.files[newname] = &fileProperties{
		isLink:  true,
		target:  oldname,
		content: &[]byte{},
	}

	return nil
}

// CreateArchive creates in memory tar.gz archive, with the content provided.
func (f *FilesMock) CreateArchive(content map[string][]byte) ([]byte, error) {
	if len(content) == 0 {
		return nil, errors.New("mock archive content must not be empty")
	}

	buf := bytes.NewBuffer(nil)
	gw := gzip.NewWriter(buf)
	tw := tar.NewWriter(gw)

	for fileName, fileContent := range content {
		err := tw.WriteHeader(&tar.Header{
			Name:     fileName,
			Size:     int64(len(fileContent)),
			Typeflag: tar.TypeReg,
		})

		if err != nil {
			return nil, err
		}

		_, err = tw.Write(fileContent)
		if err != nil {
			return nil, err
		}
	}

	err := tw.Close()
	if err != nil {
		return nil, err
	}

	err = gw.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// FileMock can be used in places where a io.Closer, io.StringWriter or io.Writer is expected.
// It is the concrete type returned from FilesMock.OpenFile()
type FileMock struct {
	absPath string
	files   *FilesMock
	content []byte
	buf     io.Reader
}

// Reads the content of the mock
func (f *FileMock) Read(b []byte) (n int, err error) {
	return f.buf.Read(b)
}

// Close mocks freeing the associated OS resources.
func (f *FileMock) Close() error {
	f.files = nil
	return nil
}

// WriteString converts the passed string to a byte array and forwards to Write().
func (f *FileMock) WriteString(s string) (n int, err error) {
	return f.Write([]byte(s))
}

// Write appends the provided byte array to the end of the current virtual file contents.
// It fails if the FileMock has been closed already, but it does not fail in case the path
// has already been removed from the FilesMock instance that created this FileMock.
// In this situation, the written contents will not become visible in the FilesMock.
func (f *FileMock) Write(p []byte) (n int, err error) {
	if f.files == nil {
		return 0, fmt.Errorf("file is closed")
	}

	f.content = append(f.content, p...)

	// It is not an error to write to a file that has been removed.
	// The kernel does reference counting, as long as someone has the file still opened,
	// it can be written to (and that entity can also still read it).
	properties, exists := f.files.files[f.absPath]
	if exists && properties.content != &dirContent {
		properties.content = &f.content
	}

	return len(p), nil
}

// OpenFile mimics the behavior os.OpenFile(), but it cannot return an instance of the os.File struct.
// Instead, it returns a pointer to a FileMock instance, which implements a number of the same methods as os.File.
// The flag parameter is checked for os.O_CREATE and os.O_APPEND and behaves accordingly.
func (f *FilesMock) OpenFile(path string, flag int, perm os.FileMode) (*FileMock, error) {
	if (f.files == nil || !f.HasFile(path)) && flag&os.O_CREATE == 0 {
		return nil, fmt.Errorf("the file '%s' does not exist: %w", path, os.ErrNotExist)
	}
	f.init()
	absPath := f.toAbsPath(path)
	properties, exists := f.files[absPath]
	if exists && properties.content == &dirContent {
		return nil, fmt.Errorf("opening directory not supported")
	}
	if !exists && flag&os.O_CREATE != 0 {
		f.associateContentAbs(absPath, &[]byte{}, perm)
		properties = f.files[absPath]
	}

	file := FileMock{
		absPath: absPath,
		files:   f,
		content: *properties.content,
	}

	if flag&os.O_TRUNC != 0 || flag&os.O_CREATE != 0 {
		file.content = []byte{}
		properties.content = &file.content
	}

	file.buf = bytes.NewBuffer(file.content)

	return &file, nil
}

func (f *FilesMock) Open(name string) (io.ReadWriteCloser, error) {
	return f.OpenFile(name, os.O_RDONLY, 0)
}

func (f *FilesMock) Create(name string) (io.ReadWriteCloser, error) {
	return f.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o666)
}

type FilesMockRelativeGlob struct {
	*FilesMock
}

// Glob of FilesMockRelativeGlob cuts current directory path part from files if pattern is relative
func (f *FilesMockRelativeGlob) Glob(pattern string) ([]string, error) {
	var matches []string
	if f.files == nil {
		return matches, nil
	}
	for path := range f.files {
		if !filepath.IsAbs(pattern) {
			path = strings.TrimLeft(path, f.Separator+f.CurrentDir)
		}
		path = strings.TrimLeft(path, f.Separator)
		matched, _ := doublestar.PathMatch(pattern, path)
		if matched {
			matches = append(matches, path)
		}
	}
	// The order in f.files is not deterministic, this would result in flaky tests.
	sort.Strings(matches)
	return matches, nil
}

func (f *FilesMock) Readlink(name string) (string, error) {
	properties, ok := f.files[name]
	if ok && properties.isLink {
		return properties.target, nil
	}
	return "", fmt.Errorf("could not retrieve target for %s", name)
}
