package gcs

import (
	"context"
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
	tasks, err := createTasks(outputParams, inputParams, gcsFolderPath, gcsSubFolder, searchFilesFunc, fileInfo)
	if err != nil {
		return fmt.Errorf("failed to create tasks: %v", err)
	}

	for _, task := range tasks {
		if err := gcsClient.UploadFile(context.Background(), gcsBucketID, task.SourcePath, task.TargetPath); err != nil {
			return fmt.Errorf("failed to persist reports: %v", err)
		}
	}
	return nil
}

func createTasks(outputParams []ReportOutputParam, inputParams map[string]string, gcsFolderPath, gcsSubFolder string, searchFilesFunc func(string) ([]string, error), fileInfo func(string) (os.FileInfo, error)) ([]Task, error) {
	var tasks []Task
	for _, param := range outputParams {
		targetFolder := path.Join(gcsFolderPath, param.StepResultType, gcsSubFolder)
		if param.ParamRef != "" {
			task, err := createTaskFromParamRef(param, inputParams, targetFolder)
			if err != nil {
				return nil, err
			}
			tasks = append(tasks, task)
		} else {
			foundTasks, err := createTasksFromFilePattern(param, targetFolder, searchFilesFunc, fileInfo)
			if err != nil {
				return nil, err
			}
			tasks = append(tasks, foundTasks...)
		}
	}
	return tasks, nil
}

func createTaskFromParamRef(param ReportOutputParam, inputParams map[string]string, targetFolder string) (Task, error) {
	paramValue, ok := inputParams[param.ParamRef]
	if !ok {
		return Task{}, fmt.Errorf("input parameter %s not found", param.ParamRef)
	}
	if paramValue == "" {
		return Task{}, fmt.Errorf("input parameter %s is empty", param.ParamRef)
	}
	return Task{SourcePath: paramValue, TargetPath: filepath.Join(targetFolder, paramValue)}, nil
}

func createTasksFromFilePattern(param ReportOutputParam, targetFolder string, searchFilesFunc func(string) ([]string, error), fileInfo func(string) (os.FileInfo, error)) ([]Task, error) {
	var tasks []Task
	foundFiles, err := searchFilesFunc(param.FilePattern)
	if err != nil {
		return nil, fmt.Errorf("error searching files with pattern %s: %v", param.FilePattern, err)
	}
	for _, sourcePath := range foundFiles {
		info, err := fileInfo(sourcePath)
		if err != nil {
			return nil, fmt.Errorf("error getting file info for %s: %v", sourcePath, err)
		}
		if info.IsDir() {
			continue
		}
		tasks = append(tasks, Task{SourcePath: sourcePath, TargetPath: filepath.Join(targetFolder, sourcePath)})
	}
	return tasks, nil
}
