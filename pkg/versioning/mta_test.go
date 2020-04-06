package versioning

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMtaInit(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		mta := Mta{}
		mta.init()
		assert.Equal(t, "mta.yaml", mta.MtaYAMLPath)
	})

	t.Run("no default", func(t *testing.T) {
		mta := Mta{MtaYAMLPath: "my/mta.yaml"}
		mta.init()
		assert.Equal(t, "my/mta.yaml", mta.MtaYAMLPath)
	})
}

func TestMtaVersioningScheme(t *testing.T) {
	mta := Mta{}
	assert.Equal(t, "semver2", mta.VersioningScheme())
}

func TestMtaGetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		mta := Mta{
			MtaYAMLPath: "my/mta.yaml",
			ReadFile:    func(filename string) ([]byte, error) { return []byte(`version: 1.2.3`), nil },
		}
		version, err := mta.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
	})

	t.Run("error case", func(t *testing.T) {
		mta := Mta{
			MtaYAMLPath: "my/mta.yaml",
			ReadFile:    func(filename string) ([]byte, error) { return []byte{}, fmt.Errorf("read error") },
		}
		_, err := mta.GetVersion()
		assert.EqualError(t, err, "failed to read file 'my/mta.yaml': read error")
	})
}

func TestMtaSetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		var content []byte
		mta := Mta{
			MtaYAMLPath: "my/mta.yaml",
			ReadFile:    func(filename string) ([]byte, error) { return []byte(`version: 1.2.3`), nil },
			WriteFile:   func(filename string, filecontent []byte, mode os.FileMode) error { content = filecontent; return nil },
		}
		err := mta.SetVersion("1.2.4")
		assert.NoError(t, err)
		assert.Contains(t, string(content), "1.2.4")
	})

	t.Run("error case", func(t *testing.T) {
		mta := Mta{
			MtaYAMLPath: "my/mta.yaml",
			ReadFile:    func(filename string) ([]byte, error) { return []byte(`version: 1.2.3`), nil },
			WriteFile:   func(filename string, filecontent []byte, mode os.FileMode) error { return fmt.Errorf("write error") },
		}
		err := mta.SetVersion("1.2.4")
		assert.EqualError(t, err, "failed to write file 'my/mta.yaml': write error")
	})
}
