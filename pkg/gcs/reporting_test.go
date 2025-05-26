//go:build unit

package gcs

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/gcs/mocks"
	"github.com/bmatcuk/doublestar"
	"github.com/stretchr/testify/assert"
)

type testFileInfo struct {
	path string
}

func (t testFileInfo) Name() string {
	return ""
}

func (t testFileInfo) Size() int64 {
	return 0
}

func (t testFileInfo) Mode() os.FileMode {
	return os.FileMode(0)
}

func (t testFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (t testFileInfo) IsDir() bool {
	if strings.HasSuffix(t.path, "test2") {
		return true
	}
	return false
}

func (t testFileInfo) Sys() interface{} {
	return nil
}

type testStepConfig struct {
	FirstParameter  string
	SecondParameter int
	ThirdParameter  string
	FourthParameter bool
}

func TestPersistReportsToGCS(t *testing.T) {
	var testCases = []struct {
		testName      string
		gcsFolderPath string
		gcsSubFolder  string
		outputParams  []ReportOutputParam
		expected      []Task
		detectedFiles []string
		uploadFileErr error
		expectedError error
	}{
		{
			testName:      "success case",
			gcsFolderPath: "test/folder/path",
			gcsSubFolder:  "sub/folder",
			outputParams: []ReportOutputParam{
				{FilePattern: "*.json", ParamRef: "", StepResultType: "general"},
				{FilePattern: "*/test*", ParamRef: "", StepResultType: ""},
				{FilePattern: "*.json", ParamRef: "firstParameter", StepResultType: "general"},
				{FilePattern: "", ParamRef: "secondParameter", StepResultType: "general"},
				{FilePattern: "", ParamRef: "thirdParameter", StepResultType: ""},
			},
			expected: []Task{
				{SourcePath: "asdf.json", TargetPath: "test/folder/path/general/sub/folder/asdf.json"},
				{SourcePath: "folder/test1", TargetPath: "test/folder/path/sub/folder/folder/test1"},
				{SourcePath: "testFolder/test3", TargetPath: "test/folder/path/sub/folder/testFolder/test3"},
				{SourcePath: "report1.json", TargetPath: "test/folder/path/general/sub/folder/report1.json"},
				{SourcePath: "test-report.json", TargetPath: "test/folder/path/general/sub/folder/test-report.json"},
				{SourcePath: "test-report2.json", TargetPath: "test/folder/path/sub/folder/test-report2.json"},
			},
			detectedFiles: []string{"asdf.json", "someFolder/someFile", "folder/test1", "folder1/test2", "testFolder/test3"},
			uploadFileErr: nil,
			expectedError: nil,
		},
		{
			testName:      "failed upload to GCS",
			gcsFolderPath: "test/folder/path",
			gcsSubFolder:  "",
			outputParams: []ReportOutputParam{
				{FilePattern: "*.json", ParamRef: "", StepResultType: "general"},
			},
			expected: []Task{
				{SourcePath: "asdf.json", TargetPath: "test/folder/path/general/asdf.json"},
			},
			detectedFiles: []string{"asdf.json", "someFolder/someFile", "folder/test1", "folder1/test2", "testFolder/test3"},
			uploadFileErr: errors.New("upload failed"),
			expectedError: errors.New("failed to persist reports: upload failed"),
		},
		{
			testName:      "failed - input parameter does not exist",
			gcsFolderPath: "test/folder/path",
			gcsSubFolder:  "",
			outputParams: []ReportOutputParam{
				{FilePattern: "", ParamRef: "missingParameter", StepResultType: "general"},
			},
			expected:      []Task{},
			detectedFiles: []string{"asdf.json", "someFolder/someFile", "folder/test1", "folder1/test2", "testFolder/test3"},
			uploadFileErr: nil,
			expectedError: errors.New("failed to create tasks: input parameter missingParameter not found"),
		},
		{
			testName:      "failed - input parameter is empty",
			gcsFolderPath: "test/folder/path",
			outputParams: []ReportOutputParam{
				{FilePattern: "", ParamRef: "emptyParameter", StepResultType: "general"},
			},
			expected:      []Task{},
			detectedFiles: []string{"asdf.json", "someFolder/someFile", "folder/test1", "folder1/test2", "testFolder/test3"},
			uploadFileErr: nil,
			expectedError: errors.New("failed to create tasks: input parameter emptyParameter is empty"),
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			inputParameters := map[string]string{
				"firstParameter":  "report1.json",
				"secondParameter": "test-report.json",
				"thirdParameter":  "test-report2.json",
				"emptyParameter":  "",
			}
			gcsBucketID := "testBucketID"
			mockedClient := &mocks.Client{}

			for _, expectation := range tt.expected {
				mockedClient.Mock.On("UploadFile", gcsBucketID, expectation.SourcePath, expectation.TargetPath).Return(
					func(pipelineId string, sourcePath string, targetPath string) error { return tt.uploadFileErr },
				).Once()
			}

			searchFn := func(path string) ([]string, error) {
				matchedFiles := []string{}
				for _, value := range tt.detectedFiles {
					match, _ := doublestar.Match(path, value)
					if match {
						matchedFiles = append(matchedFiles, value)
					}
				}
				return matchedFiles, nil
			}

			fileInfoFn := func(name string) (os.FileInfo, error) {
				return testFileInfo{name}, nil
			}

			err := PersistReportsToGCS(mockedClient, tt.outputParams, inputParameters, tt.gcsFolderPath, gcsBucketID, tt.gcsSubFolder, searchFn, fileInfoFn)
			if tt.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			}

			mockedClient.Mock.AssertNumberOfCalls(t, "UploadFile", len(tt.expected))
			mockedClient.Mock.AssertExpectations(t)
		})
	}
}
