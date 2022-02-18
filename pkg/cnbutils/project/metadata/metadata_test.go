package metadata

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/stretchr/testify/assert"
)

func TestWriteProjectMetadata(t *testing.T) {
	expectedResult := `
[source]
  type = "git"

  [source.metadata]
    refs = ["main"]

  [source.version]
    commit = "012548"
    describe = "test-commit"
`
	mockUtils := &cnbutils.MockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}

	fileutils := piperutils.Files{}

	cpeFiles := map[string]string{
		"headCommitId":  "012548",
		"commitMessage": "test-commit",
		"branch":        "main",
	}

	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	err = os.MkdirAll(filepath.Join(dir, "commonPipelineEnvironment", "git"), os.ModePerm)
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	for file, content := range cpeFiles {
		err = fileutils.FileWrite(filepath.Join(dir, "commonPipelineEnvironment", "git", file), []byte(content), os.ModePerm)
		assert.NoError(t, err)
	}

	WriteProjectMetadata(dir, mockUtils)
	assert.True(t, mockUtils.HasWrittenFile(metadataFilePath))
	result, err := mockUtils.FileRead(metadataFilePath)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, string(result))
}
