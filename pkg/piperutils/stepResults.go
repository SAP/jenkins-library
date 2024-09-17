package piperutils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
)

// Path - struct to serialize paths and some metadata back to the invoker
type Path struct {
	Name      string `json:"name"`
	Target    string `json:"target"`
	Mandatory bool   `json:"mandatory"`
	Scope     string `json:"scope"`
}

type fileWriter interface {
	WriteFile(filename string, data []byte, perm os.FileMode) error
}

// PersistReportsAndLinks stores the report paths and links in JSON format in the workspace for processing outside
func PersistReportsAndLinks(stepName, workspace string, files fileWriter, reports, links []Path) error {
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
			return fmt.Errorf("failed to marshall reports.json data for archiving: %w", err)
		}
		log.Entry().Errorln("Failed to marshall reports.json data for archiving")
	}

	if err := files.WriteFile(filepath.Join(workspace, fmt.Sprintf("%v_reports.json", stepName)), reportList, 0666); err != nil {
		return fmt.Errorf("failed to write reports.json: %w", err)
	}

	linkList, err := json.Marshal(&links)
	if err != nil {
		return fmt.Errorf("failed to marshall links.json data for archiving: %w", err)
	}
	if err := files.WriteFile(filepath.Join(workspace, fmt.Sprintf("%v_links.json", stepName)), linkList, 0666); err != nil {
		return fmt.Errorf("failed to write links.json: %w", err)
	}
	return nil
}
