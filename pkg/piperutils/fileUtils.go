package piperutils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/bmatcuk/doublestar"
)

const maxFileSize int64 = 2 * 1024 * 1024 * 1024 // 2 GB

// FileUtils ...
type FileUtils interface {
	Abs(path string) (string, error)
	DirExists(path string) (bool, error)
	FileExists(filename string) (bool, error)
	Copy(src, dest string) (int64, error)
	Move(src, dest string) error
	FileRead(path string) ([]byte, error)
	ReadFile(path string) ([]byte, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
	WriteFile(path string, content []byte, perm os.FileMode) error
	FileRemove(path string) error
	MkdirAll(path string, perm os.FileMode) error
	Chmod(path string, mode os.FileMode) error
	Chown(path string, uid, gid int) error
	Glob(pattern string) (matches []string, err error)
	Chdir(path string) error
	TempDir(string, string) (string, error)
	RemoveAll(string) error
	FileRename(string, string) error
	Getwd() (string, error)
	Symlink(oldname string, newname string) error
	SHA256(path string) (string, error)
	CurrentTime(format string) string
	Open(name string) (io.ReadWriteCloser, error)
	Create(name string) (io.ReadWriteCloser, error)
	Readlink(name string) (string, error)
	Stat(path string) (os.FileInfo, error)
	Lstat(path string) (os.FileInfo, error)
}

// Files ...
type Files struct{}

// TempDir creates a temporary directory
func (f Files) TempDir(dir, pattern string) (name string, err error) {
	if len(dir) == 0 {
		// lazy init system temp dir in case it doesn't exist
		if exists, _ := f.DirExists(os.TempDir()); !exists {
			f.MkdirAll(os.TempDir(), 0o666)
		}
	}

	return os.MkdirTemp(dir, pattern)
}

// FileExists returns true if the file system entry for the given path exists and is not a directory.
func (f Files) FileExists(filename string) (bool, error) {
	return FileExists(filename)
}

// FileExists returns true if the file system entry for the given path exists and is not a directory.
func FileExists(filename string) (bool, error) {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return !info.IsDir(), nil
}

// DirExists returns true if the file system entry for the given path exists and is a directory.
func (f Files) DirExists(path string) (bool, error) {
	info, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return info.IsDir(), nil
}

// Copy ...
func (f Files) Copy(src, dst string) (int64, error) {
	return Copy(src, dst)
}

// Move will move files from src to dst
func (f Files) Move(src, dst string) error {
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

// Chmod is a wrapper for os.Chmod().
func (f Files) Chmod(path string, mode os.FileMode) error {
	return os.Chmod(path, mode)
}

// Chown is a recursive wrapper for os.Chown().
func (f Files) Chown(path string, uid, gid int) error {
	return filepath.WalkDir(path, func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		return os.Chown(name, uid, gid)
	})
}

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
// from https://golangcode.com/unzip-files-in-go/ with the following license:
// MIT License
//
// # Copyright (c) 2017 Edd Turtle
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
func Unzip(src, dest string) ([]string, error) {
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer func() { _ = r.Close() }()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			err := os.MkdirAll(fpath, os.ModePerm)
			if err != nil {
				return filenames, fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = CopyData(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		_ = outFile.Close()
		_ = rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

// Untar will decompress a gzipped archive and then untar it, moving all files and folders
// within the tgz file (parameter 1) to an output directory (parameter 2).
// some tar like the one created from npm have an addtional package folder which need to be removed during untar
// stripComponent level acts the same like in the tar cli with level 1 corresponding to elimination of parent folder
// stripComponentLevel = 1 -> parentFolder/someFile.Txt -> someFile.Txt
// stripComponentLevel = 2 -> parentFolder/childFolder/someFile.Txt -> someFile.Txt
// when stripCompenent in 0 the untar will retain the original tar folder structure
// when stripCompmenet is greater than 0 the expectation is all files must be under that level folder and if not there is a hard check and failure condition
func Untar(src string, dest string, stripComponentLevel int) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("unable to open src: %v", err)
	}
	defer file.Close()

	if b, err := isFileGzipped(src); err == nil && b {
		zr, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("requires gzip-compressed body: %v", err)
		}

		return untar(zr, dest, stripComponentLevel)
	}

	return untar(file, dest, stripComponentLevel)
}

func untar(r io.Reader, dir string, level int) error {
	madeDir := map[string]bool{}
	tr := tar.NewReader(r)

	for {
		f, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("tar error: %v", err)
		}
		// Strip leading "/" to make the path relative
		f.Name, _ = strings.CutPrefix(f.Name, "/")
		if !validRelPath(f.Name) { // blocks path traversal attacks
			return fmt.Errorf("tar contained invalid name error %q", f.Name)
		}
		rel := filepath.FromSlash(f.Name)

		// when level X folder(s) needs to be removed we first check that the rel path must have atleast X or greater than X pathseperatorserr
		// or else we might end in index out of range
		if level > 0 {
			if strings.Count(rel, string(os.PathSeparator)) < level {
				return fmt.Errorf("files %q in tarball archive not under level %v", f.Name, level)
			}
			relSplit := strings.SplitN(rel, string(os.PathSeparator), level+1)
			rel = relSplit[level]
		}

		abs := filepath.Join(dir, rel)
		mode := f.FileInfo().Mode()

		switch {
		case mode.IsRegular():
			// Make the directory. This is redundant because it should
			// already be made by a directory entry in the tar
			// beforehand. Thus, don't check for errors; the next
			// write will fail with the same error.
			parent := filepath.Dir(abs)
			if !madeDir[parent] {
				if err = os.MkdirAll(parent, 0o755); err != nil {
					return err
				}
				madeDir[parent] = true
			}

			wf, err := os.OpenFile(abs, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode.Perm())
			if err != nil {
				return fmt.Errorf("open failed for %s: %v", abs, err)
			}

			// Reject if header size exceeds limit.
			if f.Size > maxFileSize {
				_ = wf.Close()
				_ = os.Remove(abs)
				log.Entry().Warnf("Rejecting file %s: exceeds maximum allowed size of %d bytes (2GB limit)", abs, maxFileSize)
				return fmt.Errorf("file %s exceeds maximum allowed size of %d bytes: actual %d", abs, maxFileSize, f.Size)
			}
			// Copy exactly the entry size (guarded by the header check above).
			n, err := io.CopyN(wf, tr, f.Size)
			if err != nil && err != io.EOF {
				_ = wf.Close()
				_ = os.Remove(abs)
				return fmt.Errorf("error writing to %s: %v", abs, err)
			}
			if n != f.Size {
				_ = os.Remove(abs)
				return fmt.Errorf("only wrote %d bytes to %s; expected %d", n, abs, f.Size)
			}

			if err = wf.Sync(); err != nil {
				_ = wf.Close()
				_ = os.Remove(abs)
				return fmt.Errorf("sync failed for %s: %v", abs, err)
			}

			if err = wf.Close(); err != nil {
				_ = os.Remove(abs)
				return fmt.Errorf("close failed for %s: %v", abs, err)
			}
		case mode.IsDir():
			if err = os.MkdirAll(abs, 0o755); err != nil {
				return err
			}
			madeDir[abs] = true
		case mode&fs.ModeSymlink != 0:
			if !validRelPath(f.Linkname) {
				return fmt.Errorf("tar symlink %q has invalid target %q", f.Name, f.Linkname)
			}
			if err = os.Symlink(f.Linkname, abs); err != nil {
				return err
			}
		default:
			return fmt.Errorf("tar file entry %s contained unsupported file type %v", f.Name, mode)
		}
	}
	return nil
}

