package sonar

import (
	"net/http"
	"strconv"

	"github.com/SAP/jenkins-library/pkg/log"
	sonargo "github.com/magicsong/sonargo/sonar"
	"github.com/pkg/errors"
)

// EndpointIssuesSearch API endpoint for https://sonarcloud.io/web_api/api/issues/search
const EndpointMeasuresComponent = "measures/component"

// ComponentService ...
type ComponentService struct {
	Organization string
	Project      string
	apiClient    *Requester
}

type SonarCoverage struct {
	Coverage       float32 `json:"coverage,omitempty"`
	LineCoverage   float32 `json:"lineCoverage,omitempty"`
	BranchCoverage float32 `json:"branchCoverage,omitempty"`
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

func (service *ComponentService) GetCoverage() (*SonarCoverage, error) {
	options := sonargo.MeasuresComponentOption{
		Component:  service.Project,
		MetricKeys: "coverage,branch_coverage,line_coverage",
	}
	component, response, _ := service.Component(&options)

	// reuse response verification from sonargo
	err := sonargo.CheckResponse(response)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get coverage from Sonar measures/component API")
	}
	measures := component.Component.Measures

	cov := &SonarCoverage{}

	for _, element := range measures {
		val, err := parseMeasureValuef32(*element)
		if err != nil {
			return nil, err
		}

		switch element.Metric {
		case "coverage":
			cov.Coverage = val
		case "branch_coverage":
			cov.BranchCoverage = val
		case "line_coverage":
			cov.LineCoverage = val
		default:
			log.Entry().Debugf("Received unhandled coverage metric from Sonar measures/component API. (Metric: %s, Value: %s)", element.Metric, element.Value)
		}
	}
	return cov, nil
}

// NewMeasuresComponentService returns a new instance of a service for the measures/component endpoint.
func NewMeasuresComponentService(host, token, project, organization string, client Sender) *ComponentService {
	return &ComponentService{
		Organization: organization,
		Project:      project,
		apiClient:    NewAPIClient(host, token, client),
	}
}

func parseMeasureValuef32(measure sonargo.SonarMeasure) (float32, error) {
	str := measure.Value
	f64, err := strconv.ParseFloat(str, 32)
	if err != nil {
		return 0.0, errors.Wrap(err, "Invalid value found in measure "+measure.Metric+": "+measure.Value)
	}
	return float32(f64), nil
}
