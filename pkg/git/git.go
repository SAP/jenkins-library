package git

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// UtilsWorkTree interface abstraction of git.Worktree to enable tests
type UtilsWorkTree interface {
	Add(path string) (plumbing.Hash, error)
	Commit(msg string, opts *git.CommitOptions) (plumbing.Hash, error)
	Checkout(opts *git.CheckoutOptions) error
}

// UtilsRepository interface abstraction of git.Repository to enable tests
type UtilsRepository interface {
	Worktree() (*git.Worktree, error)
	Push(o *git.PushOptions) error
}

// utilsGit interface abstraction of git to enable tests
type utilsGit interface {
	plainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error)
}

// abstractedGit abstraction of git to enable tests
var abstractedGit utilsGit = abstractionGit{}

type TheGitUtils struct {
}

// CommitSingleFile Commits the file located in the relative file path with the commitMessage to the given worktree.
// In case of errors, the error is returned. In the successful case the commit is provided.
func (TheGitUtils) CommitSingleFile(filePath, commitMessage string, worktree UtilsWorkTree) (plumbing.Hash, error) {
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
func (TheGitUtils) PushChangesToRepository(username, password string, repository UtilsRepository) error {
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
func (TheGitUtils) PlainClone(username, password, serverUrl, directory string) (UtilsRepository, error) {
	gitCloneOptions := git.CloneOptions{
		Auth: &http.BasicAuth{Username: username, Password: password},
		URL:  serverUrl,
	}
	repository, gitCloneError := abstractedGit.plainClone(directory, false, &gitCloneOptions)
	if gitCloneError != nil {
		log.Entry().WithError(gitCloneError).Error("Failed to clone git")
		return nil, gitCloneError
	}
	return repository, nil
}

// ChangeBranch checkout the provided branch.
// It will create a new branch if the branch does not exist yet.
// It will checkout "master" if no branch name if provided
func (TheGitUtils) ChangeBranch(branchName string, worktree UtilsWorkTree) error {
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

func (TheGitUtils) GetWorktree(repository UtilsRepository) (UtilsWorkTree, error) {
	worktree, err := repository.Worktree()
	if err != nil {
		log.Entry().WithError(err).Error("could not receive worktree from repository")
	}
	return worktree, err
}

type abstractionGit struct{}

func (abstractionGit) plainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {
	return git.PlainClone(path, isBare, o)
}
