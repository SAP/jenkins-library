package abaputils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

// AddonDescriptor contains fields about the addonProduct
type AddonDescriptor struct {
	AddonProduct     string      `json:"addonProduct"`
	AddonVersionYAML string      `json:"addonVersion"`
	AddonVersion     string      `json:"addonVersionAAK"`
	AddonUniqueID    string      `json:"addonUniqueID"`
	CustomerID       interface{} `json:"customerID"`
	AddonSpsLevel    string
	AddonPatchLevel  string
	TargetVectorID   string
	Repositories     []Repository `json:"repositories"`
}

// Repository contains fields for the repository/component version
type Repository struct {
	Name                string `json:"name"`
	Tag                 string `json:"tag"`
	Branch              string `json:"branch"`
	CommitID            string `json:"commitID"`
	VersionYAML         string `json:"version"`
	Version             string `json:"versionAAK"`
	PackageName         string
	PackageType         string
	SpLevel             string
	PatchLevel          string
	PredecessorCommitID string
	Status              string
	Namespace           string
	SarXMLFilePath      string
}

// ReadAddonDescriptorType is the type for ReadAddonDescriptor for mocking
type ReadAddonDescriptorType func(FileName string) (AddonDescriptor, error)

// ReadAddonDescriptor parses AddonDescriptor YAML file
func ReadAddonDescriptor(FileName string) (AddonDescriptor, error) {
	var addonDescriptor AddonDescriptor
	err := addonDescriptor.initFromYmlFile(FileName)
	return addonDescriptor, err
}

// ConstructAddonDescriptorFromJSON : Create new AddonDescriptor filled with data from JSON
func ConstructAddonDescriptorFromJSON(JSON []byte) (AddonDescriptor, error) {
	var addonDescriptor AddonDescriptor
	err := addonDescriptor.initFromJSON(JSON)
	return addonDescriptor, err
}

// initFromYmlFile : Reads from file
func (me *AddonDescriptor) initFromYmlFile(FileName string) error {
	filelocation, err := filepath.Glob(FileName)
	if err != nil || len(filelocation) != 1 {
		return errors.New(fmt.Sprintf("Could not find %v", FileName))
	}

	filename, err := filepath.Abs(filelocation[0])
	if err != nil {
		return errors.New(fmt.Sprintf("Could not get path of %v", FileName))
	}

	var yamlBytes []byte
	yamlBytes, err = ioutil.ReadFile(filename)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not read %v", FileName))
	}

	var jsonBytes []byte
	jsonBytes, err = yaml.YAMLToJSON(yamlBytes)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not parse %v", FileName))
	}

	err = me.initFromJSON(jsonBytes)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not unmarshal %v", FileName))
	}

	return nil
}

// CheckAddonDescriptorForRepositories checks AddonDescriptor struct if it contains any repositories. If not it will return an error
func CheckAddonDescriptorForRepositories(addonDescriptor AddonDescriptor) error {
	//checking if parsing went wrong
	if len(addonDescriptor.Repositories) == 0 {
		return errors.New(fmt.Sprintf("AddonDescriptor doesn't contain any repositories"))
	}
	//
	emptyRepositoryCounter := 0
	for counter, repo := range addonDescriptor.Repositories {
		if reflect.DeepEqual(Repository{}, repo) {
			emptyRepositoryCounter++
		}
		if counter+1 == len(addonDescriptor.Repositories) && emptyRepositoryCounter == len(addonDescriptor.Repositories) {
			return errors.New(fmt.Sprintf("AddonDescriptor doesn't contain any repositories"))
		}
	}
	return nil
}

// initFromJSON : Init from json
func (me *AddonDescriptor) initFromJSON(JSON []byte) error {
	return json.Unmarshal(JSON, me)
}

// AsJSON : dito
func (me *AddonDescriptor) AsJSON() []byte {
	//hopefully no errors should happen here or they are covered by the users unit tests
	jsonBytes, _ := json.Marshal(me)
	return jsonBytes
}

// SetRepositories : dito
func (me *AddonDescriptor) SetRepositories(Repositories []Repository) {
	me.Repositories = Repositories
}
