//go:build unit
// +build unit

package git

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
)

func TestCommit(t *testing.T) {
	t.Parallel()
	t.Run("successful run", func(t *testing.T) {
		t.Parallel()
		worktreeMock := WorktreeMock{}
		hash, err := commitSingleFile(".", "message", "user", &worktreeMock)
		assert.NoError(t, err)
		assert.Equal(t, plumbing.Hash([20]byte{4, 5, 6}), hash)
		assert.Equal(t, "user", worktreeMock.author)
		assert.True(t, worktreeMock.commitAll)
	})

	t.Run("error adding file", func(t *testing.T) {
		t.Parallel()
		_, err := commitSingleFile(".", "message", "user", WorktreeMockFailing{
			failingAdd: true,
		})
		assert.EqualError(t, err, "failed to add file to git: failed to add file")
	})

	t.Run("error committing file", func(t *testing.T) {
		t.Parallel()
		_, err := commitSingleFile(".", "message", "user", WorktreeMockFailing{
			failingCommit: true,
		})
		assert.EqualError(t, err, "failed to commit file: failed to commit file")
	})
}

func TestPushChangesToRepository(t *testing.T) {
	t.Parallel()
	t.Run("successful push", func(t *testing.T) {
		t.Parallel()
		err := pushChangesToRepository("user", "password", nil, RepositoryMock{
			test: t,
		}, []byte{})
		assert.NoError(t, err)
	})

	t.Run("error pushing", func(t *testing.T) {
		t.Parallel()
		err := pushChangesToRepository("user", "password", nil, RepositoryMockError{}, []byte{})
		assert.EqualError(t, err, "failed to push commit: error on push commits")
	})
}

func TestPlainClone(t *testing.T) {
	t.Parallel()
	t.Run("successful clone", func(t *testing.T) {
		t.Parallel()
		abstractedGit := &UtilsGitMock{}
		_, err := plainClone("user", "password", "URL", "", "directory", abstractedGit, []byte{})
		assert.NoError(t, err)
		assert.Equal(t, "directory", abstractedGit.path)
		assert.False(t, abstractedGit.isBare)
		assert.Equal(t, "http-basic-auth - user:*******", abstractedGit.authString)
		assert.Equal(t, "URL", abstractedGit.URL)
	})

	t.Run("error on cloning", func(t *testing.T) {
		t.Parallel()
		abstractedGit := UtilsGitMockError{}
		_, err := plainClone("user", "password", "URL", "", "directory", abstractedGit, []byte{})
		assert.EqualError(t, err, "failed to clone git: error during clone")
	})
}

func TestPlainOpenMock(t *testing.T) {
	t.Parallel()
	t.Run("successful clone", func(t *testing.T) {
		t.Parallel()
		abstractedGit := &UtilsGitMock{}
		_, err := plainOpen("directory", abstractedGit)
		assert.NoError(t, err)
		assert.Equal(t, "directory", abstractedGit.path)
	})

	t.Run("error on cloning", func(t *testing.T) {
		t.Parallel()
		abstractedGit := UtilsGitMockError{}
		_, err := plainOpen("directory", abstractedGit)
		assert.EqualError(t, err, "Unable to open git repository at 'directory': error during git plain open")
	})
}

func TestChangeBranch(t *testing.T) {
	t.Parallel()
	t.Run("checkout existing branch", func(t *testing.T) {
		t.Parallel()
		worktreeMock := &WorktreeMock{}
		err := changeBranch("otherBranch", worktreeMock)
		assert.NoError(t, err)
		assert.Equal(t, string(plumbing.NewBranchReferenceName("otherBranch")), worktreeMock.checkedOutBranch)
		assert.False(t, worktreeMock.create)
	})

	t.Run("empty branch raises error", func(t *testing.T) {
		t.Parallel()
		worktreeMock := &WorktreeMock{}
		err := changeBranch("", worktreeMock)
		assert.EqualError(t, err, "no branch name provided")
	})

	t.Run("create new branch", func(t *testing.T) {
		t.Parallel()
		err := changeBranch("otherBranch", WorktreeUtilsNewBranch{})
		assert.NoError(t, err)
	})

	t.Run("error on new branch", func(t *testing.T) {
		t.Parallel()
		err := changeBranch("otherBranch", WorktreeMockFailing{
			failingCheckout: true,
		})
		assert.EqualError(t, err, "failed to checkout branch: failed to checkout branch")
	})
}

