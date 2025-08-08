package sonar

import (
	"net/http"
	"time"

	sonargo "github.com/magicsong/sonargo/sonar"

	"github.com/SAP/jenkins-library/pkg/log"
)

// EndpointCeTask API endpoint for https://sonarcloud.io/web_api/api/ce/task
const EndpointCeTask = "ce/task"

const (
	taskStatusSuccess    = "SUCCESS"
	taskStatusFailed     = "FAILED"
	taskStatusCanceled   = "CANCELED"
	taskStatusPending    = "PENDING"
	taskStatusProcessing = "IN_PROGRESS"
)

// TaskService ...
type TaskService struct {
	TaskID       string
	PollInterval time.Duration
	apiClient    *Requester
}

// GetTask ...
func (service *TaskService) GetTask(options *sonargo.CeTaskOption) (*sonargo.CeTaskObject, *http.Response, error) {
	request, err := service.apiClient.create("GET", EndpointCeTask, options)
	if err != nil {
		return nil, nil, err
	}
	// use custom HTTP client to send request
	response, err := service.apiClient.send(request)
	if response == nil && err != nil {
		return nil, nil, err
	}
	// reuse response verrification from sonargo
	err = sonargo.CheckResponse(response)
	if err != nil {
		return nil, response, err
	}
	// decode JSON response
	result := new(sonargo.CeTaskObject)
	err = service.apiClient.decode(response, result)
	if err != nil {
		return nil, response, err
	}
	return result, response, nil
}

// HasFinished ...
func (service *TaskService) HasFinished() (bool, error) {
	options := &sonargo.CeTaskOption{
		Id: service.TaskID,
		// AdditionalFields: "warnings",
	}
	result, _, err := service.GetTask(options)
	if err != nil {
		return false, err
	}
	if result.Task.Status == taskStatusPending || result.Task.Status == taskStatusProcessing {
		return false, nil
	}
	return true, nil
}

// WaitForTask ..
func (service *TaskService) WaitForTask() error {
	log.Entry().Info("waiting for SonarQube task to complete..")
	finished, err := service.HasFinished()
	if err != nil {
		return err
	}
	for !finished {
		time.Sleep(service.PollInterval)
		finished, err = service.HasFinished()
		if err != nil {
			return err
		}
	}
	log.Entry().Info("finished.")
	return nil
}

// NewTaskService returns a new instance of a service for the task API endpoint.
func NewTaskService(host, token, task string, client Sender) *TaskService {
	return &TaskService{
		TaskID:       task,
		PollInterval: 15 * time.Second,
		apiClient:    NewAPIClient(host, token, client),
	}
}
