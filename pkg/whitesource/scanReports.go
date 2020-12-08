package whitesource

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"os"
	"path/filepath"
)

// ReportOptions defines options for downloading reports after scanning.
type ReportOptions struct {
	// ReportDirectory defines the target directory for downloading reports.
	ReportDirectory string
	// VulnerabilityReportFormat defines the requested file format of the vulnerability report (i.e. pdf).
	VulnerabilityReportFormat string
}

type scanUtils interface {
	MkdirAll(path string, perm os.FileMode) error
	FileWrite(path string, content []byte, perm os.FileMode) error
}

// DownloadReports downloads a Project's risk and vulnerability reports
func (s *Scan) DownloadReports(options ReportOptions, utils scanUtils, sys whitesource) ([]piperutils.Path, error) {
	if err := utils.MkdirAll(options.ReportDirectory, os.ModePerm); err != nil {
		return nil, err
	}

	var paths []piperutils.Path
	for _, project := range s.scannedProjects {
		vulnPath, err := downloadVulnerabilityReport(options, project, utils, sys)
		if err != nil {
			return nil, err
		}
		riskPath, err := downloadRiskReport(options, project, utils, sys)
		if err != nil {
			return nil, err
		}
		paths = append(paths, *vulnPath, *riskPath)
	}
	return paths, nil
}

func downloadVulnerabilityReport(options ReportOptions, project Project, utils scanUtils, sys whitesource) (*piperutils.Path, error) {
	reportBytes, err := sys.GetProjectVulnerabilityReport(project.Token, options.VulnerabilityReportFormat)
	if err != nil {
		return nil, err
	}

	rptFileName := fmt.Sprintf("%s-vulnerability-report.%s", project.Name, options.VulnerabilityReportFormat)
	rptFileName = filepath.Join(options.ReportDirectory, rptFileName)
	if err := utils.FileWrite(rptFileName, reportBytes, 0644); err != nil {
		return nil, err
	}

	log.Entry().Infof("Successfully downloaded vulnerability report to %s", rptFileName)
	pathName := fmt.Sprintf("%s Vulnerability Report", project.Name)
	return &piperutils.Path{Name: pathName, Target: rptFileName}, nil
}

func downloadRiskReport(options ReportOptions, project Project, utils scanUtils, sys whitesource) (*piperutils.Path, error) {
	reportBytes, err := sys.GetProjectRiskReport(project.Token)
	if err != nil {
		return nil, err
	}

	rptFileName := fmt.Sprintf("%s-risk-report.pdf", project.Name)
	rptFileName = filepath.Join(options.ReportDirectory, rptFileName)
	if err := utils.FileWrite(rptFileName, reportBytes, 0644); err != nil {
		return nil, err
	}

	log.Entry().Infof("Successfully downloaded risk report to %s", rptFileName)
	pathName := fmt.Sprintf("%s PDF Risk Report", project.Name)
	return &piperutils.Path{Name: pathName, Target: rptFileName}, nil
}
