package config

import (
	"fmt"
	"io"

	"github.com/ghodss/yaml"
)

// PipelineDefaults defines the structure of the pipeline defaults
type PipelineDefaults struct {
	Defaults []Config `json:"defaults"`
}

// ReadPipelineDefaults loads defaults and returns its content
func (d *PipelineDefaults) ReadPipelineDefaults(defaultSources []io.ReadCloser) error {

	defer func() {
		for _, def := range defaultSources {
			def.Close()
		}
	}()

	for _, def := range defaultSources {
		var c Config
		var err error

		content, err := io.ReadAll(def)
		if err != nil {
			return fmt.Errorf("error reading %v: %w", def, err)
		}

		err = yaml.Unmarshal(content, &c)
		if err != nil {
			return NewParseError(fmt.Sprintf("error unmarshalling %q: %v", content, err))
		}

		d.Defaults = append(d.Defaults, c)
	}
	return nil
}
