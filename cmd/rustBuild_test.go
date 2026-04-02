//go:build unit
// +build unit

package cmd

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

const testCargoTomlContent = `[package]
name = "my-rust-app"
version = "1.2.3"
edition = "2021"
`

type rustBuildMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock

	returnFileUploadStatus int
	returnFileUploadError  error

	clientOptions []piperhttp.ClientOptions
	fileUploads   map[string]string
}

func (r *rustBuildMockUtils) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	r.AddFile(filename, []byte("content"))
	return nil
}

func (r *rustBuildMockUtils) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *rustBuildMockUtils) SetOptions(options piperhttp.ClientOptions) {
	r.clientOptions = append(r.clientOptions, options)
}

func (r *rustBuildMockUtils) UploadRequest(method, url, file, fieldName string, header http.Header, cookies []*http.Cookie, uploadType string) (*http.Response, error) {
	r.fileUploads[file] = url
	return &http.Response{StatusCode: r.returnFileUploadStatus}, r.returnFileUploadError
}

func (r *rustBuildMockUtils) UploadFile(url, file, fieldName string, header http.Header, cookies []*http.Cookie, uploadType string) (*http.Response, error) {
	return r.UploadRequest(http.MethodPut, url, file, fieldName, header, cookies, uploadType)
}

func (r *rustBuildMockUtils) Upload(data piperhttp.UploadRequestData) (*http.Response, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *rustBuildMockUtils) getDockerImageValue(stepName string) (string, error) {
	return "rust:1-bookworm", nil
}

func newRustBuildTestsUtils() *rustBuildMockUtils {
	utils := rustBuildMockUtils{
		ExecMockRunner:         &mock.ExecMockRunner{},
		FilesMock:              &mock.FilesMock{},
		returnFileUploadStatus: 201,
		fileUploads:            map[string]string{},
	}
	return &utils
}

