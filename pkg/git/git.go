package git

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
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
func CommitSingleFile(filePath, commitMessage string, worktree *git.Worktree) (plumbing.Hash, error) {
	return commitSingleFile(filePath, commitMessage, worktree)
}

func commitSingleFile(filePath, commitMessage string, worktree utilsWorkTree) (plumbing.Hash, error) {
	_, gitAddError := worktree.Add(filePath)
	if gitAddError != nil {
		log.Entry().WithError(gitAddError).Error("Failed to add file to git")
		return [20]byte{}, gitAddError
	}

	commit, commitError := worktree.Commit(commitMessage, &git.CommitOptions{})
	if commitError != nil {
		log.Entry().WithError(commitError).Error("Failed to commit file")
		return [20]byte{}, commitError
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
	pushError := repository.Push(pushOptions)
	if pushError != nil {
		log.Entry().WithError(pushError).Error("Failed to push commit")
		return pushError
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
	repository, gitCloneError := abstractionGit.plainClone(directory, false, &gitCloneOptions)
	if gitCloneError != nil {
		log.Entry().WithError(gitCloneError).Error("Failed to clone git")
		return nil, gitCloneError
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
	checkoutError := worktree.Checkout(checkoutOptions)
	if checkoutError != nil {
		// branch might not exist, try to create branch
		checkoutOptions.Create = true
		checkoutError = worktree.Checkout(checkoutOptions)
		if checkoutError != nil {
			log.Entry().WithError(checkoutError).Error("Failed to checkout branch")
			return checkoutError
		}
	}

	return nil
}

type abstractionGit struct{}

func (abstractionGit) plainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {
	return git.PlainClone(path, isBare, o)
}
