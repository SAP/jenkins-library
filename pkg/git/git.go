package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/pkg/errors"
	"time"
)

// utilsWorkTree interface abstraction of git.Worktree to enable tests
type utilsWorkTree interface {
	Add(path string) (plumbing.Hash, error)
	Commit(msg string, opts *git.CommitOptions) (plumbing.Hash, error)
	Checkout(opts *git.CheckoutOptions) error
}

// utilsRepository interface abstraction of git.Repository to enable tests
type utilsRepository interface {
	Worktree() (*git.Worktree, error)
	Push(o *git.PushOptions) error
}

// utilsGit interface abstraction of git to enable tests
type utilsGit interface {
	plainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error)
}

// CommitSingleFile Commits the file located in the relative file path with the commitMessage to the given worktree.
// In case of errors, the error is returned. In the successful case the commit is provided.
func CommitSingleFile(filePath, commitMessage, author string, worktree *git.Worktree) (plumbing.Hash, error) {
	return commitSingleFile(filePath, commitMessage, author, worktree)
}

func commitSingleFile(filePath, commitMessage, author string, worktree utilsWorkTree) (plumbing.Hash, error) {
	_, err := worktree.Add(filePath)
	if err != nil {
		return [20]byte{}, errors.Wrap(err, "failed to add file to git")
	}

	commit, err := worktree.Commit(commitMessage, &git.CommitOptions{
		All:    true,
		Author: &object.Signature{Name: author, When: time.Now()},
	})
	if err != nil {
		return [20]byte{}, errors.Wrap(err, "failed to commit file")
	}

	return commit, nil
}

// PushChangesToRepository Pushes all committed changes in the repository to the remote repository
func PushChangesToRepository(username, password string, repository *git.Repository) error {
	return pushChangesToRepository(username, password, repository)
}

func pushChangesToRepository(username, password string, repository utilsRepository) error {
	pushOptions := &git.PushOptions{
		Auth: &http.BasicAuth{Username: username, Password: password},
	}
	err := repository.Push(pushOptions)
	if err != nil {
		return errors.Wrap(err, "failed to push commit")
	}
	return nil
}

// PlainClone Clones a non-bare repository to the provided directory
func PlainClone(username, password, serverURL, directory string) (*git.Repository, error) {
	abstractedGit := &abstractionGit{}
	return plainClone(username, password, serverURL, directory, abstractedGit)
}

func plainClone(username, password, serverURL, directory string, abstractionGit utilsGit) (*git.Repository, error) {
	gitCloneOptions := git.CloneOptions{
		Auth: &http.BasicAuth{Username: username, Password: password},
		URL:  serverURL,
	}
	repository, err := abstractionGit.plainClone(directory, false, &gitCloneOptions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to clone git")
	}
	return repository, nil
}

// ChangeBranch checkout the provided branch.
// It will create a new branch if the branch does not exist yet.
// It will checkout "master" if no branch name if provided
func ChangeBranch(branchName string, worktree *git.Worktree) error {
	return changeBranch(branchName, worktree)
}

func changeBranch(branchName string, worktree utilsWorkTree) error {
	if branchName == "" {
		branchName = "master"
	}

	var checkoutOptions = &git.CheckoutOptions{}
	checkoutOptions.Branch = plumbing.NewBranchReferenceName(branchName)
	checkoutOptions.Create = false
	err := worktree.Checkout(checkoutOptions)
	if err != nil {
		// branch might not exist, try to create branch
		checkoutOptions.Create = true
		err = worktree.Checkout(checkoutOptions)
		if err != nil {
			return errors.Wrap(err, "failed to checkout branch")
		}
	}

	return nil
}

type abstractionGit struct{}

func (abstractionGit) plainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {
	return git.PlainClone(path, isBare, o)
}
