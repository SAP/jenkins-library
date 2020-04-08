package versioning

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONfileGetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		jsonfile := JSONfile{
			Path:     "my.json",
			ReadFile: func(filename string) ([]byte, error) { return []byte(`{"version": "1.2.3"}`), nil },
		}
		version, err := jsonfile.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
	})

	t.Run("error case", func(t *testing.T) {
		jsonfile := JSONfile{
			Path:         "my.json",
			VersionField: "theversion",
			ReadFile:     func(filename string) ([]byte, error) { return []byte{}, fmt.Errorf("read error") },
		}
		_, err := jsonfile.GetVersion()
		assert.EqualError(t, err, "failed to read file 'my.json': read error")
	})
}

func TestJSONfileSetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		var content []byte
		jsonfile := JSONfile{
			Path:         "my.json",
			VersionField: "theversion",
			ReadFile:     func(filename string) ([]byte, error) { return []byte(`{"theversion": "1.2.3"}`), nil },
			WriteFile:    func(filename string, filecontent []byte, mode os.FileMode) error { content = filecontent; return nil },
		}
		err := jsonfile.SetVersion("1.2.4")
		assert.NoError(t, err)
		assert.Contains(t, string(content), `"theversion": "1.2.4"`)
	})

	t.Run("error case", func(t *testing.T) {
		jsonfile := JSONfile{
			Path:         "my.json",
			VersionField: "theversion",
			ReadFile:     func(filename string) ([]byte, error) { return []byte(`{"theversion": "1.2.3"}`), nil },
			WriteFile:    func(filename string, filecontent []byte, mode os.FileMode) error { return fmt.Errorf("write error") },
		}
		err := jsonfile.SetVersion("1.2.4")
		assert.EqualError(t, err, "failed to write file 'my.json': write error")
	})
}
