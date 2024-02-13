package contrast

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/log"
)

const (
	StatusReported = "REPORTED"
	Critical       = "CRITICAL"
	High           = "HIGH"
	Medium         = "MEDIUM"
	AuditAll       = "Audit All"
	Optional       = "Optional"
	pageSize       = 100
	startPage      = 0
)

type VulnerabilitiesResponse struct {
	Size            int             `json:"size"`
	TotalElements   int             `json:"totalElements"`
	TotalPages      int             `json:"totalPages"`
	Empty           bool            `json:"empty"`
	First           bool            `json:"first"`
	Last            bool            `json:"last"`
	Vulnerabilities []Vulnerability `json:"content"`
}

type Vulnerability struct {
	Severity string `json:"severity"`
	Status   string `json:"status"`
}

type ApplicationResponse struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Path        string `json:"path"`
	Language    string `json:"language"`
	Importance  string `json:"importance"`
}

type Contrast interface {
	GetVulnerabilities() error
	GetAppInfo(appUIUrl, server string)
}

type ContrastInstance struct {
	url    string
	apiKey string
	auth   string
}

func NewContrastInstance(url, apiKey, auth string) ContrastInstance {
	return ContrastInstance{
		url:    url,
		apiKey: apiKey,
		auth:   auth,
	}
}

func (contrast *ContrastInstance) GetVulnerabilities() ([]ContrastFindings, error) {
	url := contrast.url + "/vulnerabilities"
	client := NewContrastHttpClient(contrast.apiKey, contrast.auth)

	return getVulnerabilitiesFromClient(client, url, startPage)
}

func (contrast *ContrastInstance) GetAppInfo(appUIUrl, server string) (*ApplicationInfo, error) {
	client := NewContrastHttpClient(contrast.apiKey, contrast.auth)
	app, err := getApplicationFromClient(client, contrast.url)
	if err != nil {
		log.Entry().Errorf("failed to get application from client: %v", err)
		return nil, err
	}
	app.Url = appUIUrl
	app.Server = server
	return app, nil
}

func getApplicationFromClient(client ContrastHttpClient, url string) (*ApplicationInfo, error) {
	var appResponse ApplicationResponse
	err := client.ExecuteRequest(url, nil, &appResponse)
	if err != nil {
		return nil, err
	}

	return &ApplicationInfo{
		Id:   appResponse.Id,
		Name: appResponse.Name,
	}, nil
}

func getVulnerabilitiesFromClient(client ContrastHttpClient, url string, page int) ([]ContrastFindings, error) {
	params := map[string]string{
		"page": fmt.Sprintf("%d", page),
		"size": fmt.Sprintf("%d", pageSize),
	}
	var vulnsResponse VulnerabilitiesResponse
	err := client.ExecuteRequest(url, params, &vulnsResponse)
	if err != nil {
		return nil, err
	}

	if vulnsResponse.Empty {
		log.Entry().Info("empty vulnerabilities response")
		return []ContrastFindings{}, nil
	}

	auditAllFindings, optionalFindings := getFindings(vulnsResponse.Vulnerabilities)

	if !vulnsResponse.Last {
		findings, err := getVulnerabilitiesFromClient(client, url, page+1)
		if err != nil {
			return nil, err
		}
		accumulateFindings(auditAllFindings, optionalFindings, findings)
		return findings, nil
	}
	return []ContrastFindings{auditAllFindings, optionalFindings}, nil
}

func getFindings(vulnerabilities []Vulnerability) (ContrastFindings, ContrastFindings) {
	var auditAllFindings, optionalFindings ContrastFindings
	auditAllFindings.ClassificationName = AuditAll
	optionalFindings.ClassificationName = Optional

	for _, vuln := range vulnerabilities {
		if vuln.Severity == Critical || vuln.Severity == High || vuln.Severity == Medium {
			if vuln.Status != StatusReported {
				auditAllFindings.Audited += 1
			}
			auditAllFindings.Total += 1
		} else {
			if vuln.Status != StatusReported {
				optionalFindings.Audited += 1
			}
			optionalFindings.Total += 1
		}
	}
	return auditAllFindings, optionalFindings
}

func accumulateFindings(auditAllFindings, optionalFindings ContrastFindings, contrastFindings []ContrastFindings) {
	for i, fr := range contrastFindings {
		if fr.ClassificationName == AuditAll {
			contrastFindings[i].Total += auditAllFindings.Total
			contrastFindings[i].Audited += auditAllFindings.Audited
		}
		if fr.ClassificationName == Optional {
			contrastFindings[i].Total += optionalFindings.Total
			contrastFindings[i].Audited += optionalFindings.Audited
		}
	}
}
