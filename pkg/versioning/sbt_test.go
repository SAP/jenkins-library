package versioning

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSbtInit(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		Sbt := Sbt{}
		Sbt.init()
		assert.Equal(t, "sbtDescriptor.json", Sbt.DescriptorPath)
	})

	t.Run("no default", func(t *testing.T) {
		Sbt := Sbt{DescriptorPath: "my/sbtDescriptor.json"}
		Sbt.init()
		assert.Equal(t, "my/sbtDescriptor.json", Sbt.DescriptorPath)
	})
}

func TestSbtVersioningScheme(t *testing.T) {
	Sbt := Sbt{}
	assert.Equal(t, "semver2", Sbt.VersioningScheme())
}

func TestSbtGetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		Sbt := Sbt{
			DescriptorPath: "my/sbtDescriptor.json",
			ReadFile:       func(filename string) ([]byte, error) { return []byte(`{"version": "1.2.3"}`), nil },
		}
		version, err := Sbt.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
	})

	t.Run("error case", func(t *testing.T) {
		Sbt := Sbt{
			DescriptorPath: "my/sbtDescriptor.json",
			ReadFile:       func(filename string) ([]byte, error) { return []byte{}, fmt.Errorf("read error") },
		}
		_, err := Sbt.GetVersion()
		assert.EqualError(t, err, "failed to read file 'my/sbtDescriptor.json': read error")
	})
}

func TestSbtSetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		var content []byte
		Sbt := Sbt{
			DescriptorPath: "my/sbtDescriptor.json",
			ReadFile:       func(filename string) ([]byte, error) { return []byte(`{"version": "1.2.3"}`), nil },
			WriteFile:      func(filename string, filecontent []byte, mode os.FileMode) error { content = filecontent; return nil },
		}
		err := Sbt.SetVersion("1.2.4")
		assert.NoError(t, err)
		assert.Contains(t, string(content), "1.2.4")
	})

	t.Run("error case", func(t *testing.T) {
		Sbt := Sbt{
			DescriptorPath: "my/sbtDescriptor.json",
			ReadFile:       func(filename string) ([]byte, error) { return []byte(`{"version": "1.2.3"}`), nil },
			WriteFile:      func(filename string, filecontent []byte, mode os.FileMode) error { return fmt.Errorf("write error") },
		}
		err := Sbt.SetVersion("1.2.4")
		assert.EqualError(t, err, "failed to write file 'my/sbtDescriptor.json': write error")
	})
}
