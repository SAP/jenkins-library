package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type jsonApplyPatchUtilsMock struct {
	errorOnIndent bool
}

func (j jsonApplyPatchUtilsMock) Indent(dst *bytes.Buffer, src []byte, prefix, indent string) error {
	if j.errorOnIndent {
		return fmt.Errorf("error on Indent")
	}
	return json.Indent(dst, src, prefix, indent)
}

var schema = []byte(`
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "SAP Cloud SDK pipeline_config JSON schema",
  "type": "object",
  "properties": {
    "general": {
      "type": [
        "object",
        "null"
      ],
      "properties": {
        "productiveBranch": {
          "type": "string",
          "default": "master"
        }
      }
    }
  }
}
`)

var patch = []byte(`
[
  {
    "op": "add",
    "path": "/properties/general/properties/gitCredentialsId",
    "value": {
      "type": "string"
    }
  }
]
`)

var patchedSchema = []byte(`{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "title": "SAP Cloud SDK pipeline_config JSON schema",
    "type": "object",
    "properties": {
        "general": {
            "type": [
                "object",
                "null"
            ],
            "properties": {
                "productiveBranch": {
                    "type": "string",
                    "default": "master"
                },
		"gitCredentialsId": {
                    "type": "string"
                }
            }
        }
    }
}`)

func TestSchemaPatch(t *testing.T) {
	t.Parallel()
	t.Run("default", func(t *testing.T) {
		t.Parallel()
		options := jsonApplyPatchOptions{
			Input:  "schema.json",
			Patch:  "patch.json",
			Output: "output.json",
		}
		filesMock := mock.FilesMock{}
		filesMock.AddFile("schema.json", schema)
		filesMock.AddFile("patch.json", patch)
		err := runJsonApplyPatch(&options, &filesMock, jsonApplyPatchUtilsMock{})
		assert.NoError(t, err)
		patchedSchemaResult, err := filesMock.FileRead("output.json")
		assert.NoError(t, err)
		assert.JSONEq(t, string(patchedSchema), string(patchedSchemaResult))
	})

	t.Run("error on file write", func(t *testing.T) {
		t.Parallel()
		options := jsonApplyPatchOptions{
			Input:  "schema.json",
			Patch:  "patch.json",
			Output: "output.json",
		}
		filesMock := mock.FilesMock{
			FilesWithFailingWrites: []string{"output.json"},
		}
		filesMock.AddFile("schema.json", schema)
		filesMock.AddFile("patch.json", patch)
		err := runJsonApplyPatch(&options, &filesMock, jsonApplyPatchUtilsMock{})
		assert.EqualError(t, err, "cannot write file output.json")
	})

	t.Run("error on format json shall be ignored", func(t *testing.T) {
		t.Parallel()
		options := jsonApplyPatchOptions{
			Input:  "schema.json",
			Patch:  "patch.json",
			Output: "output.json",
		}
		filesMock := mock.FilesMock{}
		filesMock.AddFile("schema.json", schema)
		filesMock.AddFile("patch.json", patch)
		err := runJsonApplyPatch(&options, &filesMock, jsonApplyPatchUtilsMock{errorOnIndent: true})
		assert.NoError(t, err)
		patchedSchemaResult, err := filesMock.FileRead("output.json")
		assert.NoError(t, err)
		assert.JSONEq(t, string(patchedSchema), string(patchedSchemaResult))
	})

	t.Run("erroneous schema", func(t *testing.T) {
		t.Parallel()
		options := jsonApplyPatchOptions{
			Input:  "schema.json",
			Patch:  "patch.json",
			Output: "output.json",
		}
		filesMock := mock.FilesMock{}
		filesMock.AddFile("schema.json", []byte(`
			{
			    "$schema": "http://json-schema.org/draft-07/schema#",
			    "title": "SAP C"
				}
			}
			`))
		filesMock.AddFile("patch.json", patch)
		err := runJsonApplyPatch(&options, &filesMock, jsonApplyPatchUtilsMock{})
		assert.EqualError(t, err, "invalid character '}' after top-level value")
	})

	t.Run("erroneous patch file", func(t *testing.T) {
		t.Parallel()
		options := jsonApplyPatchOptions{
			Input:  "schema.json",
			Patch:  "patch.json",
			Output: "output.json",
		}
		filesMock := mock.FilesMock{}
		filesMock.AddFile("schema.json", schema)
		filesMock.AddFile("patch.json", []byte(`
			[
			    {
				"op": "add",
				"path": "/properties/general/properties/gitCredentialsId",
				"value":
				    "type": "string"
				}
			]
			`))
		err := runJsonApplyPatch(&options, &filesMock, jsonApplyPatchUtilsMock{})
		assert.EqualError(t, err, "invalid character ':' after object key:value pair")
	})

	t.Run("schema file does not exist", func(t *testing.T) {
		t.Parallel()
		options := jsonApplyPatchOptions{
			Input:  "schema.json",
			Patch:  "patch.json",
			Output: "output.json",
		}
		filesMock := mock.FilesMock{}
		err := runJsonApplyPatch(&options, &filesMock, jsonApplyPatchUtilsMock{})
		assert.EqualError(t, err, "could not read 'schema.json'")
	})

	t.Run("patch file does not exist", func(t *testing.T) {
		t.Parallel()
		options := jsonApplyPatchOptions{
			Input:  "schema.json",
			Patch:  "patch.json",
			Output: "output.json",
		}
		filesMock := mock.FilesMock{}
		filesMock.AddFile("schema.json", schema)
		err := runJsonApplyPatch(&options, &filesMock, jsonApplyPatchUtilsMock{})
		assert.EqualError(t, err, "could not read 'patch.json'")
	})
}
