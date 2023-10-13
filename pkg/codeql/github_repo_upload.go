package codeql

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"gopkg.in/yaml.v2"
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
	codeqlDatabaseYml          = "codeql-database.yml"
)

func (uploader *GitUploaderInstance) UploadProjectToGithub() (string, error) {
	tmpDir, err := os.MkdirTemp("", "tmp")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	refExists, err := doesRefExist(uploader, uploader.ref)
	if err != nil {
		return "", err
	}

	repo, err := clone(uploader, uploader.targetRepo, uploader.token, uploader.ref, tmpDir, refExists)
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

	srcLocationPrefix, err := getSourceLocationPrefix(filepath.Join(uploader.dbDir, codeqlDatabaseYml))
	if err != nil {
		return "", err
	}

	zipPath := path.Join(uploader.dbDir, SrcZip)
	err = unzip(zipPath, tmpDir, strings.Trim(srcLocationPrefix, fmt.Sprintf("%c", os.PathSeparator)))
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
		Name: "origin",
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

func (uploader *GitUploaderInstance) cloneRepo(dir string, opts *git.CloneOptions) (*git.Repository, error) {
	return git.PlainClone(dir, false, opts)
}

func (uploader *GitUploaderInstance) switchOrphan(ref string, r *git.Repository) error {
	branchName := strings.Split(ref, "/")[2:]
	newRef := plumbing.NewBranchReferenceName(strings.Join(branchName, "/"))
	return r.Storer.SetReference(plumbing.NewSymbolicReference(plumbing.HEAD, newRef))
}

func doesRefExist(uploader gitUtils, ref string) (bool, error) {
	// git ls-remote <repo>
	remoteRefs, err := uploader.listRemote()
	if err != nil {
		return false, err
	}
	for _, r := range remoteRefs {
		if string(r.Name()) == ref {
			return true, nil
		}
	}
	return false, nil
}

func clone(uploader gitUtils, url, token, ref, dir string, refExists bool) (*git.Repository, error) {
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

func unzip(zipPath, targetDir, srcDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fName := f.Name

		if runtime.GOOS == "windows" {
			fNameSplit := strings.Split(fName, "/")
			if len(fNameSplit) == 0 {
				continue
			}
			fNameSplit[0] = strings.Replace(fNameSplit[0], "_", ":", 1)
			fName = strings.Join(fNameSplit, fmt.Sprintf("%c", os.PathSeparator))
		}
		if !strings.Contains(fName, srcDir) {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		fName = strings.TrimPrefix(fName, srcDir)
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
