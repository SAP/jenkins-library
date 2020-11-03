package versioning

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

	t.Run("golang", func(t *testing.T) {
		fileExists = func(string) (bool, error) { return true, nil }
		golang, err := GetArtifact("golang", "", &Options{}, nil)

		assert.NoError(t, err)

		theType, ok := golang.(*Versionfile)
		assert.True(t, ok)
		assert.Equal(t, "VERSION", theType.path)
		assert.Equal(t, "semver2", golang.VersioningScheme())
	})

	t.Run("golang - error", func(t *testing.T) {
		fileExists = func(string) (bool, error) { return false, nil }
		_, err := GetArtifact("golang", "", &Options{}, nil)

		assert.EqualError(t, err, "no build descriptor available, supported: [VERSION version.txt go.mod]")
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

	t.Run("pip", func(t *testing.T) {
		fileExists = func(string) (bool, error) { return true, nil }
		pip, err := GetArtifact("pip", "", &Options{}, nil)

		assert.NoError(t, err)

		theType, ok := pip.(*Pip)
		assert.True(t, ok)
		assert.Equal(t, "version.txt", theType.path)
		assert.Equal(t, "pep440", pip.VersioningScheme())
	})

	t.Run("pip - error", func(t *testing.T) {
		fileExists = func(string) (bool, error) { return false, nil }
		_, err := GetArtifact("pip", "", &Options{}, nil)

		assert.EqualError(t, err, "no build descriptor available, supported: [version.txt VERSION setup.py]")
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
		res, err := customArtifact(test.file, test.field, test.section, test.scheme)

		if len(test.expectedErr) == 0 {
			assert.NoError(t, err)
			assert.Equal(t, test.expected, res)
		} else {
			assert.EqualError(t, err, test.expectedErr)
		}
	}
}
