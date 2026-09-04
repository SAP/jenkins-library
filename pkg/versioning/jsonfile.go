package versioning

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

// JSONfile defines an artifact using a json file for versioning
type JSONfile struct {
	path         string
	content      []byte
	versionField string
	readFile     func(string) ([]byte, error)
	writeFile    func(string, []byte, os.FileMode) error
}

func (j *JSONfile) init() {
	if len(j.versionField) == 0 {
		j.versionField = "version"
	}
	if j.readFile == nil {
		j.readFile = os.ReadFile
	}

	if j.writeFile == nil {
		j.writeFile = os.WriteFile
	}
}

// VersioningScheme returns the relevant versioning scheme
func (j *JSONfile) VersioningScheme() string {
	return "semver2"
}

// GetVersion returns the current version of the artifact with a JSON-based build descriptor
func (j *JSONfile) GetVersion() (string, error) {
	j.init()

	content, err := j.readFile(j.path)
	if err != nil {
		return "", fmt.Errorf("failed to read file '%v': %w", j.path, err)
	}
	j.content = content

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(content, &obj); err != nil {
		return "", fmt.Errorf("failed to read json content of file '%v': %w", j.path, err)
	}

	raw, exists := obj[j.versionField]
	if !exists || raw == nil {
		return "", nil
	}

	var version string
	if err := json.Unmarshal(raw, &version); err != nil {
		return "", fmt.Errorf("failed to parse version field '%v': %w", j.versionField, err)
	}
	return version, nil
}

// SetVersion updates the version of the artifact with a JSON-based build descriptor
func (j *JSONfile) SetVersion(version string) error {
	j.init()

	if j.content == nil {
		if _, err := j.GetVersion(); err != nil {
			return err
		}
	}

	newVal, err := json.Marshal(version)
	if err != nil {
		return fmt.Errorf("failed to marshal version value: %w", err)
	}

	updated, err := setJSONField(j.content, j.versionField, newVal)
	if err != nil {
		return fmt.Errorf("failed to create json content for '%v': %w", j.path, err)
	}

	if err := j.writeFile(j.path, updated, 0700); err != nil {
		return fmt.Errorf("failed to write file '%v': %w", j.path, err)
	}

	return nil
}

// GetCoordinates returns the coordinates
func (j *JSONfile) GetCoordinates() (Coordinates, error) {
	result := Coordinates{}
	projectVersion, err := j.GetVersion()
	if err != nil {
		return result, err
	}
	result.Version = projectVersion

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(j.content, &obj); err != nil {
		return result, fmt.Errorf("failed to read json content of file '%v': %w", j.path, err)
	}
	if raw, exists := obj["name"]; exists && raw != nil {
		var name string
		if err := json.Unmarshal(raw, &name); err == nil {
			result.ArtifactID = name
		}
	}

	return result, nil
}

// setJSONField replaces the value of a top-level key in a JSON object, preserving
// key order, then re-encodes with standard indentation.
func setJSONField(src []byte, key string, value json.RawMessage) ([]byte, error) {
	// Unmarshal into an ordered slice of key/value pairs to preserve key order.
	dec := json.NewDecoder(bytes.NewReader(src))
	dec.UseNumber()

	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return nil, fmt.Errorf("expected JSON object")
	}

	type kv struct {
		key string
		val json.RawMessage
	}
	var pairs []kv

	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		k, ok := keyTok.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected token type for key")
		}
		var rawVal json.RawMessage
		if err := dec.Decode(&rawVal); err != nil {
			return nil, err
		}
		if k == key {
			pairs = append(pairs, kv{k, value})
		} else {
			pairs = append(pairs, kv{k, rawVal})
		}
	}

	// Re-encode as an ordered object with indentation.
	var buf bytes.Buffer
	buf.WriteString("{\n")
	for i, p := range pairs {
		keyBytes, _ := json.Marshal(p.key)
		buf.WriteString("  ")
		buf.Write(keyBytes)
		buf.WriteString(": ")
		buf.Write(p.val)
		if i < len(pairs)-1 {
			buf.WriteByte(',')
		}
		buf.WriteByte('\n')
	}
	buf.WriteString("}\n")
	return buf.Bytes(), nil
}
