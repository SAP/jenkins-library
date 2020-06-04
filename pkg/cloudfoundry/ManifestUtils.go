package cloudfoundry

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"io/ioutil"
	"reflect"

	"github.com/SAP/jenkins-library/pkg/log"
)

const propApplications = "applications"
const propBuildpacks = "buildpacks"
const propBuildpack = "buildpack"

// Manifest ...
type Manifest struct {
	self     map[string]interface{}
	modified bool
	name     string
}

var m Manifest

var _readFile = ioutil.ReadFile

// ReadManifest ...
func ReadManifest(name string) (Manifest, error) {

	log.Entry().Infof("Reading manifest file  '%s'", name)

	m := Manifest{self: make(map[string]interface{}), name: name, modified: false}

	content, err := _readFile(name)
	if err != nil {
		return m, errors.Wrapf(err, "cannot read file '%v'", m.name)
	}

	err = yaml.Unmarshal(content, &m.self)

	if err != nil {
		return m, errors.Wrapf(err, "Cannot parse yaml file '%s': %s", m.name, string(content))
	}

	log.Entry().Infof("Manifest file '%s' has been parsed", m.name)

	return m, nil
}

// WriteManifest ...
func (m *Manifest) WriteManifest() error {

	d, err := yaml.Marshal(&m.self)
	if err != nil {
		return err
	}

	log.Entry().Debugf("Writing manifest file '%s'", m.name)
	err = ioutil.WriteFile(m.name, d, 0644)

	if err == nil {
		m.modified = false
	}

	log.Entry().Debugf("Manifest file '%s' has been written", m.name)
	return err
}

// GetName ...
func (m Manifest) GetName() string {
	return m.name
}

// GetApplications ...
func (m Manifest) GetApplications() ([]interface{}, error) {
	return toSlice(m.self)
}

// ApplicationHasProperty Checks if the application denoted by 'index' has the property 'name'
func (m Manifest) ApplicationHasProperty(index int, name string) (bool, error) {

	sliced, err := toSlice(m.self[propApplications])

	if err != nil {
		return false, err
	}

	if index >= len(sliced) {
		return false, fmt.Errorf("Index (%d) out of bound. Number of apps: %d", index, len(sliced))
	}

	_m, err := toMap(sliced[index])

	if err != nil {
		return false, err
	}

	_, ok := _m[name]

	return ok, nil
}

// GetApplicationProperty ...
func (m Manifest) GetApplicationProperty(index int, name string) (interface{}, error) {

	sliced, err := toSlice(m.self[propApplications])

	if err != nil {
		return "", err
	}

	if index >= len(sliced) {
		return "", fmt.Errorf("Index (%d) out of bound. Number of apps: %d", index, len(sliced))
	}

	app, err := toMap(sliced[index])

	if err != nil {
		return "", err
	}

	if app[name] != nil {
		return app[name], nil
	}

	return "", fmt.Errorf("No such property: '%s' available in application at position %d", name, index)
}

// GetAppName Gets the name of the app at 'index'
func (m Manifest) GetAppName(index int) (string, error) {

	appName, err := m.GetApplicationProperty(index, "name")

	if err != nil {
		return "", err
	}

	if name, ok := appName.(string); ok {
		return name, nil
	}

	return "", fmt.Errorf("Cannot retrieve application name for app at index %d", index)
}

// Transform ...
func (m *Manifest) Transform() error {

	sliced, err := toSlice(m.self[propApplications])
	if err != nil {
		return err
	}

	for _, app := range sliced {
		appAsMap, err := toMap(app)
		if err != nil {
			return err
		}

		err = transformApp(appAsMap, m)

		if err != nil {
			return err
		}
	}

	return nil
}

func transformApp(app map[string]interface{}, m *Manifest) error {

	appName := "n/a"

	if n, ok := app["name"].(string); ok {
		if len(n) > 0 {
			appName = n
		}
	}

	if app[propBuildpacks] == nil {
		// Revisit: not sure if a build pack is mandatory.
		// In that case we should check that app.buildpack
		// is present.
		return nil
	}

	buildPacks, err := toSlice(app[propBuildpacks])

	if err != nil {
		return err
	}

	if len(buildPacks) > 1 {
		return fmt.Errorf("More than one Cloud Foundry Buildpack is not supported. Please check manifest file '%s', application '%s'", m.name, appName)
	}

	if len(buildPacks) == 1 {
		app[propBuildpack] = buildPacks[0]
		delete(app, propBuildpacks)
		m.modified = true
	}

	return nil
}

// HasModified ...
func (m Manifest) HasModified() bool {
	return m.modified
}
func toMap(i interface{}) (map[string]interface{}, error) {

	if m, ok := i.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, fmt.Errorf("Failed to convert %v to map. Was %v", i, reflect.TypeOf(i))
}

func toSlice(i interface{}) ([]interface{}, error) {

	if s, ok := i.([]interface{}); ok {
		return s, nil
	}
	return nil, fmt.Errorf("Failed to convert %v to slice. Was %v", i, reflect.TypeOf(i))
}
