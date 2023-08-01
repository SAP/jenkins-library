package codeql

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	sapgithub "github.com/SAP/jenkins-library/pkg/github"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/google/go-github/v45/github"
)

type GithubUploader interface {
	UploadProjectToGithub() error
}

type gitService interface {
	GetRef(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error)
	CreateRef(ctx context.Context, owner, repo string, ref *github.Reference) (*github.Reference, *github.Response, error)
	CreateCommit(ctx context.Context, owner, repo string, commit *github.Commit) (*github.Commit, *github.Response, error)
	UpdateRef(ctx context.Context, owner, repo string, ref *github.Reference, force bool) (*github.Reference, *github.Response, error)
	CreateTree(ctx context.Context, owner, repo, baseTree string, entries []*github.TreeEntry) (*github.Tree, *github.Response, error)
	GetTree(ctx context.Context, owner, repo, sha string, recursive bool) (*github.Tree, *github.Response, error)
}

type gitRepositoriesService interface {
	GetCommit(ctx context.Context, owner, repo, sha string, opts *github.ListOptions) (*github.RepositoryCommit, *github.Response, error)
	ListCommits(ctx context.Context, owner, repo string, opts *github.CommitsListOptions) ([]*github.RepositoryCommit, *github.Response, error)
	Get(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error)
	DeleteFile(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContentResponse, *github.Response, error)
}

type GithubUploaderInstance struct {
	*command.Command

	serverUrl      string
	owner          string
	repository     string
	token          string
	ref            string
	sourceCommitId string
	sourceRepo     string
	dbDir          string
	trustedCerts   []string
}

func NewGithubUploaderInstance(serverUrl, owner, repository, token, ref, dbDir,
	sourceCommitId, sourceRepo string,
	trustedCerts []string) GithubUploaderInstance {
	instance := GithubUploaderInstance{
		Command:        &command.Command{},
		serverUrl:      serverUrl,
		owner:          owner,
		repository:     repository,
		token:          token,
		ref:            ref,
		sourceCommitId: sourceCommitId,
		sourceRepo:     sourceRepo,
		dbDir:          dbDir,
		trustedCerts:   trustedCerts,
	}

	instance.Stdout(log.Writer())
	instance.Stderr(log.Writer())
	return instance
}

const (
	CommitMessageRmFiles       = "branch emptying"
	CommitMessageMirroringCode = "Mirroring code for revision %s from %s"
	FileType                   = "blob"
	FileMode                   = "100644"
	TreeType                   = "tree"
	SrcZip                     = "src.zip"
)

func (repoUploader *GithubUploaderInstance) UploadProjectToGithub() (string, error) {
	apiUrl := getApiUrl(repoUploader.serverUrl)
	ctx, client, err := sapgithub.NewClient(repoUploader.token, apiUrl, "", repoUploader.trustedCerts)
	if err != nil {
		return "", err
	}

	tmpDir, err := os.MkdirTemp("", "tmp")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	sourceDir := filepath.Dir(ex)

	newRef, err := checkoutTargetRepo(ctx, client.Git, client.Repositories, repoUploader)
	if err != nil {
		return "", err
	}
	lastCommitSHA, err := emptyTargetBranch(ctx, client.Git, client.Repositories, repoUploader, *newRef.Object.SHA, "")
	if err != nil {
		return "", err
	}
	zipPath := path.Join(sourceDir, repoUploader.dbDir, SrcZip)
	err = unzip(zipPath, tmpDir, strings.Trim(sourceDir, fmt.Sprintf("%c", os.PathSeparator)), repoUploader.dbDir)
	if err != nil {
		return "", err
	}
	tree, err := addProjectFiles(ctx, client.Git, repoUploader, lastCommitSHA, tmpDir)
	if err != nil {
		return "", err
	}
	newCommitId, err := pushProjectToTargetRepo(ctx, client.Repositories, client.Git, repoUploader, newRef, tree)

	return newCommitId, err
}

// checks if target branch exists, creates a new one if necessary from the first commit of default branch
func checkoutTargetRepo(ctx context.Context, gitService gitService, repoService gitRepositoriesService, uploader *GithubUploaderInstance) (*github.Reference, error) {
	repo, _, err := repoService.Get(ctx, uploader.owner, uploader.repository)
	if err != nil {
		return nil, err
	}
	ref, _, err := gitService.GetRef(ctx, uploader.owner, uploader.repository, uploader.ref)
	if err == nil {
		return ref, nil
	}
	baseRefName := *repo.DefaultBranch
	firstCommit, err := getFirstCommit(ctx, repoService, uploader, baseRefName)
	if err != nil {
		return nil, err
	}
	newRef := &github.Reference{Ref: &uploader.ref, Object: &github.GitObject{SHA: firstCommit.SHA}}
	ref, _, err = gitService.CreateRef(ctx, uploader.owner, uploader.repository, newRef)
	return ref, err
}

