package piperenv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
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
			err := os.WriteFile(entryPath, []byte(vString), 0666)
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

		err = os.WriteFile(fmt.Sprintf("%s.json", entryPath), jsonVal, 0666)
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

	items, err := os.ReadDir(dirPath)
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
		mapKey, value, toBeEmptied, err := readFileContent(path.Join(dirPath, dirItem.Name()))
		if err != nil {
			return err
		}
		if toBeEmptied {
			err := addEmptyValueToFile(path.Join(dirPath, dirItem.Name()))
			if err != nil {
				return err
			}
			log.Entry().Debugf("Writing empty contents to file on disk: %v", path.Join(dirPath, dirItem.Name()))

			m[path.Join(prefix, mapKey)] = ""

		} else {
			m[path.Join(prefix, mapKey)] = value
		}
	}
	return nil
}

func addEmptyValueToFile(fullPath string) error {
	err := os.WriteFile(fullPath, []byte(""), 0666)
	if err != nil {
		return err
	}
	return nil
}

func readFileContent(fullPath string) (string, interface{}, bool, error) {
	toBeEmptied := false

	fileContent, err := os.ReadFile(fullPath)
	if err != nil {
		return "", nil, toBeEmptied, err
	}
	fileName := filepath.Base(fullPath)

	if strings.HasSuffix(fullPath, ".json") {
		// value is json encoded
		var value interface{}
		decoder := json.NewDecoder(bytes.NewReader(fileContent))
		decoder.UseNumber()
		err = decoder.Decode(&value)
		if err != nil {
			return "", nil, toBeEmptied, err
		}
		return strings.TrimSuffix(fileName, ".json"), value, toBeEmptied, nil
	}
	if string(fileContent) == "toBeEmptied" {
		toBeEmptied = true
	}
	return fileName, string(fileContent), toBeEmptied, nil
}
