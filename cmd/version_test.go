//go:build unit

package cmd

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func TestVersion(t *testing.T) {

	t.Run("versionAndTagInitialValues", func(t *testing.T) {

		result := runVersionCommand(t, "", "")
		assert.Contains(t, result, "commit: \"<n/a>\"")
		assert.Contains(t, result, "tag: \"<n/a>\"")
	})

	t.Run("versionAndTagSet", func(t *testing.T) {

		result := runVersionCommand(t, "16bafe", "v1.2.3")
		assert.Contains(t, result, "commit: \"16bafe\"")
		assert.Contains(t, result, "tag: \"v1.2.3\"")
	})
}

func runVersionCommand(t *testing.T, commitID, tag string) string {

	orig := os.Stdout
	defer func() { os.Stdout = orig }()

	r, w, e := os.Pipe()
	if e != nil {
		t.Error("Cannot setup pipes.")
	}

	os.Stdout = w

	//
	// needs to be set in the free wild by the build process:
	// go build -ldflags "-X github.com/SAP/jenkins-library/cmd.GitCommit=${GIT_COMMIT} -X github.com/SAP/jenkins-library/cmd.GitTag=${GIT_TAG}"
	if len(commitID) > 0 {
		GitCommit = commitID
	}
	if len(tag) > 0 {
		GitTag = tag
	}
	defer func() { GitCommit = ""; GitTag = "" }()
	//
	//

	version()

	w.Close()

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}
