package codeql

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"k8s.io/utils/strings/slices"
)

const (
	notExists = "not-exists"
	exists    = "exists"
	refsHeads = "refs/heads/"
)

type gitMock struct {
	ref string
	url string
}

func newGitMock(ref, url string) *gitMock {
	return &gitMock{ref: ref, url: url}
}

func (g *gitMock) listRemote() ([]reference, error) {
	if g.url == notExists {
		return nil, fmt.Errorf("repository not found")
	}
	list := []*referenceMock{
		{
			name: refsHeads + "ref1",
		},
		{
			name: refsHeads + "ref2",
		},
		{
			name: refsHeads + "ref3",
		},
		{
			name: refsHeads + exists,
		},
	}
	var convertedList []reference
	for _, ref := range list {
		convertedList = append(convertedList, ref)
	}
	return convertedList, nil
}

func (g *gitMock) cloneRepo(dir string, opts *git.CloneOptions) (*git.Repository, error) {
	if opts.Auth == nil {
		return nil, fmt.Errorf("error")
	}
	if opts.URL == notExists {
		return nil, fmt.Errorf("error")
	}
	return &git.Repository{}, nil
}

func (g *gitMock) switchOrphan(branch string, repo *git.Repository) error {
	return nil
}

func (g *gitMock) initRepo(dir string) (*git.Repository, error) {
	return &git.Repository{}, nil
}

type referenceMock struct {
	name string
}

func (r *referenceMock) Name() plumbing.ReferenceName {
	return plumbing.ReferenceName(r.name)
}

type repoMock struct{}

func (r *repoMock) Worktree() (*git.Worktree, error) {
	return &git.Worktree{}, nil
}

func (r *repoMock) CommitObject(commit plumbing.Hash) (*object.Commit, error) {
	return &object.Commit{Hash: commit}, nil
}

func (r *repoMock) Push(opts *git.PushOptions) error {
	if opts.Auth == nil {
		return fmt.Errorf("error")
	}
	return nil
}

type worktreeMock struct{}

func (t *worktreeMock) RemoveGlob(pattern string) error {
	return nil
}

func (t *worktreeMock) Clean(opts *git.CleanOptions) error {
	return nil
}

func (t *worktreeMock) AddWithOptions(opts *git.AddOptions) error {
	return nil
}

func (t *worktreeMock) Commit(msg string, opts *git.CommitOptions) (plumbing.Hash, error) {
	if opts.Author == nil {
		return plumbing.Hash{}, fmt.Errorf("error")
	}
	return plumbing.Hash{}, nil
}

func TestDoesRefExist(t *testing.T) {
	t.Parallel()
	t.Run("Invalid repository", func(t *testing.T) {
		ghUploader := newGitMock(refsHeads+notExists, notExists)
		_, _, err := doesRefExist(ghUploader, refsHeads+notExists)
		assert.Error(t, err)

	})
	t.Run("Ref exists", func(t *testing.T) {
		ghUploader := newGitMock(refsHeads+exists, exists)
		ok, _, err := doesRefExist(ghUploader, refsHeads+exists)
		assert.NoError(t, err)
		assert.True(t, ok)
	})
	t.Run("Ref doesn't exist", func(t *testing.T) {
		ghUploader := newGitMock(refsHeads+notExists, exists)
		ok, _, err := doesRefExist(ghUploader, refsHeads+notExists)
		assert.NoError(t, err)
		assert.False(t, ok)
	})
}

func TestClone(t *testing.T) {
	t.Parallel()
	t.Run("Created new branch", func(t *testing.T) {
		ghUploader := newGitMock(refsHeads+notExists, exists)
		repo, err := clone(ghUploader, ghUploader.url, "", ghUploader.ref, "", false, false)
		assert.NoError(t, err)
		assert.NotNil(t, repo)
	})
	t.Run("Target branch exists", func(t *testing.T) {
		ghUploader := newGitMock(refsHeads+exists, exists)
		repo, err := clone(ghUploader, ghUploader.url, "", ghUploader.ref, "", false, true)
		assert.NoError(t, err)
		assert.NotNil(t, repo)
	})
	t.Run("Repo was empty", func(t *testing.T) {
		ghUploader := newGitMock(refsHeads+exists, exists)
		repo, err := clone(ghUploader, ghUploader.url, "", ghUploader.ref, "", true, false)
		assert.NoError(t, err)
		assert.NotNil(t, repo)
	})
}

func TestClean(t *testing.T) {
	t.Parallel()
	t.Run("Success", func(t *testing.T) {
		tree := &worktreeMock{}
		err := cleanDir(tree)
		assert.NoError(t, err)
	})
}

func TestAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tree := &worktreeMock{}
		err := add(tree)
		assert.NoError(t, err)
	})
}

func TestCommit(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tree := &worktreeMock{}
		repo := &repoMock{}
		c, err := commit(repo, tree, "", "")
		assert.NoError(t, err)
		assert.NotNil(t, c)
	})
}

