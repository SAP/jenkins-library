//go:build unit
// +build unit

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

	"github.com/stretchr/testify/assert"
	"go.yaml.in/yaml/v3"
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
			t.Setenv("M2_HOME", "/opt/maven")
			utilsMock := newMtaBuildTestUtilsBundle()
			utilsMock.AddFile("mta.yaml", []byte(`ID: "myNameFromMtar"`))
			utilsMock.AddFile("/opt/maven/conf/settings.xml", []byte(snapshotSettingsXML))

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
			t.Setenv("M2_HOME", "/opt/maven")
			utilsMock := newMtaBuildTestUtilsBundle()
			utilsMock.AddFile("mta.yaml", []byte(`ID: "myNameFromMtar"`))
			utilsMock.AddFile("/opt/maven/conf/settings.xml", []byte(snapshotSettingsXML))

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
			t.Setenv("M2_HOME", "/opt/maven")
			utilsMock := newMtaBuildTestUtilsBundle()
			utilsMock.AddFile("mta.yaml", []byte(`ID: "myNameFromMtar"`))
			utilsMock.AddFile("/opt/maven/conf/settings.xml", []byte(snapshotSettingsXML))

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
			t.Setenv("M2_HOME", "/opt/maven")
			utilsMock := newMtaBuildTestUtilsBundle()
			utilsMock.sendRequestReturns = func() (*http.Response, error) {
				return &http.Response{StatusCode: 200}, nil
			}
			utilsMock.AddFile("mta.yaml", []byte(`ID: "myNameFromMtar"`))
			utilsMock.AddFile("/opt/maven/conf/settings.xml", []byte(snapshotSettingsXML))
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
			t.Setenv("M2_HOME", "/opt/maven")
			utilsMock := newMtaBuildTestUtilsBundle()
			utilsMock.AddFile("mta.yaml", []byte(`ID: "myNameFromMtar"`))
			utilsMock.AddFile("/opt/maven/conf/settings.xml", []byte(snapshotSettingsXML))
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
			assert.Equal(t, cpe.custom.mtaBuildArtifacts, `{"Coordinates":[{"groupId":"dummy","artifactId":"test","version":"dummy","packaging":"mtar","buildPath":"./","url":"dummy","purl":""}]}`)
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

// snapshotSettingsXML is a minimal SAP-Artifactory-style global settings.xml
// that has snapshot.build as the default active profile, matching the real-world
// pre-fix state.  Go's xml.Unmarshal will parse the namespace-qualified root
// element into maven.Settings when M2_HOME is set and the file is present.
const snapshotSettingsXML = `<?xml version="1.0" encoding="UTF-8"?>
<settings xmlns="http://maven.apache.org/SETTINGS/1.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://maven.apache.org/SETTINGS/1.0.0 https://maven.apache.org/xsd/settings-1.0.0.xsd">
    <activeProfiles>
        <activeProfile>snapshot.build</activeProfile>
    </activeProfiles>
</settings>`

func TestBuildMtaEffectiveProfiles(t *testing.T) {
	t.Parallel()

	t.Run("nil user profiles activates release.build", func(t *testing.T) {
		assert.Equal(t, []string{"release.build"}, buildMtaEffectiveProfiles(nil))
	})

	t.Run("empty user profiles activates release.build", func(t *testing.T) {
		assert.Equal(t, []string{"release.build"}, buildMtaEffectiveProfiles([]string{}))
	})

	t.Run("milestone opt-in passes profiles as-is", func(t *testing.T) {
		// milestone.build activates its own repository mapping; release.build must NOT
		// be co-activated (it would pull in the wrong quality repository).
		assert.Equal(t, []string{"milestone.build"}, buildMtaEffectiveProfiles([]string{"milestone.build"}))
	})

	t.Run("milestone opt-in with extra profiles passes profiles as-is", func(t *testing.T) {
		assert.Equal(t, []string{"milestone.build", "custom"},
			buildMtaEffectiveProfiles([]string{"milestone.build", "custom"}))
	})

	t.Run("snapshot opt-in passes profiles as-is", func(t *testing.T) {
		assert.Equal(t, []string{"snapshot.build"}, buildMtaEffectiveProfiles([]string{"snapshot.build"}))
	})

	t.Run("custom profiles get release.build prepended", func(t *testing.T) {
		// Non-quality-indicator profiles get release.build prepended so the
		// SAP Artifactory release repository remains accessible.
		assert.Equal(t, []string{"release.build", "my-profile"},
			buildMtaEffectiveProfiles([]string{"my-profile"}))
	})

	t.Run("multiple custom profiles get release.build prepended once", func(t *testing.T) {
		assert.Equal(t, []string{"release.build", "p1", "p2"},
			buildMtaEffectiveProfiles([]string{"p1", "p2"}))
	})
}

