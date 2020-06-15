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
			return []byte("no-route: true\napplications: [{name: 'manifestAppName', buildpacks: [sap_java_buildpack, 'another_buildpack']}]"), nil
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
