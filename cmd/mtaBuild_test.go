package cmd

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

type mtaBuildTestUtilsBundle struct {
	*mock.ExecMockRunner
	*mock.FilesMock
	projectSettingsFile            string
	globalSettingsFile             string
	registryUsedInSetNpmRegistries string
	openReturns                    string
	sendRequestReturns             func() (*http.Response, error)
}

func (m *mtaBuildTestUtilsBundle) SetNpmRegistries(defaultNpmRegistry string) error {
	m.registryUsedInSetNpmRegistries = defaultNpmRegistry
	return nil
}

func (m *mtaBuildTestUtilsBundle) InstallAllDependencies(defaultNpmRegistry string) error {
	return errors.New("Test should not install dependencies.") // TODO implement test
}

func (m *mtaBuildTestUtilsBundle) DownloadAndCopySettingsFiles(globalSettingsFile string, projectSettingsFile string) error {
	m.projectSettingsFile = projectSettingsFile
	m.globalSettingsFile = globalSettingsFile
	return nil
}

func (m *mtaBuildTestUtilsBundle) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	return errors.New("Test should not download files.")
}

func (m *mtaBuildTestUtilsBundle) Open(name string) (io.ReadWriteCloser, error) {
	if m.openReturns != "" {
		return NewMockReadCloser(m.openReturns), nil
	}
	return nil, errors.New("Test should not open files.")
}

// MockReadCloser is a struct that implements io.ReadCloser
type MockReadWriteCloser struct {
	io.Reader
	io.Writer
}

// Close is a no-op method to satisfy the io.Closer interface
func (m *MockReadWriteCloser) Close() error {
	return nil
}

// NewMockReadCloser returns a new MockReadCloser with the given data
func NewMockReadCloser(data string) io.ReadWriteCloser {
	return &MockReadWriteCloser{Reader: bytes.NewBufferString(data)}
}

func (m *mtaBuildTestUtilsBundle) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	if m.sendRequestReturns != nil {
		return m.sendRequestReturns()
	}
	return nil, errors.New("Test should not send requests.")
}