// isFileGzipped checks the first 3 bytes of the given file to determine if it is gzipped or not. Returns `true` if the file is gzipped.
func isFileGzipped(file string) (bool, error) {
	f, err := os.Open(file)
	if err != nil {
		return false, err
	}
	defer f.Close()

	b := make([]byte, 3)
	if _, err = io.ReadFull(f, b); err != nil {
		return false, err
	}

	return b[0] == 0x1f && b[1] == 0x8b && b[2] == 8, nil
}

// validRelPath validates a tar entry name to prevent path traversal and unsafe names.
// Tar headers use forward slashes, so use `path` instead of `filepath`.
func validRelPath(p string) bool {
	// Non-empty and valid UTF-8
	if p == "" || !utf8.ValidString(p) {
		return false
	}

	// Disallow NUL and all control characters (including tab and newline)
	for _, r := range p {
		if r == 0 || r < 0x20 {
			return false
		}
	}

	// Tar uses forward slashes; reject Windows separators
	if strings.Contains(p, `\`) {
		return false
	}

	// Check for absolute paths before cleaning
	if strings.HasPrefix(p, "/") {
		return false
	}

	// Check for parent directory traversal in the original path
	// We must check before cleaning because path.Clean normalizes away "../"
	components := strings.Split(p, "/")
	for _, component := range components {
		if component == ".." {
			return false
		}
	}

	// No leading "./" (but allow trailing "/" for tar directory entries)
	if strings.HasPrefix(p, "./") {
		return false
	}

	// Validate the path without trailing slash
	pathToValidate := strings.TrimSuffix(p, "/")
	if pathToValidate == "" || pathToValidate == "." {
		return false
	}

	return true
}

// Copy ...
func Copy(src, dst string) (int64, error) {
	exists, err := FileExists(src)
	if err != nil {
		return 0, err
	}

	if !exists {
		return 0, errors.New("Source file '" + src + "' does not exist")
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer func() { _ = source.Close() }()

	destination, err := os.Create(dst)
	defer func() { _ = destination.Close() }()

	if err != nil {
		return 0, err
	}
	stats, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	os.Chmod(dst, stats.Mode())
	nBytes, err := CopyData(destination, source)
	return nBytes, err
}

// FileRead is a wrapper for os.ReadFile().
func (f Files) FileRead(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// ReadFile is a wrapper for os.ReadFile() using the same name and syntax.
func (f Files) ReadFile(path string) ([]byte, error) {
	return f.FileRead(path)
}

// FileWrite is a wrapper for os.WriteFile().
func (f Files) FileWrite(path string, content []byte, perm os.FileMode) error {
	return os.WriteFile(path, content, perm)
}

// WriteFile is a wrapper for os.ReadFile() using the same name and syntax.
func (f Files) WriteFile(path string, content []byte, perm os.FileMode) error {
	return f.FileWrite(path, content, perm)
}

// FileRemove is a wrapper for os.Remove().
func (f Files) FileRemove(path string) error {
	return os.Remove(path)
}

// FileRename is a wrapper for os.Rename().
func (f Files) FileRename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

// FileOpen is a wrapper for os.OpenFile().
func (f *Files) FileOpen(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

// MkdirAll is a wrapper for os.MkdirAll().
func (f Files) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// RemoveAll is a wrapper for os.RemoveAll().
func (f Files) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Glob is a wrapper for doublestar.Glob().
func (f Files) Glob(pattern string) (matches []string, err error) {
	return doublestar.Glob(pattern)
}

// ExcludeFiles returns a slice of files, which contains only the sub-set of files that matched none
// of the glob patterns in the provided excludes list.
func ExcludeFiles(files, excludes []string) ([]string, error) {
	if len(excludes) == 0 {
		return files, nil
	}

	filteredFiles := make([]string, 0, len(files))
	for _, file := range files {
		includeFile := true
		file = filepath.FromSlash(file)
		for _, exclude := range excludes {
			matched, err := doublestar.PathMatch(exclude, file)
			if err != nil {
				return nil, fmt.Errorf("failed to match file %s to pattern %s: %w", file, exclude, err)
			}
			if matched {
				includeFile = false
				break
			}
		}
		if includeFile {
			filteredFiles = append(filteredFiles, file)
		}
	}

	return filteredFiles, nil
}

// Getwd is a wrapper for os.Getwd().
func (f Files) Getwd() (string, error) {
	return os.Getwd()
}

// Chdir is a wrapper for os.Chdir().
func (f Files) Chdir(path string) error {
	return os.Chdir(path)
}

// Stat is a wrapper for os.Stat()
func (f Files) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// Abs is a wrapper for filepath.Abs()
func (f Files) Abs(path string) (string, error) {
	return filepath.Abs(path)
}

// Symlink is a wrapper for os.Symlink
func (f Files) Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

// SHA256 computes a SHA256 for a given file
func (f Files) SHA256(path string) (string, error) {
	// Reject files larger than the global tar limit
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.Size() > maxFileSize {
		return "", fmt.Errorf("file %s exceeds maximum allowed size of %d bytes: actual %d", path, maxFileSize, info.Size())
	}

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err = io.Copy(hash, io.LimitReader(file, maxFileSize)); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", string(hash.Sum(nil))), nil
}

// CurrentTime returns the current time in the specified format
func (f Files) CurrentTime(format string) string {
	fString := format
	if len(format) == 0 {
		fString = "20060102-150405"
	}
	return fmt.Sprint(time.Now().Format(fString))
}

// Open is a wrapper for os.Open
func (f Files) Open(name string) (io.ReadWriteCloser, error) {
	return os.Open(name)
}

// Create is a wrapper for os.Create
func (f Files) Create(name string) (io.ReadWriteCloser, error) {
	return os.Create(name)
}

// Readlink wraps os.Readlink
func (f Files) Readlink(name string) (string, error) {
	return os.Readlink(name)
}

// Readlink wraps os.Readlink
func (f Files) Lstat(path string) (os.FileInfo, error) {
	return os.Lstat(path)
}
