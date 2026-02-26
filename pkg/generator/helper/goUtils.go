package helper

import (
	"io"
	"log"
	"os"

	"github.com/ghodss/yaml"
)

// StepHelperData is used to transport the needed parameters and functions from the step generator to the step generation.
type StepHelperData struct {
	OpenFile     func(s string) (io.ReadCloser, error)
	WriteFile    func(filename string, data []byte, perm os.FileMode) error
	ExportPrefix string
}

// ContextDefaultData holds the meta data and the default data for the context default parameter descriptions
type ContextDefaultData struct {
	Metadata   ContextDefaultMetadata     `json:"metadata"`
	Parameters []ContextDefaultParameters `json:"params"`
}

// ContextDefaultMetadata holds meta data for the context default parameter descripten (name, description, long description)
type ContextDefaultMetadata struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	LongDescription string `json:"longDescription,omitempty"`
}

// ContextDefaultParameters holds the description for the context default parameters
type ContextDefaultParameters struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Scope       []string `json:"scope"`
}

// ReadPipelineContextDefaultData loads step definition in yaml format
func (c *ContextDefaultData) readPipelineContextDefaultData(metadata io.ReadCloser) {
	defer metadata.Close()
	content, err := io.ReadAll(metadata)
	if err != nil {
		log.Fatalf("Error occurred: %v\n", err)
	}
	if err = yaml.Unmarshal(content, &c); err != nil {
		log.Fatalf("Error occurred: %v\n", err)
	}
}

// ReadContextDefaultMap maps the default descriptions into a map
func (c *ContextDefaultData) readContextDefaultMap() map[string]any {
	var m = make(map[string]any)

	for _, param := range c.Parameters {
		m[param.Name] = param
	}

	return m
}
