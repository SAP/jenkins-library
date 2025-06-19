//go:build unit

package terraform

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadOutputs(t *testing.T) {
	terraformOutputsJson := `{
  "boolean": {
    "sensitive": false,
    "type": "bool",
    "value": true
  },
  "list_any": {
    "sensitive": false,
    "type": [
      "tuple",
      [
        "bool",
        "string",
        "number",
        [
          "tuple",
          []
        ]
      ]
    ],
    "value": [
      true,
      "2",
      3,
      []
    ]
  },
  "list_numbers": {
    "sensitive": false,
    "type": [
      "tuple",
      [
        "number",
        "number",
        "number"
      ]
    ],
    "value": [
      1,
      2,
      3
    ]
  },
  "list_string": {
    "sensitive": false,
    "type": [
      "tuple",
      [
        "string",
        "string",
        "string"
      ]
    ],
    "value": [
      "1",
      "2",
      "3"
    ]
  },
  "map": {
    "sensitive": false,
    "type": [
      "object",
      {
        "ATTR1": "string",
        "ATTR2": [
          "object",
          {
            "ATTR3": [
              "tuple",
              []
            ]
          }
        ]
      }
    ],
    "value": {
      "ATTR1": "",
      "ATTR2": {
        "ATTR3": []
      }
    }
  },
  "secret": {
    "sensitive": true,
    "type": "string",
    "value": "this-could-be-a-password"
  },
  "string": {
    "sensitive": false,
    "type": "string",
    "value": "string"
  },
  "number": {
    "sensitive": false,
    "type": "number",
    "value": 1
  }
}
`
	outputs, err := ReadOutputs(terraformOutputsJson)
	assert.NoError(t, err)

	assert.Equal(t, 8, len(outputs))

	assert.Equal(t, true, outputs["boolean"])
	assert.Equal(t, "string", outputs["string"])
	assert.Equal(t, float64(1), outputs["number"])
}