func TestRunRustBuild(t *testing.T) {
	t.Parallel()

	cpe := rustBuildCommonPipelineEnvironment{}

	t.Run("success - build only", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{
			CargoProfile:        "release",
			TargetArchitectures: []string{"x86_64-unknown-linux-gnu"},
		}
		utils := newRustBuildTestsUtils()
		utils.AddFile("Cargo.toml", []byte(testCargoTomlContent))

		err := runRustBuild(&config, nil, utils, &cpe)
		assert.NoError(t, err)

		// rustup target add + cargo build
		assert.Equal(t, "rustup", utils.Calls[0].Exec)
		assert.Equal(t, []string{"target", "add", "x86_64-unknown-linux-gnu"}, utils.Calls[0].Params)
		assert.Equal(t, "cargo", utils.Calls[1].Exec)
		assert.Equal(t, []string{"build", "--profile", "release", "--target", "x86_64-unknown-linux-gnu"}, utils.Calls[1].Params)
	})

	t.Run("success - with tests", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{
			CargoProfile:        "release",
			RunTests:            true,
			TargetArchitectures: []string{"x86_64-unknown-linux-gnu"},
		}
		utils := newRustBuildTestsUtils()
		utils.AddFile("Cargo.toml", []byte(testCargoTomlContent))

		err := runRustBuild(&config, nil, utils, &cpe)
		assert.NoError(t, err)

		assert.Equal(t, "cargo", utils.Calls[0].Exec)
		assert.Equal(t, []string{"test", "--no-fail-fast"}, utils.Calls[0].Params)
	})

	t.Run("success - with tests and coverage tarpaulin", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{
			CargoProfile:        "release",
			RunTests:            true,
			ReportCoverage:      true,
			CoverageTool:        "tarpaulin",
			CoverageFormat:      "html",
			TargetArchitectures: []string{"x86_64-unknown-linux-gnu"},
		}
		utils := newRustBuildTestsUtils()
		utils.AddFile("Cargo.toml", []byte(testCargoTomlContent))

		err := runRustBuild(&config, nil, utils, &cpe)
		assert.NoError(t, err)

		// cargo test, cargo install cargo-tarpaulin, cargo tarpaulin
		assert.Equal(t, "cargo", utils.Calls[0].Exec)
		assert.Equal(t, []string{"test", "--no-fail-fast"}, utils.Calls[0].Params)
		assert.Equal(t, "cargo", utils.Calls[1].Exec)
		assert.Equal(t, []string{"install", "cargo-tarpaulin"}, utils.Calls[1].Params)
		assert.Equal(t, "cargo", utils.Calls[2].Exec)
		assert.Equal(t, []string{"tarpaulin", "--out", "Html", "--output-dir", "."}, utils.Calls[2].Params)
	})

	t.Run("success - with tests and coverage llvm-cov", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{
			CargoProfile:        "release",
			RunTests:            true,
			ReportCoverage:      true,
			CoverageTool:        "llvm-cov",
			CoverageFormat:      "cobertura",
			TargetArchitectures: []string{"x86_64-unknown-linux-gnu"},
		}
		utils := newRustBuildTestsUtils()
		utils.AddFile("Cargo.toml", []byte(testCargoTomlContent))

		err := runRustBuild(&config, nil, utils, &cpe)
		assert.NoError(t, err)

		assert.Equal(t, "cargo", utils.Calls[1].Exec)
		assert.Equal(t, []string{"install", "cargo-llvm-cov"}, utils.Calls[1].Params)
		assert.Equal(t, "cargo", utils.Calls[2].Exec)
		assert.Equal(t, []string{"llvm-cov", "--cobertura", "--output-path", "cobertura-coverage.xml"}, utils.Calls[2].Params)
	})

	t.Run("success - with integration tests", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{
			CargoProfile:        "release",
			RunIntegrationTests: true,
			TargetArchitectures: []string{"x86_64-unknown-linux-gnu"},
		}
		utils := newRustBuildTestsUtils()
		utils.AddFile("Cargo.toml", []byte(testCargoTomlContent))

		err := runRustBuild(&config, nil, utils, &cpe)
		assert.NoError(t, err)

		assert.Equal(t, "cargo", utils.Calls[0].Exec)
		assert.Equal(t, []string{"test", "--tests", "--no-fail-fast"}, utils.Calls[0].Params)
	})

	t.Run("success - with lint", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{
			CargoProfile:        "release",
			RunLint:             true,
			FailOnLintingError:  true,
			TargetArchitectures: []string{"x86_64-unknown-linux-gnu"},
		}
		utils := newRustBuildTestsUtils()
		utils.AddFile("Cargo.toml", []byte(testCargoTomlContent))

		err := runRustBuild(&config, nil, utils, &cpe)
		assert.NoError(t, err)

		assert.Equal(t, "cargo", utils.Calls[0].Exec)
		assert.Equal(t, []string{"clippy", "--all-targets", "--", "-D", "warnings"}, utils.Calls[0].Params)
	})

	t.Run("success - with BOM creation", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{
			CargoProfile:        "release",
			CreateBOM:           true,
			TargetArchitectures: []string{"x86_64-unknown-linux-gnu"},
		}
		utils := newRustBuildTestsUtils()
		utils.AddFile("Cargo.toml", []byte(testCargoTomlContent))

		err := runRustBuild(&config, nil, utils, &cpe)
		assert.NoError(t, err)

		assert.Equal(t, "cargo", utils.Calls[0].Exec)
		assert.Equal(t, []string{"install", "cargo-cyclonedx"}, utils.Calls[0].Params)
		assert.Equal(t, "cargo", utils.Calls[1].Exec)
		assert.Equal(t, []string{"cyclonedx", "--format", "xml"}, utils.Calls[1].Params)
	})

	t.Run("success - with publish", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{
			CargoProfile:        "release",
			TargetArchitectures: []string{"x86_64-unknown-linux-gnu"},
			Publish:             true,
			TargetRepositoryURL: "https://my.repo.local",
			ArtifactVersion:     "1.2.3",
			Output:              "my-rust-app",
		}
		utils := newRustBuildTestsUtils()
		utils.AddFile("Cargo.toml", []byte(testCargoTomlContent))
		utils.AddFile("target/x86_64-unknown-linux-gnu/release/my-rust-app", []byte("binary"))

		err := runRustBuild(&config, nil, utils, &cpe)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(utils.fileUploads))
		assert.Equal(t, "https://my.repo.local/rust/my-rust-app/1.2.3/my-rust-app-x86_64-unknown-linux-gnu", utils.fileUploads["my-rust-app-x86_64-unknown-linux-gnu"])
	})

	t.Run("success - publish with trailing slash in URL", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{
			CargoProfile:        "release",
			TargetArchitectures: []string{"x86_64-unknown-linux-gnu"},
			Publish:             true,
			TargetRepositoryURL: "https://my.repo.local/",
			ArtifactVersion:     "1.2.3",
			Output:              "my-rust-app",
		}
		utils := newRustBuildTestsUtils()
		utils.AddFile("Cargo.toml", []byte(testCargoTomlContent))
		utils.AddFile("target/x86_64-unknown-linux-gnu/release/my-rust-app", []byte("binary"))

		err := runRustBuild(&config, nil, utils, &cpe)
		assert.NoError(t, err)
		assert.Equal(t, "https://my.repo.local/rust/my-rust-app/1.2.3/my-rust-app-x86_64-unknown-linux-gnu", utils.fileUploads["my-rust-app-x86_64-unknown-linux-gnu"])
	})

	t.Run("success - multi-arch build", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{
			CargoProfile:        "release",
			TargetArchitectures: []string{"x86_64-unknown-linux-gnu", "aarch64-unknown-linux-gnu"},
		}
		utils := newRustBuildTestsUtils()
		utils.AddFile("Cargo.toml", []byte(testCargoTomlContent))

		err := runRustBuild(&config, nil, utils, &cpe)
		assert.NoError(t, err)

		// 2 x (rustup target add + cargo build) = 4 cargo/rustup calls before build settings
		assert.Equal(t, "rustup", utils.Calls[0].Exec)
		assert.Equal(t, []string{"target", "add", "x86_64-unknown-linux-gnu"}, utils.Calls[0].Params)
		assert.Equal(t, "cargo", utils.Calls[1].Exec)
		assert.Equal(t, []string{"build", "--profile", "release", "--target", "x86_64-unknown-linux-gnu"}, utils.Calls[1].Params)
		assert.Equal(t, "rustup", utils.Calls[2].Exec)
		assert.Equal(t, []string{"target", "add", "aarch64-unknown-linux-gnu"}, utils.Calls[2].Params)
		assert.Equal(t, "cargo", utils.Calls[3].Exec)
		assert.Equal(t, []string{"build", "--profile", "release", "--target", "aarch64-unknown-linux-gnu"}, utils.Calls[3].Params)
	})

	t.Run("success - with cargo features", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{
			CargoProfile:        "release",
			CargoFeatures:       []string{"feature1", "feature2"},
			TargetArchitectures: []string{"x86_64-unknown-linux-gnu"},
		}
		utils := newRustBuildTestsUtils()
		utils.AddFile("Cargo.toml", []byte(testCargoTomlContent))

		err := runRustBuild(&config, nil, utils, &cpe)
		assert.NoError(t, err)

		assert.Equal(t, "cargo", utils.Calls[1].Exec)
		assert.Contains(t, utils.Calls[1].Params, "--features")
		assert.Contains(t, utils.Calls[1].Params, "feature1,feature2")
	})

	t.Run("success - with custom build flags", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{
			CargoProfile:        "release",
			BuildFlags:          []string{"--locked", "--offline"},
			TargetArchitectures: []string{"x86_64-unknown-linux-gnu"},
		}
		utils := newRustBuildTestsUtils()
		utils.AddFile("Cargo.toml", []byte(testCargoTomlContent))

		err := runRustBuild(&config, nil, utils, &cpe)
		assert.NoError(t, err)

		assert.Equal(t, "cargo", utils.Calls[1].Exec)
		assert.Contains(t, utils.Calls[1].Params, "--locked")
		assert.Contains(t, utils.Calls[1].Params, "--offline")
	})

	t.Run("failure - build error", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{
			CargoProfile:        "release",
			TargetArchitectures: []string{"x86_64-unknown-linux-gnu"},
		}
		utils := newRustBuildTestsUtils()
		utils.AddFile("Cargo.toml", []byte(testCargoTomlContent))
		utils.ShouldFailOnCommand = map[string]error{"cargo build": fmt.Errorf("build failure")}

		err := runRustBuild(&config, nil, utils, &cpe)
		assert.EqualError(t, err, "failed to run cargo build for target x86_64-unknown-linux-gnu: build failure")
	})

	t.Run("failure - missing Cargo.toml", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{
			CargoProfile:        "release",
			TargetArchitectures: []string{"x86_64-unknown-linux-gnu"},
		}
		utils := newRustBuildTestsUtils()
		// no Cargo.toml added

		err := runRustBuild(&config, nil, utils, &cpe)
		assert.EqualError(t, err, "failed to read Cargo.toml: could not read 'Cargo.toml'")
	})

	t.Run("failure - lint failure with failOnLintingError=true", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{
			CargoProfile:        "release",
			RunLint:             true,
			FailOnLintingError:  true,
			TargetArchitectures: []string{"x86_64-unknown-linux-gnu"},
		}
		utils := newRustBuildTestsUtils()
		utils.AddFile("Cargo.toml", []byte(testCargoTomlContent))
		utils.ShouldFailOnCommand = map[string]error{"cargo clippy": fmt.Errorf("lint errors found")}

		err := runRustBuild(&config, nil, utils, &cpe)
		assert.EqualError(t, err, "cargo clippy reported linting errors: lint errors found")
	})

	t.Run("success - lint failure with failOnLintingError=false", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{
			CargoProfile:        "release",
			RunLint:             true,
			FailOnLintingError:  false,
			TargetArchitectures: []string{"x86_64-unknown-linux-gnu"},
		}
		utils := newRustBuildTestsUtils()
		utils.AddFile("Cargo.toml", []byte(testCargoTomlContent))
		utils.ShouldFailOnCommand = map[string]error{"cargo clippy": fmt.Errorf("lint errors found")}

		err := runRustBuild(&config, nil, utils, &cpe)
		assert.NoError(t, err)
	})

	t.Run("failure - publish without target repository URL", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{
			CargoProfile:        "release",
			TargetArchitectures: []string{"x86_64-unknown-linux-gnu"},
			Publish:             true,
			ArtifactVersion:     "1.2.3",
		}
		utils := newRustBuildTestsUtils()
		utils.AddFile("Cargo.toml", []byte(testCargoTomlContent))

		err := runRustBuild(&config, nil, utils, &cpe)
		assert.EqualError(t, err, "there's no target repository for binary publishing configured")
	})

	t.Run("failure - publish upload error", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{
			CargoProfile:        "release",
			TargetArchitectures: []string{"x86_64-unknown-linux-gnu"},
			Publish:             true,
			TargetRepositoryURL: "https://my.repo.local",
			ArtifactVersion:     "1.2.3",
			Output:              "my-rust-app",
		}
		utils := newRustBuildTestsUtils()
		utils.AddFile("Cargo.toml", []byte(testCargoTomlContent))
		utils.AddFile("target/x86_64-unknown-linux-gnu/release/my-rust-app", []byte("binary"))
		utils.returnFileUploadStatus = 500

		err := runRustBuild(&config, nil, utils, &cpe)
		assert.EqualError(t, err, "couldn't upload artifact, received status code 500")
	})
}

