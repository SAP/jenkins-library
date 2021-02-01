package transportrequest

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/go-git/go-git/v5/plumbing/object"
	"regexp"
	"sort"
)

// FindLabelsInCommits a label is considered to be something like
// key: label, e.g. TransportRequest: 123456
// These labels are expected to be contained in the git commit message as
// a separate line in the commit message body.
// In case several labels are found they are returned in ascending order.
func FindLabelsInCommits(commits object.CommitIter, label string) ([]string, error) {
	labelRegex, err := regexp.Compile(fmt.Sprintf(`(?m)^\s*%s\s*:\s*(\S*)\s*$`, label))
	if err != nil {
		return []string{}, fmt.Errorf("Cannot extract label: %w", err)
	}
	ids := []string{}
	err = commits.ForEach(func(c *object.Commit) error {
		for _, e := range labelRegex.FindAllStringSubmatch(c.Message, -1) {
			if len(e) < 2 { // the first entry is the full match, the second entry (at index 1) is the group
				return fmt.Errorf("Cannot extract label '%s' from commit '%s': '%s'", label, c.ID(), c.Message)
			}
			ids = append(ids, e[1])
		}
		return nil
	})
	if err != nil {
		return []string{}, fmt.Errorf("Cannot extract label: %w", err)
	}

	labels := piperutils.UniqueStrings(ids)
	sort.Strings(labels)
	return labels, nil
}
