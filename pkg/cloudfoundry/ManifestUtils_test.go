package cloudfoundry

import (
	"testing"

	"fmt"
	"github.com/stretchr/testify/assert"
)

func TestReadManifest(t *testing.T) {

	_readFile = func(filename string) ([]byte, error) {
		if filename == "myManifest.yaml" {
			return []byte("applications: [{name: 'manifestAppName'}]"), nil
		}
		return []byte{}, fmt.Errorf("File '%s' not found", filename)
	}

	manifest, err := ReadManifest("myManifest.yaml")

	appName, err := manifest.GetAppName(0)
	if assert.NoError(t, err) {
		assert.Equal(t, "manifestAppName", appName)
	}
}

func TestNoRoute(t *testing.T) {

	_readFile = func(filename string) ([]byte, error) {
		if filename == "myManifest.yaml" {
			return []byte("applications: [{name: 'manifestAppName', no-route: true}]"), nil
		}
		return []byte{}, fmt.Errorf("File '%s' not found", filename)
	}

	manifest, err := ReadManifest("myManifest.yaml")
	if !assert.NoError(t, err) {
		assert.FailNow(t, "Cannot read manifest file")
	}

	noRoute, err := manifest.GetApplicationProperty(0, "no-route")
	if assert.NoError(t, err) {
		noRouteAsBool, ok := noRoute.(bool)

		if assert.True(t, ok) && assert.NoError(t, err) {
			assert.True(t, noRouteAsBool)
		}
	}
}

func TestTransformGoodCase(t *testing.T) {

	_readFile = func(filename string) ([]byte, error) {
		if filename == "myManifest.yaml" {
			return []byte("applications: [{name: 'manifestAppName', no-route: true, buildpacks: [sap_java_buildpack]}]"), nil
		}
		return []byte{}, fmt.Errorf("File '%s' not found", filename)
	}

	manifest, err := ReadManifest("myManifest.yaml")
	assert.NoError(t, err)

	err = manifest.Transform()

	assert.NoError(t, err)
	buildpack, err := manifest.GetApplicationProperty(0, "buildpack")
	assert.NoError(t, err)
	buildpacks, err := manifest.GetApplicationProperty(0, "buildpacks")

	assert.Equal(t, "sap_java_buildpack", buildpack)
	assert.Equal(t, "", buildpacks)
	assert.True(t, manifest.HasModified())

}

func TestTransformMultipleBuildPacks(t *testing.T) {
	_readFile = func(filename string) ([]byte, error) {
		if filename == "myManifest.yaml" {
			return []byte("applications: [{name: 'manifestAppName', buildpacks: [sap_java_buildpack, 'another_buildpack']}]"), nil
		}
		return []byte{}, fmt.Errorf("File '%s' not found", filename)
	}

	manifest, err := ReadManifest("myManifest.yaml")
	assert.NoError(t, err)

	err = manifest.Transform()

	assert.EqualError(t, err, "More than one Cloud Foundry Buildpack is not supported. Please check manifest file 'myManifest.yaml', application 'manifestAppName'")
}

func TestTransformUnchanged(t *testing.T) {
	_readFile = func(filename string) ([]byte, error) {
		if filename == "myManifest.yaml" {
			return []byte("applications: [{name: 'manifestAppName', no-route: true, buildpack: sap_java_buildpack}]"), nil
		}
		return []byte{}, fmt.Errorf("File '%s' not found", filename)
	}

	manifest, err := ReadManifest("myManifest.yaml")
	assert.NoError(t, err)

	err = manifest.Transform()

	buildpack, err := manifest.GetApplicationProperty(0, "buildpack")
	assert.NoError(t, err)
	_, err = manifest.GetApplicationProperty(0, "buildpacks")
	assert.Equal(t, "sap_java_buildpack", buildpack)
	assert.EqualError(t, err, "No such property: 'buildpacks' available in application at position 0")
	assert.False(t, manifest.HasModified())
}

func TestGetManifestName(t *testing.T) {

	_readFile = func(filename string) ([]byte, error) {
		if filename == "myManifest.yaml" {
			return []byte("applications: [{name: 'firstApp'}]"), nil
		}
		return []byte{}, fmt.Errorf("File '%s' not found", filename)
	}

	manifest, err := ReadManifest("myManifest.yaml")

	if assert.NoError(t, err) {
		assert.Equal(t, "myManifest.yaml", manifest.GetName())
	}
}

func TestApplicationHasProperty(t *testing.T) {

	_readFile = func(filename string) ([]byte, error) {
		if filename == "myManifest.yaml" {
			return []byte("applications: [{name: 'firstApp'}]"), nil
		}
		return []byte{}, fmt.Errorf("File '%s' not found", filename)
	}

	manifest, err := ReadManifest("myManifest.yaml")

	if assert.NoError(t, err) {

		t.Run("Property exists", func(t *testing.T) {
			hasProp, err := manifest.ApplicationHasProperty(0, "name")
			if assert.NoError(t, err) {
				assert.True(t, hasProp)
			}
		})

		t.Run("Property does not exist", func(t *testing.T) {
			hasProp, err := manifest.ApplicationHasProperty(0, "foo")
			if assert.NoError(t, err) {
				assert.False(t, hasProp)
			}
		})
		t.Run("Index out of bounds", func(t *testing.T) {
			_, err := manifest.ApplicationHasProperty(1, "foo")
			assert.EqualError(t, err, "Index (1) out of bound. Number of apps: 1")
		})
	}
}

func TestGetApplicationsWhenNoApplicationNoIsPresent(t *testing.T) {

	_readFile = func(filename string) ([]byte, error) {
		if filename == "myManifest.yaml" {
			return []byte("noApps: true"), nil
		}
		return []byte{}, fmt.Errorf("File '%s' not found", filename)
	}

	manifest, err := ReadManifest("myManifest.yaml")
	_, err = manifest.GetApplications()

	assert.EqualError(t, err, "Failed to convert <nil> to slice. Was <nil>")
}
func TestGetApplications(t *testing.T) {

	_readFile = func(filename string) ([]byte, error) {
		if filename == "myManifest.yaml" {
			return []byte("applications: [{name: 'firstApp'}, {name: 'secondApp'}]"), nil
		}
		return []byte{}, fmt.Errorf("File '%s' not found", filename)
	}

	manifest, err := ReadManifest("myManifest.yaml")
	apps, err := manifest.GetApplications()

	if assert.NoError(t, err) {
		assert.Len(t, apps, 2)
		assert.Equal(t, map[string]interface{}{"name": "firstApp"}, apps[0])
		assert.Equal(t, map[string]interface{}{"name": "secondApp"}, apps[1])

	}
}
