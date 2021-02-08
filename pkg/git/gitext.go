package git

import (
	"io"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// LogOptionsExt extends to evaluate a revision ranage.
type LogOptionsExt struct {
	Options *git.LogOptions
	// The log fil contain commits which are can not be found in the commits
	// which are reachable from the Except hash
	// This can be used for example to get unmerged commits,
	// query the logs like `git log dev..master` or `git log dev ^master`
	Except plumbing.Hash
}

// RepositoryExt extends the go-git repository
type RepositoryExt struct {
	Repo *git.Repository
}

type commitDifferenceIterator struct {
	except     map[plumbing.Hash]struct{}
	sourceIter object.CommitIter
	start      *object.Commit
}

// LogExt extends the go-git Log.
func (r *RepositoryExt) LogExt(ox *LogOptionsExt) (object.CommitIter, error) {
	it, err := r.Repo.Log(ox.Options)

	if !ox.Except.IsZero() {
		it, err = r.logDifference(it, ox.Except, ox.Options)

		if err != nil {
			return nil, err
		}
	}

	return it, nil
}

func (r *RepositoryExt) logDifference(ci object.CommitIter, except plumbing.Hash, o *git.LogOptions) (object.CommitIter, error) {
	options := *o
	options.From = except
	exceptLogs, err := r.Repo.Log(&options)

	if err != nil {
		return nil, err
	}

	seen := make(map[plumbing.Hash]struct{})

	exceptLogs.ForEach(func(c *object.Commit) error {
		seen[c.Hash] = struct{}{}
		return nil
	})

	iter := NewCommitDifferenceIterFromIter(seen, ci)
	return iter, nil
}

// NewCommitDifferenceIterFromIter returns a commit iter that walkd the commit
// history like WalkCommitHistory but filters out the commits which are not in
// the seen hash
func NewCommitDifferenceIterFromIter(except map[plumbing.Hash]struct{}, commitIter object.CommitIter) object.CommitIter {
	iterator := new(commitDifferenceIterator)
	iterator.except = except
	iterator.sourceIter = commitIter

	return iterator
}

func (c *commitDifferenceIterator) Next() (*object.Commit, error) {
	for {
		commit, err := c.sourceIter.Next()

		if err != nil {
			return nil, err
		}

		if _, ok := c.except[commit.Hash]; ok {
			continue
		}

		return commit, nil
	}
}

func (c *commitDifferenceIterator) ForEach(cb func(*object.Commit) error) error {
	for {
		commit, nextErr := c.Next()
		if nextErr == io.EOF {
			break
		}
		if nextErr != nil {
			return nextErr
		}
		err := cb(commit)
		if err == storer.ErrStop {
			return nil
		} else if err != nil {
			return err
		}
	}
	return nil
}

func (c *commitDifferenceIterator) Close() {
	c.sourceIter.Close()
}
