package piperutils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
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
	DirExists(path string) (bool, error)
	FileExists(filename string) (bool, error)
	Copy(src, dest string) (int64, error)
	FileRead(path string) ([]byte, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
	FileRemove(path string) error
	MkdirAll(path string, perm os.FileMode) error
	Chmod(path string, mode os.FileMode) error
	Glob(pattern string) (matches []string, err error)
	Chdir(path string) error
	TempDir(string, string) (string, error)
	RemoveAll(string) error
	FileRename(string, string) error
	Getwd() (string, error)
	Symlink(oldname string, newname string) error
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
	nBytes, err := CopyData(destination, source)
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
		fmt.Errorf("unable to open src: %v", err)
	}
	return untar(file, dest, stripComponentLevel)
}

func untar(r io.Reader, dir string, level int) (err error) {
	madeDir := map[string]bool{}

	zr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("requires gzip-compressed body: %v", err)
	}
	tr := tar.NewReader(zr)
	for {
		f, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar error: %v", err)
		}
		if !validRelPath(f.Name) {
			return fmt.Errorf("tar contained invalid name error %q", f.Name)
		}
		rel := filepath.FromSlash(f.Name)

		// when level X folder(s) needs to be removed we first check that the rel path must have atleast X or greater than X pathseperatorserr
		// or else we might end in index out of range
		if level > 0 {
			if strings.Count(rel, string(os.PathSeparator)) >= level {
				relSplit := strings.SplitN(rel, string(os.PathSeparator), level+1)
				rel = relSplit[level]
			} else {
				return fmt.Errorf("files %q in tarball archive not under level %v", f.Name, level)
			}
		}

		abs := filepath.Join(dir, rel)

		fi := f.FileInfo()
		mode := fi.Mode()
		switch {
		case mode.IsRegular():
			// Make the directory. This is redundant because it should
			// already be made by a directory entry in the tar
			// beforehand. Thus, don't check for errors; the next
			// write will fail with the same error.
			dir := filepath.Dir(abs)
			if !madeDir[dir] {
				if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
					return err
				}
				madeDir[dir] = true
			}
			wf, err := os.OpenFile(abs, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode.Perm())
			if err != nil {
				return err
			}
			n, err := io.Copy(wf, tr)
			if closeErr := wf.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
			if err != nil {
				return fmt.Errorf("error writing to %s: %v", abs, err)
			}
			if n != f.Size {
				return fmt.Errorf("only wrote %d bytes to %s; expected %d", n, abs, f.Size)
			}
		case mode.IsDir():
			if err := os.MkdirAll(abs, 0755); err != nil {
				return err
			}
			madeDir[abs] = true
		default:
			return fmt.Errorf("tar file entry %s contained unsupported file type %v", f.Name, mode)
		}
	}
	return nil
}

func validRelPath(p string) bool {
	if p == "" || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") || strings.Contains(p, "../") {
		return false
	}
	return true
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

// ExcludeFiles returns a slice of files, which contains only the sub-set of files that matched none
// of the glob patterns in the provided excludes list.
func ExcludeFiles(files, excludes []string) ([]string, error) {
	if len(excludes) == 0 {
		return files, nil
	}

	var filteredFiles []string
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
