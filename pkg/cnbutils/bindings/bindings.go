// Package bindings provides utility function to create buildpack bindings folder structures
package bindings

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/mitchellh/mapstructure"
)

type binding struct {
	Type    string  `json:"type"`
	Secret  string  `json:"secret"`
	Content *string `json:"content,omitempty"`
	File    *string `json:"file,omitempty"`
}

type bindings map[string]binding

// ProcessBindings creates the given bindings in the platform directory
func ProcessBindings(utils cnbutils.BuildUtils, platformPath string, bindings map[string]interface{}) error {

	typedBindings, err := toTyped(bindings)
	if err != nil {
		return err
	}

	for name, binding := range typedBindings {
		err = processBinding(utils, platformPath, name, binding)
		if err != nil {
			return err
		}
	}

	return nil
}

func processBinding(utils cnbutils.BuildUtils, platformPath string, name string, binding binding) error {
	err := validateBinding(name, binding)
	if err != nil {
		return err
	}

	bindingDir := filepath.Join(platformPath, "bindings", name)
	err = utils.MkdirAll(bindingDir, 0755)
	if err != nil {
		return err
	}

	err = utils.FileWrite(filepath.Join(bindingDir, "type"), []byte(binding.Type), 0644)
	if err != nil {
		return err
	}

	if binding.Content != nil {
		err = utils.FileWrite(filepath.Join(bindingDir, binding.Secret), []byte(*binding.Content), 0644)
		if err != nil {
			return err
		}
	} else {
		_, err = utils.Copy(*binding.File, filepath.Join(bindingDir, binding.Secret))
		if err != nil {
			return err
		}
	}
	return nil
}

func validateBinding(name string, binding binding) error {
	if !validName(name) {
		return fmt.Errorf("invalid binding name: %s", name)
	}

	if !validName(binding.Secret) {
		return fmt.Errorf("invalid secret name: %s", binding.Secret)
	}

	if (binding.Content == nil && binding.File == nil) || (binding.Content != nil && binding.File != nil) {
		return errors.New("either 'file' or 'content' property must be specified for binding")
	}
	return nil
}

func toTyped(rawData interface{}) (bindings, error) {
	var typedBindings bindings

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused: true,
		Result:      &typedBindings,
	})
	if err != nil {
		return nil, err
	}
	err = decoder.Decode(rawData)
	if err != nil {
		return nil, err
	}

	return typedBindings, nil
}

func validName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}

	return !strings.ContainsAny(name, "/")
}
