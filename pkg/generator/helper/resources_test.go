package helper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfluxResource_StructString(t *testing.T) {
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
	measurementContent := []struct{
		measurement string
		valType     string
		name        string
		value       interface{}
	}{
		{valType: config.InfluxField, measurement: "m1" , name: "field1_1", value: i.m1.fields.field1_1},
		{valType: config.InfluxField, measurement: "m1" , name: "field1_2", value: i.m1.fields.field1_2},
		{valType: config.InfluxTag, measurement: "m1" , name: "tag1_1", value: i.m1.tags.tag1_1},
		{valType: config.InfluxTag, measurement: "m1" , name: "tag1_2", value: i.m1.tags.tag1_2},
		{valType: config.InfluxField, measurement: "m2" , name: "field2_1", value: i.m2.fields.field2_1},
		{valType: config.InfluxField, measurement: "m2" , name: "field2_2", value: i.m2.fields.field2_2},
		{valType: config.InfluxTag, measurement: "m2" , name: "tag2_1", value: i.m2.tags.tag2_1},
		{valType: config.InfluxTag, measurement: "m2" , name: "tag2_2", value: i.m2.tags.tag2_2},
	}

	errCount := 0
	for _, metric := range measurementContent {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(metric.measurement, fmt.Sprintf("%vs", metric.valType), metric.name), metric.value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting influx environment.")
			errCount++
		}
	}
	if errCount > 0 {
		log.Entry().Error("failed to persist Influx environment")
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

func TestReportsResource_StructString(t *testing.T) {
	tt := []struct {
		in       ReportsResource
		expected string
	}{
		{
			in: ReportsResource{
				Name:     "reports",
				StepName: "testStep",
				Parameters: []ReportsParameter{
					{
						FilePattern: "pattern1",
						Type:        "general",
					},
					{
						FilePattern: "pattern2",
					},
					{
						ParamRef: "testParam",
					},
				},
			},
			expected: `type testStepReports struct {
}

func (p *testStepReports) persist(stepConfig testStepOptions, gcpJsonKeyFilePath string, gcsBucketId string, gcsFolderPath string, gcsSubFolder string) {
	if gcsBucketId == "" {
		log.Entry().Info("persisting reports to GCS is disabled, because gcsBucketId is empty")
		return
	}
	log.Entry().Info("Uploading reports to Google Cloud Storage...")
	content := []gcs.ReportOutputParam{
		{FilePattern: "pattern1", ParamRef: "", StepResultType: "general"},
		{FilePattern: "pattern2", ParamRef: "", StepResultType: ""},
		{FilePattern: "", ParamRef: "testParam", StepResultType: ""},
	}

	gcsClient, err := gcs.NewClient(gcpJsonKeyFilePath, "")
	if err != nil {
		log.Entry().Errorf("creation of GCS client failed: %v", err)
        	return
	}
	defer gcsClient.Close()
	structVal := reflect.ValueOf(&stepConfig).Elem()
	inputParameters := map[string]string{}
	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Type().Field(i)
		if field.Type.String() == "string" {
			paramName := strings.Split(field.Tag.Get("json"), ",")
			paramValue, _ := structVal.Field(i).Interface().(string)
			inputParameters[paramName[0]] = paramValue
		}
	}
	if err := gcs.PersistReportsToGCS(gcsClient, content, inputParameters, gcsFolderPath, gcsBucketId, gcsSubFolder, doublestar.Glob, os.Stat); err != nil {
		log.Entry().Errorf("failed to persist reports: %v", err)
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
