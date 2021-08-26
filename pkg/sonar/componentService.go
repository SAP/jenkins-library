package sonar

import (
	"net/http"

	"github.com/SAP/jenkins-library/pkg/log"
	sonargo "github.com/magicsong/sonargo/sonar"
)

// EndpointIssuesSearch API endpoint for https://sonarcloud.io/web_api/api/issues/search
const EndpointMeasuresComponent = "measures/component"

// ComponentService ...
type ComponentService struct {
	Organization string
	Project      string
	apiClient    *Requester
}

// GetCoverage ...
func (service *ComponentService) Component(options *sonargo.MeasuresComponentOption) (*sonargo.MeasuresComponentObject, *http.Response, error) {
	request, err := service.apiClient.create("GET", EndpointMeasuresComponent, options)
	if err != nil {
		return nil, nil, err
	}
	// use custom HTTP client to send request
	response, err := service.apiClient.send(request)
	if err != nil {
		return nil, nil, err
	}
	// reuse response verrification from sonargo
	err = sonargo.CheckResponse(response)
	if err != nil {
		return nil, response, err
	}
	// decode JSON response
	result := new(sonargo.MeasuresComponentObject)
	err = service.apiClient.decode(response, result)
	if err != nil {
		return nil, response, err
	}
	return result, response, nil
}

func (service *ComponentService) GetCoverage() string {
	options := sonargo.MeasuresComponentOption{
		Component:  service.Project,
		MetricKeys: "coverage",
	}
	component, response, err := service.Component(&options)

	// reuse response verification from sonargo
	err = sonargo.CheckResponse(response)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to get coverage from Sonar")
		return ""
	}

	return component.Component.Measures[0].Value
}

// NewMeasuresComponentService returns a new instance of a service for the measures/component endpoint.
func NewMeasuresComponentService(host, token, project, organization string, client Sender) *ComponentService {
	return &ComponentService{
		Organization: organization,
		Project:      project,
		apiClient:    NewAPIClient(host, token, client),
	}
}
