//go:build unit
// +build unit

package versioning

import (
	"net/http"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type versioningMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newVersioningMockUtils() *versioningMockUtils {
	utils := versioningMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return &utils
}

func (v *versioningMockUtils) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	// so far no dedicated logic required for testing
	return nil
}

func TestGetArtifact(t *testing.T) {
	t.Run("custom", func(t *testing.T) {
		custom, err := GetArtifact("custom", "test.ini", &Options{VersionField: "theversion", VersionSection: "test"}, nil)

		assert.NoError(t, err)

		theType, ok := custom.(*INIfile)
		assert.True(t, ok)
		assert.Equal(t, "test.ini", theType.path)
		assert.Equal(t, "theversion", theType.versionField)
		assert.Equal(t, "test", theType.versionSection)
		assert.Equal(t, "semver2", custom.VersioningScheme())
	})

	t.Run("docker", func(t *testing.T) {
		docker, err := GetArtifact("docker", "test.ini", &Options{VersionSource: "custom", VersionField: "theversion", VersionSection: "test"}, nil)

		assert.NoError(t, err)

		theType, ok := docker.(*Docker)
		assert.True(t, ok)
		assert.Equal(t, "test.ini", theType.path)
		assert.Equal(t, "theversion", theType.options.VersionField)
		assert.Equal(t, "test", theType.options.VersionSection)
		assert.Equal(t, "docker", docker.VersioningScheme())
	})

	t.Run("dub", func(t *testing.T) {
		dub, err := GetArtifact("dub", "", &Options{VersionField: "theversion"}, nil)

		assert.NoError(t, err)

		theType, ok := dub.(*JSONfile)
		assert.True(t, ok)
		assert.Equal(t, "dub.json", theType.path)
		assert.Equal(t, "version", theType.versionField)
		assert.Equal(t, "semver2", dub.VersioningScheme())
	})

	t.Run("golang - version file", func(t *testing.T) {
		fileExists = func(s string) (bool, error) {
			if s == "go.mod" {
				return false, nil
			}
			return true, nil
		}
		golang, err := GetArtifact("golang", "", &Options{}, nil)

		assert.NoError(t, err)

		theType, ok := golang.(*Versionfile)
		assert.True(t, ok)
		assert.Equal(t, "VERSION", theType.path)
		assert.Equal(t, "semver2", golang.VersioningScheme())
	})

	t.Run("golang - gomod", func(t *testing.T) {
		fileExists = func(s string) (bool, error) {
			if s == "go.mod" {
				return true, nil
			}
			return false, nil
		}
		golang, err := GetArtifact("golang", "", &Options{}, nil)

		assert.NoError(t, err)

		theType, ok := golang.(*GoMod)
		assert.True(t, ok)
		assert.Equal(t, "go.mod", theType.path)
		assert.Equal(t, "semver2", golang.VersioningScheme())
	})

	t.Run("golang - error", func(t *testing.T) {
		fileExists = func(string) (bool, error) { return false, nil }
		_, err := GetArtifact("golang", "", &Options{}, nil)

		assert.EqualError(t, err, "no build descriptor available, supported: [go.mod VERSION version.txt]")
	})

	t.Run("gradle", func(t *testing.T) {
		gradle, err := GetArtifact("gradle", "", &Options{VersionField: "theversion"}, nil)

		assert.NoError(t, err)

		theType, ok := gradle.(*Gradle)
		assert.True(t, ok)
		assert.Equal(t, "gradle.properties", theType.path)
		assert.Equal(t, "theversion", theType.versionField)
		assert.Equal(t, "semver2", gradle.VersioningScheme())
	})

	t.Run("helm", func(t *testing.T) {
		helm, err := GetArtifact("helm", "testchart/Chart.yaml", &Options{}, nil)

		assert.NoError(t, err)

		theType, ok := helm.(*HelmChart)
		assert.True(t, ok)
		assert.Equal(t, "testchart/Chart.yaml", theType.path)
		assert.Equal(t, "semver2", helm.VersioningScheme())
	})

	t.Run("maven", func(t *testing.T) {
		opts := Options{
			ProjectSettingsFile: "projectsettings.xml",
			GlobalSettingsFile:  "globalsettings.xml",
			M2Path:              "m2/path",
		}
		maven, err := GetArtifact("maven", "", &opts, nil)
		assert.NoError(t, err)

		theType, ok := maven.(*Maven)
		assert.True(t, ok)
		assert.Equal(t, "pom.xml", theType.options.PomPath)
		assert.Equal(t, opts.ProjectSettingsFile, theType.options.ProjectSettingsFile)
		assert.Equal(t, opts.GlobalSettingsFile, theType.options.GlobalSettingsFile)
		assert.Equal(t, opts.M2Path, theType.options.M2Path)
		assert.Equal(t, "maven", maven.VersioningScheme())
	})

	t.Run("CAP - maven", func(t *testing.T) {
		opts := Options{
			ProjectSettingsFile:     "projectsettings.xml",
			GlobalSettingsFile:      "globalsettings.xml",
			M2Path:                  "m2/path",
			CAPVersioningPreference: "maven",
		}
		maven, err := GetArtifact("CAP", "", &opts, nil)
		assert.NoError(t, err)

		theType, ok := maven.(*Maven)
		assert.True(t, ok)
		assert.Equal(t, "pom.xml", theType.options.PomPath)
		assert.Equal(t, opts.ProjectSettingsFile, theType.options.ProjectSettingsFile)
		assert.Equal(t, opts.GlobalSettingsFile, theType.options.GlobalSettingsFile)
		assert.Equal(t, opts.M2Path, theType.options.M2Path)
		assert.Equal(t, "maven", maven.VersioningScheme())
	})

	t.Run("mta", func(t *testing.T) {
		mta, err := GetArtifact("mta", "", &Options{VersionField: "theversion"}, nil)

		assert.NoError(t, err)

		theType, ok := mta.(*YAMLfile)
		assert.True(t, ok)
		assert.Equal(t, "mta.yaml", theType.path)
		assert.Equal(t, "version", theType.versionField)
		assert.Equal(t, "semver2", mta.VersioningScheme())
	})

	t.Run("npm", func(t *testing.T) {
		npm, err := GetArtifact("npm", "", &Options{VersionField: "theversion"}, nil)

		assert.NoError(t, err)

		theType, ok := npm.(*JSONfile)
		assert.True(t, ok)
		assert.Equal(t, "package.json", theType.path)
		assert.Equal(t, "version", theType.versionField)
		assert.Equal(t, "semver2", npm.VersioningScheme())
	})

	t.Run("CAP - npm", func(t *testing.T) {
		npm, err := GetArtifact("CAP", "", &Options{VersionField: "theversion", CAPVersioningPreference: "npm"}, nil)
		assert.NoError(t, err)

		theType, ok := npm.(*JSONfile)
		assert.True(t, ok)
		assert.Equal(t, "package.json", theType.path)
		assert.Equal(t, "version", theType.versionField)
		assert.Equal(t, "semver2", npm.VersioningScheme())
	})

	t.Run("yarn", func(t *testing.T) {
		npm, err := GetArtifact("yarn", "", &Options{VersionField: "theversion"}, nil)

		assert.NoError(t, err)

		theType, ok := npm.(*JSONfile)
		assert.True(t, ok)
		assert.Equal(t, "package.json", theType.path)
		assert.Equal(t, "version", theType.versionField)
		assert.Equal(t, "semver2", npm.VersioningScheme())
	})

	t.Run("pip", func(t *testing.T) {
		fileExists = func(string) (bool, error) { return true, nil }
		pip, err := GetArtifact("pip", "", &Options{}, nil)

		assert.NoError(t, err)

		theType, ok := pip.(*Pip)
		assert.True(t, ok)
		assert.Equal(t, "setup.py", theType.path)
		assert.Equal(t, "pep440", pip.VersioningScheme())
	})

	t.Run("pip - error", func(t *testing.T) {
		fileExists = func(string) (bool, error) { return false, nil }
		_, err := GetArtifact("pip", "", &Options{}, nil)

		assert.EqualError(t, err, "no build descriptor available, supported: [setup.py version.txt VERSION]")
	})

	t.Run("sbt", func(t *testing.T) {
		fileExists = func(string) (bool, error) { return true, nil }
		sbt, err := GetArtifact("sbt", "", &Options{VersionField: "theversion"}, nil)

		assert.NoError(t, err)

		theType, ok := sbt.(*JSONfile)
		assert.True(t, ok)
		assert.Equal(t, "sbtDescriptor.json", theType.path)
		assert.Equal(t, "version", theType.versionField)
		assert.Equal(t, "semver2", sbt.VersioningScheme())
	})

	t.Run("not supported build tool", func(t *testing.T) {
		_, err := GetArtifact("nosupport", "whatever", &Options{}, nil)
		assert.EqualError(t, err, "build tool 'nosupport' not supported")
	})
}

