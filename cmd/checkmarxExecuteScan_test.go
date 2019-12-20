package cmd

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fileInfo struct {
	nam     string      // base name of the file
	siz     int64       // length in bytes for regular files; system-dependent for others
	mod     os.FileMode // file mode bits
	modtime time.Time   // modification time
	dir     bool        // abbreviation for Mode().IsDir()
	syss    interface{} // underlying data source (can return nil)
}

func (fi fileInfo) IsDir() bool {
	return fi.dir
}
func (fi fileInfo) Name() string {
	return fi.nam
}
func (fi fileInfo) Size() int64 {
	return fi.siz
}
func (fi fileInfo) ModTime() time.Time {
	return fi.modtime
}
func (fi fileInfo) Mode() os.FileMode {
	return fi.mod
}
func (fi fileInfo) Sys() interface{} {
	return fi.syss
}

func TestFilterFileGlob(t *testing.T) {
	tt := []struct {
		input    string
		fInfo    fileInfo
		expected bool
	}{
		{input: "somepath/node_modules/someOther/some.file", fInfo: fileInfo{}, expected: true},
		{input: "somepath/non_modules/someOther/some.go", fInfo: fileInfo{}, expected: false},
		{input: ".xmake/someOther/some.go", fInfo: fileInfo{}, expected: true},
		{input: "another/vendor/some.html", fInfo: fileInfo{}, expected: false},
		{input: "another/vendor/some.pdf", fInfo: fileInfo{}, expected: true},
		{input: "another/vendor/some.test", fInfo: fileInfo{}, expected: true},
		{input: "some.test", fInfo: fileInfo{}, expected: false},
		{input: "a/b/c", fInfo: fileInfo{dir: true}, expected: false},
	}

	for k, v := range tt {
		assert.Equal(t, v.expected, filterFileGlob([]string{"!**/node_modules/**", "!**/.xmake/**", "!**/*_test.go", "!**/vendor/**/*.go", "**/*.go", "**/*.html", "*.test"}, v.input, v.fInfo), fmt.Sprintf("wrong long name for run %v", k))
	}
}

func TestZipFolder(t *testing.T) {

	t.Run("zip files", func(t *testing.T) {
		var zipFileMock bytes.Buffer
		zipFolder(".", &zipFileMock, []string{"!checkmarxExecuteScan_test.go", "**/*.txt", "**/checkmarxExecuteScan.go"})

		got := zipFileMock.Len()
		want := 3204

		if got != want {
			t.Errorf("Zipping test failed expected %v but got %v", want, got)
		}
	})
}
