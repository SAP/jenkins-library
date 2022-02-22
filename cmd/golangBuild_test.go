package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/multiarch"
	"github.com/SAP/jenkins-library/pkg/telemetry"

	"github.com/stretchr/testify/assert"

	"golang.org/x/mod/modfile"
)

type golangBuildMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock

	returnFileUploadStatus int   // expected to be set upfront
	returnFileUploadError  error // expected to be set upfront

	clientOptions []piperhttp.ClientOptions // set by mock
	fileUploads   map[string]string         // set by mock
}

func (g *golangBuildMockUtils) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	return fmt.Errorf("not implemented")
}

func (g *golangBuildMockUtils) GetRepositoryURL(module string) (string, error) {
	return fmt.Sprintf("https://%s.git", module), nil
}

func (g *golangBuildMockUtils) SendRequest(method string, url string, r io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	return nil, fmt.Errorf("not implemented")
}

func (g *golangBuildMockUtils) SetOptions(options piperhttp.ClientOptions) {
	g.clientOptions = append(g.clientOptions, options)
}

func (g *golangBuildMockUtils) UploadRequest(method, url, file, fieldName string, header http.Header, cookies []*http.Cookie, uploadType string) (*http.Response, error) {
	g.fileUploads[file] = url

	response := http.Response{
		StatusCode: g.returnFileUploadStatus,
	}

	return &response, g.returnFileUploadError
}

func (g *golangBuildMockUtils) UploadFile(url, file, fieldName string, header http.Header, cookies []*http.Cookie, uploadType string) (*http.Response, error) {
	return g.UploadRequest(http.MethodPut, url, file, fieldName, header, cookies, uploadType)
}

func (g *golangBuildMockUtils) Upload(data piperhttp.UploadRequestData) (*http.Response, error) {
	return nil, fmt.Errorf("not implemented")
}

func (g *golangBuildMockUtils) getDockerImageValue(stepName string) (string, error) {
	return "golang:latest", nil
}

func newGolangBuildTestsUtils() *golangBuildMockUtils {
	utils := golangBuildMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
		//clientOptions:  []piperhttp.ClientOptions{},
		fileUploads: map[string]string{},
	}
	return &utils
}