func TestReadCargoCoordinates(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		utils := newRustBuildTestsUtils()
		utils.AddFile("Cargo.toml", []byte(testCargoTomlContent))

		coords, err := readCargoCoordinates(utils)
		assert.NoError(t, err)
		assert.Equal(t, "my-rust-app", coords.ArtifactID)
		assert.Equal(t, "1.2.3", coords.Version)
	})

	t.Run("failure - missing Cargo.toml", func(t *testing.T) {
		t.Parallel()
		utils := newRustBuildTestsUtils()

		_, err := readCargoCoordinates(utils)
		assert.EqualError(t, err, "failed to read Cargo.toml: could not read 'Cargo.toml'")
	})

	t.Run("failure - invalid TOML", func(t *testing.T) {
		t.Parallel()
		utils := newRustBuildTestsUtils()
		utils.AddFile("Cargo.toml", []byte("not [ valid toml %%%"))

		_, err := readCargoCoordinates(utils)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse Cargo.toml")
	})
}

func TestPrepareRustEnvironment(t *testing.T) {
	t.Parallel()

	t.Run("success - no TLS certs, no registry token", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{}
		utils := newRustBuildTestsUtils()

		err := prepareRustEnvironment(&config, utils)
		assert.NoError(t, err)
	})

	t.Run("success - with registry token", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{CargoRegistryToken: "mytoken"}
		utils := newRustBuildTestsUtils()

		err := prepareRustEnvironment(&config, utils)
		assert.NoError(t, err)
		assert.Contains(t, utils.Env, "CARGO_REGISTRY_TOKEN=mytoken")
	})
}

