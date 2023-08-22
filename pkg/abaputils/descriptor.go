package abaputils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

/*
 * The AddonDescriptor
 * ===================
 * contains information about the Add-on Product and the comprised Add-on Software Component Git Repositories
 *
 * Lifecycle
 * =========
 * addon.yml file is read by abapAddonAssemblyKitCheckPV|CheckCV
 * addonDesriptor is stored in CPE [commonPipelineEnvironment]
 * subsequent steps enrich and update the data in CPE
 */

// AddonDescriptor contains fields about the addonProduct
type AddonDescriptor struct {
	AddonProduct     string `json:"addonProduct"`
	AddonVersionYAML string `json:"addonVersion"`
	AddonVersion     string `json:"addonVersionAAK"`
	AddonSpsLevel    string
	AddonPatchLevel  string
	TargetVectorID   string
	Repositories     []Repository `json:"repositories"`
}

// Repository contains fields for the repository/component version
type Repository struct {
	Name                string `json:"name"`
	UseClassicCTS       bool   `json:"useClassicCTS"`
	Tag                 string `json:"tag"`
	Branch              string `json:"branch"`
	CommitID            string `json:"commitID"`
	VersionYAML         string `json:"version"`
	Version             string `json:"versionAAK"`
	AdditionalPiecelist string `json:"additionalPiecelist"`
	PackageName         string
	PackageType         string
	SpLevel             string
	PatchLevel          string
	PredecessorCommitID string
	Status              string
	Namespace           string
	SarXMLFilePath      string
	Languages           []string `json:"languages"`
	InBuildScope        bool
}

// ReadAddonDescriptorType is the type for ReadAddonDescriptor for mocking
type ReadAddonDescriptorType func(FileName string) (AddonDescriptor, error)
type readFileFunc func(FileName string) ([]byte, error)

// ReadAddonDescriptor parses AddonDescriptor YAML file
func ReadAddonDescriptor(FileName string) (AddonDescriptor, error) {
	var addonDescriptor AddonDescriptor
	err := addonDescriptor.initFromYmlFile(FileName, readFile)
	return addonDescriptor, err
}

// ConstructAddonDescriptorFromJSON : Create new AddonDescriptor filled with data from JSON
func ConstructAddonDescriptorFromJSON(JSON []byte) (AddonDescriptor, error) {
	var addonDescriptor AddonDescriptor
	err := addonDescriptor.initFromJSON(JSON)
	return addonDescriptor, err
}

func readFile(FileName string) ([]byte, error) {
	fileLocations, err := filepath.Glob(FileName)
	if err != nil || len(fileLocations) != 1 {
		return nil, errors.New(fmt.Sprintf("Could not find %v", FileName))
	}

	absoluteFilename, err := filepath.Abs(fileLocations[0])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not get path of %v", FileName))
	}

	var fileContent []byte
	fileContent, err = os.ReadFile(absoluteFilename)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not read %v", FileName))
	}

	return fileContent, nil
}

// initFromYmlFile : Reads from file
func (me *AddonDescriptor) initFromYmlFile(FileName string, readFile readFileFunc) error {
	fileContent, err := readFile(FileName)
	if err != nil {
		return err
	}

	var jsonBytes []byte
	jsonBytes, err = yaml.YAMLToJSON(fileContent)
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
		return errors.New("AddonDescriptor doesn't contain any repositories")
	}
	//
	emptyRepositoryCounter := 0
	for counter, repo := range addonDescriptor.Repositories {
		if reflect.DeepEqual(Repository{}, repo) {
			emptyRepositoryCounter++
		}
		if counter+1 == len(addonDescriptor.Repositories) && emptyRepositoryCounter == len(addonDescriptor.Repositories) {
			return errors.New("AddonDescriptor doesn't contain any repositories")
		}
	}
	return nil
}

// initFromJSON : Init from json
func (me *AddonDescriptor) initFromJSON(JSON []byte) error {
	return json.Unmarshal(JSON, me)
}

// initFromJSON : Init from json string
func (me *AddonDescriptor) InitFromJSONstring(JSONstring string) error {
	return me.initFromJSON([]byte(JSONstring))
}

// AsJSON : dito
func (me *AddonDescriptor) AsJSON() []byte {
	//hopefully no errors should happen here or they are covered by the users unit tests
	jsonBytes, _ := json.Marshal(me)
	return jsonBytes
}

// AsJSONstring : dito
func (me *AddonDescriptor) AsJSONstring() string {
	return string(me.AsJSON())
}

// SetRepositories : dito
func (me *AddonDescriptor) SetRepositories(Repositories []Repository) {
	me.Repositories = Repositories
}

// GetAakAasLanguageVector : dito
func (me *Repository) GetAakAasLanguageVector() string {
	if len(me.Languages) <= 0 {
		return `ISO-DEEN`
	}
	languageVector := `ISO-`
	for _, language := range me.Languages {
		languageVector = languageVector + language
	}
	return languageVector
}

func (me *AddonDescriptor) GetRepositoriesInBuildScope() []Repository {
	var RepositoriesInBuildScope []Repository
	for _, repo := range me.Repositories {
		if repo.InBuildScope {
			RepositoriesInBuildScope = append(RepositoriesInBuildScope, repo)
		}
	}
	return RepositoriesInBuildScope
}
