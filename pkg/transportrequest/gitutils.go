package transportrequest

import (
	"fmt"
	pipergit "github.com/SAP/jenkins-library/pkg/git"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"
	"os"
	"regexp"
)

// needs to be replaced by mocks in the tests
var getWorkDirectory = os.Getwd

// FindIDInRange finds a ID according to the label in a commit range <from>..to.
// We assume the git repo is present in the current working directory.
func FindIDInRange(label, from, to string) (string, error) {
	workdir, err := getWorkDirectory()
	if err != nil {
		return "", errors.Wrapf(err, "Cannot retrieve %s", label)
	}
	log.Entry().Infof("Opening git repo at '%s'", workdir)
	r, err := git.PlainOpen(workdir) // TODO this we need to mock also
	if err != nil {
		return "", errors.Wrapf(err, "Cannot retrieve '%s'. Unable to open git repository at '%s'", label, workdir)
	}

	cIter, err := pipergit.LogRange(r, from, to)
	if err != nil {
		return "", errors.Wrapf(err, "Cannot retrieve '%s'. Unable to resolve commits in range '%s..%s'", label, from, to)
	}

	ids, err := FindLabelsInCommits(cIter, label) // TOOD not sure if we should mock this since there are already tests ...
	if err != nil {
		return "", errors.Wrapf(err, "Cannot retrieve '%s'. Unable to traverse commits in range '%s..%s'", label, from, to)
	}

	if len(ids) > 1 {
		return "", fmt.Errorf("More than one values found for '%s' in range '%s..%s': ", label, from, to, ids)
	}
	if len(ids) == 0 {
		return "", fmt.Errorf("No values found for '%s' in range '%s..%s'", label, from, to)
	}
	return ids[0], nil
}

// FindLabelsInCommits ...
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

	return piperutils.UniqueStrings(ids), nil
}
