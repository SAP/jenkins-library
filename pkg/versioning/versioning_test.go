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
		assert.Equal(t, "test.ini", theType.Path)
		assert.Equal(t, "theversion", theType.VersionField)
		assert.Equal(t, "test", theType.VersionSection)
		assert.Equal(t, "semver2", custom.VersioningScheme())
	})

	t.Run("dub", func(t *testing.T) {
		dub, err := GetArtifact("dub", "", &Options{VersionField: "theversion"}, nil)

		assert.NoError(t, err)

		theType, ok := dub.(*JSONfile)
		assert.True(t, ok)
		assert.Equal(t, "dub.json", theType.Path)
		assert.Equal(t, "version", theType.VersionField)
		assert.Equal(t, "semver2", dub.VersioningScheme())
	})

	t.Run("golang", func(t *testing.T) {
		fileExists = func(string) (bool, error) { return true, nil }
		golang, err := GetArtifact("golang", "", &Options{}, nil)

		assert.NoError(t, err)

		theType, ok := golang.(*Versionfile)
		assert.True(t, ok)
		assert.Equal(t, "VERSION", theType.Path)
		assert.Equal(t, "semver2", golang.VersioningScheme())
	})

	t.Run("golang - error", func(t *testing.T) {
		fileExists = func(string) (bool, error) { return false, nil }
		_, err := GetArtifact("golang", "", &Options{}, nil)

		assert.EqualError(t, err, "no build descriptor available, supported: [VERSION version.txt]")
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
		assert.Equal(t, "pom.xml", theType.PomPath)
		assert.Equal(t, opts.ProjectSettingsFile, theType.ProjectSettingsFile)
		assert.Equal(t, opts.GlobalSettingsFile, theType.GlobalSettingsFile)
		assert.Equal(t, opts.M2Path, theType.M2Path)
		assert.Equal(t, "maven", maven.VersioningScheme())
	})

	t.Run("mta", func(t *testing.T) {
		mta, err := GetArtifact("mta", "", &Options{VersionField: "theversion"}, nil)

		assert.NoError(t, err)

		theType, ok := mta.(*YAMLfile)
		assert.True(t, ok)
		assert.Equal(t, "mta.yaml", theType.Path)
		assert.Equal(t, "version", theType.VersionField)
		assert.Equal(t, "semver2", mta.VersioningScheme())
	})

	t.Run("npm", func(t *testing.T) {
		npm, err := GetArtifact("npm", "", &Options{VersionField: "theversion"}, nil)

		assert.NoError(t, err)

		theType, ok := npm.(*JSONfile)
		assert.True(t, ok)
		assert.Equal(t, "package.json", theType.Path)
		assert.Equal(t, "version", theType.VersionField)
		assert.Equal(t, "semver2", npm.VersioningScheme())
	})

	t.Run("pip", func(t *testing.T) {
		fileExists = func(string) (bool, error) { return true, nil }
		pip, err := GetArtifact("pip", "", &Options{}, nil)

		assert.NoError(t, err)

		theType, ok := pip.(*Versionfile)
		assert.True(t, ok)
		assert.Equal(t, "version.txt", theType.Path)
		assert.Equal(t, "pep440", pip.VersioningScheme())
	})

	t.Run("pip - error", func(t *testing.T) {
		fileExists = func(string) (bool, error) { return false, nil }
		_, err := GetArtifact("pip", "", &Options{}, nil)

		assert.EqualError(t, err, "no build descriptor available, supported: [version.txt VERSION]")
	})

	t.Run("sbt", func(t *testing.T) {
		sbt, err := GetArtifact("sbt", "", &Options{VersionField: "theversion"}, nil)

		assert.NoError(t, err)

		theType, ok := sbt.(*JSONfile)
		assert.True(t, ok)
		assert.Equal(t, "sbtDescriptor.json", theType.Path)
		assert.Equal(t, "version", theType.VersionField)
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
		expected    Artifact
		expectedErr string
	}{
		{file: "not.supported", expectedErr: "file type not supported: 'not.supported'"},
		{file: "test.cfg", field: "testField", section: "testSection", expected: &INIfile{Path: "test.cfg", VersionField: "testField", VersionSection: "testSection"}},
		{file: "test.ini", field: "testField", section: "testSection", expected: &INIfile{Path: "test.ini", VersionField: "testField", VersionSection: "testSection"}},
		{file: "test.json", field: "testField", expected: &JSONfile{Path: "test.json", VersionField: "testField"}},
		{file: "test.yaml", field: "testField", expected: &YAMLfile{Path: "test.yaml", VersionField: "testField"}},
		{file: "test.yml", field: "testField", expected: &YAMLfile{Path: "test.yml", VersionField: "testField"}},
		{file: "test.txt", expected: &Versionfile{Path: "test.txt"}},
		{file: "test", expected: &Versionfile{Path: "test"}},
	}

	for _, test := range tt {
		res, err := customArtifact(test.file, test.field, test.section)

		if len(test.expectedErr) == 0 {
			assert.NoError(t, err)
			assert.Equal(t, test.expected, res)
		} else {
			assert.EqualError(t, err, test.expectedErr)
		}
	}
}