func TestCustomArtifact(t *testing.T) {
	tt := []struct {
		file        string
		field       string
		section     string
		scheme      string
		expected    Artifact
		expectedErr string
	}{
		{file: "not.supported", expectedErr: "file type not supported: 'not.supported'"},
		{file: "test.cfg", field: "testField", section: "testSection", expected: &INIfile{path: "test.cfg", versionField: "testField", versionSection: "testSection"}},
		{file: "test.ini", field: "testField", section: "testSection", expected: &INIfile{path: "test.ini", versionField: "testField", versionSection: "testSection"}},
		{file: "test.ini", field: "testField", section: "testSection", scheme: "maven", expected: &INIfile{path: "test.ini", versionField: "testField", versionSection: "testSection", versioningScheme: "maven"}},
		{file: "test.json", field: "testField", expected: &JSONfile{path: "test.json", versionField: "testField"}},
		{file: "test.yaml", field: "testField", expected: &YAMLfile{path: "test.yaml", versionField: "testField"}},
		{file: "test.yml", field: "testField", expected: &YAMLfile{path: "test.yml", versionField: "testField"}},
		{file: "test.txt", expected: &Versionfile{path: "test.txt"}},
		{file: "test", expected: &Versionfile{path: "test"}},
		{file: "test", scheme: "maven", expected: &Versionfile{path: "test", versioningScheme: "maven"}},
	}

	for _, test := range tt {
		t.Run(test.file, func(t *testing.T) {
			res, err := customArtifact(test.file, test.field, test.section, test.scheme)
			if len(test.expectedErr) == 0 {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, res)
			} else {
				assert.EqualError(t, err, test.expectedErr)
			}
		})

	}
}
