package git

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type TheGitUtils struct {
}

// CommitSingleFile Commits the file located in the relative file path with the commitMessage to the give repository.
// In case of errors, the error is returned. In the successful case the commit is provided.
func (f TheGitUtils) CommitSingleFile(filePath, commitMessage string, repository *git.Repository) (plumbing.Hash, error) {
	workTree, workTreeError := repository.Worktree()
	if workTreeError != nil {
		log.Entry().WithError(workTreeError).Error("Failed to get git work tree")
		return [20]byte{}, workTreeError
	}

	_, gitAddError := workTree.Add(filePath)
	if gitAddError != nil {
		log.Entry().WithError(gitAddError).Error("Failed to add file to git")
		return [20]byte{}, gitAddError
	}

	commit, commitError := workTree.Commit(commitMessage, &git.CommitOptions{})
	if commitError != nil {
		log.Entry().WithError(commitError).Error("Failed to commit file")
		return [20]byte{}, commitError
	}

	return commit, nil
}

// PushChangesToRepository Pushes all committed changes in the repository to the remote repository
func (f TheGitUtils) PushChangesToRepository(username, password string, repository *git.Repository) error {
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
func (f TheGitUtils) PlainClone(username, password, serverUrl, directory string) (*git.Repository, error) {
	gitCloneOptions := git.CloneOptions{
		Auth: &http.BasicAuth{Username: username, Password: password},
		URL:  serverUrl,
	}
	repository, gitCloneError := git.PlainClone(directory, false, &gitCloneOptions)
	if gitCloneError != nil {
		log.Entry().WithError(gitCloneError).Error("Failed to clone git")
		return nil, gitCloneError
	}
	return repository, nil
}

// ChangeBranch checkout the provided branch.
// It will create a new branch if the branch does not exist yet.
// It will checkout "master" if no branch name if provided
func (f TheGitUtils) ChangeBranch(branchName string, repository *git.Repository) error {
	if branchName == "" {
		branchName = "master"
	}

	workTree, _ := repository.Worktree()
	var checkoutOptions = &git.CheckoutOptions{}
	checkoutOptions.Branch = plumbing.NewBranchReferenceName(branchName)
	checkoutOptions.Create = false
	checkoutError := workTree.Checkout(checkoutOptions)
	if checkoutError != nil {
		// branch might not exist, try to create branch
		checkoutOptions.Create = true
		checkoutError = workTree.Checkout(checkoutOptions)
		if checkoutError != nil {
			log.Entry().WithError(checkoutError).Error("Failed to checkout branch")
			return checkoutError
		}
	}

	return nil
}
