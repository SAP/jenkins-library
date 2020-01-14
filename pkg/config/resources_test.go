package config

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStructString(t *testing.T) {
	tt := []struct {
		in       InfluxResource
		expected string
	}{
		{
			in: InfluxResource{
				Name:     "TestInflux",
				StepName: "TestStep",
				Measurements: []InfluxMeasurement{
					{
						Name:   "m1",
						Fields: []InfluxMetric{{Name: "field1_1"}, {Name: "field1_2"}},
						Tags:   []InfluxMetric{{Name: "tag1_1"}, {Name: "tag1_2"}},
					},
					{
						Name:   "m2",
						Fields: []InfluxMetric{{Name: "field2_1"}, {Name: "field2_2"}},
						Tags:   []InfluxMetric{{Name: "tag2_1"}, {Name: "tag2_2"}},
					},
				},
			},
			expected: `type TestStepTestInflux struct {
	m1 struct {
		fields struct {
			field1_1 string
			field1_2 string
		}
		tags struct {
			tag1_1 string
			tag1_2 string
		}
	}
	m2 struct {
		fields struct {
			field2_1 string
			field2_2 string
		}
		tags struct {
			tag2_1 string
			tag2_2 string
		}
	}
}

func (i *TestStepTestInflux) persist(path, resourceName string) {
	measurementContent := []config.InfluxMetricContent{
		{ValType: config.InfluxField, Measurement: "m1" , Name: "field1_1", Value: &i.m1.fields.field1_1},
		{ValType: config.InfluxField, Measurement: "m1" , Name: "field1_2", Value: &i.m1.fields.field1_2},
		{ValType: config.InfluxTag, Measurement: "m1" , Name: "tag1_1", Value: &i.m1.tags.tag1_1},
		{ValType: config.InfluxTag, Measurement: "m1" , Name: "tag1_2", Value: &i.m1.tags.tag1_2},
		{ValType: config.InfluxField, Measurement: "m2" , Name: "field2_1", Value: &i.m2.fields.field2_1},
		{ValType: config.InfluxField, Measurement: "m2" , Name: "field2_2", Value: &i.m2.fields.field2_2},
		{ValType: config.InfluxTag, Measurement: "m2" , Name: "tag2_1", Value: &i.m2.tags.tag2_1},
		{ValType: config.InfluxTag, Measurement: "m2" , Name: "tag2_2", Value: &i.m2.tags.tag2_2},
	}

	errCount := 0
	for _, metric := range measurementContent {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(metric.Measurement, fmt.Sprintf("%vs", metric.ValType), metric.Name), *metric.Value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting influx environment.")
			errCount++
		}
	}
	if errCount > 0 {
		os.Exit(1)
	}
}`,
		},
	}

	for run, test := range tt {
		t.Run(fmt.Sprintf("Run %v", run), func(t *testing.T) {
			got, err := test.in.StructString()
			assert.NoError(t, err)
			assert.Equal(t, test.expected, got)
		})

	}
}
