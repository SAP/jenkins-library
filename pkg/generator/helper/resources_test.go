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
		log.Entry().Fatal("failed to persist Influx environment")
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
						SubFolder:   "sub/folder",
					},
					{
						FilePattern: "pattern2",
					},
				},
			},
			expected: `type testStepReports struct {
}

func (p *testStepReports) persist(path, resourceName string) {
	content := []struct{
		filePattern string
		stepResultType string
		subFolder string
	}{
		{filePattern: "pattern1", stepResultType: "general", subFolder: "sub/folder"},
		{filePattern: "pattern2", stepResultType: "", subFolder: ""},
	}

	envVars := []gcs.EnvVar{
		{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: GeneralConfig.GCPJsonKeyFilePath, Modified: false},
	}
	gcsFolderPath := GeneralConfig.GCSFolderPath
	gcsBucketID := GeneralConfig.GCSBucketId
	gcsClient, err := gcs.NewClient(gcs.WithEnvVars(envVars))
	if err != nil {
		log.Entry().Fatalf("failed to persist reports: %v", err)
	}
	for _, param := range content {
		targetFolder := gcs.GetTargetFolder(gcsFolderPath, param.stepResultType, param.subFolder)
		foundFiles, err := doublestar.Glob(param.filePattern)
		if err != nil {
			log.Entry().Fatalf("failed to persist reports: %v", err)
		}
		for _, sourcePath := range foundFiles {
			fileInfo, err := os.Stat(sourcePath)
			if err != nil {
				log.Entry().Fatalf("failed to persist reports: %v", err)
			}
			if fileInfo.IsDir() {
				continue
			}
			if err := gcsClient.UploadFile(gcsBucketID, sourcePath, filepath.Join(targetFolder, sourcePath)); err != nil {
				log.Entry().Fatalf("failed to persist reports: %v", err)
			}
		}
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
