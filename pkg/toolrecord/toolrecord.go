package toolrecord

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

type keydataset struct {
	Name        string // technical name from the tool
	Value       string // technical value
	DisplayName string // user friendly name - optional
	URL         string // direct URL to navigate to this key in the tool backend - optional
}

// Toolrecord holds all data to locate a tool result
// in the tool's backend
type Toolrecord struct {
	RecordVersion int

	ToolName     string
	ToolInstance string

	// tool agnostic convenience aggregations
	// picks the most specific URL + concatenate the dimension names
	// for easy dashboard / xls creation
	DisplayName string
	DisplayURL  string

	fileUtils fileWriteUtils

	// detailed keydata - needs tool-specific parsing
	Keys []keydataset

	// place for additional context information
	Context map[string]interface{}

	// internal - not exported to the json
	workspace      string
	reportFileName string
}

type fileWriteUtils interface {
	MkdirAll(path string, perm fs.FileMode) error
	WriteFile(name string, data []byte, perm fs.FileMode) error
}

// New - initialize a new toolrecord
func New(fileUtils fileWriteUtils, workspace, toolName, toolInstance string) *Toolrecord {
	tr := Toolrecord{}
	tr.fileUtils = fileUtils

	tr.RecordVersion = 1
	tr.ToolName = toolName
	tr.ToolInstance = toolInstance
	tr.Keys = []keydataset{}
	tr.Context = make(map[string]interface{})

	tr.workspace = workspace

	reportFileName := filepath.Join(workspace,
		"toolruns",
		"toolrun_"+toolName+"_all.json")
	tr.reportFileName = reportFileName

	// keep a timestamp inside all files
	now := time.Now().UTC()
	var otr = &tr
	otr.AddContext("generatedOnUtc", now.Format("20060102150405"))
	return otr
}

// AddKeyData - add one key to the current toolrecord
// calls must follow the tool's hierachy ( e.g. org -> project)
// as DisplayName & DisplayURL are based on the call sequence
func (tr *Toolrecord) AddKeyData(keyname, keyvalue, displayname, url string) error {
	if keyname == "" {
		return errors.New("TR_ADD_KEY: empty keyname")
	}
	if keyvalue == "" {
		return fmt.Errorf("TR_ADD_KEY: empty keyvalue for %v", keyname)
	}
	keydata := keydataset{Name: keyname, Value: keyvalue, DisplayName: displayname, URL: url}
	tr.Keys = append(tr.Keys, keydata)
	return nil
}

// AddContext - add additional context information
// second call with the same label will overwrite the first call's data
func (tr *Toolrecord) AddContext(label string, data interface{}) error {
	if label == "" {
		return errors.New("TR_ADD_CONTEXT: no label supplied")
	}
	tr.Context[label] = data
	return nil
}

// GetFileName - local filename for the current record
func (tr *Toolrecord) GetFileName() string {
	return tr.reportFileName
}

// Persist - write the current record to file system
func (tr *Toolrecord) Persist() error {
	if tr.workspace == "" {
		return errors.New("TR_PERSIST: empty workspace ")
	}
	if tr.ToolName == "" {
		return errors.New("TR_PERSIST: empty toolName")
	}
	if tr.ToolInstance == "" {
		return errors.New("TR_PERSIST: empty instanceName")
	}
	// create workspace/toolrecord
	dirPath := filepath.Join(tr.workspace, "toolruns")
	err := tr.fileUtils.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("TR_PERSIST: %v", err)
	}

	// set default display data if required
	if tr.DisplayName == "" {
		tr.GenerateDefaultDisplayData()
	}

	file, err := json.Marshal(tr)
	if err != nil {
		return fmt.Errorf("TR_PERSIST: %v", err)
	}
	// no json generated ?
	if len(file) == 0 {
		return fmt.Errorf("TR_PERSIST: empty json content")
	}
	err = tr.fileUtils.WriteFile(tr.GetFileName(), file, 0o644)
	if err != nil {
		return fmt.Errorf("TR_PERSIST: %v", err)
	}
	return nil
}

// GenerateDefaultDisplayData - default aggregation for overall displayName and URL
// can be overriden by calling SetOverallDisplayData
func (tr *Toolrecord) GenerateDefaultDisplayData() {
	displayName := ""
	displayURL := ""
	for _, keyset := range tr.Keys {
		// create "name1 - name2 - name3"
		subDisplayName := keyset.DisplayName
		if subDisplayName != "" {
			if displayName != "" {
				displayName = displayName + " - "
			}
			displayName = displayName + subDisplayName
		}
		subURL := keyset.URL
		if subURL != "" {
			displayURL = subURL
		}
	}
	tr.DisplayName = displayName
	tr.DisplayURL = displayURL
}

// SetOverallDisplayData - override the default generation for DisplayName & DisplayURL
func (tr *Toolrecord) SetOverallDisplayData(newName, newURL string) {
	tr.DisplayName = newName
	tr.DisplayURL = newURL
}
