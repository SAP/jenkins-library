package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

type pipelineCreateScanSummaryUtils interface {
	FileRead(path string) ([]byte, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
	Glob(pattern string) (matches []string, err error)
}

type pipelineCreateScanSummaryUtilsBundle struct {
	*piperutils.Files
}

func newPipelineCreateScanSummaryUtils() pipelineCreateScanSummaryUtils {
	utils := pipelineCreateScanSummaryUtilsBundle{
		Files: &piperutils.Files{},
	}
	return &utils
}

func pipelineCreateScanSummary(config pipelineCreateScanSummaryOptions, telemetryData *telemetry.CustomData) {
	utils := newPipelineCreateScanSummaryUtils()

	err := runPipelineCreateScanSummary(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("failed to create scan summary")
	}
}

func runPipelineCreateScanSummary(config *pipelineCreateScanSummaryOptions, telemetryData *telemetry.CustomData, utils pipelineCreateScanSummaryUtils) error {

	pattern := reporting.StepReportDirectory + "/*.json"
	reports, _ := utils.Glob(pattern)

	scanReports := []reporting.ScanReport{}
	for _, report := range reports {
		log.Entry().Debugf("reading file %v", report)
		reportContent, err := utils.FileRead(report)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrapf(err, "failed to read report %v", report)
		}
		scanReport := reporting.ScanReport{}
		if err = json.Unmarshal(reportContent, &scanReport); err != nil {
			return errors.Wrapf(err, "failed to parse report %v", report)
		}
		scanReports = append(scanReports, scanReport)
	}

	output := []byte{}
	if len(config.PipelineLink) > 0 {
		output = []byte(fmt.Sprintf("## Pipeline Source for Details\n\nAs listed results might be incomplete, it is crucial that you check the detailed [pipeline](%v) status.\n\n", config.PipelineLink))
	}
	for _, scanReport := range scanReports {
		if (config.FailedOnly && !scanReport.SuccessfulScan) || !config.FailedOnly {
			mdReport, _ := scanReport.ToMarkdown()
			output = append(output, mdReport...)
		}
	}

	if err := utils.FileWrite(config.OutputFilePath, output, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrapf(err, "failed to write %v", config.OutputFilePath)
	}

	return nil
}
