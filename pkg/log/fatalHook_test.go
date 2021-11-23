package log

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestFatalHookLevels(t *testing.T) {
	hook := FatalHook{}
	assert.Equal(t, []logrus.Level{logrus.FatalLevel}, hook.Levels())
}

func TestFatalHookFire(t *testing.T) {
	workspace, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("Failed to create temporary workspace directory")
	}
	// clean up tmp dir
	defer os.RemoveAll(workspace)
	var logCollector *CollectorHook
	logCollector = &CollectorHook{CorrelationID: "local"}
	RegisterHook(logCollector)

	t.Run("with step name", func(t *testing.T) {
		hook := FatalHook{
			Path:          workspace,
			CorrelationID: "https://build.url",
		}
		entry := logrus.Entry{
			Data: logrus.Fields{
				"category": "testCategory",
				"stepName": "testStep",
			},
			Message: "the error message",
		}

		err := hook.Fire(&entry)

		assert.NoError(t, err)
		fileContent, err := ioutil.ReadFile(filepath.Join(workspace, "testStep_errorDetails.json"))
		assert.NoError(t, err)
		assert.NotContains(t, string(fileContent), `"category":"testCategory"`)
		assert.Contains(t, string(fileContent), `"correlationId":"https://build.url"`)
		assert.Contains(t, string(fileContent), `"message":"the error message"`)
		logInfoAvailable := false
		for _, message := range logCollector.Messages {
			logInfoAvailable = strings.Contains(message.Message, "fatal error: errorDetails{correlationId:\"https://build.url\",stepName:\"testStep\",category:\"undefined\",error:\"<nil>\",result:\"failure\",message:\"the error message\"}")
		}
		assert.True(t, logInfoAvailable)
	})

	t.Run("no step name", func(t *testing.T) {
		hook := FatalHook{
			Path:          workspace,
			CorrelationID: "https://build.url",
		}
		entry := logrus.Entry{
			Data: logrus.Fields{
				"category": "testCategory",
			},
			Message: "the error message",
		}

		err := hook.Fire(&entry)

		assert.NoError(t, err)
		fileContent, err := ioutil.ReadFile(filepath.Join(workspace, "errorDetails.json"))
		assert.NoError(t, err)
		assert.NotContains(t, string(fileContent), `"category":"testCategory"`)
		assert.Contains(t, string(fileContent), `"correlationId":"https://build.url"`)
		assert.Contains(t, string(fileContent), `"message":"the error message"`)
	})

	t.Run("file exists", func(t *testing.T) {
		hook := FatalHook{}
		entry := logrus.Entry{
			Message: "the new error message",
		}

		err := hook.Fire(&entry)

		assert.NoError(t, err)
		fileContent, err := ioutil.ReadFile(filepath.Join(workspace, "errorDetails.json"))
		assert.NoError(t, err)

		assert.Contains(t, string(fileContent), `"message":"the error message"`)
	})
}

func TestGetErrorsJson(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal("could not get current working directory")
	}
	defer os.Chdir(currentDir)

	tempDir, err := ioutil.TempDir("", "")

	if err != nil {
		t.Fatal("could not get tempDir")
	}
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatal("could not get change current working directory")
	}

	os.MkdirAll(tempDir+"/.pipeline/commonPipelineEnvironment", os.ModePerm)

	tests := []struct {
		name              string
		want              []ErrorDetails
		wantErr           bool
		writeErrorDetails []string
	}{
		{
			name:    "no error details available",
			want:    []ErrorDetails{},
			wantErr: false,
		},
		{
			name:              "found one error detail",
			want:              []ErrorDetails{{Message: "one error detail"}},
			wantErr:           false,
			writeErrorDetails: []string{"{\"Message\": \"one error detail\"}"},
		},
		{
			name:              "found multiple error details",
			want:              []ErrorDetails{{Message: "first error detail"}, {Message: "second error detail"}},
			wantErr:           false,
			writeErrorDetails: []string{"{\"Message\": \"first error detail\"}", "{\"Message\": \"second error detail\"}"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := os.Getwd()
			cpePath := "/.pipeline/commonPipelineEnvironment"
			var tempFiles []*os.File
			for _, content := range tt.writeErrorDetails {
				pathTempFile := path + cpePath
				tempFile, err := ioutil.TempFile(pathTempFile, "*errorDetails.json")
				tempFile.Write([]byte(content))
				if err != nil {
					t.Fatal("failed to create temporary file")
				}
				tempFiles = append(tempFiles, tempFile)
			}

			got, err := GetErrorsJson()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetErrorsJson() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetErrorsJson() got = %v, want %v", got, tt.want)
			}

			// cleanup
			for _, tempFile := range tempFiles {
				os.Remove(tempFile.Name())
			}
		})
	}
}

func Test_readErrorJson(t *testing.T) {
	type args struct {
		filePath string
	}

	tests := []struct {
		name            string
		args            args
		want            ErrorDetails
		wantErr         bool
		createTempFile  bool
		tempFileContent string
	}{
		{
			name: "readErrorJson successful",
			want: ErrorDetails{
				Message: "successful read file",
			},
			createTempFile:  true,
			tempFileContent: "{\"Message\":\"successful read file\"}",
			wantErr:         false,
		},
		{
			name:           "reads json file failure",
			createTempFile: false,
			wantErr:        true,
		},
		{
			name:            "unmarshal ErrorDetails failure",
			createTempFile:  true,
			tempFileContent: "{Message:successful read file}", // Malformed json object
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tempFile *os.File
			var err error
			if tt.createTempFile {
				tempFile, err = ioutil.TempFile("", "*errorDetails.json")
				if err != nil {
					t.Fatal("failed to create temporary file")
				}
				if _, err := tempFile.Write([]byte(tt.tempFileContent)); err != nil {
					t.Fatal("could not write content to temp file")
				}

			} else {
				tempFile, err = ioutil.TempFile("", "")
				if err != nil {
					t.Fatal("failed to create temporary file")
				}
			}

			got, err := readErrorJson(tempFile.Name())
			if (err != nil) != tt.wantErr {
				t.Errorf("readErrorJson() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readErrorJson() got = %v, want %v", got, tt.want)
			}
			os.RemoveAll(tempFile.Name())
		})
	}
}