func TestPush(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo := &repoMock{}
		err := push(repo, "")
		assert.NoError(t, err)
	})
}

func TestUnzip(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		targetDir, err := os.MkdirTemp("", "tmp_target")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(targetDir)
		sourceDir, err := os.MkdirTemp("", "tmp_source")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(sourceDir)
		zipPath := filepath.Join(sourceDir, "src.zip")

		srcFilenames := []string{
			filepath.Join(sourceDir, "file1"),
			filepath.Join(sourceDir, "file2"),
			filepath.Join(sourceDir, "codeqlDB"),
			filepath.Join(sourceDir, "subfolder1", "file1"),
			filepath.Join(sourceDir, "subfolder1", "file2"),
			filepath.Join(sourceDir, "subfolder2", "file1"),
		}
		err = createZIP(zipPath, srcFilenames)
		if err != nil {
			panic(err)
		}
		assert.NoError(t, unzip(zipPath, targetDir, sourceDir, "codeqlDB"))
		targetFilenames := []string{
			filepath.Join(targetDir, "file1"),
			filepath.Join(targetDir, "file2"),
			filepath.Join(targetDir, "subfolder1", "file1"),
			filepath.Join(targetDir, "subfolder1", "file2"),
			filepath.Join(targetDir, "subfolder2", "file1"),
		}
		checkExistedFiles(t, targetDir, targetFilenames)
	})

	t.Run("Empty zip", func(t *testing.T) {
		targetDir, err := os.MkdirTemp("", "tmp_target")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(targetDir)
		sourceDir, err := os.MkdirTemp("", "tmp_source")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(sourceDir)
		zipPath := filepath.Join(sourceDir, "src.zip")

		filenames := []string{}
		err = createZIP(zipPath, filenames)
		if err != nil {
			panic(err)
		}
		assert.NoError(t, unzip(zipPath, targetDir, sourceDir, "codeqlDB"))
		checkExistedFiles(t, targetDir, filenames)
	})

	t.Run("zip not found", func(t *testing.T) {
		targetDir, err := os.MkdirTemp("", "tmp_target")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(targetDir)
		sourceDir, err := os.MkdirTemp("", "tmp_source")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(sourceDir)
		zipPath := filepath.Join(sourceDir, "src.zip")

		assert.Error(t, unzip(zipPath, targetDir, sourceDir, "codeqlDB"))
	})

	t.Run("extra files in zip", func(t *testing.T) {
		targetDir, err := os.MkdirTemp("", "tmp_target")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(targetDir)
		sourceDir, err := os.MkdirTemp("", "tmp_source")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(sourceDir)
		zipPath := filepath.Join(sourceDir, "src.zip")

		srcFilenames := []string{
			filepath.Join(sourceDir, "file1"),
			filepath.Join(sourceDir, "file2"),
			filepath.Join(sourceDir, "subfolder1", "file1"),
			filepath.Join(sourceDir, "subfolder1", "file2"),
			filepath.Join(sourceDir, "subfolder2", "file1"),
			filepath.Join(targetDir, "extrafile1"),
			filepath.Join(targetDir, "extrafile2"),
			filepath.Join(targetDir, "subfolder1", "extrafile1"),
		}
		err = createZIP(zipPath, srcFilenames)
		if err != nil {
			panic(err)
		}
		assert.NoError(t, unzip(zipPath, targetDir, sourceDir, "codeqlDB"))
		targetFilenames := []string{
			filepath.Join(targetDir, "file1"),
			filepath.Join(targetDir, "file2"),
			filepath.Join(targetDir, "subfolder1", "file1"),
			filepath.Join(targetDir, "subfolder1", "file2"),
			filepath.Join(targetDir, "subfolder2", "file1"),
		}
		checkExistedFiles(t, targetDir, targetFilenames)
	})

	t.Run("file under size limit extracts successfully", func(t *testing.T) {
		targetDir, err := os.MkdirTemp("", "tmp_target")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(targetDir)
		sourceDir, err := os.MkdirTemp("", "tmp_source")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(sourceDir)
		zipPath := filepath.Join(sourceDir, "src.zip")

		// Create a file that's 10MB (well under the 100MB limit)
		filename := "large_but_valid_file.txt"
		err = createZIPWithLargeFile(zipPath, filepath.Join(sourceDir, filename), 10*1024*1024) // 10MB
		if err != nil {
			panic(err)
		}

		assert.NoError(t, unzip(zipPath, targetDir, sourceDir, "codeqlDB"))

		// Verify the file was extracted
		extractedFile := filepath.Join(targetDir, filename)
		_, err = os.Stat(extractedFile)
		assert.NoError(t, err, "File under size limit should be extracted successfully")
	})

	t.Run("file exceeding size limit is rejected", func(t *testing.T) {
		targetDir, err := os.MkdirTemp("", "tmp_target")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(targetDir)
		sourceDir, err := os.MkdirTemp("", "tmp_source")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(sourceDir)
		zipPath := filepath.Join(sourceDir, "src.zip")

		// Create a file that exceeds the 100MB limit
		filename := "too_large_file.txt"
		err = createZIPWithLargeFile(zipPath, filepath.Join(sourceDir, filename), 101*1024*1024) // 101MB
		if err != nil {
			panic(err)
		}

		err = unzip(zipPath, targetDir, sourceDir, "codeqlDB")
		assert.Error(t, err, "Files exceeding size limit should be rejected")
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "exceeds maximum allowed size", "Error should mention size limit")
			assert.Contains(t, err.Error(), filename, "Error should mention the specific file")

			// Verify the file was NOT extracted
			extractedFile := filepath.Join(targetDir, filename)
			_, statErr := os.Stat(extractedFile)
			assert.Error(t, statErr, "File exceeding size limit should not be extracted")
		}
	})

	t.Run("file exactly at size limit extracts successfully", func(t *testing.T) {
		targetDir, err := os.MkdirTemp("", "tmp_target")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(targetDir)
		sourceDir, err := os.MkdirTemp("", "tmp_source")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(sourceDir)
		zipPath := filepath.Join(sourceDir, "src.zip")

		// Create a file exactly at the 100MB limit
		filename := "exactly_limit_file.txt"
		err = createZIPWithLargeFile(zipPath, filepath.Join(sourceDir, filename), MaxFileSize) // Exactly 100MB
		if err != nil {
			panic(err)
		}

		assert.NoError(t, unzip(zipPath, targetDir, sourceDir, "codeqlDB"))

		// Verify the file was extracted
		extractedFile := filepath.Join(targetDir, filename)
		stat, err := os.Stat(extractedFile)
		assert.NoError(t, err, "File at exactly the size limit should be extracted successfully")
		assert.Equal(t, int64(MaxFileSize), stat.Size(), "Extracted file should have correct size")
	})
}

