package codeql

import (
	"archive/zip"
	"context"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-github/v45/github"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/strings/slices"
)

const (
	notExists = "not-exists"
	exists    = "exists"
	refsHeads = "refs/heads/"
)

type gitServiceMock struct{}

func (g *gitServiceMock) GetRef(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error) {
	if ref == notExists {
		return nil, nil, &github.Error{Message: "Not Found"}
	}

	return &github.Reference{
		Ref: github.String(refsHeads + ref),
		Object: &github.GitObject{
			SHA: github.String("SHAofExistingRef"),
		},
	}, nil, nil
}

func (g *gitServiceMock) CreateRef(ctx context.Context, owner, repo string, ref *github.Reference) (*github.Reference, *github.Response, error) {
	return &github.Reference{
		Ref: ref.Ref,
		Object: &github.GitObject{
			SHA: github.String("SHAofNewRef"),
		},
	}, nil, nil
}

func (g *gitServiceMock) CreateCommit(ctx context.Context, owner, repo string, commit *github.Commit) (*github.Commit, *github.Response, error) {
	return &github.Commit{
		SHA: github.String("SHAofNewCommit"),
	}, nil, nil
}

func (g *gitServiceMock) UpdateRef(ctx context.Context, owner, repo string, ref *github.Reference, force bool) (*github.Reference, *github.Response, error) {
	return ref, nil, nil
}

func (g *gitServiceMock) CreateTree(ctx context.Context, owner, repo, baseTree string, entries []*github.TreeEntry) (*github.Tree, *github.Response, error) {
	return &github.Tree{
		SHA:     github.String("SHAofNewTree"),
		Entries: entries,
	}, nil, nil
}

func (g *gitServiceMock) GetTree(ctx context.Context, owner, repo, sha string, recursive bool) (*github.Tree, *github.Response, error) {
	if sha == "emptyRef" {
		return nil, &github.Response{
			Response: &http.Response{StatusCode: 404},
		}, &github.Error{Message: "Not Found"}
	}
	if sha == "withSubfolders" {
		return &github.Tree{
				SHA: github.String("SHAofTree"),
				Entries: []*github.TreeEntry{
					{
						SHA:  github.String("SHAofTreeEntry"),
						Path: github.String("filepath1"),
						Type: github.String(TreeType),
					},
					{
						SHA:  github.String("SHAofFileEntry"),
						Path: github.String("filepath2"),
						Mode: github.String(FileMode),
						Type: github.String(FileType),
					},
				},
				Truncated: github.Bool(false),
			}, &github.Response{
				Response: &http.Response{StatusCode: 200},
			}, nil
	}
	return &github.Tree{
			SHA: github.String("SHAofTree"),
			Entries: []*github.TreeEntry{
				{
					SHA:  github.String("SHAofFileEntry"),
					Path: github.String("filepath1"),
					Mode: github.String(FileMode),
					Type: github.String(FileType),
				},
				{
					SHA:  github.String("SHAofFileEntry"),
					Path: github.String("filepath2"),
					Mode: github.String(FileMode),
					Type: github.String(FileType),
				},
			},
			Truncated: github.Bool(false),
		}, &github.Response{
			Response: &http.Response{StatusCode: 200},
		}, nil
}

type gitRepositoriesServiceMock struct{}

func (gr *gitRepositoriesServiceMock) Get(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error) {
	if repo == notExists || owner == notExists {
		return nil, nil, &github.ErrorResponse{
			Message: "Not Found",
		}
	}
	return &github.Repository{
		DefaultBranch: github.String(exists),
	}, nil, nil
}

func (gr *gitRepositoriesServiceMock) GetCommit(ctx context.Context, owner, repo, sha string, opts *github.ListOptions) (*github.RepositoryCommit, *github.Response, error) {
	return &github.RepositoryCommit{
		SHA: &sha,
		Commit: &github.Commit{
			SHA: github.String("SHAofCommit"),
		},
	}, nil, nil
}

func (gr *gitRepositoriesServiceMock) DeleteFile(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContentResponse, *github.Response, error) {
	return &github.RepositoryContentResponse{
		Commit: github.Commit{
			SHA: github.String("SHAofDeletingFile"),
		},
	}, nil, nil
}

func TestCloneTargetRepo(t *testing.T) {
	ctx := context.Background()
	ghService := gitServiceMock{}
	ghRepoService := gitRepositoriesServiceMock{}
	t.Parallel()
	t.Run("Created new branch", func(t *testing.T) {
		ghUploader := NewGithubUploaderInstance("", "", exists, "", notExists, "", "", "", []string{})
		newRef, err := cloneTargetRepo(ctx, &ghService, &ghRepoService, &ghUploader)
		assert.NoError(t, err)
		assert.NotEmpty(t, newRef)
	})
	t.Run("Target branch exists", func(t *testing.T) {
		ghUploader := NewGithubUploaderInstance("", "", exists, "", exists, "", "", "", []string{})
		newRef, err := cloneTargetRepo(ctx, &ghService, &ghRepoService, &ghUploader)
		assert.NoError(t, err)
		assert.NotEmpty(t, newRef)
	})
	t.Run("Invalid owner/repository", func(t *testing.T) {
		ghUploader := NewGithubUploaderInstance("", notExists, notExists, "", notExists, "", "", "", []string{})
		_, err := cloneTargetRepo(ctx, &ghService, &ghRepoService, &ghUploader)
		assert.Error(t, err)
	})
}

