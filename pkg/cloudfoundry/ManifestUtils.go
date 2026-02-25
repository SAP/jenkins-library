package cloudfoundry

import (
	"fmt"
	"os"
	"reflect"

	"github.com/ghodss/yaml"

	"github.com/SAP/jenkins-library/pkg/log"
)

const constPropApplications = "applications"
const constPropBuildpacks = "buildpacks"
const constPropBuildpack = "buildpack"

// Manifest ...
type Manifest interface {
	GetFileName() string
	GetAppName(index int) (string, error)
	ApplicationHasProperty(index int, name string) (bool, error)
	GetApplicationProperty(index int, name string) (interface{}, error)
	Transform() error
	IsModified() bool
	GetApplications() ([]map[string]interface{}, error)
	WriteManifest() error
}

// manifest ...
type manifest struct {
	self     map[string]interface{}
	modified bool
	name     string
}

var _readFile = os.ReadFile
var _writeFile = os.WriteFile

// ReadManifest Reads the manifest denoted by 'name'
func ReadManifest(name string) (Manifest, error) {

	log.Entry().Infof("Reading manifest file  '%s'", name)

	m := &manifest{self: make(map[string]interface{}), name: name, modified: false}

	content, err := _readFile(name)
	if err != nil {
		return m, fmt.Errorf("cannot read file '%v': %w", m.name, err)
	}

	err = yaml.Unmarshal(content, &m.self)
	if err != nil {
		return m, fmt.Errorf("Cannot parse yaml file '%s': %s: %w", m.name, string(content), err)
	}

	log.Entry().Infof("Manifest file '%s' has been parsed", m.name)

	return m, nil
}

// WriteManifest Writes the manifest to the file denoted
// by the name property (GetFileName()). The modified flag is
// resetted after the write operation.
func (m *manifest) WriteManifest() error {

	d, err := yaml.Marshal(&m.self)
	if err != nil {
		return err
	}

	log.Entry().Debugf("Writing manifest file '%s'", m.GetFileName())
	err = _writeFile(m.GetFileName(), d, 0644)

	if err == nil {
		m.modified = false
	}

	log.Entry().Debugf("Manifest file '%s' has been written", m.name)
	return err
}

// GetFileName returns the file name of the manifest.
func (m *manifest) GetFileName() string {
	return m.name
}

// GetApplications Returns all applications denoted in the manifest file.
// The applications are returned as a slice of maps. Each app is represented by
// a map.
func (m *manifest) GetApplications() ([]map[string]interface{}, error) {
	apps, err := toSlice(m.self["applications"])
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0)

	for _, app := range apps {
		if _app, ok := app.(map[string]interface{}); ok {
			result = append(result, _app)
		} else {
			return nil, fmt.Errorf("Cannot cast applications to map. Manifest file '%s' has invalid format", m.GetFileName())
		}
	}
	return result, nil
}

// ApplicationHasProperty Checks if the application denoted by 'index' has the property 'name'
func (m *manifest) ApplicationHasProperty(index int, name string) (bool, error) {

	sliced, err := toSlice(m.self[constPropApplications])
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
func (m *manifest) GetApplicationProperty(index int, name string) (interface{}, error) {

	sliced, err := toSlice(m.self[constPropApplications])
	if err != nil {
		return nil, err
	}

	if index >= len(sliced) {
		return nil, fmt.Errorf("Index (%d) out of bound. Number of apps: %d", index, len(sliced))
	}

	app, err := toMap(sliced[index])
	if err != nil {
		return nil, err
	}

	value, exists := app[name]
	if exists {
		return value, nil
	}

	return nil, fmt.Errorf("No such property: '%s' available in application at position %d", name, index)
}

// GetAppName Gets the name of the app at 'index'
func (m *manifest) GetAppName(index int) (string, error) {

	appName, err := m.GetApplicationProperty(index, "name")
	if err != nil {
		return "", err
	}

	if name, ok := appName.(string); ok {
		return name, nil
	}

	return "", fmt.Errorf("Cannot retrieve application name for app at index %d", index)
}

// Transform For each app in the manifest the first entry in the build packs list
// gets moved to the top level under the key 'buildpack'. The 'buildpacks' list is
// deleted.
func (m *manifest) Transform() error {

	sliced, err := toSlice(m.self[constPropApplications])
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

func transformApp(app map[string]interface{}, m *manifest) error {

	appName := "n/a"

	if name, ok := app["name"].(string); ok {
		if len(name) > 0 {
			appName = name
		}
	}

	if app[constPropBuildpacks] == nil {
		// Revisit: not sure if a build pack is mandatory.
		// In that case we should check that app.buildpack
		// is present.
		return nil
	}

	buildPacks, err := toSlice(app[constPropBuildpacks])
	if err != nil {
		return err
	}

	if len(buildPacks) > 1 {
		return fmt.Errorf("More than one Cloud Foundry Buildpack is not supported. Please check manifest file '%s', application '%s'", m.name, appName)
	}

	if len(buildPacks) == 1 {
		app[constPropBuildpack] = buildPacks[0]
		delete(app, constPropBuildpacks)
		m.modified = true
	}

	return nil
}

// IsModified ...
func (m *manifest) IsModified() bool {
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
