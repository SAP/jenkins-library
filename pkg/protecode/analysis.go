package protecode

import "strconv"

const (
	vulnerabilitySeverityThreshold = 7.0
)

// HasFailed checks the return status of the provided result
func HasFailed(result ResultData) bool {
	//TODO: check this in PollForResult and return error once
	return len(result.Result.Status) > 0 && result.Result.Status == statusFailed
}

// HasSevereVulnerabilities checks if any non-historic, non-triaged, non-excluded vulnerability has a CVSS score above the defined threshold
func HasSevereVulnerabilities(result Result, excludeCVEs string) bool {
	for _, component := range result.Components {
		for _, vulnerability := range component.Vulns {
			if isSevere(vulnerability) &&
				!isTriaged(vulnerability) &&
				!isExcluded(vulnerability, excludeCVEs) &&
				isExact(vulnerability) {
				return true
			}
		}
	}
	return false
}

func isSevere(vulnerability Vulnerability) bool {
	cvss3, _ := strconv.ParseFloat(vulnerability.Vuln.Cvss3Score, 64)
	if cvss3 >= vulnerabilitySeverityThreshold {
		return true
	}
	// CVSS v3 not set, fallback to CVSS v2
	parsedCvss, _ := strconv.ParseFloat(vulnerability.Vuln.Cvss, 64)
	if cvss3 == 0 && parsedCvss >= vulnerabilitySeverityThreshold {
		return true
	}
	return false
}
