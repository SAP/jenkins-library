package piperenv

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetResourceParameter(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("Failed to create temporary directory")
	}

	// clean up tmp dir
	defer os.RemoveAll(dir)

	err = SetResourceParameter(dir, "testRes", "testParam", "testVal")

	assert.NoError(t, err, "Error occurred but none expected")
	assert.Equal(t, "testVal", GetResourceParameter(dir, "testRes", "testParam"))
}

func TestSetParameter(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("Failed to create temporary directory")
	}

	// clean up tmp dir
	defer os.RemoveAll(dir)

	err = SetParameter(dir, "testParam", "testVal")

	assert.NoError(t, err, "Error occurred but none expected")
	assert.Equal(t, "testVal", GetParameter(dir, "testParam"))
}

func TestReadFromDisk(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("Failed to create temporary directory")
	}

	// clean up tmp dir
	defer os.RemoveAll(dir)

	assert.Equal(t, "", GetParameter(dir, "testParamNotExistingYet"))
}
