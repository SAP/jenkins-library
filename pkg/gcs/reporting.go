package gcs

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

type ReportOutputParam struct {
	FilePattern    string
	ParamRef       string
	StepResultType string
}

type Task struct {
	SourcePath string
	TargetPath string
}

func PersistReportsToGCS(gcsClient Client, outputParams []ReportOutputParam, inputParams map[string]string, gcsFolderPath string, gcsBucketID string, gcsSubFolder string, searchFilesFunc func(string) ([]string, error), fileInfo func(string) (os.FileInfo, error)) error {
	tasks := []Task{}
	for _, param := range outputParams {
		targetFolder := path.Join(gcsFolderPath, param.StepResultType, gcsSubFolder)
		if param.ParamRef != "" {
			paramValue, ok := inputParams[param.ParamRef]
			if !ok {
				return fmt.Errorf("there is no such input parameter as %s", param.ParamRef)
			}
			if paramValue == "" {
				return fmt.Errorf("the value of the parameter %s must not be empty", param.ParamRef)
			}
			tasks = append(tasks, Task{SourcePath: paramValue, TargetPath: filepath.Join(targetFolder, paramValue)})
		} else {
			foundFiles, err := searchFilesFunc(param.FilePattern)
			if err != nil {
				return fmt.Errorf("failed to persist reports: %v", err)
			}
			for _, sourcePath := range foundFiles {
				fileInfo, err := fileInfo(sourcePath)
				if err != nil {
					return fmt.Errorf("failed to persist reports: %v", err)
				}
				if fileInfo.IsDir() {
					continue
				}
				tasks = append(tasks, Task{SourcePath: sourcePath, TargetPath: filepath.Join(targetFolder, sourcePath)})
			}
		}
	}
	for _, task := range tasks {
		if err := gcsClient.UploadFile(gcsBucketID, task.SourcePath, task.TargetPath); err != nil {
			return fmt.Errorf("failed to persist reports: %v", err)
		}
	}
	return nil
}
