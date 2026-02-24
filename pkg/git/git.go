package git

import (
	"fmt"
	"time"

	"errors"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
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
	plainOpen(path string) (*git.Repository, error)
}

// CommitSingleFile Commits the file located in the relative file path with the commitMessage to the given worktree.
// In case of errors, the error is returned. In the successful case the commit is provided.
func CommitSingleFile(filePath, commitMessage, author string, worktree *git.Worktree) (plumbing.Hash, error) {
	return commitSingleFile(filePath, commitMessage, author, worktree)
}

func commitSingleFile(filePath, commitMessage, author string, worktree utilsWorkTree) (plumbing.Hash, error) {
	_, err := worktree.Add(filePath)
	if err != nil {
		return [20]byte{}, fmt.Errorf("failed to add file to git: %w", err)
	}

	commit, err := worktree.Commit(commitMessage, &git.CommitOptions{
		All:    true,
		Author: &object.Signature{Name: author, When: time.Now()},
	})
	if err != nil {
		return [20]byte{}, fmt.Errorf("failed to commit file: %w", err)
	}

	return commit, nil
}

// PushChangesToRepository Pushes all committed changes in the repository to the remote repository
func PushChangesToRepository(username, password string, force *bool, repository *git.Repository, caCerts []byte) error {
	return pushChangesToRepository(username, password, force, repository, caCerts)
}

func pushChangesToRepository(username, password string, force *bool, repository utilsRepository, caCerts []byte) error {
	pushOptions := &git.PushOptions{
		Auth: &http.BasicAuth{Username: username, Password: password},
	}

	if len(caCerts) > 0 {
		pushOptions.CABundle = caCerts
	}
	if force != nil {
		pushOptions.Force = *force
	}
	err := repository.Push(pushOptions)
	if err != nil {
		return fmt.Errorf("failed to push commit: %w", err)
	}
	return nil
}

// PlainClone Clones a non-bare repository to the provided directory
func PlainClone(username, password, serverURL, branchName, directory string, caCerts []byte) (*git.Repository, error) {
	abstractedGit := &abstractionGit{}
	return plainClone(username, password, serverURL, branchName, directory, abstractedGit, caCerts)
}

func plainClone(username, password, serverURL, branchName, directory string, abstractionGit utilsGit, caCerts []byte) (*git.Repository, error) {
	gitCloneOptions := git.CloneOptions{
		Auth:          &http.BasicAuth{Username: username, Password: password},
		URL:           serverURL,
		ReferenceName: plumbing.NewBranchReferenceName(branchName),
		SingleBranch:  true, // we don't need other branches, clone only branchName
	}

	if len(caCerts) > 0 {
		gitCloneOptions.CABundle = caCerts
	}

	repository, err := abstractionGit.plainClone(directory, false, &gitCloneOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to clone git: %w", err)
	}
	return repository, nil
}

// PlainOpen opens a git repository from the given path
func PlainOpen(path string) (*git.Repository, error) {
	abstractedGit := &abstractionGit{}
	return plainOpen(path, abstractedGit)
}

func plainOpen(path string, abstractionGit utilsGit) (*git.Repository, error) {
	log.Entry().Infof("Opening git repo at '%s'", path)
	r, err := abstractionGit.plainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to open git repository at '%s': %w", path, err)
	}
	return r, nil
}

// ChangeBranch checkout the provided branch.
// It will create a new branch if the branch does not exist yet.
// It will return an error if no branch name if provided
func ChangeBranch(branchName string, worktree *git.Worktree) error {
	return changeBranch(branchName, worktree)
}

func changeBranch(branchName string, worktree utilsWorkTree) error {
	if branchName == "" {
		return errors.New("no branch name provided")
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
			return fmt.Errorf("failed to checkout branch: %w", err)
		}
	}

	return nil
}

// LogRange Returns a CommitIterator providing all commits reachable from 'to', but
// not reachable by 'from'.
func LogRange(repo *git.Repository, from, to string) (object.CommitIter, error) {

	cTo, err := getCommitObject(to, repo)
	if err != nil {
		return nil, fmt.Errorf("Cannot provide log range (to: '%s' not found): %w", to, err)
	}
	cFrom, err := getCommitObject(from, repo)
	if err != nil {
		return nil, fmt.Errorf("Cannot provide log range (from: '%s' not found): %w", from, err)
	}
	ignore := []plumbing.Hash{}
	err = object.NewCommitPreorderIter(
		cFrom,
		map[plumbing.Hash]bool{},
		[]plumbing.Hash{},
	).ForEach(func(c *object.Commit) error {
		ignore = append(ignore, c.ID())
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Cannot provide log range: %w", err)
	}

	return object.NewCommitPreorderIter(cTo, map[plumbing.Hash]bool{}, ignore), nil
}

func getCommitObject(ref string, repo *git.Repository) (*object.Commit, error) {
	if len(ref) == 0 {
		// with go-git v5.1.0 we panic otherwise inside ResolveRevision
		return nil, errors.New("Cannot get a commit for an empty ref")
	}
	r, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return nil, fmt.Errorf("Trouble resolving '%s': %w", ref, err)
	}
	c, err := repo.CommitObject(*r)
	if err != nil {
		return nil, fmt.Errorf("Trouble resolving '%s': %w", ref, err)
	}
	return c, nil
}

type abstractionGit struct{}

func (abstractionGit) plainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {
	return git.PlainClone(path, isBare, o)
}

func (abstractionGit) plainOpen(path string) (*git.Repository, error) {
	return git.PlainOpen(path)
}
