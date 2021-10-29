// Package bindings provides utility function to create buildpack bindings folder structures
package bindings

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
)

type Binding struct {
	Type    string  `json:"type"`
	Secret  string  `json:"secret"`
	Content *string `json:"content,omitempty"`
	File    *string `json:"file,omitempty"`
}

func ProcessBindings(utils cnbutils.BuildUtils, platformPath string, bindings map[string]interface{}) error {
	for n, v := range bindings {
		if !validName(n) {
			return fmt.Errorf("invalid binding name: %s", n)
		}

		binding, err := toStruct(v)
		if err != nil {
			return err
		}

		if !validName(binding.Secret) {
			return fmt.Errorf("invalid secret name: %s", binding.Secret)
		}

		if (binding.Content == nil && binding.File == nil) || (binding.Content != nil && binding.File != nil) {
			return errors.New("either 'file' or 'content' property must be specified for binding")
		}

		bindingDir := filepath.Join(platformPath, "bindings", n)
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

	}

	return nil
}

func toStruct(rawData interface{}) (Binding, error) {
	var b Binding
	byteData, err := json.Marshal(rawData)
	if err != nil {
		return Binding{}, err
	}

	err = json.Unmarshal(byteData, &b)
	if err != nil {
		return Binding{}, err
	}

	return b, nil
}

func validName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}

	return !strings.ContainsAny(name, "/")
}
