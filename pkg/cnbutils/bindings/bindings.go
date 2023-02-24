// Package bindings provides utility function to create buildpack bindings folder structures
package bindings

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	k8sjson "sigs.k8s.io/json"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
)

type binding struct {
	bindingData `json:",inline"`
	Type        string        `json:"type"`
	Data        []bindingData `json:"data"`
}

type bindingData struct {
	Key     string  `json:"key"`
	Content *string `json:"content,omitempty"`
	File    *string `json:"file,omitempty"`
	FromURL *string `json:"fromUrl,omitempty"`
}

type bindings map[string]binding

// Return error if:
// 1. Content is set + File or FromURL
// 2. File is set + FromURL or Content
// 3. FromURL is set + File or Content
// 4. Everything is set
func (b *bindingData) validate() error {
	if !validName(b.Key) {
		return fmt.Errorf("invalid key: '%s'", b.Key)
	}

	if b.Content == nil && b.File == nil && b.FromURL == nil {
		return errors.New("one of 'file', 'content' or 'fromUrl' properties must be specified")
	}

	onlyOneSet := (b.Content != nil && b.File == nil && b.FromURL == nil) ||
		(b.Content == nil && b.File != nil && b.FromURL == nil) ||
		(b.Content == nil && b.File == nil && b.FromURL != nil)

	if !onlyOneSet {
		return errors.New("only one of 'content', 'file' or 'fromUrl' can be set")
	}

	return nil
}

// ProcessBindings creates the given bindings in the platform directory
func ProcessBindings(utils cnbutils.BuildUtils, httpClient piperhttp.Sender, platformPath string, bindings map[string]interface{}) error {
	typedBindings, err := toTyped(bindings)
	if err != nil {
		return errors.Wrap(err, "error while reading bindings")
	}

	for name, binding := range typedBindings {
		if len(binding.Data) == 0 {
			return fmt.Errorf("empty binding: '%s'", name)
		}
		for _, data := range binding.Data {
			err = processBinding(utils, httpClient, platformPath, name, binding.Type, data)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func processBinding(utils cnbutils.BuildUtils, httpClient piperhttp.Sender, platformPath string, name string, bindingType string, data bindingData) error {
	err := validateBinding(name, data)
	if err != nil {
		return err
	}

	bindingDir := filepath.Join(platformPath, "bindings", name)
	err = utils.MkdirAll(bindingDir, 0755)
	if err != nil {
		return errors.Wrap(err, "failed to create binding directory")
	}

	err = utils.FileWrite(filepath.Join(bindingDir, "type"), []byte(bindingType), 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write the 'type' binding file")
	}

	if data.File != nil {
		_, err = utils.Copy(*data.File, filepath.Join(bindingDir, data.Key))
		if err != nil {
			return errors.Wrap(err, "failed to copy binding file")
		}
	} else {
		var bindingContent []byte

		if data.Content == nil {
			response, err := httpClient.SendRequest(http.MethodGet, *data.FromURL, nil, nil, nil)
			if err != nil {
				return errors.Wrap(err, "failed to load binding from url")
			}

			bindingContent, err = ioutil.ReadAll(response.Body)
			defer response.Body.Close()
			if err != nil {
				return errors.Wrap(err, "error reading response")
			}
		} else {
			bindingContent = []byte(*data.Content)
		}

		err = utils.FileWrite(filepath.Join(bindingDir, data.Key), bindingContent, 0644)
		if err != nil {
			return errors.Wrap(err, "failed to write binding")
		}
	}

	return nil
}

func validateBinding(name string, data bindingData) error {
	if !validName(name) {
		return fmt.Errorf("invalid binding name: '%s'", name)
	}

	err := data.validate()
	if err != nil {
		return errors.Wrapf(err, "failed to validate binding '%s'", name)
	}
	return nil
}

func toTyped(rawMap map[string]interface{}) (bindings, error) {
	typedBindings := bindings{}

	for name, rawBinding := range rawMap {
		var b binding

		b, err := fromRaw(rawBinding)
		if err != nil {
			return nil, errors.Wrapf(err, "could not process binding '%s'", name)
		}

		if b.Key != "" {
			b.Data = append(b.Data, bindingData{
				Key:     b.Key,
				Content: b.Content,
				File:    b.File,
				FromURL: b.FromURL,
			})
		}

		typedBindings[name] = b
	}

	return typedBindings, nil
}

func fromRaw(rawData interface{}) (binding, error) {
	var new binding

	jsonValue, err := json.Marshal(rawData)
	if err != nil {
		return binding{}, err
	}

	errs, err := k8sjson.UnmarshalStrict(jsonValue, &new, k8sjson.DisallowUnknownFields)
	if err != nil {
		return binding{}, err
	}

	if len(errs) != 0 {
		for _, e := range errs {
			if err == nil {
				err = e
			} else {
				err = errors.Wrap(err, e.Error())
			}
		}
		err = errors.Wrap(err, "validation error")
		return binding{}, err
	}

	return new, nil
}

func validName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}

	return !strings.ContainsAny(name, "/")
}
