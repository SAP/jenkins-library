//go:build unit
// +build unit

package versioning

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestINIfileGetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		inifile := INIfile{
			path:           "my.cfg",
			versionSection: "test",
			readFile:       func(filename string) ([]byte, error) { return []byte("[test]\nversion = 1.2.3 "), nil },
		}
		version, err := inifile.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
	})

	t.Run("error case - read error", func(t *testing.T) {
		inifile := INIfile{
			path:     "my.cfg",
			readFile: func(filename string) ([]byte, error) { return []byte{}, fmt.Errorf("read error") },
		}
		_, err := inifile.GetVersion()
		assert.EqualError(t, err, "failed to read file 'my.cfg': read error")
	})

	t.Run("error case - load error", func(t *testing.T) {
		inifile := INIfile{
			path:     "my.cfg",
			readFile: func(filename string) ([]byte, error) { return []byte("1.2.3"), nil },
		}
		_, err := inifile.GetVersion()
		assert.EqualError(t, err, "failed to load content from file 'my.cfg': key-value delimiter not found: 1.2.3")
	})

	t.Run("error case - field not found", func(t *testing.T) {
		inifile := INIfile{
			path:     "my.cfg",
			readFile: func(filename string) ([]byte, error) { return []byte("theversion = 1.2.3"), nil },
		}
		_, err := inifile.GetVersion()
		assert.EqualError(t, err, "field 'version' not found in section ''")
	})
}

func TestINIfileSetVersion(t *testing.T) {
	t.Run("success case - flat", func(t *testing.T) {
		var content []byte
		inifile := INIfile{
			path:         "my.cfg",
			versionField: "theversion",
			readFile:     func(filename string) ([]byte, error) { return []byte("theversion = 1.2.3"), nil },
			writeFile:    func(filename string, filecontent []byte, mode os.FileMode) error { content = filecontent; return nil },
		}
		err := inifile.SetVersion("1.2.4")
		assert.NoError(t, err)
		assert.Contains(t, string(content), "theversion = 1.2.4")
	})

	t.Run("success case - section", func(t *testing.T) {
		var content []byte
		inifile := INIfile{
			path:           "my.cfg",
			versionField:   "theversion",
			versionSection: "test",
			readFile:       func(filename string) ([]byte, error) { return []byte("[test]\ntheversion = 1.2.3"), nil },
			writeFile:      func(filename string, filecontent []byte, mode os.FileMode) error { content = filecontent; return nil },
		}
		err := inifile.SetVersion("1.2.4")
		assert.NoError(t, err)
		assert.Contains(t, string(content), "[test]")
		assert.Contains(t, string(content), "1.2.4")
		assert.NotContains(t, string(content), "1.2.3")
	})

	t.Run("error case", func(t *testing.T) {
		inifile := INIfile{
			path:         "my.cfg",
			versionField: "theversion",
			readFile:     func(filename string) ([]byte, error) { return []byte("theversion = 1.2.3"), nil },
			writeFile:    func(filename string, filecontent []byte, mode os.FileMode) error { return fmt.Errorf("write error") },
		}
		err := inifile.SetVersion("1.2.4")
		assert.EqualError(t, err, "failed to write file 'my.cfg': write error")
	})
}