func TestRunGolangBuild(t *testing.T) {
	cpe := golangBuildCommonPipelineEnvironment{}

	t.Run("success - no tests", func(t *testing.T) {
		config := golangBuildOptions{
			TargetArchitectures: []string{"linux,amd64"},
			EntryPointPath:      "cmd/server.go",
		}
		utils := newGolangBuildTestsUtils()
		telemetryData := telemetry.CustomData{}

		err := runGolangBuild(&config, &telemetryData, utils, &cpe)
		assert.NoError(t, err)
		assert.Equal(t, "go", utils.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{"build", "-trimpath", "cmd/server.go"}, utils.ExecMockRunner.Calls[0].Params)
	})

	t.Run("success - tests & ldflags", func(t *testing.T) {
		config := golangBuildOptions{
			RunTests:            true,
			LdflagsTemplate:     "test",
			TargetArchitectures: []string{"linux,amd64"},
		}
		utils := newGolangBuildTestsUtils()
		telemetryData := telemetry.CustomData{}

		err := runGolangBuild(&config, &telemetryData, utils, &cpe)
		assert.NoError(t, err)
		assert.Equal(t, "go", utils.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{"install", "gotest.tools/gotestsum@latest"}, utils.ExecMockRunner.Calls[0].Params)
		assert.Equal(t, "gotestsum", utils.ExecMockRunner.Calls[1].Exec)
		assert.Equal(t, []string{"--junitfile", "TEST-go.xml", "--", fmt.Sprintf("-coverprofile=%v", coverageFile), "./..."}, utils.ExecMockRunner.Calls[1].Params)
		assert.Equal(t, "go", utils.ExecMockRunner.Calls[2].Exec)
		assert.Equal(t, []string{"build", "-trimpath", "-ldflags", "test"}, utils.ExecMockRunner.Calls[2].Params)
	})

	t.Run("success - tests with coverage", func(t *testing.T) {
		config := golangBuildOptions{
			RunTests:            true,
			ReportCoverage:      true,
			TargetArchitectures: []string{"linux,amd64"},
		}
		utils := newGolangBuildTestsUtils()
		telemetryData := telemetry.CustomData{}

		err := runGolangBuild(&config, &telemetryData, utils, &cpe)
		assert.NoError(t, err)
		assert.Equal(t, "go", utils.ExecMockRunner.Calls[2].Exec)
		assert.Equal(t, []string{"tool", "cover", "-html", coverageFile, "-o", "coverage.html"}, utils.ExecMockRunner.Calls[2].Params)
	})

	t.Run("success - integration tests", func(t *testing.T) {
		config := golangBuildOptions{
			RunIntegrationTests: true,
			TargetArchitectures: []string{"linux,amd64"},
		}
		utils := newGolangBuildTestsUtils()
		telemetryData := telemetry.CustomData{}

		err := runGolangBuild(&config, &telemetryData, utils, &cpe)
		assert.NoError(t, err)
		assert.Equal(t, "go", utils.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{"install", "gotest.tools/gotestsum@latest"}, utils.ExecMockRunner.Calls[0].Params)
		assert.Equal(t, "gotestsum", utils.ExecMockRunner.Calls[1].Exec)
		assert.Equal(t, []string{"--junitfile", "TEST-integration.xml", "--", "-tags=integration", "./..."}, utils.ExecMockRunner.Calls[1].Params)
		assert.Equal(t, "go", utils.ExecMockRunner.Calls[2].Exec)
		assert.Equal(t, []string{"build", "-trimpath"}, utils.ExecMockRunner.Calls[2].Params)
	})

	t.Run("success - publishes binaries", func(t *testing.T) {
		config := golangBuildOptions{
			TargetArchitectures:      []string{"linux,amd64"},
			Output:                   "testBin",
			Publish:                  true,
			TargetRepositoryURL:      "https://my.target.repository.local",
			TargetRepositoryUser:     "user",
			TargetRepositoryPassword: "password",
			ArtifactVersion:          "1.0.0",
		}
		utils := newGolangBuildTestsUtils()
		utils.returnFileUploadStatus = 201
		utils.FilesMock.AddFile("go.mod", []byte("module example.com/my/module"))
		telemetryData := telemetry.CustomData{}

		err := runGolangBuild(&config, &telemetryData, utils, &cpe)
		if assert.NoError(t, err) {
			assert.Equal(t, "go", utils.ExecMockRunner.Calls[0].Exec)
			assert.Equal(t, []string{"build", "-trimpath", "-o", "testBin-linux.amd64"}, utils.ExecMockRunner.Calls[0].Params)

			assert.Equal(t, 1, len(utils.fileUploads))
			assert.Equal(t, "https://my.target.repository.local/go/example.com/my/module/1.0.0/testBin-linux.amd64", utils.fileUploads["testBin-linux.amd64"])
		}
	})

	t.Run("success - publishes binaries (when TargetRepositoryURL ends with slash)", func(t *testing.T) {
		config := golangBuildOptions{
			TargetArchitectures:      []string{"linux,amd64"},
			Output:                   "testBin",
			Publish:                  true,
			TargetRepositoryURL:      "https://my.target.repository.local/",
			TargetRepositoryUser:     "user",
			TargetRepositoryPassword: "password",
			ArtifactVersion:          "1.0.0",
		}
		utils := newGolangBuildTestsUtils()
		utils.returnFileUploadStatus = 200
		utils.FilesMock.AddFile("go.mod", []byte("module example.com/my/module"))
		telemetryData := telemetry.CustomData{}

		err := runGolangBuild(&config, &telemetryData, utils, &cpe)
		if assert.NoError(t, err) {
			assert.Equal(t, "go", utils.ExecMockRunner.Calls[0].Exec)
			assert.Equal(t, []string{"build", "-trimpath", "-o", "testBin-linux.amd64"}, utils.ExecMockRunner.Calls[0].Params)

			assert.Equal(t, 1, len(utils.fileUploads))
			assert.Equal(t, "https://my.target.repository.local/go/example.com/my/module/1.0.0/testBin-linux.amd64", utils.fileUploads["testBin-linux.amd64"])
		}
	})

	t.Run("success - create BOM", func(t *testing.T) {
		config := golangBuildOptions{
			CreateBOM:           true,
			TargetArchitectures: []string{"linux,amd64"},
		}
		utils := newGolangBuildTestsUtils()
		telemetryData := telemetry.CustomData{}

		err := runGolangBuild(&config, &telemetryData, utils, &cpe)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(utils.ExecMockRunner.Calls))
		assert.Equal(t, "go", utils.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{"install", "github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest"}, utils.ExecMockRunner.Calls[0].Params)
		assert.Equal(t, "cyclonedx-gomod", utils.ExecMockRunner.Calls[1].Exec)
		assert.Equal(t, []string{"mod", "-licenses", "-test", "-output", "bom.xml"}, utils.ExecMockRunner.Calls[1].Params)
		assert.Equal(t, "go", utils.ExecMockRunner.Calls[2].Exec)
		assert.Equal(t, []string{"build", "-trimpath"}, utils.ExecMockRunner.Calls[2].Params)
	})

	t.Run("failure - install pre-requisites for testing", func(t *testing.T) {
		config := golangBuildOptions{
			RunTests: true,
		}
		utils := newGolangBuildTestsUtils()
		utils.ShouldFailOnCommand = map[string]error{"go install gotest.tools/gotestsum": fmt.Errorf("install failure")}
		telemetryData := telemetry.CustomData{}

		err := runGolangBuild(&config, &telemetryData, utils, &cpe)
		assert.EqualError(t, err, "failed to install pre-requisite: install failure")
	})

	t.Run("failure - install pre-requisites for BOM creation", func(t *testing.T) {
		config := golangBuildOptions{
			CreateBOM: true,
		}
		utils := newGolangBuildTestsUtils()
		utils.ShouldFailOnCommand = map[string]error{"go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest": fmt.Errorf("install failure")}
		telemetryData := telemetry.CustomData{}

		err := runGolangBuild(&config, &telemetryData, utils, &cpe)
		assert.EqualError(t, err, "failed to install pre-requisite: install failure")
	})

	t.Run("failure - test run failure", func(t *testing.T) {
		config := golangBuildOptions{
			RunTests: true,
		}
		utils := newGolangBuildTestsUtils()
		utils.ShouldFailOnCommand = map[string]error{"gotestsum --junitfile": fmt.Errorf("test failure")}
		telemetryData := telemetry.CustomData{}

		err := runGolangBuild(&config, &telemetryData, utils, &cpe)
		assert.EqualError(t, err, "running tests failed - junit result missing: test failure")
	})

	t.Run("failure - test failure", func(t *testing.T) {
		config := golangBuildOptions{
			RunTests: true,
		}
		utils := newGolangBuildTestsUtils()
		utils.ShouldFailOnCommand = map[string]error{"gotestsum --junitfile": fmt.Errorf("test failure")}
		utils.AddFile("TEST-go.xml", []byte("some content"))
		utils.AddFile(coverageFile, []byte("some content"))
		telemetryData := telemetry.CustomData{}

		err := runGolangBuild(&config, &telemetryData, utils, &cpe)
		assert.EqualError(t, err, "some tests failed")
	})

	t.Run("failure - prepareLdflags", func(t *testing.T) {
		config := golangBuildOptions{
			RunTests:            true,
			LdflagsTemplate:     "{{.CPE.test",
			TargetArchitectures: []string{"linux,amd64"},
		}
		utils := newGolangBuildTestsUtils()
		telemetryData := telemetry.CustomData{}

		err := runGolangBuild(&config, &telemetryData, utils, &cpe)
		assert.Contains(t, fmt.Sprint(err), "failed to parse ldflagsTemplate")
	})

	t.Run("failure - build failure", func(t *testing.T) {
		config := golangBuildOptions{
			RunIntegrationTests: true,
			TargetArchitectures: []string{"linux,amd64"},
		}
		utils := newGolangBuildTestsUtils()
		utils.ShouldFailOnCommand = map[string]error{"go build": fmt.Errorf("build failure")}
		telemetryData := telemetry.CustomData{}

		err := runGolangBuild(&config, &telemetryData, utils, &cpe)
		assert.EqualError(t, err, "failed to run build for linux.amd64: build failure")
	})

	t.Run("failure - publish - no target repository defined", func(t *testing.T) {
		config := golangBuildOptions{
			TargetArchitectures: []string{"linux,amd64"},
			Output:              "testBin",
			Publish:             true,
		}
		utils := newGolangBuildTestsUtils()
		telemetryData := telemetry.CustomData{}

		err := runGolangBuild(&config, &telemetryData, utils, &cpe)
		assert.EqualError(t, err, "there's no target repository for binary publishing configured")
	})

	t.Run("failure - publish - no go.mod file found", func(t *testing.T) {
		config := golangBuildOptions{
			TargetArchitectures:      []string{"linux,amd64"},
			Output:                   "testBin",
			Publish:                  true,
			TargetRepositoryURL:      "https://my.target.repository.local",
			TargetRepositoryUser:     "user",
			TargetRepositoryPassword: "password",
			ArtifactVersion:          "1.0.0",
		}
		utils := newGolangBuildTestsUtils()
		telemetryData := telemetry.CustomData{}

		err := runGolangBuild(&config, &telemetryData, utils, &cpe)
		assert.EqualError(t, err, "go.mod file not found")
	})

	t.Run("failure - publish - go.mod file without module path", func(t *testing.T) {
		config := golangBuildOptions{
			TargetArchitectures:      []string{"linux,amd64"},
			Output:                   "testBin",
			Publish:                  true,
			TargetRepositoryURL:      "https://my.target.repository.local",
			TargetRepositoryUser:     "user",
			TargetRepositoryPassword: "password",
			ArtifactVersion:          "1.0.0",
		}
		utils := newGolangBuildTestsUtils()
		utils.FilesMock.AddFile("go.mod", []byte(""))
		telemetryData := telemetry.CustomData{}

		err := runGolangBuild(&config, &telemetryData, utils, &cpe)
		assert.EqualError(t, err, "go.mod doesn't declare a module path")
	})

	t.Run("failure - publish - no artifactVersion set", func(t *testing.T) {
		config := golangBuildOptions{
			TargetArchitectures:      []string{"linux,amd64"},
			Output:                   "testBin",
			Publish:                  true,
			TargetRepositoryURL:      "https://my.target.repository.local",
			TargetRepositoryUser:     "user",
			TargetRepositoryPassword: "password",
		}
		utils := newGolangBuildTestsUtils()
		utils.FilesMock.AddFile("go.mod", []byte("module example.com/my/module"))
		telemetryData := telemetry.CustomData{}

		err := runGolangBuild(&config, &telemetryData, utils, &cpe)
		assert.EqualError(t, err, "no build descriptor available, supported: [VERSION version.txt go.mod]")
	})

	t.Run("failure - publish - received unexpected status code", func(t *testing.T) {
		config := golangBuildOptions{
			TargetArchitectures:      []string{"linux,amd64"},
			Output:                   "testBin",
			Publish:                  true,
			TargetRepositoryURL:      "https://my.target.repository.local",
			TargetRepositoryUser:     "user",
			TargetRepositoryPassword: "password",
			ArtifactVersion:          "1.0.0",
		}
		utils := newGolangBuildTestsUtils()
		utils.returnFileUploadStatus = 500
		utils.FilesMock.AddFile("go.mod", []byte("module example.com/my/module"))
		telemetryData := telemetry.CustomData{}

		err := runGolangBuild(&config, &telemetryData, utils, &cpe)
		assert.EqualError(t, err, "couldn't upload artifact, received status code 500")
	})

	t.Run("failure - create BOM", func(t *testing.T) {
		config := golangBuildOptions{
			CreateBOM:           true,
			TargetArchitectures: []string{"linux,amd64"},
		}
		utils := newGolangBuildTestsUtils()
		utils.ShouldFailOnCommand = map[string]error{"cyclonedx-gomod mod -licenses -test -output bom.xml": fmt.Errorf("BOM creation failure")}
		telemetryData := telemetry.CustomData{}

		err := runGolangBuild(&config, &telemetryData, utils, &cpe)
		assert.EqualError(t, err, "BOM creation failed: BOM creation failure")
	})
}