func TestRunRustBuildPerArchitecture(t *testing.T) {
	t.Parallel()

	packageName := "my-rust-app"

	t.Run("success - default output", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{CargoProfile: "release"}
		utils := newRustBuildTestsUtils()

		binary, err := runRustBuildPerArchitecture(&config, packageName, utils, "x86_64-unknown-linux-gnu")
		assert.NoError(t, err)
		assert.Equal(t, "target/x86_64-unknown-linux-gnu/release/my-rust-app", binary)

		assert.Equal(t, "rustup", utils.Calls[0].Exec)
		assert.Equal(t, []string{"target", "add", "x86_64-unknown-linux-gnu"}, utils.Calls[0].Params)
		assert.Equal(t, "cargo", utils.Calls[1].Exec)
		assert.Equal(t, []string{"build", "--profile", "release", "--target", "x86_64-unknown-linux-gnu"}, utils.Calls[1].Params)
	})

	t.Run("success - custom output renames binary", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{CargoProfile: "release", Output: "out-bin"}
		utils := newRustBuildTestsUtils()
		// add source binary so FileRename finds something
		utils.AddFile("target/x86_64-unknown-linux-gnu/release/my-rust-app", []byte("binary"))

		binary, err := runRustBuildPerArchitecture(&config, packageName, utils, "x86_64-unknown-linux-gnu")
		assert.NoError(t, err)
		assert.Equal(t, "out-bin-x86_64-unknown-linux-gnu", binary)
	})

	t.Run("success - with features", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{CargoProfile: "release", CargoFeatures: []string{"f1", "f2"}}
		utils := newRustBuildTestsUtils()

		_, err := runRustBuildPerArchitecture(&config, packageName, utils, "x86_64-unknown-linux-gnu")
		assert.NoError(t, err)
		assert.Contains(t, utils.Calls[1].Params, "--features")
		assert.Contains(t, utils.Calls[1].Params, "f1,f2")
	})

	t.Run("failure - rustup target add fails", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{CargoProfile: "release"}
		utils := newRustBuildTestsUtils()
		utils.ShouldFailOnCommand = map[string]error{"rustup target add": fmt.Errorf("target not found")}

		_, err := runRustBuildPerArchitecture(&config, packageName, utils, "x86_64-unknown-linux-gnu")
		assert.EqualError(t, err, "failed to add rust target x86_64-unknown-linux-gnu: target not found")
	})

	t.Run("failure - cargo build fails", func(t *testing.T) {
		t.Parallel()
		config := rustBuildOptions{CargoProfile: "release"}
		utils := newRustBuildTestsUtils()
		utils.ShouldFailOnCommand = map[string]error{"cargo build": fmt.Errorf("compilation error")}

		_, err := runRustBuildPerArchitecture(&config, packageName, utils, "x86_64-unknown-linux-gnu")
		assert.EqualError(t, err, "failed to run cargo build for target x86_64-unknown-linux-gnu: compilation error")
	})
}
