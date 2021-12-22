// Package bindings provides utility function to create buildpack bindings folder structures
package bindings

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/mitchellh/mapstructure"
)

type binding struct {
	Type    string  `json:"type"`
	Key     string  `json:"key"`
	Content *string `json:"content,omitempty"`
	File    *string `json:"file,omitempty"`
	FromURL *string `json:"fromUrl,omitempty"`
}

func (b *binding) validate() error {
	if !validName(b.Key) {
		return fmt.Errorf("invalid key: '%s'", b.Key)
	}

	if b.Content == nil && b.File == nil && b.FromURL == nil {
		return errors.New("one of 'file', 'content' or 'fromUrl' properties must be specified for binding")
	}

	// Return error if:
	// 1. Content is set + File or FromURL
	// 2. File is set + FromURL or Content
	// 3. FromURL is set + File or Content
	// 4. Everything is set
	if (b.Content != nil && (b.File != nil || b.FromURL != nil)) ||
		(b.File != nil && (b.FromURL != nil || b.Content != nil)) ||
		(b.FromURL != nil && (b.File != nil || b.Content != nil)) ||
		(b.Content != nil && b.File != nil && b.FromURL != nil) {
		return errors.New("only one of 'content', 'file' or 'fromUrl' can be set for a binding")
	}

	return nil
}

type bindings map[string]binding

// ProcessBindings creates the given bindings in the platform directory
func ProcessBindings(utils cnbutils.BuildUtils, httpClient piperhttp.Sender, platformPath string, bindings map[string]interface{}) error {

	typedBindings, err := toTyped(bindings)
	if err != nil {
		return errors.Wrap(err, "failed to convert map to struct")
	}

	for name, binding := range typedBindings {
		err = processBinding(utils, httpClient, platformPath, name, binding)
		if err != nil {
			return err
		}
	}

	return nil
}

func processBinding(utils cnbutils.BuildUtils, httpClient piperhttp.Sender, platformPath string, name string, binding binding) error {
	err := validateBinding(name, binding)
	if err != nil {
		return err
	}

	bindingDir := filepath.Join(platformPath, "bindings", name)
	err = utils.MkdirAll(bindingDir, 0755)
	if err != nil {
		return errors.Wrap(err, "failed to create binding directory")
	}

	err = utils.FileWrite(filepath.Join(bindingDir, "type"), []byte(binding.Type), 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write the 'type' binding file")
	}

	if binding.Content != nil {
		err = utils.FileWrite(filepath.Join(bindingDir, binding.Key), []byte(*binding.Content), 0644)
		if err != nil {
			return errors.Wrap(err, "failed to write binding")
		}
	} else if binding.File != nil {
		_, err = utils.Copy(*binding.File, filepath.Join(bindingDir, binding.Key))
		if err != nil {
			return errors.Wrap(err, "failed to copy binding file")
		}
	} else {
		response, err := httpClient.SendRequest(http.MethodGet, *binding.FromURL, nil, nil, nil)
		if err != nil {
			return errors.Wrap(err, "failed to load binding from url")
		}

		content, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return errors.Wrap(err, "error reading response")
		}
		_ = response.Body.Close()

		err = utils.FileWrite(filepath.Join(bindingDir, binding.Key), content, 0644)
		if err != nil {
			return errors.Wrap(err, "failed to write binding")
		}
	}
	return nil
}

func validateBinding(name string, binding binding) error {
	if !validName(name) {
		return fmt.Errorf("invalid binding name: '%s'", name)
	}

	return binding.validate()
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
