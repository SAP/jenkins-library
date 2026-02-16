package transportrequest

import (
	"fmt"

	gitUtils "github.com/SAP/jenkins-library/pkg/git"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"

	"os"
	"regexp"
	"sort"
	"strings"
)

var logRange = gitUtils.LogRange
var findLabelsInCommits = FindLabelsInCommits

type iTransportRequestGitUtils interface {
	PlainOpen(directory string) (*git.Repository, error)
}
type transportRequestGitUtils struct {
}

func (g *transportRequestGitUtils) PlainOpen(directory string) (*git.Repository, error) {
	r, err := gitUtils.PlainOpen(directory)
	if err != nil {
		return nil, fmt.Errorf("Unable to open git repository at '%s': %w", directory, err)
	}
	return r, nil
}

// FindIDInRange finds a ID according to the label in a commit range <from>..<to>.
// We assume the git repo is present in the current working directory.
func FindIDInRange(label, from, to string) (string, error) {

	return findIDInRange(label, from, to, &transportRequestGitUtils{})
}

func findIDInRange(label, from, to string, trGitUtils iTransportRequestGitUtils) (string, error) {

	workdir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("Cannot open git repo in current working directory '%s': %w", workdir, err)
	}
	log.Entry().Infof("Opening git repo at '%s'", workdir)

	r, err := trGitUtils.PlainOpen(workdir)
	if err != nil {
		return "", fmt.Errorf("Unable to open git repository at '%s': %w", workdir, err)
	}

	cIter, err := logRange(r, from, to)
	if err != nil {
		return "", fmt.Errorf("Cannot retrieve '%s'. Unable to resolve commits in range '%s..%s': %w", label, from, to, err)
	}

	ids, err := findLabelsInCommits(cIter, label)
	if err != nil {
		return "", fmt.Errorf("Cannot retrieve '%s'. Unable to traverse commits in range '%s..%s': %w", label, from, to, err)
	}

	if len(ids) > 1 {
		return "", fmt.Errorf("More than one values found for label '%s' in range '%s..%s': '%s'", label, from, to, ids)
	}
	if len(ids) == 0 {
		return "", fmt.Errorf("No values found for '%s' in range '%s..%s'", label, from, to)
	}
	return ids[0], nil
}

// FindLabelsInCommits a label is considered to be something like
// key: label, e.g. TransportRequest: 123456
// These labels are expected to be contained in the git commit message as
// a separate line in the commit message body.
// In case several labels are found they are returned in ascending order.
func FindLabelsInCommits(commits object.CommitIter, label string) ([]string, error) {
	labelRegex, err := regexp.Compile(finishLabel(label))
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

func finishLabel(label string) string {
	// contains prefix, like the old default
	if strings.ContainsAny(label, ":=") {
		return fmt.Sprintf(`(?m)^\s*%s\s*(\S*)\s*$`, label)
	}
	// contains key only, like the new default
	return fmt.Sprintf(`(?m)^\s*%s\s*:\s*(\S*)\s*$`, label)
}
