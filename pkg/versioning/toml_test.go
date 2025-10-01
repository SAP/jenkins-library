//go:build unit
// +build unit

package versioning

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

const (
	sampleToml = `[project]
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

func TestPipTomlGetCoordinates(t *testing.T) {
	// t.Run("success case - pyproject.toml", func(t *testing.T) {
	// 	fileUtils := mock.FilesMock{}
	// 	fileUtils.AddFile("pyproject.toml", []byte(sampleToml))

	// 	pip := Toml{
	// 		Pip: Pip{
	// 			path:       "pyproject.toml",
	// 			fileExists: fileUtils.FileExists,
	// 			readFile:   fileUtils.FileRead,
	// 			writeFile:  fileUtils.FileWrite,
	// 		},
	// 	}

	// 	coordinates, err := pip.GetCoordinates()
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, "simple-python", coordinates.ArtifactID)
	// 	assert.Equal(t, "1.2.3", coordinates.Version)
	// })
	t.Run("success case - large pyproject.toml", func(t *testing.T) {
		fileUtils := mock.FilesMock{}
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

// func TestPipTomlSetVersion(t *testing.T) {
	// t.Run("success case - pyproject.toml", func(t *testing.T) {
	// 	fileUtils := mock.FilesMock{}
	// 	fileUtils.AddFile("pyproject.toml", []byte(sampleToml))

	// 	pip := Toml{
	// 		Pip: Pip{
	// 			path:       "pyproject.toml",
	// 			fileExists: fileUtils.FileExists,
	// 			readFile:   fileUtils.FileRead,
	// 			writeFile:  fileUtils.FileWrite,
	// 		},
	// 	}

	// 	coordinates, err := pip.GetCoordinates()
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, "simple-python", coordinates.ArtifactID)
	// 	assert.Equal(t, "1.2.3", coordinates.Version)
// 	// })
// 	t.Run("success case - large pyproject.toml", func(t *testing.T) {
// 		fileUtils := mock.FilesMock{}
// 		fileUtils.AddFile("pyproject.toml", []byte(largeSampleToml))

// 		pip := Pip{
// 			path:       "pyproject.toml",
// 			fileExists: fileUtils.FileExists,
// 			readFile:   fileUtils.FileRead,
// 			writeFile:  fileUtils.FileWrite,
// 		}

// 		err := pip.SetVersion("5.0.0")
// 		assert.NoError(t, err)
// 		// assert.Equal(t, "sampleproject", coordinates.ArtifactID)
// 		// assert.Equal(t, "5.0.0", coordinates.Version)

// 		assert.True(t, fileUtils.HasWrittenFile("pyproject.toml"))
// 	})
// }