func TestEmptyTargetBranch(t *testing.T) {
	ctx := context.Background()
	ghService := gitServiceMock{}
	ghRepoService := gitRepositoriesServiceMock{}
	t.Parallel()
	t.Run("Success with not empty ref", func(t *testing.T) {
		ghUploader := NewGithubUploaderInstance("", "", "repo", "", refsHeads+exists, "", "", "", []string{})
		lastCommitSHA, err := emptyTargetBranch(ctx, &ghService, &ghRepoService, &ghUploader, "ObjectSHA", "")
		assert.NoError(t, err)
		assert.NotEmpty(t, lastCommitSHA)
	})
	t.Run("Success with subfolders in non-empty ref", func(t *testing.T) {
		ghUploader := NewGithubUploaderInstance("", "", "repo", "", refsHeads+exists, "", "", "", []string{})
		lastCommitSHA, err := emptyTargetBranch(ctx, &ghService, &ghRepoService, &ghUploader, "withSubfolders", "")
		assert.NoError(t, err)
		assert.NotEmpty(t, lastCommitSHA)
	})
	t.Run("Success with empty ref", func(t *testing.T) {
		ghUploader := NewGithubUploaderInstance("", "", "repo", "", refsHeads+exists, "", "", "", []string{})
		lastCommitSHA, err := emptyTargetBranch(ctx, &ghService, &ghRepoService, &ghUploader, "emptyRef", "")
		assert.NoError(t, err)
		assert.NotEmpty(t, lastCommitSHA)
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
			filepath.Join(sourceDir, "subfolder1", "file1"),
			filepath.Join(sourceDir, "subfolder1", "file2"),
			filepath.Join(sourceDir, "subfolder2", "file1"),
		}
		err = createZIP(zipPath, srcFilenames)
		if err != nil {
			panic(err)
		}
		assert.NoError(t, unzip(zipPath, targetDir, sourceDir))
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
		assert.NoError(t, unzip(zipPath, targetDir, sourceDir))
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

		assert.Error(t, unzip(zipPath, targetDir, sourceDir))
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
		assert.NoError(t, unzip(zipPath, targetDir, sourceDir))
		targetFilenames := []string{
			filepath.Join(targetDir, "file1"),
			filepath.Join(targetDir, "file2"),
			filepath.Join(targetDir, "subfolder1", "file1"),
			filepath.Join(targetDir, "subfolder1", "file2"),
			filepath.Join(targetDir, "subfolder2", "file1"),
		}
		checkExistedFiles(t, targetDir, targetFilenames)
	})
}

func TestAddProjectFiles(t *testing.T) {
	ctx := context.Background()
	ghService := gitServiceMock{}
	tmpDir, err := os.MkdirTemp("", "tmp_test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	filenames := []string{
		filepath.Join(tmpDir, "file1"),
		filepath.Join(tmpDir, "file2"),
		filepath.Join(tmpDir, "subfolder1", "file1"),
		filepath.Join(tmpDir, "subfolder1", "file2"),
		filepath.Join(tmpDir, "subfolder2", "file1"),
	}
	err = fillFolderWithFiles(filenames)
	if err != nil {
		panic(err)
	}

	t.Run("Success", func(t *testing.T) {
		commitSHA := "SHAofLastCommit"
		ghUploader := NewGithubUploaderInstance("", "", "", "", refsHeads+exists, "", "", "", []string{})
		tree, err := addProjectFiles(ctx, &ghService, &ghUploader, commitSHA, tmpDir)
		assert.NoError(t, err)
		assert.NotEmpty(t, tree)
		checkExistedFiles(t, tmpDir, filenames)
	})
}

func TestPushProjectToTargetRepo(t *testing.T) {
	ctx := context.Background()
	ghService := gitServiceMock{}
	ghRepoService := gitRepositoriesServiceMock{}
	ref := &github.Reference{
		Ref: github.String(refsHeads + exists),
		Object: &github.GitObject{
			SHA: github.String("SHAofRef"),
		},
	}
	tree := &github.Tree{
		SHA: github.String("SHAofTree"),
		Entries: []*github.TreeEntry{
			{
				SHA:     github.String("SHAofTreeEntry"),
				Path:    github.String("filepath"),
				Mode:    github.String(FileMode),
				Type:    github.String(FileType),
				Content: github.String("some content"),
			},
		},
	}
	t.Run("Success", func(t *testing.T) {
		ghUploader := NewGithubUploaderInstance("", "", "", "", refsHeads+exists, "", "", "", []string{})
		commitID, err := pushProjectToTargetRepo(ctx, &ghRepoService, &ghService, &ghUploader, ref, tree)
		assert.NoError(t, err)
		assert.NotEmpty(t, commitID)
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

func fillFolderWithFiles(filenames []string) error {
	for _, fileName := range filenames {
		err := ensureBaseDir(fileName)
		if err != nil {
			return err
		}
		f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return err
		}

		r := strings.NewReader("test content\n")
		_, err = io.Copy(f, r)
		f.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func ensureBaseDir(fpath string) error {
	baseDir := path.Dir(fpath)
	info, err := os.Stat(baseDir)
	if err == nil && info.IsDir() {
		return nil
	}
	return os.MkdirAll(baseDir, 0755)
}
