//go:build unit
// +build unit

package versioning

import (
	"fmt"
	"testing"

	piperMock "github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

const sampleCargoToml = `[package]
name = "gha-rust-hello-world"
version = "0.1.0"
edition = "2021"
`

const missingVersionCargoToml = `[package]
name = "gha-rust-hello-world"
edition = "2021"
`

const missingNameCargoToml = `[package]
version = "0.1.0"
edition = "2021"
`

// cargoTomlWithDepsVersion has a [dependencies] entry that pins the same version
// string as the package — SetVersion must not touch the dependency line.
const cargoTomlWithDepsVersion = `[package]
name = "gha-rust-hello-world"
version = "0.1.0"
edition = "2021"

[dependencies]
serde = { version = "0.1.0", features = ["derive"] }
`

// cargoTomlWithBinSection has a [[bin]] array-table after [package] whose header
// starts with "[" — the line-by-line parser must treat it as a section boundary.
const cargoTomlWithBinSection = `[package]
name = "gha-rust-hello-world"
version = "0.1.0"
edition = "2021"

[[bin]]
name = "cli"
path = "src/main.rs"
`

// cargoTomlVersionAfterMultiline has a [package] description using a triple-quoted
// multi-line string whose body contains a "["-prefixed line — the parser must not
// treat that line as a section header and must still find version below it.
const cargoTomlVersionAfterMultiline = `[package]
name = "gha-rust-hello-world"
description = """
See [README.md] for configuration.
"""
version = "0.1.0"
edition = "2021"
`

func TestCargoGetVersion(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		fileUtils := piperMock.FilesMock{}
		fileUtils.AddFile("Cargo.toml", []byte(sampleCargoToml))

		cargo := Cargo{path: "Cargo.toml", readFile: fileUtils.FileRead, writeFile: fileUtils.FileWrite}
		version, err := cargo.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "0.1.0", version)
	})
	t.Run("missing version field", func(t *testing.T) {
		t.Parallel()
		fileUtils := piperMock.FilesMock{}
		fileUtils.AddFile("Cargo.toml", []byte(missingVersionCargoToml))

		cargo := Cargo{path: "Cargo.toml", readFile: fileUtils.FileRead, writeFile: fileUtils.FileWrite}
		_, err := cargo.GetVersion()
		assert.ErrorContains(t, err, "no version information found in file 'Cargo.toml'")
	})
	t.Run("missing file", func(t *testing.T) {
		t.Parallel()
		fileUtils := piperMock.FilesMock{}

		cargo := Cargo{path: "Cargo.toml", readFile: fileUtils.FileRead, writeFile: fileUtils.FileWrite}
		_, err := cargo.GetVersion()
		assert.ErrorContains(t, err, "failed to read file 'Cargo.toml'")
	})
}

func TestCargoSetVersion(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		fileUtils := piperMock.FilesMock{}
		fileUtils.AddFile("Cargo.toml", []byte(sampleCargoToml))

		cargo := Cargo{path: "Cargo.toml", readFile: fileUtils.FileRead, writeFile: fileUtils.FileWrite}
		err := cargo.SetVersion("1.2.3")
		assert.NoError(t, err)

		cargo2 := Cargo{path: "Cargo.toml", readFile: fileUtils.FileRead, writeFile: fileUtils.FileWrite}
		version, err := cargo2.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
	})
	t.Run("does not modify dependency with same version", func(t *testing.T) {
		t.Parallel()
		fileUtils := piperMock.FilesMock{}
		fileUtils.AddFile("Cargo.toml", []byte(cargoTomlWithDepsVersion))

		cargo := Cargo{path: "Cargo.toml", readFile: fileUtils.FileRead, writeFile: fileUtils.FileWrite}
		err := cargo.SetVersion("2.0.0")
		assert.NoError(t, err)

		updatedBytes, _ := fileUtils.FileRead("Cargo.toml")
		updated := string(updatedBytes)
		assert.Contains(t, updated, `version = "2.0.0"`, "package version should be updated")
		assert.Contains(t, updated, `version = "0.1.0"`, "dependency version must remain unchanged")
	})

	t.Run("version before [[bin]] array-table is replaced", func(t *testing.T) {
		t.Parallel()
		fileUtils := piperMock.FilesMock{}
		fileUtils.AddFile("Cargo.toml", []byte(cargoTomlWithBinSection))

		cargo := Cargo{path: "Cargo.toml", readFile: fileUtils.FileRead, writeFile: fileUtils.FileWrite}
		err := cargo.SetVersion("2.0.0")
		assert.NoError(t, err)

		cargo2 := Cargo{path: "Cargo.toml", readFile: fileUtils.FileRead, writeFile: fileUtils.FileWrite}
		version, err := cargo2.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "2.0.0", version)
	})

	t.Run("version after triple-quoted description containing '[' is replaced", func(t *testing.T) {
		t.Parallel()
		fileUtils := piperMock.FilesMock{}
		fileUtils.AddFile("Cargo.toml", []byte(cargoTomlVersionAfterMultiline))

		cargo := Cargo{path: "Cargo.toml", readFile: fileUtils.FileRead, writeFile: fileUtils.FileWrite}
		err := cargo.SetVersion("2.0.0")
		assert.NoError(t, err)

		cargo2 := Cargo{path: "Cargo.toml", readFile: fileUtils.FileRead, writeFile: fileUtils.FileWrite}
		version, err := cargo2.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "2.0.0", version)
	})
}

func TestCargoGetCoordinates(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		fileUtils := piperMock.FilesMock{}
		fileUtils.AddFile("Cargo.toml", []byte(sampleCargoToml))

		cargo := Cargo{path: "Cargo.toml", readFile: fileUtils.FileRead, writeFile: fileUtils.FileWrite}
		coords, err := cargo.GetCoordinates()
		assert.NoError(t, err)
		assert.Equal(t, "gha-rust-hello-world", coords.ArtifactID)
		assert.Equal(t, "0.1.0", coords.Version)
	})
	t.Run("missing name", func(t *testing.T) {
		t.Parallel()
		fileUtils := piperMock.FilesMock{}
		fileUtils.AddFile("Cargo.toml", []byte(missingNameCargoToml))

		cargo := Cargo{path: "Cargo.toml", readFile: fileUtils.FileRead, writeFile: fileUtils.FileWrite}
		_, err := cargo.GetCoordinates()
		assert.ErrorContains(t, err, fmt.Sprintf("no name information found in file 'Cargo.toml'"))
	})
	t.Run("missing version", func(t *testing.T) {
		t.Parallel()
		fileUtils := piperMock.FilesMock{}
		fileUtils.AddFile("Cargo.toml", []byte(missingVersionCargoToml))

		cargo := Cargo{path: "Cargo.toml", readFile: fileUtils.FileRead, writeFile: fileUtils.FileWrite}
		_, err := cargo.GetCoordinates()
		assert.ErrorContains(t, err, "no version information found in file 'Cargo.toml'")
	})
}

func TestCargoVersioningScheme(t *testing.T) {
	t.Parallel()
	cargo := Cargo{}
	assert.Equal(t, "semver2", cargo.VersioningScheme())
}
