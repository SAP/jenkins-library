package npm

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

const (
	ignoreFilename = ".npmignore"
)

var (
	writeIgnoreFile = ioutil.WriteFile
)

func NewNPMIgnore(path string) NPMIgnore {
	if !strings.HasSuffix(path, ignoreFilename) {
		path = filepath.Join(path, ignoreFilename)
	}
	return NPMIgnore{filepath: path, values: []string{}}
}

type NPMIgnore struct {
	filepath string
	values   []string
}

func (ignorefile *NPMIgnore) Write() error {
	content := strings.Join(ignorefile.values, "\n")

	if err := writeIgnoreFile(ignorefile.filepath, []byte(content+"\n"), 0644); err != nil {
		return errors.Wrapf(err, "failed to write %s", ignorefile.filepath)
	}
	return nil
}

func (ignorefile *NPMIgnore) Load() error {
	file, err := os.Open(ignorefile.filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	ignorefile.values = lines
	return scanner.Err()
}

func (ignorefile *NPMIgnore) Add(value string) {
	ignorefile.values = append(ignorefile.values, value)
}