func TestRunGolangTests(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		config := golangBuildOptions{}
		utils := newGolangBuildTestsUtils()
		utils.AddFile("TEST-go.xml", []byte("some content"))
		utils.AddFile(coverageFile, []byte("some content"))

		success, err := runGolangTests(&config, utils)
		assert.NoError(t, err)
		assert.True(t, success)
		assert.Equal(t, "gotestsum", utils.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{"--junitfile", "TEST-go.xml", "--", fmt.Sprintf("-coverprofile=%v", coverageFile), "./..."}, utils.ExecMockRunner.Calls[0].Params)
	})

	t.Run("success - failed tests", func(t *testing.T) {
		t.Parallel()
		config := golangBuildOptions{}
		utils := newGolangBuildTestsUtils()
		utils.AddFile("TEST-go.xml", []byte("some content"))
		utils.AddFile(coverageFile, []byte("some content"))
		utils.ExecMockRunner.ShouldFailOnCommand = map[string]error{"gotestsum": fmt.Errorf("execution error")}

		success, err := runGolangTests(&config, utils)
		assert.NoError(t, err)
		assert.False(t, success)
	})

	t.Run("error - run failed, no junit", func(t *testing.T) {
		t.Parallel()
		config := golangBuildOptions{}
		utils := newGolangBuildTestsUtils()
		utils.ExecMockRunner.ShouldFailOnCommand = map[string]error{"gotestsum": fmt.Errorf("execution error")}

		_, err := runGolangTests(&config, utils)
		assert.EqualError(t, err, "running tests failed - junit result missing: execution error")
	})

	t.Run("error - run failed, no coverage", func(t *testing.T) {
		t.Parallel()
		config := golangBuildOptions{}
		utils := newGolangBuildTestsUtils()
		utils.ExecMockRunner.ShouldFailOnCommand = map[string]error{"gotestsum": fmt.Errorf("execution error")}
		utils.AddFile("TEST-go.xml", []byte("some content"))

		_, err := runGolangTests(&config, utils)
		assert.EqualError(t, err, "running tests failed - coverage output missing: execution error")
	})
}

