package sonar

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"errors"

	"github.com/SAP/jenkins-library/pkg/log"
	sonargo "github.com/magicsong/sonargo/sonar"
)

// EndpointIssuesSearch API endpoint for https://sonarcloud.io/web_api/api/measures/component
const EndpointMeasuresComponent = "measures/component"

// ComponentService ...
type ComponentService struct {
	Organization string
	Project      string
	Branch       string
	PullRequest  string
	apiClient    *Requester
}

type SonarCoverage struct {
	Coverage          float32 `json:"coverage"`
	LineCoverage      float32 `json:"lineCoverage"`
	LinesToCover      int     `json:"linesToCover"`
	UncoveredLines    int     `json:"uncoveredLines"`
	BranchCoverage    float32 `json:"branchCoverage"`
	BranchesToCover   int     `json:"branchesToCover"`
	UncoveredBranches int     `json:"uncoveredBranches"`
}

type SonarLinesOfCode struct {
	Total                int                         `json:"total"`
	LanguageDistribution []SonarLanguageDistribution `json:"languageDistribution,omitempty"`
}

type SonarLanguageDistribution struct {
	LanguageKey string `json:"languageKey,omitempty"` // Description:"key of the language as retrieved from sonarqube. All languages (key + name) are available as API https://<sonarqube-instance>/api/languages/list ",ExampleValue:"java,js,web,go"
	LinesOfCode int    `json:"linesOfCode"`
}

func (service *ComponentService) Component(options *MeasuresComponentOption) (*sonargo.MeasuresComponentObject, *http.Response, error) {
	// if PR, ignore branch name and consider PR branch name. If not PR, consider branch name
	if len(service.PullRequest) > 0 {
		options.PullRequest = service.PullRequest
	} else if len(service.Branch) > 0 {
		options.Branch = service.Branch
	}
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

func (service *ComponentService) GetLinesOfCode() (*SonarLinesOfCode, error) {
	options := MeasuresComponentOption{
		Component:  service.Project,
		MetricKeys: "ncloc_language_distribution,ncloc",
	}
	component, response, err := service.Component(&options)

	if err != nil {
		return nil, fmt.Errorf("Failed to get coverage from Sonar measures/component API: %w", err)
	}

	// reuse response verification from sonargo
	err = sonargo.CheckResponse(response)
	if err != nil {
		return nil, fmt.Errorf("Failed to get lines of code from Sonar measures/component API: %w", err)
	}
	measures := component.Component.Measures

	loc := &SonarLinesOfCode{}

	for _, element := range measures {

		var err error

		switch element.Metric {
		case "ncloc":
			loc.Total, err = parseMeasureValueInt(*element)
		case "ncloc_language_distribution":
			loc.LanguageDistribution, err = parseMeasureLanguageDistribution(*element)
		default:
			log.Entry().Debugf("Received unhandled lines of code metric from Sonar measures/component API. (Metric: %s, Value: %s)", element.Metric, element.Value)
		}
		if err != nil {
			// there was an error in the type conversion
			return nil, err
		}
	}
	return loc, nil
}

func (service *ComponentService) GetCoverage() (*SonarCoverage, error) {
	options := MeasuresComponentOption{
		Component:  service.Project,
		MetricKeys: "coverage,branch_coverage,line_coverage,uncovered_lines,lines_to_cover,conditions_to_cover,uncovered_conditions",
	}
	component, response, err := service.Component(&options)
	if err != nil {
		return nil, fmt.Errorf("Failed to get coverage from Sonar measures/component API: %w", err)
	}

	// reuse response verification from sonargo
	err = sonargo.CheckResponse(response)
	if err != nil {
		return nil, fmt.Errorf("Failed to get coverage from Sonar measures/component API: %w", err)
	}
	measures := component.Component.Measures

	cov := &SonarCoverage{}

	for _, element := range measures {

		var err error

		switch element.Metric {
		case "coverage":
			cov.Coverage, err = parseMeasureValuef32(*element)
		case "branch_coverage":
			cov.BranchCoverage, err = parseMeasureValuef32(*element)
		case "line_coverage":
			cov.LineCoverage, err = parseMeasureValuef32(*element)
		case "uncovered_lines":
			cov.UncoveredLines, err = parseMeasureValueInt(*element)
		case "lines_to_cover":
			cov.LinesToCover, err = parseMeasureValueInt(*element)
		case "conditions_to_cover":
			cov.BranchesToCover, err = parseMeasureValueInt(*element)
		case "uncovered_conditions":
			cov.UncoveredBranches, err = parseMeasureValueInt(*element)
		default:
			log.Entry().Debugf("Received unhandled coverage metric from Sonar measures/component API. (Metric: %s, Value: %s)", element.Metric, element.Value)
		}
		if err != nil {
			// there was an error in the type conversion
			return nil, err
		}
	}
	return cov, nil
}

// NewMeasuresComponentService returns a new instance of a service for the measures/component endpoint.
func NewMeasuresComponentService(host, token, project, organization, branch, pullRequest string, client Sender) *ComponentService {
	return &ComponentService{
		Organization: organization,
		Project:      project,
		Branch:       branch,
		PullRequest:  pullRequest,
		apiClient:    NewAPIClient(host, token, client),
	}
}

func parseMeasureValuef32(measure sonargo.SonarMeasure) (float32, error) {
	str := measure.Value
	f64, err := strconv.ParseFloat(str, 32)
	if err != nil {
		return 0.0, fmt.Errorf("Invalid value found in measure "+measure.Metric+": "+measure.Value, err)
	}
	return float32(f64), nil
}

func parseMeasureValueInt(measure sonargo.SonarMeasure) (int, error) {
	str := measure.Value
	val, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("Invalid value found in measure "+measure.Metric+": "+measure.Value, err)
	}
	return int(val), nil
}

func parseMeasureLanguageDistribution(measure sonargo.SonarMeasure) ([]SonarLanguageDistribution, error) {
	str := measure.Value // example: js=589;ts=16544;web=1377
	var ld []SonarLanguageDistribution
	entries := strings.Split(str, ";")

	for _, entry := range entries {

		dist := strings.Split(entry, "=")

		if len(dist) != 2 {
			return nil, errors.New("Not able to split value " + entry + " at '=' found in measure " + measure.Metric + ": " + measure.Value)
		}

		loc, err := strconv.Atoi(dist[1])
		if err != nil {
			return nil, fmt.Errorf("Not able to parse value "+dist[1]+" found in measure "+measure.Metric+": "+measure.Value, err)
		}
		ld = append(ld, SonarLanguageDistribution{LanguageKey: dist[0], LinesOfCode: loc})

	}

	return ld, nil
}
