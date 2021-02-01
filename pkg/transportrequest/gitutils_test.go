package transportrequest

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

type commitIteratorMock struct {
	commits []object.Commit
	index   int
}

func (iter *commitIteratorMock) Next() (*object.Commit, error) {
	i := iter.index
	iter.index++

	if i >= len(iter.commits) {
		return nil, io.EOF // real iterators also behave like this
	}
	return &iter.commits[i], nil
}

func (iter *commitIteratorMock) ForEach(cb func(c *object.Commit) error) error {
	for {
		c, err := iter.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		err = cb(c)
		if err == storer.ErrStop {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (iter *commitIteratorMock) Close() {

}

func TestRetrieveLabelStraightForward(t *testing.T) {

	t.Run("single commit tests", func(t *testing.T) {

		runTest := func(testConfig []string) {
			t.Run(testConfig[0], func(t *testing.T) {
				commitIter := &commitIteratorMock{
					commits: []object.Commit{
						object.Commit{
							Hash:    plumbing.NewHash("3434343434343434343434343434343434343434"),
							Message: testConfig[1],
						},
					},
				}
				labels, err := FindLabelsInCommits(commitIter, "TransportRequest")
				if assert.NoError(t, err) {
					expected := testConfig[2:]
					if assert.Len(t, labels, len(expected)) {
						assert.Subset(t, expected, labels)
					}
				}
			})
		}

		tests := [][]string{
			[]string{
				"straight forward",
				"this is a commit with TransportRequestId\n\nThis is the first line of the message body\nTransportRequest: 12345678",
				"12345678",
			},
			[]string{
				"trailing spaces after our value",
				"this is a commit with TransportRequestId\n\nThis is the first line of the message body\nTransportRequest: 12345678  ",
				"12345678",
			},
			[]string{
				"trailing text after our value",
				"this is a commit with TransportRequestId\n\nThis is the first line of the message body\nTransportRequest: 12345678 aaa",
			},
			[]string{
				"leading whitespace before our label",
				"this is a commit with TransportRequestId\n\nThis is the first line of the message body\n   TransportRequest: 12345678",
				"12345678",
			},
			[]string{
				"leading text before our label",
				"this is a commit with TransportRequestId\n\nThis is the first line of the message body\naaa TransportRequest: 12345678",
			},
			[]string{
				"whitespaces before column",
				"this is a commit with TransportRequestId\n\nThis is the first line of the message body\nTransportRequest  : 12345678",
				"12345678",
			},
			[]string{
				"no whitespaces after column",
				"this is a commit with TransportRequestId\n\nThis is the first line of the message body\nTransportRequest  :12345678",
				"12345678",
			},
			[]string{
				"two times the same id in the same commit",
				"this is a commit with TransportRequestId\n\nThis is the first line of the message body\nTransportRequest : 12345678\nTransportRequest : 12345678",
				"12345678",
			},
			[]string{
				// we report the ids, this is basically an invalid state, but needs to be filtered out by the caller
				"two different ids in the same commit",
				"this is a commit with TransportRequestId\n\nThis is the first line of the message body\nTransportRequest : 12345678\nTransportRequest : 87654321",
				"12345678", "87654321",
			},
		}

		for _, testConfig := range tests {
			runTest(testConfig)
		}
	})

	t.Run("multi commit tests", func(t *testing.T) {

		t.Run("two different ids in different commits", func(t *testing.T) {
			commitIter := &commitIteratorMock{
				commits: []object.Commit{
					object.Commit{
						Hash:    plumbing.NewHash("3434343434343434343434343434343434343434"),
						Message: "this is a commit with TransportRequestId\n\nThis is the first line of the message body\nTransportRequest: 12345678",
					},
					object.Commit{
						Hash:    plumbing.NewHash("1212121212121212121212121212121212121212"),
						Message: "this is a commit with TransportRequestId\n\nThis is the first line of the message body\nTransportRequest: 87654321",
					},
				},
			}
			labels, err := FindLabelsInCommits(commitIter, "TransportRequest")
			if assert.NoError(t, err) {
				assert.Equal(t, []string{"12345678", "87654321"}, labels)
			}
		})

		t.Run("two different ids in different commits agains, order needs to be the same", func(t *testing.T) {
			commitIter := &commitIteratorMock{
				commits: []object.Commit{
					object.Commit{
						Hash:    plumbing.NewHash("1212121212121212121212121212121212121212"),
						Message: "this is a commit with TransportRequestId\n\nThis is the first line of the message body\nTransportRequest: 87654321",
					},
					object.Commit{
						Hash:    plumbing.NewHash("3434343434343434343434343434343434343434"),
						Message: "this is a commit with TransportRequestId\n\nThis is the first line of the message body\nTransportRequest: 12345678",
					},
				},
			}
			labels, err := FindLabelsInCommits(commitIter, "TransportRequest")
			if assert.NoError(t, err) {
				assert.Equal(t, []string{"12345678", "87654321"}, labels)
			}
		})

		t.Run("the same id in different commits", func(t *testing.T) {
			commitIter := &commitIteratorMock{
				commits: []object.Commit{
					object.Commit{
						Hash:    plumbing.NewHash("3434343434343434343434343434343434343434"),
						Message: "this is a commit with TransportRequestId\n\nThis is the first line of the message body\nTransportRequest: 12345678",
					},
					object.Commit{
						Hash:    plumbing.NewHash("1212121212121212121212121212121212121212"),
						Message: "this is a commit with TransportRequestId\n\nThis is the first line of the message body\nTransportRequest: 12345678",
					},
				},
			}
			labels, err := FindLabelsInCommits(commitIter, "TransportRequest")
			if assert.NoError(t, err) {
				expected := []string{"12345678"}
				if assert.Len(t, labels, len(expected)) {
					assert.Subset(t, expected, labels)
				}
			}
		})
	})
}
