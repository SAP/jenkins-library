//go:build unit

package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
	t.Run("default", func(t *testing.T) {
		options := jsonApplyPatchOptions{
			Input:  "schema.json",
			Patch:  "patch.json",
			Output: "output.json",
		}
		filesMock := mock.FilesMock{}
		filesMock.AddFile("schema.json", schema)
		filesMock.AddFile("patch.json", patch)
		err := runJsonApplyPatch(&options, &filesMock)
		assert.NoError(t, err)
		patchedSchemaResult, err := filesMock.FileRead("output.json")
		assert.NoError(t, err)
		assert.JSONEq(t, string(patchedSchema), string(patchedSchemaResult))
	})

	t.Run("file does not exist", func(t *testing.T) {
		options := jsonApplyPatchOptions{
			Input:  "schema.json",
			Patch:  "patch.json",
			Output: "output.json",
		}
		filesMock := mock.FilesMock{}
		err := runJsonApplyPatch(&options, &filesMock)
		assert.Error(t, err)
	})
}
