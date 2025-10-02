//go:build unit
// +build unit

package versioning

import (
	"fmt"
	"testing"

	piperMock "github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	invalidToml = `[project]`
	sampleToml  = `[project]
name = "simple-python"
version = "1.2.3"
`
	largeSampleToml = `[project]
name = "sampleproject"
version = "4.0.0"
description = "A sample Python project"
license = { file = "LICENSE.txt" }

authors = [{ name = "A. Random Developer", email = "author@example.com" }]
requires-python = ">=3.9"
readme = "README.md"


maintainers = [{ name = "A. Great Maintainer", email = "maintainer@example.com" }]
keywords = [
    "sample",
    "setuptools",
    "development",
]
classifiers = [
    "Development Status :: 3 - Alpha",
    "Intended Audience :: Developers",
    "Topic :: Software Development :: Build Tools",
    "Programming Language :: Python :: 3",
    "Programming Language :: Python :: 3.9",
    "Programming Language :: Python :: 3.10",
    "Programming Language :: Python :: 3.11",
    "Programming Language :: Python :: 3.12",
    "Programming Language :: Python :: 3.13",
    "Programming Language :: Python :: 3 :: Only",
]
dependencies = ["peppercorn"]

[project.optional-dependencies]
dev = ["check-manifest"]
test = ["coverage"]

[project.urls]
Homepage = "https://github.com/pypa/sampleproject"
"Bug Reports" = "https://github.com/pypa/sampleproject/issues"
Funding = "https://donate.pypi.org"
"Say Thanks!" = "http://saythanks.io/to/example"
Source = "https://github.com/pypa/sampleproject/"

[project.scripts]
sample = "sample:main"

[build-system]
# A list of packages that are needed to build your package:
requires = ["setuptools"] # REQUIRED if [build-system] table is used
# The name of the Python object that frontends will use to perform the build:
build-backend = "setuptools.build_meta" # If not defined, then legacy behavior can happen.

[tool.uv]
package = false

[tool.setuptools]
# If there are data files included in your packages that need to be
# installed, specify them here.
package-data = { "sample" = ["*.dat"] }
`
)

func TestTomlSetVersion(t *testing.T) {
	t.Run("success case - large pyproject.toml", func(t *testing.T) {
		fileUtils := piperMock.FilesMock{}
		fileUtils.AddFile("pyproject.toml", []byte(largeSampleToml))

		toml := Toml{
			Pip: Pip{
				path:       "pyproject.toml",
				fileExists: fileUtils.FileExists,
				readFile:   fileUtils.FileRead,
				writeFile:  fileUtils.FileWrite,
			},
		}

		coordinates, err := toml.GetCoordinates()
		assert.NoError(t, err)
		assert.Equal(t, "sampleproject", coordinates.ArtifactID)
		assert.Equal(t, "4.0.0", coordinates.Version)

		// test SetVersion
		err = toml.SetVersion("5.0.0")
		assert.NoError(t, err)
		coordinates, err = toml.GetCoordinates()
		assert.NoError(t, err)
		assert.Equal(t, "sampleproject", coordinates.ArtifactID)
		assert.Equal(t, "5.0.0", coordinates.Version)
	})
}

func TestTomlGetCoordinates(t *testing.T) {
	t.Parallel()
	t.Run("success case - pyproject.toml", func(t *testing.T) {
		filename := TomlBuildDescriptor
		fileUtils := piperMock.FilesMock{}
		fileUtils.AddFile(filename, []byte(sampleToml))

		pip := Toml{
			Pip: Pip{
				path:       filename,
				fileExists: fileUtils.FileExists,
				readFile:   fileUtils.FileRead,
				writeFile:  fileUtils.FileWrite,
			},
		}

		coordinates, err := pip.GetCoordinates()
		assert.NoError(t, err)
		assert.Equal(t, "simple-python", coordinates.ArtifactID)
		assert.Equal(t, "1.2.3", coordinates.Version)
	})
	t.Run("fail - invalid pyproject.toml", func(t *testing.T) {
		filename := TomlBuildDescriptor
		fileUtils := piperMock.FilesMock{}
		fileUtils.AddFile(filename, []byte(invalidToml))

		pip := Toml{
			Pip: Pip{
				path:       filename,
				fileExists: fileUtils.FileExists,
				readFile:   fileUtils.FileRead,
				writeFile:  fileUtils.FileWrite,
			},
		}

		coordinates, err := pip.GetCoordinates()
		assert.ErrorContains(t, err, fmt.Sprintf("no version information found in file '%s'", filename))
		assert.Equal(t, "", coordinates.ArtifactID)
		assert.Equal(t, "", coordinates.Version)
	})
	t.Run("fail - empty pyproject.toml", func(t *testing.T) {
		filename := TomlBuildDescriptor
		fileUtils := piperMock.FilesMock{}
		fileUtils.AddFile(filename, []byte(""))

		toml := Toml{
			Pip: Pip{
				path:       filename,
				fileExists: fileUtils.FileExists,
				readFile:   fileUtils.FileRead,
				writeFile:  fileUtils.FileWrite,
			},
		}

		coordinates, err := toml.GetCoordinates()
		assert.ErrorContains(t, err, fmt.Sprintf("no version information found in file '%s'", filename))
		assert.Equal(t, "", coordinates.ArtifactID)
		assert.Equal(t, "", coordinates.Version)
	})
	t.Run("fail - no pyproject.toml", func(t *testing.T) {
		filename := mock.Anything
		fileUtils := piperMock.FilesMock{}
		fileUtils.AddFile(filename, []byte(""))

		toml := Toml{
			Pip: Pip{
				path:       filename,
				fileExists: fileUtils.FileExists,
				readFile:   fileUtils.FileRead,
				writeFile:  fileUtils.FileWrite,
			},
		}

		coordinates, err := toml.GetCoordinates()
		assert.ErrorContains(t, err, fmt.Sprintf("file '%s' is not a pyproject.toml", filename))
		assert.Equal(t, "", coordinates.ArtifactID)
		assert.Equal(t, "", coordinates.Version)
	})
	t.Run("fail - missing pyproject.toml", func(t *testing.T) {
		filename := TomlBuildDescriptor
		fileUtils := piperMock.FilesMock{}

		toml := Toml{
			Pip: Pip{
				path:       filename,
				fileExists: fileUtils.FileExists,
				readFile:   fileUtils.FileRead,
				writeFile:  fileUtils.FileWrite,
			},
		}

		coordinates, err := toml.GetCoordinates()
		assert.ErrorContains(t, err, fmt.Sprintf("failed to read file '%s'", filename))
		assert.Equal(t, "", coordinates.ArtifactID)
		assert.Equal(t, "", coordinates.Version)
	})
}
