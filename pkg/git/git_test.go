package git

import (
	"errors"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCommit(t *testing.T) {
	t.Parallel()
	t.Run("successful run", func(t *testing.T) {
		worktreeMock := WorktreeMock{}
		hash, err := commitSingleFile(".", "message", "user", &worktreeMock)
		assert.NoError(t, err)
		assert.Equal(t, plumbing.Hash([20]byte{4, 5, 6}), hash)
		assert.Equal(t, "user", worktreeMock.author)
		assert.True(t, worktreeMock.commitAll)
	})

	t.Run("error adding file", func(t *testing.T) {
		_, err := commitSingleFile(".", "message", "user", WorktreeMockFailing{
			failingAdd: true,
		})
		assert.Error(t, err)
		assert.EqualError(t, err, "failed to add file to git: failed to add file")
	})

	t.Run("error committing file", func(t *testing.T) {
		_, err := commitSingleFile(".", "message", "user", WorktreeMockFailing{
			failingCommit: true,
		})
		assert.Error(t, err)
		assert.EqualError(t, err, "failed to commit file: failed to commit file")
	})
}

func TestPushChangesToRepository(t *testing.T) {
	t.Parallel()
	t.Run("successful push", func(t *testing.T) {
		err := pushChangesToRepository("user", "password", RepositoryMock{
			test: t,
		})
		assert.NoError(t, err)
	})

	t.Run("error pushing", func(t *testing.T) {
		err := pushChangesToRepository("user", "password", RepositoryMockError{})
		assert.Error(t, err)
		assert.EqualError(t, err, "failed to push commit: error on push commits")
	})
}

func TestPlainClone(t *testing.T) {
	t.Parallel()
	t.Run("successful clone", func(t *testing.T) {
		abstractedGit := &UtilsGitMock{}
		_, err := plainClone("user", "password", "URL", "directory", abstractedGit)
		assert.NoError(t, err)
		assert.Equal(t, "directory", abstractedGit.path)
		assert.False(t, abstractedGit.isBare)
		assert.Equal(t, "http-basic-auth - user:*******", abstractedGit.authString)
		assert.Equal(t, "URL", abstractedGit.URL)
	})

	t.Run("error on cloning", func(t *testing.T) {
		abstractedGit := UtilsGitMockError{}
		_, err := plainClone("user", "password", "URL", "directory", abstractedGit)
		assert.Error(t, err)
		assert.EqualError(t, err, "failed to clone git: error during clone")
	})
}

func TestChangeBranch(t *testing.T) {
	t.Parallel()
	t.Run("checkout existing branch", func(t *testing.T) {
		worktreeMock := &WorktreeMock{}
		err := changeBranch("otherBranch", worktreeMock)
		assert.NoError(t, err)
		assert.Equal(t, string(plumbing.NewBranchReferenceName("otherBranch")), worktreeMock.checkedOutBranch)
		assert.False(t, worktreeMock.create)
	})

	t.Run("empty branch defaulted to master", func(t *testing.T) {
		worktreeMock := &WorktreeMock{}
		err := changeBranch("", worktreeMock)
		assert.NoError(t, err)
		assert.Equal(t, string(plumbing.NewBranchReferenceName("master")), worktreeMock.checkedOutBranch)
		assert.False(t, worktreeMock.create)
	})

	t.Run("create new branch", func(t *testing.T) {
		err := changeBranch("otherBranch", WorktreeUtilsNewBranch{})
		assert.NoError(t, err)
	})

	t.Run("error on new branch", func(t *testing.T) {
		err := changeBranch("otherBranch", WorktreeMockFailing{
			failingCheckout: true,
		})
		assert.Error(t, err)
		assert.EqualError(t, err, "failed to checkout branch: failed to checkout branch")
	})
}

type RepositoryMock struct {
	worktree *git.Worktree
	test     *testing.T
}

func (r RepositoryMock) Worktree() (*git.Worktree, error) {
	if r.worktree != nil {
		return r.worktree, nil
	}
	return &git.Worktree{}, nil
}

func (r RepositoryMock) Push(o *git.PushOptions) error {
	assert.Equal(r.test, "http-basic-auth - user:*******", o.Auth.String())
	return nil
}

type RepositoryMockError struct{}

func (RepositoryMockError) Worktree() (*git.Worktree, error) {
	return nil, errors.New("error getting worktree")
}

func (RepositoryMockError) Push(*git.PushOptions) error {
	return errors.New("error on push commits")
}

type WorktreeMockFailing struct {
	failingAdd      bool
	failingCommit   bool
	failingCheckout bool
}

func (w WorktreeMockFailing) Add(string) (plumbing.Hash, error) {
	if w.failingAdd {
		return [20]byte{}, errors.New("failed to add file")
	}
	return [20]byte{}, nil
}

func (w WorktreeMockFailing) Commit(string, *git.CommitOptions) (plumbing.Hash, error) {
	if w.failingCommit {
		return [20]byte{}, errors.New("failed to commit file")
	}
	return [20]byte{}, nil
}

func (w WorktreeMockFailing) Checkout(*git.CheckoutOptions) error {
	if w.failingCheckout {
		return errors.New("failed to checkout branch")
	}
	return nil
}

type WorktreeMock struct {
	expectedBranchName string
	checkedOutBranch   string
	create             bool
	author             string
	commitAll          bool
}

func (WorktreeMock) Add(string) (plumbing.Hash, error) {
	return [20]byte{1, 2, 3}, nil
}

func (w *WorktreeMock) Commit(_ string, options *git.CommitOptions) (plumbing.Hash, error) {
	w.author = options.Author.Name
	w.commitAll = options.All
	return [20]byte{4, 5, 6}, nil
}

func (w *WorktreeMock) Checkout(opts *git.CheckoutOptions) error {
	w.checkedOutBranch = string(opts.Branch)
	w.create = opts.Create
	return nil
}

type WorktreeUtilsNewBranch struct{}

func (WorktreeUtilsNewBranch) Add(string) (plumbing.Hash, error) {
	panic("implement me")
}

func (WorktreeUtilsNewBranch) Commit(string, *git.CommitOptions) (plumbing.Hash, error) {
	panic("implement me")
}

func (WorktreeUtilsNewBranch) Checkout(opts *git.CheckoutOptions) error {
	if opts.Create {
		return nil
	}
	return errors.New("branch already exists")
}

type UtilsGitMock struct {
	path       string
	isBare     bool
	authString string
	URL        string
}

func (u *UtilsGitMock) plainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {
	u.path = path
	u.isBare = isBare
	u.authString = o.Auth.String()
	u.URL = o.URL
	return nil, nil
}

type UtilsGitMockError struct{}

func (UtilsGitMockError) plainClone(string, bool, *git.CloneOptions) (*git.Repository, error) {
	return nil, errors.New("error during clone")
}
