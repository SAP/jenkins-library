package git

import (
	"errors"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCommitSingleFileErrorAddingFile(t *testing.T) {
	var cut = TheGitUtils{}
	_, err := cut.CommitSingleFile(".", "message", WorktreeMockFailingAdd{})
	assert.Error(t, err)
	assert.Equal(t, errors.New("failed to add file"), err)
}

func TestCommitSingleFileErrorCommittingFile(t *testing.T) {
	var cut = TheGitUtils{}
	_, err := cut.CommitSingleFile(".", "message", WorktreeMockFailingCommit{})
	assert.Error(t, err)
	assert.Equal(t, errors.New("error on commit"), err)
}

func TestCommitSingleFile(t *testing.T) {
	var cut = TheGitUtils{}
	hash, err := cut.CommitSingleFile(".", "message", WorktreeMock{
		test: t,
	})
	assert.NoError(t, err)
	assert.Equal(t, plumbing.Hash([20]byte{4, 5, 6}), hash)
}

func TestPushChangesToRepository(t *testing.T) {
	var cut = TheGitUtils{}
	err := cut.PushChangesToRepository("user", "password", RepositoryMock{
		test: t,
	})
	assert.NoError(t, err)
}

func TestPushChangesToRepositoryErrorPushing(t *testing.T) {
	var cut = TheGitUtils{}
	err := cut.PushChangesToRepository("user", "password", RepositoryMockError{})
	assert.Error(t, err)
	assert.Equal(t, errors.New("error on push commits"), err)
}

func TestPlainClone(t *testing.T) {
	var cut = TheGitUtils{}
	oldUtilsGit := abstractedGit
	defer func() {
		abstractedGit = oldUtilsGit
	}()
	abstractedGit = UtilsGitMock{
		test: t,
	}
	_, err := cut.PlainClone("user", "password", "URL", "directory")
	assert.NoError(t, err)
}

func TestPlainCloneErrorOnCloning(t *testing.T) {
	var cut = TheGitUtils{}
	oldUtilsGit := abstractedGit
	defer func() {
		abstractedGit = oldUtilsGit
	}()
	abstractedGit = UtilsGitMockError{}
	_, err := cut.PlainClone("user", "password", "URL", "directory")
	assert.Error(t, err)
	assert.Equal(t, errors.New("error during clone"), err)
}

func TestChangeBranch(t *testing.T) {
	var cut = TheGitUtils{}
	err := cut.ChangeBranch("otherBranch", WorktreeMock{
		expectedBranchName: "otherBranch",
		test:               t,
	})
	assert.NoError(t, err)
}

func TestChangeBranchMaster(t *testing.T) {
	var cut = TheGitUtils{}
	err := cut.ChangeBranch("", WorktreeMock{
		expectedBranchName: "master",
		test:               t,
	})
	assert.NoError(t, err)
}

func TestChangeBranchNewBranch(t *testing.T) {
	var cut = TheGitUtils{}
	err := cut.ChangeBranch("otherBranch", WorktreeUtilsNewBranch{})
	assert.NoError(t, err)
}

func TestChangeBranchNewBranchCannotBeCreated(t *testing.T) {
	var cut = TheGitUtils{}
	err := cut.ChangeBranch("otherBranch", WorktreeUtilsErrorCheckout{})
	assert.Error(t, err)
	assert.Equal(t, errors.New("cannot checkout branch"), err)
}

func TestGetWorktree(t *testing.T) {
	var cut = TheGitUtils{}
	testWorktree := &git.Worktree{}
	worktree, err := cut.GetWorktree(RepositoryMock{
		worktree: testWorktree,
		test:     t,
	})
	assert.NoError(t, err)
	assert.Equal(t, testWorktree, worktree)
}

func TestGetWorktreeError(t *testing.T) {
	var cut = TheGitUtils{}
	_, err := cut.GetWorktree(RepositoryMockError{})
	assert.Error(t, err)
	assert.Equal(t, errors.New("error getting worktree"), err)
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

func (RepositoryMockError) Push(o *git.PushOptions) error {
	return errors.New("error on push commits")
}

type WorktreeMockFailingAdd struct{}

func (WorktreeMockFailingAdd) Add(path string) (plumbing.Hash, error) {
	return [20]byte{}, errors.New("failed to add file")
}

func (WorktreeMockFailingAdd) Commit(msg string, opts *git.CommitOptions) (plumbing.Hash, error) {
	panic("implement me")
}

func (WorktreeMockFailingAdd) Checkout(opts *git.CheckoutOptions) error {
	panic("implement me")
}

type WorktreeMockFailingCommit struct{}

func (WorktreeMockFailingCommit) Add(path string) (plumbing.Hash, error) {
	return [20]byte{1, 2, 3}, nil
}

func (WorktreeMockFailingCommit) Commit(msg string, opts *git.CommitOptions) (plumbing.Hash, error) {
	return [20]byte{}, errors.New("error on commit")
}

func (WorktreeMockFailingCommit) Checkout(opts *git.CheckoutOptions) error {
	panic("implement me")
}

type WorktreeMock struct {
	expectedBranchName string
	test               *testing.T
}

func (WorktreeMock) Add(path string) (plumbing.Hash, error) {
	return [20]byte{1, 2, 3}, nil
}

func (WorktreeMock) Commit(msg string, opts *git.CommitOptions) (plumbing.Hash, error) {
	return [20]byte{4, 5, 6}, nil
}

func (w WorktreeMock) Checkout(opts *git.CheckoutOptions) error {
	assert.Equal(w.test, plumbing.NewBranchReferenceName(w.expectedBranchName), opts.Branch)
	assert.False(w.test, opts.Create)
	return nil
}

type WorktreeUtilsNewBranch struct{}

func (WorktreeUtilsNewBranch) Add(path string) (plumbing.Hash, error) {
	panic("implement me")
}

func (WorktreeUtilsNewBranch) Commit(msg string, opts *git.CommitOptions) (plumbing.Hash, error) {
	panic("implement me")
}

func (WorktreeUtilsNewBranch) Checkout(opts *git.CheckoutOptions) error {
	if opts.Create {
		return nil
	}
	return errors.New("branch already exists")
}

type WorktreeUtilsErrorCheckout struct{}

func (WorktreeUtilsErrorCheckout) Add(path string) (plumbing.Hash, error) {
	panic("implement me")
}

func (WorktreeUtilsErrorCheckout) Commit(msg string, opts *git.CommitOptions) (plumbing.Hash, error) {
	panic("implement me")
}

func (WorktreeUtilsErrorCheckout) Checkout(opts *git.CheckoutOptions) error {
	return errors.New("cannot checkout branch")
}

type UtilsGitMock struct {
	test *testing.T
}

func (u UtilsGitMock) plainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {
	assert.Equal(u.test, "directory", path)
	assert.False(u.test, isBare)
	assert.Equal(u.test, "http-basic-auth - user:*******", o.Auth.String())
	assert.Equal(u.test, "URL", o.URL)
	return nil, nil
}

type UtilsGitMockError struct{}

func (UtilsGitMockError) plainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {
	return nil, errors.New("error during clone")
}