func newMtaBuildTestUtilsBundle() *mtaBuildTestUtilsBundle {
	utilsBundle := mtaBuildTestUtilsBundle{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return &utilsBundle
}

func TestMtaBuild(t *testing.T) {
	cpe := mtaBuildCommonPipelineEnvironment{}
	SetConfigOptions(ConfigCommandOptions{
		OpenFile: config.OpenPiperFile,
	})

	t.Run("Application name not set", func(t *testing.T) {
		utilsMock := newMtaBuildTestUtilsBundle()
		options := mtaBuildOptions{}

		err := runMtaBuild(options, &cpe, utilsMock)

		assert.NotNil(t, err)
		assert.Equal(t, "'mta.yaml' not found in project sources and 'applicationName' not provided as parameter - cannot generate 'mta.yaml' file", err.Error())
	})

	t.Run("Provide default npm registry", func(t *testing.T) {
		utilsMock := newMtaBuildTestUtilsBundle()
		options := mtaBuildOptions{
			ApplicationName:    "myApp",
			Platform:           "CF",
			DefaultNpmRegistry: "https://example.org/npm",
			MtarName:           "myName",
			Source:             "./",
			Target:             "./",
		}

		utilsMock.AddFile("package.json", []byte(`{"name": "myName", "version": "1.2.3"}`))

		err := runMtaBuild(options, &cpe, utilsMock)

		assert.Nil(t, err)

		assert.Equal(t, "https://example.org/npm", utilsMock.registryUsedInSetNpmRegistries)
	})

	t.Run("Package json does not exist", func(t *testing.T) {
		utilsMock := newMtaBuildTestUtilsBundle()

		options := mtaBuildOptions{ApplicationName: "myApp"}

		err := runMtaBuild(options, &cpe, utilsMock)

		assert.NotNil(t, err)

		assert.Equal(t, "package.json file does not exist", err.Error())
	})

	t.Run("Write yaml file", func(t *testing.T) {
		utilsMock := newMtaBuildTestUtilsBundle()

		options := mtaBuildOptions{
			ApplicationName:    "myApp",
			Platform:           "CF",
			MtarName:           "myName",
			Source:             "./",
			Target:             "./",
			EnableSetTimestamp: true,
		}

		utilsMock.AddFile("package.json", []byte(`{"name": "myName", "version": "1.2.3"}`))

		err := runMtaBuild(options, &cpe, utilsMock)

		assert.Nil(t, err)

		type MtaResult struct {
			Version    string
			ID         string `yaml:"ID,omitempty"`
			Parameters map[string]string
			Modules    []struct {
				Name       string
				Type       string
				Parameters map[string]interface{}
			}
		}

		assert.True(t, utilsMock.HasWrittenFile("mta.yaml"))

		var result MtaResult
		mtaContent, _ := utilsMock.FileRead("mta.yaml")
		err = yaml.Unmarshal(mtaContent, &result)
		assert.NoError(t, err)
		assert.Equal(t, "myName", result.ID)
		assert.Equal(t, "1.2.3", result.Version)
		assert.Equal(t, "myApp", result.Modules[0].Name)
		assert.Regexp(t, "^1\\.2\\.3-[\\d]{14}$", result.Modules[0].Parameters["version"])
		assert.Equal(t, "myApp", result.Modules[0].Parameters["name"])
	})

	t.Run("Dont write mta yaml file when already present no timestamp placeholder", func(t *testing.T) {
		utilsMock := newMtaBuildTestUtilsBundle()

		options := mtaBuildOptions{ApplicationName: "myApp"}

		utilsMock.AddFile("package.json", []byte(`{"name": "myName", "version": "1.2.3"}`))
		utilsMock.AddFile("mta.yaml", []byte("already there"))

		_ = runMtaBuild(options, &cpe, utilsMock)

		assert.False(t, utilsMock.HasWrittenFile("mta.yaml"))
	})

	t.Run("Write mta yaml file when already present with timestamp placeholder", func(t *testing.T) {
		utilsMock := newMtaBuildTestUtilsBundle()

		options := mtaBuildOptions{
			ApplicationName:    "myApp",
			EnableSetTimestamp: true,
		}

		utilsMock.AddFile("package.json", []byte(`{"name": "myName", "version": "1.2.3"}`))
		utilsMock.AddFile("mta.yaml", []byte("already there with-${timestamp}"))

		_ = runMtaBuild(options, &cpe, utilsMock)

		assert.True(t, utilsMock.HasWrittenFile("mta.yaml"))
	})

	t.Run("Mta build mbt toolset", func(t *testing.T) {
		utilsMock := newMtaBuildTestUtilsBundle()

		cpe.mtarFilePath = ""

		options := mtaBuildOptions{ApplicationName: "myApp", Platform: "CF", MtarName: "myName.mtar", Source: "./", Target: "./"}

		utilsMock.AddFile("package.json", []byte(`{"name": "myName", "version": "1.2.3"}`))

		err := runMtaBuild(options, &cpe, utilsMock)

		assert.Nil(t, err)

		if assert.Len(t, utilsMock.Calls, 1) {
			assert.Equal(t, "mbt", utilsMock.Calls[0].Exec)
			assert.Equal(t, []string{"build", "--mtar", "myName.mtar", "--platform", "CF", "--source", filepath.FromSlash("./"), "--target", filepath.FromSlash(_ignoreError(os.Getwd()))}, utilsMock.Calls[0].Params)
		}
		assert.Equal(t, "myName.mtar", cpe.mtarFilePath)
	})

	t.Run("Source and target related tests", func(t *testing.T) {
		t.Run("Mta build mbt toolset with custom source and target paths", func(t *testing.T) {
			utilsMock := newMtaBuildTestUtilsBundle()

			cpe.mtarFilePath = ""

			options := mtaBuildOptions{
				ApplicationName: "myApp",
				Platform:        "CF",
				MtarName:        "myName.mtar",
				Source:          "mySourcePath/",
				Target:          "myTargetPath/",
			}

			utilsMock.AddFile("package.json", []byte(`{"name": "myName", "version": "1.2.3"}`))

			err := runMtaBuild(options, &cpe, utilsMock)

			assert.Nil(t, err)

			if assert.Len(t, utilsMock.Calls, 1) {
				assert.Equal(t, "mbt", utilsMock.Calls[0].Exec)
				assert.Equal(t, []string{
					"build", "--mtar", "myName.mtar", "--platform", "CF",
					"--source", filepath.FromSlash("mySourcePath/"),
					"--target", filepath.Join(_ignoreError(os.Getwd()), filepath.FromSlash("mySourcePath/myTargetPath/")),
				},
					utilsMock.Calls[0].Params)
			}
			assert.Equal(t, "mySourcePath/myTargetPath/myName.mtar", cpe.mtarFilePath)
			assert.Equal(t, "mySourcePath/mta.yaml", cpe.custom.mtaBuildToolDesc)
		})
	})

	t.Run("M2Path related tests", func(t *testing.T) {
		t.Run("Mta build mbt toolset with m2Path", func(t *testing.T) {
			utilsMock := newMtaBuildTestUtilsBundle()
			utilsMock.CurrentDir = "root_folder/workspace"
			cpe.mtarFilePath = ""

			options := mtaBuildOptions{
				ApplicationName: "myApp",
				Platform:        "CF",
				MtarName:        "myName.mtar",
				Source:          "./",
				Target:          "./",
				M2Path:          ".pipeline/local_repo",
			}

			utilsMock.AddFile("mta.yaml", []byte(`ID: "myNameFromMtar"`))

			err := runMtaBuild(options, &cpe, utilsMock)

			assert.Nil(t, err)
			assert.Contains(t, utilsMock.Env, filepath.FromSlash("MAVEN_OPTS=-Dmaven.repo.local=/root_folder/workspace/.pipeline/local_repo"))
		})
	})

	t.Run("Settings file releatd tests", func(t *testing.T) {
		t.Run("Copy global settings file", func(t *testing.T) {
			utilsMock := newMtaBuildTestUtilsBundle()
			utilsMock.AddFile("mta.yaml", []byte(`ID: "myNameFromMtar"`))

			options := mtaBuildOptions{
				ApplicationName:    "myApp",
				GlobalSettingsFile: "/opt/maven/settings.xml",
				Platform:           "CF",
				MtarName:           "myName",
				Source:             "./",
				Target:             "./",
			}

			err := runMtaBuild(options, &cpe, utilsMock)

			assert.Nil(t, err)

			assert.Equal(t, "/opt/maven/settings.xml", utilsMock.globalSettingsFile)
			assert.Equal(t, "", utilsMock.projectSettingsFile)
		})

		t.Run("Copy project settings file", func(t *testing.T) {
			utilsMock := newMtaBuildTestUtilsBundle()
			utilsMock.AddFile("mta.yaml", []byte(`ID: "myNameFromMtar"`))

			options := mtaBuildOptions{ApplicationName: "myApp", ProjectSettingsFile: "/my/project/settings.xml", Platform: "CF", MtarName: "myName", Source: "./", Target: "./"}

			err := runMtaBuild(options, &cpe, utilsMock)

			assert.Nil(t, err)

			assert.Equal(t, "/my/project/settings.xml", utilsMock.projectSettingsFile)
			assert.Equal(t, "", utilsMock.globalSettingsFile)
		})
	})

	t.Run("publish related tests", func(t *testing.T) {
		t.Run("error when no repository url", func(t *testing.T) {
			utilsMock := newMtaBuildTestUtilsBundle()
			utilsMock.AddFile("mta.yaml", []byte(`ID: "myNameFromMtar"`))

			options := mtaBuildOptions{
				ApplicationName:    "myApp",
				GlobalSettingsFile: "/opt/maven/settings.xml",
				Platform:           "CF",
				MtarName:           "myName",
				Source:             "./",
				Target:             "./",
				Publish:            true,
			}

			err := runMtaBuild(options, &cpe, utilsMock)

			assert.Equal(t, "mtaDeploymentRepositoryUser, mtaDeploymentRepositoryPassword and mtaDeploymentRepositoryURL not found, must be present", err.Error())
		})

		t.Run("error when no mtar group", func(t *testing.T) {
			utilsMock := newMtaBuildTestUtilsBundle()
			utilsMock.AddFile("mta.yaml", []byte(`ID: "myNameFromMtar"`))

			options := mtaBuildOptions{
				ApplicationName:                 "myApp",
				GlobalSettingsFile:              "/opt/maven/settings.xml",
				Platform:                        "CF",
				MtarName:                        "myName",
				Source:                          "./",
				Target:                          "./",
				Publish:                         true,
				MtaDeploymentRepositoryURL:      "dummy",
				MtaDeploymentRepositoryPassword: "dummy",
				MtaDeploymentRepositoryUser:     "dummy",
			}

			err := runMtaBuild(options, &cpe, utilsMock)

			assert.Equal(t, "mtarGroup, version not found and must be present", err.Error())
		})

		t.Run("successful publish", func(t *testing.T) {
			utilsMock := newMtaBuildTestUtilsBundle()
			utilsMock.sendRequestReturns = func() (*http.Response, error) {
				return &http.Response{StatusCode: 200}, nil
			}
			utilsMock.AddFile("mta.yaml", []byte(`ID: "myNameFromMtar"`))
			utilsMock.openReturns = `{"version":"1.2.3"}`
			options := mtaBuildOptions{
				ApplicationName:                 "myApp",
				GlobalSettingsFile:              "/opt/maven/settings.xml",
				Platform:                        "CF",
				MtarName:                        "test",
				Source:                          "./",
				Target:                          "./",
				Publish:                         true,
				MtaDeploymentRepositoryURL:      "dummy",
				MtaDeploymentRepositoryPassword: "dummy",
				MtaDeploymentRepositoryUser:     "dummy",
				MtarGroup:                       "dummy",
				Version:                         "dummy",
			}
			err := runMtaBuild(options, &cpe, utilsMock)
			assert.Nil(t, err)
		})

		t.Run("succesful build artifact", func(t *testing.T) {
			utilsMock := newMtaBuildTestUtilsBundle()
			utilsMock.AddFile("mta.yaml", []byte(`ID: "myNameFromMtar"`))
			utilsMock.openReturns = `{"version":"1.2.3"}`
			utilsMock.sendRequestReturns = func() (*http.Response, error) {
				return &http.Response{StatusCode: 200}, nil
			}
			options := mtaBuildOptions{
				ApplicationName:                 "myApp",
				GlobalSettingsFile:              "/opt/maven/settings.xml",
				Platform:                        "CF",
				MtarName:                        "test",
				Source:                          "./",
				Target:                          "./",
				Publish:                         true,
				MtaDeploymentRepositoryURL:      "dummy",
				MtaDeploymentRepositoryPassword: "dummy",
				MtaDeploymentRepositoryUser:     "dummy",
				MtarGroup:                       "dummy",
				Version:                         "dummy",
				CreateBuildArtifactsMetadata:    true,
			}
			err := runMtaBuild(options, &cpe, utilsMock)
			assert.Nil(t, err)
			assert.Equal(t, cpe.custom.mtaBuildArtifacts, `{"Coordinates":[{"groupId":"dummy","artifactId":"test","version":"dummy","packaging":"mtar","buildPath":"./","url":"dummydummy/test/dummy/test-dummy.mtar","purl":""}]}`)
		})
	})
}

func TestMtaBuildSourceDir(t *testing.T) {
	cpe := mtaBuildCommonPipelineEnvironment{}
	t.Run("getSourcePath", func(t *testing.T) {
		t.Parallel()

		t.Run("getPath dir unset", func(t *testing.T) {
			options := mtaBuildOptions{Source: "", Target: ""}
			assert.Equal(t, filepath.FromSlash("./"), getSourcePath(options))
			assert.Equal(t, filepath.FromSlash("./"), getTargetPath(options))
		})
		t.Run("getPath source set", func(t *testing.T) {
			options := mtaBuildOptions{Source: "spath", Target: ""}
			assert.Equal(t, filepath.FromSlash("spath"), getSourcePath(options))
			assert.Equal(t, filepath.FromSlash("./"), getTargetPath(options))
		})
		t.Run("getPath target set", func(t *testing.T) {
			options := mtaBuildOptions{Source: "", Target: "tpath"}
			assert.Equal(t, filepath.FromSlash("./"), getSourcePath(options))
			assert.Equal(t, filepath.FromSlash("tpath"), getTargetPath(options))
		})
		t.Run("getPath dir set to relative path", func(t *testing.T) {
			options := mtaBuildOptions{Source: "spath", Target: "tpath"}
			assert.Equal(t, filepath.FromSlash("spath"), getSourcePath(options))
			assert.Equal(t, filepath.FromSlash("tpath"), getTargetPath(options))
		})
		t.Run("getPath dir ends with seperator", func(t *testing.T) {
			options := mtaBuildOptions{Source: "spath/", Target: "tpath/"}
			assert.Equal(t, filepath.FromSlash("spath/"), getSourcePath(options))
			assert.Equal(t, filepath.FromSlash("tpath/"), getTargetPath(options))
		})
		t.Run("getPath dir set to absolute path", func(t *testing.T) {
			sourcePath := filepath.Join(_ignoreError(os.Getwd()), "spath")
			targetPath := filepath.Join(_ignoreError(os.Getwd()), "tpath")
			options := mtaBuildOptions{Source: sourcePath, Target: targetPath}
			assert.Equal(t, filepath.FromSlash(sourcePath), getSourcePath(options))
			assert.Equal(t, filepath.FromSlash(targetPath), getTargetPath(options))
		})
	})

	t.Run("find build tool descriptor from configuration", func(t *testing.T) {
		t.Parallel()
		t.Run("default mta.yaml", func(t *testing.T) {
			utilsMock := newMtaBuildTestUtilsBundle()

			utilsMock.AddFile("mta.yaml", []byte("already there"))

			_ = runMtaBuild(mtaBuildOptions{ApplicationName: "myApp"}, &cpe, utilsMock)

			assert.False(t, utilsMock.HasWrittenFile("mta.yaml"))
		})
		t.Run("create mta.yaml from config.source", func(t *testing.T) {
			utilsMock := newMtaBuildTestUtilsBundle()

			utilsMock.AddFile("package.json", []byte(`{"name": "myName", "version": "1.2.3"}`))

			_ = runMtaBuild(mtaBuildOptions{ApplicationName: "myApp", Source: "create"}, &cpe, utilsMock)

			assert.True(t, utilsMock.HasWrittenFile("create/mta.yaml"))
		})
		t.Run("read yaml from config.source", func(t *testing.T) {
			utilsMock := newMtaBuildTestUtilsBundle()

			utilsMock.AddFile("path/mta.yaml", []byte("already there"))

			_ = runMtaBuild(mtaBuildOptions{ApplicationName: "myApp", Source: "path"}, &cpe, utilsMock)

			assert.False(t, utilsMock.HasWrittenFile("path/mta.yaml"))
		})
	})

	t.Run("MTA build should enable create BOM", func(t *testing.T) {
		utilsMock := newMtaBuildTestUtilsBundle()

		options := mtaBuildOptions{ApplicationName: "myApp", Platform: "CF", DefaultNpmRegistry: "https://example.org/npm", MtarName: "myName", Source: "./", Target: "./", CreateBOM: true}
		utilsMock.AddFile("package.json", []byte(`{"name": "myName", "version": "1.2.3"}`))

		err := runMtaBuild(options, &cpe, utilsMock)
		assert.Nil(t, err)
		assert.Contains(t, utilsMock.Calls[0].Params, "--sbom-file-path")
	})
}

func TestMtaBuildMtar(t *testing.T) {
	t.Run("getMtarName", func(t *testing.T) {
		t.Parallel()

		t.Run("mtar name from yaml", func(t *testing.T) {
			utilsMock := newMtaBuildTestUtilsBundle()
			utilsMock.AddFile("mta.yaml", []byte(`ID: "nameFromMtar"`))

			assert.Equal(t, filepath.FromSlash("nameFromMtar.mtar"), _ignoreErrorForGetMtarName(getMtarName(mtaBuildOptions{MtarName: ""}, "mta.yaml", utilsMock)))
		})
		t.Run("mtar name from yaml with suffixed value", func(t *testing.T) {
			utilsMock := newMtaBuildTestUtilsBundle()
			utilsMock.AddFile("mta.yaml", []byte(`ID: "nameFromMtar.mtar"`))

			assert.Equal(t, filepath.FromSlash("nameFromMtar.mtar"), _ignoreErrorForGetMtarName(getMtarName(mtaBuildOptions{MtarName: ""}, "mta.yaml", utilsMock)))
		})
		t.Run("mtar name from config", func(t *testing.T) {
			utilsMock := newMtaBuildTestUtilsBundle()
			utilsMock.AddFile("mta.yaml", []byte(`ID: "nameFromMtar"`))

			assert.Equal(t, filepath.FromSlash("nameFromConfig.mtar"), _ignoreErrorForGetMtarName(getMtarName(mtaBuildOptions{MtarName: "nameFromConfig.mtar"}, "mta.yaml", utilsMock)))
		})
	})

	t.Run("getMtarFilePath", func(t *testing.T) {
		t.Parallel()

		t.Run("plain mtar name", func(t *testing.T) {
			assert.Equal(t, "mta.mtar", getMtarFilePath(mtaBuildOptions{Source: "", Target: ""}, "mta.mtar"))
		})
		t.Run("plain mtar name from default", func(t *testing.T) {
			assert.Equal(t, "mta.mtar", getMtarFilePath(mtaBuildOptions{Source: "./", Target: "./"}, "mta.mtar"))
		})
		t.Run("source path", func(t *testing.T) {
			assert.Equal(t, filepath.FromSlash("source/mta.mtar"), getMtarFilePath(mtaBuildOptions{Source: "source", Target: ""}, "mta.mtar"))
		})
		t.Run("target path", func(t *testing.T) {
			assert.Equal(t, filepath.FromSlash("target/mta.mtar"), getMtarFilePath(mtaBuildOptions{Source: "", Target: "target"}, "mta.mtar"))
		})
		t.Run("source and target path", func(t *testing.T) {
			assert.Equal(t, filepath.FromSlash("source/target/mta.mtar"), getMtarFilePath(mtaBuildOptions{Source: "source", Target: "target"}, "mta.mtar"))
		})
	})
}

func _ignoreError(s string, e error) string {
	return s
}

func _ignoreErrorForGetMtarName(s string, b bool, e error) string {
	return s
}
