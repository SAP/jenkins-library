package piperutils

import (
	"encoding/json"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/gcs"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
)

// Path - struct to serialize paths and some metadata back to the invoker
type Path struct {
	Name      string `json:"name"`
	Target    string `json:"target"`
	Mandatory bool   `json:"mandatory"`
	Scope     string `json:"scope"`
}

// PersistReportsAndLinks stores the report paths and links in JSON format in the workspace for processing outside
func PersistReportsAndLinks(stepName, workspace string, reports, links []Path, gcsClient gcs.Client, bucketID string) {
	if reports == nil {
		reports = []Path{}
	}
	if links == nil {
		links = []Path{}
	}

	hasMandatoryReport := false
	for _, report := range reports {
		if report.Mandatory {
			hasMandatoryReport = true
			break
		}
	}
	reportList, err := json.Marshal(&reports)
	if err != nil {
		if hasMandatoryReport {
			log.Entry().Fatalln("Failed to marshall reports.json data for archiving")
		}
		log.Entry().Errorln("Failed to marshall reports.json data for archiving")
	}
	piperenv.SetParameter(workspace, fmt.Sprintf("%v_reports.json", stepName), string(reportList))

	linkList, err := json.Marshal(&links)
	if err != nil {
		log.Entry().Errorln("Failed to marshall links.json data for archiving")
	} else {
		piperenv.SetParameter(workspace, fmt.Sprintf("%v_links.json", stepName), string(linkList))
	}
	// upload reports to Google Cloud Storage
	if gcsClient != nil {
		for _, report := range reports {
			if err := gcsClient.UploadFile(bucketID, report.Target, report.Target); err != nil {
				log.Entry().Fatalf("Failed to upload report to GCS: %v", err)
			}
			log.Entry().Infof("Report %s was uploaded to GCS", report.Target)
		}
	}
}
