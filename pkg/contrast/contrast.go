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

// Note: ContrastInstance is deprecated. Use the unified Client from client.go instead.
// The Client now supports both async (SARIF/PDF) and sync (vulnerabilities, app info) operations.

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
	auditAllFindings.Total = 0
	auditAllFindings.Audited = 0
	optionalFindings.ClassificationName = Optional
	optionalFindings.Total = 0
	optionalFindings.Audited = 0

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
