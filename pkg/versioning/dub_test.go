package versioning

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDubInit(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		dub := Dub{}
		dub.init()
		assert.Equal(t, "dub.json", dub.DubJSONPath)
	})

	t.Run("no default", func(t *testing.T) {
		dub := Dub{DubJSONPath: "my/dub.json"}
		dub.init()
		assert.Equal(t, "my/dub.json", dub.DubJSONPath)
	})
}

func TestDubVersioningScheme(t *testing.T) {
	dub := Dub{}
	assert.Equal(t, "semver2", dub.VersioningScheme())
}

func TestDubGetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		dub := Dub{
			DubJSONPath: "my/dub.json",
			ReadFile:    func(filename string) ([]byte, error) { return []byte(`{"name": "test","version": "1.2.3"}`), nil },
		}
		version, err := dub.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
	})

	t.Run("error case", func(t *testing.T) {
		dub := Dub{
			DubJSONPath: "my/dub.json",
			ReadFile:    func(filename string) ([]byte, error) { return []byte{}, fmt.Errorf("read error") },
		}
		_, err := dub.GetVersion()
		assert.EqualError(t, err, "failed to read file 'my/dub.json': read error")
	})
}

func TestDubSetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		var content []byte
		dub := Dub{
			DubJSONPath: "my/dub.json",
			ReadFile:    func(filename string) ([]byte, error) { return []byte(`{"name": "test","version": "1.2.3"}`), nil },
			WriteFile:   func(filename string, filecontent []byte, mode os.FileMode) error { content = filecontent; return nil },
		}
		err := dub.SetVersion("1.2.4")
		assert.NoError(t, err)
		assert.Contains(t, string(content), "1.2.4")
	})

	t.Run("error case", func(t *testing.T) {
		dub := Dub{
			DubJSONPath: "my/dub.json",
			ReadFile:    func(filename string) ([]byte, error) { return []byte(`{"name": "test","version": "1.2.3"}`), nil },
			WriteFile:   func(filename string, filecontent []byte, mode os.FileMode) error { return fmt.Errorf("write error") },
		}
		err := dub.SetVersion("1.2.4")
		assert.EqualError(t, err, "failed to write file 'my/dub.json': write error")
	})
}