// get the first commit of branch as we don't need the whole history
func getFirstCommit(ctx context.Context, repoService gitRepositoriesService, uploader *GithubUploaderInstance, ref string) (*github.RepositoryCommit, error) {
	page := 1
	firstCommit := &github.RepositoryCommit{}
	for page != 0 {
		opts := &github.CommitsListOptions{
			SHA: ref,
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: perPageCount,
			},
		}
		commits, response, err := repoService.ListCommits(ctx, uploader.owner, uploader.repository, opts)
		if err != nil {
			return nil, err
		}
		page = response.NextPage
		firstCommit = commits[len(commits)-1]
	}
	return firstCommit, nil
}

// delete all files from target branch (remote)
func emptyTargetBranch(ctx context.Context, gitService gitService, repoService gitRepositoriesService, uploader *GithubUploaderInstance, objectSHA, entryParentPath string) (string, error) {
	lastCommitSHA := objectSHA
	tree, resp, err := gitService.GetTree(ctx, uploader.owner, uploader.repository, objectSHA, false)
	if resp.Response.StatusCode == 404 {
		return lastCommitSHA, nil
	}
	if err != nil {
		return lastCommitSHA, err
	}

	for _, entry := range tree.Entries {
		entryPath := *entry.Path
		if entryParentPath != "" {
			entryPath = fmt.Sprintf("%s/%s", entryParentPath, *entry.Path)
		}
		if *entry.Type == TreeType {
			lastCommitSHA, err = emptyTargetBranch(ctx, gitService, repoService, uploader, *entry.SHA, entryPath)
			if err != nil {
				return lastCommitSHA, err
			}
			continue
		}
		content, _, err := repoService.DeleteFile(ctx, uploader.owner, uploader.repository, entryPath, &github.RepositoryContentFileOptions{
			Message: github.String(CommitMessageRmFiles),
			SHA:     entry.SHA,
			Branch:  &uploader.ref,
		})
		if err != nil {
			return lastCommitSHA, err
		}
		lastCommitSHA = *content.Commit.SHA
	}
	return lastCommitSHA, nil
}

func addProjectFiles(ctx context.Context, gitService gitService, uploader *GithubUploaderInstance, lastCommitSHA, dir string) (*github.Tree, error) {
	var entries []*github.TreeEntry
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fileName := strings.Trim(strings.TrimPrefix(path, dir), fmt.Sprintf("%c", os.PathSeparator))
		entries = append(entries, &github.TreeEntry{
			Path:    &fileName,
			Type:    github.String(FileType),
			Content: github.String(string(content)),
			Mode:    github.String(FileMode),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	newTree, _, err := gitService.CreateTree(ctx, uploader.owner, uploader.repository, lastCommitSHA, entries)
	return newTree, err
}

func pushProjectToTargetRepo(ctx context.Context, repoService gitRepositoriesService, gitService gitService, uploader *GithubUploaderInstance, ref *github.Reference, tree *github.Tree) (string, error) {
	parent, _, err := repoService.GetCommit(ctx, uploader.owner, uploader.repository, *ref.Object.SHA, nil)
	if err != nil {
		return "", err
	}
	parent.Commit.SHA = parent.SHA

	commit := &github.Commit{
		Message: github.String(fmt.Sprintf(CommitMessageMirroringCode, uploader.sourceCommitId, uploader.sourceRepo)),
		Tree:    tree,
		Parents: []*github.Commit{parent.Commit},
	}
	newCommit, _, err := gitService.CreateCommit(ctx, uploader.owner, uploader.repository, commit)
	if err != nil {
		return "", err
	}
	ref.Object.SHA = newCommit.SHA
	_, _, err = gitService.UpdateRef(ctx, uploader.owner, uploader.repository, ref, true)
	return *newCommit.SHA, err
}

func unzip(zipPath, targetDir, srcDir, dbDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if !strings.Contains(f.Name, srcDir) || strings.Contains(f.Name, path.Join(srcDir, dbDir)) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}

		fName := strings.TrimPrefix(f.Name, srcDir)
		fpath := filepath.Join(targetDir, fName)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModeDir)
			rc.Close()
			continue
		}
		err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm)
		if err != nil {
			rc.Close()
			return err
		}

		fNew, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(fNew, rc)
		if err != nil {
			rc.Close()
			fNew.Close()
			return err
		}
		rc.Close()
		fNew.Close()
	}
	return nil
}
