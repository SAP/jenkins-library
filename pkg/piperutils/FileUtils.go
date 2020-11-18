package piperutils

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar"
)

// FileUtils ...
type FileUtils interface {
	Abs(path string) (string, error)
	FileExists(filename string) (bool, error)
	Copy(src, dest string) (int64, error)
	FileRead(path string) ([]byte, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Chmod(path string, mode os.FileMode) error
	Glob(pattern string) (matches []string, err error)
}

// Files ...
type Files struct {
}

// TempDir creates a temporary directory
func (f Files) TempDir(dir, pattern string) (name string, err error) {
	return ioutil.TempDir(dir, pattern)
}

// FileExists returns true if the file system entry for the given path exists and is not a directory.
func (f Files) FileExists(filename string) (bool, error) {
	info, err := os.Stat(filename)

	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return !info.IsDir(), nil
}

// FileExists returns true if the file system entry for the given path exists and is not a directory.
func FileExists(filename string) (bool, error) {
	return Files{}.FileExists(filename)
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

	exists, err := f.FileExists(src)

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
	if err != nil {
		return 0, err
	}
	defer func() { _ = destination.Close() }()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

//Chmod is a wrapper for os.Chmod().
func (f Files) Chmod(path string, mode os.FileMode) error {
	return os.Chmod(path, mode)
}

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
// from https://golangcode.com/unzip-files-in-go/ with the following license:
// MIT License
//
// Copyright (c) 2017 Edd Turtle
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

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		_ = outFile.Close()
		_ = rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

// Copy ...
func Copy(src, dst string) (int64, error) {
	return Files{}.Copy(src, dst)
}

// FileRead is a wrapper for ioutil.ReadFile().
func (f Files) FileRead(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

// FileWrite is a wrapper for ioutil.WriteFile().
func (f Files) FileWrite(path string, content []byte, perm os.FileMode) error {
	return ioutil.WriteFile(path, content, perm)
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