func TestGetSourceLocationPrefix(t *testing.T) {
	t.Parallel()
	t.Run("Success", func(t *testing.T) {
		filename := "test-file.yml"
		location := "/some/location"
		err := createFile(filename, location, false)
		assert.NoError(t, err)
		defer os.Remove(filename)
		srcLocationPrefix, err := getSourceLocationPrefix(filename)
		assert.NoError(t, err)
		assert.Equal(t, location, srcLocationPrefix)
	})

	t.Run("No file found", func(t *testing.T) {
		filename := "test-file-2.yml"
		_, err := getSourceLocationPrefix(filename)
		assert.Error(t, err)
	})

	t.Run("Empty file", func(t *testing.T) {
		filename := "test-file-3.yml"
		err := createFile(filename, "", true)
		assert.NoError(t, err)
		defer os.Remove(filename)
		srcLocationPrefix, err := getSourceLocationPrefix(filename)
		assert.NoError(t, err)
		assert.Empty(t, srcLocationPrefix)
	})
}

func checkExistedFiles(t *testing.T, dir string, filenames []string) {
	counter := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == dir || info.IsDir() {
			return nil
		}
		assert.True(t, slices.Contains(filenames, path))
		counter++
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, len(filenames), counter)
}

func createZIP(zipPath string, filenames []string) error {
	archive, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)
	defer zipWriter.Close()

	for _, filename := range filenames {
		writer, err := zipWriter.Create(filename)
		if err != nil {
			return err
		}

		reader := strings.NewReader("test content\n")
		if _, err := io.Copy(writer, reader); err != nil {
			return err
		}
	}
	return nil
}

func createZIPWithLargeFile(zipPath string, filename string, sizeInBytes int64) error {
	archive, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)
	defer zipWriter.Close()

	writer, err := zipWriter.Create(filename)
	if err != nil {
		return err
	}

	// Create a large content by repeating a pattern
	pattern := "0123456789abcdefghijklmnopqrstuvwxyz\n" // 37 bytes
	bytesWritten := int64(0)
	for bytesWritten < sizeInBytes {
		remainingBytes := sizeInBytes - bytesWritten
		contentToWrite := pattern
		if remainingBytes < int64(len(pattern)) {
			contentToWrite = pattern[:remainingBytes]
		}
		n, err := writer.Write([]byte(contentToWrite))
		if err != nil {
			return err
		}
		bytesWritten += int64(n)
	}

	return nil
}

func createFile(fileName, location string, isEmpty bool) error {
	err := ensureBaseDir(fileName)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	if isEmpty {
		return nil
	}

	type codeqlDatabase struct {
		SourceLocation string `yaml:"sourceLocationPrefix"`
		OtherInfo      string `yaml:"otherInfo"`
	}
	db := codeqlDatabase{SourceLocation: location, OtherInfo: "test"}
	data, err := yaml.Marshal(db)
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	return err
}

func ensureBaseDir(fpath string) error {
	baseDir := path.Dir(fpath)
	info, err := os.Stat(baseDir)
	if err == nil && info.IsDir() {
		return nil
	}
	return os.MkdirAll(baseDir, 0755)
}
