package piperutils

import (
	"encoding/json"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
)

// Path - struct to serialize paths and some metadata back to the invoker
type Path struct {
	Name      string `json:"name"`
	Target    string `json:"target"`
	Mandatory bool   `json:"mandatory"`
}

// PersistReportsAndLinks stores the report paths and links in JSON format in the workspace for processing outside
func PersistReportsAndLinks(workspace string, reports, links []Path) {
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
	piperenv.SetParameter(workspace, "reports.json", string(reportList))

	linkList, err := json.Marshal(&links)
	if err != nil {
		log.Entry().Errorln("Failed to marshall links.json data for archiving")
	} else {
		piperenv.SetParameter(workspace, "links.json", string(linkList))
	}
}