func TestRunGolangIntegrationTests(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		config := golangBuildOptions{}
		utils := newGolangBuildTestsUtils()
		utils.AddFile("TEST-integration.xml", []byte("some content"))

		success, err := runGolangIntegrationTests(&config, utils)
		assert.NoError(t, err)
		assert.True(t, success)
		assert.Equal(t, "gotestsum", utils.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{"--junitfile", "TEST-integration.xml", "--", "-tags=integration", "./..."}, utils.ExecMockRunner.Calls[0].Params)
	})

	t.Run("success - failed tests", func(t *testing.T) {
		t.Parallel()
		config := golangBuildOptions{}
		utils := newGolangBuildTestsUtils()
		utils.AddFile("TEST-integration.xml", []byte("some content"))
		utils.ExecMockRunner.ShouldFailOnCommand = map[string]error{"gotestsum": fmt.Errorf("execution error")}

		success, err := runGolangIntegrationTests(&config, utils)
		assert.NoError(t, err)
		assert.False(t, success)
	})

	t.Run("error - run failed", func(t *testing.T) {
		t.Parallel()
		config := golangBuildOptions{}
		utils := newGolangBuildTestsUtils()
		utils.ExecMockRunner.ShouldFailOnCommand = map[string]error{"gotestsum": fmt.Errorf("execution error")}

		_, err := runGolangIntegrationTests(&config, utils)
		assert.EqualError(t, err, "running tests failed: execution error")
	})
}

