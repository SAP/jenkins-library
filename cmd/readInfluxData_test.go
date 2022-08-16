package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"

	"github.com/stretchr/testify/assert"
)

func TestRunReadInfluxData(t *testing.T) {
	t.Parallel()

	t.Run("success - no data", func(t *testing.T) {
		fileUtils := mock.FilesMock{}
		var b bytes.Buffer
		err := runReadInfluxData(&fileUtils, &b)

		assert.NoError(t, err)
		assert.Equal(t, "{}\n", b.String())
	})

	t.Run("success - with data", func(t *testing.T) {
		fileUtils := mock.FilesMock{}
		fileUtils.AddFile(".pipeline/influx/step_data/fields/field1", []byte("value1"))
		fileUtils.AddFile(".pipeline/influx/step_data/fields/field2", []byte("true"))
		fileUtils.AddFile(".pipeline/influx/step_data/fields/field3", []byte("false"))
		fileUtils.AddFile(".pipeline/influx/step_data/fields/field4.json", []byte("25"))
		fileUtils.AddFile(".pipeline/influx/custom_data/tags/tag1", []byte("tagValue1"))
		fileUtils.AddFile(".pipeline/influx/custom_data/tags/tag2", []byte("tagValue2"))
		var b bytes.Buffer
		err := runReadInfluxData(&fileUtils, &b)

		expectedJSON := `{
	"fields": {
		"step_data": {
			"field1": "value1",
			"field2": true,
			"field3": false,
			"field4": 25
		}
	},
	"tags": {
		"custom_data": {
			"tag1": "tagValue1",
			"tag2": "tagValue2"
		}
	}
}
`
		assert.NoError(t, err)
		assert.Equal(t, expectedJSON, b.String())
	})

	t.Run("success - ignoring invalid file path", func(t *testing.T) {
		fileUtils := mock.FilesMock{}
		fileUtils.AddFile(".pipeline/influx/step_data/field1", []byte("value1"))
		var b bytes.Buffer
		err := runReadInfluxData(&fileUtils, &b)

		assert.NoError(t, err)
	})

	t.Run("failure - readFile", func(t *testing.T) {
		fileUtils := mock.FilesMock{}
		fileUtils.AddFile(".pipeline/influx/step_data/fields/field1", []byte("value1"))
		fileUtils.FileReadErrors = map[string]error{".pipeline/influx/step_data/fields/field1": fmt.Errorf("read error")}
		var b bytes.Buffer
		err := runReadInfluxData(&fileUtils, &b)

		assert.EqualError(t, err, "failed to read file: read error")
	})

	t.Run("failure - unmarshal value", func(t *testing.T) {
		fileUtils := mock.FilesMock{}
		fileUtils.AddFile(".pipeline/influx/step_data/fields/field1.json", []byte("{value1"))
		var b bytes.Buffer
		err := runReadInfluxData(&fileUtils, &b)

		assert.Contains(t, fmt.Sprint(err), "failed to unmarshal json content of influx data file .pipeline/influx/step_data/fields/field1.json: ")
	})
}
