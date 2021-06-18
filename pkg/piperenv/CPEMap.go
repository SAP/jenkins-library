package piperenv

import (
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// CPEMap represents the common pipeline environment map
type CPEMap map[string]interface{}

// LoadFromDisk reads the given path from disk and populates it to the CPEMap.
func (c *CPEMap) LoadFromDisk(path string) error {
	if *c == nil {
		*c = CPEMap{}
	}
	err := dirToMap(*c, path, "")
	if err != nil {
		return err
	}
	return nil
}

// WriteToDisk writes the CPEMap to a disk and uses rootDirectory as the starting point
func (c CPEMap) WriteToDisk(rootDirectory string) error {
	err := os.MkdirAll(rootDirectory, 0777)
	if err != nil {
		return err
	}

	for k, v := range c {
		entryPath := path.Join(rootDirectory, k)
		err := os.MkdirAll(filepath.Dir(entryPath), 0777)
		if err != nil {
			return err
		}
		// if v is a string no json marshalling is needed
		if vString, ok := v.(string); ok {
			err := ioutil.WriteFile(entryPath, []byte(vString), 0666)
			if err != nil {
				return err
			}
			continue
		}
		// v is not a string. serialise v to json and add '.json' suffix
		jsonVal, err := json.Marshal(v)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(fmt.Sprintf("%s.json", entryPath), jsonVal, 0666)
		if err != nil {
			return err
		}
	}
	return nil
}

func dirToMap(m map[string]interface{}, dirPath, prefix string) error {
	if stat, err := os.Stat(dirPath); err != nil || !stat.IsDir() {
		log.Entry().Debugf("stat on %s failed. Path does not exist", dirPath)
		return nil
	}

	items, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, dirItem := range items {
		if dirItem.IsDir() {
			err := dirToMap(m, path.Join(dirPath, dirItem.Name()), dirItem.Name())
			if err != nil {
				return err
			}
			continue
		}
		// load file content and unmarshal it if needed
		mapKey, value, err := readFileContent(path.Join(dirPath, dirItem.Name()))
		if err != nil {
			return err
		}
		m[path.Join(prefix, mapKey)] = value
	}
	return nil
}

func readFileContent(fullPath string) (string, interface{}, error) {
	fileContent, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return "", nil, err
	}
	fileName := filepath.Base(fullPath)

	if strings.HasSuffix(fullPath, ".json") {
		// value is json encoded
		var value interface{}
		err = json.Unmarshal(fileContent, &value)
		if err != nil {
			return "", nil, err
		}
		return strings.TrimSuffix(fileName, ".json"), value, nil
	}
	return fileName, string(fileContent), nil
}
