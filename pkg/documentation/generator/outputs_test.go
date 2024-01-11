//go:build unit
// +build unit

package generator

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestStepOutputs(t *testing.T) {
	t.Run("no resources", func(t *testing.T) {
		stepData := config.StepData{Spec: config.StepSpec{Outputs: config.StepOutputs{Resources: []config.StepResources{}}}}
		result := stepOutputs(&stepData)
		assert.Equal(t, "", result)
	})

	t.Run("with resources", func(t *testing.T) {
		stepData := config.StepData{Spec: config.StepSpec{Outputs: config.StepOutputs{Resources: []config.StepResources{
			{Name: "commonPipelineEnvironment", Type: "piperEnvironment", Parameters: []map[string]interface{}{{"name": "param1"}, {"name": "param2"}}},
			{
				Name: "influxName",
				Type: "influx",
				Parameters: []map[string]interface{}{
					{"name": "influx1", "fields": []interface{}{
						map[string]interface{}{"name": "1_1"},
						map[string]interface{}{"name": "1_2"},
					}},
					{"name": "influx2", "fields": []interface{}{
						map[string]interface{}{"name": "2_1"},
						map[string]interface{}{"name": "2_2"},
					}},
				},
			},
		}}}}
		result := stepOutputs(&stepData)
		assert.Contains(t, result, "## Outputs")
		assert.Contains(t, result, "| influxName |")
		assert.Contains(t, result, "measurement `influx1`<br /><ul>")
		assert.Contains(t, result, "measurement `influx2`<br /><ul>")
		assert.Contains(t, result, "<li>1_1</li>")
		assert.Contains(t, result, "<li>1_2</li>")
		assert.Contains(t, result, "<li>2_1</li>")
		assert.Contains(t, result, "<li>2_2</li>")
	})
}
