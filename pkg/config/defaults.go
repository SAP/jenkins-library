package config

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/SAP/jenkins-library/pkg/log"
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

		content, err := ioutil.ReadAll(def)
		log.Entry().Info(content, "weirdContent")
		if err != nil {
			return errors.Wrapf(err, "error reading %v", def)
		}

		err = yaml.Unmarshal(content, &c)
		if err != nil {
			return NewParseError(fmt.Sprintf("error unmarshalling %q: %v", content, err))
		}

		d.Defaults = append(d.Defaults, c)
	}
	return nil
}
