//go:build unit

package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfluxResource_StructString(t *testing.T) {
	tt := []struct {
		influxResource InfluxResource
		expected       string
	}{
		{
			influxResource: InfluxResource{
				Name:     "test_influx",
				StepName: "testStep",
				Measurements: []InfluxMeasurement{
					{
						Name: "m1",
						Fields: []InfluxMetric{
							{Name: "field1"},
							{Name: "field2", Type: "string"},
						},
						Tags: []InfluxMetric{
							{Name: "tag1"},
						},
					},
				},
			},
			expected: `type testStepTest_influx struct {
	m1 struct {
		fields struct {
			field1 string
			field2 string
		}
		tags struct {
			tag1 string
		}
	}
}

func (i *testStepTest_influx) persist(path, resourceName string) {
	measurementContent := []struct{
		measurement string
		valType     string
		name        string
		value       interface{}
	}{
		{valType: config.InfluxField, measurement: "m1" , name: "field1", value: i.m1.fields.field1},
		{valType: config.InfluxField, measurement: "m1" , name: "field2", value: i.m1.fields.field2},
		{valType: config.InfluxTag, measurement: "m1" , name: "tag1", value: i.m1.tags.tag1},
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

	for _, test := range tt {
		generatedString, err := test.influxResource.StructString()
		assert.NoError(t, err)
		assert.Equal(t, test.expected, generatedString)
	}
}

func TestReportsResource_StructString(t *testing.T) {
	tt := []struct {
		reportsResource ReportsResource
		expected        string
	}{
		{
			reportsResource: ReportsResource{
				Name:     "reports",
				StepName: "testStep",
				Parameters: []ReportsParameter{
					{FilePattern: "test.json", Type: "test-type"},
					{ParamRef: "testParam"},
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
		{FilePattern: "test.json", ParamRef: "", StepResultType: "test-type"},
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

	for _, test := range tt {
		generatedString, err := test.reportsResource.StructString()
		assert.NoError(t, err)
		assert.Equal(t, test.expected, generatedString)
	}
}
