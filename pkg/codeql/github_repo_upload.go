package codeql

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"gopkg.in/yaml.v3"
)

type GitUploader interface {
	UploadProjectToGithub() (string, error)
}

type GitUploaderInstance struct {
	*command.Command

	token          string
	ref            string
	sourceCommitId string
	sourceRepo     string
	targetRepo     string
	dbDir          string
}

func NewGitUploaderInstance(token, ref, dbDir, sourceCommitId, sourceRepo, targetRepo string) (*GitUploaderInstance, error) {
	dbAbsPath, err := filepath.Abs(dbDir)
	if err != nil {
		return nil, err
	}
	instance := &GitUploaderInstance{
		Command:        &command.Command{},
		token:          token,
		ref:            ref,
		sourceCommitId: sourceCommitId,
		sourceRepo:     sourceRepo,
		targetRepo:     targetRepo,
		dbDir:          filepath.Clean(dbAbsPath),
	}

	instance.Stdout(log.Writer())
	instance.Stderr(log.Writer())
	return instance, nil
}

type gitUtils interface {
	listRemote() ([]reference, error)
	cloneRepo(dir string, opts *git.CloneOptions) (*git.Repository, error)
	switchOrphan(ref string, repo *git.Repository) error
	initRepo(dir string) (*git.Repository, error)
}

type repository interface {
	Worktree() (*git.Worktree, error)
	CommitObject(commit plumbing.Hash) (*object.Commit, error)
	Push(o *git.PushOptions) error
}

type worktree interface {
	RemoveGlob(pattern string) error
	Clean(opts *git.CleanOptions) error
	AddWithOptions(opts *git.AddOptions) error
	Commit(msg string, opts *git.CommitOptions) (plumbing.Hash, error)
}

type reference interface {
	Name() plumbing.ReferenceName
}

const (
	CommitMessageMirroringCode = "Mirroring code for revision %s from %s"
	SrcZip                     = "src.zip"
	CodeqlDatabaseYml          = "codeql-database.yml"
	OriginRemote               = "origin"
)

func (uploader *GitUploaderInstance) UploadProjectToGithub() (string, error) {
	tmpDir, err := os.MkdirTemp("", "tmp")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	refExists, repoEmpty, err := doesRefExist(uploader, uploader.ref)
	if err != nil {
		return "", err
	}

	repo, err := clone(uploader, uploader.targetRepo, uploader.token, uploader.ref, tmpDir, repoEmpty, refExists)
	if err != nil {
		return "", err
	}

	tree, err := repo.Worktree()
	if err != nil {
		return "", err
	}
	err = cleanDir(tree)
	if err != nil {
		return "", err
	}

	srcLocationPrefix, err := getSourceLocationPrefix(filepath.Join(uploader.dbDir, CodeqlDatabaseYml))
	if err != nil {
		return "", err
	}

	zipPath := path.Join(uploader.dbDir, SrcZip)
	err = unzip(zipPath, tmpDir, strings.Trim(srcLocationPrefix, fmt.Sprintf("%c", os.PathSeparator)), strings.Trim(uploader.dbDir, fmt.Sprintf("%c", os.PathSeparator)))
	if err != nil {
		return "", err
	}

	err = add(tree)
	if err != nil {
		return "", err
	}

	newCommit, err := commit(repo, tree, uploader.sourceCommitId, uploader.sourceRepo)
	if err != nil {
		return "", err
	}

	err = push(repo, uploader.token)
	if err != nil {
		return "", err
	}

	return newCommit.ID().String(), err
}

func (uploader *GitUploaderInstance) listRemote() ([]reference, error) {
	rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: OriginRemote,
		URLs: []string{uploader.targetRepo},
	})

	list, err := rem.List(&git.ListOptions{
		Auth: &http.BasicAuth{
			Username: "does-not-matter",
			Password: uploader.token,
		},
	})
	if err != nil {
		return nil, err
	}
	var convertedList []reference
	for _, ref := range list {
		convertedList = append(convertedList, ref)
	}
	return convertedList, err
}