func TestLogRange(t *testing.T) {

	against := func(t *testing.T, r *git.Repository, from, to string, expected []plumbing.Hash) {
		seen := []plumbing.Hash{}
		cIter, err := LogRange(r, from, to)
		if assert.NoError(t, err) {
			err = cIter.ForEach(func(c *object.Commit) error {
				seen = append(seen, c.ID())
				return nil
			})
			if assert.NoError(t, err) {
				if assert.Len(t, seen, len(expected)) {
					assert.Subset(t, seen, expected)
				}
			}
		}
	}

	prepareRepo := func() (r *git.Repository, hashes map[string]plumbing.Hash, err error) {

		hashes = map[string]plumbing.Hash{}

		// Creates a commit
		c := func(r *git.Repository, fs billy.Filesystem, name string) (hash plumbing.Hash, err error) {

			if val, ok := hashes[name]; ok {
				err = fmt.Errorf("Cannot create commit for name '%s'. There is already a commit available (%s) for that name", name, val)
				return
			}
			w, err := r.Worktree()
			if err != nil {
				return
			}
			f, err := fs.Create(fmt.Sprintf("commit%s.txt", name))
			if err != nil {
				return
			}
			_, err = f.Write([]byte(fmt.Sprintf("Commit %s", name)))
			if err != nil {
				return
			}
			_, err = w.Add(fmt.Sprintf("commit%s.txt", name))
			if err != nil {
				return
			}

			hash, err = w.Commit(fmt.Sprintf("Commit %s", name), &git.CommitOptions{})
			if err != nil {
				return
			}
			hashes[name] = hash
			return
		}

		// Creates a branch on the currently checked out commit
		b := func(r *git.Repository, name string) (err error) {
			w, err := r.Worktree()
			if err != nil {
				return
			}

			b := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", name))

			err = w.Checkout(&git.CheckoutOptions{Create: true, Force: false, Branch: b})

			return
		}

		fs := memfs.New()

		// create new git repo
		r, err = git.Init(memory.NewStorage(), fs)
		if err != nil {
			return
		}

		config, err := r.Config()
		if err != nil {
			return
		}

		config.User.Name = "me"
		config.User.Email = "me@example.org"

		err = r.SetConfig(config)
		if err != nil {
			return
		}

		w, err := r.Worktree()
		if err != nil {
			return
		}

		// add a commit to the repo -- A --
		hashA, err := c(r, fs, "A")
		if err != nil {
			return
		}

		_, err = r.CreateTag("initial", hashA,
			&git.CreateTagOptions{
				Tagger: &object.Signature{
					Name:  config.User.Name,
					Email: config.User.Email,
				},
				Message: "initial",
			})
		if err != nil {
			return
		}

		// another commit -- B --
		_, err = c(r, fs, "B")
		if err != nil {
			return
		}

		// checkout the first commit again
		err = w.Checkout(&git.CheckoutOptions{Hash: hashA})
		if err != nil {
			return
		}

		// add another file as sucessor of the first commit -- C --
		_, err = c(r, fs, "C")
		if err != nil {
			return
		}

		// and yet another commit -- D --
		_, err = c(r, fs, "D")
		if err != nil {
			return
		}

		err = b(r, "branch1")

		return
	}

	r, hashes, err := prepareRepo()
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("%v", err), err)
	}

	// Our repo contains these commits and branches:
	//
	//  / C - D <-- HEAD <-- branch1
	// A - B <-- master
	//
	// Tag 'initial' sits on commit A

	t.Parallel()

	t.Run("B against D", func(t *testing.T) {
		against(t, r, hashes["B"].String(), hashes["D"].String(), []plumbing.Hash{hashes["C"], hashes["D"]})
	})
	t.Run("A against B", func(t *testing.T) {
		against(t, r, hashes["A"].String(), hashes["B"].String(), []plumbing.Hash{hashes["B"]})
	})
	t.Run("B against HEAD", func(t *testing.T) {
		against(t, r, hashes["B"].String(), "HEAD", []plumbing.Hash{hashes["C"], hashes["D"]})
	})
	t.Run("B against HEAD~1", func(t *testing.T) {
		against(t, r, hashes["B"].String(), "HEAD~1", []plumbing.Hash{hashes["C"]})
	})
	t.Run("A against master", func(t *testing.T) {
		against(t, r, hashes["A"].String(), "master", []plumbing.Hash{hashes["B"]})
	})
	t.Run("master against a branch pointing to D", func(t *testing.T) {
		against(t, r, "master", "branch1", []plumbing.Hash{hashes["C"], hashes["D"]})
	})
	t.Run("A against master~1", func(t *testing.T) {
		against(t, r, hashes["A"].String(), "master~1", []plumbing.Hash{})
	})
	t.Run("Tag against C", func(t *testing.T) {
		against(t, r, "initial", hashes["C"].String(), []plumbing.Hash{hashes["C"]})
	})

	t.Run("Same ref results in empty result", func(t *testing.T) {
		against(t, r, hashes["A"].String(), hashes["A"].String(), []plumbing.Hash{})
	})
	t.Run("Invalid ref", func(t *testing.T) {
		// it is unlikely as hell, but at some time we might get a test failure here
		// when a commit with this hash has been created during preparation of the repo.
		// Maybe we should check first if a commit with that hash exists and if so try
		// another hash. paranoia :-)
		_, err := LogRange(r, "0123456789012345678901234567890123456789", "HEAD")
		assert.EqualError(t, err, "Cannot provide log range (from: '0123456789012345678901234567890123456789' not found): Trouble resolving '0123456789012345678901234567890123456789': reference not found")
	})
	t.Run("Empty string as ref", func(t *testing.T) {
		_, err := LogRange(r, "", "HEAD")
		assert.EqualError(t, err, "Cannot provide log range (from: '' not found): Cannot get a commit for an empty ref")
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

func (u *UtilsGitMock) plainOpen(path string) (*git.Repository, error) {
	u.path = path
	return nil, nil
}

type UtilsGitMockError struct{}

func (UtilsGitMockError) plainClone(string, bool, *git.CloneOptions) (*git.Repository, error) {
	return nil, errors.New("error during clone")
}

func (UtilsGitMockError) plainOpen(path string) (*git.Repository, error) {
	return nil, errors.New("error during git plain open")
}