func TestMtaBuildProfileSettingsXMLUpdate(t *testing.T) {
	const testM2Home = "/opt/test/m2"
	const settingsPath = testM2Home + "/conf/settings.xml"

	cpe := mtaBuildCommonPipelineEnvironment{}

	t.Run("default profiles rewrite settings.xml to release.build", func(t *testing.T) {
		// When no user profiles are provided the step must switch the active
		// profile from the SAP Artifactory default (snapshot.build) to release.build.
		t.Setenv("M2_HOME", testM2Home)

		utilsMock := newMtaBuildTestUtilsBundle()
		utilsMock.AddFile("mta.yaml", []byte(`ID: "myApp"`))
		utilsMock.AddFile(settingsPath, []byte(snapshotSettingsXML))

		options := mtaBuildOptions{
			ApplicationName:    "myApp",
			Platform:           "CF",
			MtarName:           "myApp.mtar",
			Source:             "./",
			Target:             "./",
			GlobalSettingsFile: "https://repo.example.com/settings.xml",
		}

		err := runMtaBuild(options, &cpe, utilsMock)
		assert.Nil(t, err)

		content, readErr := utilsMock.FileRead(settingsPath)
		if assert.Nil(t, readErr, "settings.xml must be readable after build") {
			s := string(content)
			assert.Contains(t, s, "release.build",
				"settings.xml must contain release.build as active profile")
			assert.NotContains(t, s, "snapshot.build",
				"settings.xml must not contain snapshot.build after default-profile rewrite")
		}
	})

	t.Run("milestone opt-in keeps milestone.build in settings.xml", func(t *testing.T) {
		t.Setenv("M2_HOME", testM2Home)

		utilsMock := newMtaBuildTestUtilsBundle()
		utilsMock.AddFile("mta.yaml", []byte(`ID: "myApp"`))
		utilsMock.AddFile(settingsPath, []byte(snapshotSettingsXML))

		options := mtaBuildOptions{
			ApplicationName:    "myApp",
			Platform:           "CF",
			MtarName:           "myApp.mtar",
			Source:             "./",
			Target:             "./",
			GlobalSettingsFile: "https://repo.example.com/settings.xml",
			Profiles:           []string{"milestone.build"},
		}

		err := runMtaBuild(options, &cpe, utilsMock)
		assert.Nil(t, err)

		content, readErr := utilsMock.FileRead(settingsPath)
		if assert.Nil(t, readErr, "settings.xml must be readable after build") {
			s := string(content)
			assert.Contains(t, s, "milestone.build",
				"settings.xml must contain milestone.build for milestone quality build")
			assert.NotContains(t, s, "release.build",
				"settings.xml must not contain release.build when milestone.build is the active profile")
		}
	})

	t.Run("snapshot opt-in keeps snapshot.build in settings.xml", func(t *testing.T) {
		t.Setenv("M2_HOME", testM2Home)

		utilsMock := newMtaBuildTestUtilsBundle()
		utilsMock.AddFile("mta.yaml", []byte(`ID: "myApp"`))
		utilsMock.AddFile(settingsPath, []byte(snapshotSettingsXML))

		options := mtaBuildOptions{
			ApplicationName:    "myApp",
			Platform:           "CF",
			MtarName:           "myApp.mtar",
			Source:             "./",
			Target:             "./",
			GlobalSettingsFile: "https://repo.example.com/settings.xml",
			Profiles:           []string{"snapshot.build"},
		}

		err := runMtaBuild(options, &cpe, utilsMock)
		assert.Nil(t, err)

		content, readErr := utilsMock.FileRead(settingsPath)
		if assert.Nil(t, readErr, "settings.xml must be readable after build") {
			s := string(content)
			assert.Contains(t, s, "snapshot.build",
				"settings.xml must contain snapshot.build for snapshot quality build")
		}
	})

	t.Run("custom profiles prepend release.build in settings.xml", func(t *testing.T) {
		t.Setenv("M2_HOME", testM2Home)

		utilsMock := newMtaBuildTestUtilsBundle()
		utilsMock.AddFile("mta.yaml", []byte(`ID: "myApp"`))
		utilsMock.AddFile(settingsPath, []byte(snapshotSettingsXML))

		options := mtaBuildOptions{
			ApplicationName:    "myApp",
			Platform:           "CF",
			MtarName:           "myApp.mtar",
			Source:             "./",
			Target:             "./",
			GlobalSettingsFile: "https://repo.example.com/settings.xml",
			Profiles:           []string{"my-custom-profile"},
		}

		err := runMtaBuild(options, &cpe, utilsMock)
		assert.Nil(t, err)

		content, readErr := utilsMock.FileRead(settingsPath)
		if assert.Nil(t, readErr, "settings.xml must be readable after build") {
			s := string(content)
			assert.Contains(t, s, "release.build",
				"settings.xml must contain release.build alongside custom profile")
			assert.Contains(t, s, "my-custom-profile",
				"settings.xml must contain the custom profile")
		}
	})

	t.Run("no globalSettingsFile skips settings.xml update", func(t *testing.T) {
		// When globalSettingsFile is not provided there is no $M2_HOME/conf/settings.xml
		// for mbt's internal Maven to read, so the step must not attempt to update it.
		t.Setenv("M2_HOME", testM2Home)

		utilsMock := newMtaBuildTestUtilsBundle()
		utilsMock.AddFile("mta.yaml", []byte(`ID: "myApp"`))
		// Deliberately NOT adding settings.xml to the mock — if the step tries to read
		// it, utils.FileRead will fail, which would surface as a test error.

		options := mtaBuildOptions{
			ApplicationName: "myApp",
			Platform:        "CF",
			MtarName:        "myApp.mtar",
			Source:          "./",
			Target:          "./",
			// GlobalSettingsFile intentionally empty
		}

		err := runMtaBuild(options, &cpe, utilsMock)
		assert.Nil(t, err, "build must succeed when no globalSettingsFile is provided")

		_, readErr := utilsMock.FileRead(settingsPath)
		assert.NotNil(t, readErr,
			"settings.xml must not exist (was not created/written) when no globalSettingsFile is provided")
	})
}