func TestReportGolangTestCoverage(t *testing.T) {
	t.Parallel()

	t.Run("success - cobertura", func(t *testing.T) {
		t.Parallel()
		config := golangBuildOptions{CoverageFormat: "cobertura"}
		utils := newGolangBuildTestsUtils()
		utils.AddFile(coverageFile, []byte("some content"))

		err := reportGolangTestCoverage(&config, utils)
		assert.NoError(t, err)
		assert.Equal(t, "go", utils.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{"install", "github.com/boumenot/gocover-cobertura@latest"}, utils.ExecMockRunner.Calls[0].Params)
		assert.Equal(t, "gocover-cobertura", utils.ExecMockRunner.Calls[1].Exec)
		exists, err := utils.FileExists("cobertura-coverage.xml")
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("success - cobertura exclude generated", func(t *testing.T) {
		t.Parallel()
		config := golangBuildOptions{CoverageFormat: "cobertura", ExcludeGeneratedFromCoverage: true}
		utils := newGolangBuildTestsUtils()
		utils.AddFile(coverageFile, []byte("some content"))

		err := reportGolangTestCoverage(&config, utils)
		assert.NoError(t, err)
		assert.Equal(t, "gocover-cobertura", utils.ExecMockRunner.Calls[1].Exec)
		assert.Equal(t, []string{"-ignore-gen-files"}, utils.ExecMockRunner.Calls[1].Params)
	})

	t.Run("error - cobertura installation", func(t *testing.T) {
		t.Parallel()
		config := golangBuildOptions{CoverageFormat: "cobertura", ExcludeGeneratedFromCoverage: true}
		utils := newGolangBuildTestsUtils()
		utils.ExecMockRunner.ShouldFailOnCommand = map[string]error{"go install github.com/boumenot/gocover-cobertura": fmt.Errorf("install error")}

		err := reportGolangTestCoverage(&config, utils)
		assert.EqualError(t, err, "failed to install pre-requisite: install error")
	})

	t.Run("error - cobertura missing coverage file", func(t *testing.T) {
		t.Parallel()
		config := golangBuildOptions{CoverageFormat: "cobertura", ExcludeGeneratedFromCoverage: true}
		utils := newGolangBuildTestsUtils()

		err := reportGolangTestCoverage(&config, utils)
		assert.Contains(t, fmt.Sprint(err), "failed to read coverage file")
	})

	t.Run("error - cobertura coversion", func(t *testing.T) {
		t.Parallel()
		config := golangBuildOptions{CoverageFormat: "cobertura", ExcludeGeneratedFromCoverage: true}
		utils := newGolangBuildTestsUtils()
		utils.AddFile(coverageFile, []byte("some content"))
		utils.ExecMockRunner.ShouldFailOnCommand = map[string]error{"gocover-cobertura -ignore-gen-files": fmt.Errorf("execution error")}

		err := reportGolangTestCoverage(&config, utils)
		assert.EqualError(t, err, "failed to convert coverage data to cobertura format: execution error")
	})

	t.Run("error - writing cobertura file", func(t *testing.T) {
		t.Parallel()
		config := golangBuildOptions{CoverageFormat: "cobertura", ExcludeGeneratedFromCoverage: true}
		utils := newGolangBuildTestsUtils()
		utils.AddFile(coverageFile, []byte("some content"))
		utils.FileWriteError = fmt.Errorf("write failure")

		err := reportGolangTestCoverage(&config, utils)
		assert.EqualError(t, err, "failed to create cobertura coverage file: write failure")
	})

	t.Run("success - html", func(t *testing.T) {
		t.Parallel()
		config := golangBuildOptions{}
		utils := newGolangBuildTestsUtils()

		err := reportGolangTestCoverage(&config, utils)
		assert.NoError(t, err)
		assert.Equal(t, "go", utils.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{"tool", "cover", "-html", coverageFile, "-o", "coverage.html"}, utils.ExecMockRunner.Calls[0].Params)
	})

	t.Run("error - html", func(t *testing.T) {
		t.Parallel()
		config := golangBuildOptions{}
		utils := newGolangBuildTestsUtils()
		utils.ExecMockRunner.ShouldFailOnCommand = map[string]error{"go tool cover -html cover.out -o coverage.html": fmt.Errorf("execution error")}
		utils.AddFile(coverageFile, []byte("some content"))

		err := reportGolangTestCoverage(&config, utils)
		assert.EqualError(t, err, "failed to create html coverage file: execution error")
	})
}

func TestPrepareLdflags(t *testing.T) {
	t.Parallel()
	dir, err := ioutil.TempDir("", "")
	defer os.RemoveAll(dir) // clean up
	assert.NoError(t, err, "Error when creating temp dir")

	err = os.Mkdir(filepath.Join(dir, "commonPipelineEnvironment"), 0777)
	assert.NoError(t, err, "Error when creating folder structure")

	err = ioutil.WriteFile(filepath.Join(dir, "commonPipelineEnvironment", "artifactVersion"), []byte("1.2.3"), 0666)
	assert.NoError(t, err, "Error when creating cpe file")

	t.Run("success - default", func(t *testing.T) {
		config := golangBuildOptions{LdflagsTemplate: "-X version={{ .CPE.artifactVersion }}"}
		utils := newGolangBuildTestsUtils()
		result, err := prepareLdflags(&config, utils, dir)
		assert.NoError(t, err)
		assert.Equal(t, "-X version=1.2.3", result)
	})

	t.Run("error - template parsing", func(t *testing.T) {
		config := golangBuildOptions{LdflagsTemplate: "-X version={{ .CPE.artifactVersion "}
		utils := newGolangBuildTestsUtils()
		_, err := prepareLdflags(&config, utils, dir)
		assert.Contains(t, fmt.Sprint(err), "failed to parse ldflagsTemplate")
	})
}

func TestRunGolangBuildPerArchitecture(t *testing.T) {
	t.Parallel()

	t.Run("success - default", func(t *testing.T) {
		t.Parallel()
		config := golangBuildOptions{}
		utils := newGolangBuildTestsUtils()
		ldflags := ""
		architecture, _ := multiarch.ParsePlatformString("linux,amd64")

		binaryName, err := runGolangBuildPerArchitecture(&config, utils, ldflags, architecture)
		assert.NoError(t, err)
		assert.Greater(t, len(utils.Env), 3)
		assert.Contains(t, utils.Env, "CGO_ENABLED=0")
		assert.Contains(t, utils.Env, "GOOS=linux")
		assert.Contains(t, utils.Env, "GOARCH=amd64")
		assert.Equal(t, utils.Calls[0].Exec, "go")
		assert.Equal(t, utils.Calls[0].Params[0], "build")
		assert.Empty(t, binaryName)
	})

	t.Run("success - custom params", func(t *testing.T) {
		t.Parallel()
		config := golangBuildOptions{BuildFlags: []string{"--flag1", "val1", "--flag2", "val2"}, Output: "testBin", Packages: []string{"./test/.."}}
		utils := newGolangBuildTestsUtils()
		ldflags := "-X test=test"
		architecture, _ := multiarch.ParsePlatformString("linux,amd64")

		binaryName, err := runGolangBuildPerArchitecture(&config, utils, ldflags, architecture)
		assert.NoError(t, err)
		assert.Contains(t, utils.Calls[0].Params, "-o")
		assert.Contains(t, utils.Calls[0].Params, "testBin-linux.amd64")
		assert.Contains(t, utils.Calls[0].Params, "./test/..")
		assert.Contains(t, utils.Calls[0].Params, "-ldflags")
		assert.Contains(t, utils.Calls[0].Params, "-X test=test")
		assert.Equal(t, "testBin-linux.amd64", binaryName)
	})

	t.Run("success - windows", func(t *testing.T) {
		t.Parallel()
		config := golangBuildOptions{Output: "testBin"}
		utils := newGolangBuildTestsUtils()
		ldflags := ""
		architecture, _ := multiarch.ParsePlatformString("windows,amd64")

		binaryName, err := runGolangBuildPerArchitecture(&config, utils, ldflags, architecture)
		assert.NoError(t, err)
		assert.Contains(t, utils.Calls[0].Params, "-o")
		assert.Contains(t, utils.Calls[0].Params, "testBin-windows.amd64.exe")
		assert.Equal(t, "testBin-windows.amd64.exe", binaryName)
	})

	t.Run("execution error", func(t *testing.T) {
		t.Parallel()
		config := golangBuildOptions{}
		utils := newGolangBuildTestsUtils()
		utils.ShouldFailOnCommand = map[string]error{"go build": fmt.Errorf("execution error")}
		ldflags := ""
		architecture, _ := multiarch.ParsePlatformString("linux,amd64")

		_, err := runGolangBuildPerArchitecture(&config, utils, ldflags, architecture)
		assert.EqualError(t, err, "failed to run build for linux.amd64: execution error")
	})

}

func TestPrepareGolangEnvironment(t *testing.T) {
	modTestFile := `
module private.example.com/m

require (
        example.com/public/module v1.0.0
        private1.example.com/private/repo v0.1.0
        private2.example.com/another/repo v0.2.0
)

go 1.17`

	type expectations struct {
		envVars          []string
		commandsExecuted [][]string
	}
	tests := []struct {
		name           string
		modFileContent string
		globPattern    string
		gitToken       string
		expect         expectations
	}{
		{
			name:           "success - does nothing if privateModules is not set",
			modFileContent: modTestFile,
			globPattern:    "",
			gitToken:       "secret",
			expect:         expectations{},
		},
		{
			name:           "success - goprivate is set and authentication properly configured",
			modFileContent: modTestFile,
			globPattern:    "*.example.com",
			gitToken:       "secret",
			expect: expectations{
				envVars: []string{"GOPRIVATE=*.example.com"},
				commandsExecuted: [][]string{
					[]string{"git", "config", "--global", "url.https://secret@private1.example.com/private/repo.git.insteadOf", "https://private1.example.com/private/repo.git"},
					[]string{"git", "config", "--global", "url.https://secret@private2.example.com/another/repo.git.insteadOf", "https://private2.example.com/another/repo.git"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils := newGolangBuildTestsUtils()

			goModFile, _ := modfile.Parse("go.mod", []byte(tt.modFileContent), nil)

			config := golangBuildOptions{}
			config.PrivateModules = tt.globPattern
			config.PrivateModulesGitToken = tt.gitToken

			err := prepareGolangEnvironment(&config, goModFile, utils)

			if assert.NoError(t, err) {
				assert.Subset(t, os.Environ(), tt.expect.envVars)
				assert.Equal(t, len(tt.expect.commandsExecuted), len(utils.Calls))

				for i, expectedCommand := range tt.expect.commandsExecuted {
					assert.Equal(t, expectedCommand[0], utils.Calls[i].Exec)
					assert.Equal(t, expectedCommand[1:], utils.Calls[i].Params)
				}
			}
		})
	}
}

func TestLookupGolangPrivateModulesRepositories(t *testing.T) {
	t.Parallel()

	modTestFile := `
module private.example.com/m

require (
	example.com/public/module v1.0.0
	private1.example.com/private/repo v0.1.0
	private2.example.com/another/repo v0.2.0
)

go 1.17`

	type expectations struct {
		repos        []string
		errorMessage string
	}
	tests := []struct {
		name           string
		modFileContent string
		globPattern    string
		expect         expectations
	}{
		{
			name:           "Does nothing if glob pattern is empty",
			modFileContent: modTestFile,
			expect: expectations{
				repos: []string{},
			},
		},
		{
			name:           "Does nothing if there is no go.mod file",
			globPattern:    "private.example.com",
			modFileContent: "",
			expect: expectations{
				repos: []string{},
			},
		},
		{
			name:           "Detects all private repos using a glob pattern",
			modFileContent: modTestFile,
			globPattern:    "*.example.com",
			expect: expectations{
				repos: []string{"https://private1.example.com/private/repo.git", "https://private2.example.com/another/repo.git"},
			},
		},
		{
			name:           "Detects all private repos",
			modFileContent: modTestFile,
			globPattern:    "private1.example.com,private2.example.com",
			expect: expectations{
				repos: []string{"https://private1.example.com/private/repo.git", "https://private2.example.com/another/repo.git"},
			},
		},
		{
			name:           "Detects a dedicated repo",
			modFileContent: modTestFile,
			globPattern:    "private2.example.com",
			expect: expectations{
				repos: []string{"https://private2.example.com/another/repo.git"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils := newGolangBuildTestsUtils()

			goModFile, _ := modfile.Parse("", []byte(tt.modFileContent), nil)

			repos, err := lookupGolangPrivateModulesRepositories(goModFile, tt.globPattern, utils)

			if tt.expect.errorMessage == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expect.repos, repos)
			} else {
				assert.EqualError(t, err, tt.expect.errorMessage)
			}
		})
	}
}
