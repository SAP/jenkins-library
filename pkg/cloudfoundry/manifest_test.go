//go:build unit
// +build unit

package cloudfoundry

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func writeManifest(t *testing.T, content string) string {
	t.Helper()
	path := t.TempDir() + "/manifest.yml"
	assert.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func TestReadManifest(t *testing.T) {
	path := writeManifest(t, "applications: [{name: manifestAppName}]\n")

	manifest, err := ReadManifest(path)

	if assert.NoError(t, err) {
		name, err := manifest.GetAppName(0)
		assert.NoError(t, err)
		assert.Equal(t, "manifestAppName", name)
		assert.Equal(t, path, manifest.GetFileName())
	}
}

func TestReadManifestErrors(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		_, err := ReadManifest(t.TempDir() + "/missing.yml")
		assert.Error(t, err)
	})
	t.Run("invalid YAML", func(t *testing.T) {
		path := writeManifest(t, "applications: [\n")
		_, err := ReadManifest(path)
		assert.Error(t, err)
	})
}

func TestManifestApplications(t *testing.T) {
	path := writeManifest(t, "applications: [{name: firstApp}, {name: secondApp, no-route: true}]\n")
	manifest, err := ReadManifest(path)
	if !assert.NoError(t, err) {
		return
	}

	apps, err := manifest.GetApplications()
	assert.NoError(t, err)
	assert.Equal(t, []map[string]any{
		{"name": "firstApp"},
		{"name": "secondApp", "no-route": true},
	}, apps)

	hasName, err := manifest.ApplicationHasProperty(0, "name")
	assert.NoError(t, err)
	assert.True(t, hasName)
	noRoute, err := manifest.GetApplicationProperty(1, "no-route")
	assert.NoError(t, err)
	assert.Equal(t, true, noRoute)

	_, err = manifest.GetApplicationProperty(0, "missing")
	assert.EqualError(t, err, "No such property: 'missing' available in application at position 0")
	_, err = manifest.ApplicationHasProperty(2, "name")
	assert.EqualError(t, err, "Index (2) out of bound. Number of apps: 2")
}

func TestManifestMissingApplications(t *testing.T) {
	path := writeManifest(t, "noApps: true\n")
	manifest, err := ReadManifest(path)
	if !assert.NoError(t, err) {
		return
	}

	_, err = manifest.GetApplications()
	assert.EqualError(t, err, "Failed to convert <nil> to slice. Was <nil>")
}

func TestManifestApplicationPropertyDoesNotExist(t *testing.T) {
	path := writeManifest(t, "applications: [{name: app}]\n")
	manifest, err := ReadManifest(path)
	if !assert.NoError(t, err) {
		return
	}

	hasProperty, err := manifest.ApplicationHasProperty(0, "missing")
	assert.NoError(t, err)
	assert.False(t, hasProperty)
}

func TestManifestTransform(t *testing.T) {
	t.Run("moves a single buildpack and persists the result", func(t *testing.T) {
		path := writeManifest(t, "applications: [{name: app, no-route: true, buildpacks: [sap_java_buildpack]}]\n")
		manifest, err := ReadManifest(path)
		if !assert.NoError(t, err) {
			return
		}
		assert.NoError(t, manifest.Transform())
		assert.True(t, manifest.IsModified())
		buildpack, err := manifest.GetApplicationProperty(0, "buildpack")
		assert.NoError(t, err)
		assert.Equal(t, "sap_java_buildpack", buildpack)
		_, err = manifest.GetApplicationProperty(0, "buildpacks")
		assert.EqualError(t, err, "No such property: 'buildpacks' available in application at position 0")
		assert.NoError(t, manifest.WriteManifest())
		assert.False(t, manifest.IsModified())

		content, err := os.ReadFile(path)
		assert.NoError(t, err)
		assert.Equal(t, "applications:\n    - buildpack: sap_java_buildpack\n      name: app\n      no-route: true\n", string(content))
	})

	t.Run("rejects multiple buildpacks", func(t *testing.T) {
		path := writeManifest(t, "applications: [{name: app, buildpacks: [one, two]}]\n")
		manifest, err := ReadManifest(path)
		if assert.NoError(t, err) {
			assert.Error(t, manifest.Transform())
		}
	})

	t.Run("leaves manifests with a single buildpack unchanged", func(t *testing.T) {
		path := writeManifest(t, "applications: [{name: app, buildpack: sap_java_buildpack}]\n")
		manifest, err := ReadManifest(path)
		if assert.NoError(t, err) {
			assert.NoError(t, manifest.Transform())
			assert.False(t, manifest.IsModified())
			buildpack, err := manifest.GetApplicationProperty(0, "buildpack")
			assert.NoError(t, err)
			assert.Equal(t, "sap_java_buildpack", buildpack)
			_, err = manifest.GetApplicationProperty(0, "buildpacks")
			assert.EqualError(t, err, "No such property: 'buildpacks' available in application at position 0")
		}
	})
}