func (uploader *GitUploaderInstance) initRepo(dir string) (*git.Repository, error) {
	// git init -b <ref>
	repo, err := git.PlainInitWithOptions(dir, &git.PlainInitOptions{
		InitOptions: git.InitOptions{
			DefaultBranch: plumbing.ReferenceName(uploader.ref),
		},
	})
	if err != nil {
		return nil, err
	}

	// git remote add origin <repo>
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: OriginRemote,
		URLs: []string{uploader.targetRepo},
	})
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (uploader *GitUploaderInstance) cloneRepo(dir string, opts *git.CloneOptions) (*git.Repository, error) {
	return git.PlainClone(dir, false, opts)
}

func (uploader *GitUploaderInstance) switchOrphan(ref string, r *git.Repository) error {
	branchName := strings.Split(ref, "/")[2:]
	newRef := plumbing.NewBranchReferenceName(strings.Join(branchName, "/"))
	return r.Storer.SetReference(plumbing.NewSymbolicReference(plumbing.HEAD, newRef))
}

func doesRefExist(uploader gitUtils, ref string) (bool, bool, error) {
	// git ls-remote <repo>
	remoteRefs, err := uploader.listRemote()
	if err != nil {
		if errors.Is(err, transport.ErrEmptyRemoteRepository) {
			return false, true, nil
		}
		return false, false, err
	}
	for _, r := range remoteRefs {
		if string(r.Name()) == ref {
			return true, false, nil
		}
	}
	return false, false, nil
}

func clone(uploader gitUtils, url, token, ref, dir string, repoEmpty, refExists bool) (*git.Repository, error) {
	if repoEmpty {
		return uploader.initRepo(dir)
	}

	opts := &git.CloneOptions{
		URL: url,
		Auth: &http.BasicAuth{
			Username: "does-not-matter",
			Password: token,
		},
		SingleBranch: true,
		Depth:        1,
	}
	if refExists {
		opts.ReferenceName = plumbing.ReferenceName(ref)
		// git clone -b <ref> --single-branch --depth=1 <url> <dir>
		return uploader.cloneRepo(dir, opts)
	}

	// git clone --single-branch --depth=1 <url> <dir>
	r, err := uploader.cloneRepo(dir, opts)
	if err != nil {
		return nil, err
	}

	// git switch --orphan <ref>
	err = uploader.switchOrphan(ref, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func cleanDir(t worktree) error {
	// git rm -r
	err := t.RemoveGlob("*")
	if err != nil {
		return err
	}
	// git clean -d
	err = t.Clean(&git.CleanOptions{Dir: true})
	return err
}

func add(t worktree) error {
	// git add --all
	return t.AddWithOptions(&git.AddOptions{
		All: true,
	})
}

func commit(r repository, t worktree, sourceCommitId, sourceRepo string) (*object.Commit, error) {
	// git commit --allow-empty -m <msg>
	newCommit, err := t.Commit(fmt.Sprintf(CommitMessageMirroringCode, sourceCommitId, sourceRepo), &git.CommitOptions{
		AllowEmptyCommits: true,
		Author: &object.Signature{
			When: time.Now(),
		},
	})
	if err != nil {
		return nil, err
	}
	return r.CommitObject(newCommit)
}

func push(r repository, token string) error {
	// git push
	return r.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: "does-not-matter",
			Password: token,
		},
	})
}

func unzip(zipPath, targetDir, srcDir, dbDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// normalize zip entry and input dirs to use forward slashes for comparison
		fName := filepath.ToSlash(f.Name)
		srcDirNorm := filepath.ToSlash(srcDir)
		dbDirNorm := filepath.ToSlash(dbDir)

		if !strings.Contains(fName, srcDirNorm) || strings.Contains(fName, dbDirNorm) {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		// remove the srcDir prefix (in slash form) and convert back to OS-specific paths
		fName = strings.TrimPrefix(fName, srcDirNorm)
		fName = filepath.FromSlash(fName)
		fpath := filepath.Join(targetDir, fName)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
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

func getSourceLocationPrefix(fileName string) (string, error) {
	type codeqlDatabase struct {
		SourceLocation string `yaml:"sourceLocationPrefix"`
	}
	var db codeqlDatabase
	file, err := os.ReadFile(fileName)
	if err != nil {
		return "", err
	}
	err = yaml.Unmarshal(file, &db)
	if err != nil {
		return "", err
	}

	return db.SourceLocation, nil
}
